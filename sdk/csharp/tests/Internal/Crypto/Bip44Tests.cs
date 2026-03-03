using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Internal.Crypto;
using Xunit;

namespace Seocheon.Sdk.Tests.Internal.Crypto;

public class Bip44Tests
{
    private const string TestMnemonic =
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";

    [Fact]
    public void DeriveKey_ProducesConsistentKey()
    {
        var key1 = Bip44.DeriveKey(TestMnemonic);
        var key2 = Bip44.DeriveKey(TestMnemonic);
        Assert.Equal(key1.ToBytes(), key2.ToBytes());
    }

    [Fact]
    public void DeriveKey_EmptyMnemonic_Throws()
    {
        Assert.Throws<SdkException>(() => Bip44.DeriveKey(""));
    }

    [Fact]
    public void ValidateMnemonic_ValidPhrase()
    {
        Assert.True(Bip44.ValidateMnemonic(TestMnemonic));
    }

    [Fact]
    public void ValidateMnemonic_InvalidPhrase()
    {
        Assert.False(Bip44.ValidateMnemonic("invalid mnemonic phrase"));
        Assert.False(Bip44.ValidateMnemonic(""));
    }
}
