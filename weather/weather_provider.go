package weather

import (
	"github.com/kboeckler/pictureframe/client"
	"github.com/kboeckler/pictureframe/control"
	log "github.com/sirupsen/logrus"
	"time"
)

type Provider interface {
	GetCurrentWeather() Weather
	GetUpcomingWeather() []Weather
}

func CreateWeatherProvider(weatherClient *client.WeatherClient, control *control.Control) Provider {
	impl := &weatherProviderImpl{}
	impl.weatherClient = weatherClient
	impl.control = control
	impl.init()
	return impl
}

type weatherProviderImpl struct {
	currentWeather  Weather
	upcomingWeather []Weather
	weatherClient   *client.WeatherClient
	control         *control.Control
}

func (wp *weatherProviderImpl) GetUpcomingWeather() []Weather {
	return wp.upcomingWeather
}

func (wp *weatherProviderImpl) GetCurrentWeather() Weather {
	return wp.currentWeather
}

func (wp *weatherProviderImpl) init() {
	go wp.loadWeather()
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			_ = <-ticker.C
			if !wp.control.GetHibernate() {
				wp.loadWeather()
			}
		}
	}()
}

func (wp *weatherProviderImpl) loadWeather() {
	weatherResponse, err := wp.weatherClient.GetWeather()
	if err != nil {
		log.Errorf("Error loading weather: %v\n", err)
		return
	}
	var alertEntry *client.AlertResponse
	if len(weatherResponse.Alerts) > 0 {
		alertEntry = &weatherResponse.Alerts[0]
	}
	wp.currentWeather = createWeather(alertEntry, weatherResponse.Current)
	wp.upcomingWeather = make([]Weather, 0)
	for i := 1; i <= 7; i++ {
		wp.upcomingWeather = append(wp.upcomingWeather, createWeather(alertEntry, weatherResponse.Hourly[i]))
	}
}

func createWeather(alertEntry *client.AlertResponse, weatherEntry client.WeatherEntryResponse) Weather {
	weatherTime := time.Unix(weatherEntry.Dt, 0)
	var alert string
	if alertEntry != nil {
		start := time.Unix(alertEntry.Start, 0)
		end := time.Unix(alertEntry.End, 0)
		if !end.Before(weatherTime) && !start.After(weatherTime) {
			alert = alertEntry.Event
		}
	}
	weatherTimeString := weatherTime.Format(time.RFC3339)
	weather := Weather{Time: weatherTimeString, Temperature: weatherEntry.Temp, Icon: weatherEntry.Weather[0].Icon, Precipitation: weatherEntry.Rain.LastHour, WindSpeed: toKmPerHour(weatherEntry.WindSpeed), Alert: alert}
	return weather
}

func toKmPerHour(meterPerSecond float64) float64 {
	return meterPerSecond * 3.6
}
