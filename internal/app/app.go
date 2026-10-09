package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/karabas/yakamoz/internal/author"
	"github.com/karabas/yakamoz/internal/comment"
	"github.com/karabas/yakamoz/internal/config"
	"github.com/karabas/yakamoz/internal/platform/ai"
	"github.com/karabas/yakamoz/internal/platform/httpx"
	"github.com/karabas/yakamoz/internal/platform/postgres"
	"github.com/karabas/yakamoz/internal/topic"
)

func NewRouter(cfg config.Config, logger *slog.Logger, authorRepo author.Repository, topicRepo topic.Repository, commentRepo comment.Repository, pool *postgres.Pool, reactionRepo reactionStore) http.Handler {
	authorService := author.NewService(authorRepo)
	topicService := topic.NewService(topicRepo, ai.Stub{}, cfg.DefaultLanguage, cfg.AITimeout)
	commentService := comment.NewService(commentRepo)
	authors := author.NewHandler(authorService)
	topics := topic.NewHandler(topicService, cfg.DefaultLanguage)
	comments := comment.NewHandler(commentService)
	reactions := reactionHandler{store: reactionRepo}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if pool != nil {
			if err := pool.Ping(r.Context()); err != nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, "not_ready", "database is unavailable", nil)
				return
			}
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST /api/v1/authors", authors.Create)
	mux.HandleFunc("GET /api/v1/authors", authors.List)
	mux.HandleFunc("GET /api/v1/authors/{id}", authors.GetByID)
	mux.HandleFunc("GET /api/v1/authors/by-nickname/{nickname}", authors.GetByNickname)
	mux.HandleFunc("PATCH /api/v1/authors/{id}", authors.Update)
	mux.HandleFunc("DELETE /api/v1/authors/{id}", authors.Delete)
	mux.HandleFunc("POST /api/v1/topics", topics.Create)
	mux.HandleFunc("GET /api/v1/topics", topics.List)
	mux.HandleFunc("GET /api/v1/topics/{id}", topics.GetByID)
	mux.HandleFunc("GET /api/v1/topics/by-slug/{slug}", topics.GetBySlug)
	mux.HandleFunc("PUT /api/v1/topics/{id}/translations/{lang}", topics.AddTranslation)
	mux.HandleFunc("POST /api/v1/topics/{id}/translations/{lang}/preview", topics.PreviewTranslation)
	mux.HandleFunc("POST /api/v1/topics/{id}/translate", topics.Translate)
	mux.HandleFunc("POST /api/v1/topics/{id}/publish", topics.Publish)
	mux.HandleFunc("POST /api/v1/comments", comments.Create)
	mux.HandleFunc("GET /api/v1/comments", comments.ListByTopic)
	mux.HandleFunc("GET /api/v1/reactions", reactions.List)
	mux.HandleFunc("PUT /api/v1/reactions", reactions.Set)
	mux.HandleFunc("GET /api/v1/comments/counts", comments.CountByTopics)
	mux.HandleFunc("GET /api/v1/comments/{id}", comments.GetByID)
	mux.HandleFunc("DELETE /api/v1/comments/{id}", comments.Delete)

	return httpx.Middleware(logger, mux)
}

func NewMemory(cfg config.Config, logger *slog.Logger) http.Handler {
	authorRepo := author.NewMemoryRepository()
	topicRepo := topic.NewMemoryRepository()
	commentRepo := comment.NewMemoryRepository()
	seed(context.Background(), cfg, logger, authorRepo, topicRepo, commentRepo)
	return NewRouter(cfg, logger, authorRepo, topicRepo, commentRepo, nil, newMemoryReactionStore())
}

func NewPostgres(ctx context.Context, cfg config.Config, logger *slog.Logger, pool *postgres.Pool) http.Handler {
	return NewRouter(cfg, logger, author.NewPostgresRepository(pool), topic.NewPostgresRepository(pool), comment.NewPostgresRepository(pool), pool, newPostgresReactionStore(pool))
}
