package adaptiveWorker

type resourcesUsage struct {
	loadPercent float64
	loadAvg     float64
	cpuPercent  float64
	ramPercent  float64
	cpuCount    int
}

type Job struct {
	ID            string
	Handler       func()
	SkippedCycles int
}
