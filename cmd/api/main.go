package main

import (
	"context"
	"log/slog"

	"Drop-Key/internal/db"
)

func main() {
	ctx := context.Background()
	db, err := db.InitDB(ctx)
	if err != nil {
		slog.Error("Error initialising database.", "error", err)
	}
}
