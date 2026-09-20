package absence

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	types "main/types/absence"
	"math"
	"time"
)

// ErrOverlap is returned when an employee already has an active absence request
// covering (part of) the requested period.
var ErrOverlap = errors.New("overlapping absence request already exists")

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// ── Absence types ─────────────────────────────────────────────

func (s *Store) GetAllAbsenceTypes() (*types.AbsenceResponse, error) {
	var absenceList []types.AbsenceTypes

	query := `SELECT id, code, type_name, is_paid, status FROM absence_types ORDER BY id;`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a types.AbsenceTypes
		if err := rows.Scan(&a.Id, &a.Code, &a.TypeName, &a.IsPaid, &a.Status); err != nil {
			return nil, err
		}
		absenceList = append(absenceList, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &types.AbsenceResponse{
		Data: absenceList,
	}, nil
}

// ── Absence requests ──────────────────────────────────────────

const requestSelect = `SELECT r.id, r.employee_id, r.absence_type_id, r.start_date, r.end_date,
	r.total_days, r.reason, r.status, r.created_at, r.created_by, r.approved_at, r.approved_by,
	t.type_name, t.is_paid,
	e.first_name, e.last_name
	FROM absence_requests r
	LEFT JOIN absence_types t ON t.id = r.absence_type_id
	LEFT JOIN employees e ON e.id = r.employee_id`

func scanRequest(scanner interface {
	Scan(dest ...any) error
}) (*types.AbsenceRequest, error) {
	var req types.AbsenceRequest
	// DATE columns are scanned as time.Time and normalized to YYYY-MM-DD so the
	// value can be parsed and re-used internally (and rendered cleanly by clients).
	var start, end time.Time
	err := scanner.Scan(
		&req.ID, &req.EmployeeID, &req.AbsenceTypeID,
		&start, &end, &req.TotalDays,
		&req.Reason, &req.Status, &req.CreatedAt, &req.CreatedBy,
		&req.ApprovedAt, &req.ApprovedBy,
		&req.TypeName, &req.IsPaid,
		&req.FirstName, &req.LastName,
	)
	if err != nil {
		return nil, err
	}
	req.StartDate = start.Format("2006-01-02")
	req.EndDate = end.Format("2006-01-02")
	return &req, nil
}

// querier is satisfied by both *sql.DB and *sql.Tx.
type querier interface {
	QueryRow(query string, args ...any) *sql.Row
}

// hasOverlap reports whether the employee already has a DRAFT, PENDING or
// APPROVED request intersecting [start, end]. excludeID ignores one request
// (used when editing it).
func hasOverlap(q querier, employeeID uint, start, end string, excludeID int64) (bool, error) {
	var exists bool
	err := q.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM absence_requests
			WHERE employee_id = $1
			  AND status IN ('DRAFT','PENDING','APPROVED')
			  AND start_date <= $3::date
			  AND end_date   >= $2::date
			  AND ($4 = 0 OR id <> $4)
		)`, employeeID, start, end, excludeID).Scan(&exists)
	return exists, err
}

func (s *Store) CreateRequest(req types.AbsenceRequest) (*types.AbsenceRequest, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	overlap, err := hasOverlap(tx, req.EmployeeID, req.StartDate, req.EndDate, 0)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrOverlap
	}

	var id uint
	err = tx.QueryRow(
		`INSERT INTO absence_requests (employee_id, absence_type_id, start_date, end_date, total_days, reason, status, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		req.EmployeeID, req.AbsenceTypeID, req.StartDate, req.EndDate,
		req.TotalDays, req.Reason, req.Status, req.CreatedBy,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetRequestByID(int64(id))
}

// UpdateRequest edits an existing request (owner flow: DRAFT/PENDING only) and
// recomputes its fields. It rejects periods overlapping other active requests.
func (s *Store) UpdateRequest(req types.AbsenceRequest) (*types.AbsenceRequest, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	overlap, err := hasOverlap(tx, req.EmployeeID, req.StartDate, req.EndDate, int64(req.ID))
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrOverlap
	}

	result, err := tx.Exec(`
		UPDATE absence_requests
		SET absence_type_id = $1, start_date = $2, end_date = $3,
		    total_days = $4, reason = $5, status = $6
		WHERE id = $7`,
		req.AbsenceTypeID, req.StartDate, req.EndDate,
		req.TotalDays, req.Reason, req.Status, req.ID)
	if err != nil {
		return nil, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return nil, sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetRequestByID(int64(req.ID))
}

// GetHolidays returns the holidays (global or country-specific to the employee)
// in the [from, to] range, keyed by YYYY-MM-DD.
func (s *Store) GetHolidays(employeeID uint, from, to string) (map[string]bool, error) {
	rows, err := s.db.Query(`
		SELECT h.holiday_date::text
		FROM holidays h
		WHERE h.holiday_date BETWEEN $2::date AND $3::date
		  AND (h.country_id IS NULL
		       OR h.country_id = (SELECT country FROM employees WHERE id = $1))`,
		employeeID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	holidays := map[string]bool{}
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		holidays[d] = true
	}
	return holidays, rows.Err()
}

func (s *Store) GetRequestByID(id int64) (*types.AbsenceRequest, error) {
	req, err := scanRequest(s.db.QueryRow(
		requestSelect+` WHERE r.id = $1`, id,
	))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Store) GetRequestsByEmployee(employeeID uint) ([]types.AbsenceRequest, error) {
	rows, err := s.db.Query(
		requestSelect+` WHERE r.employee_id = $1 ORDER BY r.created_at DESC`, employeeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []types.AbsenceRequest
	for rows.Next() {
		req, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, *req)
	}
	if requests == nil {
		requests = []types.AbsenceRequest{}
	}
	return requests, rows.Err()
}

func (s *Store) GetAllRequests() ([]types.AbsenceRequest, error) {
	rows, err := s.db.Query(requestSelect + ` ORDER BY r.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []types.AbsenceRequest
	for rows.Next() {
		req, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, *req)
	}
	if requests == nil {
		requests = []types.AbsenceRequest{}
	}
	return requests, rows.Err()
}

func (s *Store) UpdateRequestStatus(id int64, status string, approvedBy *uint) (*types.AbsenceRequest, error) {
	log.Printf("Updating absence request %d to status %s", id, status)

	// Load the request first so we know its previous state, days, and type
	req, err := s.GetRequestByID(id)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, sql.ErrNoRows
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var result sql.Result
	if status == "APPROVED" || status == "REJECTED" {
		result, err = tx.Exec(
			`UPDATE absence_requests SET status = $1, approved_by = $2, approved_at = NOW()
			 WHERE id = $3`,
			status, approvedBy, id,
		)
	} else {
		result, err = tx.Exec(
			`UPDATE absence_requests SET status = $1 WHERE id = $2`,
			status, id,
		)
	}

	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, sql.ErrNoRows
	}

	// ── Ledger side effects ───────────────────────────────────────
	// APPROVED  → consume days (negative ledger entry)
	if status == "APPROVED" && req.Status != "APPROVED" {
		if err := s.addLedgerEntryTx(tx, types.LeaveBalanceEntry{
			EmployeeID:    req.EmployeeID,
			AbsenceTypeID: req.AbsenceTypeID,
			EntryType:     "CONSUMED",
			Days:          -req.TotalDays,
			AccrualYear:   yearOf(req.StartDate),
			ReferenceID:   &req.ID,
		}); err != nil {
			return nil, fmt.Errorf("ledger consume: %w", err)
		}
	}

	// CANCELLED (from APPROVED) → give the days back (+)
	if status == "CANCELLED" && req.Status == "APPROVED" {
		if err := s.addLedgerEntryTx(tx, types.LeaveBalanceEntry{
			EmployeeID:    req.EmployeeID,
			AbsenceTypeID: req.AbsenceTypeID,
			EntryType:     "CANCELLED",
			Days:          req.TotalDays,
			AccrualYear:   yearOf(req.StartDate),
			ReferenceID:   &req.ID,
		}); err != nil {
			return nil, fmt.Errorf("ledger restore: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetRequestByID(id)
}

// yearOf extracts the year from a YYYY-MM-DD date string.
func yearOf(date string) int {
	if t, err := time.Parse("2006-01-02", date); err == nil {
		return t.Year()
	}
	return time.Now().Year()
}

// ── Leave balance ─────────────────────────────────────────────

func (s *Store) GetBalanceByEmployee(employeeID uint) ([]types.LeaveBalanceSummary, error) {
	rows, err := s.db.Query(`
		SELECT lb.absence_type_id,
		       t.type_name,
		       t.code,
		       t.is_paid,
		       COALESCE(SUM(CASE WHEN lb.entry_type = 'CONSUMED' THEN 0 ELSE lb.days END), 0) AS granted,
		       COALESCE(SUM(CASE WHEN lb.entry_type = 'CONSUMED' THEN -lb.days ELSE 0 END), 0) AS used,
		       COALESCE(SUM(lb.days), 0) AS available
		FROM leave_balance lb
		JOIN absence_types t ON t.id = lb.absence_type_id
		WHERE lb.employee_id = $1
		  AND (lb.expires_at IS NULL OR lb.expires_at >= CURRENT_DATE)
		GROUP BY lb.absence_type_id, t.type_name, t.code, t.is_paid
		ORDER BY t.type_name`, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []types.LeaveBalanceSummary
	for rows.Next() {
		var s types.LeaveBalanceSummary
		if err := rows.Scan(
			&s.AbsenceTypeID, &s.TypeName, &s.Code, &s.IsPaid,
			&s.GrantedDays, &s.UsedDays, &s.AvailableDays,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	if summaries == nil {
		summaries = []types.LeaveBalanceSummary{}
	}
	return summaries, rows.Err()
}

// ── Leave balance ─────────────────────────────────────────────

func (s *Store) AddLedgerEntry(entry types.LeaveBalanceEntry) (*types.LeaveBalanceEntry, error) {
	var id uint
	err := s.db.QueryRow(
		`INSERT INTO leave_balance (employee_id, absence_type_id, entry_type, days, accrual_year, expires_at, reference_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		entry.EmployeeID, entry.AbsenceTypeID, entry.EntryType, entry.Days,
		entry.AccrualYear, entry.ExpiresAt, entry.ReferenceID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	var e types.LeaveBalanceEntry
	err = s.db.QueryRow(`
		SELECT lb.id, lb.employee_id, lb.absence_type_id, lb.entry_type, lb.days,
		       lb.accrual_year, lb.expires_at, lb.reference_id, lb.created_at,
		       t.type_name
		FROM leave_balance lb
		JOIN absence_types t ON t.id = lb.absence_type_id
		WHERE lb.id = $1`, id,
	).Scan(
		&e.ID, &e.EmployeeID, &e.AbsenceTypeID, &e.EntryType, &e.Days,
		&e.AccrualYear, &e.ExpiresAt, &e.ReferenceID, &e.CreatedAt,
		&e.TypeName,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// addLedgerEntryTx inserts a ledger row inside the given transaction.
func (s *Store) addLedgerEntryTx(tx *sql.Tx, entry types.LeaveBalanceEntry) error {
	_, err := tx.Exec(
		`INSERT INTO leave_balance (employee_id, absence_type_id, entry_type, days, accrual_year, expires_at, reference_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.EmployeeID, entry.AbsenceTypeID, entry.EntryType, entry.Days,
		entry.AccrualYear, entry.ExpiresAt, entry.ReferenceID,
	)
	return err
}

// ── Leave policies ────────────────────────────────────────────

func (s *Store) GetPolicies() ([]types.LeavePolicy, error) {
	rows, err := s.db.Query(`
		SELECT id, name, grant_policy, days_per_period, period_months,
		       allow_carry_over, carry_over_max_days, carry_over_expiry_month, carry_over_expiry_day,
		       requires_balance
		FROM leave_policies ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []types.LeavePolicy
	for rows.Next() {
		var p types.LeavePolicy
		if err := rows.Scan(
			&p.ID, &p.Name, &p.GrantPolicy,
			&p.DaysPerPeriod, &p.PeriodMonths,
			&p.AllowCarryOver, &p.CarryOverMaxDays, &p.CarryOverExpiryMonth, &p.CarryOverExpiryDay,
			&p.RequiresBalance,
		); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	if policies == nil {
		policies = []types.LeavePolicy{}
	}
	return policies, rows.Err()
}

// GetActivePolicy returns the policy that applies to an employee for an absence
// type on a given date. It checks employee_leave_policy first; if no explicit
// assignment exists, it falls back to the system default by absence type code.
//
// A nil policy with a nil error means "no rule is configured for this type"
// (e.g. TRAINING or PERSONAL have no default). Callers must decide what to do —
// the balance check treats such a type as balance-required (fail closed).
func (s *Store) GetActivePolicy(employeeID, absenceTypeID uint, atDate string) (*types.LeavePolicy, error) {
	// 1) explicit assignment valid on that date
	var p types.LeavePolicy
	err := s.db.QueryRow(`
		SELECT lp.id, lp.name, lp.grant_policy, lp.days_per_period, lp.period_months,
		       lp.allow_carry_over, lp.carry_over_max_days, lp.carry_over_expiry_month, lp.carry_over_expiry_day,
		       lp.requires_balance
		FROM employee_leave_policy elp
		JOIN leave_policies lp ON lp.id = elp.policy_id
		WHERE elp.employee_id = $1 AND elp.absence_type_id = $2
		  AND elp.valid_from <= $3::date
		  AND (elp.valid_to IS NULL OR elp.valid_to >= $3::date)
		ORDER BY elp.valid_from DESC LIMIT 1`,
		employeeID, absenceTypeID, atDate,
	).Scan(
		&p.ID, &p.Name, &p.GrantPolicy,
		&p.DaysPerPeriod, &p.PeriodMonths,
		&p.AllowCarryOver, &p.CarryOverMaxDays, &p.CarryOverExpiryMonth, &p.CarryOverExpiryDay,
		&p.RequiresBalance,
	)
	if err == nil {
		return &p, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// 2) fall back to default policy per absence type code
	code := ""
	_ = s.db.QueryRow(`SELECT code FROM absence_types WHERE id = $1`, absenceTypeID).Scan(&code)

	defaultPolicy := map[string]string{
		"VACATION": "Vacation standard",
		"SICK":     "Sick unlimited",
		"PARENTAL": "Parental leave",
	}
	defaultName, ok := defaultPolicy[code]
	if !ok {
		// No explicit assignment and no system default for this absence type:
		// report "no policy" (nil, nil) instead of an error, so callers can tell
		// a missing rule apart from a broken lookup.
		return nil, nil
	}

	err = s.db.QueryRow(`
		SELECT id, name, grant_policy, days_per_period, period_months,
		       allow_carry_over, carry_over_max_days, carry_over_expiry_month, carry_over_expiry_day,
		       requires_balance
		FROM leave_policies WHERE name = $1`, defaultName,
	).Scan(
		&p.ID, &p.Name, &p.GrantPolicy,
		&p.DaysPerPeriod, &p.PeriodMonths,
		&p.AllowCarryOver, &p.CarryOverMaxDays, &p.CarryOverExpiryMonth, &p.CarryOverExpiryDay,
		&p.RequiresBalance,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("default policy '%s' not found in DB", defaultName)
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) AssignEmployeePolicy(a types.EmployeePolicyAssignment) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// close the currently-open assignment for the same employee+type (if any)
	_, err = tx.Exec(`
		UPDATE employee_leave_policy
		SET valid_to = $1::date - INTERVAL '1 day'
		WHERE employee_id = $2 AND absence_type_id = $3 AND valid_to IS NULL`,
		a.ValidFrom, a.EmployeeID, a.AbsenceTypeID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO employee_leave_policy (employee_id, absence_type_id, policy_id, valid_from, valid_to)
		VALUES ($1, $2, $3, $4, $5)`,
		a.EmployeeID, a.AbsenceTypeID, a.PolicyID, a.ValidFrom, a.ValidTo)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) GetEmployeePolicyAssignments(employeeID uint) ([]types.EmployeePolicyAssignment, error) {
	rows, err := s.db.Query(`
		SELECT elp.id, elp.employee_id, elp.absence_type_id, elp.policy_id, elp.valid_from, elp.valid_to,
		       at.type_name, lp.name
		FROM employee_leave_policy elp
		JOIN absence_types at ON at.id = elp.absence_type_id
		JOIN leave_policies lp ON lp.id = elp.policy_id
		WHERE elp.employee_id = $1
		ORDER BY elp.valid_from DESC`, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []types.EmployeePolicyAssignment
	for rows.Next() {
		var a types.EmployeePolicyAssignment
		if err := rows.Scan(
			&a.ID, &a.EmployeeID, &a.AbsenceTypeID, &a.PolicyID, &a.ValidFrom, &a.ValidTo,
			&a.AbsenceTypeName, &a.PolicyName,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	if list == nil {
		list = []types.EmployeePolicyAssignment{}
	}
	return list, rows.Err()
}

// GetAvailableForType returns the live available days for one absence type.
func (s *Store) GetAvailableForType(employeeID, absenceTypeID uint) (float64, error) {
	var avail float64
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(days), 0)
		FROM leave_balance
		WHERE employee_id = $1 AND absence_type_id = $2
		  AND (expires_at IS NULL OR expires_at >= CURRENT_DATE)`, employeeID, absenceTypeID,
	).Scan(&avail)
	return avail, err
}

// ── Annual rollover: accrual + carry-over + expiry ──────────────
//
// For target year Y (run ~Jan 1 of Y, idempotent per employee+type):
//  1. If an ACCRUAL already exists for (emp, type, Y) → skip.
//  2. Compute leftover of year Y-1 (net unused days).
//     • if carry-over allowed → keep min(leftover, cap);
//     leftover beyond the cap is written as EXPIRATION (-)
//     • if NOT allowed → all leftover is written as EXPIRATION (-)
//  3. Grant the new ACCRUAL for year Y.
func (s *Store) RolloverYear(year int) (*types.RolloverReport, error) {
	report := &types.RolloverReport{
		Year:    year,
		Accrued: []types.RolloverAction{},
		Carried: []types.RolloverAction{},
		Expired: []types.RolloverAction{},
	}

	prevYear := year - 1
	grantDate := fmt.Sprintf("%04d-01-01", year)

	// gather all (employee, absence_type) pairs that have any activity or assignment
	pairs, err := s.activeLeavePairs()
	if err != nil {
		return nil, err
	}

	for _, pair := range pairs {
		// resolve the policy active on Jan 1 of the target year
		policy, err := s.GetActivePolicy(pair.employeeID, pair.absenceTypeID, grantDate)
		if err != nil || policy == nil {
			report.SkippedCnt++
			continue
		}

		if policy.GrantPolicy != "YEARLY_GRANT" && policy.GrantPolicy != "MATERNITY_GRANT" {
			report.SkippedCnt++
			continue
		}
		if policy.DaysPerPeriod == nil {
			report.SkippedCnt++
			continue
		}

		typeName := s.typeName(pair.absenceTypeID)

		// MATERNITY: grant once ever (skip if any ACCRUAL exists)
		if policy.GrantPolicy == "MATERNITY_GRANT" {
			var exists bool
			_ = s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM leave_balance WHERE employee_id=$1 AND absence_type_id=$2 AND entry_type='ACCRUAL')`,
				pair.employeeID, pair.absenceTypeID).Scan(&exists)
			if exists {
				report.SkippedCnt++
				continue
			}
			if err := s.addLedgerEntry(types.LeaveBalanceEntry{
				EmployeeID:    pair.employeeID,
				AbsenceTypeID: pair.absenceTypeID,
				EntryType:     "ACCRUAL",
				Days:          *policy.DaysPerPeriod,
				AccrualYear:   year,
			}); err != nil {
				return nil, err
			}
			report.Accrued = append(report.Accrued, types.RolloverAction{pair.employeeID, pair.absenceTypeID, typeName, *policy.DaysPerPeriod})
			continue
		}

		// YEARLY_GRANT:
		// skip if we already accrued this year (idempotency)
		var accrued bool
		_ = s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM leave_balance WHERE employee_id=$1 AND absence_type_id=$2 AND entry_type='ACCRUAL' AND accrual_year=$3)`,
			pair.employeeID, pair.absenceTypeID, year).Scan(&accrued)
		if accrued {
			report.SkippedCnt++
			continue
		}

		// leftover from previous year (net): grants + carry of prev year + consumed etc.
		leftover, err := s.netLeftover(pair.employeeID, pair.absenceTypeID, prevYear)
		if err != nil {
			return nil, err
		}

		if leftover > 0 {
			keep := 0.0
			if policy.AllowCarryOver && policy.CarryOverMaxDays != nil {
				keep = math.Min(leftover, *policy.CarryOverMaxDays)
			} else if policy.AllowCarryOver {
				keep = leftover
			}

			// expiry date for carried days = policy carry-over date of the target year
			carryExpiry := fmt.Sprintf("%04d-%02d-%02d",
				year, intVal(policy.CarryOverExpiryMonth, 6), intVal(policy.CarryOverExpiryDay, 30))

			if keep > 0.001 {
				// push the kept (still-open) rows of the previous year to the carry expiry date
				if _, err := s.db.Exec(`
					UPDATE leave_balance SET expires_at = $1
					WHERE employee_id = $2 AND absence_type_id = $3 AND accrual_year = $4
					  AND entry_type IN ('ACCRUAL','CARRY_OVER','MANUAL_ADJUST')
					  AND days > 0 AND (expires_at IS NULL OR expires_at > $1)`, carryExpiry,
					pair.employeeID, pair.absenceTypeID, prevYear); err != nil {
					return nil, err
				}
				report.Carried = append(report.Carried, types.RolloverAction{pair.employeeID, pair.absenceTypeID, typeName, keep})
			}

			// expire whatever is beyond what we keep
			excess := leftover - keep
			if excess > 0.001 {
				if err := s.addLedgerEntry(types.LeaveBalanceEntry{
					EmployeeID:    pair.employeeID,
					AbsenceTypeID: pair.absenceTypeID,
					EntryType:     "EXPIRATION",
					Days:          -excess,
					AccrualYear:   prevYear,
				}); err != nil {
					return nil, err
				}
				report.Expired = append(report.Expired, types.RolloverAction{pair.employeeID, pair.absenceTypeID, typeName, excess})
			}
		} else if !policy.AllowCarryOver {
			// no carry-over policy but no leftover anyway: nothing to do
		}

		// accrue the new year
		if err := s.addLedgerEntry(types.LeaveBalanceEntry{
			EmployeeID:    pair.employeeID,
			AbsenceTypeID: pair.absenceTypeID,
			EntryType:     "ACCRUAL",
			Days:          *policy.DaysPerPeriod,
			AccrualYear:   year,
			ExpiresAt:     ptrStr(expiryForPolicy(policy, year)),
		}); err != nil {
			return nil, err
		}
		report.Accrued = append(report.Accrued, types.RolloverAction{pair.employeeID, pair.absenceTypeID, typeName, *policy.DaysPerPeriod})
	}

	return report, nil
}

type leavePair struct {
	employeeID    uint
	absenceTypeID uint
}

func (s *Store) activeLeavePairs() ([]leavePair, error) {
	rows, err := s.db.Query(`
		SELECT employee_id, absence_type_id FROM employee_leave_policy
		UNION
		SELECT employee_id, absence_type_id FROM leave_balance
		UNION
		SELECT DISTINCT e.id, 1 FROM employees e`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pairs []leavePair
	seen := map[string]bool{}
	for rows.Next() {
		var p leavePair
		if err := rows.Scan(&p.employeeID, &p.absenceTypeID); err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%d:%d", p.employeeID, p.absenceTypeID)
		if !seen[key] {
			seen[key] = true
			pairs = append(pairs, p)
		}
	}
	return pairs, rows.Err()
}

func (s *Store) netLeftover(employeeID, absenceTypeID uint, year int) (float64, error) {
	var total float64
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(days), 0)
		FROM leave_balance
		WHERE employee_id = $1 AND absence_type_id = $2 AND accrual_year = $3`,
		employeeID, absenceTypeID, year).Scan(&total)
	return total, err
}

func (s *Store) typeName(absenceTypeID uint) string {
	var n string
	_ = s.db.QueryRow(`SELECT type_name FROM absence_types WHERE id = $1`, absenceTypeID).Scan(&n)
	return n
}

// expiryForPolicy computes the expiry date for days granted in `year`:
// normally the same calendar year (no expiry), but for vacation the accrual
// window extends to the policy's carry-over expiry (e.g. 30 June of year+1).
func expiryForPolicy(p *types.LeavePolicy, year int) string {
	month, day := 12, 31
	yr := year
	if p.AllowCarryOver && p.CarryOverExpiryMonth != nil && p.CarryOverExpiryDay != nil {
		month = *p.CarryOverExpiryMonth
		day = *p.CarryOverExpiryDay
		yr = year + 1
	}
	return fmt.Sprintf("%04d-%02d-%02d", yr, month, day)
}

func ptrStr(s string) *string { return &s }

func intVal(p *int, fallback int) int {
	if p != nil {
		return *p
	}
	return fallback
}

// addLedgerEntry non-transactional insert used by rollover (single statements).
func (s *Store) addLedgerEntry(entry types.LeaveBalanceEntry) error {
	_, err := s.db.Exec(
		`INSERT INTO leave_balance (employee_id, absence_type_id, entry_type, days, accrual_year, expires_at, reference_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.EmployeeID, entry.AbsenceTypeID, entry.EntryType, entry.Days,
		entry.AccrualYear, entry.ExpiresAt, entry.ReferenceID,
	)
	return err
}

// ── Scheduled jobs: idempotency + monthly accrual + expiry ─────

// TryMarkScheduledRun atomically records that (job, period, subject) has run.
// It reports true only for the caller that created the marker, so concurrent
// instances cannot run the same job twice.
func (s *Store) TryMarkScheduledRun(job, period, subject string) (bool, error) {
	res, err := s.db.Exec(`
		INSERT INTO scheduler_runs (job, period, subject)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`, job, period, subject)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// RunMonthlyAccrual grants days for MONTHLY_GRANT policies for the given period
// ("YYYY-MM"). Idempotent per employee+absence type via scheduler_runs.
func (s *Store) RunMonthlyAccrual(period string) (*types.RolloverReport, error) {
	report := &types.RolloverReport{
		Accrued: []types.RolloverAction{},
		Carried: []types.RolloverAction{},
		Expired: []types.RolloverAction{},
	}

	atDate := period + "-01"
	year := yearOf(atDate)
	report.Year = year

	pairs, err := s.activeLeavePairs()
	if err != nil {
		return nil, err
	}

	for _, pair := range pairs {
		policy, err := s.GetActivePolicy(pair.employeeID, pair.absenceTypeID, atDate)
		if err != nil || policy == nil || policy.GrantPolicy != "MONTHLY_GRANT" || policy.DaysPerPeriod == nil {
			report.SkippedCnt++
			continue
		}

		subject := fmt.Sprintf("%d:%d", pair.employeeID, pair.absenceTypeID)
		marked, err := s.TryMarkScheduledRun("monthly_accrual", period, subject)
		if err != nil {
			return nil, err
		}
		if !marked {
			report.SkippedCnt++
			continue
		}

		if err := s.addLedgerEntry(types.LeaveBalanceEntry{
			EmployeeID:    pair.employeeID,
			AbsenceTypeID: pair.absenceTypeID,
			EntryType:     "ACCRUAL",
			Days:          *policy.DaysPerPeriod,
			AccrualYear:   year,
			ExpiresAt:     ptrStr(expiryForPolicy(policy, year)),
		}); err != nil {
			return nil, err
		}

		report.Accrued = append(report.Accrued, types.RolloverAction{
			EmployeeID:    pair.employeeID,
			AbsenceTypeID: pair.absenceTypeID,
			TypeName:      s.typeName(pair.absenceTypeID),
			Days:          *policy.DaysPerPeriod,
		})
	}

	return report, nil
}

// RunExpiration writes a single EXPIRATION entry for every (employee, type,
// year) group whose balance is still positive but whose expirable rows have all
// passed their expiry date. It expires the group's NET remaining days, so it
// composes with the partial expiry already written by carry-over. Idempotent:
// once the EXPIRATION row is written the group's net becomes ~0.
func (s *Store) RunExpiration() (*types.RolloverReport, error) {
	report := &types.RolloverReport{
		Accrued: []types.RolloverAction{},
		Carried: []types.RolloverAction{},
		Expired: []types.RolloverAction{},
	}

	rows, err := s.db.Query(`
		SELECT employee_id, absence_type_id, accrual_year, SUM(days)
		FROM leave_balance
		GROUP BY employee_id, absence_type_id, accrual_year
		HAVING SUM(days) > 0.001
		   AND COUNT(*) FILTER (WHERE expires_at IS NOT NULL) > 0
		   AND COUNT(*) FILTER (WHERE expires_at >= CURRENT_DATE) = 0`)
	if err != nil {
		return nil, err
	}

	type candidate struct {
		employeeID    uint
		absenceTypeID uint
		year          int
		days          float64
	}
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.employeeID, &c.absenceTypeID, &c.year, &c.days); err != nil {
			rows.Close()
			return nil, err
		}
		candidates = append(candidates, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, c := range candidates {
		subject := fmt.Sprintf("%d:%d", c.employeeID, c.absenceTypeID)
		marked, err := s.TryMarkScheduledRun("expiration", fmt.Sprintf("%d", c.year), subject)
		if err != nil {
			return nil, err
		}
		if !marked {
			report.SkippedCnt++
			continue
		}

		if err := s.addLedgerEntry(types.LeaveBalanceEntry{
			EmployeeID:    c.employeeID,
			AbsenceTypeID: c.absenceTypeID,
			EntryType:     "EXPIRATION",
			Days:          -c.days,
			AccrualYear:   c.year,
		}); err != nil {
			return nil, err
		}

		report.Expired = append(report.Expired, types.RolloverAction{
			EmployeeID:    c.employeeID,
			AbsenceTypeID: c.absenceTypeID,
			TypeName:      s.typeName(c.absenceTypeID),
			Days:          c.days,
		})
	}

	return report, nil
}
