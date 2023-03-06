package weather

type Weather struct {
	Time          string  `json:"time"`
	Temperature   float64 `json:"temperature"`
	Icon          string  `json:"icon"`
	Precipitation float64 `json:"precipitation"`
	WindSpeed     float64 `json:"wind_speed"`
	Alert         string  `json:"alert"`
}
