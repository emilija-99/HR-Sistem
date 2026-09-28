package employee

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	types "main/types/employee"
)

// openTestDB connects to TEST_DATABASE_URL and skips the test when it is not
// configured or unreachable. Run these with scripts/run-go-tests.sh.
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

// seedUser inserts a throwaway user (removed, cascading to its employee, on cleanup).
func seedUser(t *testing.T, db *sql.DB) uint {
	t.Helper()
	var id uint
	err := db.QueryRow(
		`INSERT INTO users (email, password_hash) VALUES ($1, 'x') RETURNING id`,
		fmt.Sprintf("gotest-emp-%d@hr-sistem.com", time.Now().UnixNano()),
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM user_roles WHERE user_id = $1`, id)
		db.Exec(`DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

// seedLevelPositions creates an isolated department with ROOKIE/JUNIOR/SENIOR/LEAD
// positions and returns their ids by level. Everything is removed on cleanup, so
// the test does not depend on the seeded organisation.
func seedLevelPositions(t *testing.T, db *sql.DB) map[string]uint {
	t.Helper()

	var deptID uint
	err := db.QueryRow(
		`INSERT INTO departments (id, name)
		 VALUES ((SELECT COALESCE(MAX(id), 0) + 1 FROM departments), $1)
		 RETURNING id`,
		fmt.Sprintf("TestDept-%d", time.Now().UnixMilli()),
	).Scan(&deptID)
	if err != nil {
		t.Fatalf("insert department: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM employees WHERE position_id IN (SELECT id FROM positions WHERE department_id = $1)`, deptID)
		db.Exec(`DELETE FROM positions WHERE department_id = $1`, deptID)
		db.Exec(`DELETE FROM departments WHERE id = $1`, deptID)
	})

	ids := map[string]uint{}
	for _, lvl := range []string{"ROOKIE", "JUNIOR", "SENIOR", "LEAD"} {
		var pid uint
		if err := db.QueryRow(
			`INSERT INTO positions (department_id, title, level)
			 VALUES ($1, 'Go Test Engineer', $2) RETURNING id`,
			deptID, lvl,
		).Scan(&pid); err != nil {
			t.Fatalf("insert position %s: %v", lvl, err)
		}
		ids[lvl] = pid
	}
	return ids
}

const testCountry = 182 // Serbia (seeded by migration 002)

// A new employee without an explicit supervisor is linked to the nearest level
// above them in the same department — e.g. a rookie reports to a senior when that
// is the closest higher level.
func TestCreateAssignsNearestSupervisorByLevel(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)
	pos := seedLevelPositions(t, db)

	seniorPos, rookiePos := pos["SENIOR"], pos["ROOKIE"]

	senior, err := store.Create(types.Employee{
		UserID: seedUser(t, db), FirstName: "Sara", LastName: "Senior",
		Country: testCountry, PositionID: &seniorPos,
	})
	if err != nil {
		t.Fatalf("create senior: %v", err)
	}

	rookie, err := store.Create(types.Employee{
		UserID: seedUser(t, db), FirstName: "Roko", LastName: "Rookie",
		Country: testCountry, PositionID: &rookiePos,
	})
	if err != nil {
		t.Fatalf("create rookie: %v", err)
	}
	if rookie.SupervisorID == nil || *rookie.SupervisorID != senior.ID {
		t.Fatalf("rookie supervisor = %v, want %d", rookie.SupervisorID, senior.ID)
	}
}

// An explicit supervisor chosen by HR is never overwritten by the automatic rule,
// and a LEAD (nobody above them) simply gets no supervisor.
func TestCreateWithUserKeepsExplicitSupervisor(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)
	pos := seedLevelPositions(t, db)

	juniorPos, leadPos := pos["JUNIOR"], pos["LEAD"]

	junior, err := store.Create(types.Employee{
		UserID: seedUser(t, db), FirstName: "Jova", LastName: "Junior",
		Country: testCountry, PositionID: &juniorPos,
	})
	if err != nil {
		t.Fatalf("create junior: %v", err)
	}
	// the junior is the highest level so far → no automatic supervisor
	if junior.SupervisorID != nil {
		t.Fatalf("junior supervisor = %v, want nil", *junior.SupervisorID)
	}

	leadEmail := fmt.Sprintf("gotest-lead-%d@hr-sistem.com", time.Now().UnixNano())
	t.Cleanup(func() {
		db.Exec(`DELETE FROM user_roles WHERE user_id = (SELECT id FROM users WHERE email = $1)`, leadEmail)
		db.Exec(`DELETE FROM users WHERE email = $1`, leadEmail)
	})

	lead, err := store.CreateWithUser(leadEmail, "x", "EMPLOYEE", nil, types.Employee{
		FirstName: "Lea", LastName: "Lead", Country: testCountry,
		PositionID: &leadPos, SupervisorID: &junior.ID,
	})
	if err != nil {
		t.Fatalf("create lead: %v", err)
	}
	if lead.SupervisorID == nil || *lead.SupervisorID != junior.ID {
		t.Fatalf("explicit supervisor = %v, want %d", lead.SupervisorID, junior.ID)
	}
}
