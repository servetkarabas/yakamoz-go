package author

import (
	"context"
	"errors"
	"testing"

	"github.com/karabas/yakamoz/internal/uuid"
)

func TestServiceCreateAndGet(t *testing.T) {
	service := NewService(NewMemoryRepository())
	value, err := service.Create(context.Background(), "yakamoz", "dev@example.com", "bio", "tr", RoleAuthor)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	got, err := service.GetByID(context.Background(), value.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Nickname != "yakamoz" {
		t.Fatalf("Nickname = %q", got.Nickname)
	}
}

func TestServiceAllowsManyAuthorsButOnlyOneAdminAndReviewer(t *testing.T) {
	service := NewService(NewMemoryRepository())
	for i := 0; i < 5; i++ {
		if _, err := service.Create(context.Background(), "author"+string(rune('a'+i)), "author"+string(rune('a'+i))+"@example.com", "", "tr", RoleAuthor); err != nil {
			t.Fatalf("create author %d: %v", i, err)
		}
	}
	admin, err := service.Create(context.Background(), "admin", "admin@example.com", "", "en", RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), "admin2", "admin2@example.com", "", "en", RoleAdmin); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate admin error = %v", err)
	}
	reviewer, err := service.Create(context.Background(), "reviewer", "reviewer@example.com", "", "en", RoleReviewer)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), "reviewer2", "reviewer2@example.com", "", "en", RoleReviewer); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate reviewer error = %v", err)
	}
	if _, err := service.Update(context.Background(), admin.ID, Update{Role: rolePointer(RoleReviewer)}); !errors.Is(err, ErrConflict) {
		t.Fatalf("updating admin to occupied reviewer role = %v", err)
	}
	if _, err := service.Update(context.Background(), reviewer.ID, Update{Role: rolePointer(RoleAuthor)}); err != nil {
		t.Fatalf("demote reviewer: %v", err)
	}
	if _, err := service.Create(context.Background(), "reviewer3", "reviewer3@example.com", "", "en", RoleReviewer); err != nil {
		t.Fatalf("assign released reviewer role: %v", err)
	}
}

func rolePointer(role Role) *Role {
	return &role
}

func TestServiceValidationAndConflict(t *testing.T) {
	service := NewService(NewMemoryRepository())
	tests := []struct {
		name string
		run  func() error
		want error
	}{
		{name: "validation", run: func() error {
			_, err := service.Create(context.Background(), "x", "bad", "", "tr", RoleAuthor)
			return err
		}, want: ErrInvalid},
		{name: "missing", run: func() error {
			_, err := service.GetByID(context.Background(), uuid.New())
			return err
		}, want: ErrNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	if _, err := service.Create(context.Background(), "same", "same@example.com", "", "tr", RoleAuthor); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), "same", "other@example.com", "", "tr", RoleAuthor); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate nickname error = %v", err)
	}
	if _, err := service.Create(context.Background(), "other", "same@example.com", "", "tr", RoleAuthor); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate email error = %v", err)
	}
}
