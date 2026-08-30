package typesense

import (
	"strconv"
	"strings"
)

func watchlistPosition(watchlist []string, name string) int {
	for index, packageName := range watchlist {
		if packageName == name {
			return index + 1
		}
	}
	return len(watchlist)
}

func packageID(value string) int64 {
	id, _ := strconv.ParseInt(value, 10, 64)
	return id
}

func normalizePipelinePhase(phase string) string {
	if phase == pipelinePhaseDatabase {
		return phase
	}
	return pipelinePhaseWatchlist
}

func isCollectionAlreadyExistsError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "already exists")
}
