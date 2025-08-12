package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func PingCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Run ping command",
		Run:   execPingCommand,
	}
}

func execPingCommand(cmd *cobra.Command, args []string) {
	fmt.Println("Pong...")
}
