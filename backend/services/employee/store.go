package employee

import (
	"database/sql"
	"fmt"
	types "main/types/employee"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// scanFullRow scans a row from the JOIN query into Employee (20 columns)
func scanFullRow(scanner interface {
	Scan(dest ...any) error
}) (*types.Employee, error) {
	emp := new(types.Employee)
	var supFirst, supLast *string
	// DATE columns: scan as time and normalize to YYYY-MM-DD for clean display
	// (and so <input type="date"> round-trips correctly).
	var dob, hire sql.NullTime
	err := scanner.Scan(
		&emp.ID, &emp.UserID, &emp.FirstName, &emp.LastName,
		&emp.PhoneNumber, &emp.PrivateEmail, &emp.Street, &emp.Country, &emp.City,
		&dob, &hire, &emp.PositionID, &emp.CreatedAt,
		&emp.PositionTitle, &emp.PositionLevel, &emp.DepartmentID, &emp.DepartmentName,
		&emp.SupervisorID, &supFirst, &supLast,
	)
	if err != nil {
		return nil, err
	}
	if dob.Valid {
		s := dob.Time.Format("2006-01-02")
		emp.DateOfBirth = &s
	}
	if hire.Valid {
		s := hire.Time.Format("2006-01-02")
		emp.HireDate = &s
	}
	if supFirst != nil {
		name := *supFirst
		if supLast != nil {
			name += " " + *supLast
		}
		emp.SupervisorName = &name
	}
	return emp, nil
}

func (s *Store) Create(e types.Employee) (*types.Employee, error) {
	var emp types.Employee
	err := s.db.QueryRow(
		`INSERT INTO employees
		 (user_id, first_name, last_name, phone_number, private_email, street, country, city, date_of_birth, hire_date, position_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,COALESCE($10, CURRENT_DATE),$11)
		 RETURNING id, user_id, first_name, last_name, phone_number, private_email, street, country, city, date_of_birth, hire_date, position_id, created_at`,
		e.UserID, e.FirstName, e.LastName, e.PhoneNumber, e.PrivateEmail,
		e.Street, e.Country, e.City, e.DateOfBirth, e.HireDate, e.PositionID,
	).Scan(
		&emp.ID, &emp.UserID, &emp.FirstName, &emp.LastName,
		&emp.PhoneNumber, &emp.PrivateEmail, &emp.Street, &emp.Country, &emp.City,
		&emp.DateOfBirth, &emp.HireDate, &emp.PositionID, &emp.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create employee: %w", err)
	}

	// Re-fetch with JOIN to populate position/department fields
	return s.GetByID(int64(emp.ID))
}

// CreateWithUser creates a user account (with a role) and the matching employee
// profile in a single transaction. Used by the HR "new employee" flow.
func (s *Store) CreateWithUser(email, passwordHash, roleName string, createdBy *uint, e types.Employee) (*types.Employee, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var userID uint
	err = tx.QueryRow(
		`INSERT INTO users (email, password_hash, created_by) VALUES ($1, $2, $3) RETURNING id`,
		email, passwordHash, createdBy,
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	res, err := tx.Exec(
		`INSERT INTO user_roles (user_id, role_id)
		 SELECT $1, id FROM roles WHERE name = $2
		 ON CONFLICT DO NOTHING`, userID, roleName)
	if err != nil {
		return nil, fmt.Errorf("assign role: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, fmt.Errorf("role '%s' does not exist", roleName)
	}

	var empID uint
	err = tx.QueryRow(
		`INSERT INTO employees
		 (user_id, first_name, last_name, phone_number, private_email, street, country, city, date_of_birth, hire_date, position_id, supervisor_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,COALESCE($10, CURRENT_DATE),$11,$12)
		 RETURNING id`,
		userID, e.FirstName, e.LastName, e.PhoneNumber, e.PrivateEmail,
		e.Street, e.Country, e.City, e.DateOfBirth, e.HireDate, e.PositionID, e.SupervisorID,
	).Scan(&empID)
	if err != nil {
		return nil, fmt.Errorf("create employee: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetByID(int64(empID))
}

func (s *Store) GetByID(id int64) (*types.Employee, error) {
	emp, err := scanFullRow(s.db.QueryRow(
		types.JoinsQuery+` WHERE e.id = $1`, id,
	))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("employee not found")
	}
	if err != nil {
		return nil, err
	}
	return emp, nil
}

func (s *Store) GetByUserID(userID uint) (*types.Employee, error) {
	emp, err := scanFullRow(s.db.QueryRow(
		types.JoinsQuery+` WHERE e.user_id = $1`, userID,
	))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("employee not found")
	}
	if err != nil {
		return nil, err
	}
	return emp, nil
}

func (s *Store) GetAll() ([]types.Employee, error) {
	rows, err := s.db.Query(types.JoinsQuery + ` ORDER BY e.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []types.Employee
	for rows.Next() {
		emp, err := scanFullRow(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, *emp)
	}
	return employees, rows.Err()
}

func (s *Store) Update(id int64, p types.UpdateEmployeePayload) (*types.Employee, error) {
	result, err := s.db.Exec(
		`UPDATE employees SET
			first_name    = COALESCE($1,  first_name),
			last_name     = COALESCE($2,  last_name),
			phone_number  = COALESCE($3,  phone_number),
			private_email = COALESCE($4,  private_email),
			street        = COALESCE($5,  street),
			country       = COALESCE($6,  country),
			city          = COALESCE($7,  city),
			date_of_birth = COALESCE($8,  date_of_birth),
			hire_date     = COALESCE($9,  hire_date),
			position_id   = COALESCE($10, position_id),
			supervisor_id = COALESCE($11, supervisor_id)
		 WHERE id = $12`,
		p.FirstName, p.LastName, p.PhoneNumber, p.PrivateEmail, p.Street,
		p.Country, p.City, p.DateOfBirth, p.HireDate, p.PositionID, p.SupervisorID, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update employee: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("employee not found")
	}

	// Re-fetch with JOIN to return full data
	return s.GetByID(id)
}

// ── Reference data ────────────────────────────────────────────

func (s *Store) GetCountries() ([]types.Country, error) {
	rows, err := s.db.Query(`SELECT country_id, country_name, iso FROM countries ORDER BY country_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []types.Country
	for rows.Next() {
		var c types.Country
		if err := rows.Scan(&c.CountryID, &c.CountryName, &c.Iso); err != nil {
			return nil, err
		}
		countries = append(countries, c)
	}
	return countries, rows.Err()
}

func (s *Store) GetPositions() ([]types.Position, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.department_id, p.title, p.level, p.description, d.name
		FROM positions p
		LEFT JOIN departments d ON d.id = p.department_id
		ORDER BY p.department_id, p.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []types.Position
	for rows.Next() {
		var p types.Position
		if err := rows.Scan(&p.ID, &p.DepartmentID, &p.Title, &p.Level, &p.Description, &p.DepartmentName); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, rows.Err()
}

func (s *Store) GetDepartments() ([]types.Department, error) {
	rows, err := s.db.Query(`SELECT id, name, description, status FROM departments ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []types.Department
	for rows.Next() {
		var d types.Department
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.Status); err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}
	return departments, rows.Err()
}
