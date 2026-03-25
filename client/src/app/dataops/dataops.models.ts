/** GET /backups/status */
export interface BackupsStatusData {
  restore_workflow_enabled: boolean;
  pending_restore_requests: number;
  approved_awaiting_fulfillment: number;
  counts_by_status: Record<string, number>;
  note: string;
  permissions?: string[];
}

/** GET /backups/restore-requests */
export interface BackupRestoreRequestRow {
  id: number;
  created_at: string;
  updated_at: string;
  requested_by_user_id: number;
  scope: string;
  restore_point_label: string;
  reason: string;
  status: string;
  rejection_reason?: string;
  approval_1_user_id?: number;
  approval_1_at?: string;
  approval_2_user_id?: number;
  approval_2_at?: string;
  fulfilled_at?: string;
  fulfilled_note?: string;
  request_id?: string;
}

export interface BackupRestoreListData {
  items: BackupRestoreRequestRow[];
  limit: number;
}

export interface BackupRestoreCreateData {
  id: number;
  status: string;
}

export interface BackupRestoreIdStatus {
  id: number;
  status: string;
}

export interface SecurityMeData {
  user_id: number;
  roles: string[];
  effective_permissions: string[];
}
