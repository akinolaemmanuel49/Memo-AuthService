package dtos

import "time"

type DatabaseMetrics struct {
	TotalConnections    int32 `json:"totalConnections"`
	AcquiredConnections int32 `json:"acquiredConnections"`
	IdleConnections     int32 `json:"idleConnections"`
}

type HealthCheck struct {
	Status   string          `json:"status"`
	Database DatabaseMetrics `json:"database"`
	Time     time.Time       `json:"time"`
}
