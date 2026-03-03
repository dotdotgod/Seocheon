package com.seocheon.sdk.internal.tx

import com.seocheon.sdk.constants.ChainConstants
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class GasTest {

    @Test
    fun `defaultGasForMessage returns correct gas for SubmitActivity`() {
        assertEquals(
            ChainConstants.GAS_SUBMIT_ACTIVITY,
            Gas.defaultGasForMessage("/seocheon.activity.v1.MsgSubmitActivity")
        )
    }

    @Test
    fun `defaultGasForMessage returns correct gas for Withdraw`() {
        assertEquals(
            ChainConstants.GAS_WITHDRAW,
            Gas.defaultGasForMessage("/seocheon.node.v1.MsgWithdrawNodeCommission")
        )
    }

    @Test
    fun `defaultGasForMessage returns correct gas for Send`() {
        assertEquals(
            ChainConstants.GAS_SEND,
            Gas.defaultGasForMessage("/cosmos.bank.v1beta1.MsgSend")
        )
    }

    @Test
    fun `defaultGasForMessage returns fallback for unknown type`() {
        assertEquals(
            ChainConstants.GAS_FALLBACK,
            Gas.defaultGasForMessage("/unknown.MsgType")
        )
    }

    @Test
    fun `calculateFee computes correctly`() {
        assertEquals(50000000L, Gas.calculateFee(200000, 250))
    }

    @Test
    fun `calculateFee uses default gas price`() {
        assertEquals(200000 * ChainConstants.GAS_PRICE, Gas.calculateFee(200000))
    }
}
