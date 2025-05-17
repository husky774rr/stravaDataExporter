package strava
import (
	"fmt"
	"net/http"
	"net/url"
	"stravaDataExporter/internal/config"
)
type Client struct {
	cfg *config.Config
}
func NewClient(cfg *config.Config, db any) *Client {
	return &Client{cfg: cfg}
}
func HandleOAuthCallback(c *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OAuth callback received")
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
