package client

import (
	"fmt"
	"github.com/emersion/go-webdav"
	"github.com/kboeckler/pictureframe/config"
	"github.com/studio-b12/gowebdav"
	"io"
	"net/http"
	"os"
	"time"
)

func NewWebDavClient(config *config.WebDavConfig) *WebdavClient {
	// TODO: extract http base auth client from webdav
	authClient := webdav.HTTPClientWithBasicAuth(http.DefaultClient, config.User, config.Password)
	webdavClient := gowebdav.NewClient(config.Root, config.User, config.Password)
	webdavClient.SetTimeout(10 * time.Second)
	client := WebdavClient{cfg: config, innerAuthClient: authClient, innerWebdavClient: webdavClient}
	return &client
}

type WebdavClient struct {
	cfg               *config.WebDavConfig
	innerAuthClient   webdav.HTTPClient
	innerWebdavClient *gowebdav.Client
}

func (c *WebdavClient) ReadDir(path string) ([]os.FileInfo, error) {
	return c.innerWebdavClient.ReadDir(path)
}

func (c *WebdavClient) ReadStream(path string) (io.ReadCloser, error) {
	if len(c.cfg.LightpictureBase) == 0 {
		return c.innerWebdavClient.ReadStream(path)
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/%s?width=1280&height=720", c.cfg.LightpictureBase, path), nil)
	resp, err := c.innerAuthClient.Do(request)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
