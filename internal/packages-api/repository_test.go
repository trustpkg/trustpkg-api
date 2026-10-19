package packagesApi

import (
	"strings"
	"testing"
)

func TestBuildPackageDetailQueryBindsEveryArgument(t *testing.T) {
	params := PackageDetailParams{
		Identifier: "lodash", Ecosystem: "npm", IncludeVulnerabilities: true, WithSummary: true,
		WithCurrentStatus: true, IncludeReferences: true, Severity: "HIGH", VulnerabilityMonths: 3,
	}
	query, args := buildPackageDetailQuery(params)
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d", len(args))
	}
	for i := 1; i <= len(args); i++ {
		if !strings.Contains(query, "$"+string(rune('0'+i))) {
			t.Fatalf("placeholder $%d missing", i)
		}
	}
	if strings.Contains(query, "/*") {
		t.Fatal("unreplaced template marker")
	}
}
