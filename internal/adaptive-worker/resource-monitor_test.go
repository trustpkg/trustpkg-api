package adaptiveWorker

import (
	"testing"
	"time"
)

func TestResourceMonitorKeepsOnlyTheConfiguredWindow(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	monitor := newResourceMonitor(nil, resourceWindowDuration)

	monitor.addSample(now.Add(-16*time.Second), resourcesUsage{
		cpuCount:    12,
		cpuPercent:  90,
		loadPercent: 90,
		ramPercent:  90,
	})
	monitor.addSample(now.Add(-10*time.Second), resourcesUsage{
		cpuCount:    12,
		cpuPercent:  20,
		loadPercent: 20,
		ramPercent:  40,
	})

	usage, ok := monitor.snapshot(now)
	if !ok {
		t.Fatal("expected a sample in the resource window")
	}
	if usage.cpuPercent != 20 || usage.loadPercent != 20 || usage.ramPercent != 40 {
		t.Fatalf("expired sample affected snapshot: %+v", usage)
	}
}

func TestResourcePressureLevelsMoveOneStepAtATime(t *testing.T) {
	controller := &Controller{}
	bad := &resourcesUsage{cpuPercent: 90, loadPercent: 10, ramPercent: 10}
	good := &resourcesUsage{cpuPercent: 10, loadPercent: 10, ramPercent: 10}

	controller.updatePressure(bad)
	if controller.pressureLevel != 1 {
		t.Fatalf("expected first bad sample to set level 1, got %d", controller.pressureLevel)
	}
	controller.updatePressure(bad)
	if controller.pressureLevel != 2 {
		t.Fatalf("expected second bad sample to set level 2, got %d", controller.pressureLevel)
	}
	controller.updatePressure(good)
	if controller.pressureLevel != 1 {
		t.Fatalf("expected recovery to lower level to 1, got %d", controller.pressureLevel)
	}
	controller.updatePressure(good)
	if controller.pressureLevel != 0 {
		t.Fatalf("expected recovery to lower level to 0, got %d", controller.pressureLevel)
	}
}

func TestResourcePressureUsesClearThresholds(t *testing.T) {
	controller := &Controller{pressureLevel: 1}
	nearPressure := &resourcesUsage{cpuPercent: 56, loadPercent: 10, ramPercent: 10}
	clear := &resourcesUsage{cpuPercent: 54, loadPercent: 10, ramPercent: 10}

	controller.updatePressure(nearPressure)
	if controller.pressureLevel != 1 {
		t.Fatalf("expected pressure level to remain 1 near clear threshold, got %d", controller.pressureLevel)
	}
	controller.updatePressure(clear)
	if controller.pressureLevel != 0 {
		t.Fatalf("expected pressure level to clear below threshold, got %d", controller.pressureLevel)
	}
}

func TestResourceUsageStateUsesTheWorstMetric(t *testing.T) {
	if state := getResourceUsageState(resourcesUsage{cpuPercent: 10, loadPercent: 10, ramPercent: 80}); state != usageStateRegular {
		t.Fatalf("expected RAM pressure state, got %s", state)
	}
	if state := getResourceUsageState(resourcesUsage{cpuPercent: 10, loadPercent: 80, ramPercent: 10}); state != usageStateBad {
		t.Fatalf("expected critical load state, got %s", state)
	}
}
