package paste

import (
	"context"
	"os"
	"testing"
	"time"

	"Drop-Key/internal/db"
	"Drop-Key/internal/models"
	"Drop-Key/internal/user"

	"github.com/stretchr/testify/assert"
)

func setupTestService(t *testing.T) (*pasteService, func()) {
	t.Helper()
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialize database")
	pasteRepo := NewPasteRepository(db)

	userRepo := user.NewUserRepositry(db)
	cleanup := func() {
		_, err := db.NewDropTable().Model(&models.Paste{}).IfExists().Exec(ctx)
		assert.NoError(t, err, "should drop Paste table")
		_, err = db.NewDropTable().Model(&models.User{}).IfExists().Exec(ctx)
		assert.NoError(t, err, "should drop User table")
		db.Close()
	}

	paste_service := NewPasteService(pasteRepo, userRepo)

	return paste_service, cleanup
}

func TestNewPasteService(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	assert.NotNil(t, service, "should not return a nil service.")
}

func TestCreate(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	err := service.userRepo.Create(ctx, &models.User{PublicKey: "PmUPAT+CAkeISk6GDWwhPW2d4mvpwPz/9AWaaOl30xs="})
	assert.NoError(t, err, "should create user without error.")

	NotEncryptedPaste := &models.Paste{
		Ciphertext: "encrypted-data",
		Signature:  "7rBbA8oBtMB411s/+zu0p60Z9iUi0I2JhzvCbGFQ7qxUOE/K8HjGzOAGHyjF/DMK+/PA6evav3xfLP4kWSCFBg==",
		PublicKey:  "PmUPAT+CAkeISk6GDWwhPW2d4mvpwPz/9AWaaOl30xs=",
	}
	id, err := service.Create(ctx, NotEncryptedPaste, 3600)
	assert.Error(t, err, "should retur error Ciphertext not base63 encrypted.")
	assert.Equal(t, "", id, "should return a nil paste id")

	validPaste := &models.Paste{
		Ciphertext: "VGhpcyBpcyBhIHNhbXBsZSBzZWNyZXQgbWVzc2FnZQ==",
		PublicKey:  "PmUPAT+CAkeISk6GDWwhPW2d4mvpwPz/9AWaaOl30xs=",
		Signature:  "7rBbA8oBtMB411s/+zu0p60Z9iUi0I2JhzvCbGFQ7qxUOE/K8HjGzOAGHyjF/DMK+/PA6evav3xfLP4kWSCFBg==",
	}
	id, err = service.Create(ctx, validPaste, 3600)
	assert.NoError(t, err, "should create a paste.")
	assert.NotNil(t, validPaste.ID, "should return a not nil id.")
	assert.Greater(t, validPaste.ExpiresAt, time.Now().UTC().Truncate(time.Second), "should return paste with future expiry date.")

	emptyCipherPaste := &models.Paste{
		Signature: "7rBbA8oBtMB411s/+zu0p60Z9iUi0I2JhzvCbGFQ7qxUOE/K8HjGzOAGHyjF/DMK+/PA6evav3xfLP4kWSCFBg==",
		PublicKey: "test-public-key",
	}
	id, err = service.Create(ctx, emptyCipherPaste, 3600)
	assert.Error(t, err, "should return error for paste with empty Ciphertext")
	assert.Equal(t, "", emptyCipherPaste.ID, "should return empty paste ID")

	expiredPaste := &models.Paste{
		Ciphertext: "VGhpcyBpcyBhIHNhbXBsZSBzZWNyZXQgbWVzc2FnZQ==",
		PublicKey:  "PmUPAT+CAkeISk6GDWwhPW2d4mvpwPz/9AWaaOl30xs=",
		Signature:  "signed-data",
	}
	id, err = service.Create(ctx, expiredPaste, -3600)
	assert.Error(t, err, "should return error for expired paste.")
	assert.Equal(t, "", expiredPaste.ID, "should return empty paste ID")
}
