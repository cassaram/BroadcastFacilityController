package neuronview

type Head struct {
	UUID            string   `json:"uuid"`
	Name            string   `json:"name"`
	BackgroundColor string   `json:"backgroundColor"`
	BackgroundMode  string   `json:"backgroundMode"`
	Height          int      `json:"height"`
	Width           int      `json:"width"`
	Widgets         []string `json:"widgets"`
}
