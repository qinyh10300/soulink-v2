package connection

import "time"

const (
	CategoryMeal         = "meal"
	CategoryStudy        = "study"
	CategorySport        = "sport"
	CategoryEvent        = "event"
	CategoryConversation = "conversation"
	CategoryRomantic     = "romantic"
	CategoryProject      = "project"
	CategoryCofounder    = "cofounder"
	CategoryOther        = "other"
)

const (
	VisibilityPrivate    = "private"
	VisibilityCandidates = "candidates"
	VisibilityFeed       = "feed"
)

const (
	StatusDraft     = "draft"
	StatusActive    = "active"
	StatusPaused    = "paused"
	StatusFulfilled = "fulfilled"
	StatusExpired   = "expired"
)

type MatchingProfileInput struct {
	Bio                string   `json:"bio"`
	Skills             []string `json:"skills"`
	Interests          []string `json:"interests"`
	Availability       string   `json:"availability"`
	PreferredScenes    []string `json:"preferredScenes"`
	CollaborationStyle string   `json:"collaborationStyle"`
	Boundaries         string   `json:"boundaries"`
	AIPortrait         string   `json:"aiPortrait"`
	AIPortraitApproved bool     `json:"aiPortraitApproved"`
	RomanticOptIn      bool     `json:"romanticOptIn"`
	MatchingPaused     bool     `json:"matchingPaused"`
}

type MatchingProfileView struct {
	MatchingProfileInput
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type ProjectMemo struct {
	Building         string   `json:"building"`
	Stage            string   `json:"stage"`
	WeeklyTime       string   `json:"weeklyTime"`
	Compensation     string   `json:"compensation"`
	RequiredSkills   []string `json:"requiredSkills"`
	NiceToHaveSkills []string `json:"niceToHaveSkills"`
	WorkingStyle     string   `json:"workingStyle"`
	RiskAppetite     string   `json:"riskAppetite"`
}

type IntentInput struct {
	Category        string       `json:"category"`
	Title           string       `json:"title"`
	Description     string       `json:"description"`
	Goal            string       `json:"goal"`
	TimeWindow      string       `json:"timeWindow"`
	Location        string       `json:"location"`
	Visibility      string       `json:"visibility"`
	Status          string       `json:"status"`
	Constraints     []string     `json:"constraints"`
	DealBreakers    []string     `json:"dealBreakers"`
	DesiredNextStep string       `json:"desiredNextStep"`
	ProjectMemo     *ProjectMemo `json:"projectMemo,omitempty"`
	ExpiresAt       *time.Time   `json:"expiresAt"`
}

type AuthorPreview struct {
	Nickname   string `json:"nickname"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
	Verified   bool   `json:"verified"`
}

type IntentView struct {
	ID              int64         `json:"id"`
	Author          AuthorPreview `json:"author"`
	Category        string        `json:"category"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	Goal            string        `json:"goal"`
	TimeWindow      string        `json:"timeWindow"`
	Location        string        `json:"location"`
	Visibility      string        `json:"visibility"`
	Status          string        `json:"status"`
	Constraints     []string      `json:"constraints"`
	DealBreakers    []string      `json:"dealBreakers,omitempty"`
	DesiredNextStep string        `json:"desiredNextStep"`
	ProjectMemo     *ProjectMemo  `json:"projectMemo,omitempty"`
	ExpiresAt       *time.Time    `json:"expiresAt,omitempty"`
	PublishedAt     *time.Time    `json:"publishedAt,omitempty"`
	InterestCount   int           `json:"interestCount"`
	InterestState   string        `json:"interestState"`
	SavedByMe       bool          `json:"savedByMe"`
	IsOwner         bool          `json:"isOwner"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

type RecommendationView struct {
	Intent        IntentView `json:"intent"`
	Score         int        `json:"score"`
	FitReason     string     `json:"fitReason"`
	Uncertainty   string     `json:"uncertainty"`
	SharedContext []string   `json:"sharedContext"`
}

type InterestView struct {
	UserID             int64     `json:"userId"`
	Nickname           string    `json:"nickname"`
	Department         string    `json:"department"`
	Grade              string    `json:"grade"`
	Bio                string    `json:"bio"`
	Skills             []string  `json:"skills"`
	CollaborationStyle string    `json:"collaborationStyle"`
	Status             string    `json:"status"`
	Message            string    `json:"message"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type ContactView struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type RoomParticipantView struct {
	UserID     int64        `json:"userId"`
	Nickname   string       `json:"nickname"`
	Department string       `json:"department"`
	Grade      string       `json:"grade"`
	Contact    *ContactView `json:"contact,omitempty"`
}

type MatchRoomView struct {
	ID                   int64               `json:"id"`
	IntentID             int64               `json:"intentId"`
	IntentTitle          string              `json:"intentTitle"`
	IntentCategory       string              `json:"intentCategory"`
	Me                   RoomParticipantView `json:"me"`
	Other                RoomParticipantView `json:"other"`
	Status               string              `json:"status"`
	FitReason            string              `json:"fitReason"`
	Uncertainty          string              `json:"uncertainty"`
	SharedContext        []string            `json:"sharedContext"`
	ConversationStarters []string            `json:"conversationStarters"`
	SuggestedPlaces      []string            `json:"suggestedPlaces"`
	ScheduledAt          *time.Time          `json:"scheduledAt,omitempty"`
	SelectedPlace        string              `json:"selectedPlace"`
	MyContactShared      bool                `json:"myContactShared"`
	OtherContactShared   bool                `json:"otherContactShared"`
	CreatedAt            time.Time           `json:"createdAt"`
	UpdatedAt            time.Time           `json:"updatedAt"`
}

type MatchRoomMessageView struct {
	ID             int64     `json:"id"`
	RoomID         int64     `json:"roomId"`
	SenderID       int64     `json:"senderId"`
	SenderNickname string    `json:"senderNickname"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
}
