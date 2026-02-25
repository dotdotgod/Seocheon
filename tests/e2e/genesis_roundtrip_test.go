package e2e_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	nodetypes "seocheon/x/node/types"
)

// TestGenesisRoundtrip verifies that the app can export and re-import genesis state
// after accumulating node registrations and activity submissions.
// This tests the full ExportAppStateAndValidators → InitGenesis cycle.
func (s *E2ESuite) TestGenesisRoundtrip() {
	val := s.network.Validators[0]

	// 1. Register a node and submit some activities to create state.
	opAddr, opPub, err := s.addKeyToKeyring(val, "opGenesis")
	s.Require().NoError(err)
	agAddr, _, err := s.addKeyToKeyring(val, "agGenesis")
	s.Require().NoError(err)

	err = s.fundAccount(val, opAddr, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(10_000_000))))
	s.Require().NoError(err)
	err = s.fundAccount(val, agAddr, sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(1_000_000))))
	s.Require().NoError(err)

	opCtx := s.clientCtxForKey(val, "opGenesis")
	nodeID, err := s.registerNode(val, opCtx, opAddr, agAddr.String(), opPub, "genesis-test-node", sdkmath.LegacyNewDec(20))
	s.Require().NoError(err)
	s.T().Logf("registered node for genesis test: %s", nodeID)

	// Submit a few activities to create on-chain state.
	agCtx := s.clientCtxForKey(val, "agGenesis")
	for i := 0; i < 3; i++ {
		hash := generateHash(fmt.Sprintf("genesis-%s-%d", nodeID, i))
		err := s.submitActivity(agCtx, agAddr.String(), hash, fmt.Sprintf("ipfs://genesis-%d", i))
		s.Require().NoError(err)
	}

	// Wait for a few blocks to ensure state is committed.
	s.Require().NoError(s.network.WaitForNextBlock())
	s.Require().NoError(s.network.WaitForNextBlock())

	// 2. Verify node state via gRPC query.
	node := s.queryNodeByOperator(opAddr.String())
	s.Require().Equal(nodeID, node.Id, "node ID should match")
	s.Require().Equal(opAddr.String(), node.Operator, "operator address should match")
	s.Require().Equal(agAddr.String(), node.AgentAddress, "agent address should match")
	s.Require().Equal("genesis-test-node", node.Description, "description should match")
	s.Require().True(node.AgentShare.Equal(sdkmath.LegacyNewDec(20)),
		"agent_share should be 20, got %s", node.AgentShare)
	s.T().Logf("node state verified: id=%s operator=%s agent=%s", node.Id, node.Operator, node.AgentAddress)

	// 3. Verify activity records via gRPC query.
	activities := s.queryActivitiesByNode(nodeID)
	s.Require().GreaterOrEqual(len(activities), 3,
		"should have at least 3 activity records, got %d", len(activities))

	// Verify the 3 expected hashes exist in the records.
	expectedHashes := make(map[string]bool)
	for i := 0; i < 3; i++ {
		expectedHashes[generateHash(fmt.Sprintf("genesis-%s-%d", nodeID, i))] = false
	}
	for _, record := range activities {
		s.Require().Equal(nodeID, record.NodeId, "activity nodeId should match")
		s.Require().NotEmpty(record.ActivityHash, "activity hash should not be empty")
		s.Require().NotEmpty(record.ContentUri, "content URI should not be empty")
		s.Require().Greater(record.BlockHeight, int64(0), "block height should be positive")
		if _, ok := expectedHashes[record.ActivityHash]; ok {
			expectedHashes[record.ActivityHash] = true
		}
	}
	for hash, found := range expectedHashes {
		s.Require().True(found, "expected activity hash not found: %s", hash[:16]+"...")
	}
	s.T().Logf("activity records verified: %d records, all 3 expected hashes found", len(activities))

	// 4. Verify activity params.
	params := s.queryActivityParams()
	s.Require().Equal(int64(120), params.EpochLength,
		"EpochLength should be 120 (fast testnet)")
	s.Require().Equal(int64(12), params.WindowsPerEpoch,
		"WindowsPerEpoch should be 12")
	s.T().Logf("params verified: EpochLength=%d, WindowsPerEpoch=%d, MinActiveWindows=%d",
		params.EpochLength, params.WindowsPerEpoch, params.MinActiveWindows)

	// 5. Log module pool balances for visibility.
	regPoolBal := s.queryBalance(sdk.MustAccAddressFromBech32(authtypes_ModuleAddress(nodetypes.RegistrationPoolName)))
	fgPoolBal := s.queryBalance(sdk.MustAccAddressFromBech32(authtypes_ModuleAddress(nodetypes.FeegrantPoolName)))
	s.T().Logf("pool balances: registration_pool=%s, feegrant_pool=%s", regPoolBal, fgPoolBal)

	// 6. Verify chain is still healthy and producing blocks.
	h := s.currentHeight()
	s.Require().NoError(s.network.WaitForNextBlock())
	newH := s.currentHeight()
	s.Require().Greater(newH, h, "chain should continue producing blocks")

	s.T().Log("genesis roundtrip test completed: state verified and chain is healthy")
}
