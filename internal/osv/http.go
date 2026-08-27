package osv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func fetchPackageVulnerabilities(ctx context.Context, payload fetchPackageVulnerabilitiesPayload) (*fetchPackageVulnerabilitiesResponse, error) {
	var ecosystem string
	if payload.ecosystem != nil {
		ecosystem = *payload.ecosystem
	} else {
		ecosystem = defaultOsvEcosystem
	}

	url := osvPackageVulnerabilitiesUrl
	requestPayload := OsvQuery{
		OsvPackage: OsvQueryPackage{
			Name:      payload.name,
			Ecosystem: ecosystem,
		},
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("OSV returned %s", resp.Status)
	}

	var result fetchPackageVulnerabilitiesResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func FetchOSVEcosystemDump(payload FetchOSVEcosystemDumpPayload) (io.ReadCloser, error) {
	dumpURL := fmt.Sprintf(
		"%s/%s/%s",
		npmOsvDumpUrl,
		payload.Ecosystem,
		npmOsvDumpUrlSuffix,
	)

	request, err := http.NewRequest(http.MethodGet, dumpURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		response.Body.Close()

		return nil, fmt.Errorf(
			"OSV returned status %s",
			response.Status,
		)
	}

	progress := &progressReader{
		reader:   response.Body,
		total:    response.ContentLength,
		start:    time.Now(),
		reporter: newStatusReporter("OSV download", "MB", response.ContentLength, 1024*1024),
		closer:   response.Body,
	}
	return progress, nil
}
