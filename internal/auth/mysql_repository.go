package auth

import (
	"database/sql"
	"errors"
	"strings"

	_ "github.com/go-sql-driver/mysql" // Make sure to: go get github.com/go-sql-driver/mysql
)

type MySQLUserRepository struct {
	db *sql.DB
}

func NewMySQLUserRepository(dsn string) (*MySQLUserRepository, error) {
	// MySQL DSNs often look like: user:pass@tcp(127.0.0.1:3306)/dbname
	// If your unified string has "mysql://", you may need to trim it before sql.Open
	cleanDSN := strings.TrimPrefix(dsn, "mysql://")

	db, err := sql.Open("mysql", cleanDSN)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	repo := &MySQLUserRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *MySQLUserRepository) Close() error {
	return r.db.Close()
}

func (r *MySQLUserRepository) migrate() error {
	// MySQL requires a length for VARCHAR primary keys
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS ezauth_users (
			id VARCHAR(255) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password TEXT NOT NULL,
			refresh_token TEXT,
			refresh_expiry BIGINT NOT NULL DEFAULT 0
		)
	`)
	return err
}

func (r *MySQLUserRepository) Create(user User) error {
	// MySQL uses ? instead of $1
	_, err := r.db.Exec(
		`INSERT INTO ezauth_users (id, email, password, refresh_token, refresh_expiry)
		 VALUES (?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.Password, user.RefreshToken, user.RefreshExpiry,
	)
	if err != nil {
		// MySQL error for duplicate entry is usually error 1062
		if strings.Contains(err.Error(), "Duplicate entry") {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (r *MySQLUserRepository) FindByEmail(email string) (User, error) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, email, password, refresh_token, refresh_expiry
		 FROM ezauth_users WHERE email = ?`, email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.RefreshToken, &user.RefreshExpiry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *MySQLUserRepository) Update(user User) error {
	result, err := r.db.Exec(
		`UPDATE ezauth_users SET password = ?, refresh_token = ?, refresh_expiry = ?
		 WHERE email = ?`,
		user.Password, user.RefreshToken, user.RefreshExpiry, user.Email,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}
