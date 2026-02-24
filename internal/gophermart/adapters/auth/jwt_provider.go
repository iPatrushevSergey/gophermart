package auth

import (
	"errors"
	"time"

	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/vo"

	"github.com/golang-jwt/jwt/v5"
)

var errInvalidToken = errors.New("invalid token")

// JWTProvider issues and validates JWT tokens.
type JWTProvider struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTProvider returns a new JWT token provider.
func NewJWTProvider(secret string, ttl time.Duration) *JWTProvider {
	return &JWTProvider{secret: []byte(secret), ttl: ttl}
}

type claims struct {
	UserID int64 `json:"sub"`
	jwt.RegisteredClaims
}

// Issue issues a JWT for the given user ID.
func (p *JWTProvider) Issue(userID vo.UserID) (string, error) {
	now := time.Now()
	c := claims{
		UserID: int64(userID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(p.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signedToken, err := unsignedToken.SignedString(p.secret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// Validate parses the token and returns the user ID.
func (p *JWTProvider) Validate(tokenString string) (vo.UserID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return p.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, errInvalidToken
	}
	c, ok := token.Claims.(*claims)
	if !ok {
		return 0, errInvalidToken
	}
	return vo.UserID(c.UserID), nil
}

var _ port.TokenProvider = (*JWTProvider)(nil)
