package repository

import (
	"context"
	"errors"
	"fmt"

	"tsuskills-user/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) CreateUserAndEmail(
	ctx context.Context,
	user domain.User,
	email domain.Email,
) (uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO users (id, name, password_hash, status, is_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Name, user.PasswordHash, string(user.Status),
		user.IsVerified, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert user: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO emails (user_id, addr, is_primary, is_verified)
		 VALUES ($1, $2, $3, $4)`,
		email.UserID, email.Addr, email.IsPrimary, email.IsVerified,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert email: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit tx: %w", err)
	}

	return user.ID, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, *domain.Email, error) {
	var u domain.User
	var statusStr string

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, password_hash, status, is_verified, created_at, updated_at, deleted_at
		 FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&u.ID, &u.Name, &u.PasswordHash, &statusStr, &u.IsVerified,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}
		return nil, nil, fmt.Errorf("query user by id: %w", err)
	}
	u.Status = domain.StatusUser(statusStr)

	var e domain.Email
	err = r.pool.QueryRow(ctx,
		`SELECT id, user_id, addr, is_primary, is_verified, verified_at
		 FROM emails WHERE user_id = $1 AND is_primary = TRUE LIMIT 1`, id,
	).Scan(&e.ID, &e.UserID, &e.Addr, &e.IsPrimary, &e.IsVerified, &e.VerifiedAt)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, fmt.Errorf("query email: %w", err)
	}

	return &u, &e, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, emailAddr string) (*domain.User, error) {
	var u domain.User
	var statusStr string

	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.name, u.password_hash, u.status, u.is_verified, u.created_at, u.updated_at
		 FROM users u
		 JOIN emails e ON e.user_id = u.id
		 WHERE e.addr = $1 AND u.deleted_at IS NULL
		 LIMIT 1`, emailAddr,
	).Scan(&u.ID, &u.Name, &u.PasswordHash, &statusStr, &u.IsVerified,
		&u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	u.Status = domain.StatusUser(statusStr)

	return &u, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET name = $1, password_hash = $2, status = $3,
		 is_verified = $4, updated_at = NOW() WHERE id = $5 AND deleted_at IS NULL`,
		user.Name, user.PasswordHash, string(user.Status),
		user.IsVerified, user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET deleted_at = NOW(), status = $1, updated_at = NOW()
		 WHERE id = $2 AND deleted_at IS NULL`,
		string(domain.StatusDeleted), id,
	)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) EmailExists(ctx context.Context, addr string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM emails WHERE addr = $1)`, addr,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return exists, nil
}
