package osv

import (
	"strings"
	"testing"
	"time"
)

func TestParseModifiedIDCSVStopsAtCursorAndDeduplicates(t *testing.T) {
	lastModifiedAt := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	feed := strings.NewReader("2026-08-29T12:01:00Z,GHSA-new\n2026-08-29T12:00:00Z,GHSA-same\n2026-08-29T12:00:00Z,GHSA-same\n2026-08-29T11:59:58Z,GHSA-old\n")

	changes, newest, err := parseModifiedIDCSV(feed, &lastModifiedAt)
	if err != nil {
		t.Fatalf("parseModifiedIDCSV returned error: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("got %d changes, expected 2", len(changes))
	}
	if changes[0].ID != "GHSA-new" || changes[1].ID != "GHSA-same" {
		t.Fatalf("unexpected changes: %#v", changes)
	}
	if newest == nil || !newest.Equal(time.Date(2026, 8, 29, 12, 1, 0, 0, time.UTC)) {
		t.Fatalf("unexpected newest timestamp: %v", newest)
	}
}

func TestParseModifiedIDCSVWithoutCursor(t *testing.T) {
	feed := strings.NewReader("modified,id\n2026-08-29T12:01:00.123Z,GHSA-1\n")

	changes, newest, err := parseModifiedIDCSV(feed, nil)
	if err != nil {
		t.Fatalf("parseModifiedIDCSV returned error: %v", err)
	}
	if len(changes) != 1 || changes[0].ID != "GHSA-1" {
		t.Fatalf("unexpected changes: %#v", changes)
	}
	if newest == nil || newest.Nanosecond() != 123000000 {
		t.Fatalf("unexpected newest timestamp: %v", newest)
	}
}
