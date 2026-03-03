using System.Security.Cryptography;
using Org.BouncyCastle.Asn1.Sec;
using Org.BouncyCastle.Crypto.Parameters;
using Org.BouncyCastle.Crypto.Signers;
using Org.BouncyCastle.Math;
using Seocheon.Sdk.Errors;

namespace Seocheon.Sdk.Internal.Crypto;

/// <summary>
/// secp256k1 private key operations using BouncyCastle.
/// </summary>
public sealed class PrivateKey
{
    private static readonly Org.BouncyCastle.Asn1.X9.X9ECParameters Curve =
        SecNamedCurves.GetByName("secp256k1");

    private static readonly ECDomainParameters Domain =
        new(Curve.Curve, Curve.G, Curve.N, Curve.H);

    private readonly ECPrivateKeyParameters _privateKey;

    private PrivateKey(byte[] keyBytes)
    {
        var d = new BigInteger(1, keyBytes);
        _privateKey = new ECPrivateKeyParameters(d, Domain);
    }

    /// <summary>
    /// Creates a PrivateKey from raw 32-byte key material.
    /// </summary>
    public static PrivateKey FromBytes(byte[] keyBytes)
    {
        if (keyBytes.Length != 32)
            throw SdkErrors.SigningFailed("Private key must be 32 bytes");
        return new PrivateKey(keyBytes);
    }

    /// <summary>
    /// Generates a new random private key.
    /// </summary>
    public static PrivateKey Generate()
    {
        var keyBytes = new byte[32];
        RandomNumberGenerator.Fill(keyBytes);
        return new PrivateKey(keyBytes);
    }

    /// <summary>
    /// Signs data with this private key (deterministic RFC 6979).
    /// Returns a 64-byte compact signature (r || s).
    /// </summary>
    public byte[] Sign(byte[] data)
    {
        var hash = SHA256.HashData(data);
        var signer = new ECDsaSigner(new Org.BouncyCastle.Crypto.Signers.HMacDsaKCalculator(
            new Org.BouncyCastle.Crypto.Digests.Sha256Digest()));
        signer.Init(true, _privateKey);

        var components = signer.GenerateSignature(hash);
        var r = components[0];
        var s = components[1];

        // Ensure low-S (BIP-62)
        var halfOrder = Domain.N.ShiftRight(1);
        if (s.CompareTo(halfOrder) > 0)
            s = Domain.N.Subtract(s);

        var rBytes = BigIntTo32Bytes(r);
        var sBytes = BigIntTo32Bytes(s);

        var signature = new byte[64];
        Buffer.BlockCopy(rBytes, 0, signature, 0, 32);
        Buffer.BlockCopy(sBytes, 0, signature, 32, 32);
        return signature;
    }

    /// <summary>
    /// Returns the compressed public key (33 bytes).
    /// </summary>
    public byte[] GetPubKey()
    {
        var q = Domain.G.Multiply(_privateKey.D).Normalize();
        return q.GetEncoded(true);
    }

    /// <summary>
    /// Returns the raw 32-byte private key.
    /// </summary>
    public byte[] ToBytes()
    {
        return BigIntTo32Bytes(_privateKey.D);
    }

    private static byte[] BigIntTo32Bytes(BigInteger value)
    {
        var bytes = value.ToByteArrayUnsigned();
        if (bytes.Length == 32) return bytes;
        if (bytes.Length > 32) return bytes[^32..];

        var padded = new byte[32];
        Buffer.BlockCopy(bytes, 0, padded, 32 - bytes.Length, bytes.Length);
        return padded;
    }
}
