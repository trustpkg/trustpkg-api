package adaptiveWorker

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Controller struct {
	ctx            context.Context
	mu             sync.RWMutex
	currentCycle   int
	pressureLevel  int
	jobs           []Job
	activeGroups   map[string]bool
	groupWaitGroup sync.WaitGroup
	newCron        *cron.Cron
	monitor        *resourceMonitor
}

type jobGroup struct {
	key  string
	jobs []Job
}

func (controller *Controller) ScheduleJob(options Job) {
	if options.Handler == nil {
		return
	}

	controller.mu.Lock()
	defer controller.mu.Unlock()

	controller.jobs = append(controller.jobs, options)
}

func (controller *Controller) pressureOffset() int {
	controller.mu.RLock()
	defer controller.mu.RUnlock()

	return controller.pressureLevel * pressureOffsetPerLevel
}

func (controller *Controller) updatePressure(usage *resourcesUsage) {
	if usage == nil {
		return
	}

	controller.mu.Lock()
	target := getResourcePressureTarget(*usage, controller.pressureLevel)
	if controller.pressureLevel < target {
		controller.pressureLevel++
	} else if controller.pressureLevel > target {
		controller.pressureLevel--
	}
	level := controller.pressureLevel
	controller.mu.Unlock()

	workerLogf("resources usage: cpu=%.2f%% load=%.2f%% ram=%.2f%% state=%s pressure level=%d skipped cycles offset=%d",
		usage.cpuPercent,
		usage.loadPercent,
		usage.ramPercent,
		getResourceUsageState(*usage),
		level,
		level*pressureOffsetPerLevel,
	)
}

func (controller *Controller) groupsForCycle(currentCycle int) []jobGroup {
	controller.mu.RLock()
	jobs := append([]Job(nil), controller.jobs...)
	controller.mu.RUnlock()

	groups := make([]jobGroup, 0, len(jobs))
	groupIndexes := make(map[string]int, len(jobs))
	offset := controller.pressureOffset()

	for index, scheduledJob := range jobs {
		if !shouldRunInCycle(scheduledJob, currentCycle, offset) {
			continue
		}

		key := scheduledJob.ID
		if key == "" {
			key = "job:" + strconv.Itoa(index)
		} else {
			key = "id:" + key
		}

		groupIndex, exists := groupIndexes[key]
		if !exists {
			groupIndexes[key] = len(groups)
			groups = append(groups, jobGroup{key: key, jobs: []Job{scheduledJob}})
			continue
		}

		groups[groupIndex].jobs = append(groups[groupIndex].jobs, scheduledJob)
	}

	return groups
}

func (controller *Controller) runCycle(currentCycle int) int {
	groups := controller.groupsForCycle(currentCycle)
	startedGroups := 0

	for _, group := range groups {
		controller.mu.Lock()
		if controller.activeGroups == nil {
			controller.activeGroups = make(map[string]bool)
		}
		if controller.activeGroups[group.key] {
			controller.mu.Unlock()
			continue
		}
		controller.activeGroups[group.key] = true
		controller.groupWaitGroup.Add(1)
		controller.mu.Unlock()

		startedGroups++
		go controller.runGroup(group)
	}

	return startedGroups
}

func (controller *Controller) runGroup(group jobGroup) {
	defer controller.groupWaitGroup.Done()
	defer func() {
		controller.mu.Lock()
		delete(controller.activeGroups, group.key)
		controller.mu.Unlock()
	}()

	runJobGroup(controller.ctx, group.jobs, groupHandlerDelay)
}

func runJobGroup(ctx context.Context, jobs []Job, delay time.Duration) {
	var lastStart time.Time
	for index, scheduledJob := range jobs {
		if index > 0 {
			wait := delay - time.Since(lastStart)
			if wait > 0 && !waitForContext(ctx, wait) {
				return
			}
		}

		if ctx.Err() != nil {
			return
		}
		lastStart = time.Now()
		scheduledJob.Handler()
	}
}

func (controller *Controller) runScheduledCycle() {
	controller.mu.RLock()
	localCycle := controller.currentCycle
	controller.mu.RUnlock()

	cycle, err := getWorkerCycle(controller.ctx)
	if err != nil {
		workerLogf("unable to read adaptive worker cycle: %v; using local cycle %d", err, localCycle)
		cycle = localCycle
	}

	if controller.monitor != nil {
		if usage, ok := controller.monitor.snapshot(time.Now()); ok {
			controller.updatePressure(usage)
		}
	}

	startedGroups := controller.runCycle(cycle)
	nextCycle := cycle + 1

	controller.mu.Lock()
	controller.currentCycle = nextCycle
	controller.mu.Unlock()

	if err := setWorkerCycle(controller.ctx, nextCycle); err != nil {
		workerLogf("unable to persist adaptive worker cycle: %v", err)
	}

	workerLogf("scheduled cycle=%d groups started=%d", nextCycle, startedGroups)
}

func (controller *Controller) start() {
	if _, err := controller.newCron.AddFunc(cycleTime, controller.runScheduledCycle); err != nil {
		workerLogf("unable to schedule adaptive worker cycle: %v", err)
		return
	}

	controller.newCron.Start()
	go controller.monitor.run(controller.ctx, resourceSampleInterval)
	go func() {
		<-controller.ctx.Done()
		<-controller.newCron.Stop().Done()
		controller.groupWaitGroup.Wait()
	}()
}

func waitForContext(ctx context.Context, duration time.Duration) bool {
	if duration <= 0 {
		return true
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
