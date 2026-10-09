package author

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

func (r *PostgresRepository) Create(ctx context.Context, value Author) error {
	_, err := r.pool.ExecContext(ctx, `INSERT INTO authors
		(id, nickname, email, bio, preferred_language, role, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, value.ID, value.Nickname, value.Email, value.Bio,
		value.PreferredLanguage, value.Role, value.Status, value.CreatedAt, value.UpdatedAt)
	if isUniqueViolation(err) {
		return ErrConflict
	}
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Author, error) {
	return r.scanOne(ctx, `SELECT id,nickname,email,bio,preferred_language,role,status,created_at,updated_at
		FROM authors WHERE id=$1`, id)
}

func (r *PostgresRepository) GetByNickname(ctx context.Context, nickname string) (Author, error) {
	return r.scanOne(ctx, `SELECT id,nickname,email,bio,preferred_language,role,status,created_at,updated_at
		FROM authors WHERE lower(nickname)=lower($1)`, nickname)
}

func (r *PostgresRepository) scanOne(ctx context.Context, query string, arg any) (Author, error) {
	var value Author
	err := r.pool.QueryRowContext(ctx, query, arg).Scan(&value.ID, &value.Nickname, &value.Email, &value.Bio,
		&value.PreferredLanguage, &value.Role, &value.Status, &value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Author{}, ErrNotFound
	}
	return value, err
}

func (r *PostgresRepository) List(ctx context.Context, filter ListFilter) ([]Author, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT id,nickname,email,bio,preferred_language,role,status,created_at,updated_at
		FROM authors ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, max(filter.Offset, 0))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]Author, 0)
	for rows.Next() {
		var value Author
		if err := rows.Scan(&value.ID, &value.Nickname, &value.Email, &value.Bio, &value.PreferredLanguage,
			&value.Role, &value.Status, &value.CreatedAt, &value.UpdatedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, update Update) (Author, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return Author{}, err
	}
	if update.Bio != nil {
		current.Bio = *update.Bio
	}
	if update.PreferredLanguage != nil {
		current.PreferredLanguage = *update.PreferredLanguage
	}
	if update.Role != nil {
		current.Role = *update.Role
	}
	current.UpdatedAt = time.Now().UTC()
	_, err = r.pool.ExecContext(ctx, `UPDATE authors SET bio=$2,preferred_language=$3,role=$4,updated_at=$5 WHERE id=$1`,
		id, current.Bio, current.PreferredLanguage, current.Role, current.UpdatedAt)
	if isUniqueViolation(err) {
		return Author{}, ErrConflict
	}
	return current, err
}

func (r *PostgresRepository) SetStatus(ctx context.Context, id uuid.UUID, status Status) (Author, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return Author{}, err
	}
	current.Status = status
	current.UpdatedAt = time.Now().UTC()
	_, err = r.pool.ExecContext(ctx, `UPDATE authors SET status=$2,updated_at=$3 WHERE id=$1`, id, status, current.UpdatedAt)
	return current, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.ExecContext(ctx, `DELETE FROM authors WHERE id=$1`, id)
	return err
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
