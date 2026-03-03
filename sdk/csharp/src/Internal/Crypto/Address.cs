using System.Security.Cryptography;
using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Errors;

namespace Seocheon.Sdk.Internal.Crypto;

/// <summary>
/// Cosmos address derivation: pubkey → SHA256 → RIPEMD160 → bech32("seocheon").
/// </summary>
public static class Address
{
    /// <summary>
    /// Derives a bech32 address from a compressed public key (33 bytes).
    /// </summary>
    public static string FromPubKey(byte[] pubKey)
    {
        if (pubKey.Length != 33)
            throw SdkErrors.InvalidAddress("Public key must be 33 bytes (compressed)");

        // SHA256
        var sha256Hash = SHA256.HashData(pubKey);

        // RIPEMD160
        var ripemd = new Org.BouncyCastle.Crypto.Digests.RipeMD160Digest();
        ripemd.BlockUpdate(sha256Hash, 0, sha256Hash.Length);
        var ripemdHash = new byte[20];
        ripemd.DoFinal(ripemdHash, 0);

        // Bech32 encode
        return Bech32Encode(ChainConstants.AddressPrefix, ripemdHash);
    }

    /// <summary>
    /// Validates a bech32 address with the seocheon prefix.
    /// </summary>
    public static bool Validate(string address)
    {
        if (string.IsNullOrWhiteSpace(address))
            return false;

        try
        {
            var (hrp, data) = Bech32Decode(address);
            return hrp == ChainConstants.AddressPrefix && data.Length == 20;
        }
        catch
        {
            return false;
        }
    }

    // === Bech32 Implementation ===

    private const string Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l";

    private static readonly int[] Generator = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3];

    private static int Polymod(int[] values)
    {
        var chk = 1;
        foreach (var v in values)
        {
            var b = chk >> 25;
            chk = ((chk & 0x1ffffff) << 5) ^ v;
            for (var i = 0; i < 5; i++)
            {
                if (((b >> i) & 1) != 0)
                    chk ^= Generator[i];
            }
        }
        return chk;
    }

    private static int[] HrpExpand(string hrp)
    {
        var result = new int[hrp.Length * 2 + 1];
        for (var i = 0; i < hrp.Length; i++)
        {
            result[i] = hrp[i] >> 5;
            result[i + hrp.Length + 1] = hrp[i] & 31;
        }
        result[hrp.Length] = 0;
        return result;
    }

    private static bool VerifyChecksum(string hrp, int[] data)
    {
        var values = HrpExpand(hrp).Concat(data).ToArray();
        return Polymod(values) == 1;
    }

    private static int[] CreateChecksum(string hrp, int[] data)
    {
        var values = HrpExpand(hrp).Concat(data).Concat(new int[6]).ToArray();
        var polymod = Polymod(values) ^ 1;
        var checksum = new int[6];
        for (var i = 0; i < 6; i++)
            checksum[i] = (polymod >> (5 * (5 - i))) & 31;
        return checksum;
    }

    private static byte[] ConvertBits(byte[] data, int fromBits, int toBits, bool pad)
    {
        var acc = 0;
        var bits = 0;
        var result = new List<byte>();
        var maxv = (1 << toBits) - 1;

        foreach (var value in data)
        {
            acc = (acc << fromBits) | value;
            bits += fromBits;
            while (bits >= toBits)
            {
                bits -= toBits;
                result.Add((byte)((acc >> bits) & maxv));
            }
        }

        if (pad)
        {
            if (bits > 0)
                result.Add((byte)((acc << (toBits - bits)) & maxv));
        }
        else if (bits >= fromBits || ((acc << (toBits - bits)) & maxv) != 0)
        {
            throw new FormatException("Invalid padding in bech32 conversion");
        }

        return result.ToArray();
    }

    internal static string Bech32Encode(string hrp, byte[] data)
    {
        var converted = ConvertBits(data, 8, 5, true);
        var intData = converted.Select(b => (int)b).ToArray();
        var checksum = CreateChecksum(hrp, intData);
        var combined = intData.Concat(checksum).ToArray();

        var result = hrp + "1";
        foreach (var v in combined)
            result += Charset[v];

        return result;
    }

    internal static (string hrp, byte[] data) Bech32Decode(string bech)
    {
        bech = bech.ToLowerInvariant();
        var pos = bech.LastIndexOf('1');
        if (pos < 1 || pos + 7 > bech.Length)
            throw new FormatException("Invalid bech32 string");

        var hrp = bech[..pos];
        var dataChars = bech[(pos + 1)..];
        var data = new int[dataChars.Length];

        for (var i = 0; i < dataChars.Length; i++)
        {
            var idx = Charset.IndexOf(dataChars[i]);
            if (idx < 0)
                throw new FormatException($"Invalid bech32 character: {dataChars[i]}");
            data[i] = idx;
        }

        if (!VerifyChecksum(hrp, data))
            throw new FormatException("Invalid bech32 checksum");

        // Remove checksum (last 6 values)
        var payload = data[..^6].Select(v => (byte)v).ToArray();
        var decoded = ConvertBits(payload, 5, 8, false);

        return (hrp, decoded);
    }
}
