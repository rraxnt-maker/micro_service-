package cleaner

import (
	"log"
	"time"
	"status-service/internal/storage"
)

type Cleaner struct {
	interval time.Duration
	stopCh   chan struct{}
}

func NewCleaner(interval time.Duration) *Cleaner {
	return &Cleaner{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (c *Cleaner) Start() {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		log.Printf("🧹 Cleaner started, checking every %v", c.interval)

		for {
			select {
			case <-ticker.C:
				c.clean()
			case <-c.stopCh:
				log.Println("🛑 Cleaner stopped")
				return
			}
		}
	}()
}

func (c *Cleaner) Stop() {
	close(c.stopCh)
}

func (c *Cleaner) clean() {
	count, err := storage.DB.DeleteExpiredStatuses()
	if err != nil {
		log.Printf("❌ Failed to clean expired statuses: %v", err)
	} else if count > 0 {
		log.Printf("🧹 Cleaned %d expired status(es)", count)
	}
}