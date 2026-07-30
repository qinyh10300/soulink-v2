package chat

import (
	"context"
	"net/http"
	"strings"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type Service struct {
	store *store.Store
}

type ConversationView struct {
	ID             int64  `json:"id"`
	PeerUserID     int64  `json:"peerUserId"`
	PeerNickname   string `json:"peerNickname"`
	PeerDepartment string `json:"peerDepartment"`
	PeerGrade      string `json:"peerGrade"`
	LastMessage    string `json:"lastMessage"`
	LastMessageAt  string `json:"lastMessageAt,omitempty"`
	UnreadCount    int    `json:"unreadCount"`
	CanSend        bool   `json:"canSend"`
	StatusMessage  string `json:"statusMessage,omitempty"`
	CreatedAt      string `json:"createdAt"`
}

type MessageView struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversationId"`
	SenderID       int64  `json:"senderId"`
	Content        string `json:"content"`
	ContentType    string `json:"contentType"`
	CreatedAt      string `json:"createdAt"`
}

type MessagesResult struct {
	Items   []MessageView `json:"items"`
	HasMore bool          `json:"hasMore"`
}

func NewService(dataStore *store.Store) *Service {
	return &Service{store: dataStore}
}

func (s *Service) ListConversations(ctx context.Context, userID int64) ([]ConversationView, error) {
	convs, err := s.store.ListConversations(ctx, nil, userID)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]ConversationView, 0, len(convs))
	for _, c := range convs {
		peerID := c.User2ID
		if c.User2ID == userID {
			peerID = c.User1ID
		}
		view := ConversationView{
			ID:             c.ID,
			PeerUserID:     peerID,
			PeerNickname:   c.PeerNickname,
			PeerDepartment: c.PeerDepartment,
			PeerGrade:      c.PeerGrade,
			LastMessage:    c.LastMessageContent,
			UnreadCount:    c.UnreadCount,
			CanSend:        !c.FriendshipDisabledAt.Valid,
			CreatedAt:      c.CreatedAt.Format(time.RFC3339),
		}
		if c.FriendshipDisabledAt.Valid {
			view.StatusMessage = conversationStatusMessage(c.Conversation, userID)
		}
		if c.LastMessageAt.Valid {
			view.LastMessageAt = c.LastMessageAt.Time.Format(time.RFC3339)
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *Service) ListMessages(ctx context.Context, userID, convID int64, beforeID int64, limit int) (MessagesResult, error) {
	conv, err := s.store.GetConversation(ctx, nil, convID)
	if err != nil {
		if err == store.ErrNotFound {
			return MessagesResult{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "会话不存在")
		}
		return MessagesResult{}, httpapi.InternalError(err)
	}
	if conv.User1ID != userID && conv.User2ID != userID {
		return MessagesResult{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权访问")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	messages, err := s.store.ListMessages(ctx, nil, convID, beforeID, limit+1)
	if err != nil {
		return MessagesResult{}, httpapi.InternalError(err)
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	items := make([]MessageView, 0, len(messages))
	for _, m := range messages {
		items = append(items, toMessageView(m))
	}
	return MessagesResult{Items: items, HasMore: hasMore}, nil
}

func (s *Service) MarkRead(ctx context.Context, userID, convID int64) error {
	conv, err := s.store.GetConversation(ctx, nil, convID)
	if err != nil {
		if err == store.ErrNotFound {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "会话不存在")
		}
		return httpapi.InternalError(err)
	}
	if conv.User1ID != userID && conv.User2ID != userID {
		return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权访问")
	}
	if err := s.store.MarkConversationRead(ctx, nil, convID, userID, time.Now()); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) SendMessage(ctx context.Context, userID, convID int64, content, contentType string) (store.Message, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return store.Message{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "消息内容不能为空")
	}
	if contentType == "" {
		contentType = "text"
	}

	conv, err := s.store.GetConversation(ctx, nil, convID)
	if err != nil {
		if err == store.ErrNotFound {
			return store.Message{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "会话不存在")
		}
		return store.Message{}, httpapi.InternalError(err)
	}
	if conv.User1ID != userID && conv.User2ID != userID {
		return store.Message{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权访问")
	}
	if conv.FriendshipDisabledAt.Valid {
		return store.Message{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, conversationStatusMessage(conv, userID))
	}

	msg, err := s.store.CreateMessage(ctx, nil, convID, userID, content, contentType, time.Now())
	if err != nil {
		return store.Message{}, httpapi.InternalError(err)
	}
	return msg, nil
}

func (s *Service) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	count, err := s.store.GetTotalUnreadCount(ctx, nil, userID)
	if err != nil {
		return 0, httpapi.InternalError(err)
	}
	return count, nil
}

func toMessageView(m store.Message) MessageView {
	return MessageView{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Content:        m.Content,
		ContentType:    m.ContentType,
		CreatedAt:      m.CreatedAt.Format(time.RFC3339),
	}
}

func ToMessageView(m store.Message) MessageView {
	return toMessageView(m)
}

func conversationStatusMessage(conv store.Conversation, userID int64) string {
	if conv.FriendshipDisabledByUserID.Valid && conv.FriendshipDisabledByUserID.Int64 == userID {
		return "你已经删除对方好友"
	}
	return "对方已经将你删除好友"
}
