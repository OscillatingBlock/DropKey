package user

import (
	"context"
	"os"
	"testing"

	"Drop-Key/internal/db"
	"Drop-Key/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestNewUserRespository(t *testing.T) {
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialise database")
	assert.NotNil(t, db, "should return a not-nil database")
	defer db.Close()

	repo := NewUserRepository(db)
	assert.NotNil(t, repo, "should return a not-nil repository.")
	assert.Equal(t, db, repo.db, "should set correct database instance.")
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialise database")
	assert.NotNil(t, db, "should return a not-nil database")
	defer db.Close()

	repo := NewUserRepository(db)
	user := &models.User{
		ID:        "user_id",
		PublicKey: "public_key",
	}

	err = repo.Create(ctx, user)
	assert.NoError(t, err, "should create user without error.")

	var fetched models.User
	err = db.NewSelect().Model(&fetched).Where("id = ?", user.ID).Scan(ctx)
	assert.NoError(t, err, "should retrieve user")
	assert.Equal(t, user.ID, fetched.ID, "id should match")
	assert.Equal(t, user.PublicKey, fetched.PublicKey, "public_key should match.")

	_, err = db.NewDropTable().Model(&models.User{}).IfExists().Exec(ctx)
	assert.NoError(t, err, "should drop User table")
}

func TestGetUserById(t *testing.T) {
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialise database")
	assert.NotNil(t, db, "should return a not-nil database")
	defer db.Close()

	repo := NewUserRepository(db)
	user := &models.User{
		ID:        "user_id",
		PublicKey: "public_key",
	}

	_, err = db.NewInsert().Model(user).Exec(ctx)
	assert.NoError(t, err, "should create user without error.")

	var fetched *models.User
	fetched, err = repo.GetByID(ctx, user.ID)
	assert.NoError(t, err, "should retirve user")
	assert.Equal(t, user.ID, fetched.ID, "id should match")
	assert.Equal(t, user.PublicKey, fetched.PublicKey, "public_key should match.")

	_, err = db.NewDropTable().Model(&models.User{}).IfExists().Exec(ctx)
	assert.NoError(t, err, "should drop User table")
}

func TestGetUserByPublicKey(t *testing.T) {
	ctx := context.Background()
	os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")
	db, err := db.InitDB(ctx)
	assert.NoError(t, err, "should initialise database")
	assert.NotNil(t, db, "should return a not-nil database")
	defer db.Close()

	repo := NewUserRepository(db)
	user := &models.User{
		ID:        "user_id",
		PublicKey: "public_key",
	}

	_, err = db.NewInsert().Model(user).Exec(ctx)
	assert.NoError(t, err, "should create user without error.")

	var fetched *models.User
	fetched, err = repo.GetByPublicKey(ctx, user.PublicKey)
	assert.NoError(t, err, "should retirve user")
	assert.Equal(t, user.ID, fetched.ID, "id should match")
	assert.Equal(t, user.PublicKey, fetched.PublicKey, "public_key should match.")

	_, err = db.NewDropTable().Model(&models.User{}).IfExists().Exec(ctx)
	assert.NoError(t, err, "should drop User table")
}
