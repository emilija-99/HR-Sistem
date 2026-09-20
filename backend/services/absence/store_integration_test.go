package absence

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	types "main/types/absence"
)

// openTestDB connects to the database from TEST_DATABASE_URL, skipping the test
// when it is not configured or unreachable. Run these with scripts/run-go-tests.sh.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("database not reachable: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// seedEmployee inserts a throwaway user+employee and returns the employee id,
// cleaning everything up when the test finishes.
func seedEmployee(t *testing.T, db *sql.DB) uint {
	t.Helper()

	var userID uint
	err := db.QueryRow(
		`INSERT INTO users (email, password_hash) VALUES ($1, 'x') RETURNING id`,
		fmt.Sprintf("gotest-%d@hr-sistem.com", time.Now().UnixNano()),
	).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	var empID uint
	err = db.QueryRow(
		`INSERT INTO employees (user_id, first_name, last_name, country, hire_date)
		 VALUES ($1, 'Go', 'Test', 182, CURRENT_DATE) RETURNING id`,
		userID,
	).Scan(&empID)
	if err != nil {
		t.Fatalf("insert employee: %v", err)
	}

	t.Cleanup(func() {
		db.Exec(`DELETE FROM absence_requests WHERE employee_id = $1 OR created_by = $1`, empID)
		db.Exec(`DELETE FROM leave_balance WHERE employee_id = $1`, empID)
		db.Exec(`DELETE FROM employee_leave_policy WHERE employee_id = $1`, empID)
		db.Exec(`DELETE FROM attendance WHERE employee_id = $1`, empID)
		db.Exec(`DELETE FROM user_roles WHERE user_id = $1`, userID)
		db.Exec(`DELETE FROM users WHERE id = $1`, userID) // cascades employees
	})

	return empID
}

func TestCreateRequestRejectsOverlap(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)
	empID := seedEmployee(t, db)

	base := types.AbsenceRequest{
		EmployeeID:    empID,
		AbsenceTypeID: 1,
		StartDate:     "2026-06-01",
		EndDate:       "2026-06-05",
		TotalDays:     5,
		Status:        "PENDING",
		CreatedBy:     empID,
	}
	if _, err := store.CreateRequest(base); err != nil {
		t.Fatalf("first request should succeed: %v", err)
	}

	overlapping := base
	overlapping.StartDate = "2026-06-03"
	overlapping.EndDate = "2026-06-09"
	if _, err := store.CreateRequest(overlapping); !errors.Is(err, ErrOverlap) {
		t.Fatalf("expected ErrOverlap, got %v", err)
	}

	adjacent := base
	adjacent.StartDate = "2026-06-08"
	adjacent.EndDate = "2026-06-12"
	if _, err := store.CreateRequest(adjacent); err != nil {
		t.Fatalf("adjacent (non-overlapping) request should succeed: %v", err)
	}
}

func TestGetAvailableForType(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)
	empID := seedEmployee(t, db)

	if _, err := db.Exec(
		`INSERT INTO leave_balance (employee_id, absence_type_id, entry_type, days, accrual_year)
		 VALUES ($1, 1, 'ACCRUAL', 10, 2026),
		        ($1, 1, 'CONSUMED', -4, 2026)`, empID,
	); err != nil {
		t.Fatalf("seed ledger: %v", err)
	}

	avail, err := store.GetAvailableForType(empID, 1)
	if err != nil {
		t.Fatalf("GetAvailableForType: %v", err)
	}
	if avail != 6 {
		t.Fatalf("available = %v, want 6", avail)
	}
}

// regression: the exact reported case — TRAINING (type 4) has no policy and no
// ledger entries, yet a 151-day request used to pass because the missing policy
// was treated as "no balance required".
func TestEnforceBalanceFailsClosedForTypeWithoutPolicy(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)
	empID := seedEmployee(t, db)
	h := NewHandler(db, store, nil, nil, nil)

	const (
		training = 4
		sick     = 3
		start    = "2026-11-09"
	)

	t.Run("151 days without any balance is rejected", func(t *testing.T) {
		err := h.enforceBalance(empID, training, start, 151)
		if !errors.Is(err, ErrInsufficientBalance) {
			t.Fatalf("expected ErrInsufficientBalance, got %v", err)
		}
	})

	// grant 5 training days
	if _, err := db.Exec(
		`INSERT INTO leave_balance (employee_id, absence_type_id, entry_type, days, accrual_year)
		 VALUES ($1, $2, 'ACCRUAL', 5, 2026)`, empID, training,
	); err != nil {
		t.Fatalf("seed ledger: %v", err)
	}

	t.Run("within the granted days is allowed", func(t *testing.T) {
		if err := h.enforceBalance(empID, training, start, 3); err != nil {
			t.Fatalf("3 of 5 days should be allowed, got %v", err)
		}
	})

	t.Run("above the granted days is rejected", func(t *testing.T) {
		if err := h.enforceBalance(empID, training, start, 6); !errors.Is(err, ErrInsufficientBalance) {
			t.Fatalf("expected ErrInsufficientBalance, got %v", err)
		}
	})

	t.Run("a policy with requires_balance=false stays unlimited", func(t *testing.T) {
		if err := h.enforceBalance(empID, sick, start, 999); err != nil {
			t.Fatalf("sick leave must not be limited by balance, got %v", err)
		}
	})
}

// regression: a type without an explicit assignment and without a system
// default (TRAINING = 4) must report "no policy" as (nil, nil) — not as an
// error, which callers used to swallow into an unlimited balance check.
func TestGetActivePolicyNoRuleReturnsNil(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)
	empID := seedEmployee(t, db)

	const today = "2026-09-20"

	t.Run("type without any policy", func(t *testing.T) {
		policy, err := store.GetActivePolicy(empID, 4, today)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if policy != nil {
			t.Fatalf("expected nil policy for TRAINING, got %+v", policy)
		}
		if !balanceRequired(policy) {
			t.Fatal("a type without a policy must still require balance")
		}
	})

	t.Run("vacation falls back to the default policy", func(t *testing.T) {
		policy, err := store.GetActivePolicy(empID, 1, today)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if policy == nil {
			t.Fatal("expected the 'Vacation standard' default policy")
		}
		if policy.Name != "Vacation standard" || !policy.RequiresBalance {
			t.Fatalf("unexpected policy: %+v", policy)
		}
	})

	t.Run("sick leave is unlimited by default", func(t *testing.T) {
		policy, err := store.GetActivePolicy(empID, 3, today)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if policy == nil {
			t.Fatal("expected the 'Sick unlimited' default policy")
		}
		if policy.RequiresBalance {
			t.Fatalf("sick leave must not require balance: %+v", policy)
		}
		if balanceRequired(policy) {
			t.Fatal("an unlimited policy must not require balance")
		}
	})
}
