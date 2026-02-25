package e2e_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestSpamDefense tests quota enforcement and duplicate hash rejection:
// 1. Register a node (which auto-grants feegrant → feegrant_quota=10 applies).
// 2. Submit up to the feegrant quota limit.
// 3. Verify that exceeding the quota returns an error.
// 4. Verify duplicate hash rejection.
func (s *E2ESuite) TestSpamDefense() {
	val := s.network.Validators[0]

	// Create operator and agent.
	opAddr, opPub, err := s.addKeyToKeyring(val, "opSpam")
	s.Require().NoError(err)
	agAddr, _, err := s.addKeyToKeyring(val, "agSpam")
	s.Require().NoError(err)

	// Fund accounts.
	err = s.fundAccount(val, opAddr, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(10_000_000))))
	s.Require().NoError(err)
	err = s.fundAccount(val, agAddr, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(5_000_000))))
	s.Require().NoError(err)

	// Register node (auto-grants feegrant → feegrant_quota=10 applies).
	opCtx := s.clientCtxForKey(val, "opSpam")
	nodeID, err := s.registerNode(val, opCtx, opAddr, agAddr.String(), opPub, "spam-test-node", sdkmath.LegacyZeroDec())
	s.Require().NoError(err)
	s.T().Logf("registered node for spam test: %s", nodeID)

	// Feegrant quota is 10 per epoch (DefaultFeegrantQuota).
	// Submit exactly 10 activities to exhaust the quota.
	agCtx := s.clientCtxForKey(val, "agSpam")
	feegrantQuota := 10

	for i := 0; i < feegrantQuota; i++ {
		hash := generateHash(fmt.Sprintf("spam-%s-%d", nodeID, i))
		err := s.submitActivity(agCtx, agAddr.String(), hash, fmt.Sprintf("ipfs://spam-%d", i))
		s.Require().NoError(err, "activity %d should succeed (within quota)", i)
	}
	s.T().Logf("submitted %d activities, quota exhausted", feegrantQuota)

	// 11th activity should fail with quota exceeded.
	overQuotaHash := generateHash(fmt.Sprintf("spam-%s-%d", nodeID, feegrantQuota))
	err = s.submitActivity(agCtx, agAddr.String(), overQuotaHash, "ipfs://over-quota")
	s.Require().Error(err, "activity beyond quota should be rejected")
	s.Require().Contains(err.Error(), "quota exceeded", "error should mention quota exceeded")
	s.T().Log("quota enforcement verified: over-quota activity rejected")

	// Verify duplicate hash rejection (resubmit first hash).
	duplicateHash := generateHash(fmt.Sprintf("spam-%s-0", nodeID))
	err = s.submitActivity(agCtx, agAddr.String(), duplicateHash, "ipfs://dup")
	s.Require().Error(err, "duplicate hash should be rejected")
	s.T().Log("duplicate hash correctly rejected")

	s.T().Log("spam defense test completed")
}
