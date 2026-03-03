using Seocheon.Sdk.Internal.Crypto;

namespace Seocheon.Sdk.Infrastructure.Signing;

/// <summary>
/// Mnemonic-based signing service for development and testing.
/// Derives key via BIP44 path m/44'/118'/0'/0/0.
/// </summary>
public sealed class DirectService : ISigningService
{
    private readonly PrivateKey _key;
    private readonly string _address;
    private readonly byte[] _pubKey;

    /// <summary>
    /// Creates a DirectService from a BIP39 mnemonic phrase.
    /// </summary>
    public DirectService(string mnemonic)
    {
        _key = Bip44.DeriveKey(mnemonic);
        _pubKey = _key.GetPubKey();
        _address = Address.FromPubKey(_pubKey);
    }

    /// <inheritdoc />
    public byte[] Sign(byte[] data) => _key.Sign(data);

    /// <inheritdoc />
    public string GetAddress() => _address;

    /// <inheritdoc />
    public byte[] GetPubKey() => _pubKey;
}
