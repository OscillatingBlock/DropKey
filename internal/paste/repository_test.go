package paste

import (
	"context"
	"os"
	"testing"
	"time"

	"Drop-Key/internal/db"
	"Drop-Key/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestNewPasteRespository(t *testing.T) {
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialise database")
	assert.NotNil(t, db, "should return a not-nil database")
	defer db.Close()

	repo := NewPasteRepository(db)
	assert.NotNil(t, repo, "should return a not-nil repository.")
	assert.Equal(t, db, repo.db, "should set correct database instance.")
}

func TestCreatePaste(t *testing.T) {
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialise database")
	assert.NotNil(t, db, "should return a not-nil database")
	defer db.Close()

	repo := NewPasteRepository(db)
	paste := &models.Paste{
		ID:         "test-paste",
		Ciphertext: "encrypted-data",
		Signature:  "signed-data",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}

	err = repo.Create(ctx, paste)
	assert.NoError(t, err, "should create paste without error.")

	var fetched models.Paste
	err = db.NewSelect().Model(&fetched).Where("id = ?", paste.ID).Scan(ctx)
	assert.NoError(t, err, "should retrieve paste")
	assert.Equal(t, paste.Ciphertext, fetched.Ciphertext, "ciphertext should match")
	assert.Equal(t, paste.Signature, fetched.Signature, "signature should match")
	assert.Equal(t, paste.PublicKey, fetched.PublicKey, "public_key should match")
	assert.WithinDuration(t, paste.ExpiresAt, fetched.ExpiresAt, time.Second, "expires_at should match")

	_, err = db.NewDropTable().Model(&models.Paste{}).IfExists().Exec(ctx)
	assert.NoError(t, err, "should drop Paste table")
}

func TestGetByIDPaste(t *testing.T) {
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialise database")
	assert.NotNil(t, db, "should return a not-nil database")
	defer db.Close()

	repo := NewPasteRepository(db)
	paste := &models.Paste{
		ID:         "test-paste",
		Ciphertext: "encrypted-data",
		Signature:  "signed-data",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}
	id_string := paste.ID

	err = repo.Create(ctx, paste)
	assert.NoError(t, err, "should create paste without error.")

	var fetched *models.Paste
	fetched, err = repo.GetByID(ctx, id_string)
	assert.Equal(t, paste.ID, fetched.ID, "ID should be equal")
	assert.Equal(t, paste.Ciphertext, fetched.Ciphertext, "ciphertext should be equal")
	assert.Equal(t, paste.Signature, fetched.Signature, "signature should be equal")
	assert.Equal(t, paste.PublicKey, fetched.PublicKey, "public_key should match")
	assert.WithinDuration(t, paste.ExpiresAt, fetched.ExpiresAt, time.Second, "expires_at should match")

	_, err = db.NewDropTable().Model(&models.Paste{}).IfExists().Exec(ctx)
	assert.NoError(t, err, "should drop Paste table")
}

func TestUpdatePaste(t *testing.T) {
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialise database")
	assert.NotNil(t, db, "should return a not-nil database")
	defer db.Close()

	repo := NewPasteRepository(db)
	paste := &models.Paste{
		ID:         "test-paste",
		Ciphertext: "encrypted-data",
		Signature:  "signed-data",
		PublicKey:  "test-public-key",
		ExpiresAt:  time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}
	id_string := paste.ID

	err = repo.Create(ctx, paste)
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

	err = repo.UpdatePaste(ctx, paste)

	var fetchedAgain *models.Paste
	fetchedAgain, err = repo.GetByID(ctx, id_string)
	assert.Equal(t, paste.ID, fetchedAgain.ID, "ID should be equal")
	assert.Equal(t, paste.Ciphertext, fetchedAgain.Ciphertext, "ciphertext should be equal")
	assert.Equal(t, paste.Signature, fetchedAgain.Signature, "signature should be equal")
	assert.Equal(t, paste.PublicKey, fetchedAgain.PublicKey, "public_key should match")
	assert.WithinDuration(t, paste.ExpiresAt, fetchedAgain.ExpiresAt, time.Second, "expires_at should match")

	_, err = db.NewDropTable().Model(&models.Paste{}).IfExists().Exec(ctx)
	assert.NoError(t, err, "should drop Paste table")
}
