package types

//Event type represents event
type Event struct {
	Source    string `json:"source,omitempty"`
	Component string `json:"component,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Critical  int    `json:"crit,omitempty"`
	Message   string `json:"resource,omitempty"`
	Timestamp int    `json:"timestamp,omitempty"`
}
