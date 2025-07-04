package user

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"testing"

	"Drop-Key/internal/db"
	"Drop-Key/internal/models"
	"Drop-Key/internal/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var publicKey string

func init() {
	_, pub, _ := ed25519.GenerateKey(nil)
	publicKey = base64.StdEncoding.EncodeToString(pub)
}

func setupTestService(t *testing.T) (*userService, context.Context, func()) {
	t.Helper()

	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialise db without error")

	userRepo := NewUserRepository(db)
	userService := NewUserService(userRepo)
	cleanup := func() {
		_, err = db.NewDropTable().Model(&models.User{}).IfExists().Exec(ctx)
		assert.NoError(t, err, "should drop User table")
		db.Close()
	}

	return userService, ctx, cleanup
}

func TestNewUserService(t *testing.T) {
	userService, _, cleanup := setupTestService(t)
	defer cleanup()

	assert.NotNil(t, userService, "should return a not nil userService")
}

func TestCreateUser(t *testing.T) {
	userService, ctx, cleanup := setupTestService(t)
	defer cleanup()

	t.Run("valid user", func(t *testing.T) {
		pub, _, err := ed25519.GenerateKey(nil) // Use public key (first return value)
		assert.NoError(t, err, "should generate ed25519 key without error")
		validPublicKey := base64.StdEncoding.EncodeToString(pub)

		user := &models.User{
			PublicKey: validPublicKey,
		}
		id, err := userService.CreateUser(ctx, user)
		assert.NoError(t, err, "should create user without error")
		assert.NotEmpty(t, id, "should return a non-empty user ID")
		_, err = uuid.Parse(id)
		assert.NoError(t, err, "should return a valid UUID")

		fetched, err := userService.repo.GetByID(ctx, id)
		assert.NoError(t, err, "should fetch user from database")
		assert.NotNil(t, fetched, "should return a user")
		if fetched != nil {
			assert.Equal(t, validPublicKey, fetched.PublicKey, "public key should match")
			assert.Equal(t, id, fetched.ID, "user ID should match")
		}
	})

	t.Run("empty publicKey", func(t *testing.T) {
		user := &models.User{
			PublicKey: "",
		}
		id, err := userService.CreateUser(ctx, user)
		assert.ErrorIs(t, err, utils.ErrEmptyPublicKey, "should return error")
		assert.Empty(t, id, "should return empty id")
	})

	t.Run("invalid publicKey", func(t *testing.T) {
		user := &models.User{
			PublicKey: "invalid-public-key",
		}
		id, err := userService.CreateUser(ctx, user)
		assert.ErrorIs(t, err, utils.ErrInvalidPublicKey, "should retur error")
		assert.Empty(t, id, "should return empty id")
	})
}

