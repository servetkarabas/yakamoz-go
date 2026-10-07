package comment

import (
	"context"
	"errors"
	"testing"

	"github.com/karabas/yakamoz/internal/uuid"
)

func TestServiceCreateAndList(t *testing.T) {
	service := NewService(NewMemoryRepository())
	topicID := uuid.New()
	value, err := service.Create(context.Background(), topicID, uuid.New(), "tr", "ilk yorum")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	got, err := service.ListByTopic(context.Background(), ListFilter{TopicID: topicID})
	if err != nil {
		t.Fatalf("ListByTopic() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != value.ID || got[0].Body != "ilk yorum" {
		t.Fatalf("ListByTopic() = %#v", got)
	}
}

func TestServiceValidationAndNotFound(t *testing.T) {
	service := NewService(NewMemoryRepository())
	tests := []struct {
		name string
		run  func() error
		want error
	}{
		{name: "empty body", run: func() error {
			_, err := service.Create(context.Background(), uuid.New(), uuid.New(), "tr", "  ")
			return err
		}, want: ErrInvalid},
		{name: "nil topic", run: func() error {
			_, err := service.Create(context.Background(), uuid.Nil, uuid.New(), "tr", "yorum")
			return err
		}, want: ErrInvalid},
		{name: "missing", run: func() error {
			_, err := service.GetByID(context.Background(), uuid.New())
			return err
		}, want: ErrNotFound},
		{name: "delete missing", run: func() error {
			return service.Delete(context.Background(), uuid.New())
		}, want: ErrNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
