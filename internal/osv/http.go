package osv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchPackageVulnerabilities(payload FetchPackageVulnerabilitiesPayload) (*FetchPackageVulnerabilitiesResponse, error) {
	var ecosystem string
	if payload.Ecosystem != nil {
		ecosystem = *payload.Ecosystem
	} else {
		ecosystem = defaultOsvEcosystem
	}

	url := osvPackageVulnerabilitiesUrl
	requestPayload := OsvQuery{
		OsvPackage: OsvQueryPackage{
			Name:      payload.Name,
			Ecosystem: ecosystem,
		},
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
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

	var result FetchPackageVulnerabilitiesResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}
