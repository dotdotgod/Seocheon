"""Chain parameters, token denominations, and SDK defaults for the Seocheon blockchain."""

# Epoch/Window parameters (x/activity params)
EPOCH_LENGTH: int = 17280  # blocks per epoch (~24h at 5s/block)
WINDOWS_PER_EPOCH: int = 12  # windows per epoch
MIN_ACTIVE_WINDOWS: int = 8  # minimum active windows for reward qualification
WINDOW_LENGTH: int = 1440  # blocks per window (EPOCH_LENGTH / WINDOWS_PER_EPOCH)

# Quota parameters (x/activity params)
SELF_FUNDED_QUOTA: int = 100  # epoch quota for self-funded nodes
FEEGRANT_QUOTA: int = 10  # epoch quota for feegrant nodes

# Pruning parameters
ACTIVITY_PRUNING_KEEP_BLOCKS: int = 6307200  # ~1 year

# Fee model parameters (x/activity params)
FEE_THRESHOLD_MULTIPLIER: int = 3
BASE_ACTIVITY_FEE: int = 10_000_000_000  # 1 KKOT in uppyeo
FEE_EXPONENT: int = 5000  # basis points (0.5)
MAX_ACTIVITY_FEE: int = 1_000_000_000_000  # 100 KKOT in uppyeo
MIN_FEEGRANT_QUOTA: int = 8
QUOTA_REDUCTION_RATE: int = 5000  # basis points (0.5)
FEEGRANT_FEE_EXEMPT: bool = True

# Dual reward pool parameters (x/activity params)
D_MIN: int = 3000  # basis points (0.3)
FEE_TO_ACTIVITY_POOL_RATIO: int = 8000  # basis points (0.8)

# Node registration parameters (x/node params)
MAX_REGISTRATIONS_PER_BLOCK: int = 5
REGISTRATION_COOLDOWN_BLKS: int = 100
REGISTRATION_DEPOSIT: str = "0"  # uppyeo
MAX_TAGS: int = 10
MAX_TAG_LENGTH: int = 32

# Agent permission parameters
AGENT_ALLOWED_MSG_TYPES: list[str] = [
    "/seocheon.activity.v1.MsgSubmitActivity",
    "/cosmos.bank.v1beta1.MsgSend",
]
AGENT_FEEGRANT_ALLOWED_MSG_TYPES: list[str] = [
    "/seocheon.activity.v1.MsgSubmitActivity",
]
AGENT_ADDRESS_CHANGE_COOLDOWN: int = 17280  # 1 epoch

# Time-block conversion (5s/block)
BLOCKS_PER_HOUR: int = 720
BLOCKS_PER_DAY: int = 17280
BLOCKS_PER_YEAR: int = 6307200
UNBONDING_PERIOD_BLOCKS: int = 362880  # ~21 days
FEEGRANT_EXPIRY_BLOCKS: int = 3110400  # ~180 days

# Token denomination constants (6-stage system)
# uppyeo(0) -> sal(2) -> pi(4) -> sum(6) -> hon(8) -> kkot(10)
TOKEN_BASE_DENOM: str = "uppyeo"  # base denomination (10^0)
TOKEN_SAL_DENOM: str = "sal"  # sal denomination (10^2)
TOKEN_PI_DENOM: str = "pi"  # pi denomination (10^4)
TOKEN_SUM_DENOM: str = "sum"  # sum denomination (10^6)
TOKEN_HON_DENOM: str = "hon"  # hon denomination (10^8)
TOKEN_DISPLAY_DENOM: str = "kkot"  # display denomination (10^10)

# Denomination conversion factors (base unit: uppyeo)
UPPYEO_PER_SAL: int = 100
UPPYEO_PER_PI: int = 10_000
UPPYEO_PER_SUM: int = 1_000_000
UPPYEO_PER_HON: int = 100_000_000
UPPYEO_PER_KKOT: int = 10_000_000_000

DENOM_FACTORS: dict[str, int] = {
    "uppyeo": 1,
    "sal": UPPYEO_PER_SAL,
    "pi": UPPYEO_PER_PI,
    "sum": UPPYEO_PER_SUM,
    "hon": UPPYEO_PER_HON,
    "kkot": UPPYEO_PER_KKOT,
}

# Default SDK configuration values
DEFAULT_GAS_PRICE: str = "250uppyeo"
DEFAULT_GAS_PRICE_VALUE: int = 250  # uppyeo per gas unit
DEFAULT_GAS_ADJUSTMENT: float = 1.3
DEFAULT_BROADCAST_MODE: str = "sync"
DEFAULT_CONFIRM_TIMEOUT_MS: int = 30000
DEFAULT_CONFIRM_POLL_MS: int = 1000
