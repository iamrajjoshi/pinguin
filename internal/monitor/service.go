package monitor

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iamrajjoshi/pinguin/internal/errors"
	store "github.com/iamrajjoshi/pinguin/internal/store/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MonitorService interface {
	Create(ctx context.Context, monitor *store.Monitor) error
	Get(ctx context.Context, id uuid.UUID) (*store.Monitor, error)
	List(ctx context.Context) ([]store.Monitor, error)
	Update(ctx context.Context, monitor *store.Monitor) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresMonitorService struct {
	db *pgxpool.Pool
}

func NewMonitorService(db *pgxpool.Pool) *PostgresMonitorService {
	return &PostgresMonitorService{db: db}
}

func (s *PostgresMonitorService) Create(ctx context.Context, monitor *store.Monitor) error {
	query := `
		INSERT INTO monitors (url, name, interval, enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return s.db.QueryRow(ctx, query,
		monitor.URL,
		monitor.Name,
		monitor.Interval,
		monitor.Enabled,
	).Scan(&monitor.ID, &monitor.CreatedAt, &monitor.UpdatedAt)
}

func (s *PostgresMonitorService) Get(ctx context.Context, id uuid.UUID) (*store.Monitor, error) {
	monitor := &store.Monitor{}
	query := `SELECT * FROM monitors WHERE id = $1`

	err := s.db.QueryRow(ctx, query, id).Scan(
		&monitor.ID,
		&monitor.URL,
		&monitor.Name,
		&monitor.Interval,
		&monitor.Enabled,
		&monitor.CreatedAt,
		&monitor.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return monitor, nil
}

func (s *PostgresMonitorService) List(ctx context.Context) ([]store.Monitor, error) {
	var monitors []store.Monitor
	query := `SELECT * FROM monitors ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var monitor store.Monitor
		if err := rows.Scan(&monitor.ID, &monitor.URL, &monitor.Name, &monitor.Interval, &monitor.Enabled, &monitor.CreatedAt, &monitor.UpdatedAt); err != nil {
			return nil, err
		}
		monitors = append(monitors, monitor)
	}

	return monitors, nil
}

func (s *PostgresMonitorService) Update(ctx context.Context, monitor *store.Monitor) error {
	query := `
		UPDATE monitors 
		SET url = $1, name = $2, interval = $3, enabled = $4, updated_at = $5
		WHERE id = $6`

	result, err := s.db.Exec(ctx, query,
		monitor.URL,
		monitor.Name,
		monitor.Interval,
		monitor.Enabled,
		time.Now(),
		monitor.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (s *PostgresMonitorService) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM monitors WHERE id = $1`

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}
