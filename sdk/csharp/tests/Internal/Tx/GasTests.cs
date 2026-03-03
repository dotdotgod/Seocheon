using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Internal.Tx;
using Xunit;

namespace Seocheon.Sdk.Tests.Internal.Tx;

public class GasTests
{
    [Fact]
    public void GetDefaultGas_SubmitActivity()
    {
        Assert.Equal(ChainConstants.GasSubmitActivity, Gas.GetDefaultGas(Messages.TypeMsgSubmitActivity));
    }

    [Fact]
    public void GetDefaultGas_Withdraw()
    {
        Assert.Equal(ChainConstants.GasWithdraw, Gas.GetDefaultGas(Messages.TypeMsgWithdrawRewards));
    }

    [Fact]
    public void GetDefaultGas_Send()
    {
        Assert.Equal(ChainConstants.GasSend, Gas.GetDefaultGas(Messages.TypeMsgSend));
    }

    [Fact]
    public void GetDefaultGas_Unknown_ReturnsFallback()
    {
        Assert.Equal(ChainConstants.GasFallback, Gas.GetDefaultGas("/unknown.MsgType"));
    }

    [Fact]
    public void CalculateFee()
    {
        // 200000 * 250 = 50_000_000
        Assert.Equal(50_000_000UL, Gas.CalculateFee(200_000, 250));
    }

    [Fact]
    public void ResolveGasLimit_WithOverride()
    {
        Assert.Equal(500_000UL, Gas.ResolveGasLimit(Messages.TypeMsgSubmitActivity, 500_000));
    }
}
