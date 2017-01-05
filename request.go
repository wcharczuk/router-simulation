package simulation

import "time"

// Request is an incoming work item
type Request struct {
	Arrival   time.Time
	Routed    time.Time
	Served    time.Time
	Completed time.Time

	WorkTime    time.Duration
	IsCacheMiss bool
	Resource    string
	ServedBy    string
}
