package packagesApi

import (
	"os"
	"strings"
)

var allowedSearchHeaders = map[string]struct{}{
	"accept":                    {},
	"accept-encoding":           {},
	"accept-language":           {},
	"cache-control":             {},
	"connection":                {},
	"content-length":            {},
	"content-type":              {},
	"cookie":                    {},
	"host":                      {},
	"origin":                    {},
	"pragma":                    {},
	"referer":                   {},
	"sec-ch-ua":                 {},
	"sec-ch-ua-mobile":          {},
	"sec-ch-ua-platform":        {},
	"sec-fetch-dest":            {},
	"sec-fetch-mode":            {},
	"sec-fetch-site":            {},
	"sec-fetch-user":            {},
	"upgrade-insecure-requests": {},
	"user-agent":                {},
	"cf-connecting-ip":          {},
	"cf-ray":                    {},
	"cf-visitor":                {},
	"cdn-loop":                  {},
	"x-forwarded-proto":         {},
}

var developmentSearchHeaders = map[string]struct{}{
	"postman-token": {},
}

func isAllowedSearchHeader(header string) bool {
	header = strings.ToLower(header)
	if _, allowed := allowedSearchHeaders[header]; allowed {
		return true
	}
	if !isDevelopmentEnvironment() {
		return false
	}
	_, allowed := developmentSearchHeaders[header]
	return allowed
}

func isDevelopmentEnvironment() bool {
	environment := strings.TrimSpace(os.Getenv("ENVIRONMENT"))
	if commentIndex := strings.Index(environment, "//"); commentIndex >= 0 {
		environment = strings.TrimSpace(environment[:commentIndex])
	}
	return strings.EqualFold(environment, "development")
}
