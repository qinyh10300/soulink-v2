package meta

import (
	"net/http"
	"strings"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/school"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Departments(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("meta/departments"))
		return
	}

	email := strings.TrimSpace(r.URL.Query().Get("email"))
	if email != "" {
		schoolName, ok := school.NameByEmail(email)
		if !ok {
			httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid school email"))
			return
		}
		httpapi.OK(w, h.service.DepartmentsForSchool(schoolName))
		return
	}

	schoolValue := strings.TrimSpace(r.URL.Query().Get("school"))
	if schoolValue == "" {
		httpapi.OK(w, h.service.Departments())
		return
	}

	schoolName, ok := school.ParseName(schoolValue)
	if !ok {
		httpapi.Error(w, httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "invalid school"))
		return
	}

	httpapi.OK(w, h.service.DepartmentsForSchool(schoolName))
}

func (h *Handler) Grades(w http.ResponseWriter, r *http.Request) {
	httpapi.OK(w, h.service.Grades())
}

func (h *Handler) ContactTypes(w http.ResponseWriter, r *http.Request) {
	httpapi.OK(w, h.service.ContactTypes())
}

func (h *Handler) RegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.Tsinghua)
}

func (h *Handler) PKURegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.Peking)
}

func (h *Handler) BNURegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.BeijingNormal)
}

func (h *Handler) RUCRegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.Renmin)
}

func (h *Handler) CUCRegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.Communication)
}

func (h *Handler) BFSURegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.Beiwai)
}

func (h *Handler) USTBRegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.USTB)
}

func (h *Handler) CSURegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.CSU)
}

func (h *Handler) TJURegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.Tongji)
}

func (h *Handler) SJTURegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.ShanghaiJiao)
}

func (h *Handler) FDURegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.Fudan)
}

func (h *Handler) NUSRegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.NUS)
}

func (h *Handler) NTURegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.NTU)
}

func (h *Handler) ZUELRegisteredUserCount(w http.ResponseWriter, r *http.Request) {
	h.registeredUserCountBySchool(w, r, school.ZUEL)
}

func (h *Handler) registeredUserCountBySchool(w http.ResponseWriter, r *http.Request, schoolName school.Name) {
	if h == nil || h.service == nil {
		httpapi.Error(w, httpapi.FeatureNotReady("registered user count"))
		return
	}
	count, err := h.service.registeredUserCountBySchool(r.Context(), schoolName)
	if err != nil {
		httpapi.Error(w, err)
		return
	}
	httpapi.OK(w, map[string]int{"registeredUserCount": count})
}
