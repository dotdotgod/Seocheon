package com.seocheon.sdk.errors

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class SdkErrorTest {

    @Test
    fun `error message includes code`() {
        val err = SdkError.NotConnected()
        assertEquals("[9000] SDK is not connected to chain", err.message)
    }

    @Test
    fun `error code is accessible`() {
        val err = SdkError.BroadcastFailed()
        assertEquals(9001u, err.code)
    }

    @Test
    fun `error wraps cause`() {
        val cause = RuntimeException("connection refused")
        val err = SdkError.QueryFailed(cause = cause)
        assertSame(cause, err.cause)
    }

    @Test
    fun `fromAbciCode maps known codes`() {
        assertTrue(SdkError.fromAbciCode(1101u) is SdkError.NodeNotFound)
        assertTrue(SdkError.fromAbciCode(1200u) is SdkError.SubmitterNotRegistered)
        assertTrue(SdkError.fromAbciCode(1203u) is SdkError.QuotaExceeded)
    }

    @Test
    fun `fromAbciCode returns ChainError for unknown code`() {
        val err = SdkError.fromAbciCode(9999u)
        assertTrue(err is SdkError.ChainError)
        assertEquals(9999u, err.code)
    }

    @Test
    fun `all error types are SdkError instances`() {
        val errors: List<SdkError> = listOf(
            SdkError.NotConnected(),
            SdkError.BroadcastFailed(),
            SdkError.TxTimeout(),
            SdkError.TxNotFound(),
            SdkError.SigningFailed(),
            SdkError.InvalidConfig(),
            SdkError.QueryFailed(),
            SdkError.InvalidAddress(),
            SdkError.NodeNotFound(),
            SdkError.NodeAlreadyExists(),
            SdkError.UnauthorizedOperator(),
            SdkError.UnauthorizedAgentMsg(),
            SdkError.SubmitterNotRegistered(),
            SdkError.NodeNotEligible(),
            SdkError.DuplicateActivityHash(),
            SdkError.QuotaExceeded(),
            SdkError.InvalidActivityHash(),
            SdkError.InvalidContentURI(),
        )
        errors.forEach { assertTrue(it is Exception) }
        assertEquals(18, errors.size)
    }
}
