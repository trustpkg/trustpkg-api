package adaptiveWorker

import (
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

	cpuCount, err := cpu.Counts(true)

	loadAvg, err := load.Avg()
	if err != nil {
		return nil, err
	}

	memory, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
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

func getMaxWorkers(cpuCount int) int {
	return cpuCount * maxWorkersPerCpu
}

func getMaxWorkersHeadroom(cpuCount int) int {
	maxWorkersCount := getMaxWorkers(cpuCount)

	return maxWorkersCount + int(float64(maxWorkersCount)*maxWorkerHeadroom)
}

func getPointsByUsage(value float64) int {
	switch {
	case value < 40:
		return ExcellentUsagePoints
	case value < 60:
		return GoodUsagePoints
	case value < 80:
		return RegularUsagePoints
	default:
		return badUsagePoints
	}
}

func getResourceUsageState(resources calculedResourcesUsage) string {
	sum := getPointsByUsage(resources.cpuPercent) +
		getPointsByUsage(resources.loadPercent) +
		getPointsByUsage(resources.ramPercent)

	if resources.ramPercent > 90 {
		return usageStateBad
	}

	switch {
	case sum >= 8:
		return usageStateExcellent
	case sum >= 5:
		return usageStateGood
	case sum >= 2:
		return usageStateRegular
	default:
		return usageStateBad
	}
}
