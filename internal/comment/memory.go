package comment

import (
	"context"
	"sort"
	"sync"

	"github.com/karabas/yakamoz/internal/uuid"
)

type MemoryRepository struct {
	mu      sync.RWMutex
	byID    map[uuid.UUID]Comment
	byTopic map[uuid.UUID][]uuid.UUID
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{byID: make(map[uuid.UUID]Comment), byTopic: make(map[uuid.UUID][]uuid.UUID)}
}

func (r *MemoryRepository) Create(_ context.Context, value Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[value.ID] = value
	r.byTopic[value.TopicID] = append(r.byTopic[value.TopicID], value.ID)
	return nil
}

func (r *MemoryRepository) GetByID(_ context.Context, id uuid.UUID) (Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.byID[id]
	if !ok {
		return Comment{}, ErrNotFound
	}
	return value, nil
}

func (r *MemoryRepository) ListByTopic(_ context.Context, filter ListFilter) ([]Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	values := make([]Comment, 0, len(r.byTopic[filter.TopicID]))
	for _, id := range r.byTopic[filter.TopicID] {
		values = append(values, r.byID[id])
	}
	sort.Slice(values, func(i, j int) bool { return values[i].CreatedAt.Before(values[j].CreatedAt) })
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	if filter.Offset >= len(values) {
		return []Comment{}, nil
	}
	end := len(values)
	if filter.Limit > 0 && filter.Offset+filter.Limit < end {
		end = filter.Offset + filter.Limit
	}
	return values[filter.Offset:end], nil
}

func (r *MemoryRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	delete(r.byID, id)
	ids := r.byTopic[value.TopicID]
	for i, existing := range ids {
		if existing == id {
			r.byTopic[value.TopicID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	return nil
}
