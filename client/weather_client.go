package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kboeckler/pictureframe/config"
	"io/ioutil"
	"net/http"
)

func NewWeatherClient(config *config.OpenWeatherConfig) *WeatherClient {
	return &WeatherClient{config}
}

type WeatherClient struct {
	cfg *config.OpenWeatherConfig
}

func (wc *WeatherClient) GetWeather() (*WeatherResponse, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/onecall?units=metric&exclude=minutely,daily&lat=%f&lon=%f&appid=%s", wc.cfg.Lat, wc.cfg.Lon, wc.cfg.AppId)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, createError("Error creating weather request", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, createError("Error calling weather", err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, createError("Error reading weather response", err)
	}
	weatherResponse := &WeatherResponse{}
	err = json.Unmarshal(body, weatherResponse)
	if err != nil {
		return nil, createError("Error decoding weather response", err)
	}
	return weatherResponse, nil
}

func createError(msg string, cause error) error {
	return errors.New(fmt.Sprintf("%s: %v", msg, cause.Error()))
}

type AlertResponse struct {
	SenderName  string   `json:"sender_name"`
	Event       string   `json:"event"`
	Start       int64    `json:"start"`
	End         int64    `json:"end"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type WeatherMetaResponse struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type RainHourlyResponse struct {
	LastHour float64 `json:"1h"`
}

type WeatherEntryResponse struct {
	Dt         int64                 `json:"dt"`
	Sunrise    int                   `json:"sunrise"`
	Sunset     int                   `json:"sunset"`
	Temp       float64               `json:"temp"`
	FeelsLike  float64               `json:"feels_like"`
	Pressure   int                   `json:"pressure"`
	Humidity   int                   `json:"humidity"`
	DewPoint   float64               `json:"dew_point"`
	Uvi        float64               `json:"uvi"`
	Clouds     int                   `json:"clouds"`
	Visibility int                   `json:"visibility"`
	WindSpeed  float64               `json:"wind_speed"`
	WindDeg    int                   `json:"wind_deg"`
	WindGust   float64               `json:"wind_gust"`
	Rain       RainHourlyResponse    `json:"rain"`
	Weather    []WeatherMetaResponse `json:"weather"`
	Pop        float64               `json:"pop"`
}

type WeatherResponse struct {
	Lat            float64                `json:"lat"`
	Lon            float64                `json:"lon"`
	Timezone       string                 `json:"timezone"`
	TimezoneOffset int                    `json:"timezone_offset"`
	Current        WeatherEntryResponse   `json:"current"`
	Hourly         []WeatherEntryResponse `json:"hourly"`
	Alerts         []AlertResponse        `json:"alerts"`
}
