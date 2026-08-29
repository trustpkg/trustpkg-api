package npm

import (
	"encoding/json"
	"testing"
)

func TestParsePackageDownloadsResponseSinglePackage(t *testing.T) {
	data := json.RawMessage(`{"downloads":2931290995,"start":"2026-08-01","end":"2026-08-29","package":"tslib"}`)

	result, err := parsePackageDownloadsResponse(data, []string{"tslib"})
	if err != nil {
		t.Fatal(err)
	}
	if result["tslib"].Downloads != 2931290995 {
		t.Fatalf("expected 2931290995 downloads, got %d", result["tslib"].Downloads)
	}
}

func TestParsePackageDownloadsResponseBatch(t *testing.T) {
	data := json.RawMessage(`{"tslib":{"downloads":123},"typescript":{"downloads":456}}`)

	result, err := parsePackageDownloadsResponse(data, []string{"tslib", "typescript"})
	if err != nil {
		t.Fatal(err)
	}
	if result["tslib"].Downloads != 123 || result["typescript"].Downloads != 456 {
		t.Fatalf("unexpected batch response: %+v", result)
	}
}
