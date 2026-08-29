package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func fetchNpmPackages(since int) (*NpmChangesResponse, error) {
	client := &http.Client{}

	url := npmReplicationUrl + "?since=" + strconv.Itoa(since) + "&limit=" + strconv.Itoa(limit)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if err := WaitForNpmRequest(context.Background()); err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("npm replication API returned status %s", response.Status)
	}

	var packagesData NpmChangesResponse

	err = json.NewDecoder(response.Body).Decode(&packagesData)
	if err != nil {
		return nil, err
	}

	return &packagesData, nil
}

func fetchNpmPackageDownloads(payload fetchNpmPackageDownloadsPayload) (fetchNpmPackageDownloadsReturnType, error) {
	if len(payload.packages) > npmDownloadBatchLimit {
		return nil, fmt.Errorf("too many packages in a single request: maximum is %d", npmDownloadBatchLimit)
	}

	var since string
	if payload.since == nil {
		since = defaultNpmDownloadPeriod
	} else {
		since = *payload.since
	}

	url := npmDownloadUrl + "/" + since + "/" + strings.Join(payload.packages, ",")

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	if err := WaitForNpmRequest(context.Background()); err != nil {
		return nil, err
	}

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

	var responseData json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&responseData); err != nil {
		return nil, err
	}

	return parsePackageDownloadsResponse(responseData, payload.packages)
}

func parsePackageDownloadsResponse(data json.RawMessage, packages []string) (fetchNpmPackageDownloadsReturnType, error) {
	if len(packages) == 1 {
		var packageData FetchNpmPackageDownloadsResponse
		if err := json.Unmarshal(data, &packageData); err != nil {
			return nil, err
		}

		return fetchNpmPackageDownloadsReturnType{packages[0]: &packageData}, nil
	}

	var packageData fetchNpmPackageDownloadsReturnType
	if err := json.Unmarshal(data, &packageData); err != nil {
		return nil, err
	}

	return packageData, nil
}
