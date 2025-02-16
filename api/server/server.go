package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/akinolaemmanuel49/Memo-AuthService/api/routes"
	"github.com/akinolaemmanuel49/Memo-AuthService/config"
	"github.com/akinolaemmanuel49/Memo-AuthService/internal/repository/database"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app    *fiber.App
	config *config.Config
	logger *slog.Logger
	db     *database.Database
}

func NewServer(cfg *config.Config, logger *slog.Logger, db *database.Database) *Server {
	app := fiber.New(fiber.Config{
		AppName: cfg.Service.Name,
	})
	server := &Server{
		app:    app,
		config: cfg,
		logger: cfg.Logger.With(slog.String("component", "server")),
		db:     db,
	}

	// Setup routes
	routes.SetupRoutes(app, db)

	return server
}

func (s *Server) Start() error {
	// addr := fmt.Sprintf("%s:%d", s.config.Service.Host, s.config.Service.Port)
	// s.logger.Info("starting server", "address", addr)
	addr := fmt.Sprintf(":%d", 8000)
	return s.app.Listen(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	return s.app.ShutdownWithContext(ctx)
}
