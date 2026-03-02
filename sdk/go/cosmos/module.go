// Package cosmos implements the Cosmos module for the Seocheon SDK.
// It provides standard Cosmos chain operations: balance queries, token transfers,
// block info, and transaction result queries.
package cosmos

import (
	"context"
	"encoding/json"
	"fmt"

	sdkerrors "github.com/seocheon/sdk-go/errors"
	"github.com/seocheon/sdk-go/internal/chain"
	"github.com/seocheon/sdk-go/internal/signing"
	"github.com/seocheon/sdk-go/types"
	"github.com/seocheon/sdk-go/utility"
)

// Module provides standard Cosmos operations.
type Module struct {
	client chain.Client
	signer signing.Service
}

// NewModule creates a new Cosmos module.
func NewModule(client chain.Client, signer signing.Service) *Module {
	return &Module{
		client: client,
		signer: signer,
	}
}

// GetBalance returns the token balance for an address.
// If address is empty, queries the signer's own address.
// If denom is empty, defaults to "uppyeo".
func (m *Module) GetBalance(ctx context.Context, address, denom string) (*types.BalanceResponse, error) {
	effectiveAddr := address
	if effectiveAddr == "" {
		effectiveAddr = m.signer.GetAddress()
	}
	effectiveDenom := denom
	if effectiveDenom == "" {
		effectiveDenom = "uppyeo"
	}

	path := fmt.Sprintf("/cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s", effectiveAddr, effectiveDenom)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrQueryFailed, err)
	}

	var result struct {
		Balance struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balance"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing balance response: %w", err)
	}

	balanceUppyeo := parseIntSafe(result.Balance.Amount)

	return &types.BalanceResponse{
		Address:     effectiveAddr,
		Balance:     result.Balance.Amount,
		BalanceKkot: utility.FormatKkot(balanceUppyeo),
	}, nil
}

// SendTokens sends tokens to the specified address.
// Amount should be in the smallest denomination unit (uppyeo by default).
// If denom is empty, defaults to "uppyeo".
func (m *Module) SendTokens(ctx context.Context, toAddress, amount, denom string) (*types.SendTokensResponse, error) {
	if toAddress == "" {
		return nil, sdkerrors.ErrInvalidAddress
	}

	_ = m.signer.GetAddress() // from_address
	if denom == "" {
		denom = "uppyeo"
	}

	// TODO: Implement full TX flow (MsgSend) when proto dependencies are available
	return nil, fmt.Errorf("cosmos.SendTokens: TX flow requires proto message builders (not yet implemented)")
}

// GetBlockInfo returns the latest block information.
func (m *Module) GetBlockInfo(ctx context.Context) (*types.BlockInfoResponse, error) {
	block, err := m.client.GetLatestBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrQueryFailed, err)
	}

	return &types.BlockInfoResponse{
		BlockHeight: block.Height,
		BlockTime:   block.Time,
		ChainID:     block.ChainID,
		NumTxs:      uint64(block.NumTxs),
	}, nil
}

// GetTxResult returns the result of a transaction by its hash.
func (m *Module) GetTxResult(ctx context.Context, txHash string) (*types.TxResultResponse, error) {
	if txHash == "" {
		return nil, sdkerrors.ErrTxNotFound
	}

	tx, err := m.client.GetTx(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrTxNotFound, err)
	}

	events := make([]types.TxEvent, len(tx.Events))
	for i, e := range tx.Events {
		attrs := make([]types.EventAttribute, len(e.Attributes))
		for j, a := range e.Attributes {
			attrs[j] = types.EventAttribute{
				Key:   a.Key,
				Value: a.Value,
			}
		}
		events[i] = types.TxEvent{
			Type:       e.Type,
			Attributes: attrs,
		}
	}

	return &types.TxResultResponse{
		TxHash:    tx.TxHash,
		Height:    tx.Height,
		Code:      tx.Code,
		GasUsed:   tx.GasUsed,
		GasWanted: tx.GasWanted,
		RawLog:    tx.RawLog,
		Events:    events,
	}, nil
}

func parseIntSafe(s string) int64 {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	return result
}
