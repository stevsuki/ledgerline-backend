package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/apierr"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

const (
	ContextClaims = "auth_claims"
	bearerPrefix  = "Bearer "
)

// Authenticate: verify the Bearer token, store claims in the context.
func Authenticate(tokenManager domain.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.GetHeader("Authorization"))
		if token == "" {
			apierr.Write(c, domain.ErrTokenMissing)
			return
		}

		claims, err := tokenManager.Verify(token)
		if err != nil {
			if !errors.Is(err, domain.ErrTokenExpired) {
				err = domain.ErrTokenInvalid.WithCause(err)
			}
			apierr.Write(c, err)
			return
		}

		c.Set(ContextClaims, claims)
		c.Next()
	}
}

// RequireRoles: restrict by role id (never name), mounted after Authenticate.
func RequireRoles(roleIDs ...uuid.UUID) gin.HandlerFunc {
	allowed := make(map[uuid.UUID]struct{}, len(roleIDs))
	for _, id := range roleIDs {
		allowed[id] = struct{}{}
	}

	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			apierr.Write(c, domain.ErrAuthRequired)
			return
		}
		if _, ok := allowed[claims.RoleID]; !ok {
			apierr.Write(c, domain.ErrAccessDenied)
			return
		}
		c.Next()
	}
}

// extractToken accepts only "Bearer <token>" per RFC 6750, rejecting anything else.
func extractToken(header string) string {
	rest, found := strings.CutPrefix(strings.TrimSpace(header), bearerPrefix)
	if !found {
		return ""
	}
	return strings.TrimSpace(rest)
}

// GetClaims from the context.
func GetClaims(c *gin.Context) (*domain.TokenClaims, bool) {
	v, exists := c.Get(ContextClaims)
	if !exists {
		return nil, false
	}
	claims, ok := v.(*domain.TokenClaims)
	return claims, ok
}

// GetUserID of the currently logged-in user.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	claims, ok := GetClaims(c)
	if !ok {
		return uuid.Nil, false
	}
	return claims.UserID, true
}
