package safety

import "time"

const (
	ContextIntent    = "intent"
	ContextMatchRoom = "match_room"
)

type ReportInput struct {
	ContextType string
	ContextID   int64
	Category    string
	Details     string
	BlockUser   bool
}

type ReportResult struct {
	ID        int64     `json:"id"`
	TicketID  string    `json:"ticketId"`
	Blocked   bool      `json:"blocked"`
	CreatedAt time.Time `json:"createdAt"`
}

type Capabilities struct {
	IsAdmin bool `json:"isAdmin"`
}

type ReportView struct {
	ID                   int64      `json:"id"`
	TicketID             string     `json:"ticketId"`
	ReporterID           int64      `json:"reporterId"`
	ReporterNickname     string     `json:"reporterNickname"`
	ReporterEmail        string     `json:"reporterEmail"`
	ReportedUserID       int64      `json:"reportedUserId"`
	ReportedUserNickname string     `json:"reportedUserNickname"`
	ReportedUserEmail    string     `json:"reportedUserEmail"`
	ContextType          string     `json:"contextType"`
	ContextID            int64      `json:"contextId"`
	ContextTitle         string     `json:"contextTitle"`
	Category             string     `json:"category"`
	Details              string     `json:"details"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"createdAt"`
	ResolvedAt           *time.Time `json:"resolvedAt,omitempty"`
	ResolvedBy           *int64     `json:"resolvedBy,omitempty"`
}
