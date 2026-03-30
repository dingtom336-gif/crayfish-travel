package riskcontrol

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	refundKeyPrefix = "riskcontrol:refund_count:"
	refundKeyTTL    = 31 * 24 * time.Hour // 31 days
	maxRefundsMonth = 3
)

// AntifraudChecker performs risk-control fraud checks using Redis counters.
type AntifraudChecker struct {
	rdb *goredis.Client
}

// NewAntifraudChecker creates a new AntifraudChecker.
func NewAntifraudChecker(rdb *goredis.Client) *AntifraudChecker {
	return &AntifraudChecker{rdb: rdb}
}

// CheckRefundFrequency returns an error if the identity hash has exceeded
// the maximum allowed refunds (3) in the current month.
func (ac *AntifraudChecker) CheckRefundFrequency(ctx context.Context, idHash string) error {
	key := refundKey(idHash)
	count, err := ac.rdb.Get(ctx, key).Int()
	if err != nil && err != goredis.Nil {
		return fmt.Errorf("failed to check refund frequency: %w", err)
	}

	if count >= maxRefundsMonth {
		return fmt.Errorf("refund frequency exceeded: %d refunds this month (max %d)", count, maxRefundsMonth)
	}

	return nil
}

// IncrementRefund increments the refund counter for the identity hash.
func (ac *AntifraudChecker) IncrementRefund(ctx context.Context, idHash string) error {
	key := refundKey(idHash)
	pipe := ac.rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, refundKeyTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to increment refund counter: %w", err)
	}

	return nil
}

// refundKey builds the Redis key for a given identity hash and current month.
func refundKey(idHash string) string {
	month := time.Now().Format("2006-01")
	return fmt.Sprintf("%s%s:%s", refundKeyPrefix, idHash, month)
}
