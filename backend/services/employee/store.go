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

// scanFullRow scans a row from the JOIN query into Employee (17 columns)
func scanFullRow(scanner interface {
	Scan(dest ...any) error
}) (*types.Employee, error) {
	emp := new(types.Employee)
	err := scanner.Scan(
		&emp.ID, &emp.UserID, &emp.FirstName, &emp.LastName,
		&emp.PhoneNumber, &emp.PrivateEmail, &emp.Street, &emp.Country, &emp.City,
		&emp.DateOfBirth, &emp.HireDate, &emp.PositionID, &emp.CreatedAt,
		&emp.PositionTitle, &emp.PositionLevel, &emp.DepartmentID, &emp.DepartmentName,
	)
	if err != nil {
		return nil, err
	}
	return emp, nil
}

func (s *Store) Create(e types.Employee) (*types.Employee, error) {
	var emp types.Employee
	err := s.db.QueryRow(
		`INSERT INTO employees
		 (user_id, first_name, last_name, phone_number, private_email, street, country, city, date_of_birth, hire_date, position_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
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
			position_id   = COALESCE($10, position_id)
		 WHERE id = $11`,
		p.FirstName, p.LastName, p.PhoneNumber, p.PrivateEmail, p.Street,
		p.Country, p.City, p.DateOfBirth, p.HireDate, p.PositionID, id,
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
