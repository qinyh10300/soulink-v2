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

const (
	RoomStatusOpen   = "open"
	RoomStatusMet    = "met"
	RoomStatusClosed = "closed"
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
	HiddenCategories   []string `json:"hiddenCategories"`
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
	UserID     int64  `json:"userId,omitempty"`
	Nickname   string `json:"nickname"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
	Verified   bool   `json:"verified"`
}

type ContextProfileIntent struct {
	ID              int64  `json:"id"`
	Category        string `json:"category"`
	Title           string `json:"title"`
	Goal            string `json:"goal"`
	TimeWindow      string `json:"timeWindow"`
	Location        string `json:"location"`
	DesiredNextStep string `json:"desiredNextStep"`
}

type ContextProfileView struct {
	UserID             int64                `json:"userId"`
	Nickname           string               `json:"nickname"`
	SchoolName         string               `json:"schoolName"`
	Department         string               `json:"department,omitempty"`
	Grade              string               `json:"grade,omitempty"`
	Verified           bool                 `json:"verified"`
	Bio                string               `json:"bio,omitempty"`
	Skills             []string             `json:"skills,omitempty"`
	Interests          []string             `json:"interests,omitempty"`
	Availability       string               `json:"availability,omitempty"`
	PreferredScenes    []string             `json:"preferredScenes,omitempty"`
	CollaborationStyle string               `json:"collaborationStyle,omitempty"`
	VisibleFields      []string             `json:"visibleFields"`
	Intent             ContextProfileIntent `json:"intent"`
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
	Intent        IntentView    `json:"intent"`
	Score         int           `json:"score"`
	FitReason     string        `json:"fitReason"`
	Uncertainty   string        `json:"uncertainty"`
	SharedContext []string      `json:"sharedContext"`
	Inference     InferenceView `json:"inference"`
}

type InferenceView struct {
	Source        string  `json:"source"`
	Model         string  `json:"model,omitempty"`
	Confidence    float64 `json:"confidence"`
	PromptVersion string  `json:"promptVersion"`
	Editable      bool    `json:"editable"`
}

type DailyRecommendationsView struct {
	SourceIntentID int64                `json:"sourceIntentId,omitempty"`
	GeneratedAt    *time.Time           `json:"generatedAt,omitempty"`
	RefreshAt      *time.Time           `json:"refreshAt,omitempty"`
	Algorithm      string               `json:"algorithm"`
	Items          []RecommendationView `json:"items"`
}

type RecommendationPreferenceInput struct {
	LearningEnabled bool `json:"learningEnabled"`
}

type RecommendationAffinityView struct {
	Category string `json:"category"`
	Weight   int    `json:"weight"`
	Signals  int    `json:"signals"`
	Summary  string `json:"summary"`
}

type RecommendationSignalView struct {
	Type       string    `json:"type"`
	Category   string    `json:"category"`
	Impact     int       `json:"impact"`
	Source     string    `json:"source"`
	OccurredAt time.Time `json:"occurredAt"`
}

type RecommendationPreferencesView struct {
	LearningEnabled bool                         `json:"learningEnabled"`
	Affinities      []RecommendationAffinityView `json:"affinities"`
	RecentSignals   []RecommendationSignalView   `json:"recentSignals"`
	LearningResetAt *time.Time                   `json:"learningResetAt,omitempty"`
	UpdatedAt       *time.Time                   `json:"updatedAt,omitempty"`
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
	Assistant            InferenceView       `json:"assistant"`
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
