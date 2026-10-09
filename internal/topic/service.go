package topic

import (
	"context"
	"strings"
	"sync"
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
	slug := Slug(title)
	if title == "" || language == "" || authorID == uuid.Nil || slug == "" {
		return Topic{}, ErrInvalid
	}
	now := s.now().UTC()
	translation := Translation{Language: language, Title: title, Description: description, Source: SourceHuman, TranslatedAt: now}
	translations := map[string]Translation{language: translation}
	if language != "en" {
		english, err := s.translateFields(ctx, translation, "en")
		if err != nil {
			return Topic{}, err
		}
		translations["en"] = english
	}
	value := Topic{
		ID: uuid.New(), Slug: slug, OriginalLanguage: language, Status: StatusDraft, CreatedBy: authorID,
		CreatedAt: now, UpdatedAt: now, Translations: translations,
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
	value, err = s.ensureTranslation(ctx, value, requestedLanguage)
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
	value, err = s.ensureTranslation(ctx, value, requestedLanguage)
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

func (s *Service) ensureTranslation(ctx context.Context, value Topic, language string) (Topic, error) {
	language = i18n.Normalize(language)
	if language == "" || !i18n.IsSupported(language) || language == value.OriginalLanguage {
		return value, nil
	}
	if _, ok := value.Translations[language]; ok {
		return value, nil
	}
	translation, err := s.translateFields(ctx, value.Translations[value.OriginalLanguage], language)
	if err != nil {
		return Topic{}, err
	}
	if err := s.repo.UpsertTranslation(ctx, value.ID, translation); err != nil {
		return Topic{}, err
	}
	if value.Translations == nil {
		value.Translations = make(map[string]Translation)
	}
	value.Translations[language] = translation
	return value, nil
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]Topic, error) {
	language := i18n.Normalize(filter.Language)
	filter.Language = ""
	if filter.Sort != "likes" {
		filter.Sort = "newest"
	}
	values, err := s.repo.List(ctx, filter)
	if err != nil || language == "" {
		return values, err
	}

	semaphore := make(chan struct{}, 4)
	translationErrors := make(chan error, len(values))
	var workers sync.WaitGroup
	for index := range values {
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				translationErrors <- ctx.Err()
				return
			}
			defer func() { <-semaphore }()
			translated, err := s.ensureTranslation(ctx, values[index], language)
			if err != nil {
				translationErrors <- err
				return
			}
			values[index] = translated
		}(index)
	}
	workers.Wait()
	close(translationErrors)
	for err := range translationErrors {
		if err != nil {
			return nil, err
		}
	}
	return values, nil
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

func (s *Service) PreviewTranslation(ctx context.Context, id uuid.UUID, language string) (Translation, error) {
	language = i18n.Normalize(language)
	if id == uuid.Nil || language == "" {
		return Translation{}, ErrInvalid
	}
	value, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Translation{}, err
	}
	if language == value.OriginalLanguage {
		return Translation{}, ErrInvalid
	}
	return s.translateFields(ctx, value.Translations[value.OriginalLanguage], language)
}

func (s *Service) translateFields(ctx context.Context, source Translation, target string) (Translation, error) {
	translation := Translation{Language: target, Source: SourceAI, TranslatedAt: s.now().UTC()}
	for _, field := range []struct {
		source string
		dest   *string
	}{{source.Title, &translation.Title}, {source.Description, &translation.Description}} {
		if field.source == "" {
			continue
		}
		translationContext, cancel := context.WithTimeout(ctx, s.aiTimeout)
		result, err := s.translator.Translate(translationContext, ai.TranslateRequest{
			Text: field.source, SourceLanguage: source.Language, TargetLanguage: target,
		})
		cancel()
		if err != nil {
			return Translation{}, err
		}
		*field.dest = result.Text
	}
	return translation, nil
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
		translation, err := s.translateFields(ctx, source, target)
		if err != nil {
			return Topic{}, err
		}
		if err := s.repo.UpsertTranslation(ctx, id, translation); err != nil {
			return Topic{}, err
		}
		value.Translations[target] = translation
	}
	return value, nil
}
