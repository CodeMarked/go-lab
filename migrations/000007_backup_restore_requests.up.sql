-- Backup restore request workflow with two distinct approvers (neither the requester).

CREATE TABLE IF NOT EXISTS backup_restore_requests (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  requested_by_user_id INT NOT NULL,
  scope VARCHAR(64) NOT NULL,
  restore_point_label VARCHAR(256) NOT NULL,
  reason TEXT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  rejection_reason TEXT DEFAULT NULL,
  approval_1_user_id INT DEFAULT NULL,
  approval_1_at TIMESTAMP NULL DEFAULT NULL,
  approval_2_user_id INT DEFAULT NULL,
  approval_2_at TIMESTAMP NULL DEFAULT NULL,
  fulfilled_at TIMESTAMP NULL DEFAULT NULL,
  fulfilled_note TEXT DEFAULT NULL,
  request_id VARCHAR(128) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_backup_restore_status (status),
  KEY idx_backup_restore_requested_by (requested_by_user_id),
  KEY idx_backup_restore_created (created_at),
  CONSTRAINT fk_backup_restore_requested_by FOREIGN KEY (requested_by_user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_backup_restore_approval_1 FOREIGN KEY (approval_1_user_id) REFERENCES users (id) ON DELETE SET NULL,
  CONSTRAINT fk_backup_restore_approval_2 FOREIGN KEY (approval_2_user_id) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
