CREATE TABLE IF NOT EXISTS grant_applications (
    id SERIAL PRIMARY KEY,
    parish_id INTEGER NOT NULL REFERENCES parishes(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_title VARCHAR(500) NOT NULL,
    project_description TEXT,
    building_stage VARCHAR(50),
    requested_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    estimated_project_cost DECIMAL(15,2) NOT NULL DEFAULT 0,
    project_location TEXT,
    project_start_date DATE,
    expected_completion_date DATE,
    justification TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_grant_applications_parish_id ON grant_applications(parish_id);
CREATE INDEX IF NOT EXISTS idx_grant_applications_user_id ON grant_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_grant_applications_status ON grant_applications(status);
CREATE INDEX IF NOT EXISTS idx_grant_applications_deleted_at ON grant_applications(deleted_at);
