using Seocheon.Sdk.Utils;
using Xunit;

namespace Seocheon.Sdk.Tests.Utils;

public class HashTests
{
    [Fact]
    public void ValidateActivityHash_ValidHash()
    {
        var hash = "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3";
        Assert.True(HashUtils.ValidateActivityHash(hash));
    }

    [Fact]
    public void ValidateActivityHash_WrongLength()
    {
        Assert.False(HashUtils.ValidateActivityHash("abc123"));
    }

    [Fact]
    public void ValidateActivityHash_UpperCase()
    {
        var hash = "A665A45920422F9D417E4867EFDC4FB8A04A1F3FFF1FA07E998E86F7F7A27AE3";
        Assert.False(HashUtils.ValidateActivityHash(hash));
    }

    [Fact]
    public void ValidateActivityHash_NonHex()
    {
        var hash = "zzzza45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3";
        Assert.False(HashUtils.ValidateActivityHash(hash));
    }

    [Fact]
    public void ValidateActivityHash_Empty()
    {
        Assert.False(HashUtils.ValidateActivityHash(""));
        Assert.False(HashUtils.ValidateActivityHash(null!));
    }

    [Fact]
    public void ComputeActivityHash_Bytes()
    {
        var hash = HashUtils.ComputeActivityHash("123"u8.ToArray());
        Assert.Equal(64, hash.Length);
        Assert.True(HashUtils.ValidateActivityHash(hash));
    }

    [Fact]
    public void ComputeActivityHash_String()
    {
        var hash = HashUtils.ComputeActivityHash("hello world");
        Assert.Equal(64, hash.Length);
        Assert.True(HashUtils.ValidateActivityHash(hash));
    }

    [Fact]
    public void ComputeActivityHash_Deterministic()
    {
        var h1 = HashUtils.ComputeActivityHash("test data");
        var h2 = HashUtils.ComputeActivityHash("test data");
        Assert.Equal(h1, h2);
    }
}
