package handler

import (
	"errors"
	"fmt"
	"github.com/eugeneverywhere/canopsis-test/db/mocks"
	"github.com/eugeneverywhere/canopsis-test/db/models"
	"github.com/eugeneverywhere/canopsis-test/types"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestCreateAlarm(t *testing.T) {
	db := &mocks.DB{}
	h := New(db)

	event := &types.Event{
		Source:    "test",
		Component: "test",
		Resource:  "test",
		Critical:  3,
		Message:   "msg",
		Timestamp: 1585632165,
	}

	db.On("Lock", event.Component+event.Resource).Return(nil).Once()
	db.On("Unlock", event.Component+event.Resource).Return(nil).Once()
	db.On("GetAlarm", event).Return(nil, nil).Once()
	db.On("CreateAlarm", event).Return(nil).Once()

	err := h.HandleEvent(event)
	assert.Equal(t, err, nil)
}

func TestUpdateAlarm(t *testing.T) {
	db := &mocks.DB{}
	h := New(db)
	alarm := &models.Alarm{
		Component: "test",
		Resource:  "test",
		Crit:      3,
		LastMsg:   "msg",
		FirstMsg:  "msg",
		StartTime: 1585632165,
		LastTime:  1585632165,
		Status:    "ONGOING",
	}
	event := &types.Event{
		Source:    "test",
		Component: "test",
		Resource:  "test",
		Critical:  2,
		Message:   "msg2",
		Timestamp: 1585642165,
	}

	db.On("Lock", event.Component+event.Resource).Return(nil).Once()
	db.On("Unlock", event.Component+event.Resource).Return(nil).Once()
	db.On("GetAlarm", event).Return(alarm, nil).Once()
	db.On("UpdateAlarm", event, models.OngoingStatus).Return(nil).Once()

	err := h.HandleEvent(event)
	assert.Equal(t, err, nil)
}

func TestSkipOldEvent(t *testing.T) {
	db := &mocks.DB{}
	h := New(db)
	alarm := &models.Alarm{
		Component: "test",
		Resource:  "test",
		Crit:      3,
		LastMsg:   "msg",
		FirstMsg:  "msg",
		StartTime: 1585632165,
		LastTime:  1585632165,
		Status:    "ONGOING",
	}
	event := &types.Event{
		Source:    "test",
		Component: "test",
		Resource:  "test",
		Critical:  2,
		Message:   "msg2",
		Timestamp: 1585622165, // Less than alarm LastTime
	}

	db.On("Lock", event.Component+event.Resource).Return(nil).Once()
	db.On("Unlock", event.Component+event.Resource).Return(nil).Once()
	db.On("GetAlarm", event).Return(alarm, nil).Once()

	err := h.HandleEvent(event)
	assert.Equal(t, err, nil)
}

func TestUnexpected(t *testing.T) {
	db := &mocks.DB{}
	h := New(db)
	event := &types.Event{
		Source:    "test",
		Component: "test",
		Resource:  "test",
		Critical:  -2,
		Message:   "msg2",
		Timestamp: 1585622165,
	}

	expectedErr := errors.New(fmt.Sprintf("Unexpected crit value: %v", event.Critical))
	err := h.HandleEvent(event)
	assert.Equal(t, err, expectedErr)
}

func TestResolve(t *testing.T) {
	db := &mocks.DB{}
	h := New(db)
	alarm := &models.Alarm{
		Component: "test",
		Resource:  "test",
		Crit:      3,
		LastMsg:   "msg",
		FirstMsg:  "msg",
		StartTime: 1585632165,
		LastTime:  1585632165,
		Status:    "ONGOING",
	}
	event := &types.Event{
		Source:    "test",
		Component: "test",
		Resource:  "test",
		Critical:  0,
		Message:   "msg2",
		Timestamp: 1585652165,
	}

	db.On("Lock", event.Component+event.Resource).Return(nil).Once()
	db.On("Unlock", event.Component+event.Resource).Return(nil).Once()
	db.On("GetAlarm", event).Return(alarm, nil).Once()
	db.On("UpdateAlarm", event, models.ResolvedStatus).Return(nil).Once()

	err := h.HandleEvent(event)
	assert.Equal(t, err, nil)
}
