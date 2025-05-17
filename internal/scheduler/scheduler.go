package scheduler

import (
	"time"

	"stravaDataExporter/internal/strava"
)

func StartHourlyFetch(c *strava.Client) {
	go func() {
		for {
			c.Logger.Info("Fetching data from Strava (not implemented)")
			time.Sleep(time.Hour)
		}
	}()
}

func StartDailyTokenRefresh(c *strava.Client) {
	go func() {
		for {
			err := c.RefreshAccessToken()
			if err != nil {
				c.Logger.Error("failed to refresh access token", "error", err)
			} else {
				c.Logger.Info("access token refreshed successfully (scheduled)")
			}
			time.Sleep(24 * time.Hour)
			// time.Sleep(1 * time.Minute) // For testing purposes, refresh every minute
		}
	}()
}
