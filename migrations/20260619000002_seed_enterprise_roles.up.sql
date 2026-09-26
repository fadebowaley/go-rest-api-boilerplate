INSERT INTO roles (name, description)
VALUES
    ('super_admin', 'Super Administrator with unrestricted system access'),
    ('homeland_admin', 'Homeland Mission administrator with full platform access'),
    ('grant_officer', 'Grant Officer who processes and reviews grant applications'),
    ('finance_officer', 'Finance Officer who manages disbursements and payments'),
    ('auditor', 'Auditor who reviews applications, approvals, and disbursements'),
    ('committee_member', 'Committee Member who participates in approval decisions'),
    ('regional_reviewer', 'Regional Reviewer who reviews applications at regional level'),
    ('provincial_reviewer', 'Provincial Reviewer who reviews applications at provincial level'),
    ('zonal_reviewer', 'Zonal Reviewer who reviews applications at zonal level'),
    ('area_reviewer', 'Area Reviewer who reviews applications at area level'),
    ('parish_pastor', 'Parish Pastor who submits grant applications on behalf of a parish')
ON CONFLICT (name) DO NOTHING;
