package calendar

import (
	"testing"
	"time"

	"soulink-backend/internal/store"
)

func TestBusyEventViewRedactsPrivateDetails(t *testing.T) {
	now := time.Now()
	view := busyEventView(store.CalendarEvent{
		ID:                9,
		OrganizerID:       3,
		OrganizerNickname: "不应暴露",
		EventType:         "date",
		Title:             "私人约会",
		Description:       "私人备注",
		Location:          "私人地点",
		StartsAt:          now,
		EndsAt:            now.Add(time.Hour),
		Visibility:        "private",
		CreatedAt:         now,
	})

	if view.Title != "忙碌" || view.EventType != "busy" || !view.IsBusyOnly {
		t.Fatalf("busyEventView() = %#v, want redacted busy event", view)
	}
	if view.OrganizerID != 0 || view.OrganizerNickname != "" || view.Location != "" || view.Description != "" {
		t.Fatalf("busyEventView() leaked private details: %#v", view)
	}
}

func TestNormalizeEventInputRejectsPastCreation(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	_, err := normalizeEventInput(EventInput{
		EventType:  "meetup",
		Title:      "已经开始",
		StartsAt:   start,
		EndsAt:     start.Add(time.Hour),
		Visibility: "busy_only",
	}, true)
	if err == nil {
		t.Fatal("normalizeEventInput() expected past creation error")
	}
}

func TestValidateRangeRejectsOversizedWindow(t *testing.T) {
	from := time.Now()
	if err := validateRange(from, from.Add(94*24*time.Hour)); err == nil {
		t.Fatal("validateRange() expected oversized window error")
	}
}
