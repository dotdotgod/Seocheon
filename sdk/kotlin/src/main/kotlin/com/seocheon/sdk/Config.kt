package com.seocheon.sdk

import com.seocheon.sdk.constants.ChainConstants
import com.seocheon.sdk.errors.SdkError
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * Blockchain connection settings.
 */
@Serializable
data class ChainConfig(
    @SerialName("chain_id") val chainId: String,
    @SerialName("rpc_endpoint") val rpcEndpoint: String,
    @SerialName("grpc_endpoint") val grpcEndpoint: String,
    @SerialName("gas_price") val gasPrice: String = ChainConstants.DEFAULT_GAS_PRICE,
    @SerialName("gas_adjustment") val gasAdjustment: Double = ChainConstants.DEFAULT_GAS_ADJUSTMENT,
)

/**
 * Signing mode specifies how transactions are signed.
 */
enum class SigningMode {
    VAULT,
    KEYSTORE,
    DIRECT,
}

/**
 * Transaction signing settings.
 */
@Serializable
data class SigningConfig(
    val mode: SigningMode,
    @SerialName("vault_endpoint") val vaultEndpoint: String = "",
    @SerialName("key_name") val keyName: String = "",
    @SerialName("keystore_path") val keystorePath: String = "",
    @SerialName("passphrase_env") val passphraseEnv: String = "",
    val mnemonic: String = "",
)

/**
 * Transaction broadcast settings.
 */
@Serializable
data class TxConfig(
    @SerialName("broadcast_mode") val broadcastMode: String = ChainConstants.DEFAULT_BROADCAST_MODE,
    @SerialName("confirm_timeout_ms") val confirmTimeoutMs: Long = ChainConstants.DEFAULT_CONFIRM_TIMEOUT_MS,
    @SerialName("confirm_poll_interval_ms") val confirmPollIntervalMs: Long = ChainConstants.DEFAULT_CONFIRM_POLL_MS,
)

/**
 * Top-level SDK configuration.
 */
@Serializable
data class SDKConfig(
    val chain: ChainConfig,
    val signing: SigningConfig,
    val tx: TxConfig = TxConfig(),
) {
    /**
     * Validates this configuration, throwing [SdkError.InvalidConfig] on invalid state.
     */
    fun validate() {
        if (chain.chainId.isBlank()) {
            throw SdkError.InvalidConfig("chain_id is required")
        }
        if (chain.rpcEndpoint.isBlank()) {
            throw SdkError.InvalidConfig("rpc_endpoint is required")
        }
        if (chain.grpcEndpoint.isBlank()) {
            throw SdkError.InvalidConfig("grpc_endpoint is required")
        }
        when (signing.mode) {
            SigningMode.VAULT -> {
                if (signing.vaultEndpoint.isBlank() || signing.keyName.isBlank()) {
                    throw SdkError.InvalidConfig("vault mode requires vault_endpoint and key_name")
                }
            }
            SigningMode.KEYSTORE -> {
                if (signing.keystorePath.isBlank() || signing.passphraseEnv.isBlank()) {
                    throw SdkError.InvalidConfig("keystore mode requires keystore_path and passphrase_env")
                }
            }
            SigningMode.DIRECT -> {
                if (signing.mnemonic.isBlank()) {
                    throw SdkError.InvalidConfig("direct mode requires mnemonic")
                }
            }
        }
    }

    companion object {
        /**
         * Creates an SDKConfig with default values applied.
         */
        fun default(
            chainId: String,
            rpcEndpoint: String,
            grpcEndpoint: String,
            signing: SigningConfig,
        ): SDKConfig = SDKConfig(
            chain = ChainConfig(
                chainId = chainId,
                rpcEndpoint = rpcEndpoint,
                grpcEndpoint = grpcEndpoint,
            ),
            signing = signing,
        )
    }
}
