package com.seocheon.sdk.errors

/**
 * Sealed class hierarchy for Seocheon SDK errors with error codes.
 */
sealed class SdkError(
    val code: UInt,
    val errorMessage: String,
    override val cause: Throwable? = null,
) : Exception(if (code > 0u) "[$code] $errorMessage" else errorMessage, cause) {

    // SDK-level errors (9000 series)
    class NotConnected(cause: Throwable? = null) :
        SdkError(9000u, "SDK is not connected to chain", cause)

    class BroadcastFailed(cause: Throwable? = null) :
        SdkError(9001u, "transaction broadcast failed", cause)

    class TxTimeout(cause: Throwable? = null) :
        SdkError(9002u, "transaction confirmation timeout", cause)

    class TxNotFound(cause: Throwable? = null) :
        SdkError(9003u, "transaction not found", cause)

    class SigningFailed(cause: Throwable? = null) :
        SdkError(9004u, "transaction signing failed", cause)

    class InvalidConfig(detail: String = "invalid SDK configuration", cause: Throwable? = null) :
        SdkError(9005u, detail, cause)

    class QueryFailed(detail: String = "chain query failed", cause: Throwable? = null) :
        SdkError(9006u, detail, cause)

    class InvalidAddress(cause: Throwable? = null) :
        SdkError(9007u, "invalid address format", cause)

    // x/node errors (1100 series)
    class NodeNotFound(cause: Throwable? = null) :
        SdkError(1101u, "node not found", cause)

    class NodeAlreadyExists(cause: Throwable? = null) :
        SdkError(1102u, "node already exists for this operator", cause)

    class UnauthorizedOperator(cause: Throwable? = null) :
        SdkError(1108u, "unauthorized: signer is not the node operator", cause)

    class UnauthorizedAgentMsg(cause: Throwable? = null) :
        SdkError(1109u, "agent address is not authorized for this message type", cause)

    // x/activity errors (1200 series)
    class SubmitterNotRegistered(cause: Throwable? = null) :
        SdkError(1200u, "submitter agent address is not registered to any node", cause)

    class NodeNotEligible(cause: Throwable? = null) :
        SdkError(1201u, "node is not in an eligible status (REGISTERED or ACTIVE)", cause)

    class DuplicateActivityHash(cause: Throwable? = null) :
        SdkError(1202u, "duplicate (activity_hash, content_uri) pair already exists", cause)

    class QuotaExceeded(cause: Throwable? = null) :
        SdkError(1203u, "activity quota exceeded for this epoch", cause)

    class InvalidActivityHash(cause: Throwable? = null) :
        SdkError(1204u, "activity hash must be exactly 64 hex characters (32 bytes)", cause)

    class InvalidContentURI(cause: Throwable? = null) :
        SdkError(1205u, "content URI must not be empty", cause)

    // Generic chain error
    class ChainError(code: UInt, detail: String = "chain error code $code", cause: Throwable? = null) :
        SdkError(code, detail, cause)

    companion object {
        /**
         * Maps an ABCI error code to its corresponding SdkError.
         */
        fun fromAbciCode(code: UInt): SdkError = when (code) {
            9000u -> NotConnected()
            9001u -> BroadcastFailed()
            9002u -> TxTimeout()
            9003u -> TxNotFound()
            9004u -> SigningFailed()
            9005u -> InvalidConfig()
            9006u -> QueryFailed()
            9007u -> InvalidAddress()
            1101u -> NodeNotFound()
            1102u -> NodeAlreadyExists()
            1108u -> UnauthorizedOperator()
            1109u -> UnauthorizedAgentMsg()
            1200u -> SubmitterNotRegistered()
            1201u -> NodeNotEligible()
            1202u -> DuplicateActivityHash()
            1203u -> QuotaExceeded()
            1204u -> InvalidActivityHash()
            1205u -> InvalidContentURI()
            else -> ChainError(code)
        }
    }
}
