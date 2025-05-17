package strava

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"stravaDataExporter/internal/config"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type Client struct {
	cfg          *config.Config
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	Logger       *slog.Logger
}

// IsAuthenticated checks if the client has a valid access token.
func (c *Client) IsAuthenticated() bool {
	return c.AccessToken != ""
}

// Example of a handler that requires authentication.
func (c *Client) AuthRequiredHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !c.IsAuthenticated() {
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

func NewClient(cfg *config.Config, db any) *Client {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &Client{cfg: cfg, Logger: logger}
}

func HandleOAuthCallback(c *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			c.Logger.Warn("missing code", "path", r.URL.Path)
			http.Error(w, "Missing code in query", http.StatusBadRequest)
			return
		}

		form := url.Values{}
		form.Set("client_id", c.cfg.StravaClientID)
		form.Set("client_secret", c.cfg.StravaClientSecret)
		form.Set("code", code)
		form.Set("grant_type", "authorization_code")

		resp, err := http.PostForm("https://www.strava.com/oauth/token", form)
		if err != nil {
			c.Logger.Error("failed to get token", "error", err)
			http.Error(w, "Failed to get token from Strava", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.Logger.Error("failed to read response", "error", err)
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			return
		}

		var token TokenResponse
		if err := json.Unmarshal(body, &token); err != nil {
			c.Logger.Error("failed to unmarshal token", "error", err)
			http.Error(w, "Failed to parse token JSON", http.StatusInternalServerError)
			return
		}

		c.AccessToken = token.AccessToken
		c.RefreshToken = token.RefreshToken
		c.ExpiresAt = token.ExpiresAt

		c.Logger.Info("access token obtained",
			"access_token", token.AccessToken,
			"refresh_token", token.RefreshToken,
			"expires_at", time.Unix(token.ExpiresAt, 0),
		)

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<html><body><h2>Authentication complete.</h2><p>Please close this window.</p></body></html>`)
	}
}

func HandleLogin(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectURL := "http://localhost:8080/auth/callback"
		stravaAuthURL := fmt.Sprintf(
			"https://www.strava.com/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%s&approval_prompt=auto&scope=activity:read_all",
			cfg.StravaClientID,
			url.QueryEscape(redirectURL),
		)
		http.Redirect(w, r, stravaAuthURL, http.StatusFound)
	}
}

func (c *Client) RefreshAccessToken() error {
	if c.RefreshToken == "" {
		c.Logger.Warn("no refresh token")
		return fmt.Errorf("no refresh token available")
	}

	form := url.Values{}
	form.Set("client_id", c.cfg.StravaClientID)
	form.Set("client_secret", c.cfg.StravaClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", c.RefreshToken)

	resp, err := http.PostForm("https://www.strava.com/oauth/token", form)
	if err != nil {
		c.Logger.Error("failed to refresh token", "error", err)
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("failed to read refresh response", "error", err)
		return fmt.Errorf("failed to read refresh response: %w", err)
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		c.Logger.Error("failed to unmarshal refresh token", "error", err)
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	c.AccessToken = token.AccessToken
	c.RefreshToken = token.RefreshToken
	c.ExpiresAt = token.ExpiresAt

	c.Logger.Info("access token refreshed",
		"access_token", token.AccessToken,
		"expires_at", time.Unix(token.ExpiresAt, 0),
	)

	return nil
}
