package service

import (
	"context"
	"time"
)

type DBPinger interface {
	PingContext(ctx context.Context) error
}

type HealthStatus struct {
	Name      string    `json:"name"`
	Env       string    `json:"env"`
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
}

type HealthService struct {
	appName string
	env     string
	db      DBPinger
}

func NewHealthService(appName, env string, db DBPinger) *HealthService {
	return &HealthService{
		appName: appName,
		env:     env,
		db:      db,
	}
}

func (s *HealthService) Status(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Name:      s.appName,
		Env:       s.env,
		Status:    "ok",
		Database:  "unknown",
		Timestamp: time.Now().UTC(),
	}

	if s.db == nil {
		return status
	}

	if err := s.db.PingContext(ctx); err != nil {
		status.Status = "degraded"
		status.Database = "down"
		return status
	}

	status.Database = "up"
	return status
}
