import { describe, it, expect, beforeEach } from "vitest";
import { HttpChainClient } from "../../src/infrastructure/chain-client.js";

describe("chain-client", () => {
  let client: HttpChainClient;

  beforeEach(() => {
    client = new HttpChainClient();
  });

  it("17.1: starts_disconnected", () => {
    expect(client.isConnected()).toBe(false);
  });

  it("17.2: connects_and_stores_endpoints", async () => {
    await client.connect("http://localhost:1317", "localhost:9090");
    expect(client.isConnected()).toBe(true);
    expect(client.getRpcEndpoint()).toBe("http://localhost:1317");
    expect(client.getGrpcEndpoint()).toBe("localhost:9090");
  });

  it("17.3: disconnects_and_clears_endpoints", async () => {
    await client.connect("http://localhost:1317", "localhost:9090");
    await client.disconnect();
    expect(client.isConnected()).toBe(false);
    expect(client.getRpcEndpoint()).toBe("");
    expect(client.getGrpcEndpoint()).toBe("");
  });

  it("17.4: strips_trailing_slash", async () => {
    // Verify endpoint is stored as provided (client does not strip)
    await client.connect("http://localhost:1317/", "localhost:9090");
    expect(client.getRpcEndpoint()).toBe("http://localhost:1317/");
  });

  it("17.5: throws_when_not_connected", async () => {
    await expect(
      client.broadcastTx(new Uint8Array([1, 2]), "sync"),
    ).rejects.toThrow("ChainClient is not connected");
    await expect(
      client.getAccountInfo("seocheon1test"),
    ).rejects.toThrow("ChainClient is not connected");
    await expect(client.getLatestBlock()).rejects.toThrow(
      "ChainClient is not connected",
    );
    await expect(
      client.queryRest("/test"),
    ).rejects.toThrow("ChainClient is not connected");
  });
});
