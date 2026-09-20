#!/bin/sh
# Seed / reset the fixed accounts used by the Playwright E2E suite.
#
#   scripts/seed-test-data.sh          # (re)create test accounts
#   scripts/seed-test-data.sh --clean  # remove test accounts and their data
#
# Accounts (password from E2E_PASSWORD, default "E2eTest1!"):
#   e2e-admin@hr-sistem.com     -> PLATFORM_ADMIN  (+ employee profile)
#   e2e-hr@hr-sistem.com        -> HR_ADMIN        (+ employee profile, is the supervisor)
#   e2e-manager@hr-sistem.com   -> MANAGER_PORTAL_ACCESS (+ employee profile)
#   e2e-employee@hr-sistem.com  -> EMPLOYEE        (+ employee profile + 15 vacation days)
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"  # project root (informational)
API_BASE="${E2E_API_BASE:-http://localhost:8034/api/v1}"
PASSWORD="${E2E_PASSWORD:-E2eTest1!}"
PG_CONTAINER="${PG_CONTAINER:-hr_db}"
PG_USER="${POSTGRES_USER:-hr_user}"
PG_DB="${POSTGRES_DB:-hr_db}"

ADMIN_EMAIL="e2e-admin@hr-sistem.com"
HR_EMAIL="e2e-hr@hr-sistem.com"
MGR_EMAIL="e2e-manager@hr-sistem.com"
EMP_EMAIL="e2e-employee@hr-sistem.com"

psql_cmd() { podman exec -i "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -t -A -c "$1"; }

cleanup() {
  psql_cmd "
    DELETE FROM absence_requests WHERE employee_id IN (SELECT e.id FROM employees e JOIN users u ON u.id=e.user_id WHERE u.email LIKE 'e2e-%@hr-sistem.com')
       OR created_by IN (SELECT e.id FROM employees e JOIN users u ON u.id=e.user_id WHERE u.email LIKE 'e2e-%@hr-sistem.com');
    DELETE FROM leave_balance WHERE employee_id IN (SELECT e.id FROM employees e JOIN users u ON u.id=e.user_id WHERE u.email LIKE 'e2e-%@hr-sistem.com');
    DELETE FROM employee_leave_policy WHERE employee_id IN (SELECT e.id FROM employees e JOIN users u ON u.id=e.user_id WHERE u.email LIKE 'e2e-%@hr-sistem.com');
    DELETE FROM attendance WHERE employee_id IN (SELECT e.id FROM employees e JOIN users u ON u.id=e.user_id WHERE u.email LIKE 'e2e-%@hr-sistem.com');
    DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'e2e-%@hr-sistem.com');
    DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'e2e-%@hr-sistem.com');
    DELETE FROM users WHERE email LIKE 'e2e-%@hr-sistem.com';
  " >/dev/null
}

if [ "$1" = "--clean" ]; then
  cleanup
  echo "e2e test data removed"
  exit 0
fi

# wait for the API
i=0
until curl -sf -o /dev/null "$API_BASE/absences/types"; do
  i=$((i + 1))
  if [ "$i" -gt 30 ]; then
    echo "ERROR: API not reachable at $API_BASE" >&2
    exit 1
  fi
  sleep 1
done

echo "resetting e2e accounts..."
cleanup

register()  { curl -s -X POST "$API_BASE/register" -H 'Content-Type: application/json' -d "{\"email\":\"$1\",\"password_hash\":\"$PASSWORD\"}" >/dev/null; }
token()     { curl -s -X POST "$API_BASE/login" -H 'Content-Type: application/json' -d "{\"email\":\"$1\",\"password_hash\":\"$PASSWORD\"}" | jq -r '.data.accessToken'; }
setrole()   { psql_cmd "DELETE FROM user_roles WHERE user_id=(SELECT id FROM users WHERE email='$1'); INSERT INTO user_roles (user_id, role_id) SELECT u.id, r.id FROM users u, roles r WHERE u.email='$1' AND r.name='$2';" >/dev/null; }

register "$ADMIN_EMAIL"
register "$HR_EMAIL"
register "$MGR_EMAIL"
register "$EMP_EMAIL"

setrole "$ADMIN_EMAIL" PLATFORM_ADMIN
setrole "$HR_EMAIL"    HR_ADMIN
setrole "$MGR_EMAIL"   MANAGER_PORTAL_ACCESS

HR_T="$(token "$HR_EMAIL")"
EMP_T="$(token "$EMP_EMAIL")"
ADMIN_T="$(token "$ADMIN_EMAIL")"
MGR_T="$(token "$MGR_EMAIL")"

# employee profiles (country 182 = Serbia so holidays/business days apply)
curl -s -o /dev/null -X POST "$API_BASE/employees" -H "Authorization: Bearer $ADMIN_T" -H 'Content-Type: application/json' \
  -d '{"first_name":"E2E","last_name":"Admin","country":182,"city":"Beograd","position_id":1}'
curl -s -o /dev/null -X POST "$API_BASE/employees" -H "Authorization: Bearer $MGR_T" -H 'Content-Type: application/json' \
  -d '{"first_name":"E2E","last_name":"Manager","country":182,"city":"Beograd","position_id":1}'
HR_PROF="$(curl -s -X POST "$API_BASE/employees" -H "Authorization: Bearer $HR_T" -H 'Content-Type: application/json' \
  -d '{"first_name":"E2E","last_name":"HR","country":182,"city":"Beograd","position_id":1}' | jq -r '.data.id')"
EMP_PROF="$(curl -s -X POST "$API_BASE/employees" -H "Authorization: Bearer $EMP_T" -H 'Content-Type: application/json' \
  -d '{"first_name":"E2E","last_name":"Employee","country":182,"city":"Beograd","position_id":1}' | jq -r '.data.id')"

# employee reports to HR
psql_cmd "UPDATE employees SET supervisor_id=$HR_PROF WHERE id=$EMP_PROF;" >/dev/null

# give the employee vacation days so request submission succeeds
psql_cmd "INSERT INTO leave_balance (employee_id, absence_type_id, entry_type, days, accrual_year, expires_at)
          VALUES ($EMP_PROF, 1, 'ACCRUAL', 15, 2026, '2027-06-30');" >/dev/null

echo "seeded:"
echo "  admin    $ADMIN_EMAIL / $PASSWORD"
echo "  hr       $HR_EMAIL / $PASSWORD (employee id $HR_PROF)"
echo "  manager  $MGR_EMAIL / $PASSWORD"
echo "  employee $EMP_EMAIL / $PASSWORD (employee id $EMP_PROF, 15 vacation days)"
