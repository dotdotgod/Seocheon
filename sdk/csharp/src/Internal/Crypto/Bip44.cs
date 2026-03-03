using NBitcoin;
using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Errors;

namespace Seocheon.Sdk.Internal.Crypto;

/// <summary>
/// BIP39 mnemonic to BIP44 key derivation for Cosmos (coin type 118).
/// Path: m/44'/118'/0'/0/0
/// </summary>
public static class Bip44
{
    /// <summary>
    /// Derives a secp256k1 private key from a BIP39 mnemonic.
    /// </summary>
    /// <param name="mnemonic">BIP39 mnemonic phrase (12 or 24 words).</param>
    /// <returns>32-byte private key.</returns>
    public static PrivateKey DeriveKey(string mnemonic)
    {
        if (string.IsNullOrWhiteSpace(mnemonic))
            throw SdkErrors.SigningFailed("Mnemonic is required");

        try
        {
            var mnemonicObj = new Mnemonic(mnemonic, Wordlist.English);
            var masterKey = mnemonicObj.DeriveExtKey();

            // BIP44 path: m/44'/118'/0'/0/0
            var derivedKey = masterKey
                .Derive(new KeyPath(ChainConstants.Bip44Path));

            var privateKeyBytes = derivedKey.PrivateKey.ToBytes();
            return PrivateKey.FromBytes(privateKeyBytes);
        }
        catch (SdkException)
        {
            throw;
        }
        catch (Exception ex)
        {
            throw SdkErrors.SigningFailed($"BIP44 key derivation failed: {ex.Message}", ex);
        }
    }

    /// <summary>
    /// Validates a BIP39 mnemonic phrase.
    /// </summary>
    public static bool ValidateMnemonic(string mnemonic)
    {
        if (string.IsNullOrWhiteSpace(mnemonic))
            return false;

        try
        {
            _ = new Mnemonic(mnemonic, Wordlist.English);
            return true;
        }
        catch
        {
            return false;
        }
    }

    /// <summary>
    /// Generates a new random BIP39 mnemonic (24 words).
    /// </summary>
    public static string GenerateMnemonic()
    {
        var mnemonic = new Mnemonic(Wordlist.English, WordCount.TwentyFour);
        return mnemonic.ToString();
    }
}
