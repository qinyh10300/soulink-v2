package meetup

import (
	"database/sql"
	"testing"
	"time"

	"soulink-backend/internal/store"
)

func TestRequireExpectedVersionRejectsStaleMutation(t *testing.T) {
	err := requireExpectedVersion(store.Meetup{Version: 4}, 3)
	if err == nil || !isExpectedConflict(err) {
		t.Fatalf("requireExpectedVersion() error = %v, want conflict", err)
	}
}

func TestReminderDoesNotExposeOtherParticipantsFeedback(t *testing.T) {
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	view := reminderFor(store.Meetup{
		Status:   "confirmed",
		StartsAt: now.Add(30 * time.Minute),
		EndsAt:   now.Add(90 * time.Minute),
	}, "accepted", false, now)
	if view.Stage != "soon" || view.Text == "" {
		t.Fatalf("reminderFor() = %#v, want soon reminder", view)
	}
}

func TestSignalFromFeedbackIsPrivateAndReversibleInput(t *testing.T) {
	wouldMeetAgain := false
	typeName, weight := signalFromFeedback(FeedbackInput{
		Attendance:     "attended",
		WouldMeetAgain: &wouldMeetAgain,
	})
	if typeName != "do_not_repeat" || weight >= 0 {
		t.Fatalf("signalFromFeedback() = %q, %d", typeName, weight)
	}
}

func TestNullableCalendarEventID(t *testing.T) {
	value := nullableInt64Pointer(sql.NullInt64{Int64: 7, Valid: true})
	if value == nil || *value != 7 {
		t.Fatalf("nullableInt64Pointer() = %#v", value)
	}
}
