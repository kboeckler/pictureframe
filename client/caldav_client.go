package client

import (
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/kboeckler/pictureframe/config"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func NewCalDavClient(config *config.CalDavConfig) *CaldavClient {
	// TODO: extract http base auth client from webdav
	authClient := webdav.HTTPClientWithBasicAuth(http.DefaultClient, config.User, config.Password)
	caldavClient, err := caldav.NewClient(authClient, config.BaseUrl+config.HomePath)
	if err != nil {
		log.Fatalf("Error creating caldav client: %v", err)
	}
	return &CaldavClient{config, authClient, caldavClient}
}

type CaldavClient struct {
	cfg               *config.CalDavConfig
	innerAuthClient   webdav.HTTPClient
	innerCaldavClient *caldav.Client
}

func (cc *CaldavClient) GetCalendarExportAsIcs(calendarPath string) (io.ReadCloser, error) {
	request, err := http.NewRequest("GET", cc.cfg.BaseUrl+calendarPath+"?export", nil)
	resp, err := cc.innerAuthClient.Do(request)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (cc *CaldavClient) FindCalendars() ([]caldav.Calendar, error) {
	userPrincipal, err := cc.innerCaldavClient.FindCurrentUserPrincipal()
	if err != nil {
		return nil, err
	}
	set, err := cc.innerCaldavClient.FindCalendarHomeSet(userPrincipal)
	if err != nil {
		return nil, err
	}
	calendars, err := cc.innerCaldavClient.FindCalendars(set)
	if err != nil {
		return nil, err
	}
	return calendars, nil
}
