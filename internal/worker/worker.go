// internal/worker/worker.go
package worker

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/iamrajjoshi/pinguin/internal/check"
	"github.com/iamrajjoshi/pinguin/internal/monitor"
	store "github.com/iamrajjoshi/pinguin/internal/store/models"
)

type Worker struct {
	id             int
	rdb            *redis.Client
	client         *http.Client
	queueKey       string
	monitorService monitor.MonitorService
	checkService   check.CheckService
}

func NewWorker(id int, rdb *redis.Client, monitorService monitor.MonitorService, checkService check.CheckService) *Worker {
	return &Worker{
		id:             id,
		rdb:            rdb,
		client:         &http.Client{Timeout: 30 * time.Second},
		queueKey:       "monitor:queue",
		monitorService: monitorService,
		checkService:   checkService,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Pop job from queue
			result, err := w.rdb.BRPop(ctx, 0, w.queueKey).Result()
			if err != nil {
				log.Println("Error getting job from queue:", err)
				continue
			}

			var strMonitorID string = result[1]

			monitorID, err := uuid.Parse(strMonitorID)
			if err != nil {
				log.Println("Invalid monitor ID:", err)
				continue
			}

			// Get monitor from DB
			monitor, err := w.monitorService.Get(ctx, monitorID)
			if err != nil {
				log.Println("Failed to get monitor:", err)
				continue
			}

			// Perform check
			start := time.Now()
			// TODO: Get URL from DB or maybe cache
			resp, err := w.client.Head(monitor.URL)
			duration := time.Since(start)

			check_obj := store.Check{
				Time:       time.Now(),
				MonitorID:  monitorID,
				DurationMS: int(duration.Milliseconds()),
				Success:    err == nil,
			}

			if err == nil {
				if resp.StatusCode == http.StatusOK {
					bodySize := int(resp.ContentLength)
					resp.Body.Close()

					headers := make(map[string]any)
					for k, v := range resp.Header {
						headers[k] = v[0]
					}

					check_obj.StatusCode = resp.StatusCode
					check_obj.Headers = headers
					check_obj.BodySize = &bodySize
				}
			}

			err = w.checkService.Create(ctx, &check_obj)
			if err != nil {
				log.Println("Failed to store check result:", err)
			}
		}
	}
}
