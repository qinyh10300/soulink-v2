package playground

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPostCursorRoundTrip(t *testing.T) {
	scope := postCursorScope("study", "期末", "hot")
	want := encodedPostCursor{
		Version:   1,
		Scope:     scope,
		Score:     17,
		CreatedAt: "2026-08-07T12:30:45.123456Z",
		ID:        42,
	}
	raw, err := encodePostCursor(want)
	if err != nil {
		t.Fatalf("encodePostCursor() error = %v", err)
	}
	got, err := decodePostCursor(raw, scope)
	if err != nil {
		t.Fatalf("decodePostCursor() error = %v", err)
	}
	if got.Score != want.Score || got.ID != want.ID {
		t.Fatalf("decoded cursor = %+v", got)
	}
	wantTime, _ := time.Parse(time.RFC3339Nano, want.CreatedAt)
	if !got.CreatedAt.Equal(wantTime) {
		t.Fatalf("decoded time = %s, want %s", got.CreatedAt, wantTime)
	}
}

func TestInternalMediaKey(t *testing.T) {
	key := uuid.NewString()
	if got, ok := internalMediaKey("/media/" + key); !ok || got != key {
		t.Fatalf("internalMediaKey() = %q, %v", got, ok)
	}
	for _, value := range []string{
		"https://example.com/image.png",
		"/media/not-a-uuid",
		"/media/" + key + "/extra",
	} {
		if _, ok := internalMediaKey(value); ok {
			t.Fatalf("internalMediaKey(%q) accepted invalid path", value)
		}
	}
}

func TestPostCursorRejectsChangedFilterAndMalformedValues(t *testing.T) {
	scope := postCursorScope("campus", "", "latest")
	raw, err := encodePostCursor(encodedPostCursor{
		Version:   1,
		Scope:     scope,
		CreatedAt: "2026-08-07T12:30:45Z",
		ID:        9,
	})
	if err != nil {
		t.Fatalf("encodePostCursor() error = %v", err)
	}
	if _, err := decodePostCursor(raw, postCursorScope("study", "", "latest")); err == nil {
		t.Fatal("decodePostCursor() accepted a cursor from another filter")
	}
	if _, err := decodePostCursor("not-base64", scope); err == nil {
		t.Fatal("decodePostCursor() accepted malformed data")
	}
}
