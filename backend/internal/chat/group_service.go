package chat

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

const (
	maxGroupNameLength   = 40
	maxGroupMemberCount  = 50
	maxChatMessageLength = 4000
)

type GroupMemberView struct {
	UserID     int64  `json:"userId"`
	Nickname   string `json:"nickname"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
	Role       string `json:"role"`
	IsMe       bool   `json:"isMe"`
	JoinedAt   string `json:"joinedAt"`
}

type GroupView struct {
	ID                int64             `json:"id"`
	Name              string            `json:"name"`
	OwnerID           int64             `json:"ownerId"`
	IsOwner           bool              `json:"isOwner"`
	MemberCount       int               `json:"memberCount"`
	Members           []GroupMemberView `json:"members"`
	LastMessage       string            `json:"lastMessage"`
	LastMessageAt     string            `json:"lastMessageAt,omitempty"`
	LastMessageSender string            `json:"lastMessageSender,omitempty"`
	UnreadCount       int               `json:"unreadCount"`
	CreatedAt         string            `json:"createdAt"`
}

type GroupMessageView struct {
	ID             int64  `json:"id"`
	GroupID        int64  `json:"groupId"`
	SenderID       int64  `json:"senderId"`
	SenderNickname string `json:"senderNickname"`
	Content        string `json:"content"`
	ContentType    string `json:"contentType"`
	CreatedAt      string `json:"createdAt"`
}

type GroupMessagesResult struct {
	Items   []GroupMessageView `json:"items"`
	HasMore bool               `json:"hasMore"`
}

type CreateGroupInput struct {
	Name      string
	MemberIDs []int64
}

func (s *Service) EnsureConversation(ctx context.Context, userID, peerID int64) (ConversationView, error) {
	if peerID <= 0 || peerID == userID {
		return ConversationView{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "无法创建该会话")
	}
	friendship, err := s.store.GetFriendshipBetween(ctx, nil, userID, peerID)
	if err != nil || friendship.Status != "accepted" {
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return ConversationView{}, httpapi.InternalError(err)
		}
		return ConversationView{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "成为好友后才能发消息")
	}
	conversation, err := s.store.CreateConversation(ctx, nil, userID, peerID, time.Now())
	if err != nil {
		return ConversationView{}, httpapi.InternalError(err)
	}
	conversations, err := s.ListConversations(ctx, userID)
	if err != nil {
		return ConversationView{}, err
	}
	for _, view := range conversations {
		if view.ID == conversation.ID {
			return view, nil
		}
	}
	return ConversationView{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "会话不存在")
}

func (s *Service) ListGroups(ctx context.Context, userID int64) ([]GroupView, error) {
	groups, err := s.store.ListChatGroups(ctx, nil, userID)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]GroupView, 0, len(groups))
	for _, group := range groups {
		views = append(views, groupSummaryView(group, userID))
	}
	return views, nil
}

func (s *Service) GetGroup(ctx context.Context, userID, groupID int64) (GroupView, error) {
	group, err := s.requireGroupMember(ctx, groupID, userID)
	if err != nil {
		return GroupView{}, err
	}
	members, err := s.store.ListChatGroupMembers(ctx, nil, groupID)
	if err != nil {
		return GroupView{}, httpapi.InternalError(err)
	}
	return groupDetailView(group, members, userID), nil
}

func (s *Service) CreateGroup(ctx context.Context, ownerID int64, input CreateGroupInput) (GroupView, error) {
	name, err := normalizeGroupName(input.Name)
	if err != nil {
		return GroupView{}, err
	}
	memberIDs, err := normalizeMemberIDs(ownerID, input.MemberIDs)
	if err != nil {
		return GroupView{}, err
	}
	if len(memberIDs) < 2 {
		return GroupView{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "至少选择两位好友")
	}
	if err := s.validateFriends(ctx, ownerID, memberIDs); err != nil {
		return GroupView{}, err
	}

	var group store.ChatGroup
	now := time.Now()
	err = s.store.WithTx(ctx, func(tx *sql.Tx) error {
		created, err := s.store.CreateChatGroup(ctx, tx, ownerID, name, now)
		if err != nil {
			return httpapi.InternalError(err)
		}
		group = created
		if err := s.store.UpsertChatGroupMember(ctx, tx, group.ID, ownerID, "owner", now); err != nil {
			return httpapi.InternalError(err)
		}
		for _, memberID := range memberIDs {
			if err := s.store.UpsertChatGroupMember(ctx, tx, group.ID, memberID, "member", now); err != nil {
				return httpapi.InternalError(err)
			}
		}
		return nil
	})
	if err != nil {
		return GroupView{}, err
	}
	members, err := s.store.ListChatGroupMembers(ctx, nil, group.ID)
	if err != nil {
		return GroupView{}, httpapi.InternalError(err)
	}
	return groupDetailView(group, members, ownerID), nil
}

func (s *Service) RenameGroup(ctx context.Context, userID, groupID int64, name string) (GroupView, error) {
	group, err := s.requireGroupOwner(ctx, groupID, userID)
	if err != nil {
		return GroupView{}, err
	}
	normalized, err := normalizeGroupName(name)
	if err != nil {
		return GroupView{}, err
	}
	if err := s.store.UpdateChatGroupName(ctx, nil, groupID, normalized, time.Now()); err != nil {
		return GroupView{}, httpapi.InternalError(err)
	}
	group.Name = normalized
	members, err := s.store.ListChatGroupMembers(ctx, nil, groupID)
	if err != nil {
		return GroupView{}, httpapi.InternalError(err)
	}
	return groupDetailView(group, members, userID), nil
}

func (s *Service) AddGroupMembers(ctx context.Context, userID, groupID int64, rawMemberIDs []int64) (GroupView, error) {
	group, err := s.requireGroupOwner(ctx, groupID, userID)
	if err != nil {
		return GroupView{}, err
	}
	memberIDs, err := normalizeMemberIDs(userID, rawMemberIDs)
	if err != nil || len(memberIDs) == 0 {
		return GroupView{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "请选择要加入的好友")
	}
	existing, err := s.store.ListChatGroupMembers(ctx, nil, groupID)
	if err != nil {
		return GroupView{}, httpapi.InternalError(err)
	}
	if len(existing)+len(memberIDs) > maxGroupMemberCount {
		return GroupView{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "群聊人数已达上限")
	}
	if err := s.validateFriends(ctx, userID, memberIDs); err != nil {
		return GroupView{}, err
	}
	now := time.Now()
	for _, memberID := range memberIDs {
		if err := s.store.UpsertChatGroupMember(ctx, nil, groupID, memberID, "member", now); err != nil {
			return GroupView{}, httpapi.InternalError(err)
		}
	}
	members, err := s.store.ListChatGroupMembers(ctx, nil, groupID)
	if err != nil {
		return GroupView{}, httpapi.InternalError(err)
	}
	return groupDetailView(group, members, userID), nil
}

func (s *Service) RemoveGroupMember(ctx context.Context, userID, groupID, memberID int64) error {
	group, err := s.requireGroupMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if memberID != userID && group.OwnerID != userID {
		return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "仅群主可以移除成员")
	}
	if memberID == group.OwnerID {
		return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "群主暂时不能退出群聊")
	}
	if _, err := s.store.GetChatGroupMember(ctx, nil, groupID, memberID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "群成员不存在")
		}
		return httpapi.InternalError(err)
	}
	if err := s.store.RemoveChatGroupMember(ctx, nil, groupID, memberID, time.Now()); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) ListGroupMessages(ctx context.Context, userID, groupID, beforeID int64, limit int) (GroupMessagesResult, error) {
	if _, err := s.requireGroupMember(ctx, groupID, userID); err != nil {
		return GroupMessagesResult{}, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	messages, err := s.store.ListChatGroupMessages(ctx, nil, groupID, beforeID, limit+1)
	if err != nil {
		return GroupMessagesResult{}, httpapi.InternalError(err)
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	items := make([]GroupMessageView, 0, len(messages))
	for _, message := range messages {
		items = append(items, ToGroupMessageView(message))
	}
	return GroupMessagesResult{Items: items, HasMore: hasMore}, nil
}

func (s *Service) SendGroupMessage(ctx context.Context, userID, groupID int64, content, contentType string) (store.ChatGroupMessage, []int64, error) {
	if _, err := s.requireGroupMember(ctx, groupID, userID); err != nil {
		return store.ChatGroupMessage{}, nil, err
	}
	content = strings.TrimSpace(content)
	if content == "" || utf8.RuneCountInString(content) > maxChatMessageLength {
		return store.ChatGroupMessage{}, nil, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "消息内容不符合要求")
	}
	if contentType == "" {
		contentType = "text"
	}
	if contentType != "text" {
		return store.ChatGroupMessage{}, nil, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "暂不支持该消息类型")
	}
	message, err := s.store.CreateChatGroupMessage(ctx, nil, groupID, userID, content, contentType, time.Now())
	if err != nil {
		return store.ChatGroupMessage{}, nil, httpapi.InternalError(err)
	}
	members, err := s.store.ListChatGroupMembers(ctx, nil, groupID)
	if err != nil {
		return store.ChatGroupMessage{}, nil, httpapi.InternalError(err)
	}
	memberIDs := make([]int64, 0, len(members))
	for _, member := range members {
		memberIDs = append(memberIDs, member.UserID)
	}
	return message, memberIDs, nil
}

func (s *Service) MarkGroupRead(ctx context.Context, userID, groupID int64) error {
	if _, err := s.requireGroupMember(ctx, groupID, userID); err != nil {
		return err
	}
	if err := s.store.MarkChatGroupRead(ctx, nil, groupID, userID, time.Now()); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func ToGroupMessageView(message store.ChatGroupMessage) GroupMessageView {
	return GroupMessageView{
		ID:             message.ID,
		GroupID:        message.GroupID,
		SenderID:       message.SenderID,
		SenderNickname: message.SenderNickname,
		Content:        message.Content,
		ContentType:    message.ContentType,
		CreatedAt:      message.CreatedAt.Format(time.RFC3339),
	}
}

func (s *Service) requireGroupMember(ctx context.Context, groupID, userID int64) (store.ChatGroup, error) {
	group, err := s.store.GetChatGroup(ctx, nil, groupID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.ChatGroup{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "群聊不存在")
		}
		return store.ChatGroup{}, httpapi.InternalError(err)
	}
	if _, err := s.store.GetChatGroupMember(ctx, nil, groupID, userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.ChatGroup{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权访问该群聊")
		}
		return store.ChatGroup{}, httpapi.InternalError(err)
	}
	return group, nil
}

func (s *Service) requireGroupOwner(ctx context.Context, groupID, userID int64) (store.ChatGroup, error) {
	group, err := s.requireGroupMember(ctx, groupID, userID)
	if err != nil {
		return store.ChatGroup{}, err
	}
	if group.OwnerID != userID {
		return store.ChatGroup{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "仅群主可以进行此操作")
	}
	return group, nil
}

func (s *Service) validateFriends(ctx context.Context, ownerID int64, memberIDs []int64) error {
	for _, memberID := range memberIDs {
		friendship, err := s.store.GetFriendshipBetween(ctx, nil, ownerID, memberID)
		if err != nil || friendship.Status != "accepted" {
			if err != nil && !errors.Is(err, store.ErrNotFound) {
				return httpapi.InternalError(err)
			}
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "只能邀请已成为好友的同学")
		}
	}
	return nil
}

func normalizeGroupName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > maxGroupNameLength {
		return "", httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "群名称需为 2 到 40 个字符")
	}
	return name, nil
}

func normalizeMemberIDs(ownerID int64, raw []int64) ([]int64, error) {
	seen := make(map[int64]struct{}, len(raw))
	members := make([]int64, 0, len(raw))
	for _, memberID := range raw {
		if memberID <= 0 {
			return nil, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "群成员参数错误")
		}
		if memberID == ownerID {
			continue
		}
		if _, ok := seen[memberID]; ok {
			continue
		}
		seen[memberID] = struct{}{}
		members = append(members, memberID)
	}
	if len(members)+1 > maxGroupMemberCount {
		return nil, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "群聊人数已达上限")
	}
	sort.Slice(members, func(i, j int) bool { return members[i] < members[j] })
	return members, nil
}

func groupSummaryView(group store.ChatGroupSummary, viewerID int64) GroupView {
	view := GroupView{
		ID:                group.ID,
		Name:              group.Name,
		OwnerID:           group.OwnerID,
		IsOwner:           group.OwnerID == viewerID,
		MemberCount:       group.MemberCount,
		Members:           []GroupMemberView{},
		LastMessage:       group.LastMessage,
		LastMessageSender: group.LastMessageSender,
		UnreadCount:       group.UnreadCount,
		CreatedAt:         group.CreatedAt.Format(time.RFC3339),
	}
	if group.LastMessageAt.Valid {
		view.LastMessageAt = group.LastMessageAt.Time.Format(time.RFC3339)
	}
	return view
}

func groupDetailView(group store.ChatGroup, members []store.ChatGroupMember, viewerID int64) GroupView {
	memberViews := make([]GroupMemberView, 0, len(members))
	for _, member := range members {
		memberViews = append(memberViews, GroupMemberView{
			UserID:     member.UserID,
			Nickname:   member.Nickname,
			Department: member.Department,
			Grade:      member.Grade,
			Role:       member.Role,
			IsMe:       member.UserID == viewerID,
			JoinedAt:   member.JoinedAt.Format(time.RFC3339),
		})
	}
	return GroupView{
		ID:          group.ID,
		Name:        group.Name,
		OwnerID:     group.OwnerID,
		IsOwner:     group.OwnerID == viewerID,
		MemberCount: len(memberViews),
		Members:     memberViews,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
	}
}
