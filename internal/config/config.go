package config

import (
	"os"
	"path"
)

type Config struct {
	StravaClientID     string
	StravaClientSecret string
	InfluxDBURL        string
	InfluxDBToken      string
	InfluxDBOrg        string
	InfluxDBBucket     string
	ProjectRootPath    string
	FtpFileRelPath     string
	FtpFileAbsPath     string
}

func Load() *Config {
	config := &Config{
		StravaClientID:     os.Getenv("STRAVA_CLIENT_ID"),
		StravaClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
		InfluxDBURL:        os.Getenv("INFLUXDB_URL"),
		InfluxDBToken:      os.Getenv("INFLUXDB_TOKEN"),
		InfluxDBOrg:        os.Getenv("INFLUXDB_ORG"),
		InfluxDBBucket:     os.Getenv("INFLUXDB_BUCKET"),
		ProjectRootPath:    os.Getenv("PROJECT_ROOT_PATH"),
		FtpFileRelPath:     os.Getenv("FTP_FILE_REL_PATH"),
	}
	config.FtpFileAbsPath = path.Join(config.ProjectRootPath, config.FtpFileRelPath)
	return config
}
