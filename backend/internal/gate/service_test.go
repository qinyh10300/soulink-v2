package gate

import (
	"testing"
	"time"
)

func TestIsWindowActive(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	cases := []struct {
		name string
		now  time.Time
		want bool
	}{
		{name: "before", now: start.Add(-time.Second), want: false},
		{name: "at_start", now: start, want: true},
		{name: "inside", now: start.Add(time.Hour), want: true},
		{name: "at_end", now: end, want: false},
		{name: "after", now: end.Add(time.Second), want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isWindowActive(start, end, tc.now); got != tc.want {
				t.Fatalf("isWindowActive() = %v, want %v", got, tc.want)
			}
		})
	}
}
