// NOTE AI SLOP
package db

import (
	"context"
	"os"
	"testing"
	"time"

	"Drop-Key/internal/paste"
	"Drop-Key/internal/user"
	"github.com/stretchr/testify/assert"
)

// TestInitDB tests the InitDB function
func TestInitDB(t *testing.T) {
	// Create a context for database operations
	ctx := context.Background()

	// Test case 1: Successful database initialization
	t.Run("successful initialization", func(t *testing.T) {
		// Set the DSN for your test database (update credentials if needed)
		os.Setenv("DSN", "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true")

		// Call InitDB
		db, err := InitDB(ctx)
		if err != nil {
			t.Logf("InitDB failed: %v", err) // Debug log
			t.Fatalf("should connect to database without error: %v", err)
		}
		// Check if we got a valid database object
		assert.NotNil(t, db, "should return a non-nil database object")

		// Test table creation by inserting a User and Paste
		user := &user.User{
			ID:        "test-user",
			PublicKey: "test-public-key",
		}
		_, err = db.NewInsert().Model(user).Exec(ctx)
		assert.NoError(t, err, "should insert a user without error")

		paste := &paste.Paste{
			ID:         "test-paste",
			Ciphertext: "encrypted-data",
			Signature:  "signed-data",
			PublicKey:  "test-public-key",
			ExpiresAt:  time.Now().Add(time.Hour),
		}
		_, err = db.NewInsert().Model(paste).Exec(ctx)
		assert.NoError(t, err, "should insert a paste without error")

		// Clean up: Drop tables to reset for the next test
		_, err = db.NewDropTable().Model(paste).IfExists().Exec(ctx)
		assert.NoError(t, err, "should drop Paste table")
		_, err = db.NewDropTable().Model(user).IfExists().Exec(ctx)
		assert.NoError(t, err, "should drop User table")
		// Close the database connection
		err = db.Close()
		assert.NoError(t, err, "should close database connection")
	})

	// Test case 2: Missing DSN
	t.Run("missing DSN", func(t *testing.T) {
		// Unset the DSN environment variable
		os.Unsetenv("DSN")

		// Call InitDB
		db, err := InitDB(ctx)
		// Check if we got an error
		assert.Error(t, err, "should return an error when DSN is missing")
		// Check if the error message is correct
		assert.Contains(t, err.Error(), "Failed to get DSN", "error should mention missing DSN")
		// Check if no database object was returned
		assert.Nil(t, db, "should return nil database object on error")
	})

	// Test case 3: Invalid DSN
	t.Run("invalid DSN", func(t *testing.T) {
		// Set an invalid DSN
		os.Setenv("DSN", "invalid:dsn@tcp(wronghost)/testdb")

		// Call InitDB
		db, err := InitDB(ctx)
		// Check if we got an error
		assert.Error(t, err, "should return an error with invalid DSN")
		// Check if the error message is correct
		assert.Contains(t, err.Error(), "Error while connecting to database", "error should mention connection failure")
		// Check if no database object was returned
		assert.Nil(t, db, "should return nil database object on error")
	})

	// Test case 4: Connection failure
	t.Run("connection failure", func(t *testing.T) {
		// Set a DSN with wrong credentials
		os.Setenv("DSN", "wronguser:wrongpass@tcp(localhost:3306)/testdb?parseTime=true")

		// Call InitDB
		db, err := InitDB(ctx)
		// Check if we got an error
		assert.Error(t, err, "should return an error when connection fails")
		// Check if the error message is correct
		assert.Contains(t, err.Error(), "Error while connecting to database", "error should mention connection failure")
		// Check if no database object was returned
		assert.Nil(t, db, "should return nil database object on error")
	})
}
