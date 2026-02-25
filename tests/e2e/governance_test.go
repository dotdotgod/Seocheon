package e2e_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	activitytypes "seocheon/x/activity/types"
)

// TestGovernanceParamChange tests that x/activity params can be updated
// via governance proposal end-to-end:
// 1. Query current params (MinActiveWindows=8).
// 2. Submit a MsgUpdateParams proposal to change MinActiveWindows to 6.
// 3. Both validators vote YES.
// 4. Wait for the voting period to end.
// 5. Verify the proposal passed and params are updated.
func (s *E2ESuite) TestGovernanceParamChange() {
	val0 := s.network.Validators[0]
	val1 := s.network.Validators[1]

	// --- Step 1: Query current params ---
	params := s.queryActivityParams()
	s.T().Logf("current params: EpochLength=%d, WindowsPerEpoch=%d, MinActiveWindows=%d",
		params.EpochLength, params.WindowsPerEpoch, params.MinActiveWindows)
	s.Require().Equal(int64(8), params.MinActiveWindows, "initial MinActiveWindows should be 8")

	// --- Step 2: Build MsgUpdateParams with MinActiveWindows=6 ---
	govAuthority := authtypes.NewModuleAddress("gov").String()

	// Construct params explicitly to avoid proto unknown field issues.
	newParams := activitytypes.Params{
		EpochLength:              params.EpochLength,
		WindowsPerEpoch:          params.WindowsPerEpoch,
		MinActiveWindows:         6, // changed from 8
		SelfFundedQuota:          params.SelfFundedQuota,
		FeegrantQuota:            params.FeegrantQuota,
		ActivityPruningKeepBlocks: params.ActivityPruningKeepBlocks,
		FeeThresholdMultiplier:   params.FeeThresholdMultiplier,
		BaseActivityFee:          params.BaseActivityFee,
		FeeExponent:              params.FeeExponent,
		MaxActivityFee:           params.MaxActivityFee,
		MinFeegrantQuota:         params.MinFeegrantQuota,
		QuotaReductionRate:       params.QuotaReductionRate,
		FeegrantFeeExempt:        params.FeegrantFeeExempt,
		DMin:                     params.DMin,
		FeeToActivityPoolRatio:   params.FeeToActivityPoolRatio,
	}

	updateMsg := &activitytypes.MsgUpdateParams{
		Authority: govAuthority,
		Params:    newParams,
	}

	deposit := sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(10)))
	proposalMsg, err := govv1.NewMsgSubmitProposal(
		[]sdk.Msg{updateMsg},
		deposit,
		val0.Address.String(),
		"", // metadata
		"Update MinActiveWindows to 6",
		"Lower the minimum active windows threshold from 8 to 6 for testing.",
		false, // not expedited
	)
	s.Require().NoError(err)

	// Submit proposal from val0.
	clientCtx0 := val0.ClientCtx.
		WithFromAddress(val0.Address).
		WithFromName("node0")
	resp, err := s.broadcastTx(clientCtx0, proposalMsg)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), resp.Code, "proposal submission failed: %s", resp.RawLog)
	s.T().Log("proposal submitted successfully")

	// --- Step 3: Verify proposal is in VOTING_PERIOD ---
	s.Require().NoError(s.network.WaitForNextBlock())
	proposal := s.queryGovProposal(1)
	s.Require().Equal(govv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, proposal.Status,
		"proposal should be in voting period, got %s", proposal.Status)
	s.T().Logf("proposal #1 status: %s", proposal.Status)

	// --- Step 4: Both validators vote YES ---
	voteMsg0 := govv1.NewMsgVote(val0.Address, 1, govv1.VoteOption_VOTE_OPTION_YES, "")
	resp, err = s.broadcastTx(clientCtx0, voteMsg0)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), resp.Code, "val0 vote failed: %s", resp.RawLog)
	s.T().Log("val0 voted YES")

	clientCtx1 := val0.ClientCtx.
		WithKeyring(val1.ClientCtx.Keyring).
		WithFromAddress(val1.Address).
		WithFromName("node1")
	voteMsg1 := govv1.NewMsgVote(val1.Address, 1, govv1.VoteOption_VOTE_OPTION_YES, "")
	resp, err = s.broadcastTx(clientCtx1, voteMsg1)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), resp.Code, "val1 vote failed: %s", resp.RawLog)
	s.T().Log("val1 voted YES")

	// --- Step 5: Wait for voting period to end (10s configured in patchGovGenesis) ---
	deadline := time.Now().Add(30 * time.Second)
	var finalStatus govv1.ProposalStatus
	for time.Now().Before(deadline) {
		s.Require().NoError(s.network.WaitForNextBlock())
		p := s.queryGovProposal(1)
		if p.Status != govv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
			finalStatus = p.Status
			break
		}
	}
	s.Require().Equal(govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, finalStatus,
		"proposal should have passed, got %s", finalStatus)
	s.T().Log("proposal #1 PASSED")

	// --- Step 6: Verify params updated ---
	updatedParams := s.queryActivityParams()
	s.T().Logf("updated params: MinActiveWindows=%d", updatedParams.MinActiveWindows)
	s.Require().Equal(int64(6), updatedParams.MinActiveWindows,
		"MinActiveWindows should be updated to 6")

	// Verify other params remain unchanged.
	s.Require().Equal(params.EpochLength, updatedParams.EpochLength,
		"EpochLength should not change")
	s.Require().Equal(params.WindowsPerEpoch, updatedParams.WindowsPerEpoch,
		"WindowsPerEpoch should not change")
	s.Require().Equal(params.SelfFundedQuota, updatedParams.SelfFundedQuota,
		"SelfFundedQuota should not change")
	s.Require().Equal(params.FeegrantQuota, updatedParams.FeegrantQuota,
		"FeegrantQuota should not change")
	s.Require().Equal(params.DMin, updatedParams.DMin,
		"DMin should not change")

	s.T().Log("governance param change test completed successfully")
}
