package npm

import (
	"encoding/json"
	"fmt"
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

func FetchNpmPackageDownloads(payload FetchNpmPackageDownloadsPayload) (*FetchNpmPackageDownloadsResponse, error) {
	var since string
	if payload.since == nil {
		since = defaultNpmDownloadPeriod
	} else {
		since = *payload.since
	}

	url := npmDownloadUrl + "/" + since + "/" + payload.packageName

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"npm download API returned status %s",
			response.Status,
		)
	}

	var packageData FetchNpmPackageDownloadsResponse

	if err := json.NewDecoder(response.Body).Decode(&packageData); err != nil {
		return nil, err
	}

	return &packageData, nil
}
