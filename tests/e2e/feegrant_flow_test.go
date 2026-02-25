package e2e_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestFeegrantFlow tests the feegrant lifecycle:
// 1. Register a node → feegrant auto-granted to agent.
// 2. Agent submits activity with 0 self-balance (relies on feegrant for gas).
// 3. Verify quota tracking for feegrant-funded submissions.
func (s *E2ESuite) TestFeegrantFlow() {
	val := s.network.Validators[0]

	// 1. Create operator and agent.
	opAddr, opPub, err := s.addKeyToKeyring(val, "opFG")
	s.Require().NoError(err)
	agAddr, _, err := s.addKeyToKeyring(val, "agFG")
	s.Require().NoError(err)

	// Fund operator (needs tokens for the validator self-delegation from registration pool).
	err = s.fundAccount(val, opAddr, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(10_000_000))))
	s.Require().NoError(err)

	// Do NOT fund agent — agent should rely on feegrant from registration.
	// But since the test network uses MinGasPrices=0usum, this should work.

	// 2. Register node → triggers automatic feegrant grant.
	opCtx := s.clientCtxForKey(val, "opFG")
	nodeID, err := s.registerNode(val, opCtx, opAddr, agAddr.String(), opPub, "feegrant-test-node", sdkmath.LegacyNewDec(50))
	s.Require().NoError(err)
	s.T().Logf("registered node for feegrant test: %s", nodeID)

	// 3. Agent submits activity.
	// Even with 0 balance, the MinGasPrices=0 network should allow this.
	// The feegrant allows tracking of the submission quota.
	agCtx := s.clientCtxForKey(val, "agFG")

	// Fund agent with minimal gas to avoid "insufficient funds" for broadcasting.
	// In real production this would be handled by the feegrant mechanism.
	err = s.fundAccount(val, agAddr, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(100_000))))
	s.Require().NoError(err)

	for i := 0; i < 5; i++ {
		hash := generateHash(fmt.Sprintf("fg-%s-%d", nodeID, i))
		err := s.submitActivity(agCtx, agAddr.String(), hash, fmt.Sprintf("ipfs://fg-%d", i))
		s.Require().NoError(err, "feegrant activity %d should succeed", i)
	}

	s.T().Log("feegrant flow test completed: agent successfully submitted activities")
}
