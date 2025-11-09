package router

type Crosspoint struct {
	Source           Source
	SourceLevel      Level
	Destination      Destination
	DestinationLevel Level
	Locked           bool
}
