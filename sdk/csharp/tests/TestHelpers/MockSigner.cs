using Seocheon.Sdk.Infrastructure.Signing;

namespace Seocheon.Sdk.Tests.TestHelpers;

/// <summary>
/// Mock signing service for testing.
/// </summary>
public class MockSigner : ISigningService
{
    public string Address { get; set; } = "seocheon1testaddress0000000000000000000000test";
    public byte[] PubKeyBytes { get; set; } = new byte[33];

    public MockSigner()
    {
        PubKeyBytes[0] = 0x02; // Compressed pubkey prefix
        for (var i = 1; i < 33; i++)
            PubKeyBytes[i] = (byte)(i % 256);
    }

    public byte[] Sign(byte[] data) => new byte[64]; // Dummy 64-byte signature

    public string GetAddress() => Address;

    public byte[] GetPubKey() => PubKeyBytes;
}
