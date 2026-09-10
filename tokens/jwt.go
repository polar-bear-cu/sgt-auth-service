package tokens

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Access-token payload
type Claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

func SignAccess(secret, userID, email string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		Email: email,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

func ParseAccess(secret, raw string) (*Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(raw, &claims, func(*jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	return &claims, nil
}
