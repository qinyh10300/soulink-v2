package analytics

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"soulink-backend/internal/auth"
)

type EventRecorder interface {
	Record(context.Context, EventInput) error
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(payload []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(payload)
}

func Track(
	recorder EventRecorder,
	logger *slog.Logger,
	eventName string,
	subjectType string,
	subjectParam string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trackedWriter := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(trackedWriter, r)
			if recorder == nil || trackedWriter.status < 200 || trackedWriter.status >= 300 {
				return
			}
			user, ok := auth.CurrentUser(r)
			if !ok {
				return
			}
			var subjectID int64
			if subjectParam != "" {
				parsedID, err := strconv.ParseInt(chi.URLParam(r, subjectParam), 10, 64)
				if err != nil || parsedID <= 0 {
					return
				}
				subjectID = parsedID
			}
			if err := recorder.Record(r.Context(), EventInput{
				Name:        eventName,
				ActorUserID: user.ID,
				SubjectType: subjectType,
				SubjectID:   subjectID,
				RequestID:   middleware.GetReqID(r.Context()),
				Properties:  map[string]any{"source": "api"},
			}); err != nil && logger != nil {
				logger.Warn("record product event", "event", eventName, "error", err)
			}
		})
	}
}
