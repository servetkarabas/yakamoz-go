package topic

import (
	"context"
	"strings"
	"time"

	"github.com/karabas/yakamoz/internal/platform/ai"
	"github.com/karabas/yakamoz/internal/platform/i18n"
	"github.com/karabas/yakamoz/internal/uuid"
)

type Service struct {
	repo        Repository
	translator  ai.Translator
	defaultLang string
	aiTimeout   time.Duration
	now         func() time.Time
}

type TranslateOptions struct {
	Force bool
}

func NewService(repo Repository, translator ai.Translator, defaultLanguage string, aiTimeout time.Duration) *Service {
	return &Service{repo: repo, translator: translator, defaultLang: defaultLanguage, aiTimeout: aiTimeout, now: time.Now}
}

func (s *Service) Create(ctx context.Context, title, description, language string, authorID uuid.UUID) (Topic, error) {
	language = i18n.Normalize(language)
	title = strings.TrimSpace(title)
	if title == "" || language == "" || authorID == uuid.Nil {
		return Topic{}, ErrInvalid
	}
	now := s.now().UTC()
	translation := Translation{Language: language, Title: title, Description: description, Source: SourceHuman, TranslatedAt: now}
	value := Topic{
		ID: uuid.New(), Slug: Slug(title), OriginalLanguage: language, Status: StatusDraft, CreatedBy: authorID,
		CreatedAt: now, UpdatedAt: now, Translations: map[string]Translation{language: translation},
	}
	if value.Slug == "" {
		return Topic{}, ErrInvalid
	}
	if err := s.repo.Create(ctx, value); err != nil {
		return Topic{}, err
	}
	return value, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID, requestedLanguage string) (Resolved, error) {
	value, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Resolved{}, err
	}
	return s.resolve(value, requestedLanguage), nil
}

func (s *Service) GetBySlug(ctx context.Context, slug, requestedLanguage string) (Resolved, error) {
	value, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return Resolved{}, err
	}
	return s.resolve(value, requestedLanguage), nil
}

func (s *Service) resolve(value Topic, requestedLanguage string) Resolved {
	available := make(map[string]bool, len(value.Translations))
	for language := range value.Translations {
		available[language] = true
	}
	served := i18n.Resolve(requestedLanguage, value.OriginalLanguage, s.defaultLang, available)
	return Resolved{Topic: value, Translation: value.Translations[served], ServedLanguage: served}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Topic, error) {
	filter.Language = i18n.Normalize(filter.Language)
	return s.repo.List(ctx, filter)
}

func (s *Service) AddTranslation(ctx context.Context, id uuid.UUID, language, title, description string) (Translation, error) {
	language = i18n.Normalize(language)
	if id == uuid.Nil || language == "" || strings.TrimSpace(title) == "" {
		return Translation{}, ErrInvalid
	}
	value := Translation{Language: language, Title: strings.TrimSpace(title), Description: description, Source: SourceHuman, TranslatedAt: s.now().UTC()}
	if err := s.repo.UpsertTranslation(ctx, id, value); err != nil {
		return Translation{}, err
	}
	return value, nil
}

func (s *Service) Publish(ctx context.Context, id uuid.UUID) (Topic, error) {
	return s.repo.SetStatus(ctx, id, StatusPublished)
}

func (s *Service) Archive(ctx context.Context, id uuid.UUID) (Topic, error) {
	return s.repo.SetStatus(ctx, id, StatusArchived)
}

func (s *Service) TranslateWithAI(ctx context.Context, id uuid.UUID, targetLanguages []string, options ...TranslateOptions) (Topic, error) {
	value, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Topic{}, err
	}
	force := len(options) > 0 && options[0].Force
	source := value.Translations[value.OriginalLanguage]
	for _, target := range targetLanguages {
		target = i18n.Normalize(target)
		if target == "" || target == value.OriginalLanguage {
			continue
		}
		existing, exists := value.Translations[target]
		if exists && existing.Source == SourceHuman && !force {
			continue
		}
		translation := Translation{Language: target, Source: SourceAI, TranslatedAt: s.now().UTC()}
		if source.Title != "" {
			translationContext, cancel := context.WithTimeout(ctx, s.aiTimeout)
			result, translateErr := s.translator.Translate(translationContext, ai.TranslateRequest{
				Text: source.Title, SourceLanguage: value.OriginalLanguage, TargetLanguage: target,
			})
			cancel()
			if translateErr != nil {
				return Topic{}, translateErr
			}
			translation.Title = result.Text
		}
		if source.Description != "" {
			translationContext, cancel := context.WithTimeout(ctx, s.aiTimeout)
			result, translateErr := s.translator.Translate(translationContext, ai.TranslateRequest{
				Text: source.Description, SourceLanguage: value.OriginalLanguage, TargetLanguage: target,
			})
			cancel()
			if translateErr != nil {
				return Topic{}, translateErr
			}
			translation.Description = result.Text
		}
		if err := s.repo.UpsertTranslation(ctx, id, translation); err != nil {
			return Topic{}, err
		}
		value.Translations[target] = translation
	}
	return value, nil
}
