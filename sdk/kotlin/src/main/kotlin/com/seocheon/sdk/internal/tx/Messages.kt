package com.seocheon.sdk.internal.tx

/**
 * Interface for encoding a specific message type into protobuf bytes.
 */
interface MessageEncoder {
    /** Returns the protobuf type URL. */
    fun typeUrl(): String
    /** Returns the protobuf-encoded message bytes. */
    fun encode(): ByteArray
}

/**
 * Encodes /seocheon.activity.v1.MsgSubmitActivity.
 * Fields: submitter(1, string), activity_hash(2, string), content_uri(3, string)
 */
data class MsgSubmitActivity(
    val submitter: String,
    val activityHash: String,
    val contentUri: String,
) : MessageEncoder {
    override fun typeUrl(): String = "/seocheon.activity.v1.MsgSubmitActivity"
    override fun encode(): ByteArray = Protobuf.concatBytes(
        Protobuf.encodeFieldString(1, submitter),
        Protobuf.encodeFieldString(2, activityHash),
        Protobuf.encodeFieldString(3, contentUri),
    )
}

/**
 * Encodes /seocheon.node.v1.MsgWithdrawNodeCommission.
 * Fields: operator(1, string)
 */
data class MsgWithdrawNodeCommission(
    val operator: String,
) : MessageEncoder {
    override fun typeUrl(): String = "/seocheon.node.v1.MsgWithdrawNodeCommission"
    override fun encode(): ByteArray = Protobuf.concatBytes(
        Protobuf.encodeFieldString(1, operator),
    )
}

/**
 * Represents cosmos.base.v1beta1.Coin.
 * Fields: denom(1, string), amount(2, string)
 */
data class Coin(
    val denom: String,
    val amount: String,
) {
    fun encode(): ByteArray = Protobuf.concatBytes(
        Protobuf.encodeFieldString(1, denom),
        Protobuf.encodeFieldString(2, amount),
    )
}

/**
 * Encodes /cosmos.bank.v1beta1.MsgSend.
 * Fields: from_address(1, string), to_address(2, string), amount(3, repeated Coin)
 */
data class MsgSend(
    val fromAddress: String,
    val toAddress: String,
    val amount: List<Coin>,
) : MessageEncoder {
    override fun typeUrl(): String = "/cosmos.bank.v1beta1.MsgSend"
    override fun encode(): ByteArray {
        val parts = mutableListOf(
            Protobuf.encodeFieldString(1, fromAddress),
            Protobuf.encodeFieldString(2, toAddress),
        )
        for (coin in amount) {
            parts.add(Protobuf.encodeFieldBytes(3, coin.encode()))
        }
        return Protobuf.concatBytes(*parts.toTypedArray())
    }
}

/**
 * Encodes /seocheon.node.v1.MsgConfirmDelegation.
 * Fields: delegator_address(1, string), validator_address(2, string)
 */
data class MsgConfirmDelegation(
    val delegatorAddress: String,
    val validatorAddress: String,
) : MessageEncoder {
    override fun typeUrl(): String = "/seocheon.node.v1.MsgConfirmDelegation"
    override fun encode(): ByteArray = Protobuf.concatBytes(
        Protobuf.encodeFieldString(1, delegatorAddress),
        Protobuf.encodeFieldString(2, validatorAddress),
    )
}
