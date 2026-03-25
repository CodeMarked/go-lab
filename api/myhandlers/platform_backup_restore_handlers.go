package myhandlers

import (
	"encoding/json"
	"errors"
	"log/slog"
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

// GetBackupsStatus returns restore-request workflow counts (Phase C governance; physical backup runs are operator-owned).
func GetBackupsStatus(c *gin.Context) {
	counts, err := AuthStore.BackupRestoreStatusCounts(c.Request.Context())
	if err != nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to read backup restore status", nil)
		return
	}
	pending := counts[authstore.BackupRestoreStatusPending]
	approved := counts[authstore.BackupRestoreStatusApproved]
	respond.OK(c, gin.H{
		"restore_workflow_enabled":      true,
		"pending_restore_requests":      pending,
		"approved_awaiting_fulfillment": approved,
		"counts_by_status":              counts,
		"note":                          "Physical backups/restores are run out of band; this API tracks approval workflow only.",
		"permissions":                   []string{platformrbac.PermBackupsRead},
	})
}

type backupRestoreRequestJSON struct {
	ID                int64           `json:"id"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	RequestedByUserID int             `json:"requested_by_user_id"`
	Scope             string          `json:"scope"`
	RestorePointLabel string          `json:"restore_point_label"`
	Reason            string          `json:"reason"`
	Status            string          `json:"status"`
	RejectionReason   *string         `json:"rejection_reason,omitempty"`
	Approval1UserID   *int            `json:"approval_1_user_id,omitempty"`
	Approval1At       *time.Time      `json:"approval_1_at,omitempty"`
	Approval2UserID   *int            `json:"approval_2_user_id,omitempty"`
	Approval2At       *time.Time      `json:"approval_2_at,omitempty"`
	FulfilledAt       *time.Time      `json:"fulfilled_at,omitempty"`
	FulfilledNote     *string         `json:"fulfilled_note,omitempty"`
	RequestID         *string         `json:"request_id,omitempty"`
}

func backupRestoreRowToJSON(r authstore.BackupRestoreRequestRow) backupRestoreRequestJSON {
	out := backupRestoreRequestJSON{
		ID:                r.ID,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
		RequestedByUserID: r.RequestedByUserID,
		Scope:             r.Scope,
		RestorePointLabel: r.RestorePointLabel,
		Reason:            r.Reason,
		Status:            r.Status,
	}
	if r.RejectionReason.Valid {
		s := r.RejectionReason.String
		out.RejectionReason = &s
	}
	if r.Approval1UserID.Valid {
		v := int(r.Approval1UserID.Int64)
		out.Approval1UserID = &v
	}
	if r.Approval1At.Valid {
		t := r.Approval1At.Time
		out.Approval1At = &t
	}
	if r.Approval2UserID.Valid {
		v := int(r.Approval2UserID.Int64)
		out.Approval2UserID = &v
	}
	if r.Approval2At.Valid {
		t := r.Approval2At.Time
		out.Approval2At = &t
	}
	if r.FulfilledAt.Valid {
		t := r.FulfilledAt.Time
		out.FulfilledAt = &t
	}
	if r.FulfilledNote.Valid {
		s := r.FulfilledNote.String
		out.FulfilledNote = &s
	}
	if r.RequestID.Valid {
		s := r.RequestID.String
		out.RequestID = &s
	}
	return out
}

// ListBackupRestoreRequests returns recent restore workflow rows (newest first).
func ListBackupRestoreRequests(c *gin.Context) {
	statusFilter := strings.TrimSpace(c.Query("status"))
	limit := 50
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 200 {
			respond.Error(c, http.StatusBadRequest, api.CodeValidation, "limit must be between 1 and 200", nil)
			return
		}
		limit = n
	}
	rows, err := AuthStore.ListBackupRestoreRequests(c.Request.Context(), limit, statusFilter)
	if err != nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to list backup restore requests", nil)
		return
	}
	items := make([]backupRestoreRequestJSON, 0, len(rows))
	for _, r := range rows {
		items = append(items, backupRestoreRowToJSON(r))
	}
	respond.OK(c, gin.H{"items": items, "limit": limit})
}

// CreateBackupRestoreRequestBody is JSON for POST /backups/restore-requests.
type CreateBackupRestoreRequestBody struct {
	Scope             string `json:"scope"`
	RestorePointLabel string `json:"restore_point_label"`
}

// PostBackupRestoreRequest creates a pending restore request (two approvers required; neither may be the requester).
func PostBackupRestoreRequest(c *gin.Context) {
	uid, ok := middleware.AuthUserIDFromContext(c)
	if !ok {
		respond.Error(c, http.StatusForbidden, api.CodeForbidden, "user subject required", nil)
		return
	}
	var body CreateBackupRestoreRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Error(c, http.StatusBadRequest, api.CodeValidation, "invalid JSON body", nil)
		return
	}
	scope := strings.TrimSpace(body.Scope)
	label := strings.TrimSpace(body.RestorePointLabel)
	if scope == "" || len(scope) > 64 {
		respond.Error(c, http.StatusBadRequest, api.CodeValidation, "scope is required and must be at most 64 characters", nil)
		return
	}
	if label == "" || len(label) > 256 {
		respond.Error(c, http.StatusBadRequest, api.CodeValidation, "restore_point_label is required and must be at most 256 characters", nil)
		return
	}
	reason := ""
	if v, ok := c.Get("platform_action_reason"); ok {
		if s, ok := v.(string); ok {
			reason = strings.TrimSpace(s)
		}
	}
	if reason == "" {
		respond.Error(c, http.StatusBadRequest, api.CodeValidation, "missing privileged action reason", map[string]any{
			"header": middleware.PlatformActionReasonHeader,
		})
		return
	}
	rid := requestid.FromContext(c)
	id, err := AuthStore.CreateBackupRestoreRequest(c.Request.Context(), uid, scope, label, reason, rid)
	if err != nil {
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to create backup restore request", nil)
		return
	}
	sub := c.GetString("auth_subject")
	_ = AuthStore.InsertAdminAuditEvent(c.Request.Context(), &uid, sub, "backup.restore.request", "backup_restore_request", strconv.FormatInt(id, 10),
		reason, rid, c.ClientIP(), c.GetHeader("User-Agent"), nil)
	slog.Info("backup_restore_request_created", "request_db_id", id, "requested_by", uid, "scope", scope, "request_id", rid)
	respond.OK(c, gin.H{"id": id, "status": authstore.BackupRestoreStatusPending})
}

func parseInt64Param(c *gin.Context, name string) (int64, bool) {
	s := strings.TrimSpace(c.Param(name))
	if s == "" {
		respond.Error(c, http.StatusBadRequest, api.CodeValidation, "missing path id", nil)
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 1 {
		respond.Error(c, http.StatusBadRequest, api.CodeValidation, "id must be a positive integer", nil)
		return 0, false
	}
	return n, true
}

// PostBackupRestoreApprove records one approval toward the two-approver gate.
func PostBackupRestoreApprove(c *gin.Context) {
	uid, ok := middleware.AuthUserIDFromContext(c)
	if !ok {
		respond.Error(c, http.StatusForbidden, api.CodeForbidden, "user subject required", nil)
		return
	}
	id, ok := parseInt64Param(c, "id")
	if !ok {
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
	err := AuthStore.ApproveBackupRestoreRequest(c.Request.Context(), id, uid)
	if err != nil {
		switch {
		case errors.Is(err, authstore.ErrBackupRestoreNotFound):
			respond.Error(c, http.StatusNotFound, api.CodeNotFound, "restore request not found", nil)
		case errors.Is(err, authstore.ErrBackupRestoreInvalidState):
			respond.Error(c, http.StatusConflict, api.CodeConflict, "restore request is not pending", nil)
		case errors.Is(err, authstore.ErrBackupRestoreCannotApprove):
			respond.Error(c, http.StatusForbidden, api.CodeForbidden, "requester cannot approve own restore request", nil)
		case errors.Is(err, authstore.ErrBackupRestoreDuplicateSlot):
			respond.Error(c, http.StatusConflict, api.CodeConflict, "same user cannot approve twice", nil)
		case errors.Is(err, authstore.ErrBackupRestoreAlreadyFilled):
			respond.Error(c, http.StatusConflict, api.CodeConflict, "restore request already has two approvals", nil)
		default:
			respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to approve restore request", nil)
		}
		return
	}
	row, _ := AuthStore.GetBackupRestoreRequest(c.Request.Context(), id)
	sub := c.GetString("auth_subject")
	rid := requestid.FromContext(c)
	meta, _ := json.Marshal(map[string]any{"status_after": row.Status})
	_ = AuthStore.InsertAdminAuditEvent(c.Request.Context(), &uid, sub, "backup.restore.approve", "backup_restore_request", strconv.FormatInt(id, 10),
		reason, rid, c.ClientIP(), c.GetHeader("User-Agent"), meta)
	slog.Info("backup_restore_request_approved", "request_db_id", id, "approver", uid, "status", row.Status, "request_id", rid)
	respond.OK(c, gin.H{"id": id, "status": row.Status})
}

// PostBackupRestoreReject rejects a pending request (caller must not be the requester).
func PostBackupRestoreReject(c *gin.Context) {
	uid, ok := middleware.AuthUserIDFromContext(c)
	if !ok {
		respond.Error(c, http.StatusForbidden, api.CodeForbidden, "user subject required", nil)
		return
	}
	id, ok := parseInt64Param(c, "id")
	if !ok {
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
	err := AuthStore.RejectBackupRestoreRequest(c.Request.Context(), id, uid, reason)
	if err != nil {
		switch {
		case errors.Is(err, authstore.ErrBackupRestoreNotFound):
			respond.Error(c, http.StatusNotFound, api.CodeNotFound, "restore request not found", nil)
		case errors.Is(err, authstore.ErrBackupRestoreCannotApprove):
			respond.Error(c, http.StatusForbidden, api.CodeForbidden, "requester cannot reject as approver; use cancel instead", nil)
		case errors.Is(err, authstore.ErrBackupRestoreInvalidState):
			respond.Error(c, http.StatusConflict, api.CodeConflict, "restore request is not pending", nil)
		default:
			respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to reject restore request", nil)
		}
		return
	}
	sub := c.GetString("auth_subject")
	rid := requestid.FromContext(c)
	_ = AuthStore.InsertAdminAuditEvent(c.Request.Context(), &uid, sub, "backup.restore.reject", "backup_restore_request", strconv.FormatInt(id, 10),
		reason, rid, c.ClientIP(), c.GetHeader("User-Agent"), nil)
	slog.Info("backup_restore_request_rejected", "request_db_id", id, "actor", uid, "request_id", rid)
	respond.OK(c, gin.H{"id": id, "status": authstore.BackupRestoreStatusRejected})
}

// PostBackupRestoreFulfill marks an approved request fulfilled after operators complete restore out of band.
func PostBackupRestoreFulfill(c *gin.Context) {
	uid, ok := middleware.AuthUserIDFromContext(c)
	if !ok {
		respond.Error(c, http.StatusForbidden, api.CodeForbidden, "user subject required", nil)
		return
	}
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	note := strings.TrimSpace(c.GetHeader(middleware.PlatformActionReasonHeader))
	if note == "" {
		if v, ok := c.Get("platform_action_reason"); ok {
			if s, ok := v.(string); ok {
				note = strings.TrimSpace(s)
			}
		}
	}
	err := AuthStore.FulfillBackupRestoreRequest(c.Request.Context(), id, uid, note)
	if err != nil {
		if errors.Is(err, authstore.ErrBackupRestoreInvalidState) {
			respond.Error(c, http.StatusConflict, api.CodeConflict, "restore request is not approved", nil)
			return
		}
		respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to fulfill restore request", nil)
		return
	}
	sub := c.GetString("auth_subject")
	rid := requestid.FromContext(c)
	_ = AuthStore.InsertAdminAuditEvent(c.Request.Context(), &uid, sub, "backup.restore.fulfill", "backup_restore_request", strconv.FormatInt(id, 10),
		note, rid, c.ClientIP(), c.GetHeader("User-Agent"), nil)
	slog.Info("backup_restore_request_fulfilled", "request_db_id", id, "actor", uid, "request_id", rid)
	respond.OK(c, gin.H{"id": id, "status": authstore.BackupRestoreStatusFulfilled})
}

// PostBackupRestoreCancel lets the requester cancel a pending request.
func PostBackupRestoreCancel(c *gin.Context) {
	uid, ok := middleware.AuthUserIDFromContext(c)
	if !ok {
		respond.Error(c, http.StatusForbidden, api.CodeForbidden, "user subject required", nil)
		return
	}
	id, ok := parseInt64Param(c, "id")
	if !ok {
		return
	}
	reason := strings.TrimSpace(c.GetHeader(middleware.PlatformActionReasonHeader))
	if reason == "" {
		if v, ok := c.Get("platform_action_reason"); ok {
			if s, vok := v.(string); vok {
				reason = strings.TrimSpace(s)
			}
		}
	}
	err := AuthStore.CancelBackupRestoreRequest(c.Request.Context(), id, uid)
	if err != nil {
		switch {
		case errors.Is(err, authstore.ErrBackupRestoreNotFound):
			respond.Error(c, http.StatusNotFound, api.CodeNotFound, "restore request not found", nil)
		case errors.Is(err, authstore.ErrBackupRestoreNotRequester):
			respond.Error(c, http.StatusForbidden, api.CodeForbidden, "only the requester may cancel", nil)
		case errors.Is(err, authstore.ErrBackupRestoreInvalidState):
			respond.Error(c, http.StatusConflict, api.CodeConflict, "restore request is not pending", nil)
		default:
			respond.Error(c, http.StatusInternalServerError, api.CodeInternal, "failed to cancel restore request", nil)
		}
		return
	}
	sub := c.GetString("auth_subject")
	rid := requestid.FromContext(c)
	_ = AuthStore.InsertAdminAuditEvent(c.Request.Context(), &uid, sub, "backup.restore.cancel", "backup_restore_request", strconv.FormatInt(id, 10),
		reason, rid, c.ClientIP(), c.GetHeader("User-Agent"), nil)
	slog.Info("backup_restore_request_cancelled", "request_db_id", id, "requester", uid, "request_id", rid)
	respond.OK(c, gin.H{"id": id, "status": authstore.BackupRestoreStatusCancelled})
}
