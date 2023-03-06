package main

import (
	"github.com/kboeckler/pictureframe/client"
	"github.com/kboeckler/pictureframe/config"
	c "github.com/kboeckler/pictureframe/control"
	"github.com/kboeckler/pictureframe/event"
	"github.com/kboeckler/pictureframe/picture"
	"github.com/kboeckler/pictureframe/server"
	"github.com/kboeckler/pictureframe/weather"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

func main() {
	closer := setupLogger()
	defer closer.Close()
	closer = setupMetrics()
	defer closer.Close()

	log.Infoln("Pictureframe is starting")
	fullConfig := config.NewConfig("config.json")
	control := c.NewControl(fullConfig.Server)
	webDavClient := client.NewWebDavClient(fullConfig.Webdav)
	caldavClient := client.NewCalDavClient(fullConfig.Caldav)
	weatherClient := client.NewWeatherClient(fullConfig.OpenWeather)
	pictureProvider := picture.CreatePictureProvider(fullConfig.Webdav, webDavClient, control)
	eventProvider := event.CreateEventProvider(caldavClient, control)
	weatherProvider := weather.CreateWeatherProvider(weatherClient, control)
	frameServer := server.CreateFrameServer(fullConfig.Server, control, eventProvider, pictureProvider, weatherProvider)
	frameServer.StartServer()
}

func setupLogger() io.Closer {
	logWriter, err := os.OpenFile("log.json",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(logWriter)

	return logWriter
}

func setupMetrics() io.Closer {
	metricsFileWriter, err := os.OpenFile("metrics.json",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	metricsWriter := server.NewMetricsWriter(metricsFileWriter)
	log.AddHook(server.NewMetricsHook(metricsWriter))
	ticker := time.NewTicker(15 * time.Minute)
	go func() {
		for {
			_ = <-ticker.C
			_ = metricsWriter.WriteEntry(server.Periodic, "")
		}
	}()
	return metricsFileWriter
}
