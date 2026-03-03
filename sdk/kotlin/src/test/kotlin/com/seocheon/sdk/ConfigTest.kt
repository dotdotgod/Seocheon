package com.seocheon.sdk

import com.seocheon.sdk.errors.SdkError
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class ConfigTest {

    private fun validConfig(
        mode: SigningMode = SigningMode.DIRECT,
        mnemonic: String = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
    ) = SDKConfig(
        chain = ChainConfig("test-chain", "http://localhost:26657", "http://localhost:9090"),
        signing = SigningConfig(mode = mode, mnemonic = mnemonic),
    )

    @Test
    fun `valid direct config passes validation`() {
        assertDoesNotThrow { validConfig().validate() }
    }

    @Test
    fun `empty chain_id fails validation`() {
        val config = SDKConfig(
            chain = ChainConfig("", "http://localhost:26657", "http://localhost:9090"),
            signing = SigningConfig(mode = SigningMode.DIRECT, mnemonic = "test mnemonic phrase"),
        )
        assertThrows(SdkError.InvalidConfig::class.java) { config.validate() }
    }

    @Test
    fun `empty rpc_endpoint fails validation`() {
        val config = SDKConfig(
            chain = ChainConfig("test-chain", "", "http://localhost:9090"),
            signing = SigningConfig(mode = SigningMode.DIRECT, mnemonic = "test"),
        )
        assertThrows(SdkError.InvalidConfig::class.java) { config.validate() }
    }

    @Test
    fun `empty grpc_endpoint fails validation`() {
        val config = SDKConfig(
            chain = ChainConfig("test-chain", "http://localhost:26657", ""),
            signing = SigningConfig(mode = SigningMode.DIRECT, mnemonic = "test"),
        )
        assertThrows(SdkError.InvalidConfig::class.java) { config.validate() }
    }

    @Test
    fun `vault mode requires endpoint and key_name`() {
        val config = SDKConfig(
            chain = ChainConfig("test-chain", "http://localhost:26657", "http://localhost:9090"),
            signing = SigningConfig(mode = SigningMode.VAULT),
        )
        assertThrows(SdkError.InvalidConfig::class.java) { config.validate() }
    }

    @Test
    fun `keystore mode requires path and passphrase`() {
        val config = SDKConfig(
            chain = ChainConfig("test-chain", "http://localhost:26657", "http://localhost:9090"),
            signing = SigningConfig(mode = SigningMode.KEYSTORE),
        )
        assertThrows(SdkError.InvalidConfig::class.java) { config.validate() }
    }

    @Test
    fun `direct mode requires mnemonic`() {
        val config = SDKConfig(
            chain = ChainConfig("test-chain", "http://localhost:26657", "http://localhost:9090"),
            signing = SigningConfig(mode = SigningMode.DIRECT),
        )
        assertThrows(SdkError.InvalidConfig::class.java) { config.validate() }
    }

    @Test
    fun `default config factory sets defaults`() {
        val config = SDKConfig.default(
            "test-chain",
            "http://localhost:26657",
            "http://localhost:9090",
            SigningConfig(mode = SigningMode.DIRECT, mnemonic = "test"),
        )
        assertEquals("sync", config.tx.broadcastMode)
        assertEquals(30000L, config.tx.confirmTimeoutMs)
        assertEquals(1000L, config.tx.confirmPollIntervalMs)
    }
}
