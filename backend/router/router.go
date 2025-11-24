package router

type Router interface {
	Init(map[string]interface{})
	Start()
	Stop()
	// Channel that passes any crosspoint changes reported by the router
	// If implemented, should be buffered to not halt internal processing of the router module
	SetCrosspointNotifyFunc(func(Crosspoint))
	GetLevels() []Level
	GetSources() []Source
	GetSource(srcID int) Source
	GetDestination(destID int) Destination
	GetLevel(lvlID int) Level
	GetDestinations() []Destination
	GetCrosspoints() []Crosspoint
	SetCrosspoint(destID int, destLevelID int, srcID int, srcLevelID int) error
	LockDestination(dest int, level int) error
	UnlockDestination(dest int, level int) error
}
