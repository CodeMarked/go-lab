package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codemarked/go-lab/go_CRUD_api/config"
	"github.com/gin-gonic/gin"
)

func TestCSRFSkipsWhenBearerPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		SessionCookieName: "gl_session",
		CSRFCookieName:    "gl_csrf",
		CSRFHeaderName:    "X-CSRF-Token",
	}
	r := gin.New()
	r.Use(CSRFCookieProtect(cfg))
	r.POST("/api/v1/users", func(c *gin.Context) { c.Status(http.StatusCreated) })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	req.Header.Set("Cookie", "gl_session=should-be-ignored-for-csrf")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestCSRFAllowsRegisterExemptPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		SessionCookieName: "gl_session",
		CSRFCookieName:    "gl_csrf",
		CSRFHeaderName:    "X-CSRF-Token",
	}
	r := gin.New()
	r.Use(CSRFCookieProtect(cfg))
	r.POST("/api/v1/auth/register", func(c *gin.Context) { c.Status(http.StatusCreated) })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d", rr.Code)
	}
}
