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
		codeChan := make(chan string, 1)
		errChan := make(chan error, 1)

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", serverPort))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to start callback server: %w", err)
		}

		server := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != callbackPath {
					http.NotFound(w, r)
					return
				}

				if r.URL.Query().Get("state") != state {
					errChan <- fmt.Errorf("state mismatch")
					http.Error(w, "State mismatch", http.StatusBadRequest)
					return
				}

				if errParam := r.URL.Query().Get("error"); errParam != "" {
					errDesc := r.URL.Query().Get("error_description")
					errChan <- fmt.Errorf("authorization error: %s - %s", errParam, errDesc)
					http.Error(w, fmt.Sprintf("Authorization error: %s", errDesc), http.StatusBadRequest)
					return
				}

				authCode := r.URL.Query().Get("code")
				if authCode == "" {
					errChan <- fmt.Errorf("no authorization code received")
					http.Error(w, "No authorization code", http.StatusBadRequest)
					return
				}

				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				// Embed logo as base64 (placeholder - will be replaced at build time if needed)
				html := strings.Replace(successHTML, "{{LOGO_BASE64}}", getLogoBase64(), 1)
				_, err := fmt.Fprint(w, html)
				if err != nil {
					errChan <- fmt.Errorf("failed to write response: %w", err)
					return
				}

				codeChan <- authCode
			}),
		}

		go func() {
			if err := server.Serve(listener); err != http.ErrServerClosed {
				errChan <- err
			}
		}()

		select {
		case code = <-codeChan:
		case err := <-errChan:
			_ = server.Close()
			return "", "", "", err
		case <-ctx.Done():
			_ = server.Close()
			return "", "", "", ctx.Err()
		case <-time.After(5 * time.Minute):
			_ = server.Close()
			return "", "", "", fmt.Errorf("authentication timed out")
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
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
