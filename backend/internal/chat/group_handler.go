package chat

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"soulink-backend/internal/auth"
	"soulink-backend/internal/httpapi"
)

type createConversationRequest struct {
	PeerUserID int64 `json:"peerUserId"`
}

type createGroupRequest struct {
	Name      string  `json:"name"`
	MemberIDs []int64 `json:"memberIds"`
}

type updateGroupRequest struct {
	Name string `json:"name"`
}

type addGroupMembersRequest struct {
	MemberIDs []int64 `json:"memberIds"`
}

type createChatMessageRequest struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	var request createConversationRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	conversation, err := h.service.EnsureConversation(r.Context(), user.ID, request.PeerUserID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "conversation ready", conversation)
}

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	conversationID, err := parseChatRouteID(r, "conversationId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var request createChatMessageRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	message, err := h.service.SendMessage(r.Context(), user.ID, conversationID, request.Content, request.ContentType)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	view := ToMessageView(message)
	h.broadcastDirectMessage(r, user.ID, conversationID, view)
	httpapi.Respond(w, http.StatusCreated, "message sent", view)
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	groups, err := h.service.ListGroups(r.Context(), user.ID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, groups)
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	var request createGroupRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	group, err := h.service.CreateGroup(r.Context(), user.ID, CreateGroupInput{
		Name:      request.Name,
		MemberIDs: request.MemberIDs,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	h.broadcastGroupUpdate(group)
	httpapi.Respond(w, http.StatusCreated, "group created", group)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := parseChatRouteID(r, "groupId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	group, err := h.service.GetGroup(r.Context(), user.ID, groupID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, group)
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := parseChatRouteID(r, "groupId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var request updateGroupRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	group, err := h.service.RenameGroup(r.Context(), user.ID, groupID, request.Name)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	h.broadcastGroupUpdate(group)
	httpapi.OK(w, group)
}

func (h *Handler) AddGroupMembers(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := parseChatRouteID(r, "groupId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var request addGroupMembersRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	group, err := h.service.AddGroupMembers(r.Context(), user.ID, groupID, request.MemberIDs)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	h.broadcastGroupUpdate(group)
	httpapi.OK(w, group)
}

func (h *Handler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := parseChatRouteID(r, "groupId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	memberID, err := parseChatRouteID(r, "userId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	if err := h.service.RemoveGroupMember(r.Context(), user.ID, groupID, memberID); err != nil {
		httpapi.Error(w, err)
		return
	}
	h.sendGroupUpdated(memberID, groupID)
	if group, err := h.service.GetGroup(r.Context(), user.ID, groupID); err == nil {
		h.broadcastGroupUpdate(group)
	}
	httpapi.OK(w, nil)
}

func (h *Handler) ListGroupMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := parseChatRouteID(r, "groupId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	beforeID, limit, err := parseMessagePagination(r)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	result, err := h.service.ListGroupMessages(r.Context(), user.ID, groupID, beforeID, limit)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, result)
}

func (h *Handler) CreateGroupMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := parseChatRouteID(r, "groupId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	var request createChatMessageRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	message, memberIDs, err := h.service.SendGroupMessage(r.Context(), user.ID, groupID, request.Content, request.ContentType)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	view := ToGroupMessageView(message)
	h.broadcastGroupMessage(memberIDs, view)
	httpapi.Respond(w, http.StatusCreated, "message sent", view)
}

func (h *Handler) MarkGroupRead(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusUnauthorized, httpapi.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := parseChatRouteID(r, "groupId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	if err := h.service.MarkGroupRead(r.Context(), user.ID, groupID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}

func (h *Handler) broadcastDirectMessage(r *http.Request, senderID, conversationID int64, view MessageView) {
	data, _ := json.Marshal(view)
	envelope := WSEnvelope{Type: "message", Data: data}
	h.hub.SendToUser(senderID, envelope)
	conversation, err := h.service.store.GetConversation(r.Context(), nil, conversationID)
	if err != nil {
		return
	}
	peerID := conversation.User1ID
	if peerID == senderID {
		peerID = conversation.User2ID
	}
	h.hub.SendToUser(peerID, envelope)
}

func (h *Handler) broadcastGroupMessage(memberIDs []int64, view GroupMessageView) {
	data, _ := json.Marshal(view)
	envelope := WSEnvelope{Type: "group_message", Data: data}
	for _, memberID := range memberIDs {
		h.hub.SendToUser(memberID, envelope)
	}
}

func (h *Handler) broadcastGroupUpdate(group GroupView) {
	for _, member := range group.Members {
		h.sendGroupUpdated(member.UserID, group.ID)
	}
}

func (h *Handler) sendGroupUpdated(userID, groupID int64) {
	data, _ := json.Marshal(map[string]int64{"groupId": groupID})
	h.hub.SendToUser(userID, WSEnvelope{Type: "group_updated", Data: data})
}

func parseChatRouteID(r *http.Request, key string) (int64, error) {
	value, err := strconv.ParseInt(chi.URLParam(r, key), 10, 64)
	if err != nil || value <= 0 {
		return 0, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}
	return value, nil
}

func parseMessagePagination(r *http.Request) (int64, int, error) {
	var beforeID int64
	if raw := r.URL.Query().Get("before"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return 0, 0, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		beforeID = parsed
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return 0, 0, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
		}
		limit = parsed
	}
	return beforeID, limit, nil
}
