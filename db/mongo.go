package db

import (
	"errors"
	"fmt"
	"github.com/eugeneverywhere/canopsis-test/db/models"
	"github.com/eugeneverywhere/canopsis-test/types"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/lillilli/logger"
	lock "github.com/square/mongo-lock"
	"time"
)

const alarmCollection = "test_technique"
const maxAttempts = 10
const lockSecondsTTL = 2

type mongoDB struct {
	name    string
	url     string
	session *mgo.Session
	locker  *lock.Client
	log     logger.Logger
}

// Represents database interaction
type DB interface {
	CreateOrUpdateAlarm(event *types.Event) error
	ResolveAlarm(event *types.Event) error
	UpdateAlarm(event *types.Event, status string) error
	Lock(key string) error
	Unlock(key string) error
	Connect() error
	Close()
}

func New(url string, name string) DB {
	log := logger.NewLogger("db")
	return &mongoDB{
		name:    name,
		url:     url,
		session: nil,
		locker:  nil,
		log:     log,
	}
}

func (db *mongoDB) Lock(key string) error {
	err := db.locker.XLock("alarm", key, lock.LockDetails{TTL: lockSecondsTTL})
	if err == lock.ErrAlreadyLocked {
		for i := 0; i < maxAttempts; i++ {
			time.Sleep(lockSecondsTTL * time.Second)
			err := db.locker.XLock("alarm", key, lock.LockDetails{TTL: lockSecondsTTL})
			if err == nil {
				return nil
			}
		}
	}
	return errors.New("lock attempt timeout reached")
}

func (db *mongoDB) Unlock(key string) error {
	_, err := db.locker.Unlock(key)
	return err
}

func (db *mongoDB) CreateNewAlarm(event *types.Event) error {
	alarm := &models.Alarm{
		Component: event.Component,
		Resource:  event.Resource,
		Crit:      event.Critical,
		LastMsg:   event.Message,
		FirstMsg:  event.Message,
		StartTime: event.Timestamp,
		LastTime:  event.Timestamp,
		Status:    models.OngoingStatus,
	}
	return db.session.DB(db.name).C(alarmCollection).Insert(alarm)
}

func (db *mongoDB) ResolveAlarm(event *types.Event) error {
	return db.UpdateAlarm(event, models.ResolvedStatus)
}

func (db *mongoDB) UpdateAlarm(event *types.Event, status string) error {
	return db.session.DB(db.name).C(alarmCollection).
		Update(bson.M{"component": event.Component, "resource": event.Resource, "status": models.OngoingStatus},
			bson.M{
				"last_time": event.Timestamp,
				"last_msg":  event.Message,
				"crit":      event.Critical,
				"status":    status,
			})
}

func (db *mongoDB) CreateOrUpdateAlarm(event *types.Event) error {
	err := db.Lock(getLockKey(event))
	if err != nil {
		return err
	}
	defer db.Unlock(getLockKey(event))

	alarm := &models.Alarm{}
	err = db.session.DB(db.name).C(alarmCollection).
		Find(bson.M{"component": event.Component, "resource": event.Resource, "status": models.OngoingStatus}).
		One(alarm)
	if err != nil {
		db.log.Error(err.Error())
	}

	if alarm.Component == "" {
		return db.CreateNewAlarm(event)
	}
	if event.Timestamp > alarm.LastTime {
		return db.UpdateAlarm(event, models.OngoingStatus)
	}
	return nil
}

func (db *mongoDB) Connect() (err error) {
	if db.session != nil {
		db.session.Close()
	}
	db.session, err = mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{db.url},
		Database: db.name,
	})
	if err != nil {
		return err
	}
	db.session.SetSafe(&mgo.Safe{WMode: "majority"})

	db.locker = lock.NewClient(db.session, db.name, alarmCollection)
	_ = db.locker.CreateIndexes()

	return nil
}

func (db *mongoDB) Close() {
	db.session.Close()
}

func getLockKey(event *types.Event) string {
	return fmt.Sprintf("%s%s", event.Component, event.Resource)
}