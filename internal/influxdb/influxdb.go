package influxdb
import (
	"github.com/influxdata/influxdb-client-go/v2"
	"stravaDataExporter/internal/config"
)
func NewClient(cfg *config.Config) (influxdb2.Client, error) {
	client := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	return client, nil
}
