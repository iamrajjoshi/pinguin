package store

import (
	"time"

	"github.com/google/uuid"
)

type Check struct {
	Time       time.Time      `json:"time" db:"time"`
	MonitorID  uuid.UUID      `json:"monitor_id" db:"monitor_id"`
	DurationMS int            `json:"duration_ms" db:"duration_ms"`
	StatusCode int            `json:"status_code" db:"status_code"`
	Error      *string        `json:"error" db:"error"`
	Success    bool           `json:"success" db:"success"`
	Headers    map[string]any `json:"headers" db:"headers"`
	Body       *string        `json:"body" db:"body"`
	BodySize   *int           `json:"body_size" db:"body_size"`
}
