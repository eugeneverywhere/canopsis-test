package handler

import (
	"encoding/json"
	"fmt"
	"github.com/eugeneverywhere/canopsis-test/db"
	"github.com/eugeneverywhere/canopsis-test/db/models"
	"github.com/eugeneverywhere/canopsis-test/types"
	"github.com/lillilli/logger"
)

type AlarmHandler interface {
	HandleMsg(rawMsgPayload []byte)
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
	if err := json.Unmarshal(rawMsgPayload, &event); err != nil {
		h.log.Errorf("Can't parse event %q: %v", string(rawMsgPayload), err)
		return
	}
	go h.HandleEvent(event)
}

func (h *alarmHandler) HandleEvent(event *types.Event) {
	var err error
	switch {
	case event.Critical == 0:
		err = h.ResolveAlarm(event)
	case event.Critical > 0:
		err = h.CreateOrUpdateAlarm(event)
	default:
		h.log.Info("Unexpected crit value")
	}
	if err != nil {
		h.log.Errorf("Error handling %s-%s : %s", event.Component, event.Resource, err.Error())
	}
}

func (h *alarmHandler) CreateOrUpdateAlarm(event *types.Event) error {
	err := h.db.Lock(getLockKey(event))
	if err != nil {
		return err
	}
	defer h.db.Unlock(getLockKey(event))
	alarm, err := h.db.GetAlarm(event)
	if alarm == nil {
		return h.db.CreateAlarm(event)
	}
	if event.Timestamp > alarm.LastTime {
		return h.db.UpdateAlarm(event, models.OngoingStatus)
	}
	return nil
}

func (h *alarmHandler) ResolveAlarm(event *types.Event) error {
	return h.db.UpdateAlarm(event, models.ResolvedStatus)
}

func getLockKey(event *types.Event) string {
	return fmt.Sprintf("%s%s", event.Component, event.Resource)
}
