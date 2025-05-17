package influxdb

import (
	"stravaDataExporter/internal/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func NewClient(cfg *config.Config) (influxdb2.Client, error) {
	client := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	return client, nil
}
