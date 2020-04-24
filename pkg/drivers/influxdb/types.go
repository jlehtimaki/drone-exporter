package influxdb

import "time"

type Build struct {
	Time     time.Time
	Number   int64
	WaitTime int64
	Duration int64
	Name     string
	Event    string
	Source   string
	Target   string
	Created  int64
	Started  int64
	Finished int64
	Tags     map[string]string
}

type Stage struct {
	Time     time.Time
	WaitTime int64
	Duration int64
	Created  int64
	Started  int64
	Stopped  int64
	OS       string
	Arch     string
	Status   string
	Name     string
	Tags     map[string]string
}

type Step struct {
	Time     time.Time
	Duration int64
	Started  int64
	Stopped  int64
	Number   int
	Name     string
	Status   string
	Tags     map[string]string
}
