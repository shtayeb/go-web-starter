package main

import (
	"go-web-starter/cmd/api/commands"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "app",
		Short: "A web application built with Go.",
	}

	// CLI commands
	rootCmd.AddCommand(
		commands.ServerCommand(),
		commands.SeedCommand(),
		commands.PingCommand(),
		commands.MigrateCommand(),
	)

	rootCmd.Execute()
}
