package models

import "fmt"

const OngoingStatus = "ONGOING"
const ResolvedStatus = "RESOLVED"

type Alarm struct {
	Component string `bson:"component"`
	Resource  string `bson:"resource"`
	Crit      int    `bson:"crit"`
	LastMsg   string `bson:"last_msg"`
	FirstMsg  string `bson:"first_msg"`
	StartTime int64  `bson:"start_time"`
	LastTime  int64  `bson:"last_time"`
	Status    string `bson:"status"`
}

func (a *Alarm) String() string {
	return fmt.Sprintf("Alarm:\n  Component: %s\n  Resource: %s\n  Crit: %v\n  LastMsg: %s\n  FirstMsg:  %s\n  StartTime:  %v\n  LastTime:  %v\n  Status:  %v\n",
		a.Component, a.Resource, a.Crit, a.LastMsg, a.FirstMsg, a.StartTime, a.LastTime, a.Status)
}
