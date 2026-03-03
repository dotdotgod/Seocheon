using Seocheon.Sdk.Internal.Crypto;
using Xunit;

namespace Seocheon.Sdk.Tests.Internal.Crypto;

public class KeysTests
{
    [Fact]
    public void Generate_ProducesValidKey()
    {
        var key = PrivateKey.Generate();
        var bytes = key.ToBytes();
        Assert.Equal(32, bytes.Length);
    }

    [Fact]
    public void FromBytes_RoundTrip()
    {
        var original = PrivateKey.Generate();
        var bytes = original.ToBytes();
        var restored = PrivateKey.FromBytes(bytes);
        Assert.Equal(original.ToBytes(), restored.ToBytes());
    }

    [Fact]
    public void GetPubKey_Is33Bytes()
    {
        var key = PrivateKey.Generate();
        var pubKey = key.GetPubKey();
        Assert.Equal(33, pubKey.Length);
        Assert.True(pubKey[0] == 0x02 || pubKey[0] == 0x03); // Compressed prefix
    }

    [Fact]
    public void Sign_Returns64Bytes()
    {
        var key = PrivateKey.Generate();
        var data = "test data"u8.ToArray();
        var sig = key.Sign(data);
        Assert.Equal(64, sig.Length);
    }

    [Fact]
    public void Sign_Deterministic()
    {
        var key = PrivateKey.Generate();
        var data = "test data"u8.ToArray();
        var sig1 = key.Sign(data);
        var sig2 = key.Sign(data);
        Assert.Equal(sig1, sig2); // RFC 6979 deterministic
    }
}
