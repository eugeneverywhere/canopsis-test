package db

import (
	"errors"
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
	UpdateAlarm(event *types.Event, status string) error
	GetAlarm(event *types.Event) (*models.Alarm, error)
	CreateAlarm(event *types.Event) error
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
	db.log.Debugf("Locking: %s", key)
	err := db.locker.XLock("alarm", key, lock.LockDetails{TTL: lockSecondsTTL})
	if err == lock.ErrAlreadyLocked {
		db.log.Debugf("Already locked: %s", key)
		for i := 0; i < maxAttempts; i++ {
			time.Sleep(lockSecondsTTL * time.Second)
			err := db.locker.XLock("alarm", key, lock.LockDetails{TTL: lockSecondsTTL})
			if err == nil {
				return nil
			}
		}
	}
	if err == nil {
		return nil
	}
	return errors.New("lock attempt timeout reached")
}

func (db *mongoDB) Unlock(key string) error {
	db.log.Debugf("Unlocking: %s", key)
	_, err := db.locker.Unlock(key)
	return err
}

func (db *mongoDB) CreateAlarm(event *types.Event) error {
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
	db.log.Debugf("Creating: %s", alarm)
	return db.session.DB(db.name).C(alarmCollection).Insert(alarm)
}

func (db *mongoDB) UpdateAlarm(event *types.Event, status string) error {
	db.log.Debugf("Updating: %s", event)
	return db.session.DB(db.name).C(alarmCollection).
		Update(bson.M{"component": event.Component, "resource": event.Resource, "status": models.OngoingStatus},
			bson.M{"$set": bson.M{
				"last_time": event.Timestamp,
				"last_msg":  event.Message,
				"crit":      event.Critical,
				"status":    status,
			},
			})
}

func (db *mongoDB) GetAlarm(event *types.Event) (*models.Alarm, error) {
	alarm := &models.Alarm{}
	db.log.Debugf("Searching: %s", event.Component+event.Resource)
	err := db.session.DB(db.name).C(alarmCollection).
		Find(bson.M{"component": event.Component, "resource": event.Resource, "status": models.OngoingStatus}).
		One(alarm)
	if err != nil {
		db.log.Debugf("Not found: %s", event.Component+event.Resource)
		return nil, err
	}
	db.log.Debugf("Found: %s", alarm)
	return alarm, nil
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
