package usecases

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/polar-bear-cu/sgt-auth-service/models"
	"github.com/polar-bear-cu/sgt-auth-service/repositories"
	"github.com/polar-bear-cu/sgt-auth-service/tokens"
)

var ErrRefreshInvalid = errors.New("refresh token invalid or expired")

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type UserDirectory interface {
	FindOrCreateUser(ctx context.Context, email, googleSub, name, pictureURL string) (userID string, err error)
}

type AuthUsecase struct {
	oauthConfig *oauth2.Config
	users       UserDirectory
	jwtSecret   string
	accessTTL   time.Duration
	refreshTTL  time.Duration
	refresh     repositories.RefreshTokenRepository
}

func NewAuth(
	oauthConfig *oauth2.Config,
	users UserDirectory,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
	refresh repositories.RefreshTokenRepository,
) *AuthUsecase {
	return &AuthUsecase{
		oauthConfig: oauthConfig,
		users:       users,
		jwtSecret:   jwtSecret,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
		refresh:     refresh,
	}
}

func (u *AuthUsecase) LoginURL(state string) string {
	return u.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (u *AuthUsecase) HandleCallback(ctx context.Context, code string) (TokenPair, error) {
	tok, err := u.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return TokenPair{}, fmt.Errorf("exchange code: %w", err)
	}

	info, err := fetchGoogleUserInfo(ctx, u.oauthConfig.Client(ctx, tok))
	if err != nil {
		return TokenPair{}, err
	}

	userID, err := u.users.FindOrCreateUser(ctx, info.Email, info.Sub, info.Name, info.Picture)
	if err != nil {
		return TokenPair{}, fmt.Errorf("resolve user: %w", err)
	}

	return u.issue(ctx, userID, info.Email, info.Name, info.Picture)
}

func (u *AuthUsecase) Refresh(ctx context.Context, rawRefreshToken string) (TokenPair, error) {
	hash := hashToken(rawRefreshToken)
	rec, err := u.refresh.FindByHash(ctx, hash)
	if err != nil {
		return TokenPair{}, ErrRefreshInvalid
	}
	if !rec.IsActive(time.Now()) {
		return TokenPair{}, ErrRefreshInvalid
	}
	if err := u.refresh.Revoke(ctx, hash); err != nil {
		return TokenPair{}, err
	}

	return u.issue(ctx, rec.UserID, "", "", "")
}

func (u *AuthUsecase) Logout(ctx context.Context, rawRefreshToken string) error {
	err := u.refresh.Revoke(ctx, hashToken(rawRefreshToken))
	if errors.Is(err, repositories.ErrRefreshTokenNotFound) {
		return nil
	}
	return err
}

func (u *AuthUsecase) issue(ctx context.Context, userID, email, name, picture string) (TokenPair, error) {
	access, err := tokens.SignAccess(u.jwtSecret, userID, email, name, picture, u.accessTTL)
	if err != nil {
		return TokenPair{}, err
	}

	raw, hash := newRefreshToken()
	if _, err := u.refresh.Create(ctx, models.RefreshToken{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(u.refreshTTL),
	}); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int(u.accessTTL.Seconds()),
	}, nil
}

type googleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func fetchGoogleUserInfo(ctx context.Context, client *http.Client) (googleUserInfo, error) {
	const googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return googleUserInfo{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return googleUserInfo{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return googleUserInfo{}, fmt.Errorf("userinfo: status %d", resp.StatusCode)
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return googleUserInfo{}, err
	}
	return info, nil
}

func newRefreshToken() (raw, hash string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, hashToken(raw)
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
