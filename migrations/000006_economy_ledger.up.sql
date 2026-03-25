-- Phase B: append-only economy ledger read model for operator visibility (ingestion from Marble/suite is out of band).

CREATE TABLE IF NOT EXISTS economy_ledger_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  platform_user_id INT NOT NULL,
  event_type VARCHAR(64) NOT NULL,
  amount_delta BIGINT NOT NULL DEFAULT 0,
  currency_code VARCHAR(32) NOT NULL DEFAULT 'default',
  reference_type VARCHAR(64) DEFAULT NULL,
  reference_id VARCHAR(128) DEFAULT NULL,
  meta_json JSON DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_economy_ledger_created (created_at),
  KEY idx_economy_ledger_user_created (platform_user_id, created_at),
  KEY idx_economy_ledger_event_type (event_type),
  CONSTRAINT fk_economy_ledger_user FOREIGN KEY (platform_user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
