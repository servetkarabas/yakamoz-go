package comment

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/karabas/yakamoz/internal/platform/httpx"
	"github.com/karabas/yakamoz/internal/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type commentResponse struct {
	ID        uuid.UUID `json:"id"`
	TopicID   uuid.UUID `json:"topic_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Language  string    `json:"language"`
	Body      string    `json:"body"`
	CreatedAt string    `json:"created_at"`
}

func response(value Comment) commentResponse {
	return commentResponse{ID: value.ID, TopicID: value.TopicID, AuthorID: value.AuthorID,
		Language: value.Language, Body: value.Body,
		CreatedAt: value.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00")}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TopicID  uuid.UUID `json:"topic_id"`
		AuthorID uuid.UUID `json:"author_id"`
		Language string    `json:"language"`
		Body     string    `json:"body"`
	}
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "request body is invalid: "+err.Error(), nil)
		return
	}
	value, err := h.service.Create(r.Context(), request.TopicID, request.AuthorID, request.Language, request.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, response(value))
}

func (h *Handler) ListByTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := uuid.Parse(r.URL.Query().Get("topic_id"))
	if err != nil {
		writeError(w, ErrInvalid)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	values, err := h.service.ListByTopic(r.Context(), ListFilter{TopicID: topicID, Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	result := make([]commentResponse, 0, len(values))
	for _, value := range values {
		result = append(result, response(value))
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalidID)
		return
	}
	value, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, response(value))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalidID)
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrInvalidID):
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "request validation failed", nil)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "comment not found", nil)
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
