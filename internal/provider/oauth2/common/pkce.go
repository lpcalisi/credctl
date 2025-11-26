package common

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateCodeChallenge(verifier string) string {
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

func AuthenticatePKCEFlow(ctx context.Context, params PKCEFlowParams) (code, codeVerifier, redirectURI string, err error) {
	codeVerifier, err = generateCodeVerifier()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate code verifier: %w", err)
	}
	codeChallenge := generateCodeChallenge(codeVerifier)

	state, err := generateState()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate state: %w", err)
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

	authURL, err := url.Parse(params.AuthEndpoint)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid auth endpoint: %w", err)
	}

	q := authURL.Query()
	q.Set("client_id", params.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	if len(params.Scopes) > 0 {
		q.Set("scope", strings.Join(params.Scopes, " "))
	}
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	authURL.RawQuery = q.Encode()

	if err := OpenBrowser(authURL.String()); err != nil {
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

				w.Header().Set("Content-Type", "text/html")
				_, err := fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Successful</title></head>
<body>
<h1>Authentication Successful</h1>
<p>You can close this tab and return to the terminal.</p>
<script>window.close();</script>
</body>
</html>`)
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
