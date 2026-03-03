"""Default gas limits and fee calculation."""

# Default gas limits per message type
DEFAULT_GAS_SUBMIT_ACTIVITY: int = 200000
DEFAULT_GAS_WITHDRAW_NODE_COMMISSION: int = 300000
DEFAULT_GAS_SEND: int = 100000
DEFAULT_GAS_FALLBACK: int = 200000

# Default fee parameters
DEFAULT_FEE_DENOM: str = "uppyeo"
DEFAULT_GAS_PRICE: int = 250  # 250 uppyeo per gas unit


def default_gas_for_message(type_url: str) -> int:
    """Return the default gas limit for a given message type URL."""
    _gas_map = {
        "/seocheon.activity.v1.MsgSubmitActivity": DEFAULT_GAS_SUBMIT_ACTIVITY,
        "/seocheon.node.v1.MsgWithdrawNodeCommission": DEFAULT_GAS_WITHDRAW_NODE_COMMISSION,
        "/cosmos.bank.v1beta1.MsgSend": DEFAULT_GAS_SEND,
    }
    return _gas_map.get(type_url, DEFAULT_GAS_FALLBACK)


def calculate_fee(gas_limit: int, gas_price: int) -> int:
    """Compute the fee amount from gas limit and gas price."""
    return gas_limit * gas_price
