using System.Text;
using System.Text.Json;
using Seocheon.Sdk.Errors;

namespace Seocheon.Sdk.Infrastructure.Signing;

/// <summary>
/// External vault server signing service for production use.
/// Delegates signing to a remote vault via HTTP.
/// </summary>
public sealed class VaultService : ISigningService
{
    private readonly HttpClient _http;
    private readonly string _endpoint;
    private readonly string _keyName;
    private string _address = "";
    private byte[] _pubKey = [];

    /// <summary>
    /// Creates a VaultService connected to the given endpoint.
    /// </summary>
    public VaultService(string endpoint, string keyName)
    {
        if (string.IsNullOrWhiteSpace(endpoint))
            throw SdkErrors.InvalidConfig("Vault endpoint is required");
        if (string.IsNullOrWhiteSpace(keyName))
            throw SdkErrors.InvalidConfig("Vault key name is required");

        _endpoint = endpoint.TrimEnd('/');
        _keyName = keyName;
        _http = new HttpClient { Timeout = TimeSpan.FromSeconds(30) };

        Initialize();
    }

    /// <inheritdoc />
    public byte[] Sign(byte[] data)
    {
        try
        {
            var payload = JsonSerializer.Serialize(new
            {
                key_name = _keyName,
                data = Convert.ToBase64String(data)
            });

            var content = new StringContent(payload, Encoding.UTF8, "application/json");
            var response = _http.PostAsync($"{_endpoint}/sign", content).GetAwaiter().GetResult();

            if (!response.IsSuccessStatusCode)
                throw SdkErrors.SigningFailed($"Vault sign request failed: {response.StatusCode}");

            var body = response.Content.ReadAsStringAsync().GetAwaiter().GetResult();
            using var doc = JsonDocument.Parse(body);
            var sig = doc.RootElement.GetProperty("signature").GetString()!;
            return Convert.FromBase64String(sig);
        }
        catch (SdkException)
        {
            throw;
        }
        catch (Exception ex)
        {
            throw SdkErrors.SigningFailed($"Vault sign failed: {ex.Message}", ex);
        }
    }

    /// <inheritdoc />
    public string GetAddress() => _address;

    /// <inheritdoc />
    public byte[] GetPubKey() => _pubKey;

    private void Initialize()
    {
        try
        {
            var response = _http.GetAsync($"{_endpoint}/keys/{_keyName}").GetAwaiter().GetResult();

            if (!response.IsSuccessStatusCode)
                throw SdkErrors.SigningFailed($"Vault key lookup failed: {response.StatusCode}");

            var body = response.Content.ReadAsStringAsync().GetAwaiter().GetResult();
            using var doc = JsonDocument.Parse(body);

            _address = doc.RootElement.GetProperty("address").GetString()!;
            var pubKeyBase64 = doc.RootElement.GetProperty("public_key").GetString()!;
            _pubKey = Convert.FromBase64String(pubKeyBase64);
        }
        catch (SdkException)
        {
            throw;
        }
        catch (Exception ex)
        {
            throw SdkErrors.SigningFailed($"Vault initialization failed: {ex.Message}", ex);
        }
    }
}
