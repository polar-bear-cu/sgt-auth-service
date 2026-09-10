package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/polar-bear-cu/sgt-auth-service/models"
)

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository interface {
	Create(ctx context.Context, t models.RefreshToken) (models.RefreshToken, error)
	FindByHash(ctx context.Context, hash string) (models.RefreshToken, error)
	Revoke(ctx context.Context, hash string) error
}

type RefreshTokenPostgres struct {
	db *pgxpool.Pool
}

func NewRefreshTokenPostgres(db *pgxpool.Pool) RefreshTokenRepository {
	return &RefreshTokenPostgres{db: db}
}

func (r *RefreshTokenPostgres) Create(ctx context.Context, t models.RefreshToken) (models.RefreshToken, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3) RETURNING id, created_at`,
		t.UserID, t.TokenHash, t.ExpiresAt,
	).Scan(&t.ID, &t.CreatedAt)
	return t, err
}

func (r *RefreshTokenPostgres) FindByHash(ctx context.Context, hash string) (models.RefreshToken, error) {
	var t models.RefreshToken
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		 FROM refresh_tokens WHERE token_hash = $1`, hash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.RefreshToken{}, ErrRefreshTokenNotFound
	}
	return t, err
}

func (r *RefreshTokenPostgres) Revoke(ctx context.Context, hash string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now()
		 WHERE token_hash = $1 AND revoked_at IS NULL`, hash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrRefreshTokenNotFound
	}
	return nil
}
