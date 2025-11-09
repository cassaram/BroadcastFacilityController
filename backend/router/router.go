package router

type Router interface {
	Init(interface{})
	Start()
	Stop()
	GetCrosspointNotifyChannel() chan Crosspoint
	GetLevels() []Level
	GetSources() []Source
	GetDestinations() []Destination
	GetCrosspoints() []Crosspoint
	SetCrosspoint(Destination, Level, Source, Level) error
	LockDestination(Destination, Level) error
	UnlockDestination(Destination, Level) error
}
