package types

import (
	"encoding/hex"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParams returns the default module parameters.
func DefaultParams() Params {
	return Params{
		BeaconVerificationEnabled: false,  // Start in stub mode
		MaxBeaconAgeBlocks:        17280,  // ~1 epoch (~1 day)
		DrandGenesisTime:          1692803367, // drand quicknet genesis
		DrandPeriodSeconds:        3,       // quicknet: 3s per round
		DrandChainHash:            "52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971",
		DrandPublicKey:            "", // Set via governance when BLS verification is enabled
		DrandSchemeId:             "bls-unchained-g1-rfc9380",
		CommitRevealEnabled:       false,  // Disabled by default
		MinFutureRounds:           30,
		RequestTimeoutBlocks:      5760,   // ~8 hours at 5s/block
		RequestPruningBlocks:      17280,  // ~1 epoch/day
		MaxPendingRequests:        1000,
		MaxRequestsPerRequester:   10,
		MaxFulfillsPerBlock:       50,
		MinRequestFee:             sdk.NewCoin("uppyeo", math.NewInt(1000)),
	}
}

// Validate performs basic validation of params.
func (p Params) Validate() error {
	if p.MaxBeaconAgeBlocks == 0 {
		return fmt.Errorf("max_beacon_age_blocks must be positive")
	}
	if p.DrandPeriodSeconds == 0 {
		return fmt.Errorf("drand_period_seconds must be positive")
	}
	if p.CommitRevealEnabled {
		if p.MinFutureRounds == 0 {
			return fmt.Errorf("min_future_rounds must be positive when commit-reveal is enabled")
		}
		if p.RequestTimeoutBlocks == 0 {
			return fmt.Errorf("request_timeout_blocks must be positive when commit-reveal is enabled")
		}
		if p.MaxPendingRequests == 0 {
			return fmt.Errorf("max_pending_requests must be positive when commit-reveal is enabled")
		}
		if p.MaxRequestsPerRequester == 0 {
			return fmt.Errorf("max_requests_per_requester must be positive when commit-reveal is enabled")
		}
		if p.MaxFulfillsPerBlock == 0 {
			return fmt.Errorf("max_fulfills_per_block must be positive when commit-reveal is enabled")
		}
		if !p.MinRequestFee.IsPositive() {
			return fmt.Errorf("min_request_fee must be positive when commit-reveal is enabled")
		}
	}
	return nil
}

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:              DefaultParams(),
		Beacons:             []Beacon{},
		RandomnessRequests:  []RandomnessRequest{},
		NextRequestId:       1,
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	rounds := make(map[uint64]bool)
	for i, beacon := range gs.Beacons {
		if beacon.Round == 0 {
			return fmt.Errorf("beacon %d has zero round", i)
		}
		if rounds[beacon.Round] {
			return fmt.Errorf("duplicate beacon round: %d", beacon.Round)
		}
		rounds[beacon.Round] = true

		if beacon.Randomness == "" {
			return fmt.Errorf("beacon %d (round %d) has empty randomness", i, beacon.Round)
		}
	}

	requestIDs := make(map[uint64]bool)
	for i, req := range gs.RandomnessRequests {
		if req.RequestId == 0 {
			return fmt.Errorf("randomness request %d has zero request_id", i)
		}
		if requestIDs[req.RequestId] {
			return fmt.Errorf("duplicate randomness request ID: %d", req.RequestId)
		}
		requestIDs[req.RequestId] = true

		if req.Requester == "" {
			return fmt.Errorf("randomness request %d has empty requester", i)
		}
		if len(req.CommitHash) != 64 {
			return fmt.Errorf("randomness request %d has invalid commit_hash length", i)
		}
		if _, err := hex.DecodeString(req.CommitHash); err != nil {
			return fmt.Errorf("randomness request %d has non-hex commit_hash", i)
		}
		if req.NumWords == 0 || req.NumWords > 10 {
			return fmt.Errorf("randomness request %d has invalid num_words: %d", i, req.NumWords)
		}
		if req.Status == RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_UNSPECIFIED {
			return fmt.Errorf("randomness request %d has unspecified status", i)
		}
	}

	if gs.NextRequestId == 0 {
		return fmt.Errorf("next_request_id must be positive")
	}

	return nil
}
