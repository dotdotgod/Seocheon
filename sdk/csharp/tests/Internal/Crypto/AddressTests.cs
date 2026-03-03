using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Internal.Crypto;
using Xunit;

namespace Seocheon.Sdk.Tests.Internal.Crypto;

public class AddressTests
{
    [Fact]
    public void FromPubKey_ProducesValidBech32()
    {
        var key = PrivateKey.Generate();
        var pubKey = key.GetPubKey();
        var address = Address.FromPubKey(pubKey);

        Assert.StartsWith(ChainConstants.AddressPrefix + "1", address);
        Assert.True(Address.Validate(address));
    }

    [Fact]
    public void FromPubKey_Deterministic()
    {
        var key = PrivateKey.Generate();
        var pubKey = key.GetPubKey();
        var addr1 = Address.FromPubKey(pubKey);
        var addr2 = Address.FromPubKey(pubKey);
        Assert.Equal(addr1, addr2);
    }

    [Fact]
    public void Validate_InvalidAddress()
    {
        Assert.False(Address.Validate(""));
        Assert.False(Address.Validate("invalid"));
        Assert.False(Address.Validate("cosmos1xxxxxx")); // wrong prefix
    }
}
