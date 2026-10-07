package author

import (
	"encoding/json"
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

type authorResponse struct {
	ID                uuid.UUID `json:"id"`
	Nickname          string    `json:"nickname"`
	Email             string    `json:"email"`
	Bio               string    `json:"bio"`
	PreferredLanguage string    `json:"preferred_language"`
	Status            Status    `json:"status"`
	CreatedAt         string    `json:"created_at"`
	UpdatedAt         string    `json:"updated_at"`
}

func response(value Author) authorResponse {
	return authorResponse{ID: value.ID, Nickname: value.Nickname, Email: value.Email, Bio: value.Bio,
		PreferredLanguage: value.PreferredLanguage, Status: value.Status,
		CreatedAt: value.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt: value.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00")}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Nickname          string `json:"nickname"`
		Email             string `json:"email"`
		Bio               string `json:"bio"`
		PreferredLanguage string `json:"preferred_language"`
	}
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "request body is invalid: "+err.Error(), nil)
		return
	}
	value, err := h.service.Create(r.Context(), request.Nickname, request.Email, request.Bio, request.PreferredLanguage)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, response(value))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	values, err := h.service.List(r.Context(), ListFilter{Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	result := make([]authorResponse, 0, len(values))
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

func (h *Handler) GetByNickname(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.GetByNickname(r.Context(), r.PathValue("nickname"))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, response(value))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, ErrInvalidID)
		return
	}
	var request struct {
		Bio               *string `json:"bio"`
		PreferredLanguage *string `json:"preferred_language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "request body is invalid", nil)
		return
	}
	value, err := h.service.Update(r.Context(), id, Update{Bio: request.Bio, PreferredLanguage: request.PreferredLanguage})
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
		httpx.WriteError(w, http.StatusNotFound, "not_found", "author not found", nil)
	case errors.Is(err, ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "conflict", "author already exists", nil)
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
