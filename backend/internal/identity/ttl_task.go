package identity

import (
	"context"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

const (
	// TypeIdentityCleanup is the Asynq task type for PII TTL cleanup.
	TypeIdentityCleanup = "identity:cleanup"
)

// CleanupProcessor handles scheduled PII deletion.
type CleanupProcessor struct {
	repo *Repository
}

// NewCleanupProcessor creates a new cleanup task processor.
func NewCleanupProcessor(repo *Repository) *CleanupProcessor {
	return &CleanupProcessor{repo: repo}
}

// NewCleanupTask creates a new Asynq task for PII cleanup.
func NewCleanupTask() *asynq.Task {
	return asynq.NewTask(TypeIdentityCleanup, nil)
}

// ProcessTask deletes expired identity records.
func (p *CleanupProcessor) ProcessTask(_ context.Context, _ *asynq.Task) error {
	count, err := p.repo.DeleteExpired()
	if err != nil {
		return fmt.Errorf("cleanup expired identities: %w", err)
	}
	if count > 0 {
		log.Printf("[identity:cleanup] deleted %d expired records", count)
	}
	return nil
}
