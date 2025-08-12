package commands

import (
	"context"
	"fmt"
	"go-htmx-sqlite/internal/config"
	"go-htmx-sqlite/internal/database"
	"go-htmx-sqlite/internal/queries"
	"go-htmx-sqlite/internal/service"
	"log"

	"github.com/spf13/cobra"
)

func SeedCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Seed the database with test data",
		Run:   execSeed,
	}
}

func execSeed(cmd *cobra.Command, args []string) {
	cfg := config.LoadConfigFromEnv()
	db := database.New(cfg.Database)
	defer db.Close(cfg.Database)

	q := queries.New(db.GetDB())
	authService := service.NewAuthService(q, db)
	ctx := context.Background()

	users := []struct {
		name     string
		email    string
		password string
	}{
		{"John Doe", "john@example.com", "password123"},
		{"Jane Smith", "jane@example.com", "password456"},
	}

	fmt.Println("Seeding users with accounts...")

	for _, userData := range users {
		createdUser, err := authService.SignUp(ctx, userData.name, userData.email, userData.password, true)
		if err != nil {
			log.Printf("Failed to create user %s: %v", userData.email, err)
			continue
		}
		fmt.Printf("Created user: %s (ID: %d)\n", createdUser.Email, createdUser.ID)
	}

	fmt.Println("Seeding completed!")
}
