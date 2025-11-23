package router

type Crosspoint struct {
	Destination      int  `json:"destination"`
	DestinationLevel int  `json:"destination_level"`
	Source           int  `json:"source"`
	SourceLevel      int  `json:"source_level"`
	Locked           bool `json:"locked"`
}
