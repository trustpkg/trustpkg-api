package adaptiveWorker

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestShouldRunInCycleUsesPressureOffset(t *testing.T) {
	job := Job{SkippedCycles: 29}

	if !shouldRunInCycle(job, 29, 0) {
		t.Fatal("job should run on the base schedule")
	}
	if shouldRunInCycle(job, 29, pressureOffsetPerLevel) {
		t.Fatal("job should be delayed by the pressure offset")
	}
	if !shouldRunInCycle(job, 44, pressureOffsetPerLevel) {
		t.Fatal("job should run on the delayed schedule")
	}
}

func TestGroupsForCycleSeparatesEmptyIDsAndGroupsMatchingIDs(t *testing.T) {
	controller := &Controller{
		jobs: []Job{
			{ID: "shared", Handler: func() {}},
			{ID: "shared", Handler: func() {}},
			{Handler: func() {}},
			{Handler: func() {}},
		},
	}

	groups := controller.groupsForCycle(0)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if len(groups[0].jobs) != 2 {
		t.Fatalf("expected shared group to contain 2 jobs, got %d", len(groups[0].jobs))
	}
	if len(groups[1].jobs) != 1 || len(groups[2].jobs) != 1 {
		t.Fatal("jobs without IDs must remain independent")
	}
}

func TestRunJobGroupRunsHandlersInOrder(t *testing.T) {
	var mu sync.Mutex
	var calls []string
	ctx := context.Background()

	runJobGroup(ctx, []Job{
		{Handler: func() {
			mu.Lock()
			calls = append(calls, "first")
			mu.Unlock()
		}},
		{Handler: func() {
			mu.Lock()
			calls = append(calls, "second")
			mu.Unlock()
		}},
	}, 0)

	if !reflect.DeepEqual(calls, []string{"first", "second"}) {
		t.Fatalf("expected handlers in order, got %v", calls)
	}
}

func TestRunJobGroupHonorsDelayBetweenHandlers(t *testing.T) {
	starts := make(chan time.Time, 2)
	ctx := context.Background()

	runJobGroup(ctx, []Job{
		{Handler: func() { starts <- time.Now() }},
		{Handler: func() { starts <- time.Now() }},
	}, 10*time.Millisecond)

	first := <-starts
	second := <-starts
	if second.Sub(first) < 10*time.Millisecond {
		t.Fatalf("expected at least 10ms between handler starts, got %s", second.Sub(first))
	}
}

func TestRunCycleDoesNotOverlapAnActiveGroup(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	controller := &Controller{
		ctx:          context.Background(),
		jobs:         []Job{{ID: "shared", Handler: func() { close(started); <-release }}},
		activeGroups: make(map[string]bool),
	}

	if groups := controller.runCycle(0); groups != 1 {
		t.Fatalf("expected one started group, got %d", groups)
	}
	<-started

	if groups := controller.runCycle(1); groups != 0 {
		t.Fatalf("expected active group to be skipped, got %d started groups", groups)
	}

	close(release)
	controller.groupWaitGroup.Wait()
}

func TestRunCycleStartsDifferentGroupsIndependently(t *testing.T) {
	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	release := make(chan struct{})
	controller := &Controller{
		ctx: context.Background(),
		jobs: []Job{
			{ID: "first", Handler: func() { close(firstStarted); <-release }},
			{ID: "second", Handler: func() { close(secondStarted); <-release }},
		},
		activeGroups: make(map[string]bool),
	}

	if groups := controller.runCycle(0); groups != 2 {
		t.Fatalf("expected two started groups, got %d", groups)
	}
	<-firstStarted
	<-secondStarted

	close(release)
	controller.groupWaitGroup.Wait()
}
