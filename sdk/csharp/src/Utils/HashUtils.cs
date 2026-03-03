using System.Security.Cryptography;
using System.Text.RegularExpressions;
using Seocheon.Sdk.Constants;

namespace Seocheon.Sdk.Utils;

/// <summary>
/// SHA-256 hash utilities for activity hash validation and computation.
/// </summary>
public static partial class HashUtils
{
    [GeneratedRegex("^[0-9a-f]{64}$")]
    private static partial Regex HexPattern();

    /// <summary>
    /// Validates that a string is a valid 64-character lowercase hex SHA-256 hash.
    /// </summary>
    public static bool ValidateActivityHash(string hash)
    {
        if (string.IsNullOrEmpty(hash))
            return false;

        if (hash.Length != ChainConstants.ActivityHashLength)
            return false;

        return HexPattern().IsMatch(hash);
    }

    /// <summary>
    /// Computes the SHA-256 hash of the given data and returns the lowercase hex string.
    /// </summary>
    public static string ComputeActivityHash(byte[] data)
    {
        var hashBytes = SHA256.HashData(data);
        return Convert.ToHexStringLower(hashBytes);
    }

    /// <summary>
    /// Computes the SHA-256 hash of the given UTF-8 string.
    /// </summary>
    public static string ComputeActivityHash(string data)
    {
        return ComputeActivityHash(System.Text.Encoding.UTF8.GetBytes(data));
    }
}
