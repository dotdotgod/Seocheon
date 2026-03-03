package com.seocheon.sdk

import com.seocheon.sdk.errors.SdkError
import com.seocheon.sdk.infrastructure.MockChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class SeocheonSdkTest {

    private val mockSigner = object : SigningService {
        override suspend fun sign(txBytes: ByteArray): ByteArray = ByteArray(64) { 0x01 }
        override fun getAddress(): String = "seocheon1mockaddress"
        override fun getPubKey(): ByteArray = ByteArray(33) { if (it == 0) 0x02 else 0x01 }
    }

    private fun createSdk(): SeocheonSdk {
        val config = SDKConfig(
            chain = ChainConfig("test-chain", "http://localhost:26657", "http://localhost:9090"),
            signing = SigningConfig(mode = SigningMode.DIRECT, mnemonic = "test mnemonic phrase"),
        )
        return SeocheonSdk.createWithClient(config, MockChainClient(), mockSigner)
    }

    @Test
    fun `SDK creates with valid config and mock client`() {
        val sdk = createSdk()
        assertNotNull(sdk)
    }

    @Test
    fun `SDK has all 5 modules`() {
        val sdk = createSdk()
        assertNotNull(sdk.activity)
        assertNotNull(sdk.node)
        assertNotNull(sdk.epoch)
        assertNotNull(sdk.rewards)
        assertNotNull(sdk.cosmos)
    }

    @Test
    fun `SDK connect and disconnect`() = runTest {
        val sdk = createSdk()
        assertFalse(sdk.isConnected())
        sdk.connect()
        assertTrue(sdk.isConnected())
        sdk.disconnect()
        assertFalse(sdk.isConnected())
    }

    @Test
    fun `SDK getAddress returns signer address`() {
        val sdk = createSdk()
        assertEquals("seocheon1mockaddress", sdk.getAddress())
    }

    @Test
    fun `SDK getConfig returns config`() {
        val sdk = createSdk()
        val config = sdk.getConfig()
        assertEquals("test-chain", config.chain.chainId)
    }
}
