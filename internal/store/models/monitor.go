package store

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Monitor struct {
	ID        uuid.UUID `json:"id" db:"id"`
	URL       string    `json:"url" db:"url"`
	Name      string    `json:"name" db:"name"`
	Interval  int       `json:"interval" db:"interval"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (m *Monitor) Validate() error {
	if m.URL == "" {
		return errors.New("url is required")
	}
	return nil
}
