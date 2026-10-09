package topic

import (
	"errors"
	"strings"
	"time"

	"github.com/karabas/yakamoz/internal/uuid"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

type Source string

const (
	SourceHuman Source = "human"
	SourceAI    Source = "ai"
)

var (
	ErrInvalid  = errors.New("invalid topic")
	ErrNotFound = errors.New("topic not found")
	ErrConflict = errors.New("topic conflict")
)

type Translation struct {
	Language     string    `json:"language"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Source       Source    `json:"source"`
	TranslatedAt time.Time `json:"translated_at"`
}

type Topic struct {
	ID               uuid.UUID              `json:"id"`
	Slug             string                 `json:"slug"`
	OriginalLanguage string                 `json:"original_language"`
	Status           Status                 `json:"status"`
	CreatedBy        uuid.UUID              `json:"created_by"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	Translations     map[string]Translation `json:"translations"`
}

func (t Topic) Validate() error {
	if t.ID == uuid.Nil || t.CreatedBy == uuid.Nil || strings.TrimSpace(t.Slug) == "" || strings.TrimSpace(t.OriginalLanguage) == "" {
		return ErrInvalid
	}
	if len(t.Translations) == 0 {
		return ErrInvalid
	}
	translation, ok := t.Translations[t.OriginalLanguage]
	if !ok || strings.TrimSpace(translation.Title) == "" {
		return ErrInvalid
	}
	return nil
}

type ListFilter struct {
	Language string
	Status   Status
	Sort     string
	Limit    int
	Offset   int
}

type Resolved struct {
	Topic          Topic       `json:"topic"`
	Translation    Translation `json:"translation"`
	ServedLanguage string      `json:"served_language"`
}
