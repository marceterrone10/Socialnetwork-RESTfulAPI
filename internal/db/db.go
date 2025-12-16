package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func New(addr string, maxOpenConns int, maxIdleConns int, maxLifetime string) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	duration, err := time.ParseDuration(maxLifetime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Intentar conectar con retry
	var lastErr error
	for i := 0; i < 3; i++ {
		if err := db.PingContext(ctx); err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		return db, nil
	}

	return nil, lastErr
}
