package repository

import (
	"context"
	"fmt"
	"time"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users WHERE email = $1 LIMIT 1
	`
	var user model.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) ListManagers(ctx context.Context) ([]model.User, error) {
	query := `SELECT id, name, email, role FROM users WHERE role = 'admissions' ORDER BY id ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// ListAll returns all users in the system (for admin panel)
func (r *UserRepository) ListAll(ctx context.Context) ([]model.User, error) {
	query := `SELECT id, name, email, role, created_at, updated_at FROM users ORDER BY id ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

// UpdateUser modifies an existing user. If passwordHash is empty, password is left unchanged.
func (r *UserRepository) UpdateUser(ctx context.Context, id int, name, email, role, passwordHash string) (*model.User, error) {
	var (
		err error
		now = time.Now()
	)
	if passwordHash == "" {
		_, err = r.db.Pool.Exec(ctx, `UPDATE users SET name=$1, email=$2, role=$3, updated_at=$4 WHERE id=$5`,
			name, email, role, now, id)
	} else {
		_, err = r.db.Pool.Exec(ctx, `UPDATE users SET name=$1, email=$2, role=$3, password_hash=$4, updated_at=$5 WHERE id=$6`,
			name, email, role, passwordHash, now, id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	var u model.User
	err = r.db.Pool.QueryRow(ctx,
		`SELECT id, name, email, role, created_at, updated_at FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}
	return &u, nil
}

// DeleteUser removes a user by id. Leads assigned to that user keep assignee_id (FK is ON DELETE NO ACTION),
// so we null those out first to keep referential integrity.
func (r *UserRepository) DeleteUser(ctx context.Context, id int) error {
	_, _ = r.db.Pool.Exec(ctx, `UPDATE leads SET assignee_id = NULL WHERE assignee_id = $1`, id)
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// CreateUser inserts a new user and returns the created user
func (r *UserRepository) CreateUser(ctx context.Context, name, email, passwordHash, role string) (*model.User, error) {
	query := `
		INSERT INTO users (name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	var user model.User
	err := r.db.Pool.QueryRow(ctx, query, name, email, passwordHash, role, now, now).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	user.Name = name
	user.Email = email
	user.Role = role
	return &user, nil
}
