package author

import (
	"context"
	"errors"
	"testing"

	"github.com/karabas/yakamoz/internal/uuid"
)

func TestServiceCreateAndGet(t *testing.T) {
	service := NewService(NewMemoryRepository())
	value, err := service.Create(context.Background(), "yakamoz", "dev@example.com", "bio", "tr")
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

func TestServiceValidationAndConflict(t *testing.T) {
	service := NewService(NewMemoryRepository())
	tests := []struct {
		name string
		run  func() error
		want error
	}{
		{name: "validation", run: func() error {
			_, err := service.Create(context.Background(), "x", "bad", "", "tr")
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
	if _, err := service.Create(context.Background(), "same", "same@example.com", "", "tr"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), "same", "other@example.com", "", "tr"); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate nickname error = %v", err)
	}
	if _, err := service.Create(context.Background(), "other", "same@example.com", "", "tr"); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate email error = %v", err)
	}
}
