package common

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

//go:embed success.html
var successHTML string

//go:embed logo.png
var logoBytes []byte

// getLogoBase64 returns the logo as a base64 encoded string
func getLogoBase64() string {
	return base64.StdEncoding.EncodeToString(logoBytes)
}

// GenerateCodeVerifier generates a PKCE code verifier
func GenerateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateCodeChallenge generates a PKCE code challenge from a verifier
func GenerateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// CallbackResult contains the result of a callback from an authentication flow
type CallbackResult struct {
	Params url.Values // All query parameters received in the callback
}

// StartCallbackServer starts a local HTTP server and waits for a callback
// This is a generic function that can be used by multiple authentication flows
// It returns all query parameters received without performing any validation
func StartCallbackServer(ctx context.Context, port int, path string) (*CallbackResult, error) {
	resultChan := make(chan *CallbackResult, 1)
	errChan := make(chan error, 1)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != path {
				http.NotFound(w, r)
				return
			}

			// Capture all query parameters
			params := r.URL.Query()

			// Show success page
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			html := strings.Replace(successHTML, "{{LOGO_BASE64}}", getLogoBase64(), 1)
			_, err := fmt.Fprint(w, html)
			if err != nil {
				errChan <- fmt.Errorf("failed to write response: %w", err)
				return
			}

			resultChan <- &CallbackResult{Params: params}
		}),
	}

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	var result *CallbackResult
	select {
	case result = <-resultChan:
	case err := <-errChan:
		_ = server.Close()
		return nil, err
	case <-ctx.Done():
		_ = server.Close()
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		_ = server.Close()
		return nil, fmt.Errorf("authentication timed out")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)

	return result, nil
}

// AuthCodeFlowParams contains parameters for authorization code flow
type AuthCodeFlowParams struct {
	AuthEndpoint string
	ClientID     string
	Scopes       []string
	RedirectURI  string
	RedirectPort int
	UsePKCE      bool // If true, use PKCE extension
}

// AuthenticateAuthCodeFlow performs OAuth2 authorization code flow (with optional PKCE)
func AuthenticateAuthCodeFlow(ctx context.Context, params AuthCodeFlowParams) (code, codeVerifier, redirectURI string, err error) {
	state, err := generateState()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Generate PKCE parameters if enabled
	var codeChallenge string
	if params.UsePKCE {
		codeVerifier, err = GenerateCodeVerifier()
		if err != nil {
			return "", "", "", fmt.Errorf("failed to generate code verifier: %w", err)
		}
		codeChallenge = GenerateCodeChallenge(codeVerifier)
	}

	redirectURI = params.RedirectURI
	serverPort := params.RedirectPort
	callbackPath := "/callback"
	isLocalhost := true

	if redirectURI == "" {
		redirectURI = fmt.Sprintf("http://localhost:%d/callback", params.RedirectPort)
	} else {
		parsedURI, err := url.Parse(redirectURI)
		if err != nil {
			return "", "", "", fmt.Errorf("invalid redirect_uri: %w", err)
		}

		hostname := parsedURI.Hostname()
		isLocalhost = hostname == "localhost" || hostname == "127.0.0.1"

		if isLocalhost {
			if parsedURI.Port() != "" {
				port, err := strconv.Atoi(parsedURI.Port())
				if err != nil {
					return "", "", "", fmt.Errorf("invalid port in redirect_uri: %w", err)
				}
				serverPort = port
			}
			if parsedURI.Path != "" {
				callbackPath = parsedURI.Path
			}
		}
	}

	// Build authorization URL using oauth2.Config
	config := &oauth2.Config{
		ClientID:    params.ClientID,
		RedirectURL: redirectURI,
		Scopes:      params.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL: params.AuthEndpoint,
		},
	}

	// Add PKCE parameters if enabled
	var authCodeOptions []oauth2.AuthCodeOption
	if params.UsePKCE {
		authCodeOptions = append(authCodeOptions,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}

	authURL := config.AuthCodeURL(state, authCodeOptions...)

	if err := OpenBrowser(authURL); err != nil {
		return "", "", "", fmt.Errorf("failed to open browser: %w", err)
	}

	if isLocalhost {
		// Use the generic callback server
		result, err := StartCallbackServer(ctx, serverPort, callbackPath)
		if err != nil {
			return "", "", "", err
		}

		// Validate state parameter
		if result.Params.Get("state") != state {
			return "", "", "", fmt.Errorf("state mismatch")
		}

		// Check for OAuth2 error response
		if errParam := result.Params.Get("error"); errParam != "" {
			errDesc := result.Params.Get("error_description")
			return "", "", "", fmt.Errorf("authorization error: %s - %s", errParam, errDesc)
		}

		// Extract authorization code
		code = result.Params.Get("code")
		if code == "" {
			return "", "", "", fmt.Errorf("no authorization code received")
		}
	} else {
		fmt.Fprintf(os.Stderr, "\nExternal redirect URI detected: %s\n", redirectURI)
		fmt.Fprintf(os.Stderr, "After authorizing, extract the 'code' parameter from the callback URL and enter it below.\n\n")
		fmt.Fprintf(os.Stderr, "Enter authorization code: ")

		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return "", "", "", fmt.Errorf("failed to read authorization code")
		}
		code = strings.TrimSpace(scanner.Text())
		if code == "" {
			return "", "", "", fmt.Errorf("authorization code cannot be empty")
		}
	}

	return code, codeVerifier, redirectURI, nil
}
