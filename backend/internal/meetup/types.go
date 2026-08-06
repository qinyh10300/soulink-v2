package meetup

import "time"

const (
	ContextMatchRoom = "match_room"
	ContextDirect    = "direct"
	ContextGroup     = "group"
)

type AvailabilityWindowInput struct {
	WindowType  string
	Timezone    string
	StartsAt    *time.Time
	EndsAt      *time.Time
	Weekday     int
	StartMinute int
	EndMinute   int
	ValidFrom   *time.Time
	ValidUntil  *time.Time
}

type AvailabilityWindowView struct {
	ID          int64   `json:"id"`
	WindowType  string  `json:"windowType"`
	Timezone    string  `json:"timezone"`
	StartsAt    *string `json:"startsAt,omitempty"`
	EndsAt      *string `json:"endsAt,omitempty"`
	Weekday     int     `json:"weekday,omitempty"`
	StartMinute int     `json:"startMinute,omitempty"`
	EndMinute   int     `json:"endMinute,omitempty"`
	ValidFrom   *string `json:"validFrom,omitempty"`
	ValidUntil  *string `json:"validUntil,omitempty"`
}

type ContextRef struct {
	Type string `json:"type"`
	ID   int64  `json:"id"`
}

type SuggestionInput struct {
	Context         ContextRef
	From            time.Time
	To              time.Time
	DurationMinutes int
}

type TimeSuggestion struct {
	StartsAt    string `json:"startsAt"`
	EndsAt      string `json:"endsAt"`
	Label       string `json:"label"`
	Explanation string `json:"explanation"`
}

type CampusPlace struct {
	Name    string   `json:"name"`
	Area    string   `json:"area"`
	BestFor []string `json:"bestFor"`
}

type SuggestionResult struct {
	ParticipantCount         int              `json:"participantCount"`
	MissingAvailabilityCount int              `json:"missingAvailabilityCount"`
	Candidates               []TimeSuggestion `json:"candidates"`
	Places                   []CampusPlace    `json:"places"`
}

type CreateMeetupInput struct {
	Context  ContextRef
	Title    string
	Category string
	StartsAt time.Time
	EndsAt   time.Time
	Location string
}

type RescheduleInput struct {
	StartsAt        time.Time
	EndsAt          time.Time
	Location        string
	ExpectedVersion int
}

type ParticipantView struct {
	UserID      int64  `json:"userId"`
	Nickname    string `json:"nickname"`
	Department  string `json:"department"`
	Grade       string `json:"grade"`
	Status      string `json:"status"`
	IsMe        bool   `json:"isMe"`
	IsOrganizer bool   `json:"isOrganizer"`
}

type FeedbackInput struct {
	Attendance     string
	Rating         *int
	WouldMeetAgain *bool
	SafetyConcern  bool
	PrivateNotes   string
}

type FeedbackView struct {
	Attendance     string `json:"attendance"`
	Rating         *int   `json:"rating,omitempty"`
	WouldMeetAgain *bool  `json:"wouldMeetAgain,omitempty"`
	SafetyConcern  bool   `json:"safetyConcern"`
	PrivateNotes   string `json:"privateNotes"`
	SubmittedAt    string `json:"submittedAt"`
}

type ReminderView struct {
	Stage string `json:"stage"`
	Text  string `json:"text"`
}

type MeetupView struct {
	ID              int64             `json:"id"`
	Context         ContextRef        `json:"context"`
	OrganizerID     int64             `json:"organizerId"`
	Title           string            `json:"title"`
	Category        string            `json:"category"`
	StartsAt        string            `json:"startsAt"`
	EndsAt          string            `json:"endsAt"`
	Location        string            `json:"location"`
	Status          string            `json:"status"`
	Version         int               `json:"version"`
	CalendarEventID *int64            `json:"calendarEventId,omitempty"`
	Participants    []ParticipantView `json:"participants"`
	MyStatus        string            `json:"myStatus"`
	CanRespond      bool              `json:"canRespond"`
	CanReschedule   bool              `json:"canReschedule"`
	CanCancel       bool              `json:"canCancel"`
	CanComplete     bool              `json:"canComplete"`
	Reminder        ReminderView      `json:"reminder"`
	SuggestedPlaces []CampusPlace     `json:"suggestedPlaces"`
	MyFeedback      *FeedbackView     `json:"myFeedback,omitempty"`
	CreatedAt       string            `json:"createdAt"`
	UpdatedAt       string            `json:"updatedAt"`
}

type interval struct {
	Start time.Time
	End   time.Time
}

type contextDetails struct {
	Ref             ContextRef
	Title           string
	Category        string
	ParticipantIDs  []int64
	SuggestedPlaces []CampusPlace
}
