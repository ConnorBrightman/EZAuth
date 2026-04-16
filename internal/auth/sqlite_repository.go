package auth

import (
	"database/sql"
	"errors"
	"strings"

	_ "modernc.org/sqlite"
)

type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository opens (or creates) a SQLite database at path and
// runs the schema migration.
func NewSQLiteUserRepository(path string) (*SQLiteUserRepository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &SQLiteUserRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteUserRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteUserRepository) migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id             TEXT PRIMARY KEY,
			email          TEXT UNIQUE NOT NULL,
			password       TEXT NOT NULL,
			refresh_token  TEXT NOT NULL DEFAULT '',
			refresh_expiry INTEGER NOT NULL DEFAULT 0
		)
	`)
	return err
}

func (r *SQLiteUserRepository) Create(user User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password, refresh_token, refresh_expiry)
		 VALUES (?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.Password, user.RefreshToken, user.RefreshExpiry,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (r *SQLiteUserRepository) FindByEmail(email string) (User, error) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, email, password, refresh_token, refresh_expiry
		 FROM users WHERE email = ?`, email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.RefreshToken, &user.RefreshExpiry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *SQLiteUserRepository) Update(user User) error {
	result, err := r.db.Exec(
		`UPDATE users SET password = ?, refresh_token = ?, refresh_expiry = ?
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
