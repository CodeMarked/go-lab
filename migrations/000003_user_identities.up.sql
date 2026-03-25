-- Links external OIDC subjects (issuer + sub) to local users.id for Auth0 / other IdPs.
CREATE TABLE IF NOT EXISTS user_identities (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  issuer VARCHAR(512) NOT NULL,
  subject VARCHAR(512) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_user_identities_issuer_subject (issuer(255), subject(255)),
  CONSTRAINT fk_user_identities_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  KEY idx_user_identities_user (user_id)
) ENGINE=InnoDB;
