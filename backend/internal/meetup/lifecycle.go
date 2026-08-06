package meetup

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/analytics"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

func (s *Service) Respond(
	ctx context.Context,
	userID int64,
	meetupID int64,
	action string,
	expectedVersion int,
) (MeetupView, error) {
	action = strings.ToLower(strings.TrimSpace(action))
	if action != "accept" && action != "decline" {
		return MeetupView{}, invalidParams("请选择接受或拒绝")
	}
	now := s.now()
	confirmed := false
	var category string
	var participantCount int
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		record, err := s.store.GetMeetupForUpdate(ctx, tx, userID, meetupID)
		if err != nil {
			return err
		}
		if err := requireExpectedVersion(record, expectedVersion); err != nil {
			return err
		}
		if record.Status != "proposed" && record.Status != "reschedule_requested" {
			return conflict("这项约见已不能回复")
		}
		category = record.Category
		participantStatus := "accepted"
		if action == "decline" {
			participantStatus = "declined"
		}
		if err := s.store.UpdateMeetupParticipantStatus(
			ctx,
			tx,
			record.ID,
			userID,
			participantStatus,
			now,
		); err != nil {
			return err
		}
		participants, err := s.store.ListMeetupParticipants(ctx, tx, record.ID)
		if err != nil {
			return err
		}
		participantCount = len(participants)
		if action == "decline" {
			if record.CalendarEventID.Valid {
				if err := s.store.CancelCalendarEvent(ctx, tx, record.CalendarEventID.Int64, now); err != nil {
					return err
				}
			}
			return s.store.UpdateMeetup(ctx, tx, store.UpdateMeetupParams{
				MeetupID:        record.ID,
				ExpectedVersion: record.Version,
				StartsAt:        record.StartsAt,
				EndsAt:          record.EndsAt,
				Location:        record.Location,
				Status:          "declined",
				Now:             now,
			})
		}
		for _, participant := range participants {
			if participant.Status != "accepted" {
				return nil
			}
		}
		calendarEventID, err := s.createCalendarProjection(ctx, tx, record, participants, now)
		if err != nil {
			return err
		}
		confirmed = true
		return s.store.UpdateMeetup(ctx, tx, store.UpdateMeetupParams{
			MeetupID:        record.ID,
			ExpectedVersion: record.Version,
			StartsAt:        record.StartsAt,
			EndsAt:          record.EndsAt,
			Location:        record.Location,
			Status:          "confirmed",
			CalendarEventID: sql.NullInt64{Int64: calendarEventID, Valid: true},
			Now:             now,
		})
	})
	if err != nil {
		return MeetupView{}, wrapStoreTxError(err)
	}
	if confirmed {
		s.record(ctx, analytics.EventMeetupConfirmed, userID, meetupID, category, participantCount, "confirmed")
	}
	return s.Get(ctx, userID, meetupID)
}

func (s *Service) Reschedule(
	ctx context.Context,
	userID int64,
	meetupID int64,
	input RescheduleInput,
) (MeetupView, error) {
	record, err := s.store.GetMeetup(ctx, nil, userID, meetupID)
	if err != nil {
		return MeetupView{}, mapMeetupStoreError(err)
	}
	if err := s.validateMeetupDetails(
		record.Title,
		input.StartsAt,
		input.EndsAt,
		strings.TrimSpace(input.Location),
		record.CalendarEventID.Int64,
	); err != nil {
		return MeetupView{}, err
	}
	participants, err := s.store.ListMeetupParticipants(ctx, nil, meetupID)
	if err != nil {
		return MeetupView{}, httpapi.InternalError(err)
	}
	participantIDs := make([]int64, 0, len(participants))
	for _, participant := range participants {
		participantIDs = append(participantIDs, participant.UserID)
	}
	busy, err := s.store.ListBusyIntervals(
		ctx,
		nil,
		participantIDs,
		input.StartsAt,
		input.EndsAt,
		record.CalendarEventID.Int64,
	)
	if err != nil {
		return MeetupView{}, httpapi.InternalError(err)
	}
	if hasBusyConflict(busy, input.StartsAt, input.EndsAt) {
		return MeetupView{}, conflict("新时段与至少一位参与者已接受的日程冲突")
	}

	now := s.now()
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		current, err := s.store.GetMeetupForUpdate(ctx, tx, userID, meetupID)
		if err != nil {
			return err
		}
		if err := requireExpectedVersion(current, input.ExpectedVersion); err != nil {
			return err
		}
		if current.Status != "proposed" && current.Status != "reschedule_requested" &&
			current.Status != "confirmed" {
			return conflict("这项约见已不能改期")
		}
		if current.CalendarEventID.Valid {
			if err := s.store.CancelCalendarEvent(ctx, tx, current.CalendarEventID.Int64, now); err != nil {
				return err
			}
		}
		if err := s.store.ResetMeetupParticipants(ctx, tx, meetupID, userID, now); err != nil {
			return err
		}
		return s.store.UpdateMeetup(ctx, tx, store.UpdateMeetupParams{
			MeetupID:        meetupID,
			ExpectedVersion: current.Version,
			StartsAt:        input.StartsAt,
			EndsAt:          input.EndsAt,
			Location:        strings.TrimSpace(input.Location),
			Status:          "reschedule_requested",
			Now:             now,
		})
	})
	if err != nil {
		return MeetupView{}, wrapStoreTxError(err)
	}
	s.record(ctx, analytics.EventMeetupProposed, userID, meetupID, record.Category, len(participants), "reschedule_requested")
	return s.Get(ctx, userID, meetupID)
}

func (s *Service) Cancel(
	ctx context.Context,
	userID int64,
	meetupID int64,
	expectedVersion int,
	reason string,
) (MeetupView, error) {
	reason = strings.TrimSpace(reason)
	if utf8.RuneCountInString(reason) > maxCancellationLength {
		return MeetupView{}, invalidParams("取消原因不能超过 240 个字")
	}
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		record, err := s.store.GetMeetupForUpdate(ctx, tx, userID, meetupID)
		if err != nil {
			return err
		}
		if err := requireExpectedVersion(record, expectedVersion); err != nil {
			return err
		}
		if record.Status != "proposed" && record.Status != "reschedule_requested" &&
			record.Status != "confirmed" {
			return conflict("这项约见已不能取消")
		}
		if record.CalendarEventID.Valid {
			if err := s.store.CancelCalendarEvent(ctx, tx, record.CalendarEventID.Int64, now); err != nil {
				return err
			}
		}
		return s.store.UpdateMeetup(ctx, tx, store.UpdateMeetupParams{
			MeetupID:        record.ID,
			ExpectedVersion: record.Version,
			StartsAt:        record.StartsAt,
			EndsAt:          record.EndsAt,
			Location:        record.Location,
			Status:          "cancelled",
			CancelledBy:     sql.NullInt64{Int64: userID, Valid: true},
			CancelledReason: reason,
			Now:             now,
		})
	})
	if err != nil {
		return MeetupView{}, wrapStoreTxError(err)
	}
	return s.Get(ctx, userID, meetupID)
}

func (s *Service) Complete(
	ctx context.Context,
	userID int64,
	meetupID int64,
	expectedVersion int,
	outcome string,
) (MeetupView, error) {
	outcome = strings.ToLower(strings.TrimSpace(outcome))
	if outcome != "completed" && outcome != "no_show" {
		return MeetupView{}, invalidParams("完成结果不正确")
	}
	var category string
	var participantCount int
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		record, err := s.store.GetMeetupForUpdate(ctx, tx, userID, meetupID)
		if err != nil {
			return err
		}
		if err := requireExpectedVersion(record, expectedVersion); err != nil {
			return err
		}
		if record.Status != "confirmed" {
			return conflict("只有已确认的约见才能记录结果")
		}
		if now.Before(record.StartsAt) {
			return invalidParams("约见开始后才能记录结果")
		}
		participants, err := s.store.ListMeetupParticipants(ctx, tx, meetupID)
		if err != nil {
			return err
		}
		category = record.Category
		participantCount = len(participants)
		return s.store.UpdateMeetup(ctx, tx, store.UpdateMeetupParams{
			MeetupID:        record.ID,
			ExpectedVersion: record.Version,
			StartsAt:        record.StartsAt,
			EndsAt:          record.EndsAt,
			Location:        record.Location,
			Status:          outcome,
			CalendarEventID: record.CalendarEventID,
			Now:             now,
		})
	})
	if err != nil {
		return MeetupView{}, wrapStoreTxError(err)
	}
	eventName := analytics.EventMeetupCompleted
	if outcome == "no_show" {
		eventName = analytics.EventMeetupNoShow
	}
	s.record(ctx, eventName, userID, meetupID, category, participantCount, outcome)
	return s.Get(ctx, userID, meetupID)
}

func (s *Service) SubmitFeedback(
	ctx context.Context,
	userID int64,
	meetupID int64,
	input FeedbackInput,
) (MeetupView, error) {
	input.Attendance = strings.ToLower(strings.TrimSpace(input.Attendance))
	input.PrivateNotes = strings.TrimSpace(input.PrivateNotes)
	if input.Attendance != "attended" && input.Attendance != "no_show" &&
		input.Attendance != "cancelled" {
		return MeetupView{}, invalidParams("见面情况不正确")
	}
	if input.Rating != nil && (*input.Rating < 1 || *input.Rating > 5) {
		return MeetupView{}, invalidParams("评分需要在 1 到 5 之间")
	}
	if utf8.RuneCountInString(input.PrivateNotes) > maxFeedbackNotesLength {
		return MeetupView{}, invalidParams("私密备注不能超过 1000 个字")
	}
	record, err := s.store.GetMeetup(ctx, nil, userID, meetupID)
	if err != nil {
		return MeetupView{}, mapMeetupStoreError(err)
	}
	if s.now().Before(record.StartsAt) && record.Status != "cancelled" {
		return MeetupView{}, invalidParams("约见开始后才能提交反馈")
	}
	if record.Status == "proposed" || record.Status == "reschedule_requested" ||
		record.Status == "declined" {
		return MeetupView{}, conflict("这项约见还不能提交反馈")
	}
	participants, err := s.store.ListMeetupParticipants(ctx, nil, meetupID)
	if err != nil {
		return MeetupView{}, httpapi.InternalError(err)
	}
	participantIDs := make([]int64, 0, len(participants))
	for _, participant := range participants {
		participantIDs = append(participantIDs, participant.UserID)
	}
	rating := sql.NullInt64{}
	if input.Rating != nil {
		rating = sql.NullInt64{Int64: int64(*input.Rating), Valid: true}
	}
	wouldMeetAgain := sql.NullBool{}
	if input.WouldMeetAgain != nil {
		wouldMeetAgain = sql.NullBool{Bool: *input.WouldMeetAgain, Valid: true}
	}
	signalType, weight := signalFromFeedback(input)
	now := s.now()
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := s.store.GetMeetupForUpdate(ctx, tx, userID, meetupID); err != nil {
			return err
		}
		if _, err := s.store.UpsertMeetupFeedback(ctx, tx, store.UpsertMeetupFeedbackParams{
			MeetupID:       meetupID,
			UserID:         userID,
			Attendance:     input.Attendance,
			Rating:         rating,
			WouldMeetAgain: wouldMeetAgain,
			SafetyConcern:  input.SafetyConcern,
			PrivateNotes:   input.PrivateNotes,
			Now:            now,
		}); err != nil {
			return err
		}
		if signalType == "" {
			return s.store.RevokeMeetupLearningSignals(ctx, tx, meetupID, userID, now)
		}
		return s.store.ReplaceMeetupLearningSignals(
			ctx,
			tx,
			meetupID,
			userID,
			participantIDs,
			signalType,
			weight,
			now,
		)
	})
	if err != nil {
		return MeetupView{}, wrapStoreTxError(err)
	}
	s.record(
		ctx,
		analytics.EventMeetingFeedbackSubmitted,
		userID,
		meetupID,
		record.Category,
		len(participants),
		input.Attendance,
	)
	if input.Attendance == "no_show" {
		s.record(ctx, analytics.EventMeetupNoShow, userID, meetupID, record.Category, len(participants), "feedback")
	}
	return s.Get(ctx, userID, meetupID)
}

func (s *Service) RevokeFeedback(
	ctx context.Context,
	userID int64,
	meetupID int64,
) (MeetupView, error) {
	now := s.now()
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := s.store.GetMeetupForUpdate(ctx, tx, userID, meetupID); err != nil {
			return err
		}
		if err := s.store.RevokeMeetupFeedback(ctx, tx, meetupID, userID, now); err != nil {
			return err
		}
		return s.store.RevokeMeetupLearningSignals(ctx, tx, meetupID, userID, now)
	})
	if err != nil {
		return MeetupView{}, wrapStoreTxError(err)
	}
	return s.Get(ctx, userID, meetupID)
}

func (s *Service) createCalendarProjection(
	ctx context.Context,
	tx *sql.Tx,
	record store.Meetup,
	participants []store.MeetupParticipant,
	now time.Time,
) (int64, error) {
	event, err := s.store.CreateCalendarEvent(ctx, tx, store.CreateCalendarEventParams{
		OrganizerID: record.OrganizerID,
		EventType:   calendarEventType(record.Category),
		Title:       "SouLink 约见 · " + record.Title,
		Description: "由 SouLink 双向确认的校内约见",
		Location:    record.Location,
		StartsAt:    record.StartsAt,
		EndsAt:      record.EndsAt,
		Visibility:  "busy_only",
		Now:         now,
	})
	if err != nil {
		return 0, err
	}
	for _, participant := range participants {
		if participant.UserID == record.OrganizerID {
			continue
		}
		if err := s.store.AddCalendarAttendee(ctx, tx, event.ID, participant.UserID, now); err != nil {
			return 0, err
		}
		if err := s.store.UpdateCalendarAttendeeStatus(
			ctx,
			tx,
			event.ID,
			participant.UserID,
			"accepted",
			now,
		); err != nil {
			return 0, err
		}
	}
	return event.ID, nil
}

func isExpectedConflict(err error) bool {
	var appErr *httpapi.AppError
	return errors.As(err, &appErr) && appErr.HTTPStatus == 409
}
