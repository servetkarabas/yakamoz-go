package comment

import (
	"context"

	"github.com/karabas/yakamoz/internal/uuid"
)

type ListFilter struct {
	TopicID uuid.UUID
	Limit   int
	Offset  int
}

type Repository interface {
	Create(context.Context, Comment) error
	GetByID(context.Context, uuid.UUID) (Comment, error)
	ListByTopic(context.Context, ListFilter) ([]Comment, error)
	CountByTopics(context.Context, []uuid.UUID) (map[uuid.UUID]int, error)
	Delete(context.Context, uuid.UUID) error
}
