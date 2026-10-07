package author

import (
	"context"

	"github.com/karabas/yakamoz/internal/uuid"
)

type ListFilter struct {
	Limit  int
	Offset int
}

type Repository interface {
	Create(context.Context, Author) error
	GetByID(context.Context, uuid.UUID) (Author, error)
	GetByNickname(context.Context, string) (Author, error)
	List(context.Context, ListFilter) ([]Author, error)
	Update(context.Context, uuid.UUID, Update) (Author, error)
	SetStatus(context.Context, uuid.UUID, Status) (Author, error)
	Delete(context.Context, uuid.UUID) error
}
