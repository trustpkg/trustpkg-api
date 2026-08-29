package osv

import (
	"strconv"
	"strings"
	"time"
)

type ParsedVulnerability struct {
	OSVID       string
	CVEID       *string
	Summary     *string
	Description *string
	Severity    *string
	CVSSScore   *float64
	CVSSVector  *string
	PublishedAt *time.Time
	ModifiedAt  *time.Time
	WithdrawnAt *time.Time

	Affected   []AffectedPackage
	References []string
}

type AffectedPackage struct {
	Ecosystem           string  `json:"ecosystem"`
	PackageName         string  `json:"package_name"`
	IntroducedVersion   *string `json:"introduced_version,omitempty"`
	FixedVersion        *string `json:"fixed_version,omitempty"`
	LastAffectedVersion *string `json:"last_affected_version,omitempty"`
}

func parseVulnerability(raw OsvVuln) ParsedVulnerability {
	parsed := ParsedVulnerability{OSVID: raw.ID, Summary: optionalString(raw.Summary), Description: optionalString(raw.Details), PublishedAt: optionalTime(raw.Published), ModifiedAt: optionalTime(raw.Modified), WithdrawnAt: raw.Withdrawn}
	for _, alias := range raw.Aliases {
		if strings.HasPrefix(alias, "CVE-") {
			parsed.CVEID = optionalString(alias)
			break
		}
	}
	for _, severity := range raw.Severity {
		if parsed.CVSSVector == nil && severity.Score != "" {
			parsed.CVSSVector = optionalString(severity.Score)
		}
		if score, err := strconv.ParseFloat(severity.Score, 64); err == nil {
			parsed.CVSSScore = &score
		}
	}
	if raw.DatabaseSpecific.Severity != "" {
		parsed.Severity = optionalString(raw.DatabaseSpecific.Severity)
	}
	for _, affected := range raw.Affected {
		for _, versionRange := range affected.Ranges {
			for _, version := range parseAffectedRange(versionRange.Events) {
				parsed.Affected = append(parsed.Affected, AffectedPackage{
					Ecosystem: affected.Package.Ecosystem, PackageName: affected.Package.Name,
					IntroducedVersion:   version.introduced,
					FixedVersion:        version.fixed,
					LastAffectedVersion: version.lastAffected,
				})
			}
		}
	}
	for _, reference := range raw.References {
		if reference.URL != "" {
			parsed.References = append(parsed.References, reference.URL)
		}
	}
	return parsed
}

type affectedVersion struct {
	introduced   *string
	fixed        *string
	lastAffected *string
}

func parseAffectedRange(events []OsvEvent) []affectedVersion {
	versions := make([]affectedVersion, 0, len(events))
	var introduced *string

	for _, event := range events {
		if event.Introduced != "" {
			if introduced != nil {
				versions = append(versions, affectedVersion{introduced: introduced})
			}
			introduced = optionalString(event.Introduced)
		}

		if event.Fixed != "" {
			if introduced == nil {
				introduced = optionalString("0")
			}
			versions = append(versions, affectedVersion{
				introduced: introduced,
				fixed:      optionalString(event.Fixed),
			})
			introduced = nil
		}

		if event.LastAffected != "" {
			if introduced == nil {
				introduced = optionalString("0")
			}
			versions = append(versions, affectedVersion{
				introduced:   introduced,
				lastAffected: optionalString(event.LastAffected),
			})
			introduced = nil
		}
	}

	if introduced != nil {
		versions = append(versions, affectedVersion{introduced: introduced})
	}
	return versions
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func optionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
