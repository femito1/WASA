// File: service/api/auth.go
package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// jwtSecret is the signing key for JWT tokens.
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// CustomClaims defines the structure of our JWT claims.
type CustomClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token for the given user ID.
// The token will be valid for 24 hours.
func GenerateToken(userID uint64) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ExtractUserIDFromToken extracts the user ID from the request's Authorization header.
// It expects the header to be of the form "Bearer <token>".
func ExtractUserIDFromToken(r *http.Request) (uint64, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, errors.New("missing Authorization header")
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return 0, errors.New("invalid Authorization header format")
	}
	tokenStr := parts[1]
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims.UserID, nil
	}
	return 0, errors.New("invalid token")
}
