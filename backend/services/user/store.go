package user

import (
	"database/sql"
	"fmt"
	"log"
	types "main/types/user"
)

type RefreshTokenStore interface {
	SaveRefreshToken(userID uint, tokenHash string) error
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	log.Printf("Getting user by email: %s", email)
	u := new(types.User)
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, is_active, created_at, created_by, updated_at, updated_by
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.IsActive, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	log.Printf("Found user: %v", u)
	return u, nil
}

func scanRowIntoUser(rows *sql.Rows) (*types.User, error) {
	user := new(types.User)

	err := rows.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.IsActive,
		&user.CreatedAt,
		&user.CreatedBy,
		&user.UpdatedAt,
		&user.UpdatedBy,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
func (s *Store) GetUserByID(id int) (*types.User, error) {
	u := new(types.User)
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, is_active, created_at, created_by, updated_at, updated_by
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.IsActive, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) CreateUser(u types.User) (uint, error) {
	// log.Print(u)
	query := `
        INSERT INTO users (email, password_hash)
        VALUES ($1,$2)
        RETURNING id;
    `

	var id uint
	err := s.db.QueryRow(
		query,
		u.Email,
		u.Password,
	).Scan(&id)
	return id, err
}

func (s *Store) CreateUserWithRole(u types.User, roleName string, createdBy *uint) (*types.User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var userID uint
	err = tx.QueryRow(
		`INSERT INTO users (email, password_hash, created_by) VALUES ($1, $2, $3) RETURNING id`,
		u.Email, u.Password, createdBy,
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	result, err := tx.Exec(
		`INSERT INTO user_roles (user_id, role_id)
		 SELECT $1, id FROM roles WHERE name = $2
		 ON CONFLICT DO NOTHING`, userID, roleName,
	)
	if err != nil {
		return nil, fmt.Errorf("assign role: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("role '%s' does not exist", roleName)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	u.ID = userID
	return &u, nil
}

func (s *Store) EmailExists(email string) bool {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
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

func (s *Store) SaveRefreshToken(userID uint, tokenHash string) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES (gen_random_uuid(), $1, $2, NOW() + INTERVAL '7 days');
	`
	_, err := s.db.Exec(query, userID, tokenHash)
	return err
}

func (s *Store) GetUserIDByRefreshToken(tokenHash string) (uint, error) {
	var userID uint

	query := `
		SELECT user_id
		FROM refresh_tokens
		WHERE token_hash = $1
		  AND revoked = false
		  AND expires_at > NOW()
		LIMIT 1;
	`

	err := s.db.QueryRow(query, tokenHash).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *Store) GetUserPremissions(roleName types.PermissionRequest) (*types.UserPermissions, error) {
	var perms []types.Permission

	query := `SELECT DISTINCT p.id, p.code
				FROM permissions p
				WHERE p.id IN (
				    SELECT rp.permission_id
				    FROM role_permissions rp
				    WHERE rp.role_id = (
				        SELECT r.id
				        FROM roles r
				        WHERE r.name = $1
				    )
				)
				ORDER BY p.id;`

	rows, err := s.db.Query(query, roleName.RoleName)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var p types.Permission
		if err := rows.Scan(&p.ID, &p.Code); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &types.UserPermissions{
		Permissions: perms,
	}, nil
}

func (s *Store) ChangeUserStatus(userID uint, isActive bool) (*types.User, error) {
	u := new(types.User)
	err := s.db.QueryRow(
		`UPDATE users SET is_active = $1 WHERE id = $2
		 RETURNING id, email, password_hash, is_active, created_at, created_by, updated_at, updated_by`,
		isActive, userID,
	).Scan(&u.ID, &u.Email, &u.Password, &u.IsActive, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}
	return u, nil
}

func (s *Store) GetAllUsers() ([]types.User, error) {
	var users []types.User

	rows, err := s.db.Query(
		`SELECT id, email, password_hash, is_active, created_at, created_by, updated_at, updated_by
			 FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user, err := scanRowIntoUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
func (s *Store) GetUserByIDWithRole(id int64) (*types.User, string, error) {
	u := new(types.User)
	var roleName sql.NullString
	err := s.db.QueryRow(
		`SELECT u.id, u.email, u.password_hash, u.is_active, u.created_at, u.created_by, u.updated_at, u.updated_by, r.name
		 FROM users u
		 LEFT JOIN user_roles ur ON ur.user_id = u.id
		 LEFT JOIN roles r ON r.id = ur.role_id
		 WHERE u.id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.IsActive, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy, &roleName)

	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	return u, roleName.String, nil
}
