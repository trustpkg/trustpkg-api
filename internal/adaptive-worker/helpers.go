package adaptiveWorker

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

func CheckResourceUsage() (*resourcesUsage, error) {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}
	if len(cpuPercent) == 0 {
		return nil, fmt.Errorf("CPU usage measurement returned no values")
	}

	cpuCount, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	loadAvg, err := load.Avg()
	if err != nil {
		return nil, err
	}

	memory, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	if cpuCount == 0 {
		return nil, fmt.Errorf("CPU usage measurement returned zero CPUs")
	}

	loadPercent := loadAvg.Load1 / float64(cpuCount) * 100

	return &resourcesUsage{
		loadPercent: loadPercent,
		loadAvg:     loadAvg.Load1,
		cpuPercent:  cpuPercent[0],
		ramPercent:  memory.UsedPercent,
		cpuCount:    cpuCount,
	}, nil
}

func getResourceUsageState(resources resourcesUsage) string {
	if resources.cpuPercent >= loadCriticalThreshold ||
		resources.loadPercent >= loadCriticalThreshold ||
		resources.ramPercent >= ramCriticalThreshold {
		return usageStateBad
	}

	if resources.cpuPercent >= loadPressureThreshold ||
		resources.loadPercent >= loadPressureThreshold ||
		resources.ramPercent >= ramPressureThreshold {
		return usageStateRegular
	}

	return usageStateExcellent
}

func getResourcePressureLevel(resources resourcesUsage) int {
	switch getResourceUsageState(resources) {
	case usageStateBad:
		return maxPressureLevel
	case usageStateRegular:
		return 1
	default:
		return 0
	}
}

func getResourcePressureTarget(resources resourcesUsage, currentLevel int) int {
	target := getResourcePressureLevel(resources)
	if target > currentLevel {
		return target
	}

	if currentLevel >= maxPressureLevel &&
		(resources.cpuPercent >= loadCriticalClear ||
			resources.loadPercent >= loadCriticalClear ||
			resources.ramPercent >= ramCriticalClear) {
		return maxPressureLevel
	}

	if currentLevel >= 1 &&
		(resources.cpuPercent >= loadPressureClear ||
			resources.loadPercent >= loadPressureClear ||
			resources.ramPercent >= ramPressureClear) {
		return 1
	}

	return target
}
