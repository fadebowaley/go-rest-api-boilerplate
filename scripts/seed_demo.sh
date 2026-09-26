#!/bin/bash
# seed_demo.sh - Creates demo data for Applico
# Compatible with bash 3.2+ (no associative arrays)
set -euo pipefail

BASE_URL="http://localhost:8080/api/v1"
DB_CMD="docker exec -i go_api_db psql -U postgres -d grab"
TENANT_ID=1

PASS()  { echo "  [PASS] $1"; }
FAIL()  { echo "  [FAIL] $1"; exit 1; }
INFO()  { echo "  [INFO] $1"; }
STEP()  { echo ""; echo "=== $1 ==="; }
API_JQ() { local t=$1; shift; curl -s -H "Authorization: Bearer $t" -H "X-Tenant-ID: $TENANT_ID" "$@"; }

extract_token() {
  echo "$1" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null
}

user_id_from_register() {
  echo "$1" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('user',{}).get('id',''))" 2>/dev/null
}

login_user() {
  local email="$1" pass="$2"
  curl -s -X POST "$BASE_URL/auth/login" -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$pass\"}"
}

register_user() {
  local email="$1" pass="$2" name="$3"
  curl -s -X POST "$BASE_URL/auth/register" -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$pass\",\"name\":\"$name\",\"organization\":\"RCCG Homeland Mission\"}"
}

# ---------------------------------------------------------------------------
STEP "Checking environment"
curl -sf "http://localhost:8080/health" > /dev/null 2>&1 || FAIL "API not running"
PASS "API is healthy"
echo "SELECT 1" | $DB_CMD > /dev/null 2>&1 || FAIL "Cannot reach DB"
PASS "Database accessible"

# ---------------------------------------------------------------------------
STEP "1. Ensure custom roles exist"
$DB_CMD <<'SQL'
INSERT INTO roles (name, description)
VALUES
  ('tenant_admin', 'Tenant administrator'),
  ('program_manager', 'Program manager'),
  ('reviewer', 'Reviewer'),
  ('applicant', 'Applicant')
ON CONFLICT (name) DO NOTHING;
SQL
PASS "Roles ensured"

# ---------------------------------------------------------------------------
STEP "2. Register tenant users"

# We iterate over user definitions using simple indexed arrays
USER_KEYS=(tenantadmin programmgr reviewer finance auditor applicant)
USER_EMAILS=(tenantadmin@rccg.org programmgr@rccg.org reviewer@rccg.org finance@rccg.org auditor@rccg.org applicant@rccg.org)
USER_PASSES=(Admin123! Mgr123! Review123! Finance123! Audit123! Apply123!)
USER_NAMES=("Tenant Admin" "Program Manager" "Reviewer" "Finance Officer" "Auditor" "Applicant User")

for i in "${!USER_KEYS[@]}"; do
  key="${USER_KEYS[$i]}"
  email="${USER_EMAILS[$i]}"
  pass="${USER_PASSES[$i]}"
  name="${USER_NAMES[$i]}"
  INFO "Registering $email"
  resp=$(register_user "$email" "$pass" "$name")
  tok=$(extract_token "$resp")
  if [[ -z "$tok" ]]; then
    INFO "  -> may already exist, logging in"
    resp=$(login_user "$email" "$pass")
    tok=$(extract_token "$resp")
    if [[ -z "$tok" ]]; then
      echo "  $resp" | python3 -m json.tool 2>/dev/null || true
      FAIL "Cannot register $email"
    fi
  fi
  # Store in global scope by writing to temp file (bash 3 compatible)
  eval "TOKEN_${key}=\"$tok\""
  uid=$(user_id_from_register "$resp")
  eval "UID_${key}=$uid"
  PASS "$email (user_id=$uid)"
done

# Extract tokens and UIDs
TOKEN_ADMIN="${TOKEN_tenantadmin}"
TOKEN_APPLICANT="${TOKEN_applicant}"
TOKEN_FINANCE="${TOKEN_finance}"
TOKEN_REVIEWER="${TOKEN_reviewer}"
TOKEN_PGMGR="${TOKEN_programmgr}"

UID_ADMIN="${UID_tenantadmin}"
UID_PGMGR="${UID_programmgr}"
UID_REVIEWER="${UID_reviewer}"
UID_FINANCE="${UID_finance}"
UID_AUDITOR="${UID_auditor}"
UID_APPLICANT="${UID_applicant}"

# ---------------------------------------------------------------------------
STEP "3. Assign roles to users (SQL bypasses RBAC bootstrap)"

# Use direct SQL since API RBAC prevents bootstrapping
assign_role_sql() {
  local uid="$1" role="$2"
  [[ -z "$uid" || "$uid" == "null" ]] && { return; }
  $DB_CMD <<SQL
INSERT INTO user_roles (user_id, role_id)
SELECT $uid, id FROM roles WHERE name = '$role'
ON CONFLICT DO NOTHING;
SQL
}

assign_role_sql "$UID_ADMIN" "homeland_admin"
assign_role_sql "$UID_ADMIN" "tenant_admin"
assign_role_sql "$UID_PGMGR" "program_manager"
assign_role_sql "$UID_REVIEWER" "reviewer"
assign_role_sql "$UID_FINANCE" "finance_officer"
assign_role_sql "$UID_AUDITOR" "auditor"
assign_role_sql "$UID_APPLICANT" "applicant"
assign_role_sql "$UID_ADMIN" "grant_officer"
assign_role_sql "$UID_PGMGR" "grant_officer"
assign_role_sql "$UID_REVIEWER" "grant_officer"
PASS "Roles assigned"

# ---------------------------------------------------------------------------
STEP "4. Add users to tenant (SQL bypasses API)"

add_to_tenant_sql() {
  local uid="$1"
  [[ -z "$uid" || "$uid" == "null" ]] && { return; }
  $DB_CMD <<SQL
INSERT INTO tenant_users (tenant_id, user_id)
SELECT $TENANT_ID, $uid
ON CONFLICT DO NOTHING;
SQL
}

add_to_tenant_sql "$UID_ADMIN"
add_to_tenant_sql "$UID_PGMGR"
add_to_tenant_sql "$UID_REVIEWER"
add_to_tenant_sql "$UID_FINANCE"
add_to_tenant_sql "$UID_AUDITOR"
add_to_tenant_sql "$UID_APPLICANT"
PASS "Users added to tenant"

# Re-login users after role assignment so JWT tokens contain their roles
INFO "Re-logging users after role assignments"
TOKEN_ADMIN=$(extract_token "$(login_user "tenantadmin@rccg.org" "Admin123!")")
TOKEN_PGMGR=$(extract_token "$(login_user "programmgr@rccg.org" "Mgr123!")")
TOKEN_REVIEWER=$(extract_token "$(login_user "reviewer@rccg.org" "Review123!")")
TOKEN_FINANCE=$(extract_token "$(login_user "finance@rccg.org" "Finance123!")")
TOKEN_AUDITOR=$(extract_token "$(login_user "auditor@rccg.org" "Audit123!")")
TOKEN_APPLICANT=$(extract_token "$(login_user "applicant@rccg.org" "Apply123!")")
PASS "All users re-logged in with fresh tokens"

# ---------------------------------------------------------------------------
STEP "5. Create grant program"

PROGRAM_RESP=$(API_JQ "$TOKEN_ADMIN" -X POST "$BASE_URL/programs" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Church Building Support Grant",
    "description": "Support for church building and renovation projects",
    "eligibility_rules": {"max_amount": 50000},
    "currency": "USD"
  }')
PROGRAM_ID=$(echo "$PROGRAM_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [[ -z "$PROGRAM_ID" ]]; then
  PROGRAM_ID=$(API_JQ "$TOKEN_ADMIN" "$BASE_URL/programs" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; print(d[0]['id'] if d else '')" 2>/dev/null)
  INFO "Using existing program id=$PROGRAM_ID"
else
  PASS "Program created (id=$PROGRAM_ID)"
fi

# ---------------------------------------------------------------------------
STEP "6. Create form template and publish"

FORM_RESP=$(API_JQ "$TOKEN_ADMIN" -X POST "$BASE_URL/forms/templates" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Church Building Grant Application",
    "description": "Standard application form",
    "schema_definition": {
      "fields": [
        {"name": "project_title", "type": "text", "label": "Project Title", "required": true},
        {"name": "amount", "type": "number", "label": "Requested Amount", "required": true},
        {"name": "location", "type": "text", "label": "Project Location", "required": true},
        {"name": "description", "type": "textarea", "label": "Project Description", "required": true}
      ]
    }
  }')
FORM_ID=$(echo "$FORM_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [[ -z "$FORM_ID" ]]; then
  FORM_ID=$(API_JQ "$TOKEN_ADMIN" "$BASE_URL/forms/templates" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; print(d[0]['id'] if d else '')" 2>/dev/null)
  INFO "Using existing form id=$FORM_ID"
else
  PASS "Form template created (id=$FORM_ID)"
fi

API_JQ "$TOKEN_ADMIN" -X POST "$BASE_URL/forms/templates/$FORM_ID/publish" > /dev/null 2>&1 && \
  PASS "Form published" || INFO "Form already published"

# ---------------------------------------------------------------------------
STEP "7. Create workflow template with steps"

WF_RESP=$(API_JQ "$TOKEN_ADMIN" -X POST "$BASE_URL/workflows" \
  -H "Content-Type: application/json" \
  -d '{"name": "Standard Grant Review", "description": "3-step review process"}')
WF_ID=$(echo "$WF_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [[ -z "$WF_ID" ]]; then
  WF_ID=$(API_JQ "$TOKEN_ADMIN" "$BASE_URL/workflows" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; print(d[0]['id'] if d else '')" 2>/dev/null)
  INFO "Using existing workflow id=$WF_ID"
else
  PASS "Workflow created (id=$WF_ID)"
fi

STEPS_EXIST=$(API_JQ "$TOKEN_ADMIN" "$BASE_URL/workflows/$WF_ID/steps" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('data',[])))" 2>/dev/null)
if [[ "$STEPS_EXIST" -eq 0 ]]; then
  S1=$(API_JQ "$TOKEN_ADMIN" -X POST "$BASE_URL/workflows/$WF_ID/steps" \
    -H "Content-Type: application/json" \
    -d '{"name": "Initial Review", "step_order": 1, "assignee_roles": ["reviewer"]}' | \
    python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
  PASS "Step 1 created (id=$S1)"

  S2=$(API_JQ "$TOKEN_ADMIN" -X POST "$BASE_URL/workflows/$WF_ID/steps" \
    -H "Content-Type: application/json" \
    -d '{"name": "Committee Review", "step_order": 2, "assignee_roles": ["program_manager","reviewer"]}' | \
    python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
  PASS "Step 2 created (id=$S2)"

  S3=$(API_JQ "$TOKEN_ADMIN" -X POST "$BASE_URL/workflows/$WF_ID/steps" \
    -H "Content-Type: application/json" \
    -d '{"name": "Final Approval", "step_order": 3, "assignee_roles": ["tenant_admin"]}' | \
    python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])" 2>/dev/null)
  PASS "Step 3 created (id=$S3)"

  echo "UPDATE workflow_steps SET is_final_approval = true WHERE id = $S3;" | $DB_CMD
  PASS "Step 3 marked final approval"
else
  STEP_IDS=$(API_JQ "$TOKEN_ADMIN" "$BASE_URL/workflows/$WF_ID/steps" | python3 -c "
import sys,json
steps=sorted(json.load(sys.stdin)['data'], key=lambda s: s.get('step_order',0))
for s in steps: print(s['id'])
" 2>/dev/null)
  S1=$(echo "$STEP_IDS" | sed -n '1p')
  S2=$(echo "$STEP_IDS" | sed -n '2p')
  S3=$(echo "$STEP_IDS" | sed -n '3p')
  INFO "Using existing steps: $S1 $S2 $S3"
fi

# Link workflow to program
API_JQ "$TOKEN_ADMIN" -X PUT "$BASE_URL/programs/$PROGRAM_ID" \
  -H "Content-Type: application/json" \
  -d "{\"workflow_template_id\": $WF_ID}" > /dev/null 2>&1 && \
  PASS "Workflow linked to program" || INFO "Already linked"

# ---------------------------------------------------------------------------
STEP "8. Register applicant entity"

APPLICANT_RESP=$(API_JQ "$TOKEN_ADMIN" -X POST "$BASE_URL/applicants" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "parish",
    "name": "RCCG Central Parish",
    "email": "parish@rccg.org",
    "phone": "+234-800-RCCG",
    "address": "1 Cathedral Avenue, Lagos"
  }')
APPLICANT_ID=$(echo "$APPLICANT_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [[ -z "$APPLICANT_ID" ]]; then
  APPLICANT_ID=$(API_JQ "$TOKEN_ADMIN" "$BASE_URL/applicants" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; print(d[0]['id'] if d else '')" 2>/dev/null)
  INFO "Using existing applicant id=$APPLICANT_ID"
else
  PASS "Applicant created (id=$APPLICANT_ID)"
fi

# ---------------------------------------------------------------------------
STEP "9. Create grant application"

GRANT_RESP=$(API_JQ "$TOKEN_APPLICANT" -X POST "$BASE_URL/grants" \
  -H "Content-Type: application/json" \
  -d "{
    \"program_id\": $PROGRAM_ID,
    \"applicant_id\": $APPLICANT_ID,
    \"project_title\": \"Cathedral Renovation Phase 2\",
    \"requested_amount\": 25000,
    \"estimated_project_cost\": 35000,
    \"project_location\": \"Lagos, Nigeria\",
    \"project_description\": \"Renovation of the main cathedral building\"
  }")
GRANT_ID=$(echo "$GRANT_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [[ -z "$GRANT_ID" ]]; then
  echo "  $GRANT_RESP" | python3 -m json.tool 2>/dev/null || true
  FAIL "Could not create grant"
fi
PASS "Grant created (id=$GRANT_ID, draft)"

# ---------------------------------------------------------------------------
STEP "10. Submit the grant application"

SUBMIT_RESP=$(API_JQ "$TOKEN_APPLICANT" -X POST "$BASE_URL/grants/$GRANT_ID/submit" -H "Content-Type: application/json")
STATUS=$(echo "$SUBMIT_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('status',''))" 2>/dev/null)
if [[ "$STATUS" == "submitted" ]]; then
  PASS "Grant submitted"
else
  echo "  $SUBMIT_RESP" | python3 -m json.tool 2>/dev/null || true
  FAIL "Could not submit grant"
fi

# ---------------------------------------------------------------------------
STEP "11. Workflow lifecycle via SQL"

WF_ROW=$(echo "SELECT id, current_step_id, status FROM application_workflows WHERE application_id = $GRANT_ID LIMIT 1;" | $DB_CMD -t -A -F'|')
WF_APP_ID=$(echo "$WF_ROW" | cut -d'|' -f1)
WF_CUR_STEP=$(echo "$WF_ROW" | cut -d'|' -f2)
[[ -z "$WF_APP_ID" ]] && FAIL "No workflow started"
PASS "Workflow started (id=$WF_APP_ID, step=$WF_CUR_STEP)"

# Step IDs ordered
STEP_IDS=$(API_JQ "$TOKEN_ADMIN" "$BASE_URL/workflows/$WF_ID/steps" | python3 -c "
import sys,json
steps=sorted(json.load(sys.stdin)['data'], key=lambda s: s.get('step_order',0))
for s in steps: print(s['id'])
" 2>/dev/null)
S1=$(echo "$STEP_IDS" | sed -n '1p')
S2=$(echo "$STEP_IDS" | sed -n '2p')
S3=$(echo "$STEP_IDS" | sed -n '3p')

approve_step_sql() {
  local step_id="$1" actor_uid="$2" comment="$3"
  $DB_CMD <<SQL
INSERT INTO workflow_actions (tenant_id, application_workflow_id, step_id, actor_id, action, comment, created_at)
SELECT $TENANT_ID, $WF_APP_ID, $step_id, $actor_uid, 'approved', '$comment', NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM workflow_actions
  WHERE application_workflow_id = $WF_APP_ID AND step_id = $step_id AND actor_id = $actor_uid
);
SQL
  local count=$($DB_CMD -t -A -c "SELECT COUNT(*) FROM workflow_actions WHERE application_workflow_id = $WF_APP_ID AND step_id = $step_id AND action = 'approved';")
  if [[ "$count" -gt 0 ]]; then PASS "Step $step_id approved"; else FAIL "Step $step_id not approved"; fi
}

approve_step_sql "$S1" "$UID_REVIEWER" "Initial review complete"
$DB_CMD <<SQL
UPDATE application_workflows SET current_step_id = $S2, updated_at = NOW() WHERE id = $WF_APP_ID AND current_step_id = $S1;
SQL
PASS "Advanced to step 2"

approve_step_sql "$S2" "$UID_PGMGR" "Committee review passed"
$DB_CMD <<SQL
UPDATE application_workflows SET current_step_id = $S3, updated_at = NOW() WHERE id = $WF_APP_ID AND current_step_id = $S2;
SQL
PASS "Advanced to step 3"

approve_step_sql "$S3" "$UID_ADMIN" "Final approval granted"
$DB_CMD <<SQL
UPDATE application_workflows SET
  status = 'completed', current_step_id = NULL, completed_at = NOW(), updated_at = NOW()
WHERE id = $WF_APP_ID;
UPDATE grant_applications SET status = 'approved', updated_at = NOW() WHERE id = $GRANT_ID;
SQL
PASS "Workflow completed, grant approved"

# ---------------------------------------------------------------------------
STEP "12. Create disbursement"

DISP_RESP=$(API_JQ "$TOKEN_FINANCE" -X POST "$BASE_URL/grants/$GRANT_ID/disbursements" \
  -H "Content-Type: application/json" \
  -d "{\"grant_application_id\": $GRANT_ID, \"amount\": 25000, \"currency\": \"USD\", \"notes\": \"First tranche payment\"}")
DISP_ID=$(echo "$DISP_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [[ -n "$DISP_ID" ]]; then
  PASS "Disbursement created (id=$DISP_ID)"
  echo "UPDATE grant_applications SET status = 'disbursed', updated_at = NOW() WHERE id = $GRANT_ID;" | $DB_CMD
  PASS "Grant marked disbursed"
else
  echo "  $DISP_RESP" | python3 -m json.tool 2>/dev/null || true
  FAIL "Could not create disbursement"
fi

# ---------------------------------------------------------------------------
STEP "13. Submit project update"

UPDATE_RESP=$(API_JQ "$TOKEN_APPLICANT" -X POST "$BASE_URL/grants/$GRANT_ID/updates" \
  -H "Content-Type: application/json" \
  -d "{\"grant_application_id\": $GRANT_ID, \"title\": \"Foundation work completed\", \"description\": \"Foundation renovation complete\", \"notes\": \"Next: roofing\"}")
UPDATE_ID=$(echo "$UPDATE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [[ -n "$UPDATE_ID" ]]; then
  PASS "Project update created (id=$UPDATE_ID)"
else
  echo "  $UPDATE_RESP" | python3 -m json.tool 2>/dev/null || true
  FAIL "Could not create project update"
fi

# ---------------------------------------------------------------------------
STEP "14. Verify final state"

echo ""
echo "========== DEMO DATA SUMMARY =========="
echo "Tenant:         RCCG Homeland Mission (id=$TENANT_ID)"
echo "Program:        Church Building Support Grant (id=$PROGRAM_ID)"
echo "Form Template:  Church Building Grant Application (id=$FORM_ID)"
echo "Workflow:       Standard Grant Review (id=$WF_ID)"
echo "Applicant:      RCCG Central Parish (id=$APPLICANT_ID)"
echo "Grant:          Cathedral Renovation Phase 2 (id=$GRANT_ID)"
echo "Disbursement:   id=$DISP_ID"
echo "Project Update: id=$UPDATE_ID"
echo ""
echo "Users:"
for i in "${!USER_KEYS[@]}"; do
  email="${USER_EMAILS[$i]}"
  pass="${USER_PASSES[$i]}"
  name="${USER_NAMES[$i]}"
  echo "  $name <$email> / $pass"
done
echo ""
echo "========================================"
