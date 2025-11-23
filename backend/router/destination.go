package router

type Destination struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Levels []int  `json:"levels"`
}
