package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"time"
)

const (
	Auth0Domain  = "YOUR_AUTH0_TENANT.us.auth0.com"
	ClientID     = "YOUR_CLIENT_ID"
	APIAudience  = "http://localhost:8080/api"
	CallbackPort = "8085"
	CallbackURL  = "http://localhost:8085/callback"
	APIPort      = "8080"
)

// Login runs the full PKCE authorization code flow.
func Login() error {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return fmt.Errorf("generating code verifier: %w", err)
	}
	challenge := GenerateCodeChallenge(verifier)

	state, err := GenerateState()
	if err != nil {
		return fmt.Errorf("generating state: %w", err)
	}

	authURL := buildAuthorizeURL(challenge, state)
	fmt.Printf("Opening browser to:\n%s\n\n", authURL)
	if err := exec.Command("open", authURL).Start(); err != nil {
		fmt.Printf("Could not open browser automatically. Please visit the URL above.\n")
	}

	token, err := runCallbackServer(state, verifier)
	if err != nil {
		return fmt.Errorf("callback flow: %w", err)
	}

	if err := SaveToken(token); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	fmt.Println("Login successful! Token saved.")
	return nil
}

func buildAuthorizeURL(challenge, state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", ClientID)
	params.Set("redirect_uri", CallbackURL)
	params.Set("audience", APIAudience)
	params.Set("scope", "openid profile email offline_access")
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")
	params.Set("state", state)
	return fmt.Sprintf("https://%s/authorize?%s", Auth0Domain, params.Encode())
}

func runCallbackServer(expectedState, verifier string) (*TokenResponse, error) {
	var (
		result *TokenResponse
		runErr error
	)

	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":" + CallbackPort, Handler: mux}

	done := make(chan struct{})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		defer close(done)

		q := r.URL.Query()
		if q.Get("state") != expectedState {
			runErr = fmt.Errorf("state mismatch: possible CSRF attack")
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}

		code := q.Get("code")
		if code == "" {
			runErr = fmt.Errorf("no code in callback: %s", q.Get("error_description"))
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		token, err := exchangeCode(code, verifier)
		if err != nil {
			runErr = err
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			return
		}

		result = token
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body><h1>Login successful!</h1><p>You can close this tab and return to the terminal.</p></body></html>`)
	})

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return nil, fmt.Errorf("callback server error: %w", err)
	case <-done:
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("timed out waiting for Auth0 callback")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	if runErr != nil {
		return nil, runErr
	}
	return result, nil
}

// Logout deletes the local token and opens the Auth0 logout URL in the browser.
func Logout() error {
	if err := DeleteToken(); err != nil {
		return fmt.Errorf("deleting token: %w", err)
	}
	logoutURL := fmt.Sprintf("https://%s/v2/logout?client_id=%s&returnTo=http://localhost:8085", Auth0Domain, ClientID)
	if err := exec.Command("open", logoutURL).Start(); err != nil {
		fmt.Println("Could not open browser for Auth0 logout. Token deleted locally.")
	}
	fmt.Println("Logged out.")
	return nil
}

func exchangeCode(code, verifier string) (*TokenResponse, error) {
	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("client_id", ClientID)
	params.Set("code", code)
	params.Set("code_verifier", verifier)
	params.Set("redirect_uri", CallbackURL)

	resp, err := http.PostForm(
		fmt.Sprintf("https://%s/oauth/token", Auth0Domain),
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("token endpoint returned %d: %v", resp.StatusCode, errBody)
	}

	var t TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}
	return &t, nil
}
