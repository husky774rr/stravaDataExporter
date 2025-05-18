package main

import (
	"log"
	"net/http"
	"stravaDataExporter/internal/config"
	"stravaDataExporter/internal/influxdb"
	"stravaDataExporter/internal/scheduler"
	"stravaDataExporter/internal/strava"
)

func main() {
	cfg := config.Load()
	dbClient, err := influxdb.NewClient(cfg)
	if err != nil {
		log.Fatalf("InfluxDB connection failed: %v", err)
	}
	stravaClient := strava.NewClient(cfg, &dbClient)
	if err := stravaClient.LoadFTPHistoricalData(); err != nil {
		log.Fatalf("Failed to load FTP historical data: %v", err)
	}

	go scheduler.StartRegularlyFetch(stravaClient)
	go scheduler.StartAccessTokenRefresh(stravaClient)

	http.HandleFunc("/auth/callback", strava.HandleOAuthCallback(stravaClient))
	http.HandleFunc("/auth/login", strava.HandleLogin(cfg))
	http.HandleFunc("/", stravaClient.AuthRequiredHandler(nil))
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
