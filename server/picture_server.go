package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kboeckler/pictureframe/config"
	"github.com/kboeckler/pictureframe/control"
	"github.com/kboeckler/pictureframe/event"
	"github.com/kboeckler/pictureframe/picture"
	"github.com/kboeckler/pictureframe/weather"
	"github.com/sinhashubham95/go-actuator"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func CreateFrameServer(serverConfig *config.ServerConfig, control *control.Control, eventProvider event.Provider, pictureProvider picture.Provider, weatherProvider weather.Provider) *FrameServer {
	return &FrameServer{serverConfig, control, eventProvider, pictureProvider, weatherProvider}
}

type FrameServer struct {
	config          *config.ServerConfig
	control         *control.Control
	eventProvider   event.Provider
	pictureProvider picture.Provider
	weatherProvider weather.Provider
}

func (pf *FrameServer) listEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(pf.eventProvider.GetEvents())
	if err != nil {
		log.Errorf("Error writing listEvents response: %v\n", err)
	}
}

func (pf *FrameServer) getPicture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, max-age=0, no-cache")
	err := pf.pictureProvider.WritePicture(w)
	if err != nil {
		log.Errorf("Error writing getPicture response: %v\n", err)
	}
}

func (pf *FrameServer) getCurrentWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(pf.weatherProvider.GetCurrentWeather())
	if err != nil {
		log.Errorf("Error writing getCurrentWeather response: %v\n", err)
	}
}

func (pf *FrameServer) getUpcomingWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(pf.weatherProvider.GetUpcomingWeather())
	if err != nil {
		log.Errorf("Error writing getUpcomingWeather response: %v\n", err)
	}
}

func (pf *FrameServer) getLog(w http.ResponseWriter, r *http.Request) {
	levelFilter, hasLevelFilter := r.URL.Query()["level"]
	entries := make([]LogEntry, 0)
	file, err := os.Open("log.json")
	if err != nil {
		log.Errorf("Unable to log config.json: %v", err)
		_, err = w.Write(make([]byte, 0))
		if err != nil {
			log.Printf("Error writing getLog empty response: %v\n", err)
		}
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Warnf("Error closing log.json file read stream: %v", err)
		}
	}(file)
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Warnf("Error reading line: %v\n", err)
			break
		}
		entry := LogEntry{}
		err = json.Unmarshal(line, &entry)
		if err != nil {
			log.Warnf("Error decoding line: %v\n", err)
			continue
		}
		if hasLevelFilter && !strings.EqualFold(entry.Level, levelFilter[0]) {
			continue
		}
		entries = append(entries, entry)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(entries)
	if err != nil {
		log.Errorf("Error writing getLog response: %v\n", err)
	}
}

func (pf *FrameServer) getOrSetHibernate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" || r.Method == "PUT" {
		var bodyAsMap map[string]interface{}
		requestBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Error reading request body: %v", err)
			w.WriteHeader(500)
		}
		err = json.Unmarshal(requestBytes, &bodyAsMap)
		if err != nil {
			w.WriteHeader(400)
			_, err := w.Write([]byte("Not a json body"))
			if err != nil {
				log.Errorf("Error writing getOrSetHibernate empty response: %v", err)
			}
			return
		}
		hibernateValue, isPresent := bodyAsMap["hibernate"]
		hibernateBool, isBool := hibernateValue.(bool)
		if !isPresent || !isBool {
			w.WriteHeader(400)
			_, err := w.Write([]byte("No boolean value \"hibernate\" present"))
			if err != nil {
				log.Errorf("Error writing getOrSetHibernate response: %v", err)
			}
			return
		}
		pf.control.SetHibernate(hibernateBool)
		w.WriteHeader(204)
	} else if r.Method == "GET" {
		hibernateResponse := struct {
			Hibernate bool `json:"hibernate"`
		}{}
		hibernateResponse.Hibernate = pf.control.GetHibernate()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := json.NewEncoder(w).Encode(hibernateResponse)
		if err != nil {
			log.Errorf("Error writing getOrSetHibernate response: %v\n", err)
		}
	} else {
		w.WriteHeader(405)
	}
}

func (pf *FrameServer) getAppState(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		appStateResponse := Appstate{}
		appStateResponse.Hibernation = HibernationState{pf.control.GetHibernate()}
		appStateResponse.CalendarMappings = pf.config.CalendarFaMappings
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := json.NewEncoder(w).Encode(appStateResponse)
		if err != nil {
			log.Errorf("Error writing getAppState response: %v\n", err)
		}
	} else {
		w.WriteHeader(405)
	}
}

func (pf *FrameServer) StartServer() {
	actuatorHandler := actuator.GetActuatorHandler(&actuator.Config{})
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Path("/log").HandlerFunc(pf.getLog)
	myRouter.Path("/app_state").HandlerFunc(pf.getAppState)
	myRouter.Path("/hibernate").HandlerFunc(pf.getOrSetHibernate)
	myRouter.PathPrefix("/actuator").Handler(actuatorHandler)
	myRouter.Path("/events").HandlerFunc(pf.listEvents)
	myRouter.Path("/picture").HandlerFunc(pf.getPicture)
	myRouter.Path("/weather").HandlerFunc(pf.getCurrentWeather)
	myRouter.Path("/upcoming_weather").HandlerFunc(pf.getUpcomingWeather)
	myRouter.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("www"))))
	address := fmt.Sprintf("%s:%d", pf.config.Ip, pf.config.Port)
	srv := &http.Server{
		Handler: myRouter,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Infof("Listening on %s\n", "http://"+address)
	log.Fatal(srv.ListenAndServe())
}
