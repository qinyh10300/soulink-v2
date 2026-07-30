package reveal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadScheduleSetBuildsInclusiveBatchTimes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reveal_schedule.json")
	contents := `{
  "timezone": "Asia/Shanghai",
  "schedules": [
    {
      "questionnaireKey": "fdu-date",
      "questionnaireVersion": "v0.1",
      "startAt": "2026-03-27T21:00:00+08:00",
      "endAt": "2026-03-27T22:00:00+08:00",
      "batchCount": 7
    }
  ]
}`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write reveal schedule file: %v", err)
	}

	set, err := LoadScheduleSet(path)
	if err != nil {
		t.Fatalf("LoadScheduleSet() error = %v", err)
	}

	window, ok := set.Window("fdu-date", "v0.1")
	if !ok {
		t.Fatal("Window() ok = false, want true")
	}
	batchTimes := window.BatchTimes()
	if len(batchTimes) != 7 {
		t.Fatalf("len(batchTimes) = %d, want 7", len(batchTimes))
	}
	for index := 1; index < len(batchTimes); index++ {
		if batchTimes[index].Sub(batchTimes[index-1]) != 10*time.Minute {
			t.Fatalf("batch interval = %s, want 10m", batchTimes[index].Sub(batchTimes[index-1]))
		}
	}
	if batchTimes[0].Format(time.RFC3339) != "2026-03-27T21:00:00+08:00" {
		t.Fatalf("first batch time = %s, want 2026-03-27T21:00:00+08:00", batchTimes[0].Format(time.RFC3339))
	}
	if batchTimes[len(batchTimes)-1].Format(time.RFC3339) != "2026-03-27T22:00:00+08:00" {
		t.Fatalf("last batch time = %s, want 2026-03-27T22:00:00+08:00", batchTimes[len(batchTimes)-1].Format(time.RFC3339))
	}
}

func TestWindowPrefersExactQuestionnaireKey(t *testing.T) {
	set := &ScheduleSet{
		Timezone: "Asia/Shanghai",
		Schedules: []Schedule{
			{
				QuestionnaireVersion: "v0.1",
				StartAt:              "2026-03-27T21:00:00+08:00",
				EndAt:                "2026-03-27T22:00:00+08:00",
				BatchCount:           7,
			},
			{
				QuestionnaireKey:     "fdu-date",
				QuestionnaireVersion: "v0.1",
				StartAt:              "2026-03-27T20:00:00+08:00",
				EndAt:                "2026-03-27T21:00:00+08:00",
				BatchCount:           3,
			},
		},
	}

	window, ok := set.Window("fdu-date", "v0.1")
	if !ok {
		t.Fatal("Window() ok = false, want true")
	}
	if window.BatchCount != 3 {
		t.Fatalf("window.BatchCount = %d, want 3", window.BatchCount)
	}
	if window.StartAt.Format(time.RFC3339) != "2026-03-27T20:00:00+08:00" {
		t.Fatalf("window.StartAt = %s, want 2026-03-27T20:00:00+08:00", window.StartAt.Format(time.RFC3339))
	}
}

func TestChoosePairedRevealAtKeepsOneOfTheExistingTimes(t *testing.T) {
	left := time.Date(2026, 3, 27, 21, 0, 0, 0, time.FixedZone("CST", 8*3600))
	right := time.Date(2026, 3, 27, 21, 10, 0, 0, time.FixedZone("CST", 8*3600))
	chosen := choosePairedRevealAt(42, 1001, 1002, &left, &right, Window{}, false)
	if chosen == nil {
		t.Fatal("choosePairedRevealAt() = nil, want one of the existing times")
	}
	if !chosen.Equal(left) && !chosen.Equal(right) {
		t.Fatalf("chosen reveal time = %s, want %s or %s", chosen.Format(time.RFC3339), left.Format(time.RFC3339), right.Format(time.RFC3339))
	}
}
