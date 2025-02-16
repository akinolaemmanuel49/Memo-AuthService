package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/akinolaemmanuel49/Memo-AuthService/logging"
	"github.com/joho/godotenv"
)

// Environment represents possible runtime environments
type Environment string

const (
	Production  Environment = "production"
	Staging     Environment = "staging"
	Testing     Environment = "testing"
	Development Environment = "development"
)

// IsValid checks if the environment value is valid
func (e Environment) IsValid() bool {
	switch e {
	case Production, Staging, Testing, Development:
		return true
	default:
		return false
	}
}

// String implements the Stringer interface for Environment
func (e Environment) String() string {
	return string(e)
}

// LogConfig holds logging-specific configuration
type LogConfig struct {
	Level slog.Level
	Json  bool   // Whether to use JSON formatter
	File  string // Optional log file path
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name string
	Host string
	Port int
}

// DatabaseConfig holds database-specific configuration
type DatabaseConfig struct {
	URI               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// Config holds all configuration values
type Config struct {
	Environment Environment
	Logger      *slog.Logger
	LogConfig   LogConfig
	Service     ServiceConfig
	Database    DatabaseConfig
}

// setupLogger configures the slog logger based on environment
func setupLogger(env Environment, logCfg LogConfig, serviceName string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     logCfg.Level,
		AddSource: true,
	}

	var handler slog.Handler
	if logCfg.Json {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	if logCfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(logCfg.File), 0o755); err == nil {
			file, err := os.OpenFile(logCfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
			if err == nil {
				handler = logging.NewMultiHandler(handler, slog.NewJSONHandler(file, opts))
			}
		}
	}

	logger := slog.New(handler)

	// Add default context
	logger = logger.With(
		slog.String("environment", env.String()),
		slog.String("service", serviceName),
	)

	return logger
}

// LoadConfig initializes and loads the configuration
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Get environment from ENV var, default to development
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	cfg.Environment = Environment(strings.ToLower(env))
	if !cfg.Environment.IsValid() {
		return nil, fmt.Errorf("invalid environment: %s", env)
	}

	// Load environment variables from .env file
	err := godotenv.Load(fmt.Sprintf(".env.%s", cfg.Environment))
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Load service configuration
	cfg.Service.Name = getEnv("SERVICE_NAME", "authService")
	cfg.Service.Host = getEnv("SERVICE_HOST", "localhost")
	cfg.Service.Port = getEnvAsInt("SERVICE_PORT", 8000)

	// Load database configuration
	if getEnv("DATABASE_URI", "") == "" {
		return nil, fmt.Errorf("DATABASE_URI environment variable is required")
	}
	cfg.Database.URI = getEnv("DATABASE_URI", "")
	cfg.Database.MaxConns = int32(getEnvAsInt("DATABASE_MAX_CONNS", 25))
	cfg.Database.MinConns = int32(getEnvAsInt("DATABASE_MIN_CONNS", 5))
	cfg.Database.MaxConnLifetime = getEnvAsDuration("DATABASE_MAX_CONN_LIFETIME", "1h")
	cfg.Database.MaxConnIdleTime = getEnvAsDuration("DATABASE_MAX_CONN_IDLE_TIME", "30m")
	cfg.Database.HealthCheckPeriod = getEnvAsDuration("DATABASE_HEALTH_CHECK_PERIOD", "1m")

	// Setup logging configuration based on environment
	cfg.LogConfig = LogConfig{
		Level: getLogLevel(cfg.Environment),
		Json:  cfg.Environment == Production || cfg.Environment == Staging,
		File:  getEnv("LOG_FILE", ""),
	}

	// Initialize the logger with service name
	cfg.Logger = setupLogger(cfg.Environment, cfg.LogConfig, cfg.Service.Name)

	return cfg, nil
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt reads an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// getEnvAsDuration reads an environment variable as a time.Duration or returns a default value
func getEnvAsDuration(key, defaultValue string) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		duration, err := time.ParseDuration(defaultValue)
		if err != nil {
			return 0
		}
		return duration
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return duration
}

// getLogLevel returns appropriate log level based on environment
func getLogLevel(env Environment) slog.Level {
	switch env {
	case Production, Staging:
		return slog.LevelInfo
	case Testing:
		return slog.LevelWarn
	default:
		return slog.LevelDebug
	}
}
