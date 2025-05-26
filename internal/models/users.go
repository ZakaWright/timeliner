package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	IsActive     bool      `json:"is_active`
	CreatedAt    time.Time `json:"created_at"`
}

// Allows for interaction with the DB
type UserModel struct {
	DB  *pgxpool.Pool
	CTX context.Context
}

// interact with DB
func (m UserModel) GetByID(id int64) (*User, error) {
	query := `
		SELECT user_id, username, is_active
		FROM users 
		WHERE user_id=$1
		`
	var user User
	err := m.DB.QueryRow(m.CTX, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (m UserModel) Insert(username string, password string) (*User, error) {

	password_hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password %v", err)
	}
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING user_id, username
	`
	var user User
	err = m.DB.QueryRow(m.CTX, query, username, string(password_hash)).Scan(
		&user.ID,
		&user.PasswordHash,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") /*|| strings.Contains(err.Error(), "unique")*/ {
			return nil, errors.New("username must be unique")
		}
		return nil, fmt.Errorf("failed to create user %v", err)
	}
	return &user, nil
}

func (m UserModel) GetActiveUsers() ([]*User, error) {
	query := `
		SELECT user_id, username
		FROM users
		WHERE is_active=true
		ORDER BY username
	`

	rows, err := m.DB.Query(m.CTX, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username cannot be empty")
	}

	// TODO more
	return nil
}
