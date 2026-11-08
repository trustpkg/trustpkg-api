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

func TestIsAllowedSearchHeaderAcceptsProxyIPHeaders(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	for _, header := range []string{
		"Forwarded",
		"True-Client-IP",
		"X-Client-IP",
		"X-Cluster-Client-IP",
		"X-Forwarded-For",
		"X-Forwarded-Host",
		"X-Forwarded-Port",
		"X-Forwarded-Proto",
		"X-Original-Forwarded-For",
		"X-Real-IP",
		"CF-Connecting-IP",
		"CF-Connecting-IPv6",
		"CF-Pseudo-IPv4",
		"CF-IPCountry",
		"CF-Ray",
		"CF-Visitor",
		"CDN-Loop",
	} {
		t.Run(header, func(t *testing.T) {
			if !isAllowedSearchHeader(header) {
				t.Fatalf("expected %s header to be allowed", header)
			}
		})
	}
	if isAllowedSearchHeader("X-Unrecognized-Header") {
		t.Fatal("expected unrecognized header to be rejected")
	}
}
