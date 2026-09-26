BEGIN;

ALTER TABLE grant_programs DROP COLUMN IF EXISTS form_template_id;
ALTER TABLE grant_applications DROP COLUMN IF EXISTS form_version_id;
ALTER TABLE grant_applications DROP COLUMN IF EXISTS form_template_id;
DROP TABLE IF EXISTS grant_application_versions;
DROP TABLE IF EXISTS form_versions;
DROP TABLE IF EXISTS form_templates;

COMMIT;
