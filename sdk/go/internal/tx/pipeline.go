package tx

import (
	"context"
	"fmt"
	"time"
)

// PipelineConfig holds configuration for the TX pipeline.
type PipelineConfig struct {
	ChainID         string
	GasPrice        uint64
	ConfirmTimeout  time.Duration
	PollInterval    time.Duration
}

// DefaultPipelineConfig returns a PipelineConfig with sensible defaults.
func DefaultPipelineConfig(chainID string) PipelineConfig {
	return PipelineConfig{
		ChainID:        chainID,
		GasPrice:       DefaultGasPrice,
		ConfirmTimeout: DefaultConfirmTimeout,
		PollInterval:   DefaultPollInterval,
	}
}

// ExecuteTx executes the full 4-phase TX pipeline:
//   Phase 1 - Assembly: query account, build TxBody + AuthInfo + SignDoc
//   Phase 2 - Signing: sign the SignDoc
//   Phase 3 - Broadcast: encode TxRaw and broadcast
//   Phase 4 - Confirmation: poll for TX inclusion
func ExecuteTx(ctx context.Context, querier ChainQuerier, signer Signer, cfg PipelineConfig, req TxRequest) (*TxResult, error) {
	// Phase 1: Assembly
	address := signer.GetAddress()
	pubKey, err := signer.GetPubKey()
	if err != nil {
		return nil, fmt.Errorf("getting public key: %w", err)
	}

	accountNumber, sequence, err := querier.GetAccountInfo(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("getting account info for %s: %w", address, err)
	}

	// Determine gas limit
	gasLimit := req.GasLimit
	if gasLimit == 0 {
		gasLimit = DefaultGasForMessage(req.Message.TypeURL())
	}

	// Determine fee
	feeAmount := req.FeeAmount
	if feeAmount == 0 {
		gasPrice := cfg.GasPrice
		if gasPrice == 0 {
			gasPrice = DefaultGasPrice
		}
		feeAmount = CalculateFee(gasLimit, gasPrice)
	}

	feeDenom := req.FeeDenom
	if feeDenom == "" {
		feeDenom = DefaultFeeDenom
	}

	// Encode TxBody
	bodyBytes := EncodeTxBody([]MessageEncoder{req.Message}, req.Memo, req.TimeoutHeight)

	// Encode AuthInfo
	feeCoins := []Coin{{Denom: feeDenom, Amount: fmt.Sprintf("%d", feeAmount)}}
	authInfoBytes := EncodeAuthInfo(pubKey, sequence, feeCoins, gasLimit)

	// Encode SignDoc
	signDocBytes := EncodeSignDoc(bodyBytes, authInfoBytes, cfg.ChainID, accountNumber)

	// Phase 2: Signing
	signature, err := signer.Sign(signDocBytes)
	if err != nil {
		return nil, fmt.Errorf("signing transaction: %w", err)
	}

	// Phase 3: Broadcast
	txRawBytes := EncodeTxRaw(bodyBytes, authInfoBytes, signature)

	txHash, code, rawLog, err := querier.BroadcastTxSync(ctx, txRawBytes)
	if err != nil {
		return nil, fmt.Errorf("broadcasting transaction: %w", err)
	}

	// Check broadcast result code
	if code != 0 {
		return &TxResult{
			TxHash: txHash,
			Code:   code,
			RawLog: rawLog,
		}, fmt.Errorf("broadcast failed with code %d: %s", code, rawLog)
	}

	// Phase 4: Confirmation
	result, err := PollTxConfirmation(ctx, querier, txHash, cfg.ConfirmTimeout, cfg.PollInterval)
	if err != nil {
		// Return partial result with TX hash even on timeout
		return &TxResult{
			TxHash: txHash,
		}, fmt.Errorf("confirming transaction: %w", err)
	}

	return result, nil
}
