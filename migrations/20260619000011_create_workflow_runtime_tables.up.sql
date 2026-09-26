BEGIN;

CREATE TABLE IF NOT EXISTS application_workflows (
    id SERIAL PRIMARY KEY,
    application_id INTEGER NOT NULL REFERENCES grant_applications(id) ON DELETE CASCADE,
    workflow_template_id INTEGER NOT NULL REFERENCES workflow_templates(id) ON DELETE CASCADE,
    current_step_id INTEGER REFERENCES workflow_steps(id),
    status VARCHAR(30) NOT NULL DEFAULT 'in_progress',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflow_actions (
    id SERIAL PRIMARY KEY,
    application_workflow_id INTEGER NOT NULL REFERENCES application_workflows(id) ON DELETE CASCADE,
    step_id INTEGER NOT NULL REFERENCES workflow_steps(id),
    actor_id INTEGER NOT NULL REFERENCES users(id),
    action VARCHAR(30) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_app_workflows_application ON application_workflows(application_id);
CREATE INDEX IF NOT EXISTS idx_app_workflows_current_step ON application_workflows(current_step_id);
CREATE INDEX IF NOT EXISTS idx_app_workflows_status ON application_workflows(status);
CREATE INDEX IF NOT EXISTS idx_workflow_actions_workflow ON workflow_actions(application_workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_actions_step ON workflow_actions(step_id);
CREATE INDEX IF NOT EXISTS idx_workflow_actions_actor ON workflow_actions(actor_id);
CREATE INDEX IF NOT EXISTS idx_workflow_actions_action ON workflow_actions(action);

COMMIT;
