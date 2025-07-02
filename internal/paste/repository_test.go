package paste

import (
	"context"
	"os"
	"testing"
	"time"

	"Drop-Key/internal/db"
	"Drop-Key/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

func setupTestDB(t *testing.T) (*bun.DB, *pasteRepository, func()) {
	t.Helper()
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialize database")
	repo := NewPasteRepository(db)
	cleanup := func() {
		_, err := db.NewDropTable().Model(&models.Paste{}).IfExists().Exec(ctx)
		assert.NoError(t, err, "should drop Paste table")
		db.Close()
	}
	return db, repo, cleanup
}

func TestNewPasteRespository(t *testing.T) {
	db, repo, cleanup := setupTestDB(t)
	defer cleanup()

	assert.NotNil(t, repo, "should return a not-nil repository.")
	assert.Equal(t, db, repo.db, "should set correct database instance.")
}

func TestCreatePaste(t *testing.T) {
	db, repo, cleanup := setupTestDB(t)
	defer cleanup()

	paste := &models.Paste{
		ID:         "test-paste",
		Ciphertext: "encrypted-data",
		Signature:  "signed-data",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}

	ctx := context.Background()

	err := repo.Create(ctx, paste)
	assert.NoError(t, err, "should create paste without error.")

	var fetched models.Paste
	err = db.NewSelect().Model(&fetched).Where("id = ?", paste.ID).Scan(ctx)
	assert.NoError(t, err, "should retrieve paste")
	assert.Equal(t, paste.Ciphertext, fetched.Ciphertext, "ciphertext should match")
	assert.Equal(t, paste.Signature, fetched.Signature, "signature should match")
	assert.Equal(t, paste.PublicKey, fetched.PublicKey, "public_key should match")
	assert.WithinDuration(t, paste.ExpiresAt, fetched.ExpiresAt, time.Second, "expires_at should match")
}

func TestGetByIDPaste(t *testing.T) {
	ctx := context.Background()
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	paste := &models.Paste{
		ID:         "test-paste",
		Ciphertext: "encrypted-data",
		Signature:  "signed-data",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}
	id_string := paste.ID

	err := repo.Create(ctx, paste)
	assert.NoError(t, err, "should create paste without error.")

	var fetched *models.Paste
	fetched, err = repo.GetByID(ctx, id_string)
	assert.Equal(t, paste.ID, fetched.ID, "ID should be equal")
	assert.Equal(t, paste.Ciphertext, fetched.Ciphertext, "ciphertext should be equal")
	assert.Equal(t, paste.Signature, fetched.Signature, "signature should be equal")
	assert.Equal(t, paste.PublicKey, fetched.PublicKey, "public_key should match")
	assert.WithinDuration(t, paste.ExpiresAt, fetched.ExpiresAt, time.Second, "expires_at should match")

	expiredPaste := &models.Paste{
		ID:         "expired-paste",
		Ciphertext: "expired-data",
		Signature:  "expired-signature",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(-time.Hour).Truncate(time.Second),
	}
	err = repo.Create(ctx, expiredPaste)
	assert.NoError(t, err, "should insert expired paste")

	fetched, err = repo.GetByID(ctx, expiredPaste.ID)
	assert.ErrorIs(t, err, ErrPasteExpired, "should return ErrPasteExpired for expired paste")
	assert.Nil(t, fetched, "should return nil for expired paste")

	fetched, err = repo.GetByID(ctx, "non-existent")
	assert.Error(t, err, "should return error for non-existent paste")
	assert.NotErrorIs(t, err, ErrPasteExpired, "error should not be ErrPasteExpired for non-existent paste")
	assert.Nil(t, fetched, "should return nil for non-existent paste")
}

func TestUpdatePaste(t *testing.T) {
	ctx := context.Background()
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	paste := &models.Paste{
		ID:         "test-paste",
		Ciphertext: "encrypted-data",
		Signature:  "signed-data",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}
	id_string := paste.ID

	err := repo.Create(ctx, paste)
	assert.NoError(t, err, "should create paste without error.")

	var fetched *models.Paste
	fetched, err = repo.GetByID(ctx, id_string)
	assert.Equal(t, paste.ID, fetched.ID, "ID should be equal")
	assert.Equal(t, paste.Ciphertext, fetched.Ciphertext, "ciphertext should be equal")
	assert.Equal(t, paste.Signature, fetched.Signature, "signature should be equal")
	assert.Equal(t, paste.PublicKey, fetched.PublicKey, "public_key should match")
	assert.WithinDuration(t, paste.ExpiresAt, fetched.ExpiresAt, time.Second, "expires_at should match")

	paste.Ciphertext = "updated-encrypted-data"
	paste.Signature = "updated-signed-data"

	err = repo.Update(ctx, paste)

	var fetchedAgain *models.Paste
	fetchedAgain, err = repo.GetByID(ctx, id_string)
	assert.Equal(t, paste.ID, fetchedAgain.ID, "ID should be equal")
	assert.Equal(t, paste.Ciphertext, fetchedAgain.Ciphertext, "ciphertext should be equal")
	assert.Equal(t, paste.Signature, fetchedAgain.Signature, "signature should be equal")
	assert.Equal(t, paste.PublicKey, fetchedAgain.PublicKey, "public_key should match")
	assert.WithinDuration(t, paste.ExpiresAt, fetchedAgain.ExpiresAt, time.Second, "expires_at should match")
}

func TestGetByPublicKey(t *testing.T) {
	_, repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	paste1 := &models.Paste{
		ID:         "test-paste-1",
		Ciphertext: "encrypted-data-1",
		Signature:  "signed-data-1",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}
	paste2 := &models.Paste{
		ID:         "test-paste-2",
		Ciphertext: "encrypted-data-2",
		Signature:  "signed-data-2",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}

	err := repo.Create(ctx, paste1)
	assert.NoError(t, err, "should create first paste without error")
	err = repo.Create(ctx, paste2)
	assert.NoError(t, err, "should create second paste without error")

	pastes, err := repo.GetByPublicKey(ctx, "test-public-key")
	assert.NoError(t, err, "should retrieve pastes without error")
	assert.Len(t, pastes, 2, "should return two pastes")
	assert.Equal(t, paste1.ID, pastes[0].ID, "first paste ID should match")
	assert.Equal(t, paste2.ID, pastes[1].ID, "second paste ID should match")

	expiredPaste := &models.Paste{
		ID:         "expired-paste",
		Ciphertext: "expired-data",
		Signature:  "expired-signature",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(-time.Hour).Truncate(time.Second),
	}
	err = repo.Create(ctx, expiredPaste)
	assert.NoError(t, err, "should create expired paste")

	pastes, err = repo.GetByPublicKey(ctx, "test-public-key")
	assert.NoError(t, err, "should retrieve pastes without error")
	assert.Len(t, pastes, 2, "should return only non-expired pastes")
}
