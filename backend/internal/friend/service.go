package friend

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/store"
)

type NotifyFunc func(userID int64, eventType string, data any)

type Service struct {
	store  *store.Store
	notify NotifyFunc
}

type UserView struct {
	UserID     int64  `json:"userId"`
	Nickname   string `json:"nickname"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
}

type FriendView struct {
	FriendshipID int64    `json:"friendshipId"`
	User         UserView `json:"user"`
	CreatedAt    string   `json:"createdAt"`
}

type RequestView struct {
	FriendshipID int64    `json:"friendshipId"`
	User         UserView `json:"user"`
	CreatedAt    string   `json:"createdAt"`
}

type AcceptResult struct {
	FriendshipID   int64 `json:"friendshipId"`
	ConversationID int64 `json:"conversationId"`
}

func NewService(dataStore *store.Store, notify NotifyFunc) *Service {
	return &Service{store: dataStore, notify: notify}
}

func (s *Service) SendRequest(ctx context.Context, requesterID int64, nickname string) (store.Friendship, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return store.Friendship{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}

	addressee, err := s.store.GetUserByNickname(ctx, nil, nickname)
	if err != nil {
		if err == store.ErrNotFound {
			return store.Friendship{}, httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "用户不存在")
		}
		return store.Friendship{}, httpapi.InternalError(err)
	}
	if requesterID == addressee.ID {
		return store.Friendship{}, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "不能添加自己为好友")
	}

	existing, err := s.store.GetFriendshipBetween(ctx, nil, requesterID, addressee.ID)
	if err == nil {
		switch existing.Status {
		case "accepted":
			return store.Friendship{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "你们已经是好友了")
		case "pending":
			return store.Friendship{}, httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "已有待处理的好友请求")
		case "blocked":
			return store.Friendship{}, httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无法发送好友请求")
		}
	} else if err != store.ErrNotFound {
		return store.Friendship{}, httpapi.InternalError(err)
	}

	now := time.Now()
	friendship, err := s.store.CreateFriendship(ctx, nil, requesterID, addressee.ID, now)
	if err != nil {
		return store.Friendship{}, httpapi.InternalError(err)
	}

	if s.notify != nil {
		s.notify(addressee.ID, "friend_request", map[string]any{
			"friendshipId": friendship.ID,
			"fromUserId":   requesterID,
		})
	}

	return friendship, nil
}

func (s *Service) AcceptRequest(ctx context.Context, currentUserID, friendshipID int64) (AcceptResult, error) {
	var result AcceptResult
	err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
		f, err := s.store.GetFriendship(ctx, tx, friendshipID)
		if err != nil {
			if err == store.ErrNotFound {
				return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "好友请求不存在")
			}
			return httpapi.InternalError(err)
		}
		if f.AddresseeID != currentUserID {
			return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权操作")
		}
		if f.Status != "pending" {
			return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "该请求已处理")
		}

		now := time.Now()
		if err := s.store.UpdateFriendshipStatus(ctx, tx, friendshipID, "accepted", now); err != nil {
			return httpapi.InternalError(err)
		}

		conv, err := s.store.CreateConversation(ctx, tx, f.RequesterID, f.AddresseeID, now)
		if err != nil {
			return httpapi.InternalError(err)
		}
		result = AcceptResult{FriendshipID: friendshipID, ConversationID: conv.ID}
		return nil
	})
	if err != nil {
		return AcceptResult{}, err
	}

	if s.notify != nil {
		f, fErr := s.store.GetFriendship(ctx, nil, friendshipID)
		if fErr == nil {
			s.notify(f.RequesterID, "friend_accepted", map[string]any{
				"friendshipId":   friendshipID,
				"conversationId": result.ConversationID,
				"byUserId":       currentUserID,
			})
		}
	}

	return result, nil
}

func (s *Service) RejectRequest(ctx context.Context, currentUserID, friendshipID int64) error {
	f, err := s.store.GetFriendship(ctx, nil, friendshipID)
	if err != nil {
		if err == store.ErrNotFound {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "好友请求不存在")
		}
		return httpapi.InternalError(err)
	}
	if f.AddresseeID != currentUserID {
		return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权操作")
	}
	if f.Status != "pending" {
		return httpapi.NewError(http.StatusConflict, httpapi.CodeConflict, "该请求已处理")
	}
	if err := s.store.UpdateFriendshipStatus(ctx, nil, friendshipID, "rejected", time.Now()); err != nil {
		return httpapi.InternalError(err)
	}
	return nil
}

func (s *Service) ListFriends(ctx context.Context, userID int64) ([]FriendView, error) {
	friends, err := s.store.ListAcceptedFriends(ctx, nil, userID)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	views := make([]FriendView, 0, len(friends))
	for _, f := range friends {
		peerID := f.AddresseeID
		if peerID == userID {
			peerID = f.RequesterID
		}
		views = append(views, FriendView{
			FriendshipID: f.ID,
			User: UserView{
				UserID:     peerID,
				Nickname:   f.PeerNickname,
				Email:      f.PeerEmail,
				Department: f.PeerDepartment,
				Grade:      f.PeerGrade,
			},
			CreatedAt: f.UpdatedAt.Format(time.RFC3339),
		})
	}
	return views, nil
}

func (s *Service) ListReceivedRequests(ctx context.Context, userID int64) ([]RequestView, error) {
	reqs, err := s.store.ListReceivedRequests(ctx, nil, userID)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return toRequestViews(reqs, func(f store.FriendshipWithUser) int64 { return f.RequesterID }), nil
}

func (s *Service) ListSentRequests(ctx context.Context, userID int64) ([]RequestView, error) {
	reqs, err := s.store.ListSentRequests(ctx, nil, userID)
	if err != nil {
		return nil, httpapi.InternalError(err)
	}
	return toRequestViews(reqs, func(f store.FriendshipWithUser) int64 { return f.AddresseeID }), nil
}

func (s *Service) RemoveFriend(ctx context.Context, currentUserID, friendshipID int64) error {
	f, err := s.store.GetFriendship(ctx, nil, friendshipID)
	if err != nil {
		if err == store.ErrNotFound {
			return httpapi.NewError(http.StatusNotFound, httpapi.CodeUserNotFound, "好友关系不存在")
		}
		return httpapi.InternalError(err)
	}
	if f.RequesterID != currentUserID && f.AddresseeID != currentUserID {
		return httpapi.NewError(http.StatusForbidden, httpapi.CodeForbidden, "无权操作")
	}

	peerUserID := f.RequesterID
	if peerUserID == currentUserID {
		peerUserID = f.AddresseeID
	}

	now := time.Now()
	disabledConversation, err := func() (store.Conversation, error) {
		var conversation store.Conversation
		err := s.store.WithTx(ctx, func(tx *sql.Tx) error {
			updated, err := s.store.DisableConversationBetween(ctx, tx, f.RequesterID, f.AddresseeID, currentUserID, now)
			if err == nil {
				conversation = updated
			} else if err != store.ErrNotFound {
				return httpapi.InternalError(err)
			}
			if err := s.store.DeleteFriendship(ctx, tx, friendshipID); err != nil {
				return httpapi.InternalError(err)
			}
			return nil
		})
		return conversation, err
	}()
	if err != nil {
		return err
	}

	if s.notify != nil && disabledConversation.ID > 0 {
		payload := map[string]any{
			"conversationId":   disabledConversation.ID,
			"disabledByUserId": currentUserID,
		}
		s.notify(currentUserID, "conversation_updated", payload)
		s.notify(peerUserID, "conversation_updated", payload)
	}

	return nil
}

func toRequestViews(items []store.FriendshipWithUser, peerIDFn func(store.FriendshipWithUser) int64) []RequestView {
	views := make([]RequestView, 0, len(items))
	for _, f := range items {
		views = append(views, RequestView{
			FriendshipID: f.ID,
			User: UserView{
				UserID:     peerIDFn(f),
				Nickname:   f.PeerNickname,
				Email:      f.PeerEmail,
				Department: f.PeerDepartment,
				Grade:      f.PeerGrade,
			},
			CreatedAt: f.CreatedAt.Format(time.RFC3339),
		})
	}
	return views
}
