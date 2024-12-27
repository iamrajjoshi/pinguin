package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/iamrajjoshi/pinguin/internal/check"
	"github.com/iamrajjoshi/pinguin/internal/monitor"
	"github.com/redis/go-redis/v9"
)

type Scheduler struct {
	rdb            *redis.Client
	monitorService *monitor.PostgresMonitorService
	checkService   *check.PostgresCheckService

	monitorSchedule  string // Sorted set for scheduling
	monitorWorkQueue string // List for worker queue
}

const monitorScheduleRedisKey = "monitor:scheduled"
const monitorWorkQueueRedisKey = "monitor:queue"

func NewScheduler(rdb *redis.Client, monitorService *monitor.PostgresMonitorService, checkService *check.PostgresCheckService) *Scheduler {
	return &Scheduler{
		rdb:              rdb,
		monitorService:   monitorService,
		checkService:     checkService,
		monitorSchedule:  monitorScheduleRedisKey,
		monitorWorkQueue: monitorWorkQueueRedisKey,
	}
}

// Schedule adds or updates a monitor's next check time
func (s *Scheduler) Schedule(ctx context.Context, monitorID uuid.UUID, interval time.Duration) error {
	// Calculate offset to distribute checks (to avoid thundering herd problem)
	offset := func(monitorID string, interval int) int {
		sum := 0
		for _, b := range []byte(monitorID) {
			sum += int(b)
		}
		return sum % interval
	}

	// Next run time is next interval plus offset
	now := time.Now()
	nextRun := now.Add(interval).Truncate(interval)
	nextRunPlusOffset := nextRun.Add(time.Duration(offset(monitorID.String(), int(interval.Seconds()))) * time.Second)

	member := redis.Z{
		Member: monitorID,
		Score:  float64(nextRunPlusOffset.Unix()),
	}

	return s.rdb.ZAdd(ctx, s.monitorSchedule, member).Err()
}

func (s *Scheduler) ScheduleOnStartup(ctx context.Context) error {
	// Get all enabled monitors
	monitors, err := s.monitorService.GetGeneric(ctx, "enabled = true")
	if err == nil {
		for _, monitor := range monitors {
			lastCheck, err := s.checkService.GetLastCheck(ctx, monitor.ID)
			if err != nil {
				log.Println("Error getting last check:", err)
				continue
			}
			lastCheckTime := lastCheck.Time
			interval := time.Duration(monitor.Interval) * time.Second
			// If last check time is more than interval ago, schedule immediately
			if lastCheckTime.Add(interval).Before(time.Now()) {
				s.Schedule(ctx, monitor.ID, interval)
			} else {
				// Schedule for next interval
				s.Schedule(ctx, monitor.ID, interval)
			}
		}
	}
	return err
}

func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Get all jobs that need to run now
			now := float64(time.Now().Unix())
			monitorIDs, err := s.rdb.ZRangeByScore(ctx, s.monitorSchedule, &redis.ZRangeBy{
				Min: "-inf",
				Max: fmt.Sprintf("%f", now),
			}).Result()

			if err != nil {
				log.Println("Error getting scheduled monitors:", err)
				continue
			}

			// Get all monitors
			// TODO: Cache this?
			monitors, err := s.monitorService.GetManyWithStrings(ctx, monitorIDs)
			if err != nil {
				log.Println("Error getting monitors:", err)
				continue
			}

			for _, monitor := range monitors {
				// Queue just the monitor ID
				if _, err := s.rdb.LPush(ctx, s.monitorWorkQueue, monitor.ID.String()).Result(); err != nil {
					log.Println("Error queuing monitor:", err)
					continue
				}

				// Reschedule next run
				s.Schedule(ctx, monitor.ID, time.Duration(monitor.Interval)*time.Second)
			}
		}
	}
}
