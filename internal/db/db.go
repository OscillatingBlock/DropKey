package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"Drop-Key/internal/paste"
	"Drop-Key/internal/user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

func InitDB(ctx context.Context) (*bun.DB, error) {
	// Setup Connection to database
	DSN := os.Getenv("DSN")
	if DSN == "" {
		slog.Error("Failed to get DSN from .env file.")
		return nil, fmt.Errorf("Failed to get DSN from .env file")
	}
	sqldb, err := sql.Open("mysql", DSN)
	if err != nil {
		slog.Error("Failed to open MySQL connection: ", "error", err)
		return nil, fmt.Errorf("Error while opening MySQL connection, error: %w", err)
	}

	// ping database for verification
	err = sqldb.PingContext(ctx)
	if err != nil {
		slog.Error("Connection to database failed.", "error", err)
		return nil, fmt.Errorf("Error while connecting to database, error: %w", err)
	}

	// setting connection pool
	var db *bun.DB
	db = bun.NewDB(sqldb, mysqldialect.New())
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)

	// Initialise Database tables
	var paste paste.Paste
	_, err = db.NewCreateTable().Model(&paste).IfNotExists().Exec(ctx)
	if err != nil {
		slog.Error("Error while creating Paste table", "table", "paste", "error", err)
		return nil, fmt.Errorf("Error while creating paste table, error %w", err)
	}

	var user user.User
	_, err = db.NewCreateTable().Model(&user).IfNotExists().Exec(ctx)
	if err != nil {
		slog.Error("Error while creating User table", "table", "user", "error", err)
		return nil, fmt.Errorf("Error while creating users table, error %w", err)
	}

	return db, nil
}
