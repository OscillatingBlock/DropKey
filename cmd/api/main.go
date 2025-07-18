package main

import (
	"context"
	"log/slog"

	"Drop-Key/internal/db"
	"Drop-Key/internal/paste"
	"Drop-Key/internal/router"
	"Drop-Key/internal/user"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("ERROR Failed to get DSN from .env file.")
	}
	ctx := context.Background()
	db, err := db.InitDB(ctx)
	if err != nil {
		slog.Error("Error initialising database.", "error", err)
	}
	pasteRepo := paste.NewPasteRepository(db)
	userRepo := user.NewUserRepository(db)
	pasteService := paste.NewPasteService(pasteRepo, userRepo)
	userService := user.NewUserService(userRepo)

	pasteHandler := paste.NewPasteHandler(pasteService)
	userHandler := user.NewUserHandler(userService)

	e := router.Router(pasteHandler, userHandler)

	e.Logger.Fatal(e.Start("127.0.0.1:8081"))
}
