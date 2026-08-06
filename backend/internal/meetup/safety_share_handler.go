package meetup

import (
	"net/http"

	"soulink-backend/internal/httpapi"
)

type createTrustedContactRequest struct {
	Label        string `json:"label"`
	ContactType  string `json:"contactType"`
	ContactValue string `json:"contactValue"`
}

type shareMeetupSafetyRequest struct {
	TrustedContactID int64 `json:"trustedContactId"`
}

func (h *Handler) ListTrustedContacts(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	contacts, err := h.service.ListTrustedContacts(r.Context(), userID)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, contacts)
}

func (h *Handler) CreateTrustedContact(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var request createTrustedContactRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	contact, err := h.service.CreateTrustedContact(r.Context(), userID, TrustedContactInput{
		Label: request.Label, ContactType: request.ContactType, ContactValue: request.ContactValue,
	})
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "trusted contact created", contact)
}

func (h *Handler) DeleteTrustedContact(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	contactID, err := parseMeetupRouteID(r, "contactId")
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	if err := h.service.DeleteTrustedContact(r.Context(), userID, contactID); err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, nil)
}

func (h *Handler) ShareSafetySummary(w http.ResponseWriter, r *http.Request) {
	userID, meetupID, ok := meetupRequestIDs(w, r)
	if !ok {
		return
	}
	var request shareMeetupSafetyRequest
	if err := httpapi.DecodeJSON(r, &request); err != nil {
		httpapi.Error(w, err)
		return
	}
	share, err := h.service.ShareMeetupSafetySummary(
		r.Context(), userID, meetupID, request.TrustedContactID,
	)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.Respond(w, http.StatusCreated, "meetup safety summary shared", share)
}
