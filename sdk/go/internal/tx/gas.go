package tx

// Default gas limits per message type.
const (
	DefaultGasSubmitActivity          uint64 = 200000
	DefaultGasWithdrawNodeCommission  uint64 = 300000
	DefaultGasSend                    uint64 = 100000
	DefaultGasFallback                uint64 = 200000
)

// Default fee parameters.
const (
	DefaultFeeDenom    = "uppyeo"
	DefaultGasPrice    uint64 = 250 // 250 uppyeo per gas unit
)

// DefaultGasForMessage returns the default gas limit for a given message type URL.
func DefaultGasForMessage(typeURL string) uint64 {
	switch typeURL {
	case "/seocheon.activity.v1.MsgSubmitActivity":
		return DefaultGasSubmitActivity
	case "/seocheon.node.v1.MsgWithdrawNodeCommission":
		return DefaultGasWithdrawNodeCommission
	case "/cosmos.bank.v1beta1.MsgSend":
		return DefaultGasSend
	default:
		return DefaultGasFallback
	}
}

// CalculateFee computes the fee amount from gas limit and gas price.
func CalculateFee(gasLimit, gasPrice uint64) uint64 {
	return gasLimit * gasPrice
}
