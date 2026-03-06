package com.seocheon.sdk

import com.seocheon.sdk.errors.SdkError
import com.seocheon.sdk.infrastructure.ChainClient
import com.seocheon.sdk.infrastructure.OkHttpChainClient
import com.seocheon.sdk.infrastructure.signing.DirectService
import com.seocheon.sdk.infrastructure.signing.KeystoreService
import com.seocheon.sdk.infrastructure.signing.SigningService
import com.seocheon.sdk.infrastructure.signing.VaultService
import com.seocheon.sdk.internal.tx.PipelineConfig
import com.seocheon.sdk.modules.*

/**
 * Main Seocheon SDK client.
 * Provides access to 5 modules: activity, node, epoch, rewards, cosmos.
 */
class SeocheonSdk private constructor(
    private val config: SDKConfig,
    private val chainClient: ChainClient,
    private val signer: SigningService,
) {
    val activity: ActivityModule
    val node: NodeModule
    val epoch: EpochModule
    val rewards: RewardsModule
    val cosmos: CosmosModule

    private var connected = false

    init {
        val pipelineConfig = PipelineConfig(
            chainId = config.chain.chainId,
            gasPrice = parseGasPrice(config.chain.gasPrice),
            confirmTimeoutMs = config.tx.confirmTimeoutMs,
            pollIntervalMs = config.tx.confirmPollIntervalMs,
        )

        activity = ActivityModule(chainClient, signer, pipelineConfig)
        node = NodeModule(chainClient, signer, pipelineConfig)
        epoch = EpochModule(chainClient)
        rewards = RewardsModule(chainClient, signer, pipelineConfig)
        cosmos = CosmosModule(chainClient, signer, pipelineConfig)
    }

    /**
     * Connects to the blockchain.
     */
    suspend fun connect() {
        chainClient.connect()
        connected = true
    }

    /**
     * Disconnects from the blockchain.
     */
    suspend fun disconnect() {
        chainClient.disconnect()
        connected = false
    }

    /**
     * Returns true if the SDK is connected.
     */
    fun isConnected(): Boolean = connected && chainClient.isConnected()

    /**
     * Returns the SDK configuration.
     */
    fun getConfig(): SDKConfig = config

    /**
     * Returns the signer's address.
     */
    fun getAddress(): String = signer.getAddress()

    companion object {
        /**
         * Creates a new SeocheonSdk instance.
         * Validates config and creates the appropriate signing service.
         */
        suspend fun create(config: SDKConfig): SeocheonSdk {
            config.validate()

            val chainClient = OkHttpChainClient(config.chain)
            val signer = createSigner(config.signing)

            return SeocheonSdk(config, chainClient, signer)
        }

        /**
         * Creates a SeocheonSdk with a custom chain client (for testing).
         */
        fun createWithClient(config: SDKConfig, chainClient: ChainClient, signer: SigningService): SeocheonSdk {
            return SeocheonSdk(config, chainClient, signer)
        }

        private suspend fun createSigner(signing: SigningConfig): SigningService = when (signing.mode) {
            SigningMode.DIRECT -> DirectService(signing.mnemonic)
            SigningMode.KEYSTORE -> KeystoreService.create(signing.keystorePath, signing.passphraseEnv)
            SigningMode.VAULT -> VaultService.create(signing.vaultEndpoint, signing.keyName)
        }

        private fun parseGasPrice(gasPrice: String): Long {
            val numStr = gasPrice.replace(Regex("[^0-9]"), "")
            return numStr.toLongOrNull() ?: 250L
        }
    }
}
