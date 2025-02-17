package handler_test

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/akinolaemmanuel49/Memo-AuthService/config"
	"github.com/akinolaemmanuel49/Memo-AuthService/internal/handler"
	"github.com/akinolaemmanuel49/Memo-AuthService/internal/repository/database"
	"github.com/akinolaemmanuel49/Memo-AuthService/internal/repository/dtos"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *database.Database {
	ctx := context.Background()

	os.Setenv("ENVIRONMENT", "testing")

	// Change to the project root directory
	err := os.Chdir("../.")
	if err != nil {
		log.Fatalf("Error changing working directory: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize the database connection
	db, err := database.New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Ensure database is clean before running tests
	cleanupDB(db)

	t.Cleanup(func() {
		cleanupDB(db)
	})

	return db
}

// Cleans up test tables before/after each test
func cleanupDB(db *database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queries := []string{
		// "DELETE FROM users;",
		// "DELETE FROM sessions;",
		// "DELETE FROM logs;",
	}
	for _, query := range queries {
		// _, err := db.Pool.Exec(ctx, query)
		_, err := db.Pool().Exec(ctx, query)
		if err != nil {
			log.Printf("Failed to clean up test database: %v", err)
		}
	}
}

func TestHandleHealthCheck_WithRealDB(t *testing.T) {
	app := fiber.New()
	db := setupTestDB(t)

	app.Get("/health", handler.HandleHealthCheck(db))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response dtos.HealthCheck
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "ok", response.Status)
	assert.WithinDuration(t, time.Now(), response.Time, time.Second)
}
