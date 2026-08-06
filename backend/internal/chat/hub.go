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
	clients map[int64]map[*Client]struct{}
	logger  *slog.Logger
}

func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.Default()
	}
	return &Hub{
		clients: make(map[int64]map[*Client]struct{}),
		logger:  logger,
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	connections := h.clients[client.userID]
	if connections == nil {
		connections = make(map[*Client]struct{})
		h.clients[client.userID] = connections
	}
	connections[client] = struct{}{}
	connectionCount := len(connections)
	h.mu.Unlock()
	h.logger.Info("ws client registered", "user_id", client.userID, "connection_count", connectionCount)
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	connections := h.clients[client.userID]
	if _, ok := connections[client]; !ok {
		h.mu.Unlock()
		return
	}
	delete(connections, client)
	connectionCount := len(connections)
	if connectionCount == 0 {
		delete(h.clients, client.userID)
	}
	h.mu.Unlock()
	h.logger.Info("ws client unregistered", "user_id", client.userID, "connection_count", connectionCount)
}

// SendToUser sends an envelope to a specific online user. Returns false if user is offline.
func (h *Hub) SendToUser(userID int64, env WSEnvelope) bool {
	h.mu.RLock()
	connections := h.clients[userID]
	clients := make([]*Client, 0, len(connections))
	for client := range connections {
		clients = append(clients, client)
	}
	h.mu.RUnlock()
	if len(clients) == 0 {
		return false
	}
	for _, client := range clients {
		client.Send(env)
	}
	return true
}

// BroadcastToUsers sends an envelope to multiple users.
func (h *Hub) BroadcastToUsers(userIDs []int64, env WSEnvelope) {
	for _, uid := range userIDs {
		h.SendToUser(uid, env)
	}
}
