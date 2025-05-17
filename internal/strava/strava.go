package strava
import (
	"fmt"
	"net/http"
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
		fmt.Fprintln(w, "OAuth callback")
	}
}
func HandleLogin(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Redirect to Strava OAuth")
	}
}
