using Seocheon.Sdk.Errors;
using Xunit;

namespace Seocheon.Sdk.Tests;

public class ConfigTests
{
    private static SdkConfig ValidConfig() => new()
    {
        Chain = new ChainConfig
        {
            ChainId = "seocheon-1",
            RpcEndpoint = "http://localhost:26657",
            GrpcEndpoint = "http://localhost:1317"
        },
        Signing = new SigningConfig
        {
            Mode = SigningMode.Direct,
            Mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        }
    };

    [Fact]
    public void ValidConfig_PassesValidation()
    {
        var config = ValidConfig();
        config.Validate(); // Should not throw
    }

    [Fact]
    public void MissingChainId_Throws()
    {
        var config = ValidConfig() with
        {
            Chain = ValidConfig().Chain with { ChainId = "" }
        };
        var ex = Assert.Throws<SdkException>(config.Validate);
        Assert.Equal(SdkErrors.CodeInvalidConfig, ex.Code);
    }

    [Fact]
    public void MissingRpcEndpoint_Throws()
    {
        var config = ValidConfig() with
        {
            Chain = ValidConfig().Chain with { RpcEndpoint = "" }
        };
        Assert.Throws<SdkException>(config.Validate);
    }

    [Fact]
    public void MissingGrpcEndpoint_Throws()
    {
        var config = ValidConfig() with
        {
            Chain = ValidConfig().Chain with { GrpcEndpoint = "" }
        };
        Assert.Throws<SdkException>(config.Validate);
    }

    [Fact]
    public void VaultMode_RequiresEndpointAndKeyName()
    {
        var config = ValidConfig() with
        {
            Signing = new SigningConfig { Mode = SigningMode.Vault }
        };
        Assert.Throws<SdkException>(config.Validate);
    }

    [Fact]
    public void KeystoreMode_RequiresPathAndPassphraseEnv()
    {
        var config = ValidConfig() with
        {
            Signing = new SigningConfig { Mode = SigningMode.Keystore }
        };
        Assert.Throws<SdkException>(config.Validate);
    }

    [Fact]
    public void DirectMode_RequiresMnemonic()
    {
        var config = ValidConfig() with
        {
            Signing = new SigningConfig { Mode = SigningMode.Direct }
        };
        Assert.Throws<SdkException>(config.Validate);
    }

    [Fact]
    public void InvalidBroadcastMode_Throws()
    {
        var config = ValidConfig() with
        {
            Tx = new TxConfig { BroadcastMode = "invalid" }
        };
        Assert.Throws<SdkException>(config.Validate);
    }
}
