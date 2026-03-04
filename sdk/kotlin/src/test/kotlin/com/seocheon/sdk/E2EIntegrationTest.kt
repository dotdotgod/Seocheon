package com.seocheon.sdk

import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Assumptions.assumeTrue
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Tag
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance

/**
 * E2E integration tests for the Seocheon Kotlin SDK.
 *
 * Skip conditions:
 *   - SEOCHEON_GRPC not set
 *   - SEOCHEON_MNEMONIC not set
 *
 * Run with:
 *   ./gradlew e2eTest
 */
@Tag("e2e")
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class E2EIntegrationTest {

    companion object {
        val GRPC: String = System.getenv("SEOCHEON_GRPC") ?: ""
        val MNEMONIC: String = System.getenv("SEOCHEON_MNEMONIC") ?: ""
        val RPC: String = System.getenv("SEOCHEON_RPC") ?: "http://localhost:26657"
        val CHAIN_ID: String = System.getenv("SEOCHEON_CHAIN_ID") ?: "seocheon-e2e"
    }

    private lateinit var sdk: SeocheonSdk

    private fun skipIfMissing() {
        assumeTrue(GRPC.isNotBlank(), "E2E 스킵: SEOCHEON_GRPC 미설정")
        assumeTrue(MNEMONIC.isNotBlank(), "E2E 스킵: SEOCHEON_MNEMONIC 미설정")
    }

    private fun buildConfig(): SDKConfig = SDKConfig.default(
        chainId = CHAIN_ID,
        rpcEndpoint = RPC,
        grpcEndpoint = GRPC,
        signing = SigningConfig(
            mode = SigningMode.DIRECT,
            mnemonic = MNEMONIC,
        ),
    )

    @BeforeAll
    fun setUp(): Unit = runBlocking {
        skipIfMissing()
        sdk = SeocheonSdk.create(buildConfig())
        sdk.connect()
    }

    @Test
    fun `connect - isConnected가 true를 반환해야 한다`() {
        skipIfMissing()
        assertTrue(sdk.isConnected(), "Connect 후 isConnected() = false")
    }

    @Test
    fun `getBlockInfo - 양수 blockHeight를 반환해야 한다`(): Unit = runBlocking {
        skipIfMissing()
        val block = sdk.cosmos.getBlockInfo()
        assertTrue(block.blockHeight > 0, "블록 높이가 양수여야 함: ${block.blockHeight}")
        println("최신 블록: height=${block.blockHeight} chainId=${block.chainId}")
    }

    @Test
    fun `node search - x-node 엔드포인트가 응답해야 한다`(): Unit = runBlocking {
        skipIfMissing()
        val resp = sdk.node.search(limit = 10)
        assertNotNull(resp)
        println("x/node 조회 성공: total=${resp.totalCount}")
    }

    @Test
    fun `epoch getInfo - 에포크 정보를 반환해야 한다`(): Unit = runBlocking {
        skipIfMissing()
        val info = sdk.epoch.getInfo()
        assertTrue(info.blockHeight > 0, "에포크 블록 높이가 양수여야 함: ${info.blockHeight}")
        println("에포크: epoch=${info.epochNumber} window=${info.windowNumber} height=${info.blockHeight}")
    }
}
