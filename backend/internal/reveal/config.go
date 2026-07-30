package reveal

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type ScheduleSet struct {
	Timezone  string     `json:"timezone"`
	Schedules []Schedule `json:"schedules"`
}

type Schedule struct {
	QuestionnaireKey     string `json:"questionnaireKey,omitempty"`
	QuestionnaireVersion string `json:"questionnaireVersion"`
	StartAt              string `json:"startAt"`
	EndAt                string `json:"endAt"`
	BatchCount           int    `json:"batchCount"`
}

type Window struct {
	QuestionnaireKey     string
	QuestionnaireVersion string
	Timezone             string
	StartAt              time.Time
	EndAt                time.Time
	BatchCount           int
}

func LoadScheduleSet(path string) (*ScheduleSet, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return &ScheduleSet{Timezone: "Asia/Shanghai"}, nil
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read reveal schedule %s: %w", path, err)
	}

	var set ScheduleSet
	if err := json.Unmarshal(contents, &set); err != nil {
		return nil, fmt.Errorf("parse reveal schedule %s: %w", path, err)
	}
	if strings.TrimSpace(set.Timezone) == "" {
		set.Timezone = "Asia/Shanghai"
	}

	windows := make([]Window, 0, len(set.Schedules))
	for index, item := range set.Schedules {
		window, err := buildWindow(set.Timezone, item)
		if err != nil {
			return nil, fmt.Errorf("parse reveal schedule item %d: %w", index+1, err)
		}
		windows = append(windows, window)
	}

	sort.SliceStable(windows, func(i, j int) bool {
		if windows[i].QuestionnaireVersion == windows[j].QuestionnaireVersion {
			if windows[i].QuestionnaireKey == windows[j].QuestionnaireKey {
				return windows[i].StartAt.Before(windows[j].StartAt)
			}
			return windows[i].QuestionnaireKey < windows[j].QuestionnaireKey
		}
		return windows[i].QuestionnaireVersion < windows[j].QuestionnaireVersion
	})

	normalized := make([]Schedule, 0, len(windows))
	for _, window := range windows {
		normalized = append(normalized, Schedule{
			QuestionnaireKey:     window.QuestionnaireKey,
			QuestionnaireVersion: window.QuestionnaireVersion,
			StartAt:              window.StartAt.Format(time.RFC3339),
			EndAt:                window.EndAt.Format(time.RFC3339),
			BatchCount:           window.BatchCount,
		})
	}
	set.Schedules = normalized
	return &set, nil
}

func (s *ScheduleSet) Window(questionnaireKey, questionnaireVersion string) (Window, bool) {
	if s == nil {
		return Window{}, false
	}

	questionnaireKey = strings.TrimSpace(questionnaireKey)
	questionnaireVersion = strings.TrimSpace(questionnaireVersion)
	if questionnaireVersion == "" {
		return Window{}, false
	}

	var versionOnly *Window
	for _, item := range s.Schedules {
		window, err := buildWindow(s.Timezone, item)
		if err != nil {
			continue
		}
		if window.QuestionnaireVersion != questionnaireVersion {
			continue
		}
		if window.QuestionnaireKey == questionnaireKey && window.QuestionnaireKey != "" {
			return window, true
		}
		if window.QuestionnaireKey == "" && versionOnly == nil {
			copy := window
			versionOnly = &copy
		}
	}
	if versionOnly != nil {
		return *versionOnly, true
	}
	return Window{}, false
}

func (w Window) BatchTimes() []time.Time {
	if w.BatchCount <= 0 || w.EndAt.Before(w.StartAt) {
		return nil
	}
	if w.BatchCount == 1 || w.EndAt.Equal(w.StartAt) {
		return []time.Time{w.StartAt}
	}

	step := w.EndAt.Sub(w.StartAt) / time.Duration(w.BatchCount-1)
	times := make([]time.Time, 0, w.BatchCount)
	for index := 0; index < w.BatchCount; index++ {
		if index == w.BatchCount-1 {
			times = append(times, w.EndAt)
			continue
		}
		times = append(times, w.StartAt.Add(step*time.Duration(index)))
	}
	return times
}

func buildWindow(defaultTimezone string, item Schedule) (Window, error) {
	questionnaireVersion := strings.TrimSpace(item.QuestionnaireVersion)
	if questionnaireVersion == "" {
		return Window{}, fmt.Errorf("questionnaireVersion is required")
	}
	if item.BatchCount <= 0 {
		return Window{}, fmt.Errorf("batchCount must be greater than zero")
	}

	timezone := strings.TrimSpace(defaultTimezone)
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return Window{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}

	startAt, err := parseConfiguredTime(item.StartAt, location)
	if err != nil {
		return Window{}, fmt.Errorf("parse startAt: %w", err)
	}
	endAt, err := parseConfiguredTime(item.EndAt, location)
	if err != nil {
		return Window{}, fmt.Errorf("parse endAt: %w", err)
	}
	if endAt.Before(startAt) {
		return Window{}, fmt.Errorf("endAt must not be earlier than startAt")
	}

	return Window{
		QuestionnaireKey:     strings.TrimSpace(item.QuestionnaireKey),
		QuestionnaireVersion: questionnaireVersion,
		Timezone:             timezone,
		StartAt:              startAt,
		EndAt:                endAt,
		BatchCount:           item.BatchCount,
	}, nil
}

func parseConfiguredTime(value string, location *time.Location) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("time value is required")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
	}
	for _, layout := range layouts {
		if layout == time.RFC3339 {
			parsed, err := time.Parse(layout, value)
			if err == nil {
				return parsed, nil
			}
			continue
		}
		parsed, err := time.ParseInLocation(layout, value, location)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format %q", value)
}
