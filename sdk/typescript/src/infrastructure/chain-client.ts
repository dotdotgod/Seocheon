export interface AccountInfo {
  account_number: number;
  sequence: number;
}

export interface BroadcastResponse {
  tx_hash: string;
  code: number;
  raw_log: string;
}

export interface BlockResponse {
  header: {
    height: number;
    time: string;
    chain_id: string;
  };
  data: {
    txs: Uint8Array[];
  };
}

export interface TxResponse {
  tx_hash: string;
  height: number;
  tx_result: {
    code: number;
    gas_used: number;
    gas_wanted: number;
    log: string;
    events: Array<{
      type: string;
      attributes: Array<{ key: string; value: string }>;
    }>;
    msg_responses?: Array<Record<string, unknown>>;
  };
}

export interface ChainClient {
  connect(rpcEndpoint: string, grpcEndpoint: string): Promise<void>;
  disconnect(): Promise<void>;
  isConnected(): boolean;

  queryRest<T>(path: string): Promise<T>;
  broadcastTx(txBytes: Uint8Array, mode: string): Promise<BroadcastResponse>;
  getAccountInfo(address: string): Promise<AccountInfo>;

  getLatestBlock(): Promise<BlockResponse>;
  getBlockByHeight(height: number): Promise<BlockResponse>;
  getTx(txHash: string): Promise<TxResponse | null>;
}

export class HttpChainClient implements ChainClient {
  private rpcEndpoint: string = "";
  private grpcEndpoint: string = "";
  private connected: boolean = false;

  async connect(rpcEndpoint: string, grpcEndpoint: string): Promise<void> {
    this.rpcEndpoint = rpcEndpoint;
    this.grpcEndpoint = grpcEndpoint;
    this.connected = true;
  }

  async disconnect(): Promise<void> {
    this.connected = false;
    this.rpcEndpoint = "";
    this.grpcEndpoint = "";
  }

  isConnected(): boolean {
    return this.connected;
  }

  async queryRest<T>(path: string): Promise<T> {
    this.ensureConnected();
    const url = `${this.rpcEndpoint}${path}`;
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    return response.json() as Promise<T>;
  }

  async broadcastTx(
    txBytes: Uint8Array,
    mode: string,
  ): Promise<BroadcastResponse> {
    this.ensureConnected();
    const txBytesBase64 = uint8ArrayToBase64(txBytes);
    const broadcastMode =
      mode === "async" ? "BROADCAST_MODE_ASYNC" : "BROADCAST_MODE_SYNC";

    const url = `${this.rpcEndpoint}/cosmos/tx/v1beta1/txs`;
    const response = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        tx_bytes: txBytesBase64,
        mode: broadcastMode,
      }),
    });

    if (!response.ok) {
      const body = await response.text();
      throw new Error(`broadcast failed: HTTP ${response.status}: ${body}`);
    }

    const result = (await response.json()) as {
      tx_response: {
        txhash: string;
        code: number;
        raw_log: string;
      };
    };

    return {
      tx_hash: result.tx_response.txhash,
      code: result.tx_response.code,
      raw_log: result.tx_response.raw_log,
    };
  }

  async getAccountInfo(address: string): Promise<AccountInfo> {
    this.ensureConnected();
    const url = `${this.rpcEndpoint}/cosmos/auth/v1beta1/accounts/${address}`;
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(
        `account info query failed: HTTP ${response.status}: ${response.statusText}`,
      );
    }

    const result = (await response.json()) as {
      account: {
        account_number?: string;
        sequence?: string;
        base_account?: {
          account_number: string;
          sequence: string;
        };
      };
    };

    // Handle both direct accounts and nested base_account (e.g. vesting accounts)
    const acct = result.account.base_account ?? result.account;
    return {
      account_number: parseInt(acct.account_number ?? "0", 10),
      sequence: parseInt(acct.sequence ?? "0", 10),
    };
  }

  async getLatestBlock(): Promise<BlockResponse> {
    this.ensureConnected();
    const url = `${this.rpcEndpoint}/cosmos/base/tendermint/v1beta1/blocks/latest`;
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(
        `latest block query failed: HTTP ${response.status}: ${response.statusText}`,
      );
    }

    const result = (await response.json()) as {
      block: {
        header: { height: string; time: string; chain_id: string };
        data: { txs: string[] | null };
      };
    };

    return {
      header: {
        height: parseInt(result.block.header.height, 10),
        time: result.block.header.time,
        chain_id: result.block.header.chain_id,
      },
      data: {
        txs: (result.block.data.txs ?? []).map(base64ToUint8Array),
      },
    };
  }

  async getBlockByHeight(height: number): Promise<BlockResponse> {
    this.ensureConnected();
    const url = `${this.rpcEndpoint}/cosmos/base/tendermint/v1beta1/blocks/${height}`;
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(
        `block query failed: HTTP ${response.status}: ${response.statusText}`,
      );
    }

    const result = (await response.json()) as {
      block: {
        header: { height: string; time: string; chain_id: string };
        data: { txs: string[] | null };
      };
    };

    return {
      header: {
        height: parseInt(result.block.header.height, 10),
        time: result.block.header.time,
        chain_id: result.block.header.chain_id,
      },
      data: {
        txs: (result.block.data.txs ?? []).map(base64ToUint8Array),
      },
    };
  }

  async getTx(txHash: string): Promise<TxResponse | null> {
    this.ensureConnected();
    const url = `${this.rpcEndpoint}/cosmos/tx/v1beta1/txs/${txHash}`;
    const response = await fetch(url);

    if (!response.ok) {
      if (response.status === 404) return null;
      const body = await response.text();
      if (body.includes("not found") || body.includes("not exist")) {
        return null;
      }
      throw new Error(
        `tx query failed: HTTP ${response.status}: ${response.statusText}`,
      );
    }

    const result = (await response.json()) as {
      tx_response: {
        txhash: string;
        height: string;
        code: number;
        gas_used: string;
        gas_wanted: string;
        raw_log: string;
        events: Array<{
          type: string;
          attributes: Array<{ key: string; value: string }>;
        }>;
      };
    };

    return {
      tx_hash: result.tx_response.txhash,
      height: parseInt(result.tx_response.height, 10),
      tx_result: {
        code: result.tx_response.code,
        gas_used: parseInt(result.tx_response.gas_used, 10),
        gas_wanted: parseInt(result.tx_response.gas_wanted, 10),
        log: result.tx_response.raw_log,
        events: result.tx_response.events ?? [],
      },
    };
  }

  getRpcEndpoint(): string {
    return this.rpcEndpoint;
  }

  getGrpcEndpoint(): string {
    return this.grpcEndpoint;
  }

  private ensureConnected(): void {
    if (!this.connected) {
      throw new Error("ChainClient is not connected");
    }
  }
}

function uint8ArrayToBase64(bytes: Uint8Array): string {
  if (typeof Buffer !== "undefined") {
    return Buffer.from(bytes).toString("base64");
  }
  let binary = "";
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function base64ToUint8Array(b64: string): Uint8Array {
  if (typeof Buffer !== "undefined") {
    return new Uint8Array(Buffer.from(b64, "base64"));
  }
  const binary = atob(b64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}
