package meetup

import (
	"fmt"
	"sort"
	"time"

	"soulink-backend/internal/store"
)

const defaultTimezone = "Asia/Shanghai"

func buildCommonSuggestions(
	windows []store.AvailabilityWindow,
	busy []store.BusyInterval,
	participantIDs []int64,
	from time.Time,
	to time.Time,
	duration time.Duration,
) ([]TimeSuggestion, int) {
	byUser := make(map[int64][]interval, len(participantIDs))
	for _, window := range windows {
		byUser[window.UserID] = append(
			byUser[window.UserID],
			materializeWindow(window, from, to)...,
		)
	}

	missing := 0
	var common []interval
	for index, userID := range participantIDs {
		available := mergeIntervals(byUser[userID])
		if len(available) == 0 {
			missing++
		}
		if index == 0 {
			common = available
		} else {
			common = intersectIntervals(common, available)
		}
		if len(common) == 0 && index > 0 {
			break
		}
	}
	if missing > 0 || len(common) == 0 {
		return []TimeSuggestion{}, missing
	}

	busyIntervals := make([]interval, 0, len(busy))
	for _, item := range busy {
		busyIntervals = append(busyIntervals, interval{Start: item.StartsAt, End: item.EndsAt})
	}
	common = subtractIntervals(common, mergeIntervals(busyIntervals))
	return chooseSuggestions(common, duration, len(participantIDs)), missing
}

func materializeWindow(window store.AvailabilityWindow, from, to time.Time) []interval {
	if window.WindowType == "one_time" && window.StartsAt.Valid && window.EndsAt.Valid {
		candidate, ok := clippedInterval(
			interval{Start: window.StartsAt.Time, End: window.EndsAt.Time},
			from,
			to,
		)
		if !ok {
			return nil
		}
		return []interval{candidate}
	}
	if window.WindowType != "weekly" || !window.Weekday.Valid ||
		!window.StartMinute.Valid || !window.EndMinute.Valid {
		return nil
	}

	timezone := window.Timezone
	if timezone == "" {
		timezone = defaultTimezone
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil
	}
	localFrom := from.In(location)
	localTo := to.In(location)
	day := time.Date(localFrom.Year(), localFrom.Month(), localFrom.Day(), 0, 0, 0, 0, location)
	lastDay := time.Date(localTo.Year(), localTo.Month(), localTo.Day(), 0, 0, 0, 0, location)
	result := make([]interval, 0, 8)
	for !day.After(lastDay) {
		weekday := int(day.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		if weekday == int(window.Weekday.Int64) && withinValidity(day, window, location) {
			start := day.Add(time.Duration(window.StartMinute.Int64) * time.Minute)
			end := day.Add(time.Duration(window.EndMinute.Int64) * time.Minute)
			candidate, ok := clippedInterval(interval{Start: start, End: end}, from, to)
			if ok {
				result = append(result, candidate)
			}
		}
		day = day.AddDate(0, 0, 1)
	}
	return result
}

func withinValidity(day time.Time, window store.AvailabilityWindow, location *time.Location) bool {
	dayKey := day.Format("2006-01-02")
	if window.ValidFrom.Valid && dayKey < window.ValidFrom.Time.In(location).Format("2006-01-02") {
		return false
	}
	if window.ValidUntil.Valid && dayKey > window.ValidUntil.Time.In(location).Format("2006-01-02") {
		return false
	}
	return true
}

func clippedInterval(candidate interval, from, to time.Time) (interval, bool) {
	if candidate.Start.Before(from) {
		candidate.Start = from
	}
	if candidate.End.After(to) {
		candidate.End = to
	}
	return candidate, candidate.End.After(candidate.Start)
}

func mergeIntervals(values []interval) []interval {
	if len(values) == 0 {
		return []interval{}
	}
	values = append([]interval(nil), values...)
	sort.Slice(values, func(i, j int) bool {
		if values[i].Start.Equal(values[j].Start) {
			return values[i].End.Before(values[j].End)
		}
		return values[i].Start.Before(values[j].Start)
	})
	merged := []interval{values[0]}
	for _, current := range values[1:] {
		last := &merged[len(merged)-1]
		if !current.Start.After(last.End) {
			if current.End.After(last.End) {
				last.End = current.End
			}
			continue
		}
		merged = append(merged, current)
	}
	return merged
}

func intersectIntervals(left, right []interval) []interval {
	left = mergeIntervals(left)
	right = mergeIntervals(right)
	result := make([]interval, 0)
	for i, j := 0, 0; i < len(left) && j < len(right); {
		start := left[i].Start
		if right[j].Start.After(start) {
			start = right[j].Start
		}
		end := left[i].End
		if right[j].End.Before(end) {
			end = right[j].End
		}
		if end.After(start) {
			result = append(result, interval{Start: start, End: end})
		}
		if left[i].End.Before(right[j].End) {
			i++
		} else {
			j++
		}
	}
	return result
}

func subtractIntervals(available, busy []interval) []interval {
	result := mergeIntervals(available)
	for _, block := range mergeIntervals(busy) {
		next := make([]interval, 0, len(result)+1)
		for _, candidate := range result {
			if !block.End.After(candidate.Start) || !block.Start.Before(candidate.End) {
				next = append(next, candidate)
				continue
			}
			if block.Start.After(candidate.Start) {
				next = append(next, interval{Start: candidate.Start, End: block.Start})
			}
			if block.End.Before(candidate.End) {
				next = append(next, interval{Start: block.End, End: candidate.End})
			}
		}
		result = next
	}
	return result
}

func chooseSuggestions(windows []interval, duration time.Duration, participantCount int) []TimeSuggestion {
	location, _ := time.LoadLocation(defaultTimezone)
	all := make([]interval, 0, 24)
	for _, window := range windows {
		cursor := ceilToHalfHour(window.Start.In(location))
		for !cursor.Add(duration).After(window.End.In(location)) && len(all) < 60 {
			all = append(all, interval{Start: cursor, End: cursor.Add(duration)})
			cursor = cursor.Add(30 * time.Minute)
		}
	}
	if len(all) == 0 {
		return []TimeSuggestion{}
	}

	selected := make([]interval, 0, 3)
	usedDays := make(map[string]struct{})
	for _, candidate := range all {
		day := candidate.Start.Format("2006-01-02")
		if _, exists := usedDays[day]; exists {
			continue
		}
		selected = append(selected, candidate)
		usedDays[day] = struct{}{}
		if len(selected) == 3 {
			break
		}
	}
	for _, candidate := range all {
		if len(selected) == 3 {
			break
		}
		alreadySelected := false
		for _, existing := range selected {
			if existing.Start.Equal(candidate.Start) {
				alreadySelected = true
				break
			}
		}
		if !alreadySelected {
			selected = append(selected, candidate)
		}
	}

	result := make([]TimeSuggestion, 0, len(selected))
	for _, candidate := range selected {
		result = append(result, TimeSuggestion{
			StartsAt: candidate.Start.UTC().Format(time.RFC3339),
			EndsAt:   candidate.End.UTC().Format(time.RFC3339),
			Label: candidate.Start.Format("1月2日 周") + chineseWeekday(candidate.Start) +
				candidate.Start.Format(" 15:04") + " - " + candidate.End.Format("15:04"),
			Explanation: fmt.Sprintf(
				"%d 位参与者均标记可约，且与已接受的私人日程不冲突",
				participantCount,
			),
		})
	}
	return result
}

func ceilToHalfHour(value time.Time) time.Time {
	minute := value.Minute()
	if minute == 0 || minute == 30 {
		return value.Truncate(time.Minute)
	}
	add := 30 - minute%30
	return value.Truncate(time.Minute).Add(time.Duration(add) * time.Minute)
}

func chineseWeekday(value time.Time) string {
	return [...]string{"日", "一", "二", "三", "四", "五", "六"}[value.Weekday()]
}

func hasBusyConflict(busy []store.BusyInterval, startsAt, endsAt time.Time) bool {
	for _, item := range busy {
		if item.StartsAt.Before(endsAt) && item.EndsAt.After(startsAt) {
			return true
		}
	}
	return false
}
