package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v3"
	"github.com/auth0/go-jwt-middleware/v3/jwks"
	"github.com/auth0/go-jwt-middleware/v3/validator"
)

const (
	auth0Domain = "YOUR_AUTH0_TENANT.us.auth0.com"
	apiAudience = "http://localhost:8080/api"
	apiPort     = "8080"
)

func Serve() error {
	issuerURL, err := url.Parse("https://" + auth0Domain + "/")
	if err != nil {
		return fmt.Errorf("parsing issuer URL: %w", err)
	}

	provider, err := jwks.NewCachingProvider(
		jwks.WithIssuerURL(issuerURL),
	)
	if err != nil {
		return fmt.Errorf("setting up JWKS provider: %w", err)
	}

	jwtValidator, err := validator.New(
		validator.WithKeyFunc(provider.KeyFunc),
		validator.WithAlgorithm(validator.RS256),
		validator.WithIssuer(issuerURL.String()),
		validator.WithAudience(apiAudience),
	)
	if err != nil {
		return fmt.Errorf("setting up JWT validator: %w", err)
	}

	middleware, err := jwtmiddleware.New(
		jwtmiddleware.WithValidator(jwtValidator),
	)
	if err != nil {
		return fmt.Errorf("setting up JWT middleware: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/data", middleware.CheckJWT(http.HandlerFunc(dataHandler)))

	fmt.Printf("API server listening on http://localhost:%s\n", apiPort)
	return http.ListenAndServe(":"+apiPort, mux)
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"friend": "Marco"})
}
