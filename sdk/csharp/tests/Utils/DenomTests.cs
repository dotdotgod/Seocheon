using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Utils;
using Xunit;

namespace Seocheon.Sdk.Tests.Utils;

public class DenomTests
{
    [Theory]
    [InlineData(1, "uppyeo", "uppyeo", 1)]
    [InlineData(100, "uppyeo", "sal", 1)]
    [InlineData(10000, "uppyeo", "pi", 1)]
    [InlineData(1000000, "uppyeo", "sum", 1)]
    [InlineData(100000000, "uppyeo", "hon", 1)]
    [InlineData(10000000000, "uppyeo", "kkot", 1)]
    public void ConvertDenom_UppyeoToAll(long amount, string from, string to, long expected)
    {
        Assert.Equal(expected, ConvertUtils.ConvertDenom(amount, from, to));
    }

    [Theory]
    [InlineData(1, "kkot", "uppyeo", 10000000000)]
    [InlineData(1, "hon", "uppyeo", 100000000)]
    [InlineData(1, "sum", "uppyeo", 1000000)]
    [InlineData(1, "pi", "uppyeo", 10000)]
    [InlineData(1, "sal", "uppyeo", 100)]
    public void ConvertDenom_AllToUppyeo(long amount, string from, string to, long expected)
    {
        Assert.Equal(expected, ConvertUtils.ConvertDenom(amount, from, to));
    }

    [Fact]
    public void ConvertDenom_SameUnit()
    {
        Assert.Equal(42, ConvertUtils.ConvertDenom(42, "kkot", "kkot"));
    }

    [Fact]
    public void ConvertDenom_CrossConversion()
    {
        // 1 kkot = 100 hon
        Assert.Equal(100, ConvertUtils.ConvertDenom(1, "kkot", "hon"));
        // 1 hon = 100 sum
        Assert.Equal(100, ConvertUtils.ConvertDenom(1, "hon", "sum"));
    }

    [Fact]
    public void ConvertDenom_InvalidDenom_Throws()
    {
        Assert.Throws<ArgumentException>(() => ConvertUtils.ConvertDenom(1, "invalid", "kkot"));
        Assert.Throws<ArgumentException>(() => ConvertUtils.ConvertDenom(1, "kkot", "invalid"));
    }

    [Theory]
    [InlineData(10000000000, "1.0000000000")]
    [InlineData(0, "0.0000000000")]
    [InlineData(1, "0.0000000001")]
    [InlineData(15000000000, "1.5000000000")]
    [InlineData(100, "0.0000000100")]
    public void FormatKkot(long uppyeo, string expected)
    {
        Assert.Equal(expected, ConvertUtils.FormatKkot(uppyeo));
    }

    [Theory]
    [InlineData("1.0", 10000000000)]
    [InlineData("0.0000000001", 1)]
    [InlineData("1.5", 15000000000)]
    [InlineData("100", 1000000000000)]
    public void ParseKkot(string kkot, long expected)
    {
        Assert.Equal(expected, ConvertUtils.ParseKkot(kkot));
    }

    [Fact]
    public void ParseKkot_EmptyThrows()
    {
        Assert.Throws<FormatException>(() => ConvertUtils.ParseKkot(""));
    }

    [Fact]
    public void FormatHumanReadable()
    {
        var result = ConvertUtils.FormatHumanReadable(10000000000);
        Assert.Equal("1.0000000000 KKOT", result);
    }

    [Fact]
    public void GetDenomNames_Returns6()
    {
        var names = ConvertUtils.GetDenomNames();
        Assert.Equal(6, names.Count);
        Assert.Equal("uppyeo", names[0]);
        Assert.Equal("kkot", names[5]);
    }
}
