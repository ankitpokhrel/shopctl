// Package auth sets up the authentication flow for the Shopify app.
//
// See https://help.shopify.com/en/api/getting-started/authentication/oauth
package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/oauth2"

	"github.com/ankitpokhrel/shopctl/pkg/browser"
)

var (
	// As this app isn't submitted to the Shopify app store yet, users have to configure their own private app to get the
	// client ID and Secret. We'll try to find a way to simplify the setup process so that users won’t have to create and
	// configure their own Shopify app just to use this tool. However, this will still be one of the possibilities.
	//
	// Client ID of the Shopify app.
	oauthClientID = os.Getenv("SHOPCTL_CLIENT_ID")
	// Client Secret of the Shopify app.
	oauthClientSecret = os.Getenv("SHOPCTL_CLIENT_SECRET")
	// See https://shopify.dev/docs/api/usage/access-scopes#authenticated-access-scopes for
	// the list of available scopes. Scopes will be added as necessary.
	scopes = []string{
		// Product, ProductVariant, Collection, Inventory.
		"write_products",
		"read_product_listings",
		"read_inventory",

		// Customer data.
		"write_customers",

		// Files.
		"write_files",
	}
)

//go:embed templates/success.html
var successHTML string

// Flow is an oAuth flow.
type Flow struct {
	Token *oauth2.Token

	cfg   *oauth2.Config
	state string
}

// NewFlow creates a new oAuth flow for the given store.
func NewFlow(store string) *Flow {
	redirectURL := os.Getenv("SHOPCTL_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://127.0.0.1/shopctl/auth/callback"
	}
	return &Flow{
		cfg: &oauth2.Config{
			ClientID:     oauthClientID,
			ClientSecret: oauthClientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     shopEndpoint(store),
			Scopes:       scopes,
		},
	}
}

func (f *Flow) handleCallback(w http.ResponseWriter, r *http.Request, done chan struct{}) {
	defer close(done)

	if valid := verifySignature(r, oauthClientSecret); !valid {
		http.Error(w, "Invalid signature", http.StatusBadRequest)
	}

	shop := r.URL.Query().Get("shop")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	// Verify nonce and hostname.
	if state != f.state || validateShopURL(shop) {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Exchange code for the access token.
	token, err := f.cfg.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	f.Token = token

	w.Header().Set("Content-Type", "text/html")
	if _, err = fmt.Fprint(w, successHTML); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// InitiateAuth initiates the Shopify authentication process.
func (f *Flow) Initiate() error {
	state, err := generateState(8) //nolint:mnd
	if err != nil {
		return err
	}
	f.state = state

	authURL := f.cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Print instructions for the user.
	fmt.Printf("\n! Initiating Shopify oAuth flow\n")
	fmt.Printf("Press \033[1mEnter\033[0m to open authentication URL in your browser...")

	// Wait for an enter key.
	if _, err = fmt.Scanln(); err != nil {
		return err
	}

	if err := browser.Browse(authURL); err != nil {
		fmt.Printf("\n! Failed opening a web browser at %s\n", authURL)
		fmt.Printf("Please try entering the URL in your browser manually\n")
	}

	if err := f.configureCallbackServer(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to start a callback server: %s\n", err)
	}

	return nil
}

func (f *Flow) configureCallbackServer() error {
	mux := http.NewServeMux()
	done := make(chan struct{})

	mux.HandleFunc("/shopctl/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		f.handleCallback(w, r, done)
	})

	server := &http.Server{
		Addr:    "", // TODO: Make the port dynamic.
		Handler: mux,
	}

	go func() {
		_ = server.ListenAndServe()
	}()

	<-done

	return server.Close()
}

// shopEndpoint configures and returns a new oauth2.Endpoint for the given shopify shop.
func shopEndpoint(shop string) oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  "https://" + shop + "/admin/oauth/authorize",
		TokenURL: "https://" + shop + "/admin/oauth/access_token",
	}
}

// generateState generates a cryptographically secure random string to use as the OAuth state/nonce parameter.
func generateState(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(bytes)), nil
}

// We'd need to check if the request actually came from the Shopify by validating the provided HMAC SHA256 signature.
// See https://shopify.dev/docs/apps/build/authentication-authorization/access-tokens/authorization-code-grant#step-1-verify-the-installation-request
func verifySignature(r *http.Request, secret string) bool {
	query := r.URL.Query()

	if len(query) == 0 || len(query["hmac"]) < 1 {
		return false
	}

	sig := query["hmac"][0]
	query.Del("hmac")

	sigComputed := query.Encode()

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(sigComputed))

	sha := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(sha), []byte(sig))
}

// validateShopURL validates if the provided shop URL is a valid Shopify hostname.
func validateShopURL(shopURL string) bool {
	// Regex to match internal Shopify shop hostname format.
	regex := `^https?:\/\/[a-zA-Z0-9][a-zA-Z0-9\-]*\.myshopify\.com\/?$`

	re := regexp.MustCompile(regex)

	return re.MatchString(shopURL)
}
