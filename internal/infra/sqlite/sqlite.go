package sqlite

import (
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	ent "github.com/cgund98/voer/internal/entity/db"
)

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

	err = db.AutoMigrate(&ent.Package{}, &ent.PackageVersion{}, &ent.PackageVersionFile{}, &ent.Message{}, &ent.MessageVersion{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
