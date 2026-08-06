package app

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"soulink-backend/internal/analytics"
	"soulink-backend/internal/auth"
	"soulink-backend/internal/calendar"
	"soulink-backend/internal/chat"
	"soulink-backend/internal/connection"
	"soulink-backend/internal/feedback"
	"soulink-backend/internal/friend"
	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/match"
	"soulink-backend/internal/meetup"
	"soulink-backend/internal/meta"
	"soulink-backend/internal/playground"
	"soulink-backend/internal/profile"
	"soulink-backend/internal/questionnaire"
	"soulink-backend/internal/safety"
)

type Dependencies struct {
	Logger               *slog.Logger
	AllowedOrigins       []string
	AuthMiddleware       func(http.Handler) http.Handler
	MatchWriteGuard      func(http.Handler) http.Handler
	EventRecorder        analytics.EventRecorder
	AnalyticsHandler     *analytics.Handler
	AuthHandler          *auth.Handler
	CalendarHandler      *calendar.Handler
	FeedbackHandler      *feedback.Handler
	ProfileHandler       *profile.Handler
	MetaHandler          *meta.Handler
	QuestionnaireHandler *questionnaire.Handler
	MatchHandler         *match.Handler
	MeetupHandler        *meetup.Handler
	FriendHandler        *friend.Handler
	ChatHandler          *chat.Handler
	PlaygroundHandler    *playground.Handler
	ConnectionHandler    *connection.Handler
	SafetyHandler        *safety.Handler
	WSHandler            http.HandlerFunc
}

// NewRouter 创建 Web API 路由。
func NewRouter(deps Dependencies) http.Handler {
	authMiddleware := deps.AuthMiddleware
	if authMiddleware == nil {
		authMiddleware = passthroughMiddleware
	}
	matchWriteGuard := deps.MatchWriteGuard
	if matchWriteGuard == nil {
		matchWriteGuard = passthroughMiddleware
	}
	track := func(eventName string, subjectType string, subjectParam string) func(http.Handler) http.Handler {
		return analytics.Track(deps.EventRecorder, deps.Logger, eventName, subjectType, subjectParam)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(CORS(deps.AllowedOrigins))
	r.Use(RequestLogger(deps.Logger))
	r.Use(Recoverer(deps.Logger))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpapi.OK(w, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(api chi.Router) {
		if deps.WSHandler != nil {
			api.Get("/ws", deps.WSHandler)
		}

		api.Post("/auth/login", deps.AuthHandler.Login)
		api.Post("/auth/register/send-code", deps.AuthHandler.SendRegisterCode)
		api.Post("/auth/register/verify-code", deps.AuthHandler.VerifyRegisterCode)
		api.Post("/auth/register", deps.AuthHandler.Register)
		api.Post("/auth/password/forgot/send-code", deps.AuthHandler.SendForgotPasswordCode)
		api.Post("/auth/password/forgot/reset", deps.AuthHandler.ResetForgotPassword)

		api.Get("/meta/departments", deps.MetaHandler.Departments)
		api.Get("/meta/grades", deps.MetaHandler.Grades)
		api.Get("/meta/contact-types", deps.MetaHandler.ContactTypes)
		api.Get("/meta/registered-user-count", deps.MetaHandler.RegisteredUserCount)
		api.Get("/meta/pku-registered-user-count", deps.MetaHandler.PKURegisteredUserCount)
		api.Get("/meta/bnu-registered-user-count", deps.MetaHandler.BNURegisteredUserCount)
		api.Get("/meta/ruc-registered-user-count", deps.MetaHandler.RUCRegisteredUserCount)
		api.Get("/meta/cuc-registered-user-count", deps.MetaHandler.CUCRegisteredUserCount)
		api.Get("/meta/bfsu-registered-user-count", deps.MetaHandler.BFSURegisteredUserCount)
		api.Get("/meta/ustb-registered-user-count", deps.MetaHandler.USTBRegisteredUserCount)
		api.Get("/meta/csu-registered-user-count", deps.MetaHandler.CSURegisteredUserCount)
		api.Get("/meta/tju-registered-user-count", deps.MetaHandler.TJURegisteredUserCount)
		api.Get("/meta/sjtu-registered-user-count", deps.MetaHandler.SJTURegisteredUserCount)
		api.Get("/meta/fdu-registered-user-count", deps.MetaHandler.FDURegisteredUserCount)
		api.Get("/meta/nus-registered-user-count", deps.MetaHandler.NUSRegisteredUserCount)
		api.Get("/meta/ntu-registered-user-count", deps.MetaHandler.NTURegisteredUserCount)
		api.Get("/meta/zuel-registered-user-count", deps.MetaHandler.ZUELRegisteredUserCount)

		api.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			if deps.AnalyticsHandler != nil {
				r.Get("/admin/product-metrics", deps.AnalyticsHandler.Dashboard)
			}

			r.Post("/auth/logout", deps.AuthHandler.Logout)
			r.Get("/auth/me", deps.AuthHandler.Me)

			r.Get("/users/me/profile", deps.ProfileHandler.GetProfile)
			r.With(matchWriteGuard).Put("/users/me/profile", deps.ProfileHandler.UpdateProfile)
			r.Get("/users/{userId}/profile", deps.ProfileHandler.GetPublicProfile)
			if deps.CalendarHandler != nil {
				r.Get("/users/{userId}/calendar", deps.CalendarHandler.ListUserSchedule)
			}

			if deps.MeetupHandler != nil {
				r.Get("/users/me/availability", deps.MeetupHandler.ListAvailability)
				r.Put("/users/me/availability", deps.MeetupHandler.ReplaceAvailability)
				r.Post("/meetup-suggestions", deps.MeetupHandler.SuggestTimes)
				r.Route("/meetups", func(r chi.Router) {
					r.Get("/", deps.MeetupHandler.List)
					r.Post("/", deps.MeetupHandler.Create)
					r.Get("/{meetupId}", deps.MeetupHandler.Get)
					r.Put("/{meetupId}/response", deps.MeetupHandler.Respond)
					r.Put("/{meetupId}/reschedule", deps.MeetupHandler.Reschedule)
					r.Put("/{meetupId}/cancel", deps.MeetupHandler.Cancel)
					r.Put("/{meetupId}/complete", deps.MeetupHandler.Complete)
					r.Post("/{meetupId}/feedback", deps.MeetupHandler.SubmitFeedback)
					r.Delete("/{meetupId}/feedback", deps.MeetupHandler.RevokeFeedback)
				})
			}
			r.Get("/users/me/contacts", deps.ProfileHandler.GetContacts)
			r.With(matchWriteGuard).Put("/users/me/contacts/{contactType}", deps.ProfileHandler.UpsertContact)
			r.With(matchWriteGuard).Delete("/users/me/contacts/{contactType}", deps.ProfileHandler.DeleteContact)
			r.Post("/users/me/password/send-code", deps.AuthHandler.SendChangePasswordCode)
			r.Post("/users/me/password", deps.AuthHandler.ChangePassword)
			r.Post("/users/me/delete/send-code", deps.AuthHandler.SendDeleteAccountCode)
			r.Delete("/users/me", deps.AuthHandler.DeleteAccount)
			r.Post("/feedback", deps.FeedbackHandler.Submit)

			if deps.ConnectionHandler != nil {
				r.Post("/intent-interpretations", deps.ConnectionHandler.Interpret)
				r.Get("/users/me/matching-profile", deps.ConnectionHandler.GetMatchingProfile)
				r.Put("/users/me/matching-profile", deps.ConnectionHandler.UpdateMatchingProfile)
				r.Get("/users/me/recommendation-preferences", deps.ConnectionHandler.GetRecommendationPreferences)
				r.Put("/users/me/recommendation-preferences", deps.ConnectionHandler.UpdateRecommendationPreferences)
				r.Post("/users/me/recommendation-preferences/reset", deps.ConnectionHandler.ResetRecommendationPreferences)
				r.With(track(analytics.EventRecommendationsViewed, "", "")).Get("/recommendations/daily", deps.ConnectionHandler.DailyRecommendations)

				r.Route("/intents", func(r chi.Router) {
					r.Get("/feed", deps.ConnectionHandler.ListFeed)
					r.Get("/saved", deps.ConnectionHandler.ListSaved)
					r.Get("/mine", deps.ConnectionHandler.ListMine)
					r.With(track(analytics.EventIntentCreated, "", "")).Post("/", deps.ConnectionHandler.CreateIntent)
					r.Get("/{intentId}", deps.ConnectionHandler.GetIntent)
					r.Put("/{intentId}", deps.ConnectionHandler.UpdateIntent)
					r.Delete("/{intentId}", deps.ConnectionHandler.ExpireIntent)
					r.Post("/{intentId}/save", deps.ConnectionHandler.SaveIntent)
					r.Delete("/{intentId}/save", deps.ConnectionHandler.UnsaveIntent)
					r.Post("/{intentId}/dismiss", deps.ConnectionHandler.DismissIntent)
					r.With(track(analytics.EventInterestExpressed, "intent", "intentId")).Post("/{intentId}/interest", deps.ConnectionHandler.ExpressInterest)
					r.Delete("/{intentId}/interest", deps.ConnectionHandler.WithdrawInterest)
					r.Get("/{intentId}/interests", deps.ConnectionHandler.ListInterests)
					r.With(
						track(analytics.EventInterestAccepted, "intent", "intentId"),
						track(analytics.EventMatchCreated, "intent", "intentId"),
					).Post("/{intentId}/interests/{candidateId}/accept", deps.ConnectionHandler.AcceptInterest)
					r.Post("/{intentId}/interests/{candidateId}/decline", deps.ConnectionHandler.DeclineInterest)
					r.With(track(analytics.EventRecommendationsViewed, "intent", "intentId")).Get("/{intentId}/recommendations", deps.ConnectionHandler.Recommendations)
				})

				r.Route("/match-rooms", func(r chi.Router) {
					r.Get("/", deps.ConnectionHandler.ListMatchRooms)
					r.Get("/{roomId}", deps.ConnectionHandler.GetMatchRoom)
					r.Put("/{roomId}/plan", deps.ConnectionHandler.UpdateMatchRoomPlan)
					r.Put("/{roomId}/status", deps.ConnectionHandler.UpdateMatchRoomStatus)
					r.Put("/{roomId}/contact-sharing", deps.ConnectionHandler.SetContactSharing)
					r.Get("/{roomId}/messages", deps.ConnectionHandler.ListRoomMessages)
					r.Post("/{roomId}/messages", deps.ConnectionHandler.CreateRoomMessage)
				})
			}

			if deps.SafetyHandler != nil {
				r.Get("/safety/capabilities", deps.SafetyHandler.Capabilities)
				r.With(track(analytics.EventSafetyReportSubmitted, "", "")).Post("/safety/reports", deps.SafetyHandler.SubmitReport)
				r.Get("/admin/safety-reports", deps.SafetyHandler.ListReports)
				r.Put("/admin/safety-reports/{reportId}", deps.SafetyHandler.UpdateReport)
			}

			if deps.CalendarHandler != nil {
				r.Route("/calendar/events", func(r chi.Router) {
					r.Get("/", deps.CalendarHandler.ListEvents)
					r.Post("/", deps.CalendarHandler.CreateEvent)
					r.Put("/{eventId}", deps.CalendarHandler.UpdateEvent)
					r.Delete("/{eventId}", deps.CalendarHandler.CancelEvent)
					r.Put("/{eventId}/rsvp", deps.CalendarHandler.Respond)
				})
			}

			r.Get("/questionnaires/current", deps.QuestionnaireHandler.GetCurrent)
			r.Get("/questionnaires/current/progress", deps.QuestionnaireHandler.GetProgress)
			r.Get("/questionnaires/current/portrait", deps.QuestionnaireHandler.GetPortrait)
			r.With(matchWriteGuard).Put("/questionnaires/current/participation", deps.QuestionnaireHandler.UpdateParticipation)
			r.With(matchWriteGuard).Put("/questionnaires/current/answers", deps.QuestionnaireHandler.SaveAnswers)
			r.With(matchWriteGuard).Put("/questionnaires/current/answers/{questionId}", deps.QuestionnaireHandler.SaveAnswer)

			r.Get("/matches/current", deps.MatchHandler.GetCurrent)
			r.Get("/matches/history", deps.MatchHandler.GetHistory)
			r.Get("/matches/cycle/status", deps.MatchHandler.GetCycleStatus)
			r.Post("/matches/{matchId}/contact", deps.MatchHandler.ContactMatchedPartner)
			r.Get("/matches/{matchId}", deps.MatchHandler.GetDetail)

			r.Post("/friends/request", deps.FriendHandler.SendRequest)
			r.Post("/friends/accept", deps.FriendHandler.AcceptRequest)
			r.Post("/friends/reject", deps.FriendHandler.RejectRequest)
			r.Get("/friends", deps.FriendHandler.ListFriends)
			r.Get("/friends/requests/received", deps.FriendHandler.ListReceivedRequests)
			r.Get("/friends/requests/sent", deps.FriendHandler.ListSentRequests)
			r.Delete("/friends/{friendshipId}", deps.FriendHandler.RemoveFriend)

			r.Get("/conversations", deps.ChatHandler.ListConversations)
			r.Post("/conversations", deps.ChatHandler.CreateConversation)
			r.Get("/conversations/{conversationId}/messages", deps.ChatHandler.ListMessages)
			r.Post("/conversations/{conversationId}/messages", deps.ChatHandler.CreateMessage)
			r.Post("/conversations/{conversationId}/read", deps.ChatHandler.MarkRead)
			r.Get("/conversations/unread-count", deps.ChatHandler.GetUnreadCount)

			r.Route("/chat-groups", func(r chi.Router) {
				r.Get("/", deps.ChatHandler.ListGroups)
				r.Post("/", deps.ChatHandler.CreateGroup)
				r.Get("/{groupId}", deps.ChatHandler.GetGroup)
				r.Put("/{groupId}", deps.ChatHandler.UpdateGroup)
				r.Post("/{groupId}/members", deps.ChatHandler.AddGroupMembers)
				r.Delete("/{groupId}/members/{userId}", deps.ChatHandler.RemoveGroupMember)
				r.Get("/{groupId}/messages", deps.ChatHandler.ListGroupMessages)
				r.Post("/{groupId}/messages", deps.ChatHandler.CreateGroupMessage)
				r.Post("/{groupId}/read", deps.ChatHandler.MarkGroupRead)
			})

			if deps.PlaygroundHandler != nil {
				r.Get("/playground/posts", deps.PlaygroundHandler.ListPosts)
				r.Post("/playground/posts", deps.PlaygroundHandler.CreatePost)
				r.Post("/playground/posts/{postId}/likes", deps.PlaygroundHandler.LikePost)
				r.Delete("/playground/posts/{postId}/likes", deps.PlaygroundHandler.UnlikePost)
				r.Post("/playground/posts/{postId}/comments", deps.PlaygroundHandler.CreateComment)
				r.Post("/playground/comments/{commentId}/likes", deps.PlaygroundHandler.LikeComment)
				r.Delete("/playground/comments/{commentId}/likes", deps.PlaygroundHandler.UnlikeComment)
			}
		})
	})

	return r
}
