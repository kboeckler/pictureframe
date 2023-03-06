package server

import (
	"github.com/sinhashubham95/go-actuator"
	"time"
)

type MetricsDump struct {
	Time    time.Time          `json:"time"`
	Type    string             `json:"type"`
	Context string             `json:"context"`
	Metrics *actuator.MemStats `json:"metrics"`
}

const (
	Log      = "LOG"
	Periodic = "PERIODIC"
)
