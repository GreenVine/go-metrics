package database

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"gorm.io/gorm"
)

const maxAlertGenRetries = 100
const maxMetricsPerBatch = 1000

var alertGen *AlertGenerator

// AlertGenerator handles the batched processing of metrics for alerts.
type AlertGenerator struct {
	sync.Mutex
	ctx context.Context

	metricQueue []*MetricRecord
}

// InitAlertGenerator creates a new instance.
func InitAlertGenerator(ctx context.Context, interval time.Duration) {
	alertGen = &AlertGenerator{
		ctx:         ctx,
		metricQueue: make([]*MetricRecord, 0),
	}

	go alertGen.Start(interval)
}

// Start begins the background goroutine that processes metrics for alerts.
func (a *AlertGenerator) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.processMetricsForAlerts()
		case <-a.ctx.Done():
			log.Println("Shutting down alert generator...")

			// Process any remaining metrics before shutdown.
			a.processMetricsForAlerts()
			return
		}
	}
}

// QueueMetric adds a metric to the processing queue.
func (a *AlertGenerator) QueueMetric(metric *MetricRecord) {
	a.Lock()
	defer a.Unlock()

	a.metricQueue = append(a.metricQueue, metric)
}

// processMetricsForAlerts processes queued metrics with a cap per attempt.
func (a *AlertGenerator) processMetricsForAlerts() {
	a.Lock()
	if len(a.metricQueue) < 1 {
		a.Unlock()
		return
	}

	// Determine how many metrics to process within this attempt
	batchSize := min(len(a.metricQueue), maxMetricsPerBatch)

	// Take metrics from the head of the queue.
	currentBatch := a.metricQueue[:batchSize]
	a.metricQueue = a.metricQueue[batchSize:]
	remaining := len(a.metricQueue)
	a.Unlock()

	// Process all metrics in a single transaction.
	ctx, cancel := context.WithTimeout(a.ctx, 1*time.Minute)
	defer cancel()

	if err := retry.Do(
		func() error {
			return WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				log.Printf(
					"Processing %d metrics for alert generation (remaining = %d).",
					batchSize, remaining,
				)

				return checkAndMaybeCreateAlerts(tx, currentBatch)
			})
		},
		retry.Context(ctx),
		retry.Attempts(maxAlertGenRetries),
		retry.LastErrorOnly(true),
	); err != nil {
		log.Printf("Failed to generate alerts: %v", err)
	}
}
