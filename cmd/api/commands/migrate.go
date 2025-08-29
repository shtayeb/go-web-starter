package commands

import (
	"database/sql"
	"fmt"
	"go-web-starter/internal/config"
	"go-web-starter/internal/database"
	"go-web-starter/sql/postgres/migrations"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func MigrateCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "migrate",
		Short:         "Run database migrations",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          execMigrate,
	}
}
func execMigrate(cmd *cobra.Command, args []string) error {
	cfg := config.LoadConfigFromEnv()
	cmd.Println("Starting database migrations...")
	db := database.New(cfg.Database)
	defer db.Close(cfg.Database)

	sqlDB := db.GetDB()

	if err := runMigrations(sqlDB); err != nil {
		cmd.Println("Migration error:", err)
		return fmt.Errorf("migration failed: %w", err)
	}
	cmd.Println("Migrations completed successfully!")
	return nil
}

func runMigrations(db *sql.DB) error {
	dialect := "postgres"
	// Get the directory of the current source file
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)
	// Navigate to project root and then to postgres migrations
	projectRoot := filepath.Join(currentDir, "..", "..", "..")
	migrationsDir := filepath.Join(projectRoot, "sql", "postgres", "migrations")

	fmt.Printf("Running goose migrations (dialect=%s, dir=%s)\n", dialect, migrationsDir)

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	goose.SetBaseFS(migrations.FS)
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("Goose migrations applied successfully")

	return nil
}
