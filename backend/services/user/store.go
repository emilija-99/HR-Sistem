package user

import (
	"database/sql"
	"fmt"
	"log"
	"main/types"
	"main/utils"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	row := s.db.QueryRow(
		"SELECT id, first_name, last_name, email, password, created_at FROM users WHERE email=$1",
		email,
	)

	u := new(types.User)
	log.Print(u)
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
	)

	log.Print(err)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	log.Print(err)

	if err != nil {
		return nil, err
	}
	log.Print("RES")
	return u, nil
}

func scanRowIntoUser(rows *sql.Rows) (*types.User, error) {
	user := new(types.User)

	err := rows.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) GetUserByID(id int) (*types.User, error) {
	rows, err := s.db.Query("SELECT * from users where id = $1", id)
	if err != nil {
		return nil, err
	}

	u := new(types.User)

	for rows.Next() {
		u, err = scanRowIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}
func (s *Store) CreateUser(u types.User) (uint, error) {
	log.Print(u)
	query := `
        INSERT INTO users (first_name, last_name, email, password)
        VALUES ($1,$2,$3,$4)
        RETURNING id;
    `

	var id uint
	err := s.db.QueryRow(
		query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Password,
	).Scan(&id)
	log.Println(id, err)
	return id, err
}

func (s *Store) CreateUserWithRole(u types.User, roleName string) (*types.User, error) {
	log.Print("CREATEUSERWITHROLE", u, roleName)
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var userID uint
	err = tx.QueryRow(`
        INSERT INTO users (first_name, last_name, email, password)
        VALUES ($1,$2,$3,$4)
        RETURNING id
    `,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Password,
	).Scan(&userID)

	log.Print("userID: %d", userID)

	if err != nil {
		return nil, err
	}
	log.Printf("Error creaint user: %x", err)

	result, err := tx.Exec(`
        INSERT INTO user_roles (user_id, role_id)
        SELECT $1, id FROM roles WHERE name=$2
        ON CONFLICT DO NOTHING
    `, userID, roleName)

	fmt.Print("result: %+v", result)
	if err != nil {
		return nil, err
	}

	fmt.Printf("error inserting role: %+x", err)
	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return nil, fmt.Errorf("role does not exist")
	}

	fmt.Printf("rows: ", rows)
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	u.ID = userID
	return &u, nil
}

func (s *Store) EmailExists(email string) string {
	users, err := s.db.Exec("SELECT * FROM users where email = $1", email)
	if err != nil {
		return utils.ErrBadRequest
	}

	rows, err := users.RowsAffected()
	if err != nil {
		return utils.ErrBadRequest
	}

	if rows > 0 {
		return utils.ErrEmailAlreadyExists
	}

	return ""
}

func (s *Store) GetUserRole(userID uint) (string, error) {
	row := s.db.QueryRow(`
		SELECT r.name
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`, userID)

	var role string
	err := row.Scan(&role)
	return role, err
}

func (s *Store) AssignRole(userID uint, roleName string) error {
	query := `
        INSERT INTO user_roles (user_id, role_id)
        SELECT $1, id FROM roles WHERE name = $2
        ON CONFLICT DO NOTHING;
    `

	result, err := s.db.Exec(query, userID, roleName)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// critical check
	if rows == 0 {
		return fmt.Errorf("role '%s' does not exist", roleName)
	}

	return nil
}
