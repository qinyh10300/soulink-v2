package engagement

import "time"

const defaultTimezone = "Asia/Shanghai"

type PreferencesView struct {
	IsAdmin               bool     `json:"isAdmin"`
	NotificationsEnabled  bool     `json:"notificationsEnabled"`
	NewMessages           bool     `json:"newMessages"`
	MeetupInvites         bool     `json:"meetupInvites"`
	MeetupReminders       bool     `json:"meetupReminders"`
	DailyPicks            bool     `json:"dailyPicks"`
	PermissionState       string   `json:"permissionState"`
	Timezone              string   `json:"timezone"`
	SelectedScenes        []string `json:"selectedScenes"`
	OnboardingCompleted   bool     `json:"onboardingCompleted"`
	OnboardingCompletedAt *string  `json:"onboardingCompletedAt,omitempty"`
	UpdatedAt             *string  `json:"updatedAt,omitempty"`
}

type NotificationPreferencesInput struct {
	NotificationsEnabled bool
	NewMessages          bool
	MeetupInvites        bool
	MeetupReminders      bool
	DailyPicks           bool
	PermissionState      string
	Timezone             string
}

type OnboardingInput struct {
	SelectedScenes []string
}

type CampusThemeView struct {
	Key           string `json:"key"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	DefaultPrompt string `json:"defaultPrompt"`
	ScheduleHint  string `json:"scheduleHint"`
	IconName      string `json:"iconName"`
	IsActive      bool   `json:"isActive"`
	DisplayOrder  int    `json:"displayOrder"`
	Selected      bool   `json:"selected"`
	UpdatedAt     string `json:"updatedAt"`
}

type UpdateCampusThemeInput struct {
	Title         string
	Description   string
	DefaultPrompt string
	ScheduleHint  string
	IconName      string
	IsActive      bool
	DisplayOrder  int
}

type NotificationView struct {
	ID           int64   `json:"id"`
	Kind         string  `json:"kind"`
	Title        string  `json:"title"`
	Body         string  `json:"body"`
	ActionPath   string  `json:"actionPath"`
	ScheduledFor string  `json:"scheduledFor"`
	DeliveredAt  *string `json:"deliveredAt,omitempty"`
	ReadAt       *string `json:"readAt,omitempty"`
	CreatedAt    string  `json:"createdAt"`
}

type NotificationPage struct {
	Items       []NotificationView `json:"items"`
	UnreadCount int64              `json:"unreadCount"`
	NextCursor  *int64             `json:"nextCursor,omitempty"`
}

type CampusThemeMetricView struct {
	Key                  string  `json:"key"`
	Title                string  `json:"title"`
	Category             string  `json:"category"`
	SupplyUsers          int64   `json:"supplyUsers"`
	DemandUsers          int64   `json:"demandUsers"`
	GroupIntents         int64   `json:"groupIntents"`
	GroupsFormed         int64   `json:"groupsFormed"`
	GroupFormationRate   float64 `json:"groupFormationRate"`
	MeetupsEligible      int64   `json:"meetupsEligible"`
	MeetupsCompleted     int64   `json:"meetupsCompleted"`
	MeetupCompletionRate float64 `json:"meetupCompletionRate"`
}

type CampusThemeMetricsView struct {
	Days        int                     `json:"days"`
	From        time.Time               `json:"from"`
	To          time.Time               `json:"to"`
	Themes      []CampusThemeMetricView `json:"themes"`
	GeneratedAt time.Time               `json:"generatedAt"`
}
