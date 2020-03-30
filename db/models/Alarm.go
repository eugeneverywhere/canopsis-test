package models

const OngoingStatus = "ONGOING"
const ResolvedStatus = "RESOLVED"

type Alarm struct {
	Component string `bson:"component"`
	Resource  string `bson:"resource"`
	Crit      int    `bson:"crit"`
	LastMsg   string `bson:"last_msg"`
	FirstMsg  string `bson:"first_msg"`
	StartTime int    `bson:"start_time"`
	LastTime  int    `bson:"last_time"`
	Status    string `bson:"status"`
}
