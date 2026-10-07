package topic

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/karabas/yakamoz/internal/platform/httpx"
	"github.com/karabas/yakamoz/internal/platform/i18n"
	"github.com/karabas/yakamoz/internal/uuid"
)

type Handler struct {
	service         *Service
	defaultLanguage string
}

func NewHandler(service *Service, defaultLanguage string) *Handler {
	return &Handler{service: service, defaultLanguage: defaultLanguage}
}

type topicResponse struct {
	ID               uuid.UUID `json:"id"`
	Slug             string    `json:"slug"`
	OriginalLanguage string    `json:"original_language"`
	Status           Status    `json:"status"`
	CreatedBy        uuid.UUID `json:"created_by"`
	CreatedAt        string    `json:"created_at"`
	UpdatedAt        string    `json:"updated_at"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	ServedLanguage   string    `json:"served_language"`
}

func topicResult(value Resolved) topicResponse {
	return topicResponse{ID: value.Topic.ID, Slug: value.Topic.Slug, OriginalLanguage: value.Topic.OriginalLanguage,
		Status: value.Topic.Status, CreatedBy: value.Topic.CreatedBy,
		CreatedAt: value.Topic.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt: value.Topic.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		Title:     value.Translation.Title, Description: value.Translation.Description, ServedLanguage: value.ServedLanguage}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Language    string    `json:"language"`
		AuthorID    uuid.UUID `json:"author_id"`
	}
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "request body is invalid", nil)
		return
	}
	value, err := h.service.Create(r.Context(), request.Title, request.Description, request.Language, request.AuthorID)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, topicResult(Resolved{Topic: value, Translation: value.Translations[value.OriginalLanguage], ServedLanguage: value.OriginalLanguage}))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	values, err := h.service.List(r.Context(), ListFilter{Language: r.URL.Query().Get("lang"), Status: Status(r.URL.Query().Get("status")), Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	result := make([]topicResponse, 0, len(values))
	for _, value := range values {
		result = append(result, topicResult(h.service.resolve(value, i18n.Requested(r, h.defaultLanguage))))
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalid)
		return
	}
	value, err := h.service.GetByID(r.Context(), id, i18n.Requested(r, h.defaultLanguage))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, topicResult(value))
}

func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.GetBySlug(r.Context(), r.PathValue("slug"), i18n.Requested(r, h.defaultLanguage))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, topicResult(value))
}

func (h *Handler) AddTranslation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalid)
		return
	}
	var request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "request body is invalid", nil)
		return
	}
	value, err := h.service.AddTranslation(r.Context(), id, r.PathValue("lang"), request.Title, request.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, value)
}

func (h *Handler) Translate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalid)
		return
	}
	var request struct {
		TargetLanguages []string `json:"target_languages"`
		Force           bool     `json:"force"`
	}
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "request body is invalid", nil)
		return
	}
	value, err := h.service.TranslateWithAI(r.Context(), id, request.TargetLanguages, TranslateOptions{Force: request.Force})
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, value)
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalid)
		return
	}
	value, err := h.service.Publish(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, value)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "request validation failed", nil)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "topic not found", nil)
	case errors.Is(err, ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "conflict", "topic already exists", nil)
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
