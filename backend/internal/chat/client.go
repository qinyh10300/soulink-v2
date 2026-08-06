package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
	sendBufLen = 64
)

// Client represents a single WebSocket connection.
type Client struct {
	userID  int64
	conn    *websocket.Conn
	hub     *Hub
	service *Service
	send    chan WSEnvelope
	logger  *slog.Logger
}

func NewClient(userID int64, conn *websocket.Conn, hub *Hub, service *Service, logger *slog.Logger) *Client {
	return &Client{
		userID:  userID,
		conn:    conn,
		hub:     hub,
		service: service,
		send:    make(chan WSEnvelope, sendBufLen),
		logger:  logger,
	}
}

// Run starts the read and write loops. Blocks until connection closes.
func (c *Client) Run(ctx context.Context) {
	c.hub.Register(c)
	defer c.hub.Unregister(c)

	go c.writePump(ctx)
	c.readPump(ctx)
}

func (c *Client) Send(env WSEnvelope) {
	select {
	case c.send <- env:
	default:
		c.logger.Warn("ws send buffer full, dropping message", "userId", c.userID)
	}
}

func (c *Client) Close() {
	c.conn.Close(websocket.StatusGoingAway, "replaced by new connection")
}

func (c *Client) readPump(ctx context.Context) {
	defer c.conn.CloseNow()

	for {
		var env WSEnvelope
		if err := wsjson.Read(ctx, c.conn, &env); err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				c.logger.Info("ws client disconnected normally", "userId", c.userID)
			} else {
				c.logger.Warn("ws read error", "userId", c.userID, "error", err)
			}
			return
		}
		c.handleIncoming(ctx, env)
	}
}

func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case env, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := wsjson.Write(writeCtx, c.conn, env)
			cancel()
			if err != nil {
				c.logger.Warn("ws write error", "userId", c.userID, "error", err)
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				c.logger.Warn("ws ping error", "userId", c.userID, "error", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) handleIncoming(ctx context.Context, env WSEnvelope) {
	switch env.Type {
	case "message":
		c.handleSendMessage(ctx, env.Data)
	case "group_message":
		c.handleSendGroupMessage(ctx, env.Data)
	case "read":
		c.handleMarkRead(ctx, env.Data)
	case "group_read":
		c.handleMarkGroupRead(ctx, env.Data)
	default:
		c.logger.Warn("ws unknown message type", "type", env.Type, "userId", c.userID)
	}
}

func (c *Client) handleSendGroupMessage(ctx context.Context, data json.RawMessage) {
	var payload struct {
		GroupID     int64  `json:"groupId"`
		Content     string `json:"content"`
		ContentType string `json:"contentType"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		c.logger.Warn("ws invalid group message payload", "userId", c.userID, "error", err)
		return
	}

	message, memberIDs, err := c.service.SendGroupMessage(ctx, c.userID, payload.GroupID, payload.Content, payload.ContentType)
	if err != nil {
		c.logger.Warn("ws send group message failed", "userId", c.userID, "error", err)
		return
	}

	messageData, _ := json.Marshal(ToGroupMessageView(message))
	envelope := WSEnvelope{Type: "group_message", Data: messageData}
	for _, memberID := range memberIDs {
		c.hub.SendToUser(memberID, envelope)
	}
}

func (c *Client) handleSendMessage(ctx context.Context, data json.RawMessage) {
	var payload struct {
		ConversationID int64  `json:"conversationId"`
		Content        string `json:"content"`
		ContentType    string `json:"contentType"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		c.logger.Warn("ws invalid message payload", "userId", c.userID, "error", err)
		return
	}

	msg, err := c.service.SendMessage(ctx, c.userID, payload.ConversationID, payload.Content, payload.ContentType)
	if err != nil {
		c.logger.Warn("ws send message failed", "userId", c.userID, "error", err)
		return
	}

	msgView := ToMessageView(msg)
	msgData, _ := json.Marshal(msgView)
	outEnv := WSEnvelope{Type: "message", Data: msgData}

	// Send to the sender (confirmation) and the recipient.
	c.Send(outEnv)
	conv, err := c.service.store.GetConversation(ctx, nil, payload.ConversationID)
	if err == nil {
		peerID := conv.User1ID
		if peerID == c.userID {
			peerID = conv.User2ID
		}
		c.hub.SendToUser(peerID, outEnv)
	}
}

func (c *Client) handleMarkRead(ctx context.Context, data json.RawMessage) {
	var payload struct {
		ConversationID int64 `json:"conversationId"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}
	if err := c.service.MarkRead(ctx, c.userID, payload.ConversationID); err != nil {
		return
	}

	readData, _ := json.Marshal(map[string]any{
		"conversationId": payload.ConversationID,
		"userId":         c.userID,
		"readAt":         time.Now().Format(time.RFC3339),
	})
	outEnv := WSEnvelope{Type: "read", Data: readData}

	conv, err := c.service.store.GetConversation(ctx, nil, payload.ConversationID)
	if err == nil {
		peerID := conv.User1ID
		if peerID == c.userID {
			peerID = conv.User2ID
		}
		c.hub.SendToUser(peerID, outEnv)
	}
}

func (c *Client) handleMarkGroupRead(ctx context.Context, data json.RawMessage) {
	var payload struct {
		GroupID int64 `json:"groupId"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}
	if err := c.service.MarkGroupRead(ctx, c.userID, payload.GroupID); err != nil {
		return
	}

	readData, _ := json.Marshal(map[string]any{
		"groupId": payload.GroupID,
		"userId":  c.userID,
		"readAt":  time.Now().Format(time.RFC3339),
	})
	c.Send(WSEnvelope{Type: "group_read", Data: readData})
}
