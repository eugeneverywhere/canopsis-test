package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eugeneverywhere/canopsis-test/db"
	"github.com/eugeneverywhere/canopsis-test/db/models"
	"github.com/eugeneverywhere/canopsis-test/types"
	"github.com/lillilli/logger"
)

type AlarmHandler interface {
	HandleMsg(rawMsgPayload []byte)
	HandleEvent(event *types.Event) error
}

type alarmHandler struct {
	db  db.DB
	log logger.Logger
}

func New(db db.DB) AlarmHandler {
	log := logger.NewLogger("handler")
	return &alarmHandler{
		db:  db,
		log: log,
	}
}

func (h *alarmHandler) HandleMsg(rawMsgPayload []byte) {

	event := &types.Event{}
	if err := json.Unmarshal(rawMsgPayload, event); err != nil {
		h.log.Errorf("Can't parse event %q: %v", string(rawMsgPayload), err)
		return
	}
	go h.processEvent(event)
}

func (h *alarmHandler) processEvent(event *types.Event) {
	err := h.HandleEvent(event)
	if err != nil {
		h.log.Errorf("Error handling %s%s : %s", event.Component, event.Resource, err.Error())
	}
}

func (h *alarmHandler) HandleEvent(event *types.Event) error {
	h.log.Debugf("handling %s", event)
	switch {
	case event.Critical == 0:
		return h.ResolveAlarm(event)
	case event.Critical > 0:
		return h.CreateOrUpdateAlarm(event)
	default:
		return errors.New(fmt.Sprintf("Unexpected crit value: %v", event.Critical))
	}
}

func (h *alarmHandler) CreateOrUpdateAlarm(event *types.Event) error {
	err := h.db.Lock(getLockKey(event))
	if err != nil {
		return err
	}
	defer h.db.Unlock(getLockKey(event))
	alarm, err := h.db.GetAlarm(event)
	if err != nil || alarm == nil {
		return h.db.CreateAlarm(event)
	}
	if event.Timestamp > alarm.LastTime {
		return h.db.UpdateAlarm(event, models.OngoingStatus)
	}
	return nil
}

func (h *alarmHandler) ResolveAlarm(event *types.Event) error {
	h.log.Debugf("Resolving %s", event.Component+event.Resource)
	err := h.db.Lock(getLockKey(event))
	if err != nil {
		return err
	}
	defer h.db.Unlock(getLockKey(event))

	alarm, err := h.db.GetAlarm(event)
	if err != nil {
		return err
	}
	if alarm != nil && event.Timestamp > alarm.LastTime && alarm.Status == models.OngoingStatus {
		err = h.db.UpdateAlarm(event, models.ResolvedStatus)
	}
	return err
}

func getLockKey(event *types.Event) string {
	return fmt.Sprintf("%s%s", event.Component, event.Resource)
}
