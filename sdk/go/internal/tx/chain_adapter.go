package tx

import (
	"context"
	"fmt"

	"github.com/seocheon/sdk-go/internal/chain"
)

// ChainClientAdapter adapts chain.Client to the ChainQuerier interface.
type ChainClientAdapter struct {
	client chain.Client
}

// NewChainClientAdapter creates a new adapter wrapping a chain.Client.
func NewChainClientAdapter(client chain.Client) *ChainClientAdapter {
	return &ChainClientAdapter{client: client}
}

func (a *ChainClientAdapter) GetAccountInfo(ctx context.Context, address string) (uint64, uint64, error) {
	info, err := a.client.GetAccountInfo(ctx, address)
	if err != nil {
		return 0, 0, fmt.Errorf("querying account info: %w", err)
	}
	return info.AccountNumber, info.Sequence, nil
}

func (a *ChainClientAdapter) BroadcastTxSync(ctx context.Context, txBytes []byte) (string, uint32, string, error) {
	// The chain client expects raw bytes; it handles base64 encoding internally
	resp, err := a.client.BroadcastTx(ctx, txBytes, "sync")
	if err != nil {
		return "", 0, "", fmt.Errorf("broadcasting tx: %w", err)
	}
	return resp.TxHash, resp.Code, resp.RawLog, nil
}

func (a *ChainClientAdapter) GetTxResult(ctx context.Context, txHash string) (*TxResult, error) {
	resp, err := a.client.GetTx(ctx, txHash)
	if err != nil {
		return nil, err
	}

	events := make([]TxEvent, len(resp.Events))
	for i, e := range resp.Events {
		attrs := make([]EventAttribute, len(e.Attributes))
		for j, attr := range e.Attributes {
			attrs[j] = EventAttribute{Key: attr.Key, Value: attr.Value}
		}
		events[i] = TxEvent{Type: e.Type, Attributes: attrs}
	}

	return &TxResult{
		TxHash:    resp.TxHash,
		Height:    resp.Height,
		Code:      resp.Code,
		GasUsed:   resp.GasUsed,
		GasWanted: resp.GasWanted,
		RawLog:    resp.RawLog,
		Events:    events,
	}, nil
}

