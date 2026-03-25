DROP TABLE IF EXISTS operator_case_actions;
DROP TABLE IF EXISTS operator_case_notes;
DROP TABLE IF EXISTS operator_cases;
DELETE ur FROM user_platform_roles ur
  INNER JOIN platform_roles r ON ur.role_id = r.id AND r.name = 'gm_liveops';
DELETE FROM platform_roles WHERE name = 'gm_liveops';
