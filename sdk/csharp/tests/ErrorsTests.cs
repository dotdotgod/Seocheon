using Seocheon.Sdk.Errors;
using Xunit;

namespace Seocheon.Sdk.Tests;

public class ErrorsTests
{
    [Fact]
    public void SdkException_FormatsWithCode()
    {
        var ex = new SdkException(9001, "broadcast failed");
        Assert.Equal("[9001] broadcast failed", ex.ToString());
    }

    [Fact]
    public void SdkException_FormatsWithoutCode()
    {
        var ex = new SdkException(0, "generic error");
        Assert.Equal("generic error", ex.ToString());
    }

    [Fact]
    public void Factory_NotConnected()
    {
        var ex = SdkErrors.NotConnected();
        Assert.Equal(SdkErrors.CodeNotConnected, ex.Code);
        Assert.Contains("not connected", ex.Message);
    }

    [Fact]
    public void Factory_TxTimeout()
    {
        var ex = SdkErrors.TxTimeout("ABCD1234");
        Assert.Equal(SdkErrors.CodeTxTimeout, ex.Code);
        Assert.Contains("ABCD1234", ex.Message);
    }

    [Fact]
    public void AbciCodeToError_MapsNodeErrors()
    {
        var ex = SdkErrors.AbciCodeToError(1101);
        Assert.Equal((uint)1101, ex.Code);
        Assert.Contains("Node not found", ex.Message);
    }

    [Fact]
    public void AbciCodeToError_MapsActivityErrors()
    {
        var ex = SdkErrors.AbciCodeToError(1203);
        Assert.Equal((uint)1203, ex.Code);
        Assert.Contains("quota exceeded", ex.Message, StringComparison.OrdinalIgnoreCase);
    }
}
