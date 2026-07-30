package questionnaire

import "time"

type QuestionType string

type CollectionStatus string

const (
	QuestionTypeSingleChoice   QuestionType = "single_choice"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeNumber         QuestionType = "number"
	QuestionTypeRange          QuestionType = "range"
	QuestionTypeScale          QuestionType = "scale"
	QuestionTypeText           QuestionType = "text"

	CollectionStatusAlwaysOpen CollectionStatus = "always_open"
	CollectionStatusScheduled  CollectionStatus = "scheduled"
	CollectionStatusOpen       CollectionStatus = "open"
	CollectionStatusClosed     CollectionStatus = "closed"
)

type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type SectionMeta struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle,omitempty"`
	Description string `json:"description,omitempty"`
	Tip         string `json:"tip,omitempty"`
}

type Question struct {
	ID                 string       `json:"id"`
	Order              int          `json:"order"`
	Section            string       `json:"section"`
	Title              string       `json:"title"`
	Type               QuestionType `json:"type"`
	Widget             string       `json:"widget,omitempty"`
	Required           bool         `json:"required"`
	Options            []Option     `json:"options,omitempty"`
	Min                *int         `json:"min,omitempty"`
	Max                *int         `json:"max,omitempty"`
	LeftLabel          string       `json:"leftLabel,omitempty"`
	RightLabel         string       `json:"rightLabel,omitempty"`
	MinSelections      *int         `json:"minSelections,omitempty"`
	MaxSelections      *int         `json:"maxSelections,omitempty"`
	MaxLength          *int         `json:"maxLength,omitempty"`
	Placeholder        string       `json:"placeholder,omitempty"`
	AllowUnlimited     bool         `json:"allowUnlimited,omitempty"`
	SupportsImportance bool         `json:"supportsImportance,omitempty"`
}

type Catalog struct {
	ID                 int
	Key                string
	Title              string
	Version            string
	CollectionStartsAt *time.Time
	CollectionEndsAt   *time.Time
	CollectionTimezone string
	Sections           []SectionMeta
	Questions          []Question
	TotalCount         int
	RequiredCount      int
	byID               map[string]Question
}

// QuestionByID 根据稳定题号查询题目定义。
func (c *Catalog) QuestionByID(questionID string) (Question, bool) {
	if c == nil || c.byID == nil {
		return Question{}, false
	}
	question, ok := c.byID[questionID]
	return question, ok
}

// CollectionWindow 判断指定时间点下问卷是否处于开放收集窗口。
func (c *Catalog) CollectionWindow(now time.Time) (CollectionStatus, bool) {
	if c == nil {
		return CollectionStatusAlwaysOpen, true
	}
	if now.IsZero() {
		now = time.Now()
	}
	if c.CollectionStartsAt == nil && c.CollectionEndsAt == nil {
		return CollectionStatusAlwaysOpen, true
	}
	if c.CollectionStartsAt != nil && now.Before(*c.CollectionStartsAt) {
		return CollectionStatusScheduled, false
	}
	if c.CollectionEndsAt != nil && now.After(*c.CollectionEndsAt) {
		return CollectionStatusClosed, false
	}
	return CollectionStatusOpen, true
}
