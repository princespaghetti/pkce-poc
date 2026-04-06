package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"pkce-poc/internal/api"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the protected API server on localhost:8080",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := api.Serve(); err != nil {
			fmt.Fprintln(os.Stderr, "Server error:", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
