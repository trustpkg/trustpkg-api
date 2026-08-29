package npm

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/redis/go-redis/v9"
	"github.com/trustpkg/trustpkg-api/db"
)

type Collector struct {
	ctx            context.Context
	whiteListState *collectionState
	baseState      *baseCollectionState
}

func newCollector(ctx context.Context) *Collector {
	return &Collector{
		ctx: ctx,
	}
}

func CollectorPipeline(ctx context.Context) error {
	if redisGetCollectorCooldown(ctx) {
		npmCollectorLog.Print("Collector is in cooldown")
		return nil
	}

	collector := newCollector(ctx)

	state, err := collector.getherCollectionState()
	if err != nil {
		return err
	}
	if state.areGathered {
		collector.runBasicPipeline()
	} else {
		collector.runWhitelistPipeline()
	}

	return nil
}

func (collector *Collector) getherCollectionState() (*collectionState, error) {
	pipe := db.Redis.Pipeline()

	areGatheredCmd := redisGetWatchlistGathered(collector.ctx, pipe)
	lastPackageCmd := redisGetLastPackage(collector.ctx, pipe)
	queueCmd := redisGetPackagesQueue(collector.ctx, pipe)
	allUnscopedGatheredCmd := redisGetAllUnscopedGathered(collector.ctx, pipe)

	if _, err := pipe.Exec(collector.ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	areGathered, err := areGatheredCmd.Bool()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return nil, err
		}
		areGathered = false
	}
	if areGathered {
		return &collectionState{
			areGathered: true,
			lastPackage: nil,
			queue:       nil,
		}, nil
	}

	var lastPackage *string
	lastPackageValue, err := lastPackageCmd.Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return nil, err
		}
		lastPackageValue = ""
	}
	if lastPackageValue == "" {
		lastPackage = nil
	} else {
		lastPackage = &lastPackageValue
	}

	var queue []string
	data := queueCmd.Val()
	if data != "" {
		if err := json.Unmarshal([]byte(data), &queue); err != nil {
			npmCollectorLog.Printf("Queue unmarshal error: %v", err)
			queue = nil
		}
	}

	allUnscopedGathered, err := allUnscopedGatheredCmd.Bool()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return nil, err
		}
		allUnscopedGathered = false
	}

	collector.whiteListState = &collectionState{
		areGathered:         areGathered,
		lastPackage:         lastPackage,
		queue:               queue,
		allUnscopedGathered: allUnscopedGathered,
	}

	return &collectionState{
		areGathered:         areGathered,
		lastPackage:         lastPackage,
		queue:               queue,
		allUnscopedGathered: allUnscopedGathered,
	}, nil
}

func (collector *Collector) runWhitelistPipeline() {
	state := collector.whiteListState

	if state.allUnscopedGathered {
		if len(state.queue) == 0 {
			npmCollectorLog.Print("Scoped package queue is empty")
			return
		}

		startIndex := 0
		if state.lastPackage != nil {
			indexOfLastScoped := slices.Index(state.queue, *state.lastPackage)
			if indexOfLastScoped == len(state.queue)-1 {
				npmCollectorLog.Printf("All Watchlist packages gathered successfully, last one is: %s", *state.lastPackage)
				return
			}
			if indexOfLastScoped >= 0 {
				startIndex = indexOfLastScoped + 1
			}
		}
		currentPackage := state.queue[startIndex]

		data, err := fetchNpmPackageDownloads(fetchNpmPackageDownloadsPayload{
			packages: []string{currentPackage},
		})
		if err != nil {
			npmCollectorLog.Printf("Error fetching npm package downloads: %s", err.Error())
			return
		}

		transaction, err := db.Pool.Begin(collector.ctx)
		if err != nil {
			npmCollectorLog.Printf("Error saving data to postgres: %s", err.Error())

			return
		}

		if err := dbInsertPackagesPopularity(collector.ctx, transaction, parseDataToPayload(data)); err != nil {
			transaction.Rollback(collector.ctx)
			npmCollectorLog.Printf("Error saving data to postgres: %s", err.Error())
			return
		}
		if err := transaction.Commit(collector.ctx); err != nil {
			npmCollectorLog.Printf("Error saving data to postgres: %s", err.Error())

			return
		}

		pipe := db.Redis.Pipeline()

		redisSetLastPackage(collector.ctx, pipe, currentPackage)

		if currentPackage == state.queue[len(state.queue)-1] {
			npmCollectorLog.Printf("All Watchlist packages gathered successfully, last one is: %s", currentPackage)

			redisSetWatchlistGathered(collector.ctx, pipe)
		}

		if _, err := pipe.Exec(collector.ctx); err != nil {
			npmCollectorLog.Printf("Error executing Redis pipeline: %v", err)
		}

		npmCollectorLog.Printf("Finished processing current scoped package: %s", currentPackage)

		return
	} else {
		all := collector.getRestOfWhitelist()
		_, unscopedPackages := getDividedPackagesByScope(all)
		scopedPackages, _ := getDividedPackagesByScope(packagesWatchlist[:])
		if len(unscopedPackages) == 0 {
			npmCollectorLog.Print("No unscoped whitelist packages remain")
			return
		}

		batchSize := min(len(unscopedPackages), npmDownloadBatchLimit)
		slicedUnscopedQueue := unscopedPackages[:batchSize]

		lastPackage := slicedUnscopedQueue[len(slicedUnscopedQueue)-1]

		data, err := fetchNpmPackageDownloads(fetchNpmPackageDownloadsPayload{
			packages: slicedUnscopedQueue,
		})
		if err != nil {
			npmCollectorLog.Printf("Error fetching npm package downloads: %s", err.Error())
			return
		}

		transaction, err := db.Pool.Begin(collector.ctx)
		if err != nil {
			npmCollectorLog.Printf("Error saving data to postgres: %s", err.Error())

			return
		}

		if err := dbInsertPackagesPopularity(collector.ctx, transaction, parseDataToPayload(data)); err != nil {
			transaction.Rollback(collector.ctx)
			npmCollectorLog.Printf("Error saving data to postgres: %s", err.Error())
			return
		}
		if err := transaction.Commit(collector.ctx); err != nil {
			npmCollectorLog.Printf("Error saving data to postgres: %s", err.Error())

			return
		}

		pipe := db.Redis.Pipeline()

		redisSetPackagesQueue(collector.ctx, pipe, scopedPackages)

		if batchSize == len(unscopedPackages) {
			if len(scopedPackages) == 0 {
				npmCollectorLog.Print("No scoped whitelist packages remain")
				return
			}
			firstOfScoped := scopedPackages[0]
			index := slices.Index(packagesWatchlist[:], firstOfScoped)

			lastBeforeFirstOfScoped := packagesWatchlist[index-1]

			redisSetAllUnscopedGathered(collector.ctx, pipe)
			redisSetLastPackage(collector.ctx, pipe, lastBeforeFirstOfScoped)

			npmCollectorLog.Println("All Unscoped whitelist packages gathered successfully")
		} else {
			redisSetLastPackage(collector.ctx, pipe, lastPackage)
		}

		if _, err := pipe.Exec(collector.ctx); err != nil {
			npmCollectorLog.Printf("Error executing Redis pipeline: %v", err)
		}

		npmCollectorLog.Printf("Finished processing current batch of whitelist packages, current batch: %d, last package: %s", len(slicedUnscopedQueue), lastPackage)

		return
	}
}

func (collector *Collector) runBasicPipeline() {
	state := redisGetBasicState(collector.ctx)
	if state.shouldRunScoped {
		collector.runBasicScopedPipeline(state)
		return
	}

	packages, err := dbSelectUnscopedPackages(collector.ctx)
	if err != nil {
		npmCollectorLog.Printf("Error selecting unscoped packages: %s", err.Error())
		return
	}
	if len(packages) == 0 {
		collector.setCooldown()
		return
	}

	if !collector.saveBasicPopularity(packages) {
		return
	}
	pipe := db.Redis.Pipeline()
	redisSetBasicLastPackage(collector.ctx, pipe, packages[len(packages)-1])
	if err := redisSetBasicQueue(collector.ctx, pipe, nil); err != nil {
		npmCollectorLog.Printf("Error serializing basic package queue: %s", err.Error())
		return
	}
	redisSetShouldRunScoped(collector.ctx, pipe, true)
	if _, err := pipe.Exec(collector.ctx); err != nil {
		npmCollectorLog.Printf("Error saving basic state to Redis: %s", err.Error())
	}

	npmCollectorLog.Printf("Finished processing basic unscoped batch: %d", len(packages))
}

func (collector *Collector) runBasicScopedPipeline(state baseCollectionState) {
	queue := state.scopedQueue
	if len(queue) == 0 {
		var err error
		queue, err = dbSelectScopedPackages(collector.ctx)
		if err != nil {
			npmCollectorLog.Printf("Error selecting scoped packages: %s", err.Error())
			return
		}
		if len(queue) == 0 {
			collector.setCooldown()
			return
		}
	}

	currentPackage := queue[0]
	if !collector.saveBasicPopularity([]string{currentPackage}) {
		return
	}

	remainingQueue := queue[1:]
	pipe := db.Redis.Pipeline()
	redisSetBasicLastPackage(collector.ctx, pipe, currentPackage)
	if err := redisSetBasicQueue(collector.ctx, pipe, remainingQueue); err != nil {
		npmCollectorLog.Printf("Error serializing scoped package queue: %s", err.Error())
		return
	}
	redisSetShouldRunScoped(collector.ctx, pipe, len(remainingQueue) > 0)
	if _, err := pipe.Exec(collector.ctx); err != nil {
		npmCollectorLog.Printf("Error saving scoped state to Redis: %s", err.Error())
	}

	npmCollectorLog.Printf("Finished processing basic scoped package: %s, remaining: %d", currentPackage, len(remainingQueue))
}

func (collector *Collector) saveBasicPopularity(packages []string) bool {
	data, err := fetchNpmPackageDownloads(fetchNpmPackageDownloadsPayload{packages: packages})
	if err != nil {
		npmCollectorLog.Printf("Error fetching basic package downloads: %s", err.Error())
		return false
	}

	transaction, err := db.Pool.Begin(collector.ctx)
	if err != nil {
		npmCollectorLog.Printf("Error starting basic popularity transaction: %s", err.Error())
		return false
	}

	if err := dbInsertPackagesPopularity(collector.ctx, transaction, parseDataToPayload(data)); err != nil {
		transaction.Rollback(collector.ctx)
		npmCollectorLog.Printf("Error saving basic popularity: %s", err.Error())
		return false
	}
	if err := transaction.Commit(collector.ctx); err != nil {
		npmCollectorLog.Printf("Error committing basic popularity: %s", err.Error())
		return false
	}

	return true
}

func (collector *Collector) setCooldown() {
	pipe := db.Redis.Pipeline()
	redisSetCollectorCooldown(collector.ctx, pipe)
	if _, err := pipe.Exec(collector.ctx); err != nil {
		npmCollectorLog.Printf("Error setting collector cooldown: %s", err.Error())
		return
	}
	npmCollectorLog.Printf("No packages available, collector cooldown set for %s", cooldownOneDayDuration)

}

func (collector *Collector) getRestOfWhitelist() []string {
	state := collector.whiteListState

	if state.lastPackage == nil {
		return packagesWatchlist[:]
	}

	for index, pkg := range packagesWatchlist {
		if pkg == *state.lastPackage {
			return packagesWatchlist[index+1:]
		}
	}

	return packagesWatchlist[:]
}
