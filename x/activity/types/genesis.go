package types

import "fmt"

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		Activities: []ActivityRecord{},
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	hashSet := make(map[string]bool)
	for i, activity := range gs.Activities {
		if activity.NodeId == "" {
			return fmt.Errorf("activity %d has empty node_id", i)
		}
		if activity.ActivityHash == "" {
			return fmt.Errorf("activity %d has empty activity_hash", i)
		}
		if len(activity.ActivityHash) != 64 {
			return fmt.Errorf("activity %d has invalid activity_hash length: %d (expected 64)", i, len(activity.ActivityHash))
		}
		if activity.ContentUri == "" {
			return fmt.Errorf("activity %d has empty content_uri", i)
		}

		// Check for duplicate hashes within the same epoch.
		key := fmt.Sprintf("%s:%d:%s", activity.NodeId, activity.Epoch, activity.ActivityHash)
		if hashSet[key] {
			return fmt.Errorf("duplicate activity hash for node %s in epoch %d: %s", activity.NodeId, activity.Epoch, activity.ActivityHash)
		}
		hashSet[key] = true
	}

	return nil
}
