package author

import (
	"context"
	"strings"
	"sync"

	"github.com/karabas/yakamoz/internal/uuid"
)

type MemoryRepository struct {
	mu         sync.RWMutex
	byID       map[uuid.UUID]Author
	byNickname map[string]uuid.UUID
	byEmail    map[string]uuid.UUID
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{byID: make(map[uuid.UUID]Author), byNickname: make(map[string]uuid.UUID), byEmail: make(map[string]uuid.UUID)}
}

func (r *MemoryRepository) Create(_ context.Context, value Author) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	nickname := strings.ToLower(value.Nickname)
	email := strings.ToLower(value.Email)
	if _, ok := r.byNickname[nickname]; ok {
		return ErrConflict
	}
	if _, ok := r.byEmail[email]; ok {
		return ErrConflict
	}
	if value.Role != RoleAuthor && r.hasRole(value.Role, uuid.Nil) {
		return ErrConflict
	}
	r.byID[value.ID] = value
	r.byNickname[nickname] = value.ID
	r.byEmail[email] = value.ID
	return nil
}

func (r *MemoryRepository) hasRole(role Role, except uuid.UUID) bool {
	for id, value := range r.byID {
		if id != except && value.Role == role {
			return true
		}
	}
	return false
}

func (r *MemoryRepository) GetByID(_ context.Context, id uuid.UUID) (Author, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.byID[id]
	if !ok {
		return Author{}, ErrNotFound
	}
	return value, nil
}

func (r *MemoryRepository) GetByNickname(_ context.Context, nickname string) (Author, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byNickname[strings.ToLower(nickname)]
	if !ok {
		return Author{}, ErrNotFound
	}
	return r.byID[id], nil
}

func (r *MemoryRepository) List(_ context.Context, filter ListFilter) ([]Author, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	values := make([]Author, 0, len(r.byID))
	for _, value := range r.byID {
		values = append(values, value)
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	if filter.Offset >= len(values) {
		return []Author{}, nil
	}
	end := len(values)
	if filter.Limit > 0 && filter.Offset+filter.Limit < end {
		end = filter.Offset + filter.Limit
	}
	return values[filter.Offset:end], nil
}

func (r *MemoryRepository) Update(_ context.Context, id uuid.UUID, update Update) (Author, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.byID[id]
	if !ok {
		return Author{}, ErrNotFound
	}
	if update.Bio != nil {
		value.Bio = *update.Bio
	}
	if update.PreferredLanguage != nil {
		value.PreferredLanguage = *update.PreferredLanguage
	}
	if update.Role != nil {
		if *update.Role != RoleAuthor && r.hasRole(*update.Role, id) {
			return Author{}, ErrConflict
		}
		value.Role = *update.Role
	}
	r.byID[id] = value
	return value, nil
}

func (r *MemoryRepository) SetStatus(_ context.Context, id uuid.UUID, status Status) (Author, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.byID[id]
	if !ok {
		return Author{}, ErrNotFound
	}
	value.Status = status
	r.byID[id] = value
	return value, nil
}

func (r *MemoryRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	delete(r.byID, id)
	delete(r.byNickname, strings.ToLower(value.Nickname))
	delete(r.byEmail, strings.ToLower(value.Email))
	return nil
}
