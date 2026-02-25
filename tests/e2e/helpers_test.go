package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	activitytypes "seocheon/x/activity/types"
	nodetypes "seocheon/x/node/types"
)

// authtypes_ModuleAddress returns the bech32 address for a module account name.
func authtypes_ModuleAddress(moduleName string) string {
	return authtypes.NewModuleAddress(moduleName).String()
}

// broadcastTx builds, signs, and broadcasts a transaction using the provided client context.
// It waits for the next block after broadcast.
func (s *E2ESuite) broadcastTx(clientCtx client.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	txf := tx.Factory{}.
		WithChainID(s.cfg.ChainID).
		WithTxConfig(s.cfg.TxConfig).
		WithKeybase(clientCtx.Keyring).
		WithAccountRetriever(s.cfg.AccountRetriever).
		WithGas(500000).
		WithGasAdjustment(1.5).
		WithFees("0usum")

	// Set account number and sequence.
	txf, err := txf.Prepare(clientCtx)
	if err != nil {
		return nil, fmt.Errorf("prepare tx: %w", err)
	}

	txBuilder, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, fmt.Errorf("build unsigned tx: %w", err)
	}

	err = tx.Sign(context.Background(), txf, clientCtx.FromName, txBuilder, true)
	if err != nil {
		return nil, fmt.Errorf("sign tx: %w", err)
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("encode tx: %w", err)
	}

	resp, err := clientCtx.BroadcastTxSync(txBytes)
	if err != nil {
		return nil, fmt.Errorf("broadcast tx: %w", err)
	}

	// CheckTx failure (ante handler).
	if resp.Code != 0 {
		return resp, nil
	}

	// Wait for the tx to be included in a block.
	if err := s.network.WaitForNextBlock(); err != nil {
		return resp, fmt.Errorf("wait for block: %w", err)
	}

	// Query the full tx result to capture DeliverTx errors.
	txResult, err := authtx.QueryTx(clientCtx, resp.TxHash)
	if err != nil {
		// If query fails, return original response (tx may not be indexed yet).
		return resp, nil
	}

	return txResult, nil
}

// addKeyToKeyring creates a new key in the validator's keyring and returns
// the address and public key.
func (s *E2ESuite) addKeyToKeyring(val *network.Validator, keyName string) (sdk.AccAddress, cryptotypes.PubKey, error) {
	kr := val.ClientCtx.Keyring
	info, _, err := kr.NewMnemonic(keyName, keyring.English, sdk.GetConfig().GetFullBIP44Path(), keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	if err != nil {
		return nil, nil, err
	}
	pubKey, err := info.GetPubKey()
	if err != nil {
		return nil, nil, err
	}
	addr, err := info.GetAddress()
	if err != nil {
		return nil, nil, err
	}
	return addr, pubKey, nil
}

// fundAccount sends coins from the validator to the target address.
func (s *E2ESuite) fundAccount(val *network.Validator, addr sdk.AccAddress, coins sdk.Coins) error {
	msg := banktypes.NewMsgSend(val.Address, addr, coins)
	clientCtx := val.ClientCtx.
		WithFromAddress(val.Address).
		WithFromName("node0")

	resp, err := s.broadcastTx(clientCtx, msg)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("fund account failed: code=%d log=%s", resp.Code, resp.RawLog)
	}
	return nil
}

// registerNode sends a MsgRegisterNode from the operator, using agentAddr for the agent.
// Returns the node ID.
// Note: pubKey param is ignored; a fresh ed25519 key is generated for the validator consensus key
// because Cosmos SDK staking requires ed25519, while keyring generates secp256k1.
func (s *E2ESuite) registerNode(
	val *network.Validator,
	operatorClientCtx client.Context,
	operatorAddr sdk.AccAddress,
	agentAddr string,
	pubKey cryptotypes.PubKey,
	description string,
	agentShare sdkmath.LegacyDec,
) (string, error) {
	// Generate a fresh ed25519 key for validator consensus (staking requires ed25519).
	consensusPubKey := ed25519.GenPrivKey().PubKey()
	anyPubKey, err := codectypes.NewAnyWithValue(consensusPubKey)
	if err != nil {
		return "", fmt.Errorf("pack pubkey: %w", err)
	}

	msg := &nodetypes.MsgRegisterNode{
		Operator:        operatorAddr.String(),
		AgentAddress:    agentAddr,
		ConsensusPubkey: anyPubKey,
		Description:     description,
		AgentShare:      agentShare,
	}

	resp, err := s.broadcastTx(operatorClientCtx, msg)
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("register node failed: code=%d log=%s", resp.Code, resp.RawLog)
	}

	// Derive node ID deterministically (same algorithm as x/node/keeper).
	hash := sha256.Sum256([]byte("seocheon-node:" + operatorAddr.String()))
	nodeID := hex.EncodeToString(hash[:16])
	return nodeID, nil
}

// submitActivity sends a MsgSubmitActivity from the agent address.
func (s *E2ESuite) submitActivity(
	clientCtx client.Context,
	submitter string,
	hash string,
	contentURI string,
) error {
	msg := &activitytypes.MsgSubmitActivity{
		Submitter:    submitter,
		ActivityHash: hash,
		ContentUri:   contentURI,
	}

	resp, err := s.broadcastTx(clientCtx, msg)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("submit activity failed: code=%d log=%s", resp.Code, resp.RawLog)
	}
	return nil
}

// generateHash generates a deterministic 64-char hex hash from a seed string.
func generateHash(seed string) string {
	h := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(h[:])
}

// queryBalance queries the usum balance for the given address via the first validator.
func (s *E2ESuite) queryBalance(addr sdk.AccAddress) sdk.Coin {
	val := s.network.Validators[0]
	queryClient := banktypes.NewQueryClient(val.ClientCtx)
	resp, err := queryClient.Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: addr.String(),
		Denom:   "usum",
	})
	s.Require().NoError(err)
	return *resp.Balance
}

// waitForHeight waits until the network reaches the given block height.
func (s *E2ESuite) waitForHeight(h int64) {
	_, err := s.network.WaitForHeightWithTimeout(h, 5*time.Minute)
	s.Require().NoError(err, "timeout waiting for height %d", h)
}

// currentHeight returns the current block height.
func (s *E2ESuite) currentHeight() int64 {
	h, err := s.network.LatestHeight()
	s.Require().NoError(err)
	return h
}

// clientCtxForKey returns a client.Context configured to sign with the given key name
// from the given validator's keyring.
func (s *E2ESuite) clientCtxForKey(val *network.Validator, keyName string) client.Context {
	kr := val.ClientCtx.Keyring
	info, err := kr.Key(keyName)
	s.Require().NoError(err)
	addr, err := info.GetAddress()
	s.Require().NoError(err)

	return val.ClientCtx.
		WithFromAddress(addr).
		WithFromName(keyName).
		WithBroadcastMode(flags.BroadcastSync)
}

// queryActivityParams queries the x/activity module parameters.
func (s *E2ESuite) queryActivityParams() activitytypes.Params {
	val := s.network.Validators[0]
	qc := activitytypes.NewQueryClient(val.ClientCtx)
	resp, err := qc.Params(context.Background(), &activitytypes.QueryParamsRequest{})
	s.Require().NoError(err)
	return resp.Params
}

// queryNodeByOperator queries a node by its operator address.
func (s *E2ESuite) queryNodeByOperator(operatorAddr string) nodetypes.Node {
	val := s.network.Validators[0]
	qc := nodetypes.NewQueryClient(val.ClientCtx)
	resp, err := qc.NodeByOperator(context.Background(), &nodetypes.QueryNodeByOperatorRequest{
		Operator: operatorAddr,
	})
	s.Require().NoError(err)
	return resp.GetNode()
}

// queryActivitiesByNode queries activity records for a given node ID.
func (s *E2ESuite) queryActivitiesByNode(nodeID string) []activitytypes.ActivityRecord {
	val := s.network.Validators[0]
	qc := activitytypes.NewQueryClient(val.ClientCtx)
	resp, err := qc.ActivitiesByNode(context.Background(), &activitytypes.QueryActivitiesByNodeRequest{
		NodeId: nodeID,
	})
	s.Require().NoError(err)
	return resp.GetActivities()
}

// queryGovProposal queries a governance proposal by ID.
func (s *E2ESuite) queryGovProposal(proposalID uint64) *govv1.Proposal {
	val := s.network.Validators[0]
	qc := govv1.NewQueryClient(val.ClientCtx)
	resp, err := qc.Proposal(context.Background(), &govv1.QueryProposalRequest{
		ProposalId: proposalID,
	})
	s.Require().NoError(err)
	return resp.GetProposal()
}
