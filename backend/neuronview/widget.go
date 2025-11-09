package neuronview

type Widget struct {
	UUID      string          `json:"uuid"`
	Name      string          `json:"name"`
	Elements  []WidgetElement `json:"elements"`
	Geometry  WidgetGeometry  `json:"geometry"`
	GroupUUID string          `json:"groupUuid"`
}

type WidgetGeometry struct {
	Height float32 `json:"height"`
	Width  float32 `json:"width"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
}

type WidgetElement struct {
	Geometry   WidgetGeometry   `json:"geometry"`
	Properties WidgetProperties `json:"properties"`
	Type       string           `json:"type"`
	Visible    bool             `json:"visible"`
}

type WidgetProperties struct {
	BackgroundColor string   `json:"backgroundColor,omitempty"`
	BorderColor     string   `json:"borderColor,omitempty"`
	BorderSize      string   `json:"borderSize,omitempty"`
	Text            string   `json:"text,omitempty"`
	TextColor       string   `json:"textColor,omitempty"`
	FitMode         string   `json:"fitMode,omitempty"`
	ChannelMapping  []string `json:"channelMapping,omitempty"`
	Horizontal      string   `json:"horizontal,omitempty"`
}
