using Seocheon.Sdk.Internal.Tx;
using Xunit;

namespace Seocheon.Sdk.Tests.Internal.Tx;

public class ProtobufTests
{
    [Theory]
    [InlineData(0UL, new byte[] { 0x00 })]
    [InlineData(1UL, new byte[] { 0x01 })]
    [InlineData(127UL, new byte[] { 0x7F })]
    [InlineData(128UL, new byte[] { 0x80, 0x01 })]
    [InlineData(300UL, new byte[] { 0xAC, 0x02 })]
    public void EncodeVarint(ulong value, byte[] expected)
    {
        Assert.Equal(expected, Protobuf.EncodeVarint(value));
    }

    [Fact]
    public void EncodeTag_Field1_Varint()
    {
        var tag = Protobuf.EncodeTag(1, Protobuf.WireVarint);
        Assert.Equal(new byte[] { 0x08 }, tag); // (1 << 3) | 0 = 8
    }

    [Fact]
    public void EncodeTag_Field1_LengthDelimited()
    {
        var tag = Protobuf.EncodeTag(1, Protobuf.WireLengthDelimited);
        Assert.Equal(new byte[] { 0x0A }, tag); // (1 << 3) | 2 = 10
    }

    [Fact]
    public void EncodeString_EmptyReturnsEmpty()
    {
        Assert.Empty(Protobuf.EncodeString(1, ""));
    }

    [Fact]
    public void EncodeString_ValidString()
    {
        var encoded = Protobuf.EncodeString(1, "hi");
        // tag(0x0A) + length(0x02) + "hi"
        Assert.Equal(new byte[] { 0x0A, 0x02, 0x68, 0x69 }, encoded);
    }

    [Fact]
    public void EncodeUint64_ZeroReturnsEmpty()
    {
        Assert.Empty(Protobuf.EncodeUint64(1, 0));
    }
}
