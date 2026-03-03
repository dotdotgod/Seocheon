using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Internal.Crypto;

namespace Seocheon.Sdk.Infrastructure.Signing;

/// <summary>
/// AES-256-GCM encrypted keystore signing service.
/// </summary>
public sealed class KeystoreService : ISigningService
{
    private readonly PrivateKey _key;
    private readonly string _address;
    private readonly byte[] _pubKey;

    /// <summary>
    /// Creates a KeystoreService by decrypting a keystore file.
    /// </summary>
    /// <param name="keystorePath">Path to the encrypted keystore JSON file.</param>
    /// <param name="passphrase">Passphrase to decrypt the keystore.</param>
    public KeystoreService(string keystorePath, string passphrase)
    {
        if (!File.Exists(keystorePath))
            throw SdkErrors.InvalidConfig($"Keystore file not found: {keystorePath}");

        var keyBytes = DecryptKeystore(keystorePath, passphrase);
        _key = PrivateKey.FromBytes(keyBytes);
        _pubKey = _key.GetPubKey();
        _address = Address.FromPubKey(_pubKey);
    }

    /// <inheritdoc />
    public byte[] Sign(byte[] data) => _key.Sign(data);

    /// <inheritdoc />
    public string GetAddress() => _address;

    /// <inheritdoc />
    public byte[] GetPubKey() => _pubKey;

    /// <summary>
    /// Creates an encrypted keystore file from a private key.
    /// </summary>
    public static void CreateKeystore(string keystorePath, byte[] privateKey, string passphrase)
    {
        var salt = RandomNumberGenerator.GetBytes(32);
        var derivedKey = DeriveKey(passphrase, salt);

        var nonce = RandomNumberGenerator.GetBytes(12);
        var tag = new byte[16];

        using var aes = new AesGcm(derivedKey, 16);
        var ciphertext = new byte[privateKey.Length];
        aes.Encrypt(nonce, privateKey, ciphertext, tag);

        var keystore = new
        {
            version = 1,
            crypto = new
            {
                cipher = "aes-256-gcm",
                salt = Convert.ToBase64String(salt),
                nonce = Convert.ToBase64String(nonce),
                tag = Convert.ToBase64String(tag),
                ciphertext = Convert.ToBase64String(ciphertext)
            }
        };

        var json = JsonSerializer.Serialize(keystore, new JsonSerializerOptions { WriteIndented = true });
        File.WriteAllText(keystorePath, json);
    }

    private static byte[] DecryptKeystore(string path, string passphrase)
    {
        try
        {
            var json = File.ReadAllText(path);
            using var doc = JsonDocument.Parse(json);
            var crypto = doc.RootElement.GetProperty("crypto");

            var salt = Convert.FromBase64String(crypto.GetProperty("salt").GetString()!);
            var nonce = Convert.FromBase64String(crypto.GetProperty("nonce").GetString()!);
            var tag = Convert.FromBase64String(crypto.GetProperty("tag").GetString()!);
            var ciphertext = Convert.FromBase64String(crypto.GetProperty("ciphertext").GetString()!);

            var derivedKey = DeriveKey(passphrase, salt);

            using var aes = new AesGcm(derivedKey, 16);
            var plaintext = new byte[ciphertext.Length];
            aes.Decrypt(nonce, ciphertext, tag, plaintext);

            return plaintext;
        }
        catch (SdkException)
        {
            throw;
        }
        catch (Exception ex)
        {
            throw SdkErrors.SigningFailed($"Failed to decrypt keystore: {ex.Message}", ex);
        }
    }

    private static byte[] DeriveKey(string passphrase, byte[] salt)
    {
        using var pbkdf2 = new Rfc2898DeriveBytes(
            Encoding.UTF8.GetBytes(passphrase),
            salt,
            100_000,
            HashAlgorithmName.SHA256);
        return pbkdf2.GetBytes(32);
    }
}
