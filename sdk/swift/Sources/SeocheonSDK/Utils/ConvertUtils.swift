import Foundation

/// Token denomination conversion utilities.
public enum ConvertUtils {
    /// Converts an amount between Seocheon token denominations.
    public static func convertDenom(amount: Int64, from: String, to: String) throws -> Int64 {
        let fromFactor = try denomFactor(from)
        let toFactor = try denomFactor(to)

        if from == to { return amount }

        // Convert to base (uppyeo) first, then to target
        let baseAmount = amount * fromFactor
        if toFactor == 1 { return baseAmount }
        return baseAmount / toFactor
    }

    /// Converts an uppyeo amount to a human-readable KKOT string.
    /// Example: 10000000000 uppyeo -> "1.0000000000"
    public static func formatKkot(_ uppyeoAmount: Int64) -> String {
        let intPart = uppyeoAmount / ChainConstants.uppyeoPerKkot
        var decPart = uppyeoAmount % ChainConstants.uppyeoPerKkot
        if decPart < 0 { decPart = -decPart }
        return String(format: "%lld.%010lld", intPart, decPart)
    }

    /// Parses a KKOT string to uppyeo amount.
    /// Example: "1.0000000000" -> 10000000000
    public static func parseKkot(_ kkot: String) throws -> Int64 {
        let parts = kkot.split(separator: ".", maxSplits: 1, omittingEmptySubsequences: false)
        guard parts.count <= 2 else {
            throw SDKError.invalidConfig("invalid kkot format: \(kkot)")
        }

        guard let intPart = Int64(parts[0]) else {
            throw SDKError.invalidConfig("invalid character in kkot integer part")
        }

        var decPart: Int64 = 0
        if parts.count == 2 {
            var decStr = String(parts[1])
            // Pad or truncate to 10 decimal places
            while decStr.count < 10 { decStr += "0" }
            decStr = String(decStr.prefix(10))
            guard let parsed = Int64(decStr) else {
                throw SDKError.invalidConfig("invalid character in kkot decimal part")
            }
            decPart = parsed
        }

        return intPart * ChainConstants.uppyeoPerKkot + decPart
    }

    /// Returns the denomination names in order.
    public static let denomNames: [String] = [
        ChainConstants.tokenBaseDenom,
        ChainConstants.tokenSalDenom,
        ChainConstants.tokenPiDenom,
        ChainConstants.tokenSumDenom,
        ChainConstants.tokenHonDenom,
        ChainConstants.tokenDisplayDenom,
    ]

    // MARK: - Private

    private static func denomFactor(_ denom: String) throws -> Int64 {
        switch denom {
        case ChainConstants.tokenBaseDenom: return 1
        case ChainConstants.tokenSalDenom: return ChainConstants.uppyeoPerSal
        case ChainConstants.tokenPiDenom: return ChainConstants.uppyeoPerPi
        case ChainConstants.tokenSumDenom: return ChainConstants.uppyeoPerSum
        case ChainConstants.tokenHonDenom: return ChainConstants.uppyeoPerHon
        case ChainConstants.tokenDisplayDenom: return ChainConstants.uppyeoPerKkot
        default:
            throw SDKError.invalidConfig("unknown denomination: \(denom) (supported: uppyeo, sal, pi, sum, hon, kkot)")
        }
    }
}
