package sqlite

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/cgund98/voer/internal/infra/logging"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type GooseLogger struct {
	Logger *slog.Logger
}

func (l *GooseLogger) Fatalf(msg string, args ...any) {
	l.Logger.Error(fmt.Sprintf(msg, args...), "source", "goose")
}

func (l *GooseLogger) Printf(msg string, args ...any) {
	l.Logger.Info(fmt.Sprintf(msg, args...), "source", "goose")
}

// genUserSpecificDbPath generates a user-specific SQLite database path. Should work for both Linux and MacOS.
func genUserSpecificDbPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".voer", "voer.db"), nil
}

// NewDB creates a new SQLite database connection and auto-migrates the database schema
func NewDB(dbPath string) (*gorm.DB, error) {
	// Default path to OS-specific SQLite database file
	if dbPath == "" {
		newPath, err := genUserSpecificDbPath()
		if err != nil {
			return nil, err
		}
		dbPath = newPath
	}

	// Create the database directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, err
		}
	}

	// Open the database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Enable foreign key constraints
	if res := db.Exec("PRAGMA foreign_keys = ON", nil); res.Error != nil {
		return nil, res.Error
	}

	// Run migrations
	goose.SetLogger(&GooseLogger{Logger: logging.Logger})

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}

	if err := goose.Up(sqlDb, "migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
