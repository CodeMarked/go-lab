package myhandlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codemarked/go-lab/api/api"
	"github.com/codemarked/go-lab/api/authstore"
	"github.com/codemarked/go-lab/api/middleware"
	"github.com/codemarked/go-lab/api/platformrbac"
	"github.com/codemarked/go-lab/api/requestid"
	"github.com/codemarked/go-lab/api/respond"
	"github.com/gin-gonic/gin"
)

// ListPlayersStub returns an empty list; authoritative player/gameplay data lives in suite repos (see docs/data-ownership.md).
func ListPlayersStub(c *gin.Context) {
	respond.OK(c, gin.H{
		"items":       []any{},
		"note":        "Stub; link players to platform users via platform_user_id in suite databases.",
		"permissions": []string{platformrbac.PermPlayersRead},
	})
}

// ListCharactersStub returns an empty list; Marble owns authoritative character state.
func ListCharactersStub(c *gin.Context) {
	respond.OK(c, gin.H{
		"items":       []any{},
		"note":        "Stub; characters are Marble-authoritative per data-ownership.md.",
		"permissions": []string{platformrbac.PermCharactersRead},
	})
}

type securityMeRolesData struct {
	UserID               int      `json:"user_id"`
	Roles                []string `json:"roles"`
	EffectivePermissions []string `json:"effective_permissions"`
}

// GetSecurityMeRoles returns the caller's platform roles and derived permission set.
func GetSecurityMeRoles(c *gin.Context) {
	uid, ok := middleware.AuthUserIDFromContext(c)
	if !ok {
		respond.Error(c, http.StatusForbidden, api.CodeForbidden, "user subject required", nil)
		return
	}
	rolesAny, _ := c.Get("platform_roles")
	roles, _ := rolesAny.([]string)
	var perms []string
	seen := make(map[string]struct{})
	for _, r := range roles {
		for _, p := range platformrbac.PermissionsForRole(r) {
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			perms = append(perms, p)
		}
	}
	respond.OK(c, securityMeRolesData{
		UserID:               uid,
		Roles:                roles,
		EffectivePermissions: perms,
	})
}

type adminAuditEventRow struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	ActorUserID  *int      `json:"actor_user_id,omitempty"`
	AuthSubject  string    `json:"auth_subject"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   *string   `json:"resource_id,omitempty"`
	Reason       *string   `json:"reason,omitempty"`
	RequestID    *string   `json:"request_id,omitempty"`
}

// ListAdminAuditEvents returns recent immutable control-plane audit rows (newest first).
func ListAdminAuditEvents(c *gin.Context) {
	const limit = 100
	rows, err := Database.Db.QueryContext(c.Request.Context(),
		`SELECT id, created_at, actor_user_id, auth_subject, action, resource_type, resource_id, reason, request_id
		 FROM admin_audit_events
		 ORDER BY id DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to list admin audit events", nil)
		return
	}
	defer func() { _ = rows.Close() }()

	out := make([]adminAuditEventRow, 0)
	for rows.Next() {
		var r adminAuditEventRow
		var actor sql.NullInt64
		var resID, reason, reqID sql.NullString
		if err := rows.Scan(&r.ID, &r.CreatedAt, &actor, &r.AuthSubject, &r.Action, &r.ResourceType, &resID, &reason, &reqID); err != nil {
			respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to decode admin audit events", nil)
			return
		}
		if actor.Valid {
			v := int(actor.Int64)
			r.ActorUserID = &v
		}
		if resID.Valid {
			s := resID.String
			r.ResourceID = &s
		}
		if reason.Valid {
			s := reason.String
			r.Reason = &s
		}
		if reqID.Valid {
			s := reqID.String
			r.RequestID = &s
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to list admin audit events", nil)
		return
	}
	respond.OK(c, gin.H{"items": out, "limit": limit})
}

// GetEconomyLedger lists append-only economy ledger rows (read-only; newest id first).
func GetEconomyLedger(c *gin.Context) {
	limit := 50
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			respond.Error(c, http.StatusBadRequest, api.CodeValidation, "limit must be between 1 and 100", nil)
			return
		}
		limit = n
	}

	var beforeID *int64
	if v := strings.TrimSpace(c.Query("before_id")); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 1 {
			respond.Error(c, http.StatusBadRequest, api.CodeValidation, "before_id must be a positive integer", nil)
			return
		}
		beforeID = &n
	}

	var platformUserID *int
	if v := strings.TrimSpace(c.Query("platform_user_id")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			respond.Error(c, http.StatusBadRequest, api.CodeValidation, "platform_user_id must be a positive integer", nil)
			return
		}
		platformUserID = &n
	}

	eventType := strings.TrimSpace(c.Query("event_type"))
	if len(eventType) > 64 {
		respond.Error(c, http.StatusBadRequest, api.CodeValidation, "event_type too long", nil)
		return
	}

	var fromTime, toTime *time.Time
	if v := strings.TrimSpace(c.Query("from")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			respond.Error(c, http.StatusBadRequest, api.CodeValidation, "from must be RFC3339 date-time", nil)
			return
		}
		fromTime = &t
	}
	if v := strings.TrimSpace(c.Query("to")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			respond.Error(c, http.StatusBadRequest, api.CodeValidation, "to must be RFC3339 date-time", nil)
			return
		}
		toTime = &t
	}

	rows, err := AuthStore.ListEconomyLedgerEvents(c.Request.Context(), authstore.EconomyLedgerQuery{
		Limit:          limit,
		BeforeID:       beforeID,
		PlatformUserID: platformUserID,
		EventType:      eventType,
		FromTime:       fromTime,
		ToTime:         toTime,
	})
	if err != nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to list economy ledger events", nil)
		return
	}

	type ledgerItem struct {
		ID             int64           `json:"id"`
		CreatedAt      time.Time       `json:"created_at"`
		PlatformUserID int             `json:"platform_user_id"`
		EventType      string          `json:"event_type"`
		AmountDelta    int64           `json:"amount_delta"`
		CurrencyCode   string          `json:"currency_code"`
		ReferenceType  *string         `json:"reference_type,omitempty"`
		ReferenceID    *string         `json:"reference_id,omitempty"`
		Meta           json.RawMessage `json:"meta,omitempty"`
	}
	items := make([]ledgerItem, 0, len(rows))
	for _, r := range rows {
		it := ledgerItem{
			ID:             r.ID,
			CreatedAt:      r.CreatedAt,
			PlatformUserID: r.PlatformUserID,
			EventType:      r.EventType,
			AmountDelta:    r.AmountDelta,
			CurrencyCode:   r.CurrencyCode,
		}
		if r.ReferenceType.Valid {
			s := r.ReferenceType.String
			it.ReferenceType = &s
		}
		if r.ReferenceID.Valid {
			s := r.ReferenceID.String
			it.ReferenceID = &s
		}
		if len(r.MetaJSON) > 0 {
			it.Meta = json.RawMessage(append([]byte(nil), r.MetaJSON...))
		}
		items = append(items, it)
	}
	respond.OK(c, gin.H{"items": items, "limit": limit})
}

// SupportAckRequest is the JSON body for POST /support/ack (reason may also be sent via X-Platform-Action-Reason).
type SupportAckRequest struct {
	Message string `json:"message"`
}

// PostSupportAck records a privileged support acknowledgment in admin_audit_events (mutation example).
func PostSupportAck(c *gin.Context) {
	uid, ok := middleware.AuthUserIDFromContext(c)
	if !ok {
		respond.Error(c, http.StatusForbidden, api.CodeForbidden, "user subject required", nil)
		return
	}
	reason := strings.TrimSpace(c.GetHeader(middleware.PlatformActionReasonHeader))
	if reason == "" {
		if v, ok := c.Get("platform_action_reason"); ok {
			if s, ok := v.(string); ok {
				reason = strings.TrimSpace(s)
			}
		}
	}
	var body SupportAckRequest
	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&body) // optional extra message
	}
	meta := map[string]any{}
	if strings.TrimSpace(body.Message) != "" {
		meta["message"] = strings.TrimSpace(body.Message)
	}
	var metaJSON []byte
	if len(meta) > 0 {
		var err error
		metaJSON, err = json.Marshal(meta)
		if err != nil {
			respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to encode audit meta", nil)
			return
		}
	}
	sub := c.GetString("auth_subject")
	err := AuthStore.InsertAdminAuditEvent(c.Request.Context(), &uid, sub, "support.ack", "support", "",
		reason, requestid.FromContext(c), c.ClientIP(), c.GetHeader("User-Agent"), metaJSON)
	if err != nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to write admin audit event", nil)
		return
	}
	respond.OK(c, gin.H{"ok": true, "recorded": "support.ack"})
}
