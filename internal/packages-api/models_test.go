package packagesApi

import "testing"

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
