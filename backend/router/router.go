package router

type Router interface {
	Init(interface{})
	Start()
	Stop()
	// Channel that passes any crosspoint changes reported by the router
	// If implemented, should be buffered to not halt internal processing of the router module
	SetCrosspointNotifyChannel(chan Crosspoint)
	GetLevels() []Level
	GetSources() []Source
	GetDestinations() []Destination
	GetCrosspoints() []Crosspoint
	SetCrosspoint(Destination, Level, Source, Level) error
	LockDestination(Destination, Level) error
	UnlockDestination(Destination, Level) error
}
