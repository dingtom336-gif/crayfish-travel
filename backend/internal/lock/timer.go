package lock

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const pollInterval = 30 * time.Second

// LockTimer polls for expired locks and auto-releases them.
type LockTimer struct {
	db           interface{ Query(string, ...any) (rows, error) }
	rdb          *goredis.Client
	orchestrator *Orchestrator
}

// rows is a minimal interface for sql.Rows used by LockTimer.
type rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
}

// NewLockTimer creates a new LockTimer.
func NewLockTimer(orchestrator *Orchestrator, rdb *goredis.Client) *LockTimer {
	return &LockTimer{
		orchestrator: orchestrator,
		rdb:          rdb,
	}
}

// StartExpiryWatcher starts a goroutine that polls for expired locks every 30 seconds.
// It blocks until the context is cancelled.
func (lt *LockTimer) StartExpiryWatcher(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Println("lock timer: expiry watcher started")

	for {
		select {
		case <-ctx.Done():
			log.Println("lock timer: expiry watcher stopped")
			return
		case <-ticker.C:
			lt.releaseExpired(ctx)
		}
	}
}

// releaseExpired finds and releases all expired locks.
func (lt *LockTimer) releaseExpired(ctx context.Context) {
	rows, err := lt.orchestrator.db.Query(`
		SELECT id, trace_id FROM lock_sessions
		WHERE state = $1 AND expires_at < NOW()`,
		StateLocked,
	)
	if err != nil {
		log.Printf("lock timer: failed to query expired locks: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var lockID uuid.UUID
		var traceID string
		if err := rows.Scan(&lockID, &traceID); err != nil {
			log.Printf("lock timer: failed to scan row: %v", err)
			continue
		}

		log.Printf("lock timer: [%s] releasing expired lock %s", traceID, lockID)
		if err := lt.orchestrator.ReleaseLock(ctx, lockID, traceID); err != nil {
			log.Printf("lock timer: [%s] failed to release lock %s: %v", traceID, lockID, err)
		}
	}
}
