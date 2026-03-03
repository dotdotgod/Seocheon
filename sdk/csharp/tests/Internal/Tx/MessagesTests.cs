using Seocheon.Sdk.Internal.Tx;
using Xunit;

namespace Seocheon.Sdk.Tests.Internal.Tx;

public class MessagesTests
{
    [Fact]
    public void EncodeMsgSubmitActivity_NotEmpty()
    {
        var msg = Messages.EncodeMsgSubmitActivity(
            "seocheon1test",
            "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
            "https://example.com/report"
        );
        Assert.NotEmpty(msg);
    }

    [Fact]
    public void EncodeMsgSubmitActivity_ContainsFields()
    {
        var msg = Messages.EncodeMsgSubmitActivity("addr", "hash64chars", "uri");
        // Should contain 3 string fields (tags 0x0A, 0x12, 0x1A)
        Assert.Contains((byte)0x0A, msg); // Field 1
        Assert.Contains((byte)0x12, msg); // Field 2
        Assert.Contains((byte)0x1A, msg); // Field 3
    }

    [Fact]
    public void EncodeMsgWithdrawRewards_NotEmpty()
    {
        var msg = Messages.EncodeMsgWithdrawRewards("seocheon1operator");
        Assert.NotEmpty(msg);
    }

    [Fact]
    public void EncodeMsgSend_NotEmpty()
    {
        var msg = Messages.EncodeMsgSend("seocheon1from", "seocheon1to", "1000", "uppyeo");
        Assert.NotEmpty(msg);
    }

    [Fact]
    public void EncodeMsgSend_ContainsAllFields()
    {
        var msg = Messages.EncodeMsgSend("from", "to", "100", "uppyeo");
        Assert.Contains((byte)0x0A, msg); // Field 1 (from)
        Assert.Contains((byte)0x12, msg); // Field 2 (to)
        Assert.Contains((byte)0x1A, msg); // Field 3 (amount)
    }

    [Fact]
    public void MessageTypeUrls_AreCorrect()
    {
        Assert.Equal("/seocheon.activity.v1.MsgSubmitActivity", Messages.TypeMsgSubmitActivity);
        Assert.Equal("/seocheon.node.v1.MsgWithdrawNodeCommission", Messages.TypeMsgWithdrawRewards);
        Assert.Equal("/cosmos.bank.v1beta1.MsgSend", Messages.TypeMsgSend);
    }
}
