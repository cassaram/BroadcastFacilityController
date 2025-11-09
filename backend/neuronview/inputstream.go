package neuronview

type InputStream struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	Enable     bool   `json:"enable"`
	Info       string `json:"info"`
	NMOSLabel  string `json:"nmosLabel"`
	PreDefined bool   `json:"predefined"`
	SDP        string `json:"sdp"`
	Type       string `json:"type"`
}
