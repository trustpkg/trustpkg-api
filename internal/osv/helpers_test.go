package osv

import "testing"

func TestParseAffectedRangeBuildsIntervals(t *testing.T) {
	versions := parseAffectedRange([]OsvEvent{
		{Introduced: "1.2.0"},
		{Fixed: "1.5.0"},
		{Introduced: "2.0.0"},
		{LastAffected: "2.3.0"},
	})

	if len(versions) != 2 {
		t.Fatalf("expected 2 intervals, got %d", len(versions))
	}
	assertAffectedVersion(t, versions[0], "1.2.0", "1.5.0", "")
	assertAffectedVersion(t, versions[1], "2.0.0", "", "2.3.0")
}

func TestParseAffectedRangeKeepsOpenInterval(t *testing.T) {
	versions := parseAffectedRange([]OsvEvent{{Introduced: "0"}})

	if len(versions) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(versions))
	}
	assertAffectedVersion(t, versions[0], "0", "", "")
}

func TestParseAffectedRangeDefaultsMissingIntroducedToZero(t *testing.T) {
	versions := parseAffectedRange([]OsvEvent{{Fixed: "1.5.0"}, {LastAffected: "2.0.0"}})

	if len(versions) != 2 {
		t.Fatalf("expected 2 intervals, got %d", len(versions))
	}
	assertAffectedVersion(t, versions[0], "0", "1.5.0", "")
	assertAffectedVersion(t, versions[1], "0", "", "2.0.0")
}

func assertAffectedVersion(t *testing.T, version affectedVersion, introduced, fixed, lastAffected string) {
	t.Helper()
	if optionalValue(version.introduced) != introduced || optionalValue(version.fixed) != fixed || optionalValue(version.lastAffected) != lastAffected {
		t.Fatalf("got introduced=%q fixed=%q last_affected=%q", optionalValue(version.introduced), optionalValue(version.fixed), optionalValue(version.lastAffected))
	}
}

func optionalValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
