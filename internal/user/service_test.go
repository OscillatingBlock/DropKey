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
		pub, _, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err, "should generate ed25519 key without error")
		validPublicKey := base64.StdEncoding.EncodeToString(pub)

		user := &models.User{
			PublicKey: validPublicKey,
		}
		id, err := userService.Create(ctx, user)
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
		id, err := userService.Create(ctx, user)
		assert.ErrorIs(t, err, utils.ErrEmptyPublicKey, "should return error")
		assert.Empty(t, id, "should return empty id")
	})

	t.Run("invalid publicKey", func(t *testing.T) {
		user := &models.User{
			PublicKey: "invalid-public-key",
		}
		id, err := userService.Create(ctx, user)
		assert.ErrorIs(t, err, utils.ErrInvalidPublicKey, "should retur error")
		assert.Empty(t, id, "should return empty id")
	})
}

func TestGetByID(t *testing.T) {
	userService, ctx, cleanup := setupTestService(t)
	defer cleanup()

	t.Run("valid user id", func(t *testing.T) {
		pub, _, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err, "should generate ed25519 key without error")
		validPublicKey := base64.StdEncoding.EncodeToString(pub)

		user := &models.User{
			PublicKey: validPublicKey,
		}
		id, err := userService.Create(ctx, user)
		assert.NoError(t, err, "should create user without error")

		fetchedUser, err := userService.GetByID(ctx, id)
		assert.NoError(t, err, "shhould return user")
		assert.Equal(t, user.ID, fetchedUser.ID, "ID should match")
		assert.Equal(t, user.PublicKey, fetchedUser.PublicKey, "publicKey should match")
	})

	t.Run("invalid user id", func(t *testing.T) {
		fetchedUser, err := userService.GetByID(ctx, "invalid-user-id")
		assert.ErrorIs(t, err, utils.ErrInvalidUserID, "shhould return error")
		assert.Nil(t, fetchedUser, "should return nil user")
	})

	t.Run("empty user id", func(t *testing.T) {
		fetchedUser, err := userService.GetByID(ctx, "")
		assert.ErrorIs(t, err, utils.ErrEmptyUserID, "shhould return error")
		assert.Nil(t, fetchedUser, "should return nil user")
	})
}

func TestAuthenticate(t *testing.T) {
	userService, ctx, cleanup := setupTestService(t)
	defer cleanup()

	t.Run("valid user id", func(t *testing.T) {
		pub, priv, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err, "should generate ed25519 key without error")
		validPublicKey := base64.StdEncoding.EncodeToString(pub)
		msg := []byte("a hard challenge")
		signature := ed25519.Sign(priv, msg)
		sig := base64.StdEncoding.EncodeToString(signature)

		user := &models.User{
			PublicKey: validPublicKey,
		}
		id, err := userService.Create(ctx, user)
		assert.NoError(t, err, "should create user without error")

		msgB64 := base64.StdEncoding.EncodeToString(msg)
		ok, err := userService.Authenticate(ctx, id, sig, msgB64)
		assert.NoError(t, err, "should Authenticate without error")
		assert.Equal(t, true, ok, "authentication should be true")
	})

	t.Run("invalid user id", func(t *testing.T) {
		_, priv, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err, "should generate key")

		challenge := []byte("a hard challenge")
		signature := ed25519.Sign(priv, challenge)
		sig := base64.StdEncoding.EncodeToString(signature)

		fakeUserID := "non-existent-user-id"

		ok, err := userService.Authenticate(ctx, fakeUserID, sig, string(challenge))
		assert.ErrorIs(t, err, utils.ErrUserNotFound, "should return user not found error")
		assert.False(t, ok, "authentication should fail")
	})

	t.Run("empty user id", func(t *testing.T) {
		_, priv, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err, "should generate key")

		challenge := []byte("a hard challenge")
		signature := ed25519.Sign(priv, challenge)
		sig := base64.StdEncoding.EncodeToString(signature)

		ok, err := userService.Authenticate(ctx, "", sig, string(challenge))
		assert.ErrorIs(t, err, utils.ErrEmptyUserID, "should return empty user id error")
		assert.False(t, ok, "authentication should fail")
	})
}

func TestGetByPublicKey(t *testing.T) {
	userService, ctx, cleanup := setupTestService(t)
	defer cleanup()

	t.Run("valid public key", func(t *testing.T) {
		pub, _, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err, "keygen should not fail")
		pubB64 := base64.StdEncoding.EncodeToString(pub)

		user := &models.User{PublicKey: pubB64}
		_, err = userService.Create(ctx, user)
		assert.NoError(t, err, "should create user")

		fetched, err := userService.GetByPublicKey(ctx, pubB64)
		assert.NoError(t, err, "should fetch user by public key")
		assert.Equal(t, user.PublicKey, fetched.PublicKey, "public key should match")
	})

	t.Run("empty public key", func(t *testing.T) {
		user, err := userService.GetByPublicKey(ctx, "")
		assert.ErrorIs(t, err, utils.ErrEmptyPublicKey)
		assert.Nil(t, user)
	})

	t.Run("invalid base64 public key", func(t *testing.T) {
		user, err := userService.GetByPublicKey(ctx, "not-base64!")
		assert.ErrorIs(t, err, utils.ErrInvalidPublicKey)
		assert.Nil(t, user)
	})

	t.Run("non-existent user", func(t *testing.T) {
		pub, _, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err)
		pubB64 := base64.StdEncoding.EncodeToString(pub)

		user, err := userService.GetByPublicKey(ctx, pubB64)
		assert.ErrorIs(t, err, utils.ErrUserNotFound)
		assert.Nil(t, user)
	})
}
