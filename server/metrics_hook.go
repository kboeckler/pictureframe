package server

import (
	"encoding/json"
	"fmt"
	"github.com/sinhashubham95/go-actuator"
	log "github.com/sirupsen/logrus"
	"io"
	"runtime"
	"time"
)

type MetricsWriter interface {
	WriteEntry(entryType string, entryContext string) error
}

func NewMetricsWriter(out io.Writer) MetricsWriter {
	return &metricsWriterImpl{out}
}

type MetricsHook struct {
	writer MetricsWriter
}

func NewMetricsHook(writer MetricsWriter) *MetricsHook {
	return &MetricsHook{writer}
}

func (m *MetricsHook) Levels() []log.Level {
	return log.AllLevels
}

func (m *MetricsHook) Fire(entry *log.Entry) error {
	_ = m.writer.WriteEntry(Log, fmt.Sprintf("%s: %s", entry.Level.String(), entry.Message))
	return nil
}

type metricsWriterImpl struct {
	out io.Writer
}

func (m *metricsWriterImpl) WriteEntry(entryType string, entryContext string) error {
	metricsEntry := MetricsDump{
		Time:    time.Now(),
		Type:    entryType,
		Context: entryContext,
		Metrics: getRuntimeMetrics(),
	}
	return json.NewEncoder(m.out).Encode(metricsEntry)
}

func getRuntimeMetrics() *actuator.MemStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	bySize := make([]actuator.BySizeElement, 0, len(memStats.BySize))
	for _, size := range memStats.BySize {
		bySize = append(bySize, actuator.BySizeElement{
			Size:         size.Size,
			MAllocations: size.Mallocs,
			Frees:        size.Frees,
		})
	}

	return &actuator.MemStats{
		Alloc:         memStats.Alloc,
		TotalAlloc:    memStats.TotalAlloc,
		Sys:           memStats.Sys,
		Lookups:       memStats.Lookups,
		MAllocations:  memStats.Mallocs,
		Frees:         memStats.Frees,
		HeapAlloc:     memStats.HeapAlloc,
		HeapSys:       memStats.HeapSys,
		HeapIdle:      memStats.HeapIdle,
		HeapInuse:     memStats.HeapInuse,
		HeapReleased:  memStats.HeapReleased,
		HeapObjects:   memStats.HeapObjects,
		StackInuse:    memStats.StackInuse,
		StackSys:      memStats.StackSys,
		MSpanInuse:    memStats.MSpanInuse,
		MSpanSys:      memStats.MSpanSys,
		MCacheInuse:   memStats.MCacheInuse,
		MCacheSys:     memStats.MCacheSys,
		BuckHashSys:   memStats.BuckHashSys,
		GCSys:         memStats.GCSys,
		OtherSys:      memStats.OtherSys,
		NextGC:        memStats.NextGC,
		LastGC:        memStats.LastGC,
		PauseTotalNs:  memStats.PauseTotalNs,
		PauseNs:       memStats.PauseNs,
		PauseEnd:      memStats.PauseEnd,
		NumGC:         memStats.NumGC,
		NumForcedGC:   memStats.NumForcedGC,
		GCCPUFraction: memStats.GCCPUFraction,
		EnableGC:      memStats.EnableGC,
		DebugGC:       memStats.DebugGC,
		BySize:        bySize,
	}

}
