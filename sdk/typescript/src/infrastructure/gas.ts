// Default gas limits and fee parameters for Seocheon TX messages.

export const DEFAULT_GAS_SUBMIT_ACTIVITY = 200000;
export const DEFAULT_GAS_WITHDRAW_NODE_COMMISSION = 300000;
export const DEFAULT_GAS_SEND = 100000;
export const DEFAULT_GAS_FALLBACK = 200000;

export const DEFAULT_FEE_DENOM = "uppyeo";
export const DEFAULT_GAS_PRICE = 250; // 250 uppyeo per gas unit

/** Returns the default gas limit for a given message type URL. */
export function defaultGasForMessage(typeUrl: string): number {
  switch (typeUrl) {
    case "/seocheon.activity.v1.MsgSubmitActivity":
      return DEFAULT_GAS_SUBMIT_ACTIVITY;
    case "/seocheon.node.v1.MsgWithdrawNodeCommission":
      return DEFAULT_GAS_WITHDRAW_NODE_COMMISSION;
    case "/cosmos.bank.v1beta1.MsgSend":
      return DEFAULT_GAS_SEND;
    default:
      return DEFAULT_GAS_FALLBACK;
  }
}

/** Computes the fee amount from gas limit and gas price. */
export function calculateFee(gasLimit: number, gasPrice: number): number {
  return gasLimit * gasPrice;
}
