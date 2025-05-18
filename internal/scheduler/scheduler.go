package scheduler

import (
	"time"

	"stravaDataExporter/internal/strava"
)

func StartRegularlyFetch(c *strava.Client) {
	go func() {
		for {
			c.Logger.Info("Waik up and fetch data from Strava. Start StartRegularlyFetch")

			if c.IsAuthenticated() {
				if err := c.FetchActivities(); err != nil {
					c.Logger.Error("failed to fetch activities", "error", err)
				} else {
					c.Logger.Info("activities fetched successfully")
				}
			} else {
				c.Logger.Info("client is not authenticated, skipping fetch")
			}

			c.Logger.Info("Sleep for next Fetching. Finish StartRegularlyFetch")
			// time.Sleep(time.Hour)
			time.Sleep(time.Minute)
		}
	}()
}

func StartAccessTokenRefresh(c *strava.Client) {
	go func() {
		for {
			err := c.RefreshAccessToken()
			if err != nil {
				c.Logger.Error("failed to refresh access token", "error", err)
			} else {
				c.Logger.Info("access token refreshed successfully (scheduled)")
			}
			time.Sleep(24 * time.Hour)
		}
	}()
}
