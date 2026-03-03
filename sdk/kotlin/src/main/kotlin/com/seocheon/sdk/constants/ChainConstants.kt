package com.seocheon.sdk.constants

/**
 * Chain parameters, token denominations, and default SDK configuration values
 * for the Seocheon blockchain.
 */
object ChainConstants {

    // Epoch/Window parameters (x/activity params)
    const val EPOCH_LENGTH: Long = 17280          // blocks per epoch (~24h at 5s/block)
    const val WINDOWS_PER_EPOCH: Long = 12        // windows per epoch
    const val MIN_ACTIVE_WINDOWS: Long = 8        // minimum active windows for reward qualification
    const val WINDOW_LENGTH: Long = 1440           // blocks per window (EPOCH_LENGTH / WINDOWS_PER_EPOCH)

    // Quota parameters (x/activity params)
    const val SELF_FUNDED_QUOTA: Long = 100        // epoch quota for self-funded nodes
    const val FEEGRANT_QUOTA: Long = 10            // epoch quota for feegrant nodes

    // Pruning parameters
    const val ACTIVITY_PRUNING_KEEP_BLOCKS: Long = 6307200  // ~1 year

    // Fee model parameters (x/activity params)
    const val FEE_THRESHOLD_MULTIPLIER: Long = 3
    const val BASE_ACTIVITY_FEE: Long = 10000000000L         // 1 KKOT in uppyeo
    const val FEE_EXPONENT: Long = 5000                       // basis points (0.5)
    const val MAX_ACTIVITY_FEE: Long = 1000000000000L         // 100 KKOT in uppyeo
    const val MIN_FEEGRANT_QUOTA: Long = 8
    const val QUOTA_REDUCTION_RATE: Long = 5000               // basis points (0.5)
    const val FEEGRANT_FEE_EXEMPT: Boolean = true

    // Dual reward pool parameters (x/activity params)
    const val D_MIN: Long = 3000                   // basis points (0.3)
    const val FEE_TO_ACTIVITY_POOL_RATIO: Long = 8000  // basis points (0.8)

    // Node registration parameters (x/node params)
    const val MAX_REGISTRATIONS_PER_BLOCK: Long = 5
    const val REGISTRATION_COOLDOWN_BLKS: Long = 100
    const val REGISTRATION_DEPOSIT: String = "0"   // uppyeo
    const val MAX_TAGS: Long = 10
    const val MAX_TAG_LENGTH: Long = 32

    // Agent permission parameters
    val AGENT_ALLOWED_MSG_TYPES: List<String> = listOf(
        "/seocheon.activity.v1.MsgSubmitActivity",
        "/cosmos.bank.v1beta1.MsgSend",
    )
    val AGENT_FEEGRANT_ALLOWED_MSG_TYPES: List<String> = listOf(
        "/seocheon.activity.v1.MsgSubmitActivity",
    )

    const val AGENT_ADDRESS_CHANGE_COOLDOWN: Long = 17280  // 1 epoch

    // Time-block conversion (5s/block)
    const val BLOCKS_PER_HOUR: Long = 720
    const val BLOCKS_PER_DAY: Long = 17280
    const val BLOCKS_PER_YEAR: Long = 6307200
    const val UNBONDING_PERIOD_BLOCKS: Long = 362880   // ~21 days
    const val FEEGRANT_EXPIRY_BLOCKS: Long = 3110400   // ~180 days

    // Token denomination constants (6-stage system)
    // uppyeo(0) -> sal(2) -> pi(4) -> sum(6) -> hon(8) -> kkot(10)
    const val TOKEN_BASE_DENOM: String = "uppyeo"    // base denomination (10^0)
    const val TOKEN_SAL_DENOM: String = "sal"        // sal denomination (10^2)
    const val TOKEN_PI_DENOM: String = "pi"          // pi denomination (10^4)
    const val TOKEN_SUM_DENOM: String = "sum"        // sum denomination (10^6)
    const val TOKEN_HON_DENOM: String = "hon"        // hon denomination (10^8)
    const val TOKEN_DISPLAY_DENOM: String = "kkot"   // display denomination (10^10)

    // Denomination conversion factors (base unit: uppyeo)
    const val UPPYEO_PER_SAL: Long = 100
    const val UPPYEO_PER_PI: Long = 10000
    const val UPPYEO_PER_SUM: Long = 1000000
    const val UPPYEO_PER_HON: Long = 100000000
    const val UPPYEO_PER_KKOT: Long = 10000000000L

    // Gas defaults
    const val GAS_SUBMIT_ACTIVITY: Long = 200000
    const val GAS_WITHDRAW: Long = 300000
    const val GAS_SEND: Long = 100000
    const val GAS_FALLBACK: Long = 200000
    const val GAS_PRICE: Long = 250

    // Fee denomination
    const val FEE_DENOM: String = "uppyeo"

    // Default SDK configuration values
    const val DEFAULT_GAS_PRICE: String = "250uppyeo"
    const val DEFAULT_GAS_ADJUSTMENT: Double = 1.3
    const val DEFAULT_BROADCAST_MODE: String = "sync"
    const val DEFAULT_CONFIRM_TIMEOUT_MS: Long = 30000
    const val DEFAULT_CONFIRM_POLL_MS: Long = 1000

    // BIP44 path
    const val BIP44_COIN_TYPE: Int = 118
    const val BIP44_PATH: String = "m/44'/118'/0'/0/0"

    // Address prefix
    const val ADDRESS_PREFIX: String = "seocheon"
}
