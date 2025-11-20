package db

import (
	"context"
	"fmt"
	"os"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pgx connection pool using DATABASE_URL or a sensible default.
// กำหนดให้เรียกใช้ตอนเริ่มโปรแกรม เพื่อให้มี pool สำหรับใช้ทั้งระบบ
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// ค่า default สำหรับเครื่องของคุณ ปรับได้ตามต้องการ
		dsn = "postgres://in:in@localhost:5432/lindb"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}

	// Ping เพื่อเช็กว่าเชื่อมต่อได้จริง
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}

	return pool, nil
}
