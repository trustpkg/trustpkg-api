package npm

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func fetchNpmPackages(since int) (*NpmChangesResponse, error) {
	client := &http.Client{}

	url := npmReplicationUrl + "?since=" + strconv.Itoa(since) + "&limit=" + strconv.Itoa(limit)

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
