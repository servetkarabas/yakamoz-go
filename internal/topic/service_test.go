package topic

import (
	"context"
	"testing"
	"time"

	"github.com/karabas/yakamoz/internal/platform/ai"
	"github.com/karabas/yakamoz/internal/uuid"
)

func TestServiceCreatesEnglishAndRequestedTranslations(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo, ai.Stub{}, "tr", time.Second)
	value, err := service.Create(context.Background(), "İstanbul", "şehir", "tr", uuid.New())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if value.Slug != "istanbul" {
		t.Fatalf("Slug = %q", value.Slug)
	}
	if value.Translations["en"].Title != "[en] İstanbul" || value.Translations["en"].Description != "[en] şehir" {
		t.Fatalf("English translation = %#v", value.Translations["en"])
	}
	resolved, err := service.GetByID(context.Background(), value.ID, "de")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if resolved.ServedLanguage != "de" || resolved.Translation.Source != SourceAI {
		t.Fatalf("AI translation = %#v", resolved)
	}
	if resolved.Translation.Description != "[de] şehir" {
		t.Fatalf("AI description = %q", resolved.Translation.Description)
	}
}

func TestServiceListGeneratesRequestedLanguage(t *testing.T) {
	service := NewService(NewMemoryRepository(), ai.Stub{}, "tr", time.Second)
	if _, err := service.Create(context.Background(), "Başlık", "Açıklama", "tr", uuid.New()); err != nil {
		t.Fatal(err)
	}

	values, err := service.List(context.Background(), ListFilter{Language: "de"})
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 || values[0].Translations["de"].Title != "[de] Başlık" || values[0].Translations["de"].Description != "[de] Açıklama" {
		t.Fatalf("translated topics = %#v", values)
	}
}

func TestServicePreviewsTranslationWithoutSaving(t *testing.T) {
	repo := NewMemoryRepository()
	service := NewService(repo, ai.Stub{}, "tr", time.Second)
	value, err := service.Create(context.Background(), "Başlık", "Açıklama", "tr", uuid.New())
	if err != nil {
		t.Fatal(err)
	}

	preview, err := service.PreviewTranslation(context.Background(), value.ID, "de")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Title != "[de] Başlık" || preview.Description != "[de] Açıklama" || preview.Source != SourceAI {
		t.Fatalf("preview = %#v", preview)
	}

	stored, err := repo.GetByID(context.Background(), value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := stored.Translations["de"]; exists {
		t.Fatalf("preview was persisted: %#v", stored.Translations["de"])
	}
}

func TestServicePreservesHumanTranslationUnlessForced(t *testing.T) {
	service := NewService(NewMemoryRepository(), ai.Stub{}, "tr", time.Second)
	value, err := service.Create(context.Background(), "Başlık", "Açıklama", "tr", uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddTranslation(context.Background(), value.ID, "en", "Human title", "Human description"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.TranslateWithAI(context.Background(), value.ID, []string{"en"}); err != nil {
		t.Fatal(err)
	}
	resolved, err := service.GetByID(context.Background(), value.ID, "en")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Translation.Source != SourceHuman || resolved.Translation.Title != "Human title" ||
		resolved.Translation.Description != "Human description" {
		t.Fatalf("human translation was overwritten: %#v", resolved.Translation)
	}
	if _, err := service.TranslateWithAI(context.Background(), value.ID, []string{"en"}, TranslateOptions{Force: true}); err != nil {
		t.Fatal(err)
	}
	resolved, err = service.GetByID(context.Background(), value.ID, "en")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Translation.Source != SourceAI || resolved.Translation.Title != "[en] Başlık" ||
		resolved.Translation.Description != "[en] Açıklama" {
		t.Fatalf("forced translation was not applied: %#v", resolved.Translation)
	}
}

func TestServiceSlugConflict(t *testing.T) {
	service := NewService(NewMemoryRepository(), ai.Stub{}, "en", time.Second)
	authorID := uuid.New()
	if _, err := service.Create(context.Background(), "ğüşiöç", "", "tr", authorID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(context.Background(), "gusioc", "", "tr", authorID); err != ErrConflict {
		t.Fatalf("conflict error = %v", err)
	}
}
