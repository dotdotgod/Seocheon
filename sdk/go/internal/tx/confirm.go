package tx

import (
	"context"
	"fmt"
	"time"
)

// DefaultConfirmTimeout is the default TX confirmation timeout.
const DefaultConfirmTimeout = 30 * time.Second

// DefaultPollInterval is the default polling interval for TX confirmation.
const DefaultPollInterval = 1 * time.Second

// PollTxConfirmation polls for a transaction result until confirmed or timeout.
func PollTxConfirmation(ctx context.Context, querier ChainQuerier, txHash string, timeout, pollInterval time.Duration) (*TxResult, error) {
	if timeout == 0 {
		timeout = DefaultConfirmTimeout
	}
	if pollInterval == 0 {
		pollInterval = DefaultPollInterval
	}

	deadline := time.After(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while waiting for tx %s: %w", txHash, ctx.Err())
		case <-deadline:
			return nil, fmt.Errorf("timeout waiting for tx %s after %v", txHash, timeout)
		case <-ticker.C:
			result, err := querier.GetTxResult(ctx, txHash)
			if err != nil {
				// TX not yet indexed, continue polling
				continue
			}
			return result, nil
		}
	}
}
