package packagesApi

import "testing"

func TestIsAllowedSearchHeaderPostmanTokenDevelopmentOnly(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")
	if !isAllowedSearchHeader("Postman-Token") {
		t.Fatal("expected Postman-Token to be allowed in development")
	}

	t.Setenv("ENVIRONMENT", "production")
	if isAllowedSearchHeader("Postman-Token") {
		t.Fatal("expected Postman-Token to be rejected outside development")
	}
}

func TestIsAllowedSearchHeaderAcceptsStandardHeaders(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	if !isAllowedSearchHeader("Accept") {
		t.Fatal("expected standard Accept header to be allowed")
	}
}
