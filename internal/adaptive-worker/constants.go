package adaptiveWorker

import "time"

const (
	cycleTime              = "@every 1s"
	resourceSampleInterval = time.Second
	resourceWindowDuration = 15 * time.Second
	groupHandlerDelay      = 2 * time.Second
	pressureOffsetPerLevel = 15
	maxPressureLevel       = 2

	loadPressureThreshold = 60
	loadCriticalThreshold = 80
	ramPressureThreshold  = 80
	ramCriticalThreshold  = 90
	loadPressureClear     = 55
	loadCriticalClear     = 75
	ramPressureClear      = 75
	ramCriticalClear      = 85

	usageStateBad       string = "bad"
	usageStateRegular   string = "regular"
	usageStateGood      string = "good"
	usageStateExcellent string = "excellent"

	adaptiveWorkerRedisKey = "adaptive-worker"
)
