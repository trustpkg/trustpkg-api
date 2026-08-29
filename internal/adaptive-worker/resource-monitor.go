package adaptiveWorker

import (
	"context"
	"sync"
	"time"
)

type resourceSample struct {
	at    time.Time
	usage resourcesUsage
}

type resourceSampler func() (*resourcesUsage, error)

type resourceMonitor struct {
	mu      sync.RWMutex
	maxAge  time.Duration
	sample  resourceSampler
	samples []resourceSample
}

func newResourceMonitor(sample resourceSampler, maxAge time.Duration) *resourceMonitor {
	return &resourceMonitor{
		maxAge: maxAge,
		sample: sample,
	}
}

func (monitor *resourceMonitor) run(ctx context.Context, interval time.Duration) {
	monitor.collect(time.Now())

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			monitor.collect(now)
		}
	}
}

func (monitor *resourceMonitor) collect(now time.Time) {
	if monitor.sample == nil {
		return
	}

	usage, err := monitor.sample()
	if err != nil {
		workerLogf("unable to collect resource usage: %v", err)
		return
	}
	if usage == nil {
		return
	}

	monitor.addSample(now, *usage)
}

func (monitor *resourceMonitor) addSample(at time.Time, usage resourcesUsage) {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	monitor.samples = append(monitor.samples, resourceSample{at: at, usage: usage})
	monitor.removeExpired(at)
}

func (monitor *resourceMonitor) snapshot(at time.Time) (*resourcesUsage, bool) {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	monitor.removeExpired(at)
	if len(monitor.samples) == 0 {
		return nil, false
	}

	var usage resourcesUsage
	for _, sample := range monitor.samples {
		usage.cpuCount += sample.usage.cpuCount
		usage.loadAvg += sample.usage.loadAvg
		usage.loadPercent += sample.usage.loadPercent
		usage.cpuPercent += sample.usage.cpuPercent
		usage.ramPercent += sample.usage.ramPercent
	}

	count := float64(len(monitor.samples))
	usage.cpuCount = int(float64(usage.cpuCount) / count)
	usage.loadAvg /= count
	usage.loadPercent /= count
	usage.cpuPercent /= count
	usage.ramPercent /= count

	return &usage, true
}

func (monitor *resourceMonitor) removeExpired(now time.Time) {
	cutoff := now.Add(-monitor.maxAge)
	firstValid := 0
	for firstValid < len(monitor.samples) && monitor.samples[firstValid].at.Before(cutoff) {
		firstValid++
	}
	if firstValid > 0 {
		monitor.samples = append([]resourceSample(nil), monitor.samples[firstValid:]...)
	}
}
