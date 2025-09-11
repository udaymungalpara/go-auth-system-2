package storage

import (
	"database/sql"
	"go-auth-system/src/models"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	*sql.DB
}

func NewPostgresDB(dataSourceName string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{db}, nil
}

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

// Change signature to match interface
func (db *PostgresDB) CreateUser(user *models.User) error {
	query := `INSERT INTO users (email, password_hash, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id`
	return db.QueryRow(query, user.Email, user.PasswordHash, user.FirstName, user.LastName).Scan(&user.ID)
}

// Add this method to implement Storage interface
func (db *PostgresDB) GetUserByID(id string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, first_name, last_name, is_email_verified, email_verified_at, failed_login_count, locked_until, created_at, updated_at, last_login_at FROM users WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.IsEmailVerified, &user.EmailVerifiedAt, &user.FailedLoginCount, &user.LockedUntil, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (db *PostgresDB) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, first_name, last_name, is_email_verified, email_verified_at, failed_login_count, locked_until, created_at, updated_at, last_login_at FROM users WHERE email = $1`
	err := db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.IsEmailVerified, &user.EmailVerifiedAt, &user.FailedLoginCount, &user.LockedUntil, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Change signature to match interface
func (db *PostgresDB) UpdateUser(user *models.User) error {
	query := `UPDATE users SET email = $1, password_hash = $2, first_name = $3, last_name = $4, is_email_verified = $5, email_verified_at = $6, failed_login_count = $7, locked_until = $8, updated_at = CURRENT_TIMESTAMP, last_login_at = $9 WHERE id = $10`
	_, err := db.Exec(query, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.IsEmailVerified, user.EmailVerifiedAt, user.FailedLoginCount, user.LockedUntil, user.LastLoginAt, user.ID)
	return err
}

func (db *PostgresDB) UpdateUserPassword(userID int, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err := db.Exec(query, passwordHash, userID)
	return err
}
