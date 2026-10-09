package topic

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/karabas/yakamoz/internal/platform/postgres"
	"github.com/karabas/yakamoz/internal/uuid"
)

type PostgresRepository struct {
	pool *postgres.Pool
}

func NewPostgresRepository(pool *postgres.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, value Topic) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO topics
		(id,slug,original_language,status,created_by,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, value.ID, value.Slug, value.OriginalLanguage, value.Status,
		value.CreatedBy, value.CreatedAt, value.UpdatedAt)
	if isUniqueViolation(err) {
		return ErrConflict
	}
	if err != nil {
		return err
	}
	for _, translation := range value.Translations {
		if _, err := tx.ExecContext(ctx, `INSERT INTO topic_translations
			(topic_id,language,title,description,source,translated_at)
			VALUES ($1,$2,$3,$4,$5,$6)`, value.ID, translation.Language, translation.Title,
			translation.Description, translation.Source, translation.TranslatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Topic, error) {
	return r.get(ctx, `WHERE t.id=$1`, id)
}

func (r *PostgresRepository) GetBySlug(ctx context.Context, slug string) (Topic, error) {
	return r.get(ctx, `WHERE t.slug=$1`, strings.ToLower(slug))
}

func (r *PostgresRepository) get(ctx context.Context, predicate string, arg any) (Topic, error) {
	var value Topic
	err := r.pool.QueryRowContext(ctx, `SELECT t.id,t.slug,t.original_language,t.status,t.created_by,t.created_at,t.updated_at
		FROM topics t `+predicate, arg).Scan(&value.ID, &value.Slug, &value.OriginalLanguage, &value.Status,
		&value.CreatedBy, &value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Topic{}, ErrNotFound
	}
	if err != nil {
		return Topic{}, err
	}
	value.Translations, err = r.translations(ctx, value.ID)
	return value, err
}

func (r *PostgresRepository) translations(ctx context.Context, id uuid.UUID) (map[string]Translation, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT language,title,description,source,translated_at
		FROM topic_translations WHERE topic_id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make(map[string]Translation)
	for rows.Next() {
		var value Translation
		if err := rows.Scan(&value.Language, &value.Title, &value.Description, &value.Source, &value.TranslatedAt); err != nil {
			return nil, err
		}
		values[value.Language] = value
	}
	return values, rows.Err()
}

func (r *PostgresRepository) List(ctx context.Context, filter ListFilter) ([]Topic, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT DISTINCT t.id,t.slug,t.original_language,t.status,t.created_by,t.created_at,t.updated_at,
		CASE WHEN $3='likes' THEN COALESCE(rc.likes,0) ELSE 0 END AS sort_likes
		FROM topics t LEFT JOIN topic_translations tt ON tt.topic_id=t.id
		LEFT JOIN (
			SELECT target_id,COUNT(*) FILTER (WHERE reaction='like') AS likes
			FROM reactions WHERE target_type='topic' GROUP BY target_id
		) rc ON rc.target_id=t.id
		WHERE ($1='' OR t.status=$1) AND ($2='' OR tt.language=$2)
		ORDER BY sort_likes DESC,t.created_at DESC LIMIT $4 OFFSET $5`, filter.Status, filter.Language, filter.Sort, limit, max(filter.Offset, 0))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]Topic, 0)
	for rows.Next() {
		var value Topic
		var sortLikes int
		if err := rows.Scan(&value.ID, &value.Slug, &value.OriginalLanguage, &value.Status, &value.CreatedBy,
			&value.CreatedAt, &value.UpdatedAt, &sortLikes); err != nil {
			return nil, err
		}
		value.Translations, err = r.translations(ctx, value.ID)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func (r *PostgresRepository) UpsertTranslation(ctx context.Context, id uuid.UUID, value Translation) error {
	_, err := r.pool.ExecContext(ctx, `INSERT INTO topic_translations
		(topic_id,language,title,description,source,translated_at) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (topic_id,language) DO UPDATE SET title=EXCLUDED.title,description=EXCLUDED.description,
		source=EXCLUDED.source,translated_at=EXCLUDED.translated_at`, id, value.Language, value.Title,
		value.Description, value.Source, value.TranslatedAt)
	return err
}

func (r *PostgresRepository) SetStatus(ctx context.Context, id uuid.UUID, status Status) (Topic, error) {
	updated := time.Now().UTC()
	result, err := r.pool.ExecContext(ctx, `UPDATE topics SET status=$2,updated_at=$3 WHERE id=$1`, id, status, updated)
	if err != nil {
		return Topic{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Topic{}, err
	}
	if affected == 0 {
		return Topic{}, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func max(value, minimum int) int {
	if value < minimum {
		return minimum
	}
	return value
}

func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint"))
}
