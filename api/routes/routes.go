package routes

import (
	"github.com/akinolaemmanuel49/Memo-AuthService/internal/handler"
	"github.com/akinolaemmanuel49/Memo-AuthService/internal/repository/database"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up the routes for the server.
func SetupRoutes(app *fiber.App, db *database.Database) {
	app.Get("/health", handler.HandleHealthCheck(db))
}
