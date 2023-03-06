package config

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type Config struct {
	Webdav      *WebDavConfig      `json:"webdav"`
	Caldav      *CalDavConfig      `json:"caldav"`
	OpenWeather *OpenWeatherConfig `json:"openweather"`
	Server      *ServerConfig      `json:"server"`
}

type WebDavConfig struct {
	Root             string   `json:"root"`
	User             string   `json:"user"`
	Password         string   `json:"password"`
	Folders          []string `json:"folders"`
	ExcludeFolders   []string `json:"excludefolders"`
	MaxFilesizeMb    float64  `json:"maxFilesizeMb"`
	LightpictureBase string   `json:"lightpictureBase"`
}

type CalDavConfig struct {
	BaseUrl  string `json:"baseUrl"`
	HomePath string `json:"homePath"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type OpenWeatherConfig struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	AppId string  `json:"appid"`
}

type ServerConfig struct {
	Ip                 string            `json:"ip"`
	Port               int               `json:"port"`
	HibernateOff       []string          `json:"hibernateOff"`
	HibernateOn        []string          `json:"hibernateOn"`
	CalendarFaMappings map[string]string `json:"calendarFaMappings"`
}

func NewConfig(filename string) Config {
	w, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Unable to read config.json: %v", err)
	}
	config, err := configFromJSON(w)
	if err != nil {
		log.Fatalf("Unable to parse config file: %v", err)
	}
	return *config
}

const NOVALUE = -9999
const NOVALUEF = -9999.0

func configFromJSON(jsonKey []byte) (*Config, error) {
	config := createEmptyConfig()
	if err := json.Unmarshal(jsonKey, config); err != nil {
		return nil, err
	}
	webdavConfig := config.Webdav
	caldavConfig := config.Caldav
	weatherConfig := config.OpenWeather
	serverConfig := config.Server

	if len(webdavConfig.Root) < 1 {
		return nil, errors.New("webdav: missing root URL in the config file")
	}
	if len(webdavConfig.User) < 1 {
		return nil, errors.New("webdav: missing user in the webdav file")
	}
	if len(webdavConfig.Password) < 1 {
		return nil, errors.New("webdav: missing password in the config file")
	}
	if webdavConfig.Folders == nil || len(webdavConfig.Folders) < 1 {
		return nil, errors.New("webdav: missing folders in the config file")
	}
	if webdavConfig.ExcludeFolders == nil {
		return nil, errors.New("webdav: missing exclude folders in the config file")
	}
	if webdavConfig.MaxFilesizeMb == NOVALUEF {
		return nil, errors.New("webdav: missing maxfilesize in the config file")
	}
	if len(caldavConfig.BaseUrl) < 1 {
		return nil, errors.New("caldav: missing base URL in the config file")
	}
	if len(caldavConfig.HomePath) < 1 {
		return nil, errors.New("caldav: missing home path in the config file")
	}
	if len(caldavConfig.User) < 1 {
		return nil, errors.New("caldav: missing user in the config file")
	}
	if len(caldavConfig.Password) < 1 {
		return nil, errors.New("caldav: missing password in the config file")
	}
	if weatherConfig.Lat == NOVALUEF {
		return nil, errors.New("openweather: missing lat in the config file")
	}
	if weatherConfig.Lon == NOVALUEF {
		return nil, errors.New("openweather: missing lon in the config file")
	}
	if len(weatherConfig.AppId) < 1 {
		return nil, errors.New("openweather: missing apppid in the config file")
	}
	if serverConfig.Port == NOVALUE {
		return nil, errors.New("server: missing port in the config file")
	}
	if len(serverConfig.Ip) < 1 {
		return nil, errors.New("server: missing ip in the config file")
	}
	return config, nil
}

func createEmptyConfig() *Config {
	config := Config{}
	config.OpenWeather = &OpenWeatherConfig{}
	config.OpenWeather.Lat = NOVALUEF
	config.OpenWeather.Lon = NOVALUEF
	config.Webdav = &WebDavConfig{}
	config.Webdav.MaxFilesizeMb = NOVALUEF
	config.Server = &ServerConfig{}
	config.Server.Port = NOVALUE
	config.Server.CalendarFaMappings = make(map[string]string)
	config.Caldav = &CalDavConfig{}
	return &config
}
