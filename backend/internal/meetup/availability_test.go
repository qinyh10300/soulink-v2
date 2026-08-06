package meetup

import (
	"database/sql"
	"testing"
	"time"

	"soulink-backend/internal/store"
)

func TestBuildCommonSuggestionsKeepsPrivacyAndAvoidsBusyTime(t *testing.T) {
	location, err := time.LoadLocation(defaultTimezone)
	if err != nil {
		t.Fatal(err)
	}
	from := time.Date(2026, 8, 10, 0, 0, 0, 0, location)
	to := from.Add(48 * time.Hour)
	windows := []store.AvailabilityWindow{
		weeklyWindow(1, 1, 18*60, 21*60),
		weeklyWindow(2, 1, 19*60, 22*60),
		weeklyWindow(1, 2, 18*60, 21*60),
		weeklyWindow(2, 2, 18*60, 21*60),
	}
	busy := []store.BusyInterval{{
		UserID:   1,
		StartsAt: time.Date(2026, 8, 10, 19, 0, 0, 0, location),
		EndsAt:   time.Date(2026, 8, 10, 20, 0, 0, 0, location),
	}}

	result, missing := buildCommonSuggestions(
		windows,
		busy,
		[]int64{1, 2},
		from,
		to,
		60*time.Minute,
	)
	if missing != 0 {
		t.Fatalf("missing = %d, want 0", missing)
	}
	if len(result) == 0 || len(result) > 3 {
		t.Fatalf("suggestions = %d, want 1..3", len(result))
	}
	for _, suggestion := range result {
		start, parseErr := time.Parse(time.RFC3339, suggestion.StartsAt)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if start.Equal(busy[0].StartsAt) {
			t.Fatalf("suggestion overlaps private busy interval: %#v", suggestion)
		}
		if suggestion.Explanation == "" {
			t.Fatal("suggestion explanation is empty")
		}
	}
}

func TestBuildCommonSuggestionsDoesNotRevealWhoHasNoAvailability(t *testing.T) {
	location, _ := time.LoadLocation(defaultTimezone)
	from := time.Date(2026, 8, 10, 0, 0, 0, 0, location)
	result, missing := buildCommonSuggestions(
		[]store.AvailabilityWindow{weeklyWindow(1, 1, 18*60, 21*60)},
		nil,
		[]int64{1, 2},
		from,
		from.Add(24*time.Hour),
		time.Hour,
	)
	if len(result) != 0 || missing != 1 {
		t.Fatalf("result = %#v, missing = %d, want empty and 1", result, missing)
	}
}

func TestSubtractIntervalsSplitsAroundBusyBlock(t *testing.T) {
	start := time.Date(2026, 8, 10, 18, 0, 0, 0, time.UTC)
	result := subtractIntervals(
		[]interval{{Start: start, End: start.Add(4 * time.Hour)}},
		[]interval{{Start: start.Add(time.Hour), End: start.Add(2 * time.Hour)}},
	)
	if len(result) != 2 || !result[0].End.Equal(start.Add(time.Hour)) ||
		!result[1].Start.Equal(start.Add(2*time.Hour)) {
		t.Fatalf("subtractIntervals() = %#v", result)
	}
}

func weeklyWindow(userID int64, weekday, startMinute, endMinute int64) store.AvailabilityWindow {
	return store.AvailabilityWindow{
		UserID:      userID,
		WindowType:  "weekly",
		Timezone:    defaultTimezone,
		Weekday:     sql.NullInt64{Int64: weekday, Valid: true},
		StartMinute: sql.NullInt64{Int64: startMinute, Valid: true},
		EndMinute:   sql.NullInt64{Int64: endMinute, Valid: true},
	}
}
