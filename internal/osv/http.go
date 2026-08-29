package osv

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	return fetchOSVEcosystemDump(context.Background(), payload)
}

func fetchOSVEcosystemDump(ctx context.Context, payload FetchOSVEcosystemDumpPayload) (io.ReadCloser, error) {
	dumpURL := fmt.Sprintf(
		"%s/%s/%s",
		osvDumpURL,
		url.PathEscape(payload.Ecosystem),
		osvDumpURLSuffix,
	)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, dumpURL, nil)
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

func fetchOSVModifiedIDCSV(ctx context.Context, ecosystem string) (io.ReadCloser, error) {
	requestURL := fmt.Sprintf("%s/%s/modified_id.csv", osvDumpURL, url.PathEscape(ecosystem))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("OSV modified feed returned status %s", response.Status)
	}

	return response.Body, nil
}

func fetchOSVVulnerability(ctx context.Context, ecosystem string, id string) (*OsvVuln, int, error) {
	requestURL := fmt.Sprintf("%s/%s/%s.json", osvDumpURL, url.PathEscape(ecosystem), url.PathEscape(id))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, 0, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil, response.StatusCode, nil
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, response.StatusCode, fmt.Errorf("OSV vulnerability returned status %s", response.Status)
	}

	var vulnerability OsvVuln
	if err := json.NewDecoder(response.Body).Decode(&vulnerability); err != nil {
		return nil, response.StatusCode, fmt.Errorf("decode OSV vulnerability %s: %w", id, err)
	}

	return &vulnerability, response.StatusCode, nil
}

func parseModifiedIDCSV(reader io.Reader, lastModifiedAt *time.Time) ([]ModifiedVulnerability, *time.Time, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = 2

	var cutoff time.Time
	if lastModifiedAt != nil {
		cutoff = lastModifiedAt.Add(-osvCursorOverlap)
	}

	changes := make([]ModifiedVulnerability, 0)
	seen := make(map[string]struct{})
	var newest *time.Time
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			return changes, newest, nil
		}
		if err != nil {
			return nil, nil, err
		}
		if record[0] == "modified" || record[0] == "timestamp" {
			continue
		}

		modifiedAt, err := time.Parse(time.RFC3339Nano, record[0])
		if err != nil {
			return nil, nil, fmt.Errorf("parse modified timestamp %q: %w", record[0], err)
		}
		if lastModifiedAt != nil && modifiedAt.Before(cutoff) {
			return changes, newest, nil
		}
		if newest == nil || modifiedAt.After(*newest) {
			copyOfTimestamp := modifiedAt
			newest = &copyOfTimestamp
		}
		if record[1] == "" {
			return nil, nil, fmt.Errorf("modified feed contains an empty vulnerability ID")
		}
		if _, exists := seen[record[1]]; exists {
			continue
		}
		seen[record[1]] = struct{}{}
		changes = append(changes, ModifiedVulnerability{ID: record[1], ModifiedAt: modifiedAt})
	}
}
