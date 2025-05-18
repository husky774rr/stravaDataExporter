package influxdb

import (
	"context"
	"stravaDataExporter/internal/config"
	"stravaDataExporter/internal/model"
	"strconv"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func NewClient(cfg *config.Config) (influxdb2.Client, error) {
	client := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	return client, nil
}

func SaveActivity(c *influxdb2.Client, activities []model.Activity, org, bucket string) []error {
	errors := make([]error, 0)
	writeAPI := (*c).WriteAPIBlocking(org, bucket)
	for _, activity := range activities {
		p := influxdb2.NewPointWithMeasurement("activity").
			AddTag("id", strconv.FormatInt(activity.ID, 10)).
			AddField("start_time", activity.StartTime).
			AddField("duration", activity.Duration).
			AddField("distance", activity.Distance).
			AddField("elevation", activity.Elevation).
			AddField("average_watt", activity.AverageWatt).
			AddField("ftp", activity.FTP).
			AddField("tss", activity.TSS).
			AddField("np", activity.NP)
		if err := writeAPI.WritePoint(context.Background(), p); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
