package com.seocheon.sdk.utils

import com.seocheon.sdk.constants.ChainConstants
import java.math.BigInteger

/**
 * Token denomination conversion utilities for the 6-stage system.
 * uppyeo(10^0) -> sal(10^2) -> pi(10^4) -> sum(10^6) -> hon(10^8) -> kkot(10^10)
 */
object ConvertUtils {

    private val DENOM_FACTORS: Map<String, BigInteger> = mapOf(
        ChainConstants.TOKEN_BASE_DENOM to BigInteger.ONE,
        ChainConstants.TOKEN_SAL_DENOM to BigInteger.valueOf(ChainConstants.UPPYEO_PER_SAL),
        ChainConstants.TOKEN_PI_DENOM to BigInteger.valueOf(ChainConstants.UPPYEO_PER_PI),
        ChainConstants.TOKEN_SUM_DENOM to BigInteger.valueOf(ChainConstants.UPPYEO_PER_SUM),
        ChainConstants.TOKEN_HON_DENOM to BigInteger.valueOf(ChainConstants.UPPYEO_PER_HON),
        ChainConstants.TOKEN_DISPLAY_DENOM to BigInteger.valueOf(ChainConstants.UPPYEO_PER_KKOT),
    )

    /**
     * Converts an amount between Seocheon token denominations.
     *
     * @param amount The amount to convert.
     * @param from Source denomination.
     * @param to Target denomination.
     * @return The converted amount.
     * @throws IllegalArgumentException if denomination is unknown.
     */
    fun convertDenom(amount: BigInteger, from: String, to: String): BigInteger {
        val fromFactor = denomFactor(from)
        val toFactor = denomFactor(to)

        if (from == to) return amount

        // Convert to base (uppyeo) first, then to target
        val baseAmount = amount * fromFactor
        if (toFactor == BigInteger.ONE) return baseAmount

        return baseAmount / toFactor
    }

    /**
     * Converts an uppyeo amount to a human-readable KKOT string.
     * Example: 10000000000 uppyeo -> "1.0000000000"
     */
    fun formatKkot(uppyeoAmount: Long): String {
        val intPart = uppyeoAmount / ChainConstants.UPPYEO_PER_KKOT
        var decPart = uppyeoAmount % ChainConstants.UPPYEO_PER_KKOT
        if (decPart < 0) decPart = -decPart
        return "%d.%010d".format(intPart, decPart)
    }

    /**
     * Parses a KKOT string to uppyeo amount.
     * Example: "1.0000000000" -> 10000000000
     */
    fun parseKkot(kkot: String): Long {
        val parts = kkot.split(".")
        require(parts.size <= 2) { "invalid kkot format: $kkot" }

        val intPart = parts[0].toLongOrNull()
            ?: throw IllegalArgumentException("invalid integer part in kkot: ${parts[0]}")

        var decPart = 0L
        if (parts.size == 2) {
            var dec = parts[1]
            // Pad or truncate to 10 decimal places
            while (dec.length < 10) dec += "0"
            dec = dec.substring(0, 10)
            decPart = dec.toLongOrNull()
                ?: throw IllegalArgumentException("invalid decimal part in kkot: ${parts[1]}")
        }

        return intPart * ChainConstants.UPPYEO_PER_KKOT + decPart
    }

    /**
     * Returns the base-unit multiplier for the given denomination.
     */
    fun denomFactor(denom: String): BigInteger {
        return DENOM_FACTORS[denom]
            ?: throw IllegalArgumentException(
                "unknown denomination: $denom (supported: uppyeo, sal, pi, sum, hon, kkot)"
            )
    }

    /**
     * Returns all supported denomination names.
     */
    fun supportedDenoms(): List<String> = DENOM_FACTORS.keys.toList()
}
