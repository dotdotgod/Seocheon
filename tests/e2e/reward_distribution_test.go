package e2e_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestRewardDistribution tests the dual reward pool pipeline:
// 1. Register two nodes (A: eligible, B: ineligible)
// 2. Submit activities: A in 8+ windows, B in <8 windows
// 3. Wait for epoch boundary
// 4. Verify that only eligible node A receives activity rewards.
func (s *E2ESuite) TestRewardDistribution() {
	val := s.network.Validators[0]

	// --- Create Node A (will be eligible) ---
	opAddrA, opPubA, err := s.addKeyToKeyring(val, "opA")
	s.Require().NoError(err)
	agAddrA, _, err := s.addKeyToKeyring(val, "agA")
	s.Require().NoError(err)

	err = s.fundAccount(val, opAddrA, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(10_000_000))))
	s.Require().NoError(err)
	err = s.fundAccount(val, agAddrA, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(1_000_000))))
	s.Require().NoError(err)

	opCtxA := s.clientCtxForKey(val, "opA")
	nodeIDA, err := s.registerNode(val, opCtxA, opAddrA, agAddrA.String(), opPubA, "node-A", sdkmath.LegacyNewDec(30))
	s.Require().NoError(err)
	s.T().Logf("registered node A: %s (agent_share=30%%)", nodeIDA)

	// --- Create Node B (will be ineligible) ---
	opAddrB, opPubB, err := s.addKeyToKeyring(val, "opB")
	s.Require().NoError(err)
	agAddrB, _, err := s.addKeyToKeyring(val, "agB")
	s.Require().NoError(err)

	err = s.fundAccount(val, opAddrB, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(10_000_000))))
	s.Require().NoError(err)
	err = s.fundAccount(val, agAddrB, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(1_000_000))))
	s.Require().NoError(err)

	opCtxB := s.clientCtxForKey(val, "opB")
	nodeIDB, err := s.registerNode(val, opCtxB, opAddrB, agAddrB.String(), opPubB, "node-B", sdkmath.LegacyZeroDec())
	s.Require().NoError(err)
	s.T().Logf("registered node B: %s (agent_share=0%%)", nodeIDB)

	// --- Submit activities ---
	agCtxA := s.clientCtxForKey(val, "agA")
	agCtxB := s.clientCtxForKey(val, "agB")

	startHeight := s.currentHeight()
	windowLength := int64(10) // 120 / 12

	// Node A: submit in 8 windows (eligible).
	for w := 0; w < 8; w++ {
		targetHeight := startHeight + int64(w)*windowLength
		if targetHeight > s.currentHeight() {
			s.waitForHeight(targetHeight)
		}

		hash := generateHash(fmt.Sprintf("rewardA-%d", w))
		err := s.submitActivity(agCtxA, agAddrA.String(), hash, fmt.Sprintf("ipfs://a-%d", w))
		s.Require().NoError(err, "node A activity window %d", w)
	}

	// Node B: submit in only 4 windows (ineligible, needs 8).
	for w := 0; w < 4; w++ {
		targetHeight := startHeight + int64(w)*windowLength
		if targetHeight > s.currentHeight() {
			s.waitForHeight(targetHeight)
		}

		hash := generateHash(fmt.Sprintf("rewardB-%d", w))
		err := s.submitActivity(agCtxB, agAddrB.String(), hash, fmt.Sprintf("ipfs://b-%d", w))
		s.Require().NoError(err, "node B activity window %d", w)
	}

	// --- Record balances before epoch boundary ---
	balOpA_before := s.queryBalance(opAddrA)
	balAgA_before := s.queryBalance(agAddrA)
	balOpB_before := s.queryBalance(opAddrB)
	balAgB_before := s.queryBalance(agAddrB)

	s.T().Logf("pre-epoch balances: opA=%s agA=%s opB=%s agB=%s",
		balOpA_before, balAgA_before, balOpB_before, balAgB_before)

	// --- Wait for epoch boundary ---
	currentH := s.currentHeight()
	epochLength := int64(120)
	epochBoundary := ((currentH / epochLength) + 1) * epochLength
	s.T().Logf("waiting for epoch boundary at %d", epochBoundary)
	s.waitForHeight(epochBoundary + 2) // +2 for safety

	// --- Verify balances after epoch ---
	balOpA_after := s.queryBalance(opAddrA)
	balAgA_after := s.queryBalance(agAddrA)
	balOpB_after := s.queryBalance(opAddrB)
	balAgB_after := s.queryBalance(agAddrB)

	s.T().Logf("post-epoch balances: opA=%s agA=%s opB=%s agB=%s",
		balOpA_after, balAgA_after, balOpB_after, balAgB_after)

	// Node A should have received rewards (operator and/or agent portion).
	// At minimum, the combined balance should have increased.
	totalA_before := balOpA_before.Amount.Add(balAgA_before.Amount)
	totalA_after := balOpA_after.Amount.Add(balAgA_after.Amount)
	s.T().Logf("node A total change: %s -> %s (delta=%s)",
		totalA_before, totalA_after, totalA_after.Sub(totalA_before))

	// Node B should NOT have received activity rewards (ineligible).
	// Their balance may change from staking rewards but not from activity rewards.
	totalB_before := balOpB_before.Amount.Add(balAgB_before.Amount)
	totalB_after := balOpB_after.Amount.Add(balAgB_after.Amount)
	deltaB := totalB_after.Sub(totalB_before)
	s.T().Logf("node B total change: %s -> %s (delta=%s)",
		totalB_before, totalB_after, deltaB)

	// --- Assertions ---
	// Activity reward pool was funded with 1,000,000 usum at genesis.
	// Only Node A (1 eligible node) receives the full pool.
	// Agent share = 30%: agent gets 300,000, operator gets 700,000.
	// Using GTE to account for possible mint inflation dust.
	deltaOpA := balOpA_after.Amount.Sub(balOpA_before.Amount)
	deltaAgA := balAgA_after.Amount.Sub(balAgA_before.Amount)
	deltaA := totalA_after.Sub(totalA_before)

	s.T().Logf("node A deltas: operator=%s, agent=%s, total=%s", deltaOpA, deltaAgA, deltaA)

	s.Require().True(deltaOpA.GTE(sdkmath.NewInt(700_000)),
		"operator A should receive >= 700,000 usum (70%% of pool), got %s", deltaOpA)
	s.Require().True(deltaAgA.GTE(sdkmath.NewInt(300_000)),
		"agent A should receive >= 300,000 usum (30%% of pool), got %s", deltaAgA)
	s.Require().True(deltaA.GTE(sdkmath.NewInt(1_000_000)),
		"node A total should receive >= 1,000,000 usum (full pool), got %s", deltaA)

	// Node B is ineligible: should receive no activity rewards.
	// Only possible delta is staking reward dust, which should be small.
	s.Require().True(deltaB.LT(sdkmath.NewInt(100_000)),
		"node B total delta should be < 100,000 usum (no activity rewards), got %s", deltaB)

	s.T().Log("reward distribution test completed: assertions passed")
}
