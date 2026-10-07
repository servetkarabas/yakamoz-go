package topic

import (
	"context"
	"strings"
	"sync"

	"github.com/karabas/yakamoz/internal/uuid"
)

type MemoryRepository struct {
	mu     sync.RWMutex
	byID   map[uuid.UUID]Topic
	bySlug map[string]uuid.UUID
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{byID: make(map[uuid.UUID]Topic), bySlug: make(map[string]uuid.UUID)}
}

func clone(value Topic) Topic {
	translations := value.Translations
	value.Translations = make(map[string]Translation, len(translations))
	for key, translation := range translations {
		value.Translations[key] = translation
	}
	return value
}

func (r *MemoryRepository) Create(_ context.Context, value Topic) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.bySlug[value.Slug]; ok {
		return ErrConflict
	}
	r.byID[value.ID] = clone(value)
	r.bySlug[value.Slug] = value.ID
	return nil
}

func (r *MemoryRepository) GetByID(_ context.Context, id uuid.UUID) (Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.byID[id]
	if !ok {
		return Topic{}, ErrNotFound
	}
	return clone(value), nil
}

func (r *MemoryRepository) GetBySlug(_ context.Context, slug string) (Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.bySlug[strings.ToLower(slug)]
	if !ok {
		return Topic{}, ErrNotFound
	}
	return clone(r.byID[id]), nil
}

func (r *MemoryRepository) List(_ context.Context, filter ListFilter) ([]Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	values := make([]Topic, 0, len(r.byID))
	for _, value := range r.byID {
		if filter.Status != "" && value.Status != filter.Status {
			continue
		}
		if filter.Language != "" {
			if _, ok := value.Translations[filter.Language]; !ok {
				continue
			}
		}
		values = append(values, clone(value))
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	if filter.Offset >= len(values) {
		return []Topic{}, nil
	}
	end := len(values)
	if filter.Limit > 0 && filter.Offset+filter.Limit < end {
		end = filter.Offset + filter.Limit
	}
	return values[filter.Offset:end], nil
}

func (r *MemoryRepository) UpsertTranslation(_ context.Context, id uuid.UUID, value Translation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	topic, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	if topic.Translations == nil {
		topic.Translations = make(map[string]Translation)
	}
	topic.Translations[value.Language] = value
	r.byID[id] = clone(topic)
	return nil
}

func (r *MemoryRepository) SetStatus(_ context.Context, id uuid.UUID, status Status) (Topic, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.byID[id]
	if !ok {
		return Topic{}, ErrNotFound
	}
	value.Status = status
	r.byID[id] = clone(value)
	return clone(value), nil
}
