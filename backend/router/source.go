package router

type Source struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Levels []int  `json:"levels"`
}
