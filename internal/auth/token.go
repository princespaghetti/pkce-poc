package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	keychainService = "pkce-poc"
	keychainUser    = "token"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

func SaveToken(t *TokenResponse) error {
	data, err := json.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}
	if err := keyring.Set(keychainService, keychainUser, string(data)); err != nil {
		return fmt.Errorf("saving token to keychain: %w", err)
	}
	return nil
}

func LoadToken() (*TokenResponse, error) {
	data, err := keyring.Get(keychainService, keychainUser)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading token from keychain: %w", err)
	}
	var t TokenResponse
	if err := json.Unmarshal([]byte(data), &t); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	return &t, nil
}

func DeleteToken() error {
	if err := keyring.Delete(keychainService, keychainUser); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("deleting token from keychain: %w", err)
	}
	return nil
}

// IsExpired returns true if the access token expires within bufferSeconds from now.
func IsExpired(accessToken string, bufferSeconds int64) bool {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return true
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return true
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return true
	}
	return time.Now().Unix() >= claims.Exp-bufferSeconds
}

// Refresh exchanges a refresh token for a new token set and saves it.
func Refresh(refreshToken string) (*TokenResponse, error) {
	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("client_id", ClientID)
	params.Set("refresh_token", refreshToken)

	resp, err := http.PostForm(
		fmt.Sprintf("https://%s/oauth/token", Auth0Domain),
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("token refresh returned %d: %v", resp.StatusCode, errBody)
	}

	var t TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, fmt.Errorf("decoding refresh response: %w", err)
	}
	if err := SaveToken(&t); err != nil {
		return nil, fmt.Errorf("saving refreshed token: %w", err)
	}
	return &t, nil
}
