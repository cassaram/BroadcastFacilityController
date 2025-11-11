package router

type Crosspoint struct {
	Destination      int
	DestinationLevel int
	Source           int
	SourceLevel      int
	Locked           bool
}
