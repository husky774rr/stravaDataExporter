package model

import "time"

type Activity struct {
	ID          int64
	StartTime   time.Time
	Duration    float64 // 秒
	Distance    float64 // メートル
	Elevation   float64 // 獲得標高
	AverageWatt float64
	FTP         float64
	TSS         float64
	NP          float64
}
