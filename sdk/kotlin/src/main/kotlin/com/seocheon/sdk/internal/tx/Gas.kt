package com.seocheon.sdk.internal.tx

import com.seocheon.sdk.constants.ChainConstants

/**
 * Gas estimation and fee calculation utilities.
 */
object Gas {

    /** Returns the default gas limit for a given message type URL. */
    fun defaultGasForMessage(typeUrl: String): Long = when (typeUrl) {
        "/seocheon.activity.v1.MsgSubmitActivity" -> ChainConstants.GAS_SUBMIT_ACTIVITY
        "/seocheon.node.v1.MsgWithdrawNodeCommission" -> ChainConstants.GAS_WITHDRAW
        "/cosmos.bank.v1beta1.MsgSend" -> ChainConstants.GAS_SEND
        else -> ChainConstants.GAS_FALLBACK
    }

    /** Computes the fee amount from gas limit and gas price. */
    fun calculateFee(gasLimit: Long, gasPrice: Long = ChainConstants.GAS_PRICE): Long =
        gasLimit * gasPrice
}
