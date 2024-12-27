package check

import (
	"context"

	"github.com/google/uuid"
	store "github.com/iamrajjoshi/pinguin/internal/store/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckService interface {
	Create(ctx context.Context, check *store.Check) error
	GetLastCheck(ctx context.Context, monitorID uuid.UUID) (store.Check, error)
}

type PostgresCheckService struct {
	db *pgxpool.Pool
}

func NewCheckService(db *pgxpool.Pool) *PostgresCheckService {
	return &PostgresCheckService{db: db}
}

func (s *PostgresCheckService) Create(ctx context.Context, check *store.Check) error {
	query := `
		INSERT INTO checks (time, monitor_id, duration_ms, success, status_code, headers, body, body_size)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.db.Exec(ctx, query, check.Time, check.MonitorID, check.DurationMS, check.Success, check.StatusCode, check.Headers, check.Body, check.BodySize)
	return err
}

func (s *PostgresCheckService) GetLastCheck(ctx context.Context, monitorID uuid.UUID) (store.Check, error) {
	query := `SELECT * FROM checks WHERE monitor_id = $1 ORDER BY time DESC LIMIT 1`
	var lastCheck store.Check
	err := s.db.QueryRow(ctx, query, monitorID).Scan(&lastCheck.Time, &lastCheck.MonitorID, &lastCheck.DurationMS, &lastCheck.Success, &lastCheck.StatusCode, &lastCheck.Headers, &lastCheck.Body, &lastCheck.BodySize)
	return lastCheck, err
}
