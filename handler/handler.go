package handler

import (
	"encoding/json"
	"github.com/eugeneverywhere/canopsis-test/db"
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
		err = h.db.ResolveAlarm(event)
	case event.Critical > 0:
		err = h.db.CreateOrUpdateAlarm(event)
	default:
		h.log.Info("Unexpected crit value")
	}
	if err != nil {
		h.log.Errorf("Error handling %s-%s : %s", event.Component, event.Resource, err.Error())
	}
}
