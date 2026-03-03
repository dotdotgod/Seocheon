import Foundation

/// Protocol for encoding a specific message type into protobuf bytes.
internal protocol MessageEncoder {
    /// The protobuf type URL.
    var typeURL: String { get }
    /// Returns the protobuf-encoded message bytes.
    func encode() -> Data
}

/// Encodes /seocheon.activity.v1.MsgSubmitActivity.
internal struct MsgSubmitActivity: MessageEncoder {
    let submitter: String
    let activityHash: String
    let contentURI: String

    var typeURL: String { "/seocheon.activity.v1.MsgSubmitActivity" }

    func encode() -> Data {
        Protobuf.concatBytes(
            Protobuf.encodeFieldString(1, value: submitter),
            Protobuf.encodeFieldString(2, value: activityHash),
            Protobuf.encodeFieldString(3, value: contentURI)
        )
    }
}

/// Encodes /seocheon.node.v1.MsgWithdrawNodeCommission.
internal struct MsgWithdrawNodeCommission: MessageEncoder {
    let `operator`: String

    var typeURL: String { "/seocheon.node.v1.MsgWithdrawNodeCommission" }

    func encode() -> Data {
        Protobuf.encodeFieldString(1, value: self.operator)
    }
}

/// Represents a cosmos.base.v1beta1.Coin.
internal struct Coin {
    let denom: String
    let amount: String

    func encode() -> Data {
        Protobuf.concatBytes(
            Protobuf.encodeFieldString(1, value: denom),
            Protobuf.encodeFieldString(2, value: amount)
        )
    }
}

/// Encodes /cosmos.bank.v1beta1.MsgSend.
internal struct MsgSend: MessageEncoder {
    let fromAddress: String
    let toAddress: String
    let amount: [Coin]

    var typeURL: String { "/cosmos.bank.v1beta1.MsgSend" }

    func encode() -> Data {
        var parts: [Data] = [
            Protobuf.encodeFieldString(1, value: fromAddress),
            Protobuf.encodeFieldString(2, value: toAddress),
        ]
        for coin in amount {
            parts.append(Protobuf.encodeFieldBytes(3, data: coin.encode()))
        }
        return Protobuf.concatByteArray(parts)
    }
}
