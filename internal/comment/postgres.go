package comment

import (
	"context"
	"database/sql"
	"errors"

	"github.com/karabas/yakamoz/internal/platform/postgres"
	"github.com/karabas/yakamoz/internal/uuid"
)

type PostgresRepository struct {
	pool *postgres.Pool
}

func NewPostgresRepository(pool *postgres.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, value Comment) error {
	_, err := r.pool.ExecContext(ctx, `INSERT INTO comments
		(id, topic_id, author_id, language, body, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`, value.ID, value.TopicID, value.AuthorID, value.Language, value.Body, value.CreatedAt)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Comment, error) {
	var value Comment
	err := r.pool.QueryRowContext(ctx, `SELECT id,topic_id,author_id,language,body,created_at
		FROM comments WHERE id=$1`, id).Scan(&value.ID, &value.TopicID, &value.AuthorID,
		&value.Language, &value.Body, &value.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Comment{}, ErrNotFound
	}
	return value, err
}

func (r *PostgresRepository) ListByTopic(ctx context.Context, filter ListFilter) ([]Comment, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT id,topic_id,author_id,language,body,created_at
		FROM comments WHERE topic_id=$1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		filter.TopicID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]Comment, 0)
	for rows.Next() {
		var value Comment
		if err := rows.Scan(&value.ID, &value.TopicID, &value.AuthorID, &value.Language,
			&value.Body, &value.CreatedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.ExecContext(ctx, `DELETE FROM comments WHERE id=$1`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return ErrNotFound
	}
	return err
}
