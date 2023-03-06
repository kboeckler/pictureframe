package event

import (
	"github.com/apognu/gocal"
	"github.com/kboeckler/pictureframe/client"
	"github.com/kboeckler/pictureframe/control"
	log "github.com/sirupsen/logrus"
	"sort"
	"time"
)

type Provider interface {
	GetEvents() []Event
}

func CreateEventProvider(caldavClient *client.CaldavClient, control *control.Control) Provider {
	impl := &eventProviderImpl{}
	impl.caldavClient = caldavClient
	impl.control = control
	impl.init()
	return impl
}

type eventProviderImpl struct {
	events       []Event
	caldavClient *client.CaldavClient
	control      *control.Control
}

func (ep *eventProviderImpl) GetEvents() []Event {
	return ep.events
}

func (ep *eventProviderImpl) init() {
	go ep.loadEvents()
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			_ = <-ticker.C
			if !ep.control.GetHibernate() {
				ep.loadEvents()
			}
		}
	}()
}

func (ep *eventProviderImpl) loadEvents() {
	log.Println("Loading events")

	items := make([]Event, 0)

	calendars, err := ep.caldavClient.FindCalendars()
	if err != nil {
		log.Errorf("Error finding calendars: %v", err)
	}
	for _, calendar := range calendars {
		calenderIcs, err := ep.caldavClient.GetCalendarExportAsIcs(calendar.Path)
		if err != nil {
			log.Error("Error trying to load ics for calendar %s: %v", calendar.Path, err)
		}

		c := gocal.NewParser(calenderIcs)

		start, end := time.Now(), time.Now().Add(3*30*24*time.Hour)
		c.Start, c.End = &start, &end
		err = c.Parse()
		if err != nil {
			log.Errorf("Error trying to parse Ics: %v", err)
		}
		err = calenderIcs.Close()
		if err != nil {
			log.Warnf("Error trying to close Ics Response: %v", err)
		}

		for _, e := range c.Events {
			date := e.Start.Local().Format(time.RFC3339)
			event := Event{Summary: e.Summary, Start: date, Type: calendar.Name}
			items = append(items, event)
		}
	}
	sort.Sort(eventList{items})
	ep.events = items
}

type eventList struct {
	events []Event
}

func (e eventList) Len() int {
	return len(e.events)
}

func (e eventList) Less(i, j int) bool {
	return e.events[i].Start < e.events[j].Start
}

func (e eventList) Swap(i, j int) {
	temp := e.events[i]
	e.events[i] = e.events[j]
	e.events[j] = temp
}
