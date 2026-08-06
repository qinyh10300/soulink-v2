package safety

import "time"

const (
	ContextIntent    = "intent"
	ContextMatchRoom = "match_room"
	ContextDirect    = "direct"
	ContextGroup     = "group"
)

type ReportInput struct {
	ContextType    string
	ContextID      int64
	Category       string
	Details        string
	BlockUser      bool
	ReportedUserID int64
}

type ReportUpdateInput struct {
	Status      string
	ReviewNotes string
	AssignToMe  bool
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

type ReportAppealView struct {
	ID          int64     `json:"id"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	ReviewNotes string    `json:"reviewNotes,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SafetyCaseView struct {
	ID           int64             `json:"id"`
	TicketID     string            `json:"ticketId"`
	Category     string            `json:"category"`
	ContextType  string            `json:"contextType"`
	ContextTitle string            `json:"contextTitle"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
	Appeal       *ReportAppealView `json:"appeal,omitempty"`
}

type ReportAppealInput struct {
	Reason string
}

type ReportAppealReviewInput struct {
	Status      string
	ReviewNotes string
}

type ReportView struct {
	ID                   int64             `json:"id"`
	TicketID             string            `json:"ticketId"`
	ReporterID           int64             `json:"reporterId"`
	ReporterNickname     string            `json:"reporterNickname"`
	ReporterEmail        string            `json:"reporterEmail"`
	ReportedUserID       int64             `json:"reportedUserId"`
	ReportedUserNickname string            `json:"reportedUserNickname"`
	ReportedUserEmail    string            `json:"reportedUserEmail"`
	ContextType          string            `json:"contextType"`
	ContextID            int64             `json:"contextId"`
	ContextTitle         string            `json:"contextTitle"`
	Category             string            `json:"category"`
	Details              string            `json:"details"`
	Status               string            `json:"status"`
	AssignedTo           *int64            `json:"assignedTo,omitempty"`
	AssignedToNickname   string            `json:"assignedToNickname,omitempty"`
	ReviewNotes          string            `json:"reviewNotes"`
	ContextSnapshot      any               `json:"contextSnapshot"`
	RiskSignals          []string          `json:"riskSignals"`
	Appeal               *ReportAppealView `json:"appeal,omitempty"`
	CreatedAt            time.Time         `json:"createdAt"`
	UpdatedAt            time.Time         `json:"updatedAt"`
	ResolvedAt           *time.Time        `json:"resolvedAt,omitempty"`
	ResolvedBy           *int64            `json:"resolvedBy,omitempty"`
}
