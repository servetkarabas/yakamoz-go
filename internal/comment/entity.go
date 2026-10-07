package comment

import (
	"errors"
	"strings"
	"time"

	"github.com/karabas/yakamoz/internal/uuid"
)

var (
	ErrInvalid   = errors.New("invalid comment")
	ErrNotFound  = errors.New("comment not found")
	ErrInvalidID = errors.New("invalid comment id")
)

type Comment struct {
	ID        uuid.UUID `json:"id"`
	TopicID   uuid.UUID `json:"topic_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Language  string    `json:"language"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func (c Comment) Validate() error {
	if c.TopicID == uuid.Nil || c.AuthorID == uuid.Nil {
		return ErrInvalid
	}
	if strings.TrimSpace(c.Language) == "" || strings.TrimSpace(c.Body) == "" {
		return ErrInvalid
	}
	return nil
}
