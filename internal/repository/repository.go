package repository

import (
	"log/slog"

	"github.com/akinolaemmanuel49/Memo-AuthService/internal/repository/database"
)

type Repository struct {
	db     database.Database
	logger *slog.Logger
}

func NewRepository(db database.Database, logger *slog.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}
