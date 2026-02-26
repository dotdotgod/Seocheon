# 클래스/인터페이스 정의

> **담당**: SDK 개발자
> **관련 문서**: [아키텍처](01_architecture.md) · [메서드](03_methods.md) · [상수](04_constants.md) · [전체 목차](README.md)

---

## 2.1 최상위 클래스

### `SeocheonSDK`

SDK 엔트리포인트. 설정을 받아 체인에 연결하고, 5개 모듈에 대한 접근을 제공한다.

```
class SeocheonSDK:
  // 생성자
  constructor(config: SDKConfig)

  // 연결
  connect() → void                    // ChainClient 초기화, 연결 확인
  disconnect() → void                 // 연결 종료, 리소스 정리

  // 모듈 접근 (읽기 전용 프로퍼티)
  activity: ActivityModule
  epoch: EpochModule
  node: NodeModule
  rewards: RewardsModule
  cosmos: CosmosModule

  // 상태 조회
  isConnected() → bool
  getConfig() → SDKConfig
```

### `SDKConfig`

```
struct SDKConfig:
  chain: ChainConfig
  signing: SigningConfig
  tx: TxConfig

struct ChainConfig:
  chain_id: string                    // 체인 ID (예: "seocheon-1")
  rpc_endpoint: string                // CometBFT RPC URL (예: "http://localhost:26657")
  grpc_endpoint: string               // Cosmos gRPC URL (예: "http://localhost:9090")
  gas_price: string                   // 가스 가격 (예: "0.025usum")
  gas_adjustment: float64             // 가스 보정 계수 (예: 1.3)

struct SigningConfig:
  mode: "vault" | "keystore" | "direct"
  // vault 모드
  vault_endpoint?: string             // vault-server URL
  key_name?: string                   // vault 키 이름
  // keystore 모드
  keystore_path?: string              // 키스토어 파일 경로
  passphrase_env?: string             // 패스프레이즈 환경변수명
  // direct 모드 (테스트 전용)
  mnemonic?: string                   // 니모닉 문자열

struct TxConfig:
  broadcast_mode: "sync" | "async"    // 기본 "sync"
  confirm_timeout_ms: uint64          // TX 확정 대기 타임아웃 (기본 30000)
  confirm_poll_interval_ms: uint64    // 폴링 간격 (기본 1000)
```

## 2.2 모듈 클래스

```
class ActivityModule:
  submit(activity_hash: string, content_uri: string) → SubmitActivityResponse
  getActivities(node_id?: string, epoch_number?: int64) → GetActivitiesResponse
  getQuota() → GetQuotaResponse

class EpochModule:
  getInfo() → EpochInfoResponse
  getQualification(node_id?: string, epoch_number?: int64) → QualificationResponse

class NodeModule:
  getInfo(node_id?: string) → NodeInfoResponse
  search(tag?: string, status?: string, limit?: uint32, order_by?: string) → NodeSearchResponse

class RewardsModule:
  getPending(node_id?: string) → PendingRewardsResponse
  withdraw() → WithdrawRewardsResponse

class CosmosModule:
  getBalance(address?: string, denom?: string) → BalanceResponse
  sendTokens(to_address: string, amount: string, denom?: string) → SendTokensResponse
  getBlockInfo() → BlockInfoResponse
  getTxResult(tx_hash: string) → TxResultResponse
```

## 2.3 인프라 인터페이스

### `SigningService`

TX 서명을 담당하는 인터페이스. 3개 구현체 제공.

```
interface SigningService:
  sign(tx_bytes: bytes) → bytes       // 서명된 TX 바이트 반환
  getAddress() → string               // 서명 주소 반환
  getPubKey() → bytes                  // 공개키 반환

// 구현체 1: Vault 모드 (권장, 프로덕션)
class VaultSigningService implements SigningService:
  constructor(vault_endpoint: string, key_name: string)

// 구현체 2: Keystore 모드 (중간 보안)
class KeystoreSigningService implements SigningService:
  constructor(keystore_path: string, passphrase: string)

// 구현체 3: Direct 모드 (테스트 전용)
class DirectSigningService implements SigningService:
  constructor(mnemonic: string)
```

### `ChainClient`

Cosmos 체인과의 통신을 추상화하는 인터페이스.

```
interface ChainClient:
  // 연결
  connect(rpc_endpoint: string, grpc_endpoint: string) → void
  disconnect() → void

  // 쿼리
  queryAbci(path: string, data: bytes) → AbciQueryResponse
  queryGrpc(service: string, method: string, request: bytes) → bytes

  // TX
  broadcastTx(tx_bytes: bytes, mode: string) → BroadcastResponse
  getAccountInfo(address: string) → AccountInfo

  // 블록
  getLatestBlock() → BlockResponse
  getBlockByHeight(height: int64) → BlockResponse
  getTx(tx_hash: string) → TxResponse
```

## 2.4 데이터 타입

### Response 구조체

```
struct SubmitActivityResponse:
  tx_hash: string
  block_height: int64
  window_number: int64                // SDK 파생 필드
  epoch_number: int64                 // SDK 파생 필드
  quota_remaining: uint64             // SDK 파생 필드

struct GetActivitiesResponse:
  activities: ActivityItem[]
  total_count: uint64

struct ActivityItem:
  activity_hash: string
  content_uri: string
  block_height: int64
  window_number: int64                // SDK 파생 필드: block_height → window 계산
  tx_hash: string                     // SDK 파생 필드: TX 인덱서에서 조회

struct GetQuotaResponse:
  epoch_number: int64
  quota_total: uint64
  quota_used: uint64
  quota_remaining: uint64
  is_feegrant: bool
  feegrant_expiry: int64?             // SDK 파생 필드: feegrant 쿼리로 병합

struct EpochInfoResponse:
  block_height: int64                 // SDK 파생 필드
  epoch_number: int64
  epoch_start_block: int64
  epoch_end_block: int64              // SDK 파생 필드
  epoch_progress: string              // SDK 파생 필드 (예: "14400/17280")
  window_number: int64
  window_start_block: int64           // SDK 파생 필드
  window_end_block: int64             // SDK 파생 필드
  window_progress: string             // SDK 파생 필드 (예: "720/1440")
  blocks_until_next_window: int64     // SDK 파생 필드
  blocks_until_next_epoch: int64

struct QualificationResponse:
  epoch_number: int64
  total_windows: int64                // 에포크 총 윈도우 수 (12)
  elapsed_windows: int64              // SDK 파생 필드
  active_windows: uint64
  required_windows: int64             // min_active_windows (8)
  is_qualified: bool
  remaining_needed: int64             // SDK 파생 필드
  can_still_qualify: bool             // SDK 파생 필드
  window_detail: WindowActivity[]     // SDK 파생 필드

struct WindowActivity:
  window_number: int64
  submission_count: uint64
  has_activity: bool

struct NodeInfoResponse:
  node_id: string
  operator: string
  agent_address: string
  status: string                      // "REGISTERED" | "ACTIVE" | "INACTIVE" | "JAILED"
  description: string
  website: string
  tags: string[]
  commission_rate: string
  agent_share: string
  total_delegation: string            // SDK 파생 필드: x/staking 쿼리로 병합
  self_delegation: string             // SDK 파생 필드: x/staking 쿼리로 병합
  validator_address: string
  registered_at: int64

struct NodeSearchResponse:
  nodes: NodeSummary[]
  total_count: uint64

struct NodeSummary:
  node_id: string
  status: string
  tags: string[]
  total_delegation: string            // SDK 파생 필드: x/staking 쿼리로 병합
  description: string

struct PendingRewardsResponse:
  delegation_reward: string           // SDK 파생 필드: x/distribution 쿼리
  activity_reward: string             // SDK 파생 필드: x/activity 보상 풀
  total_reward: string
  commission_total: string
  operator_share: string
  agent_share: string

struct WithdrawRewardsResponse:
  tx_hash: string
  withdrawn_total: string
  to_operator: string
  to_agent: string

struct BalanceResponse:
  address: string
  balance: string                     // usum 단위
  balance_kkot: string                // SDK 파생 필드: usum → KKOT 변환

struct SendTokensResponse:
  tx_hash: string
  block_height: int64

struct BlockInfoResponse:
  block_height: int64
  block_time: string                  // ISO 8601
  chain_id: string
  num_txs: uint64

struct TxResultResponse:
  tx_hash: string
  height: int64
  code: uint32                        // 0 = 성공
  gas_used: uint64
  gas_wanted: uint64
  raw_log: string
  events: TxEvent[]

struct TxEvent:
  type: string
  attributes: EventAttribute[]

struct EventAttribute:
  key: string
  value: string
```

### Proto ↔ SDK 타입 매핑 테이블

| Proto 메시지 | SDK 타입 | 비고 |
|-------------|---------|------|
| `seocheon.node.v1.Node` | `NodeInfoResponse` | `total_delegation`, `self_delegation` SDK 병합 |
| `seocheon.node.v1.NodeStatus` (enum) | `string` ("REGISTERED" 등) | SDK에서 문자열로 변환 |
| `seocheon.activity.v1.ActivityRecord` | `ActivityItem` | `window_number`, `tx_hash` SDK 계산/조회 |
| `seocheon.activity.v1.EpochActivitySummary` | `QualificationResponse` 일부 | `elapsed_windows` 등 SDK 계산 |
| `seocheon.activity.v1.QueryEpochInfoResponse` | `EpochInfoResponse` | 4필드만 proto, 나머지 SDK 계산 |
| `seocheon.activity.v1.QueryNodeEpochActivityResponse` | `GetQuotaResponse` | `feegrant_expiry` SDK 별도 조회 |
| `seocheon.node.v1.MsgRegisterNode` | (SDK 내부 전용) | RegisterNode TX용 |
| `seocheon.activity.v1.MsgSubmitActivity` | `SubmitActivityResponse` | 입력 3필드, 응답 5필드(SDK 파생) |
| `seocheon.node.v1.MsgWithdrawNodeCommission` | `WithdrawRewardsResponse` | proto 응답 2필드 → SDK 4필드 |
| `cosmos.bank.v1beta1.MsgSend` | `SendTokensResponse` | Cosmos 표준 |
| `cosmos.base.query.v1beta1.PageRequest` | `PaginationRequest` | SDK 내부 Pagination |
| `cosmos.base.query.v1beta1.PageResponse` | `PaginationResponse` | SDK 내부 Pagination |
