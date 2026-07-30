package chat

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// WSEnvelope is the JSON frame exchanged over WebSocket.
type WSEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Hub manages active WebSocket clients and routes messages between them.
type Hub struct {
	mu      sync.RWMutex
	clients map[int64]*Client // userID -> Client
	logger  *slog.Logger
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[int64]*Client),
		logger:  logger,
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.clients[client.userID]; ok {
		old.Close()
	}
	h.clients[client.userID] = client
	h.logger.Info("ws client registered", "userId", client.userID)
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if existing, ok := h.clients[client.userID]; ok && existing == client {
		delete(h.clients, client.userID)
		h.logger.Info("ws client unregistered", "userId", client.userID)
	}
}

// SendToUser sends an envelope to a specific online user. Returns false if user is offline.
func (h *Hub) SendToUser(userID int64, env WSEnvelope) bool {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	client.Send(env)
	return true
}

// BroadcastToUsers sends an envelope to multiple users.
func (h *Hub) BroadcastToUsers(userIDs []int64, env WSEnvelope) {
	for _, uid := range userIDs {
		h.SendToUser(uid, env)
	}
}
