package handler

import (
	"context"
	"time"

	"github.com/akinolaemmanuel49/Memo-AuthService/internal/repository/database"
	"github.com/gofiber/fiber/v2"
)

func HandleHealthCheck(db *database.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		databaseHealthMetric, err := db.HealthCheck(context.Background())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status": "error",
				"error":  err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":   "ok",
			"database": databaseHealthMetric,
			"time":     time.Now(),
		})
	}
}
