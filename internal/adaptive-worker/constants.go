package adaptiveWorker

const (
	minWorkers       = 2
	maxWorkersPerCpu = 16

	maxWorkerHeadroom = 0.2

	maxCpuUsage    = 0.6
	maxMemoryUsage = 0.8

	timeFromRequestWorker = 30

	skipCount = 2

	badUsagePoints       = 0
	RegularUsagePoints   = 1
	GoodUsagePoints      = 2
	ExcellentUsagePoints = 3

	usageStateBad       string = "bad"
	usageStateRegular   string = "regular"
	usageStateGood      string = "good"
	usageStateExcellent string = "excellent"
)
