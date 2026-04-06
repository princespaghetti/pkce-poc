package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"pkce-poc/internal/auth"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Call the protected API endpoint using the stored access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := auth.LoadToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading token:", err)
			os.Exit(1)
		}
		if token == nil {
			fmt.Fprintln(os.Stderr, "No token found. Run 'login' first.")
			os.Exit(1)
		}

		if auth.IsExpired(token.AccessToken, 30) {
			if token.RefreshToken == "" {
				fmt.Fprintln(os.Stderr, "Token expired and no refresh token available. Run 'login' again.")
				os.Exit(1)
			}
			newToken, err := auth.Refresh(token.RefreshToken)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Token refresh failed — run 'login' again:", err)
				os.Exit(1)
			}
			token = newToken
		}

		req, err := http.NewRequest(http.MethodGet, "http://localhost:"+auth.APIPort+"/api/data", nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error building request:", err)
			os.Exit(1)
		}
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error calling API:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			fmt.Fprintln(os.Stderr, "401 Unauthorized — your token may be expired. Run 'login' again.")
			os.Exit(1)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading response:", err)
			os.Exit(1)
		}
		fmt.Print(string(body))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
