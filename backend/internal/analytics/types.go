package analytics

import "time"

const (
	EventIntentCreated            = "intent_created"
	EventIntentPublished          = "intent_published"
	EventRecommendationsViewed    = "recommendations_viewed"
	EventInterestExpressed        = "interest_expressed"
	EventInterestAccepted         = "interest_accepted"
	EventMatchCreated             = "match_created"
	EventMeetupProposed           = "meetup_proposed"
	EventMeetupConfirmed          = "meetup_confirmed"
	EventMeetupCompleted          = "meetup_completed"
	EventMeetupNoShow             = "meetup_no_show"
	EventMeetingFeedbackSubmitted = "meeting_feedback_submitted"
	EventPlazaGroupCreated        = "plaza_group_created"
	EventGroupMatchCreated        = "group_match_created"
	EventSafetyReportSubmitted    = "safety_report_submitted"
	EventUserBlocked              = "user_blocked"
)

type EventInput struct {
	Name        string
	ActorUserID int64
	SubjectType string
	SubjectID   int64
	RequestID   string
	Properties  map[string]any
	OccurredAt  time.Time
}

type Period struct {
	Days int       `json:"days"`
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type NorthStarMetrics struct {
	MeetupsCompleted  int64   `json:"meetupsCompleted"`
	MeetupsConfirmed  int64   `json:"meetupsConfirmed"`
	MatchToMeetupRate float64 `json:"matchToMeetupRate"`
	RepeatIntentRate  float64 `json:"repeatIntentRate"`
}

type FunnelStep struct {
	Key                    string  `json:"key"`
	Label                  string  `json:"label"`
	Count                  int64   `json:"count"`
	ConversionFromPrevious float64 `json:"conversionFromPrevious"`
}

type ActivityMetrics struct {
	RegisteredUsers   int64 `json:"registeredUsers"`
	ActiveUsers       int64 `json:"activeUsers"`
	IntentCreators    int64 `json:"intentCreators"`
	RepeatIntentUsers int64 `json:"repeatIntentUsers"`
}

type SafetyMetrics struct {
	OpenReports int64   `json:"openReports"`
	ReportRate  float64 `json:"reportRate"`
}

type EventMetric struct {
	EventName    string `json:"eventName"`
	Count        int64  `json:"count"`
	UniqueActors int64  `json:"uniqueActors"`
}

type DailyMetric struct {
	Date                  string `json:"date"`
	ActiveUsers           int64  `json:"activeUsers"`
	IntentsCreated        int64  `json:"intentsCreated"`
	RecommendationsViewed int64  `json:"recommendationsViewed"`
	InterestsExpressed    int64  `json:"interestsExpressed"`
	MatchesCreated        int64  `json:"matchesCreated"`
	MeetupsConfirmed      int64  `json:"meetupsConfirmed"`
	MeetupsCompleted      int64  `json:"meetupsCompleted"`
	SafetyReports         int64  `json:"safetyReports"`
}

type Dashboard struct {
	Period      Period           `json:"period"`
	NorthStar   NorthStarMetrics `json:"northStar"`
	Activity    ActivityMetrics  `json:"activity"`
	Safety      SafetyMetrics    `json:"safety"`
	Funnel      []FunnelStep     `json:"funnel"`
	Events      []EventMetric    `json:"events"`
	Daily       []DailyMetric    `json:"daily"`
	GeneratedAt time.Time        `json:"generatedAt"`
}
