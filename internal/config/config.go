package config
import "os"
type Config struct {
	StravaClientID     string
	StravaClientSecret string
	InfluxDBURL        string
	InfluxDBToken      string
	InfluxDBOrg        string
	InfluxDBBucket     string
}
func Load() *Config {
	return &Config{
		StravaClientID:     os.Getenv("STRAVA_CLIENT_ID"),
		StravaClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
		InfluxDBURL:        os.Getenv("INFLUXDB_URL"),
		InfluxDBToken:      os.Getenv("INFLUXDB_TOKEN"),
		InfluxDBOrg:        os.Getenv("INFLUXDB_ORG"),
		InfluxDBBucket:     os.Getenv("INFLUXDB_BUCKET"),
	}
}
