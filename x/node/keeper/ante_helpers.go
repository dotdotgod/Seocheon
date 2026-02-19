package keeper

import "context"

// IsRegisteredAgent checks if the given address is a registered agent_address.
// Used by the AgentPermissionDecorator ante handler.
func (k Keeper) IsRegisteredAgent(ctx context.Context, address string) bool {
	has, _ := k.AgentIndex.Has(ctx, address)
	return has
}

// GetAllowedAgentMsgTypes returns the list of message type URLs that agent addresses are allowed to execute.
// Used by the AgentPermissionDecorator ante handler.
func (k Keeper) GetAllowedAgentMsgTypes(ctx context.Context) ([]string, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return params.AgentAllowedMsgTypes, nil
}
