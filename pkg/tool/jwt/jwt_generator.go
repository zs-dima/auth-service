package jwt_tool

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	model "github.com/zs-dima/auth-service/internal/gen/db"

	"github.com/golang-jwt/jwt"

	"github.com/google/uuid"
)

type ITokenGenerator interface {
	GenerateAccessToken(
		user *model.User,
		deviceId *uuid.UUID,
		installationId *uuid.UUID,
		jwtSecretKey string,
	) (string, error)
	GenerateRefreshToken() (string, time.Time, error)
}

type TokenGenerator struct{}

func (gen *TokenGenerator) GenerateAccessToken(
	user *model.User,
	deviceId *uuid.UUID,
	installationId *uuid.UUID,
	jwtSecretKey string,
) (string, error) {
	claims := &jwt.MapClaims{
		"sub":          user.Name,
		"aud":          Audience,
		"iss":          Issuer,
		"jti":          uuid.New().String(),
		"role":         user.Role,
		"userEmail":    user.Email,
		"userId":       user.ID,
		"device":       deviceId,
		"installation": installationId,
		"exp":          time.Now().Add(time.Hour * 24).Unix(), // Token expires after 24 hours
	}
	// Audience aud; IssuedAt int64 iat; Issuer iss; NotBefore int64 nbf
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtSecretKey))
}

// GenerateToken generates a base64 encoded securely random string
func (gen *TokenGenerator) GenerateRefreshToken() (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Hour * 24 * 7) // Refresh token expires after 7 days

	b := make([]byte, TokenLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", expiresAt, fmt.Errorf("error generating token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(b)
	return token, expiresAt, nil
}

func ExtractAuthInfo(ctx context.Context) *JwtAuthInfo {
	return ctx.Value(UserClaimsKey).(*JwtAuthInfo)
}
