package control

import (
	"github.com/kboeckler/pictureframe/config"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

type Control struct {
	config    *config.ServerConfig
	hibernate bool
}

func NewControl(config *config.ServerConfig) *Control {
	return &Control{config: config, hibernate: false}
}

func (c *Control) SetHibernate(value bool) {
	log.Infof("Setting hibernation state from %v to %v", c.hibernate, value)
	if value == true {
		c.activateHibernation()
	} else {
		c.disableHibernation()
	}
}

func (c *Control) GetHibernate() bool {
	return c.hibernate
}

func (c *Control) activateHibernation() {
	c.hibernate = true
	for _, command := range c.config.HibernateOn {
		_, err := exec.Command(command).Output()
		if err != nil {
			log.Warnf("Error executing hibernation command %s: %v", command, err)
		}
	}
}

func (c *Control) disableHibernation() {
	c.hibernate = false
	for _, command := range c.config.HibernateOff {
		_, err := exec.Command(command).Output()
		if err != nil {
			log.Warnf("Error executing hibernation command %s: %v", command, err)
		}
	}
}
