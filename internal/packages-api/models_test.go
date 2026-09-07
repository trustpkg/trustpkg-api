package packagesApi

import "testing"

func TestSearchParamsValidateAllowsSurroundingWhitespace(t *testing.T) {
	params := SearchParams{Query: " error ", Limit: 10}

	if err := params.validate(); err != nil {
		t.Fatalf("expected surrounding whitespace to be allowed: %v", err)
	}
}
