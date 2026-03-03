using Seocheon.Sdk.Internal.Tx;
using Xunit;

namespace Seocheon.Sdk.Tests.Internal.Tx;

public class EnvelopeTests
{
    [Fact]
    public void EncodeAny_ContainsTypeUrl()
    {
        var any = Envelope.EncodeAny("/test.Type", [0x01, 0x02]);
        Assert.NotEmpty(any);
        Assert.Contains((byte)0x0A, any); // Field 1 (type_url)
    }

    [Fact]
    public void EncodeTxBody_WithMessage()
    {
        var msgAny = Envelope.EncodeAny("/test.Msg", [0x01]);
        var body = Envelope.EncodeTxBody(msgAny);
        Assert.NotEmpty(body);
    }

    [Fact]
    public void EncodeTxBody_WithMemo()
    {
        var msgAny = Envelope.EncodeAny("/test.Msg", [0x01]);
        var body = Envelope.EncodeTxBody(msgAny, memo: "test memo");
        Assert.NotEmpty(body);
        // Should be larger than without memo
        var bodyNoMemo = Envelope.EncodeTxBody(msgAny);
        Assert.True(body.Length > bodyNoMemo.Length);
    }

    [Fact]
    public void EncodeAuthInfo_NotEmpty()
    {
        var pubKey = new byte[33];
        pubKey[0] = 0x02;
        var authInfo = Envelope.EncodeAuthInfo(pubKey, 0, 200000, 50000000, "uppyeo");
        Assert.NotEmpty(authInfo);
    }

    [Fact]
    public void EncodeSignDoc_NotEmpty()
    {
        var signDoc = Envelope.EncodeSignDoc(
            [0x01], // body
            [0x02], // auth_info
            "seocheon-1",
            42
        );
        Assert.NotEmpty(signDoc);
    }

    [Fact]
    public void EncodeTxRaw_NotEmpty()
    {
        var txRaw = Envelope.EncodeTxRaw(
            [0x01],       // body
            [0x02],       // auth_info
            new byte[64]  // signature
        );
        Assert.NotEmpty(txRaw);
    }
}
