// Message encoders for Seocheon and Cosmos TX messages.
// Each encoder implements the MessageEncoder interface for protobuf wire format.

import {
  encodeFieldString,
  encodeFieldBytes,
  concatBytes,
} from "./protobuf.js";

/** MessageEncoder encodes a specific message type into protobuf bytes. */
export interface MessageEncoder {
  /** Returns the protobuf type URL. */
  typeUrl(): string;
  /** Returns the protobuf-encoded message bytes. */
  encode(): Uint8Array;
}

/**
 * MsgSubmitActivity encodes /seocheon.activity.v1.MsgSubmitActivity.
 * Fields: submitter(1, string), activity_hash(2, string), content_uri(3, string)
 */
export class MsgSubmitActivity implements MessageEncoder {
  constructor(
    public readonly submitter: string,
    public readonly activityHash: string,
    public readonly contentUri: string,
  ) {}

  typeUrl(): string {
    return "/seocheon.activity.v1.MsgSubmitActivity";
  }

  encode(): Uint8Array {
    return concatBytes(
      encodeFieldString(1, this.submitter),
      encodeFieldString(2, this.activityHash),
      encodeFieldString(3, this.contentUri),
    );
  }
}

/**
 * MsgWithdrawNodeCommission encodes /seocheon.node.v1.MsgWithdrawNodeCommission.
 * Fields: operator(1, string)
 */
export class MsgWithdrawNodeCommission implements MessageEncoder {
  constructor(public readonly operator: string) {}

  typeUrl(): string {
    return "/seocheon.node.v1.MsgWithdrawNodeCommission";
  }

  encode(): Uint8Array {
    return concatBytes(encodeFieldString(1, this.operator));
  }
}

/**
 * MsgConfirmDelegation encodes /seocheon.node.v1.MsgConfirmDelegation.
 * Fields: delegator_address(1, string), validator_address(2, string)
 */
export class MsgConfirmDelegation implements MessageEncoder {
  constructor(
    public readonly delegatorAddress: string,
    public readonly validatorAddress: string,
  ) {}

  typeUrl(): string {
    return "/seocheon.node.v1.MsgConfirmDelegation";
  }

  encode(): Uint8Array {
    return concatBytes(
      encodeFieldString(1, this.delegatorAddress),
      encodeFieldString(2, this.validatorAddress),
    );
  }
}

/**
 * Coin represents cosmos.base.v1beta1.Coin.
 * Fields: denom(1, string), amount(2, string)
 */
export class Coin {
  constructor(
    public readonly denom: string,
    public readonly amount: string,
  ) {}

  encode(): Uint8Array {
    return concatBytes(
      encodeFieldString(1, this.denom),
      encodeFieldString(2, this.amount),
    );
  }
}

/**
 * MsgSend encodes /cosmos.bank.v1beta1.MsgSend.
 * Fields: from_address(1, string), to_address(2, string), amount(3, repeated Coin)
 */
export class MsgSend implements MessageEncoder {
  constructor(
    public readonly fromAddress: string,
    public readonly toAddress: string,
    public readonly amount: Coin[],
  ) {}

  typeUrl(): string {
    return "/cosmos.bank.v1beta1.MsgSend";
  }

  encode(): Uint8Array {
    const parts: Uint8Array[] = [
      encodeFieldString(1, this.fromAddress),
      encodeFieldString(2, this.toAddress),
    ];
    for (const coin of this.amount) {
      parts.push(encodeFieldBytes(3, coin.encode()));
    }
    return concatBytes(...parts);
  }
}
