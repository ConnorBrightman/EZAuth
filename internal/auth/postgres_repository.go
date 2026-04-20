package auth

import (
	"database/sql"
	"errors"
	"strings"

	_ "github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(dsn string) (*PostgresUserRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	repo := &PostgresUserRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *PostgresUserRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresUserRepository) migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS ezauth_users (
			id             TEXT PRIMARY KEY,
			email          TEXT UNIQUE NOT NULL,
			password       TEXT NOT NULL,
			refresh_token  TEXT NOT NULL DEFAULT '',
			refresh_expiry BIGINT NOT NULL DEFAULT 0
		)
	`)
	return err
}

func (r *PostgresUserRepository) Create(user User) error {
	_, err := r.db.Exec(
		`INSERT INTO ezauth_users (id, email, password, refresh_token, refresh_expiry)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.Password, user.RefreshToken, user.RefreshExpiry,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) FindByEmail(email string) (User, error) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, email, password, refresh_token, refresh_expiry
		 FROM ezauth_users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.RefreshToken, &user.RefreshExpiry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *PostgresUserRepository) Update(user User) error {
	result, err := r.db.Exec(
		`UPDATE ezauth_users SET password = $1, refresh_token = $2, refresh_expiry = $3
		 WHERE email = $4`,
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
