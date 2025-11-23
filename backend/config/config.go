package config

type RouterConfig struct {
	ID              int                    `json:"id"`
	DisplayName     string                 `json:"display_name"`
	ShortName       string                 `json:"short_name"`
	Type            string                 `json:"type"`
	Config          map[string]interface{} `json:"config"`
	AlternateLevels map[string][]int       `json:"alternate_levels"`
}

type ConfigFile struct {
	LogLevel string         `json:"log_level"`
	Routers  []RouterConfig `json:"routers"`
}
