package app

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/karabas/yakamoz/internal/platform/httpx"
	"github.com/karabas/yakamoz/internal/platform/postgres"
	"github.com/karabas/yakamoz/internal/uuid"
)

type reactionTarget string

type reactionKind string

const (
	reactionTopic   reactionTarget = "topic"
	reactionComment reactionTarget = "comment"
	reactionLike    reactionKind   = "like"
	reactionDislike reactionKind   = "dislike"
)

type reactionSummary struct {
	Likes        int          `json:"likes"`
	Dislikes     int          `json:"dislikes"`
	UserReaction reactionKind `json:"user_reaction"`
}

type reactionStore interface {
	Set(context.Context, reactionTarget, uuid.UUID, uuid.UUID, reactionKind) error
	List(context.Context, reactionTarget, []uuid.UUID, uuid.UUID) (map[uuid.UUID]reactionSummary, error)
}

type reactionKey struct {
	target    reactionTarget
	itemID    uuid.UUID
	visitorID uuid.UUID
}

type memoryReactionStore struct {
	mu     sync.RWMutex
	values map[reactionKey]reactionKind
}

func newMemoryReactionStore() *memoryReactionStore {
	return &memoryReactionStore{values: make(map[reactionKey]reactionKind)}
}

func (s *memoryReactionStore) Set(_ context.Context, target reactionTarget, itemID, visitorID uuid.UUID, value reactionKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := reactionKey{target: target, itemID: itemID, visitorID: visitorID}
	if value == "" {
		delete(s.values, key)
	} else {
		s.values[key] = value
	}
	return nil
}

func (s *memoryReactionStore) List(_ context.Context, target reactionTarget, itemIDs []uuid.UUID, visitorID uuid.UUID) (map[uuid.UUID]reactionSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[uuid.UUID]reactionSummary, len(itemIDs))
	wanted := make(map[uuid.UUID]bool, len(itemIDs))
	for _, id := range itemIDs {
		wanted[id] = true
		result[id] = reactionSummary{}
	}
	for key, value := range s.values {
		if key.target != target || !wanted[key.itemID] {
			continue
		}
		summary := result[key.itemID]
		if value == reactionLike {
			summary.Likes++
		} else {
			summary.Dislikes++
		}
		if key.visitorID == visitorID {
			summary.UserReaction = value
		}
		result[key.itemID] = summary
	}
	return result, nil
}

type postgresReactionStore struct {
	pool *postgres.Pool
}

func newPostgresReactionStore(pool *postgres.Pool) *postgresReactionStore {
	return &postgresReactionStore{pool: pool}
}

func (s *postgresReactionStore) Set(ctx context.Context, target reactionTarget, itemID, visitorID uuid.UUID, value reactionKind) error {
	if value == "" {
		_, err := s.pool.ExecContext(ctx, `DELETE FROM reactions WHERE target_type=$1 AND target_id=$2 AND visitor_id=$3`, target, itemID, visitorID)
		return err
	}
	_, err := s.pool.ExecContext(ctx, `INSERT INTO reactions (target_type,target_id,visitor_id,reaction)
		VALUES ($1,$2,$3,$4) ON CONFLICT (target_type,target_id,visitor_id)
		DO UPDATE SET reaction=EXCLUDED.reaction`, target, itemID, visitorID, value)
	return err
}

func (s *postgresReactionStore) List(ctx context.Context, target reactionTarget, itemIDs []uuid.UUID, visitorID uuid.UUID) (map[uuid.UUID]reactionSummary, error) {
	result := make(map[uuid.UUID]reactionSummary, len(itemIDs))
	if len(itemIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, len(itemIDs))
	args := make([]any, 0, len(itemIDs)+2)
	args = append(args, target, visitorID)
	for i, id := range itemIDs {
		result[id] = reactionSummary{}
		placeholders[i] = "$" + strconv.Itoa(i+3)
		args = append(args, id)
	}
	query := `SELECT target_id, COUNT(*) FILTER (WHERE reaction='like'),
		COUNT(*) FILTER (WHERE reaction='dislike'),
		COALESCE(MAX(CASE WHEN visitor_id=$2 THEN reaction END), '')
		FROM reactions WHERE target_type=$1 AND target_id IN (` + strings.Join(placeholders, ",") + `)
		GROUP BY target_id`
	rows, err := s.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var summary reactionSummary
		if err := rows.Scan(&id, &summary.Likes, &summary.Dislikes, &summary.UserReaction); err != nil {
			return nil, err
		}
		result[id] = summary
	}
	return result, rows.Err()
}

type reactionHandler struct {
	store reactionStore
}

func validReactionTarget(target reactionTarget) bool {
	return target == reactionTopic || target == reactionComment
}

func validReactionKind(value reactionKind) bool {
	return value == "" || value == reactionLike || value == reactionDislike
}

func (h reactionHandler) List(w http.ResponseWriter, r *http.Request) {
	target := reactionTarget(r.URL.Query().Get("target"))
	if !validReactionTarget(target) {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "reaction target is invalid", nil)
		return
	}
	values := strings.Split(r.URL.Query().Get("target_ids"), ",")
	if len(values) == 1 && values[0] == "" {
		values = nil
	}
	if len(values) > 100 {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "too many reaction targets", nil)
		return
	}
	itemIDs := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := uuid.Parse(value)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "reaction target is invalid", nil)
			return
		}
		itemIDs = append(itemIDs, id)
	}
	visitorID := uuid.Nil
	if value := r.URL.Query().Get("visitor_id"); value != "" {
		parsed, err := uuid.Parse(value)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "validation_error", "visitor id is invalid", nil)
			return
		}
		visitorID = parsed
	}
	valuesByID, err := h.store.List(r.Context(), target, itemIDs, visitorID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
		return
	}
	result := make(map[string]reactionSummary, len(valuesByID))
	for id, summary := range valuesByID {
		result[id.String()] = summary
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h reactionHandler) Set(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Target    reactionTarget `json:"target"`
		TargetID  uuid.UUID      `json:"target_id"`
		VisitorID uuid.UUID      `json:"visitor_id"`
		Reaction  reactionKind   `json:"reaction"`
	}
	if err := httpx.DecodeJSON(r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "request body is invalid", nil)
		return
	}
	if !validReactionTarget(request.Target) || request.TargetID == uuid.Nil || request.VisitorID == uuid.Nil || !validReactionKind(request.Reaction) {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "reaction request is invalid", nil)
		return
	}
	if err := h.store.Set(r.Context(), request.Target, request.TargetID, request.VisitorID, request.Reaction); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
		return
	}
	values, err := h.store.List(r.Context(), request.Target, []uuid.UUID{request.TargetID}, request.VisitorID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, values[request.TargetID])
}
