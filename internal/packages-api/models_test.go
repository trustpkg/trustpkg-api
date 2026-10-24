package packagesApi

import (
	"encoding/json"
	"testing"
)

func TestSearchParamsValidateAllowsSurroundingWhitespace(t *testing.T) {
	params := SearchParams{Query: " error ", Limit: 10}

	if err := params.validate(); err != nil {
		t.Fatalf("expected surrounding whitespace to be allowed: %v", err)
	}
}

func TestListParamsValidateVulnerabilityMonths(t *testing.T) {
	for _, months := range []int{0, 1, 12} {
		params := ListParams{
			Limit: 10, Page: 1, SortBy: "name", Order: "asc", Ecosystem: "npm",
			VulnerabilityMonths: months,
		}
		if err := params.validate(); err != nil {
			t.Fatalf("expected %d vulnerability months to be valid: %v", months, err)
		}
	}

	for _, months := range []int{-1, 13} {
		params := ListParams{
			Limit: 10, Page: 1, SortBy: "name", Order: "asc", Ecosystem: "npm",
			VulnerabilityMonths: months,
		}
		if err := params.validate(); err == nil {
			t.Fatalf("expected %d vulnerability months to be invalid", months)
		}
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		totalPackages int64
		limit         int
		expected      int
	}{
		{totalPackages: 0, limit: 10, expected: 0},
		{totalPackages: 1, limit: 10, expected: 1},
		{totalPackages: 10, limit: 10, expected: 1},
		{totalPackages: 11, limit: 10, expected: 2},
	}

	for _, test := range tests {
		if actual := calculateTotalPages(test.totalPackages, test.limit); actual != test.expected {
			t.Fatalf("expected %d pages for %d packages with limit %d, got %d", test.expected, test.totalPackages, test.limit, actual)
		}
	}
}

func TestPackageDetailParamsValidate(t *testing.T) {
	bad := []string{"", "a b", "../x", "a'; drop", "@a/b/c"}
	for _, id := range bad {
		p := PackageDetailParams{Identifier: id, Ecosystem: "npm"}
		if p.validate() == nil {
			t.Fatalf("%q should be invalid", id)
		}
	}
	for _, id := range []string{"lodash", "@types/node", "lodash.get"} {
		p := PackageDetailParams{Identifier: id, Ecosystem: "npm"}
		if err := p.validate(); err != nil {
			t.Fatalf("%q: %v", id, err)
		}
	}
}

func TestPackageDetailRequestDecoding(t *testing.T) {
	var request PackageDetailRequest
	body := `{"includeVulnerabilities":true,"vulnerabilityMonths":"all","severity":"high"}`
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatal(err)
	}
	params := request.params("lodash")
	if !params.IncludeVulnerabilities || params.VulnerabilityMonths != 0 || params.Severity != "HIGH" || params.Ecosystem != "npm" {
		t.Fatalf("unexpected params: %+v", params)
	}

	for _, invalid := range []string{`{"vulnerabilityMonths":13}`, `{"vulnerabilityMonths":"x"}`, `{"vulnerabilityMonths":-1}`} {
		if json.Unmarshal([]byte(invalid), &PackageDetailRequest{}) == nil {
			t.Fatalf("%s should be invalid", invalid)
		}
	}
}
