package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func fetchNpmPackagesByCurrentSequence(ctx context.Context) (*NpmChangesResponse, error) {
	client := &http.Client{}

	var sequence = 0

	sequenceFromDb, err := selectLastSequence(ctx)
	if err != nil {
		panic(err)
	}
	if sequenceFromDb != 0 {
		sequence = sequenceFromDb
	}

	fmt.Println("NPM current sequecne: ", sequence)

	url := npmReplicationUrl + "?since=" + strconv.Itoa(sequence) + "&limit=" + limit

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var packagesData NpmChangesResponse

	err = json.NewDecoder(response.Body).Decode(&packagesData)
	if err != nil {
		return nil, err
	}

	return &packagesData, nil
}
