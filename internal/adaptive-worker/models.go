package adaptiveWorker

type resourcesUsage struct {
	loadPercent float64
	loadAvg     float64
	cpuPercent  float64
	ramPercent  float64
	cpuCount    int
}

type calculatedResourcesUsage struct {
	loadPercent float64
	cpuPercent  float64
	ramPercent  float64
}

type jobHandler func(concurrency int)

type Job struct {
	Handler          jobHandler
	ReservationRatio float64
	SkippedCycles    int
}
