package comment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/karabas/yakamoz/internal/uuid"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) Create(ctx context.Context, topicID, authorID uuid.UUID, language, body string) (Comment, error) {
	value := Comment{
		ID:        uuid.New(),
		TopicID:   topicID,
		AuthorID:  authorID,
		Language:  strings.TrimSpace(language),
		Body:      body,
		CreatedAt: s.now().UTC(),
	}
	if err := value.Validate(); err != nil {
		return Comment{}, err
	}
	if err := s.repo.Create(ctx, value); err != nil {
		return Comment{}, err
	}
	return value, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Comment, error) {
	if id == uuid.Nil {
		return Comment{}, ErrInvalidID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByTopic(ctx context.Context, filter ListFilter) ([]Comment, error) {
	if filter.TopicID == uuid.Nil {
		return nil, ErrInvalid
	}
	return s.repo.ListByTopic(ctx, filter)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidID
	}
	return s.repo.Delete(ctx, id)
}

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
