namespace Seocheon.Sdk.Infrastructure.Signing;

/// <summary>
/// Signing service interface for transaction authorization.
/// </summary>
public interface ISigningService
{
    /// <summary>
    /// Signs the given bytes and returns a 64-byte compact signature.
    /// </summary>
    byte[] Sign(byte[] data);

    /// <summary>
    /// Returns the signer's bech32 address.
    /// </summary>
    string GetAddress();

    /// <summary>
    /// Returns the signer's compressed public key (33 bytes).
    /// </summary>
    byte[] GetPubKey();
}
