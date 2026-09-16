#!/bin/sh
# Verify that a user with multiple roles gets a single deterministic role
# (highest privilege wins) on every login/refresh.
BASE=http://localhost:8034/api/v1
PASS='TestTest1!'
TS=$(date +%s)
JQ=/usr/bin/jq
BOTH="rolefix-$TS@hr-sistem.com"
ONLY="employeefix-$TS@hr-sistem.com"

reg()  { curl -s -X POST "$BASE/register" -H 'Content-Type: application/json' -d "{\"email\":\"$1\",\"password_hash\":\"$PASS\"}"; }
login(){ curl -s -X POST "$BASE/login"    -H 'Content-Type: application/json' -d "{\"email\":\"$1\",\"password_hash\":\"$PASS\"}"; }
sqlq() { podman exec hr_db psql -U hr_user -d hr_db -t -A -c "$1"; }

BOTH_ID=$(reg "$BOTH" | "$JQ" -r .id)
ONLY_ID=$(reg "$ONLY" | "$JQ" -r .id)

# give the "both" user HR_ADMIN on top of the auto-assigned EMPLOYEE
sqlq "INSERT INTO user_roles (user_id, role_id) SELECT $BOTH_ID, id FROM roles WHERE name='HR_ADMIN' ON CONFLICT DO NOTHING;" >/dev/null

echo "roles for both-user:"
sqlq "SELECT r.name FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=$BOTH_ID ORDER BY r.name;"

echo; echo "5 logins as the two-role user (expect HR_ADMIN every time):"
i=0
while [ $i -lt 5 ]; do
  login "$BOTH" | "$JQ" -r '.data.user.role'
  i=$((i+1))
done

BOTH_T=$(login "$BOTH" | "$JQ" -r '.data.accessToken')
ONLY_T=$(login "$ONLY" | "$JQ" -r '.data.accessToken')

echo; echo "effective roles:"
echo "  both-user   -> $(login "$BOTH" | "$JQ" -r '.data.user.role')"
echo "  only-user   -> $(login "$ONLY" | "$JQ" -r '.data.user.role')"

echo; echo "GET /employees (hr sees it, employee must not):"
echo "  both-user  HTTP $(curl -s -o /dev/null -w '%{http_code}' "$BASE/employees" -H "Authorization: Bearer $BOTH_T")"
echo "  only-user  HTTP $(curl -s -o /dev/null -w '%{http_code}' "$BASE/employees" -H "Authorization: Bearer $ONLY_T")"

echo; echo "GET /attendance (admin only):"
echo "  both-user  HTTP $(curl -s -o /dev/null -w '%{http_code}' "$BASE/attendance" -H "Authorization: Bearer $BOTH_T")"
echo "  only-user  HTTP $(curl -s -o /dev/null -w '%{http_code}' "$BASE/attendance" -H "Authorization: Bearer $ONLY_T")"

sqlq "DELETE FROM refresh_tokens WHERE user_id IN ($BOTH_ID,$ONLY_ID); DELETE FROM user_roles WHERE user_id IN ($BOTH_ID,$ONLY_ID); DELETE FROM users WHERE id IN ($BOTH_ID,$ONLY_ID);" >/dev/null
echo; echo "cleanup done"
