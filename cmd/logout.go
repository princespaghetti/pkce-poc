package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"pkce-poc/internal/auth"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Delete the stored token and log out from Auth0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.Logout(); err != nil {
			fmt.Fprintln(os.Stderr, "Logout failed:", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
