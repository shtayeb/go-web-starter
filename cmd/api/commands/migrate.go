package commands

import (
	"database/sql"
	"fmt"
	"go-web-starter/internal/config"
	"go-web-starter/internal/database"
	"log"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func MigrateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run:   execMigrate,
	}
}

func execMigrate(cmd *cobra.Command, args []string) {
	cfg := config.LoadConfigFromEnv()
	db := database.New(cfg.Database)
	defer db.Close(cfg.Database)

	sqlDB := db.GetDB()

	if err := runMigrations(sqlDB, cfg.Database.Type); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migrations completed successfully!")
}

func runMigrations(db *sql.DB, dbType string) error {
	var migrationsDir string
	var dialect string

	fmt.Println(dbType)

	switch dbType {
	case "sqlite", "sqlite3":
		dialect = "sqlite3"
		// Get the directory of the current source file
		_, filename, _, _ := runtime.Caller(0)
		currentDir := filepath.Dir(filename)
		// Navigate to project root and then to sqlite migrations
		projectRoot := filepath.Join(currentDir, "..", "..", "..")
		migrationsDir = filepath.Join(projectRoot, "sql", "sqlite", "migrations")
	case "postgres", "postgresql":
		fallthrough
	default:
		dialect = "postgres"
		// Get the directory of the current source file
		_, filename, _, _ := runtime.Caller(0)
		currentDir := filepath.Dir(filename)
		// Navigate to project root and then to postgres migrations
		projectRoot := filepath.Join(currentDir, "..", "..", "..")
		migrationsDir = filepath.Join(projectRoot, "sql", "postgres", "migrations")
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
