DELETE FROM roles WHERE name IN (
    'super_admin',
    'homeland_admin',
    'grant_officer',
    'finance_officer',
    'auditor',
    'committee_member',
    'regional_reviewer',
    'provincial_reviewer',
    'zonal_reviewer',
    'area_reviewer',
    'parish_pastor'
);
