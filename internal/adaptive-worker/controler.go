package adaptiveWorker

type Controler struct {
	simultaneousWorkers int
}

func (controler *Controler) Adjust(count int) {
	controler.simultaneousWorkers = count
}

func (controler *Controler) CalculateWorkers() error {
	resourceUsage, err := CheckResourceUsage()
	if err != nil {
		controler.Adjust(minWorkers)

		return err
	}

	maxSimultaneousWorkers := getMaxWorkers(resourceUsage.cpuCount)
	maxSimultaneousWorkersHeadroom := getMaxWorkersHeadroom(resourceUsage.cpuCount)

	calculedData := calculedResourcesUsage{
		loadPercent: resourceUsage.loadPercent,
		cpuPercent:  resourceUsage.cpuPercent,
		ramPercent:  resourceUsage.ramPercent,
	}

	isAtBoundary := controler.simultaneousWorkers >= maxSimultaneousWorkers
	canEnterHeadroom := isAtBoundary && controler.simultaneousWorkers+skipCount <= maxSimultaneousWorkersHeadroom

	var simultaneousWorkers int
	if controler.simultaneousWorkers == 0 {
		simultaneousWorkers = minWorkers
	} else {
		simultaneousWorkers = controler.simultaneousWorkers
	}

	if getResourceUsageState(calculedData) == usageStateExcellent {
		if isAtBoundary && canEnterHeadroom {
			controler.Adjust(simultaneousWorkers + skipCount)
		} else {
			controler.Adjust(simultaneousWorkers)
		}

		return nil
	}

	if getResourceUsageState(calculedData) == usageStateGood {
		controler.Adjust(simultaneousWorkers + skipCount)

		return nil
	}

	if getResourceUsageState(calculedData) == usageStateRegular {
		controler.Adjust(simultaneousWorkers)

		return nil
	}

	if getResourceUsageState(calculedData) == usageStateBad {
		controler.Adjust(simultaneousWorkers - skipCount)

		return nil
	}

	controler.Adjust(minWorkers)

	return nil
}
