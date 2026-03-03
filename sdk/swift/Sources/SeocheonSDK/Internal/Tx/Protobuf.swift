import Foundation

/// Minimal protobuf wire format encoder.
/// Only supports types needed for Cosmos TX encoding:
/// - varint (wire type 0)
/// - length-delimited (wire type 2)
internal enum Protobuf {
    /// Encodes a UInt64 as a protobuf varint.
    static func encodeVarint(_ v: UInt64) -> Data {
        var value = v
        var buf = Data()
        while value >= 0x80 {
            buf.append(UInt8(value & 0x7F) | 0x80)
            value >>= 7
        }
        buf.append(UInt8(value))
        return buf
    }

    /// Encodes a varint field (wire type 0).
    static func encodeFieldVarint(_ fieldNumber: Int, value: UInt64) -> Data {
        guard value != 0 else { return Data() }
        let tag = encodeVarint(UInt64(fieldNumber << 3 | 0))
        return tag + encodeVarint(value)
    }

    /// Encodes a length-delimited field (wire type 2).
    static func encodeFieldBytes(_ fieldNumber: Int, data: Data) -> Data {
        guard !data.isEmpty else { return Data() }
        let tag = encodeVarint(UInt64(fieldNumber << 3 | 2))
        let length = encodeVarint(UInt64(data.count))
        return tag + length + data
    }

    /// Encodes a string field (wire type 2).
    static func encodeFieldString(_ fieldNumber: Int, value: String) -> Data {
        guard !value.isEmpty else { return Data() }
        return encodeFieldBytes(fieldNumber, data: Data(value.utf8))
    }

    /// Concatenates multiple Data slices, skipping empties.
    static func concatBytes(_ parts: Data...) -> Data {
        var result = Data()
        for p in parts {
            result.append(p)
        }
        return result
    }

    /// Concatenates an array of Data slices.
    static func concatByteArray(_ parts: [Data]) -> Data {
        var result = Data()
        for p in parts {
            result.append(p)
        }
        return result
    }
}
