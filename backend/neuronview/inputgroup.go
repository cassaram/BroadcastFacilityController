package neuronview

type InputGroup struct {
	UUID       string            `json:"uuid"`
	Name       string            `json:"name"`
	PreDefined bool              `json:"predefined"`
	VideoUUID  string            `json:"videoUuid"`
	AudioUUID  string            `json:"audioUuid"`
	DataUUID   string            `json:"dataUuid"`
	Bindings   []ProtocolBinding `json:"bindings"`
}
