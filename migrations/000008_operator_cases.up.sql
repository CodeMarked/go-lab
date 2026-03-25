-- Operator cases for player/character workflows (platform governance; Marble remains gameplay authority).

CREATE TABLE IF NOT EXISTS operator_cases (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  status VARCHAR(32) NOT NULL DEFAULT 'open',
  priority VARCHAR(16) NOT NULL DEFAULT 'normal',
  subject_platform_user_id INT NOT NULL,
  subject_character_ref VARCHAR(128) DEFAULT NULL,
  title VARCHAR(256) NOT NULL,
  description TEXT,
  created_by_user_id INT NOT NULL,
  assigned_to_user_id INT DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_operator_cases_status (status),
  KEY idx_operator_cases_subject_user (subject_platform_user_id),
  KEY idx_operator_cases_created (created_at),
  CONSTRAINT fk_operator_cases_subject_user FOREIGN KEY (subject_platform_user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_operator_cases_created_by FOREIGN KEY (created_by_user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_operator_cases_assigned_to FOREIGN KEY (assigned_to_user_id) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS operator_case_notes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  case_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  body TEXT NOT NULL,
  created_by_user_id INT NOT NULL,
  PRIMARY KEY (id),
  KEY idx_operator_case_notes_case (case_id),
  CONSTRAINT fk_operator_case_notes_case FOREIGN KEY (case_id) REFERENCES operator_cases (id) ON DELETE CASCADE,
  CONSTRAINT fk_operator_case_notes_author FOREIGN KEY (created_by_user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Append-only privileged actions: sanction, recovery_request, appeal_resolve (JSON payload + reason copy).
CREATE TABLE IF NOT EXISTS operator_case_actions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  case_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  action_kind VARCHAR(32) NOT NULL,
  payload_json JSON DEFAULT NULL,
  reason TEXT,
  actor_user_id INT NOT NULL,
  PRIMARY KEY (id),
  KEY idx_operator_case_actions_case (case_id),
  KEY idx_operator_case_actions_kind (action_kind),
  CONSTRAINT fk_operator_case_actions_case FOREIGN KEY (case_id) REFERENCES operator_cases (id) ON DELETE CASCADE,
  CONSTRAINT fk_operator_case_actions_actor FOREIGN KEY (actor_user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO platform_roles (name) VALUES ('gm_liveops');
