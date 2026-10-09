package author

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

func (s *Service) Create(ctx context.Context, nickname, email, bio, preferredLanguage string, role Role) (Author, error) {
	if role == "" {
		role = RoleAuthor
	}
	value := Author{
		ID: uuid.New(), Nickname: strings.TrimSpace(nickname), Email: strings.TrimSpace(email), Bio: bio,
		PreferredLanguage: preferredLanguage, Role: role, Status: StatusActive, CreatedAt: s.now().UTC(),
	}
	value.UpdatedAt = value.CreatedAt
	if err := value.Validate(); err != nil {
		return Author{}, err
	}
	if err := s.repo.Create(ctx, value); err != nil {
		return Author{}, err
	}
	return value, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Author, error) {
	if id == uuid.Nil {
		return Author{}, ErrInvalidID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByNickname(ctx context.Context, nickname string) (Author, error) {
	if strings.TrimSpace(nickname) == "" {
		return Author{}, ErrInvalid
	}
	return s.repo.GetByNickname(ctx, nickname)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Author, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, update Update) (Author, error) {
	if id == uuid.Nil {
		return Author{}, ErrInvalidID
	}
	if update.PreferredLanguage != nil && strings.TrimSpace(*update.PreferredLanguage) == "" {
		return Author{}, ErrInvalid
	}
	if update.Role != nil && !update.Role.Valid() {
		return Author{}, ErrInvalid
	}
	value, err := s.repo.Update(ctx, id, update)
	if err != nil {
		return Author{}, err
	}
	value.UpdatedAt = s.now().UTC()
	return value, nil
}

func (s *Service) Suspend(ctx context.Context, id uuid.UUID) (Author, error) {
	return s.setStatus(ctx, id, StatusSuspended)
}

func (s *Service) Activate(ctx context.Context, id uuid.UUID) (Author, error) {
	return s.setStatus(ctx, id, StatusActive)
}

func (s *Service) setStatus(ctx context.Context, id uuid.UUID, status Status) (Author, error) {
	if id == uuid.Nil {
		return Author{}, ErrInvalidID
	}
	value, err := s.repo.SetStatus(ctx, id, status)
	if err != nil {
		return Author{}, err
	}
	value.UpdatedAt = s.now().UTC()
	return value, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidID
	}
	return s.repo.Delete(ctx, id)
}

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
