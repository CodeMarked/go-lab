package middleware

import (
	"net/http"
	"strings"

	"github.com/codemarked/go-lab/go_CRUD_api/api"
	"github.com/codemarked/go-lab/go_CRUD_api/auth"
	"github.com/codemarked/go-lab/go_CRUD_api/requestid"
	"github.com/gin-gonic/gin"
)

// BearerAuth validates Authorization: Bearer <JWT> using the token service.
func BearerAuth(ts *auth.TokenService) gin.HandlerFunc {
	if ts == nil {
		return func(c *gin.Context) {
			writeAuthError(c, http.StatusInternalServerError, api.CodeInternal, "authentication not configured")
			c.Abort()
		}
	}
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			writeAuthError(c, http.StatusUnauthorized, api.CodeUnauthorized, "missing authorization header")
			c.Abort()
			return
		}
		const prefix = "Bearer "
		if len(raw) < len(prefix) || !strings.EqualFold(raw[:len(prefix)], prefix) {
			writeAuthError(c, http.StatusUnauthorized, api.CodeUnauthorized, "invalid authorization scheme")
			c.Abort()
			return
		}
		token := strings.TrimSpace(raw[len(prefix):])
		if token == "" {
			writeAuthError(c, http.StatusUnauthorized, api.CodeUnauthorized, "empty bearer token")
			c.Abort()
			return
		}
		claims, err := ts.ParseAccessToken(token)
		if err != nil {
			writeAuthError(c, http.StatusUnauthorized, api.CodeUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}
		sub := strings.TrimSpace(claims.Subject)
		if sub == "" {
			writeAuthError(c, http.StatusUnauthorized, api.CodeUnauthorized, "invalid token subject")
			c.Abort()
			return
		}
		c.Set("auth_subject", sub)
		c.Next()
	}
}

func writeAuthError(c *gin.Context, status int, code, message string) {
	c.JSON(status, api.ErrorEnvelope{
		Error: api.ErrBody{Code: code, Message: message},
		Meta:  api.Meta{RequestID: requestid.FromContext(c)},
	})
}
