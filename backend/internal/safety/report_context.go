package safety

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/store"
)

const (
	reportSnapshotVersion  = 1
	reportSnapshotMessages = 10
	reportSnapshotRunes    = 500
)

type resolvedReportContext struct {
	reportedUserID int64
	intentID       sql.NullInt64
	roomID         sql.NullInt64
	conversationID sql.NullInt64
	groupID        sql.NullInt64
	snapshot       json.RawMessage
	riskSignals    json.RawMessage
}

type reportContextSnapshot struct {
	Version      int                     `json:"version"`
	ContextType  string                  `json:"contextType"`
	ContextTitle string                  `json:"contextTitle"`
	CapturedAt   time.Time               `json:"capturedAt"`
	Content      *reportContentSnapshot  `json:"content,omitempty"`
	Messages     []reportMessageSnapshot `json:"messages,omitempty"`
}

type reportContentSnapshot struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type reportMessageSnapshot struct {
	ID         int64     `json:"id"`
	SenderRole string    `json:"senderRole"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (s *Service) resolveReportContext(
	ctx context.Context,
	reporterID int64,
	input ReportInput,
	capturedAt time.Time,
) (resolvedReportContext, error) {
	resolved := resolvedReportContext{}
	snapshot := reportContextSnapshot{
		Version:     reportSnapshotVersion,
		ContextType: input.ContextType,
		CapturedAt:  capturedAt,
	}
	riskTexts := []string{input.Details}

	switch input.ContextType {
	case ContextIntent:
		intent, err := s.store.GetConnectionIntent(ctx, nil, reporterID, input.ContextID)
		if err != nil {
			return resolved, mapStoreError(err, "找不到这个意图")
		}
		resolved.reportedUserID = intent.AuthorID
		resolved.intentID = sql.NullInt64{Int64: intent.ID, Valid: true}
		snapshot.ContextTitle = truncateReportText(intent.Title)
		snapshot.Content = &reportContentSnapshot{
			Category:    truncateReportText(intent.Category),
			Title:       truncateReportText(intent.Title),
			Description: truncateReportText(intent.Description),
		}
		riskTexts = append(riskTexts, intent.Title, intent.Description)

	case ContextMatchRoom:
		room, err := s.store.GetMatchRoom(ctx, nil, reporterID, input.ContextID)
		if err != nil {
			return resolved, mapStoreError(err, "找不到这个匹配房间")
		}
		resolved.reportedUserID = room.OwnerID
		if resolved.reportedUserID == reporterID {
			resolved.reportedUserID = room.CandidateID
		}
		resolved.roomID = sql.NullInt64{Int64: room.ID, Valid: true}
		snapshot.ContextTitle = truncateReportText(room.IntentTitle)
		messages, err := s.store.ListMatchRoomMessages(ctx, nil, room.ID, reporterID, 100)
		if err != nil {
			return resolved, mapStoreError(err, "找不到这个匹配房间")
		}
		if len(messages) > reportSnapshotMessages {
			messages = messages[len(messages)-reportSnapshotMessages:]
		}
		for _, message := range messages {
			snapshot.Messages = append(snapshot.Messages, reportMessageSnapshot{
				ID: message.ID, SenderRole: reportSenderRole(message.SenderID, reporterID, resolved.reportedUserID),
				Content: truncateReportText(message.Content), CreatedAt: message.CreatedAt,
			})
			riskTexts = append(riskTexts, message.Content)
		}

	case ContextDirect:
		conversation, err := s.store.GetConversation(ctx, nil, input.ContextID)
		if err != nil {
			return resolved, mapStoreError(err, "找不到这个会话")
		}
		if reporterID != conversation.User1ID && reporterID != conversation.User2ID {
			return resolved, forbidden("你不在这个会话中")
		}
		resolved.reportedUserID = conversation.User1ID
		if resolved.reportedUserID == reporterID {
			resolved.reportedUserID = conversation.User2ID
		}
		resolved.conversationID = sql.NullInt64{Int64: conversation.ID, Valid: true}
		snapshot.ContextTitle = "一对一会话"
		messages, err := s.store.ListMessages(ctx, nil, conversation.ID, 0, reportSnapshotMessages)
		if err != nil {
			return resolved, mapStoreError(err, "找不到这个会话")
		}
		for index := len(messages) - 1; index >= 0; index-- {
			message := messages[index]
			snapshot.Messages = append(snapshot.Messages, reportMessageSnapshot{
				ID: message.ID, SenderRole: reportSenderRole(message.SenderID, reporterID, resolved.reportedUserID),
				Content: truncateReportText(message.Content), CreatedAt: message.CreatedAt,
			})
			riskTexts = append(riskTexts, message.Content)
		}

	case ContextGroup:
		if input.ReportedUserID <= 0 {
			return resolved, invalidParams("请选择需要举报的群成员")
		}
		group, err := s.store.GetChatGroup(ctx, nil, input.ContextID)
		if err != nil {
			return resolved, mapStoreError(err, "找不到这个群聊")
		}
		if _, err := s.store.GetChatGroupMember(ctx, nil, group.ID, reporterID); err != nil {
			return resolved, forbidden("你不在这个群聊中")
		}
		if _, err := s.store.GetChatGroupMember(ctx, nil, group.ID, input.ReportedUserID); err != nil {
			return resolved, invalidParams("请选择当前群聊中的成员")
		}
		resolved.reportedUserID = input.ReportedUserID
		resolved.groupID = sql.NullInt64{Int64: group.ID, Valid: true}
		snapshot.ContextTitle = truncateReportText(group.Name)
		messages, err := s.store.ListChatGroupMessages(ctx, nil, group.ID, 0, 100)
		if err != nil {
			return resolved, mapStoreError(err, "找不到这个群聊")
		}
		selected := make([]store.ChatGroupMessage, 0, reportSnapshotMessages)
		for _, message := range messages {
			if message.SenderID != reporterID && message.SenderID != resolved.reportedUserID {
				continue
			}
			selected = append(selected, message)
			if len(selected) == reportSnapshotMessages {
				break
			}
		}
		for index := len(selected) - 1; index >= 0; index-- {
			message := selected[index]
			snapshot.Messages = append(snapshot.Messages, reportMessageSnapshot{
				ID: message.ID, SenderRole: reportSenderRole(message.SenderID, reporterID, resolved.reportedUserID),
				Content: truncateReportText(message.Content), CreatedAt: message.CreatedAt,
			})
			riskTexts = append(riskTexts, message.Content)
		}

	default:
		return resolved, invalidParams("未知的举报场景")
	}

	if resolved.reportedUserID == reporterID {
		return resolved, invalidParams("不能举报自己的内容")
	}
	encodedSnapshot, err := json.Marshal(snapshot)
	if err != nil {
		return resolved, err
	}
	encodedSignals, err := json.Marshal(deriveReportRiskSignals(riskTexts...))
	if err != nil {
		return resolved, err
	}
	resolved.snapshot = encodedSnapshot
	resolved.riskSignals = encodedSignals
	return resolved, nil
}

func reportSenderRole(senderID, reporterID, reportedUserID int64) string {
	if senderID == reporterID {
		return "reporter"
	}
	if senderID == reportedUserID {
		return "reported_user"
	}
	return "other"
}

func truncateReportText(value string) string {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) <= reportSnapshotRunes {
		return value
	}
	runes := []rune(value)
	return string(runes[:reportSnapshotRunes])
}

func deriveReportRiskSignals(values ...string) []string {
	text := strings.ToLower(strings.Join(values, "\n"))
	patterns := []struct {
		name     string
		keywords []string
	}{
		{name: "spam", keywords: []string{"广告", "推广", "返现", "刷单", "代理", "优惠券"}},
		{name: "contact_inducement", keywords: []string{"加微信", "加我微信", "vx", "v信", "qq", "手机号", "私聊我"}},
		{name: "harassment", keywords: []string{"骚扰", "傻逼", "废物", "滚开", "恶心", "垃圾人"}},
		{name: "threat", keywords: []string{"威胁", "打死", "弄死", "杀了", "报复", "你等着"}},
		{name: "privacy", keywords: []string{"身份证", "住址", "宿舍号", "曝光", "人肉", "照片发出去"}},
		{name: "coercion", keywords: []string{"不许拒绝", "必须来", "不来就", "别告诉别人", "强迫", "胁迫"}},
	}
	result := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		for _, keyword := range pattern.keywords {
			if strings.Contains(text, keyword) {
				result = append(result, pattern.name)
				break
			}
		}
	}
	return result
}
