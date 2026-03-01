// Epoch/Window parameters (x/activity params)
export const EPOCH_LENGTH = 17280;
export const WINDOWS_PER_EPOCH = 12;
export const MIN_ACTIVE_WINDOWS = 8;
export const WINDOW_LENGTH = EPOCH_LENGTH / WINDOWS_PER_EPOCH; // 1440

// Quota (x/activity params)
export const SELF_FUNDED_QUOTA = 100;
export const FEEGRANT_QUOTA = 10;

// Pruning (x/activity params)
export const ACTIVITY_PRUNING_KEEP_BLOCKS = 6307200;

// Fee model (x/activity params)
export const FEE_THRESHOLD_MULTIPLIER = 3;
export const BASE_ACTIVITY_FEE = 10000000000; // uppyeo, 1 KKOT
export const FEE_EXPONENT = 5000; // basis points, 0.5
export const MAX_ACTIVITY_FEE = 1000000000000; // uppyeo, 100 KKOT
export const MIN_FEEGRANT_QUOTA = 8;
export const QUOTA_REDUCTION_RATE = 5000; // basis points, 0.5
export const FEEGRANT_FEE_EXEMPT = true;

// Dual reward pool (x/activity params)
export const D_MIN = 3000; // basis points, 0.3
export const FEE_TO_ACTIVITY_POOL_RATIO = 8000; // basis points, 0.8

// Node registration (x/node params)
export const MAX_REGISTRATIONS_PER_BLOCK = 5;
export const REGISTRATION_COOLDOWN_BLOCKS = 100;
export const REGISTRATION_DEPOSIT = "0"; // uppyeo
export const MAX_TAGS = 10;
export const MAX_TAG_LENGTH = 32;

// Agent permissions (x/node params)
export const AGENT_ALLOWED_MSG_TYPES = [
  "/seocheon.activity.v1.MsgSubmitActivity",
  "/cosmos.bank.v1beta1.MsgSend",
] as const;
export const AGENT_FEEGRANT_ALLOWED_MSG_TYPES = [
  "/seocheon.activity.v1.MsgSubmitActivity",
] as const;
export const AGENT_ADDRESS_CHANGE_COOLDOWN = 17280; // 1 epoch

// Time-block conversion (5s/block)
export const BLOCKS_PER_HOUR = 720;
export const BLOCKS_PER_DAY = 17280;
export const BLOCKS_PER_YEAR = 6307200;
export const UNBONDING_PERIOD_BLOCKS = 362880; // ~21 days
export const FEEGRANT_EXPIRY_BLOCKS = 3110400; // ~180 days

// Token denomination (6-stage system)
// uppyeo(0) -> sal(2) -> pi(4) -> sum(6) -> hon(8) -> kkot(10)
export const TOKEN_BASE_DENOM = "uppyeo";
export const TOKEN_SAL_DENOM = "sal";
export const TOKEN_PI_DENOM = "pi";
export const TOKEN_SUM_DENOM = "sum";
export const TOKEN_HON_DENOM = "hon";
export const TOKEN_DISPLAY_DENOM = "kkot";

// Conversion factors (base unit: uppyeo)
export const UPPYEO_PER_SAL = 100n;
export const UPPYEO_PER_PI = 10000n;
export const UPPYEO_PER_SUM = 1000000n;
export const UPPYEO_PER_HON = 100000000n;
export const UPPYEO_PER_KKOT = 10000000000n;

// Activity hash
export const ACTIVITY_HASH_LENGTH = 64;
