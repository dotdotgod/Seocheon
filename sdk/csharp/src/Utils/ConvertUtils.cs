using System.Globalization;
using Seocheon.Sdk.Constants;

namespace Seocheon.Sdk.Utils;

/// <summary>
/// 6-stage denomination conversion utilities.
/// uppyeo(10^0) → sal(10^2) → pi(10^4) → sum(10^6) → hon(10^8) → kkot(10^10)
/// </summary>
public static class ConvertUtils
{
    private static readonly Dictionary<string, int> DenomIndex = new(StringComparer.OrdinalIgnoreCase)
    {
        ["uppyeo"] = 0,
        ["sal"] = 1,
        ["pi"] = 2,
        ["sum"] = 3,
        ["hon"] = 4,
        ["kkot"] = 5,
    };

    /// <summary>
    /// Converts an amount between two denominations.
    /// </summary>
    /// <param name="amount">Amount in source denomination.</param>
    /// <param name="from">Source denomination name.</param>
    /// <param name="to">Target denomination name.</param>
    /// <returns>Amount in target denomination.</returns>
    /// <exception cref="ArgumentException">If denomination name is invalid.</exception>
    public static long ConvertDenom(long amount, string from, string to)
    {
        if (!DenomIndex.TryGetValue(from, out var fromIdx))
            throw new ArgumentException($"Unknown denomination: {from}", nameof(from));
        if (!DenomIndex.TryGetValue(to, out var toIdx))
            throw new ArgumentException($"Unknown denomination: {to}", nameof(to));

        if (fromIdx == toIdx)
            return amount;

        var fromFactor = ChainConstants.DenomFactors[fromIdx];
        var toFactor = ChainConstants.DenomFactors[toIdx];

        // Convert to uppyeo first, then to target
        var uppyeo = amount * fromFactor;
        return uppyeo / toFactor;
    }

    /// <summary>
    /// Formats an uppyeo amount as a human-readable KKOT string with 10 decimal places.
    /// Example: 10_000_000_000 → "1.0000000000"
    /// </summary>
    public static string FormatKkot(long uppyeoAmount)
    {
        var wholePart = uppyeoAmount / ChainConstants.UppyeoPerKkot;
        var fractionalPart = Math.Abs(uppyeoAmount % ChainConstants.UppyeoPerKkot);
        var sign = uppyeoAmount < 0 && wholePart == 0 ? "-" : "";
        return $"{sign}{wholePart}.{fractionalPart:D10}";
    }

    /// <summary>
    /// Parses a KKOT string (e.g., "1.5") to uppyeo amount.
    /// </summary>
    /// <exception cref="FormatException">If the string format is invalid.</exception>
    public static long ParseKkot(string kkot)
    {
        if (string.IsNullOrWhiteSpace(kkot))
            throw new FormatException("Empty KKOT string");

        var parts = kkot.Split('.');
        var wholePart = long.Parse(parts[0], CultureInfo.InvariantCulture);
        long fractionalPart = 0;

        if (parts.Length == 2)
        {
            var fracStr = parts[1].PadRight(10, '0');
            if (fracStr.Length > 10)
                fracStr = fracStr[..10];
            fractionalPart = long.Parse(fracStr, CultureInfo.InvariantCulture);
        }
        else if (parts.Length > 2)
        {
            throw new FormatException($"Invalid KKOT format: {kkot}");
        }

        var sign = wholePart < 0 || kkot.StartsWith('-') ? -1L : 1L;
        return sign * (Math.Abs(wholePart) * ChainConstants.UppyeoPerKkot + fractionalPart);
    }

    /// <summary>
    /// Formats an amount in the given denomination to a human-readable string.
    /// Example: FormatHumanReadable(1500000000, "uppyeo") → "0.1500000000 KKOT"
    /// </summary>
    public static string FormatHumanReadable(long amount, string denom = "uppyeo")
    {
        var uppyeo = ConvertDenom(amount, denom, "uppyeo");
        return $"{FormatKkot(uppyeo)} KKOT";
    }

    /// <summary>
    /// Returns all valid denomination names.
    /// </summary>
    public static IReadOnlyList<string> GetDenomNames() => ChainConstants.DenomNames;
}
