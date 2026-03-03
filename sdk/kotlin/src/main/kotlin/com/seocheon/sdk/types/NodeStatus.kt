package com.seocheon.sdk.types

/**
 * Represents the status of a registered node.
 */
enum class NodeStatus(val value: String) {
    UNSPECIFIED("UNSPECIFIED"),
    REGISTERED("REGISTERED"),
    ACTIVE("ACTIVE"),
    INACTIVE("INACTIVE"),
    JAILED("JAILED");

    companion object {
        /**
         * Converts a proto enum integer to NodeStatus.
         */
        fun fromInt(status: Int): NodeStatus = when (status) {
            1 -> REGISTERED
            2 -> ACTIVE
            3 -> INACTIVE
            4 -> JAILED
            else -> UNSPECIFIED
        }

        /**
         * Converts a string to NodeStatus, case-insensitive.
         */
        fun fromString(status: String): NodeStatus =
            entries.find { it.value.equals(status, ignoreCase = true) } ?: UNSPECIFIED
    }
}
