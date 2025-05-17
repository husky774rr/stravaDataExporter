package scheduler
import (
	"log"
	"time"
	"stravaDataExporter/internal/strava"
)
func StartHourlyFetch(c *strava.Client) {
	go func() {
		for {
			log.Println("Fetching data from Strava...")
			time.Sleep(time.Hour)
		}
	}()
}
func StartDailyTokenRefresh(c *strava.Client) {
	go func() {
		for {
			log.Println("Refreshing access token...")
			time.Sleep(24 * time.Hour)
		}
	}()
}
