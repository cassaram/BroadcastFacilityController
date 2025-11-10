package router

type Router interface {
	Init(interface{})
	Start()
	Stop()
	// Returns a channel which crosspoint notifications are written to when routes are changed.
	// This channel must be cleared by the calling program regardless of if this feature is implemented.
	GetCrosspointNotifyChannel() chan Crosspoint
	GetLevels() []Level
	GetSources() []Source
	GetDestinations() []Destination
	GetCrosspoints() []Crosspoint
	SetCrosspoint(Destination, Level, Source, Level) error
	LockDestination(Destination, Level) error
	UnlockDestination(Destination, Level) error
}
