package adaptiveWorker

const (
	defaultMinWorkers = 2
	maxWorkersPerCpu  = 16

	maxWorkerHeadroom = 0.2

	cycleTime = "@every 30s"

	skipCount = 2

	badUsagePoints       = 0
	RegularUsagePoints   = 1
	GoodUsagePoints      = 2
	ExcellentUsagePoints = 3

	usageStateBad       string = "bad"
	usageStateRegular   string = "regular"
	usageStateGood      string = "good"
	usageStateExcellent string = "excellent"

	adaptiveWorkerRedisKey = "adaptive-worker"
)
