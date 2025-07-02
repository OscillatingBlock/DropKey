package paste

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"Drop-Key/internal/db"
	"Drop-Key/internal/models"
	"Drop-Key/internal/user"
	"Drop-Key/internal/utils"

	"github.com/stretchr/testify/assert"
)

var (
	publicKey   string = "PmUPAT+CAkeISk6GDWwhPW2d4mvpwPz/9AWaaOl30xs="
	signature   string = "7rBbA8oBtMB411s/+zu0p60Z9iUi0I2JhzvCbGFQ7qxUOE/K8HjGzOAGHyjF/DMK+/PA6evav3xfLP4kWSCFBg=="
	cipher_text string = "VGhpcyBpcyBhIHNhbXBsZSBzZWNyZXQgbWVzc2FnZQ=="
)

func setupTestService(t *testing.T) (*pasteService, func()) {
	t.Helper()
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialize database")
	pasteRepo := NewPasteRepository(db)

	userRepo := user.NewUserRepository(db)
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
	err := service.userRepo.Create(ctx, &models.User{PublicKey: publicKey})
	assert.NoError(t, err, "should create user without error.")

	t.Run("invalid cyphertext not base 64", func(t *testing.T) {
		NotEncryptedPaste := &models.Paste{
			Ciphertext: "encrypted-data",
			Signature:  signature,
			PublicKey:  publicKey,
		}
		id, err := service.Create(ctx, NotEncryptedPaste, 3600)
		assert.ErrorIs(t, err, utils.ErrPasteInvalidCiphertext, "should retur error Ciphertext not base63 encrypted.")
		assert.Equal(t, "", id, "should return a nil paste id")
	})

	t.Run("valid paste", func(t *testing.T) {
		validPaste := &models.Paste{
			Ciphertext: cipher_text,
			PublicKey:  publicKey,
			Signature:  signature,
		}
		id, err := service.Create(ctx, validPaste, 3600)
		assert.NoError(t, err, "should create a paste.")
		assert.NotEmpty(t, id, "should return a non-empty ID")
		assert.NotEmpty(t, validPaste.ID, "should set paste ID")
		assert.WithinDuration(t, time.Now().UTC().Add(time.Hour), validPaste.ExpiresAt, time.Second, "expires_at should be in the future")
	})

	t.Run("empty Ciphertext", func(t *testing.T) {
		emptyCipherPaste := &models.Paste{
			Signature: signature,
			PublicKey: "test-public-key",
		}
		id, err := service.Create(ctx, emptyCipherPaste, 3600)
		assert.ErrorIs(t, err, utils.ErrPasteEmptyCiphertext, "should return ErrPasteEmptyCiphertext")
		assert.Empty(t, id, "should return empty ID")
	})

	t.Run("expired paste", func(t *testing.T) {
		paste := &models.Paste{
			Ciphertext: cipher_text,
			PublicKey:  publicKey,
			Signature:  signature,
		}
		id, err := service.Create(ctx, paste, -3600)
		assert.ErrorIs(t, err, utils.ErrPasteExpiredAlready, "should return ErrPasteExpiredAlready")
		assert.Empty(t, id, "should return empty ID")
	})
}

func TestGetByID(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	err := service.userRepo.Create(ctx, &models.User{PublicKey: publicKey})
	assert.NoError(t, err, "should create user without error.")

	paste := &models.Paste{
		Ciphertext: cipher_text,
		PublicKey:  publicKey,
		Signature:  signature,
	}
	id, err := service.Create(ctx, paste, 3600)
	assert.NoError(t, err, "should create a paste.")
	assert.NotEmpty(t, id, "should return a non-empty ID")
	assert.NotEmpty(t, paste, "should set paste ID")
	assert.WithinDuration(t, time.Now().UTC().Add(time.Hour), paste.ExpiresAt, time.Second, "expires_at should be in the future")

	t.Run("valid ID", func(t *testing.T) {
		fetched, err := service.GetByID(ctx, id)
		assert.NoError(t, err, "should retrieve paste without error")
		assert.NotNil(t, fetched, "should return a non-nil paste")
		assert.Equal(t, id, fetched.ID, "ID should match")
		assert.Equal(t, paste.Ciphertext, fetched.Ciphertext, "ciphertext should match")
	})

	t.Run("Invalid ID", func(t *testing.T) {
		fetched, err := service.GetByID(ctx, "invlid-id")
		assert.ErrorIs(t, utils.ErrPasteInvalidID, err, "should return error for invalid id")
		assert.Nil(t, fetched, "should return nil paste.")
	})

	t.Run("non-existent ID", func(t *testing.T) {
		_, err := service.GetByID(ctx, "123e4567-e89b-12d3-a456-426614174000")
		assert.ErrorIs(t, err, utils.ErrPasteNotFound, "should return ErrPasteNotFound")
	})

	t.Run("expired paste", func(t *testing.T) {
		expiredPaste := &models.Paste{
			Ciphertext: "VGhpcyBpcyBhIHNhbXBsZSBzZWNyZXQgbWVzc2FnZQ==",
			PublicKey:  publicKey,
			Signature:  "7rBbA8oBtMB411s/+zu0p60Z9iUi0I2JhzvCbGFQ7qxUOE/K8HjGzOAGHyjF/DMK+/PA6evav3xfLP4kWSCFBg==",
		}
		id, err := service.Create(ctx, expiredPaste, -3600)
		assert.ErrorIs(t, err, utils.ErrPasteExpiredAlready, "should return ErrPasteExpiredAlready")
		assert.Empty(t, id, "should return empty ID")
	})
}

func TestGetbyPublicKey(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()
	ctx := context.Background()

	publicKey := "kwf9K6p0OPrF5wP9l1t8ZIEd52/RCLy6XhN8ZMKGT0A="
	err := service.userRepo.Create(ctx, &models.User{PublicKey: publicKey})
	assert.NoError(t, err, "should create user without error")

	paste1 := &models.Paste{
		Ciphertext: "VGhpcyBpcyBhIHNhbXBsZSBzZWNyZXQgbWVzc2FnZQ==",
		PublicKey:  publicKey,
		Signature:  "u2INPxBa85E9DJUt4UbaG8KBbeNsXoeS6bxGj+zfmvlTDeJP6jZASJfobkcqzXKJ0xWC4+vcHlGwQ78iJUYXAw==",
	}
	paste2 := &models.Paste{
		Ciphertext: "VGhpcyBpcyBhIG5ldyBzYW1wbGUgc2VjcmV0IG1lc3NhZ2U=",
		PublicKey:  publicKey,
		Signature:  "wWfCM9soU3tkDTc2sJfQcKqfz60j3t4waMizzWskpsSyq6HcA609X2Kf66bZPQ/iEeY3C/ldnMKsHsrlHCBrAw==",
	}
	id1, err := service.Create(ctx, paste1, 3600)
	assert.NoError(t, err, "should create first paste without error")
	id2, err := service.Create(ctx, paste2, 3600)
	assert.NoError(t, err, "should create second paste without error")

	t.Run("valid public key", func(t *testing.T) {
		pastes, err := service.GetByPublicKey(ctx, publicKey)
		assert.NoError(t, err, "should retrieve pastes without error")
		assert.Len(t, pastes, 2, "should return two pastes")
		if assert.Len(t, pastes, 2) {
			expectedIDs := []string{id1, id2}
			assert.Contains(t, expectedIDs, pastes[0].ID, "first paste ID should be one of the expected IDs")
			assert.Contains(t, expectedIDs, pastes[1].ID, "second paste ID should be one of the expected IDs")
			assert.NotEqual(t, pastes[0].ID, pastes[1].ID, "paste IDs should be unique")
		}
	})

	t.Run("empty public key", func(t *testing.T) {
		_, err := service.GetByPublicKey(ctx, "")
		assert.ErrorIs(t, err, utils.ErrEmptyPublicKey, "should return ErrEmptyPublicKey")
	})

	t.Run("invalid public key", func(t *testing.T) {
		_, err := service.GetByPublicKey(ctx, "not-base64")
		assert.ErrorIs(t, err, utils.ErrInvalidPublicKey, "should return ErrInvalidPublicKey")
	})

	t.Run("non-existent user", func(t *testing.T) {
		nonExistentKey, _, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err, "should generate key pair")
		nonExistentKeyB64 := base64.StdEncoding.EncodeToString(nonExistentKey)
		_, err = service.GetByPublicKey(ctx, nonExistentKeyB64)
		assert.ErrorIs(t, err, utils.ErrUserNotFoundForPublicKey, "should return ErrUserNotFoundForPublicKey")
	})
}

func TestUpdate(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	msg := []byte("This is a sample secret message")
	pub, priv, _ := ed25519.GenerateKey(nil)
	sig := ed25519.Sign(priv, msg)
	publicKey = base64.StdEncoding.EncodeToString(pub)

	ctx := context.Background()
	err := service.userRepo.Create(ctx, &models.User{PublicKey: publicKey})

	assert.NoError(t, err, "should create user without error.")

	paste := &models.Paste{
		Ciphertext: base64.StdEncoding.EncodeToString(msg),
		PublicKey:  base64.StdEncoding.EncodeToString(pub),
		Signature:  base64.StdEncoding.EncodeToString(sig),
	}
	id, err := service.Create(ctx, paste, 3600)
	assert.NoError(t, err, "should create a paste.")

	t.Run("valid update", func(t *testing.T) {
		new_msg := []byte("This is a new sample secret message")
		new_sig := ed25519.Sign(priv, new_msg)
		paste.Ciphertext = base64.StdEncoding.EncodeToString(new_msg)
		paste.Signature = base64.StdEncoding.EncodeToString(new_sig)

		err := service.Update(ctx, paste)
		assert.NoError(t, err, "should update user wihout error.")

		fetched, err := service.GetByID(ctx, id)
		assert.NoError(t, err, "should fetch without error")
		assert.Equal(t, id, fetched.ID, "id should match")
		assert.Equal(t, paste.Ciphertext, fetched.Ciphertext, "ciphertext should match")
		assert.Equal(t, paste.Signature, fetched.Signature, "signature should match")
	})
	t.Run("invalid ID", func(t *testing.T) {
		paste := &models.Paste{
			ID:         "invalid-uuid",
			Ciphertext: "VGhpcyBpcyBhbiB1cGRhdGVkIHNhbXBsZQ==",
			PublicKey:  string(pub),
			Signature:  base64.StdEncoding.EncodeToString(ed25519.Sign(priv, []byte("This is an updated sample"))),
		}
		err := service.Update(ctx, paste)
		assert.ErrorIs(t, err, utils.ErrPasteInvalidID, "should return ErrPasteInvalidID")
	})

	t.Run("empty ciphertext", func(t *testing.T) {
		paste := &models.Paste{
			ID:         id,
			Ciphertext: "",
			PublicKey:  publicKey,
			Signature:  base64.StdEncoding.EncodeToString(ed25519.Sign(priv, []byte(""))),
		}
		err := service.Update(ctx, paste)
		assert.ErrorIs(t, err, utils.ErrPasteEmptyCiphertext, "should return ErrPasteEmptyCiphertext")
	})
}
