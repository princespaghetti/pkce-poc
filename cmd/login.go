package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"pkce-poc/internal/auth"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Auth0 using PKCE and save the access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.Login(); err != nil {
			fmt.Fprintln(os.Stderr, "Login failed:", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
