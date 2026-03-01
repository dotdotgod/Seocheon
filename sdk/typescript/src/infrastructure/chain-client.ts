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
    _txBytes: Uint8Array,
    _mode: string,
  ): Promise<BroadcastResponse> {
    this.ensureConnected();
    throw new Error("broadcastTx requires CosmJS integration");
  }

  async getAccountInfo(_address: string): Promise<AccountInfo> {
    this.ensureConnected();
    throw new Error("getAccountInfo requires CosmJS integration");
  }

  async getLatestBlock(): Promise<BlockResponse> {
    this.ensureConnected();
    throw new Error("getLatestBlock requires CosmJS integration");
  }

  async getBlockByHeight(_height: number): Promise<BlockResponse> {
    this.ensureConnected();
    throw new Error("getBlockByHeight requires CosmJS integration");
  }

  async getTx(_txHash: string): Promise<TxResponse | null> {
    this.ensureConnected();
    throw new Error("getTx requires CosmJS integration");
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
