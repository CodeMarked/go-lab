package myhandlers

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/codemarked/go-lab/go_CRUD_api/api"
	"github.com/codemarked/go-lab/go_CRUD_api/auth"
	"github.com/codemarked/go-lab/go_CRUD_api/config"
	"github.com/codemarked/go-lab/go_CRUD_api/requestid"
	"github.com/codemarked/go-lab/go_CRUD_api/respond"
	"github.com/gin-gonic/gin"
)

// TokenSvc is set from main after the auth service is constructed.
var TokenSvc *auth.TokenService

// IssueToken handles POST /api/v1/auth/token (client_credentials).
func IssueToken(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			GrantType    string `json:"grant_type" binding:"required"`
			ClientID     string `json:"client_id" binding:"required"`
			ClientSecret string `json:"client_secret" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			respond.Error(c, http.StatusBadRequest, api.CodeValidation, "invalid request body", map[string]any{"field": "body"})
			return
		}
		if strings.TrimSpace(body.GrantType) != "client_credentials" {
			respond.Error(c, http.StatusBadRequest, api.CodeValidation, "unsupported grant_type", map[string]any{"grant_type": body.GrantType})
			return
		}
		if !secureStringEq(strings.TrimSpace(body.ClientID), cfg.PlatformClientID) ||
			!secureStringEq(body.ClientSecret, cfg.PlatformClientSecret) {
			respond.Error(c, http.StatusUnauthorized, api.CodeUnauthorized, "invalid client credentials", nil)
			return
		}
		if TokenSvc == nil {
			respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "token service not configured", nil)
			return
		}
		issueTokenResponse(c, cfg.PlatformClientID)
	}
}

// IssueBootstrapToken handles POST /api/v1/auth/bootstrap.
// It mints a token server-side without exposing platform credentials to the SPA.
//
// TODO(core-auth): remove this endpoint once all SPAs use POST /api/v1/auth/login + cookie sessions.
// Gate with AUTH_BOOTSTRAP_ENABLED=false for operators ready to enforce real user auth only.
func IssueBootstrapToken(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.AuthBootstrapEnabled {
			respond.Error(c, http.StatusForbidden, api.CodeForbidden, "bootstrap auth disabled; use user login", nil)
			return
		}
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "" {
			respond.Error(c, http.StatusForbidden, api.CodeForbidden, "origin required", nil)
			return
		}
		if !originAllowed(origin, cfg.CORSAllowedOrigins) {
			slog.Warn("auth_bootstrap_denied",
				"request_id", requestid.FromContext(c),
				"origin", origin,
				"reason", "origin_not_allowed",
			)
			respond.Error(c, http.StatusForbidden, api.CodeForbidden, "origin not allowed", nil)
			return
		}
		slog.Info("auth_bootstrap_used",
			"request_id", requestid.FromContext(c),
			"origin", origin,
			"path", c.Request.URL.Path,
			"client_ip", strings.TrimSpace(c.ClientIP()),
			"deprecation", "temporary_bridge",
			"replacement_flow", "POST /api/v1/auth/login",
		)
		if AuthStore != nil {
			ctx := c.Request.Context()
			_ = AuthStore.InsertAudit(ctx, "auth_bootstrap_used", nil, clientIP(c), c.Request.UserAgent(), origin, nil)
		}
		issueBootstrapTokenResponse(c, cfg.PlatformClientID)
	}
}

func issueTokenResponse(c *gin.Context, clientID string) {
	if TokenSvc == nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "token service not configured", nil)
		return
	}
	subject := "client:" + clientID
	rawToken, exp, err := TokenSvc.MintAccessToken(subject)
	if err != nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to issue token", nil)
		return
	}
	expiresIn := int(time.Until(exp).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}
	respond.JSONOK(c, http.StatusOK, gin.H{
		"access_token": rawToken,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
	})
}

func issueBootstrapTokenResponse(c *gin.Context, clientID string) {
	if TokenSvc == nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "token service not configured", nil)
		return
	}
	subject := "client:" + clientID
	rawToken, exp, err := TokenSvc.MintAccessToken(subject)
	if err != nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to issue token", nil)
		return
	}
	expiresIn := int(time.Until(exp).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}
	respond.JSONOK(c, http.StatusOK, gin.H{
		"access_token": rawToken,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
		"bootstrap": gin.H{
			"temporary":      true,
			"deprecated":     true,
			"migrate_to":     "POST /api/v1/auth/login",
			"disable_env":    "AUTH_BOOTSTRAP_ENABLED=false",
			"session_cookie": "gl_session (see SESSION_COOKIE_NAME)",
		},
	})
}

func originAllowed(origin string, allowlist []string) bool {
	for _, allowed := range allowlist {
		if strings.EqualFold(strings.TrimSpace(allowed), origin) {
			return true
		}
	}
	return false
}

func secureStringEq(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
