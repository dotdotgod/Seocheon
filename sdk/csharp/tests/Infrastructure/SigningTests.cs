using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Internal.Crypto;
using Seocheon.Sdk.Infrastructure.Signing;
using Xunit;

namespace Seocheon.Sdk.Tests.Infrastructure;

public class SigningTests
{
    private const string TestMnemonic =
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";

    [Fact]
    public void DirectService_FromMnemonic()
    {
        var svc = new DirectService(TestMnemonic);
        Assert.NotEmpty(svc.GetAddress());
        Assert.StartsWith(ChainConstants.AddressPrefix, svc.GetAddress());
    }

    [Fact]
    public void DirectService_SignReturns64Bytes()
    {
        var svc = new DirectService(TestMnemonic);
        var sig = svc.Sign("test"u8.ToArray());
        Assert.Equal(64, sig.Length);
    }

    [Fact]
    public void DirectService_PubKeyIs33Bytes()
    {
        var svc = new DirectService(TestMnemonic);
        Assert.Equal(33, svc.GetPubKey().Length);
    }

    [Fact]
    public void DirectService_ConsistentAddress()
    {
        var svc1 = new DirectService(TestMnemonic);
        var svc2 = new DirectService(TestMnemonic);
        Assert.Equal(svc1.GetAddress(), svc2.GetAddress());
    }

    [Fact]
    public void DirectService_DifferentMnemonics_DifferentAddresses()
    {
        var mnemonic2 = "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong";
        var svc1 = new DirectService(TestMnemonic);
        var svc2 = new DirectService(mnemonic2);
        Assert.NotEqual(svc1.GetAddress(), svc2.GetAddress());
    }

    [Fact]
    public void DirectService_AddressIsValidBech32()
    {
        var svc = new DirectService(TestMnemonic);
        var address = svc.GetAddress();
        Assert.True(Address.Validate(address));
    }

    [Fact]
    public void KeystoreService_CreateAndLoad()
    {
        var tempPath = Path.GetTempFileName();
        try
        {
            var key = PrivateKey.Generate();
            KeystoreService.CreateKeystore(tempPath, key.ToBytes(), "testpass");

            var svc = new KeystoreService(tempPath, "testpass");
            Assert.NotEmpty(svc.GetAddress());
            Assert.Equal(33, svc.GetPubKey().Length);
        }
        finally
        {
            File.Delete(tempPath);
        }
    }

    [Fact]
    public void KeystoreService_WrongPassphrase_Throws()
    {
        var tempPath = Path.GetTempFileName();
        try
        {
            var key = PrivateKey.Generate();
            KeystoreService.CreateKeystore(tempPath, key.ToBytes(), "correct");

            Assert.Throws<Errors.SdkException>(() => new KeystoreService(tempPath, "wrong"));
        }
        finally
        {
            File.Delete(tempPath);
        }
    }
}
