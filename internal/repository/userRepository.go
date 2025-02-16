package repository

import (
	"log/slog"

	"github.com/akinolaemmanuel49/Memo-Microservices/AuthService/internal/repository/database"
)

type UserRepository struct {
	db     *database.Database
	logger *slog.Logger
}

func NewUserRepository(db *database.Database, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}
