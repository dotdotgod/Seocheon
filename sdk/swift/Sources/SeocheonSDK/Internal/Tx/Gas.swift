import Foundation

/// Default gas limits and fee calculations.
internal enum Gas {
    static let defaultGasSubmitActivity: UInt64 = 200_000
    static let defaultGasWithdrawNodeCommission: UInt64 = 300_000
    static let defaultGasSend: UInt64 = 100_000
    static let defaultGasFallback: UInt64 = 200_000

    static let defaultFeeDenom = "uppyeo"
    static let defaultGasPrice: UInt64 = 250

    /// Returns the default gas limit for a given message type URL.
    static func defaultGasForMessage(_ typeURL: String) -> UInt64 {
        switch typeURL {
        case "/seocheon.activity.v1.MsgSubmitActivity":
            return defaultGasSubmitActivity
        case "/seocheon.node.v1.MsgWithdrawNodeCommission":
            return defaultGasWithdrawNodeCommission
        case "/cosmos.bank.v1beta1.MsgSend":
            return defaultGasSend
        default:
            return defaultGasFallback
        }
    }

    /// Computes the fee amount from gas limit and gas price.
    static func calculateFee(gasLimit: UInt64, gasPrice: UInt64) -> UInt64 {
        return gasLimit * gasPrice
    }
}
