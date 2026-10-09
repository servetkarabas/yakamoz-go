package author

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/karabas/yakamoz/internal/uuid"
)

type Status string

type Role string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"

	RoleAuthor   Role = "author"
	RoleReviewer Role = "reviewer"
	RoleAdmin    Role = "admin"
)

var (
	ErrInvalid   = errors.New("invalid author")
	ErrNotFound  = errors.New("author not found")
	ErrConflict  = errors.New("author conflict")
	ErrInvalidID = errors.New("invalid author id")
)

type Author struct {
	ID                uuid.UUID `json:"id"`
	Nickname          string    `json:"-"`
	Email             string    `json:"-"`
	Bio               string    `json:"bio"`
	PreferredLanguage string    `json:"preferred_language"`
	Role              Role      `json:"role"`
	Status            Status    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (a Author) Validate() error {
	if len([]rune(strings.TrimSpace(a.Nickname))) < 2 || strings.ContainsAny(a.Nickname, " \t\r\n") {
		log.Println("Invalid nickname")
		return ErrInvalid
	}
	if !strings.Contains(a.Email, "@") || strings.TrimSpace(a.Email) != a.Email {
		log.Println("Invalid email")
		return ErrInvalid
	}
	if a.PreferredLanguage == "" {
		log.Println("Invalid preferred language")
		return ErrInvalid
	}
	if !a.Role.Valid() {
		return ErrInvalid
	}
	return nil
}

func (r Role) Valid() bool {
	return r == RoleAuthor || r == RoleReviewer || r == RoleAdmin
}

type Update struct {
	Bio               *string
	PreferredLanguage *string
	Role              *Role
}
