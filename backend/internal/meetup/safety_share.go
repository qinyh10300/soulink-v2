package meetup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	stdmail "net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/httpapi"
	appmail "soulink-backend/internal/mail"
	"soulink-backend/internal/store"
)

var trustedPhonePattern = regexp.MustCompile(`^\+?[0-9][0-9 -]{5,19}$`)

type TrustedContactInput struct {
	Label        string
	ContactType  string
	ContactValue string
}

type TrustedContactView struct {
	ID           int64  `json:"id"`
	Label        string `json:"label"`
	ContactType  string `json:"contactType"`
	ContactValue string `json:"contactValue"`
	UpdatedAt    string `json:"updatedAt"`
}

type MeetupSafetySummary struct {
	MeetupID         int64  `json:"meetupId"`
	StartsAt         string `json:"startsAt"`
	EndsAt           string `json:"endsAt"`
	Location         string `json:"location"`
	ParticipantCount int    `json:"participantCount"`
	Status           string `json:"status"`
}

type MeetupSafetyShareView struct {
	ID             int64               `json:"id"`
	TrustedContact TrustedContactView  `json:"trustedContact"`
	Summary        MeetupSafetySummary `json:"summary"`
	DeliveryStatus string              `json:"deliveryStatus"`
	ManualText     string              `json:"manualText,omitempty"`
	CreatedAt      string              `json:"createdAt"`
}

func (s *Service) ListTrustedContacts(ctx context.Context, userID int64) ([]TrustedContactView, error) {
	contacts, err := s.store.ListTrustedContacts(ctx, nil, userID)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]TrustedContactView, 0, len(contacts))
	for _, contact := range contacts {
		views = append(views, trustedContactView(contact))
	}
	return views, nil
}

func (s *Service) CreateTrustedContact(
	ctx context.Context,
	userID int64,
	input TrustedContactInput,
) (TrustedContactView, error) {
	input.Label = strings.TrimSpace(input.Label)
	input.ContactType = strings.TrimSpace(strings.ToLower(input.ContactType))
	input.ContactValue = strings.TrimSpace(input.ContactValue)
	if err := validateTrustedContact(input); err != nil {
		return TrustedContactView{}, err
	}
	created, err := s.store.CreateTrustedContact(ctx, nil, store.TrustedContact{
		UserID: userID, Label: input.Label, ContactType: input.ContactType,
		ContactValue: input.ContactValue,
	}, s.now())
	if err != nil {
		if strings.Contains(err.Error(), "trusted_contacts_distinct_value") {
			return TrustedContactView{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "这个可信联系人已经存在")
		}
		return TrustedContactView{}, httpapi.InternalError(err)
	}
	return trustedContactView(created), nil
}

func (s *Service) DeleteTrustedContact(ctx context.Context, userID, contactID int64) error {
	if contactID <= 0 {
		return invalidParams("可信联系人编号不正确")
	}
	if err := s.store.DeleteTrustedContact(ctx, nil, userID, contactID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "可信联系人不存在")
		}
		if strings.Contains(err.Error(), "meetup_safety_shares") {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "已有安全分享使用这个联系人，暂时不能删除")
		}
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) ShareMeetupSafetySummary(
	ctx context.Context,
	userID, meetupID, contactID int64,
) (MeetupSafetyShareView, error) {
	meetup, err := s.store.GetMeetup(ctx, nil, userID, meetupID)
	if err != nil {
		return MeetupSafetyShareView{}, mapMeetupStoreError(err)
	}
	if meetup.Status != "confirmed" {
		return MeetupSafetyShareView{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "只有已确认的约见可以分享安全摘要")
	}
	participants, err := s.store.ListMeetupParticipants(ctx, nil, meetupID)
	if err != nil {
		return MeetupSafetyShareView{}, httpapi.InternalError(err)
	}
	accepted := false
	for _, participant := range participants {
		if participant.UserID == userID && participant.Status == "accepted" {
			accepted = true
			break
		}
	}
	if !accepted {
		return MeetupSafetyShareView{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "接受约见后才能分享安全摘要")
	}
	contact, err := s.store.GetTrustedContact(ctx, nil, userID, contactID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return MeetupSafetyShareView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "可信联系人不存在")
		}
		return MeetupSafetyShareView{}, httpapi.InternalError(err)
	}
	summary := MeetupSafetySummary{
		MeetupID: meetup.ID, StartsAt: meetup.StartsAt.Format(time.RFC3339),
		EndsAt: meetup.EndsAt.Format(time.RFC3339), Location: meetup.Location,
		ParticipantCount: len(participants), Status: meetup.Status,
	}
	rawSummary, err := json.Marshal(summary)
	if err != nil {
		return MeetupSafetyShareView{}, httpapi.InternalError(err)
	}
	now := s.now()
	created, err := s.store.CreateMeetupSafetyShare(ctx, nil, store.MeetupSafetyShare{
		MeetupID: meetupID, SharerID: userID, TrustedContactID: contactID,
		Summary: rawSummary, DeliveryStatus: "recorded",
	}, now)
	if err != nil {
		return MeetupSafetyShareView{}, httpapi.InternalError(err)
	}
	manualText := safetyShareText(summary)
	if contact.ContactType == "email" {
		if s.mailer == nil {
			_ = s.store.UpdateMeetupSafetyShareDelivery(ctx, nil, created.ID, "failed", sql.NullTime{})
			return MeetupSafetyShareView{}, httpapi.FeatureNotReady("meetup safety mail")
		}
		user, err := s.store.GetUserByID(ctx, nil, userID)
		if err != nil {
			return MeetupSafetyShareView{}, httpapi.InternalError(err)
		}
		err = s.mailer.Send(ctx, appmail.Message{
			To: contact.ContactValue, Subject: "SouLink 约见安全摘要",
			Text: fmt.Sprintf("%s 主动向你分享了一条 SouLink 校内约见安全摘要。\n\n%s\n\n该摘要不包含聊天内容、约见标题或其他参与者资料。", user.Nickname, manualText),
		})
		if err != nil {
			_ = s.store.UpdateMeetupSafetyShareDelivery(ctx, nil, created.ID, "failed", sql.NullTime{})
			return MeetupSafetyShareView{}, httpapi.NewError(http.StatusBadGateway, httpapi.CodeInternalError, "安全摘要邮件发送失败，请稍后重试")
		}
		created.DeliveryStatus = "sent"
		created.DeliveredAt = sql.NullTime{Time: now, Valid: true}
		if err := s.store.UpdateMeetupSafetyShareDelivery(ctx, nil, created.ID, "sent", created.DeliveredAt); err != nil {
			return MeetupSafetyShareView{}, httpapi.InternalError(err)
		}
		manualText = ""
	}
	return MeetupSafetyShareView{
		ID: created.ID, TrustedContact: trustedContactView(contact), Summary: summary,
		DeliveryStatus: created.DeliveryStatus, ManualText: manualText,
		CreatedAt: created.CreatedAt.Format(time.RFC3339),
	}, nil
}

func validateTrustedContact(input TrustedContactInput) error {
	if utf8.RuneCountInString(input.Label) < 1 || utf8.RuneCountInString(input.Label) > 64 ||
		utf8.RuneCountInString(input.ContactValue) < 3 || utf8.RuneCountInString(input.ContactValue) > 255 {
		return invalidParams("可信联系人信息不完整")
	}
	switch input.ContactType {
	case "email":
		address, err := stdmail.ParseAddress(input.ContactValue)
		if err != nil || !strings.EqualFold(address.Address, input.ContactValue) {
			return invalidParams("可信联系人邮箱不正确")
		}
	case "phone":
		if !trustedPhonePattern.MatchString(input.ContactValue) {
			return invalidParams("可信联系人手机号不正确")
		}
	case "wechat", "other":
	default:
		return invalidParams("可信联系人类型不正确")
	}
	return nil
}

func safetyShareText(summary MeetupSafetySummary) string {
	location := summary.Location
	if strings.TrimSpace(location) == "" {
		location = "校内地点待确认"
	}
	zone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		zone = time.FixedZone("CST", 8*60*60)
	}
	start, _ := time.Parse(time.RFC3339, summary.StartsAt)
	end, _ := time.Parse(time.RFC3339, summary.EndsAt)
	return fmt.Sprintf(
		"时间：%s 至 %s\n地点：%s\n参与人数：%d\n状态：已确认",
		start.In(zone).Format("2006-01-02 15:04"), end.In(zone).Format("15:04"),
		location, summary.ParticipantCount,
	)
}

func trustedContactView(contact store.TrustedContact) TrustedContactView {
	return TrustedContactView{
		ID: contact.ID, Label: contact.Label, ContactType: contact.ContactType,
		ContactValue: contact.ContactValue, UpdatedAt: contact.UpdatedAt.Format(time.RFC3339),
	}
}
