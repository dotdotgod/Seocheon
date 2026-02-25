package e2e_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestNodeActivityFlow tests the full cross-keeper flow:
// account creation -> node registration -> activity submission across windows -> epoch boundary.
func (s *E2ESuite) TestNodeActivityFlow() {
	val := s.network.Validators[0]

	// 1. Create operator and agent keys.
	operatorAddr, operatorPubKey, err := s.addKeyToKeyring(val, "operator1")
	s.Require().NoError(err)
	agentAddr, _, err := s.addKeyToKeyring(val, "agent1")
	s.Require().NoError(err)

	// 2. Fund operator account (needs tokens for delegation via registration).
	err = s.fundAccount(val, operatorAddr, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(10_000_000))))
	s.Require().NoError(err)

	// 3. Fund agent account with some gas for activity submission.
	err = s.fundAccount(val, agentAddr, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(1_000_000))))
	s.Require().NoError(err)

	// 4. Register node.
	operatorCtx := s.clientCtxForKey(val, "operator1")
	nodeID, err := s.registerNode(
		val, operatorCtx, operatorAddr, agentAddr.String(),
		operatorPubKey, "test-node-1", sdkmath.LegacyNewDec(30),
	)
	s.Require().NoError(err)
	s.T().Logf("registered node: %s", nodeID)

	// 5. Submit activities across multiple windows to achieve eligibility.
	// Fast params: EpochLength=120, WindowsPerEpoch=12 → window=10 blocks.
	// Need activities in 8 out of 12 windows.
	agentCtx := s.clientCtxForKey(val, "agent1")
	startHeight := s.currentHeight()
	s.T().Logf("starting activity submission at height %d", startHeight)

	// Submit one activity per window for 8 windows.
	// Each window is 10 blocks. We submit at the start of each window.
	windowLength := int64(10) // 120 / 12
	for w := 0; w < 8; w++ {
		// Wait until we are in the target window.
		targetHeight := startHeight + int64(w)*windowLength
		if targetHeight > s.currentHeight() {
			s.waitForHeight(targetHeight)
		}

		hash := generateHash(fmt.Sprintf("activity-%s-%d", nodeID, w))
		uri := fmt.Sprintf("ipfs://activity-%d", w)

		err := s.submitActivity(agentCtx, agentAddr.String(), hash, uri)
		s.Require().NoError(err, "failed to submit activity in window %d", w)
		s.T().Logf("submitted activity in window %d at height ~%d", w, s.currentHeight())
	}

	// 6. Wait for epoch boundary.
	// Epoch 0 ends at block 120. Find the next epoch boundary.
	currentH := s.currentHeight()
	epochLength := int64(120)
	epochBoundary := ((currentH / epochLength) + 1) * epochLength
	s.T().Logf("waiting for epoch boundary at height %d (current: %d)", epochBoundary, currentH)
	s.waitForHeight(epochBoundary + 1) // +1 to ensure EndBlocker ran

	s.T().Log("epoch boundary reached, node activity flow test passed")
}
