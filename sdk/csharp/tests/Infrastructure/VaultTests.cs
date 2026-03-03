using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Infrastructure.Signing;
using Xunit;

namespace Seocheon.Sdk.Tests.Infrastructure;

public class VaultTests
{
    [Fact]
    public void VaultService_EmptyEndpoint_Throws()
    {
        Assert.Throws<SdkException>(() => new VaultService("", "keyname"));
    }

    [Fact]
    public void VaultService_EmptyKeyName_Throws()
    {
        Assert.Throws<SdkException>(() => new VaultService("http://vault:8200", ""));
    }

    [Fact]
    public void VaultService_UnreachableEndpoint_Throws()
    {
        // Should throw during initialization (fetching key info)
        Assert.Throws<SdkException>(() => new VaultService("http://127.0.0.1:1", "test-key"));
    }
}
