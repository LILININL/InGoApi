package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository กำหนดพฤติกรรมที่ layer อื่น (เช่น service) เรียกใช้ข้อมูลผู้ใช้
type Repository interface {
	Create(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	UpdatePassword(ctx context.Context, userID int, newHash string) error
	List(ctx context.Context) ([]User, error)
}

// repo เป็น implementation ที่ใช้ pgxpool
type repo struct {
	pool *pgxpool.Pool
}

// NewRepository คืนค่า repository ที่พร้อมใช้งานกับฐานข้อมูล
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repo{pool: pool}
}

func (r *repo) Create(ctx context.Context, u User) error {
	const query = `
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, u.Email, u.PasswordHash, u.Name)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *repo) FindByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT id, email, password_hash, name, created_at
		FROM users
		WHERE email = $1
	`

	row := r.pool.QueryRow(ctx, query, email)

	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("user not found: %w", err)
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}

func (r *repo) UpdatePassword(ctx context.Context, userID int, newHash string) error {
	const query = `
		UPDATE users
		SET password_hash = $1
		WHERE id = $2
	`

	if _, err := r.pool.Exec(ctx, query, newHash, userID); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (r *repo) List(ctx context.Context) ([]User, error) {
	const query = `
		SELECT id, email, password_hash, name, created_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate users: %w", rows.Err())
	}

	return users, nil
}
