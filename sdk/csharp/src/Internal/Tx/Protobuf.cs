using System.Text;

namespace Seocheon.Sdk.Internal.Tx;

/// <summary>
/// Minimal protobuf encoding (varint, length-delimited, fixed fields).
/// </summary>
public static class Protobuf
{
    // Wire types
    public const int WireVarint = 0;
    public const int WireFixed64 = 1;
    public const int WireLengthDelimited = 2;
    public const int WireFixed32 = 5;

    /// <summary>
    /// Encodes a uint64 as a varint.
    /// </summary>
    public static byte[] EncodeVarint(ulong value)
    {
        var result = new List<byte>(10);
        while (value > 0x7F)
        {
            result.Add((byte)(value | 0x80));
            value >>= 7;
        }
        result.Add((byte)value);
        return result.ToArray();
    }

    /// <summary>
    /// Encodes a field tag (field number + wire type).
    /// </summary>
    public static byte[] EncodeTag(int fieldNumber, int wireType)
    {
        return EncodeVarint((ulong)((fieldNumber << 3) | wireType));
    }

    /// <summary>
    /// Encodes a string field.
    /// </summary>
    public static byte[] EncodeString(int fieldNumber, string value)
    {
        if (string.IsNullOrEmpty(value))
            return [];

        var bytes = Encoding.UTF8.GetBytes(value);
        return EncodeBytes(fieldNumber, bytes);
    }

    /// <summary>
    /// Encodes a bytes field.
    /// </summary>
    public static byte[] EncodeBytes(int fieldNumber, byte[] value)
    {
        if (value.Length == 0)
            return [];

        var tag = EncodeTag(fieldNumber, WireLengthDelimited);
        var length = EncodeVarint((ulong)value.Length);

        var result = new byte[tag.Length + length.Length + value.Length];
        Buffer.BlockCopy(tag, 0, result, 0, tag.Length);
        Buffer.BlockCopy(length, 0, result, tag.Length, length.Length);
        Buffer.BlockCopy(value, 0, result, tag.Length + length.Length, value.Length);
        return result;
    }

    /// <summary>
    /// Encodes a uint64 varint field.
    /// </summary>
    public static byte[] EncodeUint64(int fieldNumber, ulong value)
    {
        if (value == 0)
            return [];

        var tag = EncodeTag(fieldNumber, WireVarint);
        var data = EncodeVarint(value);

        var result = new byte[tag.Length + data.Length];
        Buffer.BlockCopy(tag, 0, result, 0, tag.Length);
        Buffer.BlockCopy(data, 0, result, tag.Length, data.Length);
        return result;
    }

    /// <summary>
    /// Encodes a nested message field.
    /// </summary>
    public static byte[] EncodeMessage(int fieldNumber, byte[] messageBytes)
    {
        return EncodeBytes(fieldNumber, messageBytes);
    }

    /// <summary>
    /// Concatenates multiple encoded fields.
    /// </summary>
    public static byte[] Concat(params byte[][] fields)
    {
        var totalLength = 0;
        foreach (var f in fields)
            totalLength += f.Length;

        var result = new byte[totalLength];
        var offset = 0;
        foreach (var f in fields)
        {
            Buffer.BlockCopy(f, 0, result, offset, f.Length);
            offset += f.Length;
        }
        return result;
    }
}
