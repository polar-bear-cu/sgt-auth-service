package repositories

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/polar-bear-cu/sgt-auth-service/models"
)

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository interface {
	Create(ctx context.Context, t models.RefreshToken) (models.RefreshToken, error)
	FindByHash(ctx context.Context, hash string) (models.RefreshToken, error)
	Revoke(ctx context.Context, hash string) error
}

type inMemoryRefreshToken struct {
	mu     sync.Mutex
	byHash map[string]models.RefreshToken
	seq    int
}

func NewInMemoryRefreshToken() RefreshTokenRepository {
	return &inMemoryRefreshToken{byHash: map[string]models.RefreshToken{}}
}

func (r *inMemoryRefreshToken) Create(_ context.Context, t models.RefreshToken) (models.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	t.ID = fmt.Sprintf("mem-%d", r.seq)
	t.CreatedAt = time.Now()
	r.byHash[t.TokenHash] = t
	return t, nil
}

func (r *inMemoryRefreshToken) FindByHash(_ context.Context, hash string) (models.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.byHash[hash]
	if !ok {
		return models.RefreshToken{}, ErrRefreshTokenNotFound
	}
	return t, nil
}

func (r *inMemoryRefreshToken) Revoke(_ context.Context, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.byHash[hash]
	if !ok {
		return ErrRefreshTokenNotFound
	}
	now := time.Now()
	t.RevokedAt = &now
	r.byHash[hash] = t
	return nil
}
