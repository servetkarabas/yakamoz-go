package topic

import (
	"context"

	"github.com/karabas/yakamoz/internal/uuid"
)

type Repository interface {
	Create(context.Context, Topic) error
	GetByID(context.Context, uuid.UUID) (Topic, error)
	GetBySlug(context.Context, string) (Topic, error)
	List(context.Context, ListFilter) ([]Topic, error)
	UpsertTranslation(context.Context, uuid.UUID, Translation) error
	SetStatus(context.Context, uuid.UUID, Status) (Topic, error)
}
