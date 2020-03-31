package types

import "fmt"

//Event type represents event
type Event struct {
	Source    string `json:"source,omitempty"`
	Component string `json:"component,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Critical  int    `json:"crit,omitempty"`
	Message   string `json:"message,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

func (e *Event) String() string {
	return fmt.Sprintf("Event:\n  Source: %v\n  Component: %v\n  Resource: %v\n  Crit: %v\n  Msg: %v\n  Timestamp:%v\n ",
		e.Source, e.Component, e.Resource, e.Critical, e.Message, e.Timestamp)
}
