package adaptiveWorker

import (
	"context"
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
)

type Controller struct {
	ctx                 context.Context
	currentCycle        int
	simultaneousWorkers int
	jobs                []job
	newCron             *cron.Cron
}

func (controller *Controller) adjust(count int) {
	var synchronization sync.Mutex

	synchronization.Lock()
	defer synchronization.Unlock()

	controller.simultaneousWorkers = count
}

func (controller *Controller) calculateWorkers() error {
	inferredMinWorkers := inferMinSimultaneousWorkers(controller.currentCycle, controller.jobs)

	resourceUsage, err := CheckResourceUsage()
	if err != nil {
		controller.adjust(inferredMinWorkers)

		return err
	}

	fmt.Println("resources usage - START----------------------->")
	fmt.Println("cpu count: ", resourceUsage.cpuCount)
	fmt.Println("cpu percent: ", resourceUsage.cpuPercent)
	fmt.Println("loadAvg: ", resourceUsage.loadAvg)
	fmt.Println("load percent: ", resourceUsage.loadPercent)
	fmt.Println("rem percent: ", resourceUsage.ramPercent)
	fmt.Println("resources usage - END-----------------------<")

	maxSimultaneousWorkers := getMaxWorkers(resourceUsage.cpuCount)

	calculatedData := calculatedResourcesUsage{
		loadPercent: resourceUsage.loadPercent,
		cpuPercent:  resourceUsage.cpuPercent,
		ramPercent:  resourceUsage.ramPercent,
	}

	canIncrease := controller.simultaneousWorkers+skipCount <= maxSimultaneousWorkers

	var simultaneousWorkers int
	if controller.simultaneousWorkers < defaultMinWorkers {
		simultaneousWorkers = inferredMinWorkers
	} else {
		simultaneousWorkers = controller.simultaneousWorkers
	}

	fmt.Println("current simultaneous gorutines: ", simultaneousWorkers)

	if getResourceUsageState(calculatedData) == usageStateExcellent || getResourceUsageState(calculatedData) == usageStateGood {
		if canIncrease {
			controller.adjust(simultaneousWorkers + skipCount)
		} else {
			controller.adjust(simultaneousWorkers)
		}

		return nil
	}

	if getResourceUsageState(calculatedData) == usageStateRegular {
		controller.adjust(simultaneousWorkers)

		return nil
	}

	if getResourceUsageState(calculatedData) == usageStateBad {
		if simultaneousWorkers-skipCount >= inferredMinWorkers {
			controller.adjust(simultaneousWorkers - skipCount)
		} else {
			controller.adjust(simultaneousWorkers)
		}

		return nil
	}

	return nil
}

func (controller *Controller) setCurrentCycle(cycle int) {
	var synchronization sync.Mutex

	synchronization.Lock()
	defer synchronization.Unlock()

	controller.currentCycle = cycle
}

func (controller *Controller) runCycle() {
	for _, currentJob := range controller.jobs {
		currentJob.calculateCycleConcurrency(
			controller.simultaneousWorkers,
			controller.currentCycle,
			inferMinSimultaneousWorkers(controller.currentCycle, controller.jobs),
		)

		currentJob.scheduledJob.Handler(currentJob.concurrency)
	}
}

func (controller *Controller) ScheduleJob(options Job) {
	var synchronization sync.Mutex

	synchronization.Lock()
	defer synchronization.Unlock()

	newJob := job{
		scheduledJob: options,
		concurrency:  0,
		canRun:       false,
	}

	controller.jobs = append(controller.jobs, newJob)
}

func (controller *Controller) start() {
	controller.newCron.AddFunc(cycleTime, func() {
		cycle, _ := getWorkerCycle(controller.ctx)
		controller.setCurrentCycle(cycle)
		controller.calculateWorkers()
		controller.runCycle()

		setWorkerCycle(controller.ctx, cycle+1)
	})

	controller.newCron.Start()
}
