package types

import (
	"time"

	"github.com/fatih/structs"
)

type Driver interface {
	Close() error
	Batch(measurement string, fieldList []map[string]interface{}) error
	LastBuildId(slug string) int
}

type Point interface {
	GetTime() time.Time
	GetFields() map[string]interface{}
	GetTags() map[string]string
	GetMeasurement() string
}

func stripTags(point Point) map[string]interface{} {
	r := structs.Map(point)
	delete(r, "Tags")
	return r
}

type Tags map[string]string
type Build struct {
	Time     time.Time
	Status   string
	BuildId  int64
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
	Tags     Tags
}

func (p Build) GetTime() time.Time {
	return p.Time
}

func (p Build) GetFields() map[string]interface{} {
	return stripTags(p)
}

func (p Build) GetTags() map[string]string {
	return p.Tags
}

func (p Build) GetMeasurement() string {
	return "builds"
}

type Stage struct {
	Time      time.Time
	BuildId   int64
	WaitTime  int64
	Duration  int64
	Created   int64
	Started   int64
	Stopped   int64
	OS        string
	Arch      string
	Status    string
	StatusInt int64
	Name      string
	Tags      Tags
}

func (p Stage) GetTime() time.Time {
	return p.Time
}

func (p Stage) GetFields() map[string]interface{} {
	return stripTags(p)
}

func (p Stage) GetTags() map[string]string {
	return p.Tags
}

func (p Stage) GetMeasurement() string {
	return "stages"
}

type Step struct {
	Time     time.Time
	BuildId  int64
	Duration int64
	Started  int64
	Stopped  int64
	Number   int
	Name     string
	Status   string
	Tags     Tags
}

func (p Step) GetTime() time.Time {
	return p.Time
}

func (p Step) GetFields() map[string]interface{} {
	return stripTags(p)
}

func (p Step) GetTags() map[string]string {
	return p.Tags
}

func (p Step) GetMeasurement() string {
	return "steps"
}
