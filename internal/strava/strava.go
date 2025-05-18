package strava

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"stravaDataExporter/internal/config"
	"stravaDataExporter/internal/influxdb"
	"stravaDataExporter/internal/model"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
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
	dbClient     *influxdb2.Client
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	Logger       *slog.Logger
	FTPRecords   []FTPRecord
}

func NewClient(cfg *config.Config, dbClient *influxdb2.Client) *Client {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &Client{cfg: cfg, dbClient: dbClient, Logger: logger}
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
		if nil != next {
			next(w, r)
		}
	}
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
			"expires_in", token.ExpiresIn,
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

func (c *Client) FetchActivities() error {
	activities, err := c.FetchActivitiesFromStrava()
	if err != nil {
		return err
	}

	c.CalculateActivities(activities)
	c.AggregateSummary(activities)

	activitiesJsons, err := json.Marshal(activities)
	if err == nil {
		c.Logger.Info("activities fetched from Strava", "activities", string(activitiesJsons))
	} else {
		c.Logger.Error("failed to marshal activities to JSON", "error", err)
	}

	if err := influxdb.SaveActivity(c.dbClient, activities, c.cfg.InfluxDBOrg, c.cfg.InfluxDBBucket); err != nil {
		c.Logger.Error("failed to save activity to InfluxDB", "error", err)
	}
	return nil
}

// FetchActivitiesFromStrava は1時間ごとにStrava APIからデータを取得（仮）
func (c *Client) FetchActivitiesFromStrava() ([]model.Activity, error) {
	c.Logger.Info("Fetching activities from Strava")

	since := time.Now().Add((-180 * 24) * time.Hour).Unix()
	until := time.Now().Unix()

	apiUrl := "https://www.strava.com/api/v3/athlete/activities"
	params := url.Values{}
	params.Set("after", strconv.FormatInt(since, 10))
	params.Set("before", strconv.FormatInt(until, 10))
	params.Set("per_page", "200")

	req, err := http.NewRequest("GET", apiUrl+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var activities []model.Activity
	if err := json.Unmarshal(body, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func (c *Client) CalculateActivities(activities []model.Activity) {
	for i := range activities {
		// FTPを取得
		activities[i].FTP = c.GetFTP(activities[i].StartTime)
		// TSS, NPを計算
		ComputeTSSNP(&activities[i])
	}
}

// ComputeTSSNP は FTP をもとに TSS, NP を算出する
func ComputeTSSNP(a *model.Activity) {
	a.NP = a.AverageWatt // 仮：NP=平均ワット
	a.TSS = (a.Duration / 3600) * (a.NP / a.FTP) * (a.NP / a.FTP) * 100
}

// AggregateSummary は週/月/年ごとの集計を行う（仮）
func (c *Client) AggregateSummary(activities []model.Activity) {
	weekly := make(map[string]float64)
	monthly := make(map[string]float64)
	yearly := make(map[string]float64)

	for _, a := range activities {
		y, m, _ := a.StartTime.Date()
		_, w := a.StartTime.ISOWeek()
		weekKey := fmt.Sprintf("%04d-W%02d", y, w)
		monthKey := fmt.Sprintf("%04d-%02d", y, m)
		yearKey := fmt.Sprintf("%04d", y)

		weekly[weekKey] += a.TSS
		monthly[monthKey] += a.TSS
		yearly[yearKey] += a.TSS
	}

	fmt.Println("Weekly TSS:", weekly)
	fmt.Println("Monthly TSS:", monthly)
	fmt.Println("Yearly TSS:", yearly)
}

type FTPRecord struct {
	Date time.Time
	FTP  float64
}

func (c *Client) LoadFTPHistoricalData() error {
	file, err := os.Open(c.cfg.FtpFileAbsPath)
	if err != nil {
		return err
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = 2
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var ftpData []FTPRecord
	for _, rec := range records[1:] {
		date, err := time.Parse("2006-01-02", rec[0])
		if err != nil {
			return err
		}
		ftp, err := strconv.ParseFloat(rec[1], 64)
		if err != nil {
			return err
		}
		ftpData = append(ftpData, FTPRecord{Date: date, FTP: ftp})
	}
	c.FTPRecords = ftpData
	return nil
}

func (c *Client) GetFTP(activityDate time.Time) float64 {
	var ftp float64 = 150.0 // default if not found
	for _, rec := range c.FTPRecords {
		if activityDate.Before(rec.Date) {
			break
		}
		ftp = rec.FTP
	}
	return ftp
}
