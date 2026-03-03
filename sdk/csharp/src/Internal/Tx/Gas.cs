using Seocheon.Sdk.Constants;

namespace Seocheon.Sdk.Internal.Tx;

/// <summary>
/// Gas estimation based on message type.
/// </summary>
public static class Gas
{
    private static readonly Dictionary<string, ulong> DefaultGasLimits = new()
    {
        [Messages.TypeMsgSubmitActivity] = ChainConstants.GasSubmitActivity,
        [Messages.TypeMsgWithdrawRewards] = ChainConstants.GasWithdraw,
        [Messages.TypeMsgSend] = ChainConstants.GasSend,
    };

    /// <summary>
    /// Returns the default gas limit for a given message type.
    /// </summary>
    public static ulong GetDefaultGas(string messageTypeUrl)
    {
        return DefaultGasLimits.GetValueOrDefault(messageTypeUrl, ChainConstants.GasFallback);
    }

    /// <summary>
    /// Calculates fee from gas limit and gas price.
    /// </summary>
    public static ulong CalculateFee(ulong gasLimit, ulong gasPrice = ChainConstants.DefaultGasPrice)
    {
        return gasLimit * gasPrice;
    }

    /// <summary>
    /// Resolves gas limit: uses override if provided, otherwise uses default for message type.
    /// </summary>
    public static ulong ResolveGasLimit(string messageTypeUrl, ulong overrideGas = 0)
    {
        return overrideGas > 0 ? overrideGas : GetDefaultGas(messageTypeUrl);
    }
}
