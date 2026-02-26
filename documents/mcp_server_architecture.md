# Seocheon Client SDK & MCP 서버 아키텍처

> **담당**: 에이전트 인프라 개발자 (TypeScript)
> **관련 문서**: [에이전트 아키텍처](agent_architecture.md) · [Activity Protocol](blockchain/04_activity_protocol.md) · [노드 모듈](blockchain/03_node_module.md) · [인덱서](blockchain/13_indexer_architecture.md) · [구현 가이드](blockchain/09_implementation.md)

Seocheon 체인과 상호작용하기 위한 Client SDK와 MCP 서버 설계이다. **Client SDK가 기반 계층**이고, MCP 서버는 SDK를 감싸는 얇은 래퍼이다.

---

## 계층 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│  AI 에이전트 (OpenClaw, Claude 등)                            │  ← MCP 프로토콜로 접근
├─────────────────────────────────────────────────────────────┤
│  MCP 서버 (seocheon-server)                                  │  ← SDK 함수를 MCP 도구로 노출
│    얇은 래퍼: MCP 프로토콜 처리 + 도구 스키마 정의              │
├─────────────────────────────────────────────────────────────┤
│  Client SDK (@seocheon/sdk)                                  │  ← 핵심 비즈니스 로직
│    TX 구성·서명·브로드캐스트, 쿼리, 필드 계산, 에러 처리         │
├─────────────────────────────────────────────────────────────┤
│  CosmJS (@cosmjs/stargate)                                   │  ← Cosmos 표준 클라이언트
├─────────────────────────────────────────────────────────────┤
│  Cosmos gRPC / CometBFT RPC                                  │
└─────────────────────────────────────────────────────────────┘
```

### 각 계층의 역할

| 계층 | 역할 | 대상 |
|------|------|------|
| **MCP 서버** | SDK 함수를 MCP 도구 스키마로 노출. 프로토콜 변환만 담당 | AI 에이전트 (LLM 기반) |
| **Client SDK** | TX 구성·서명·브로드캐스트, 쿼리 실행, 파생 필드 계산, 에러 처리. 모든 비즈니스 로직이 여기에 있다 | 개발자, MCP 서버, 커스텀 에이전트 |
| **CosmJS** | Cosmos SDK 체인 RPC/gRPC 통신, 프로토콜 인코딩 | SDK 내부 사용 |

### 왜 SDK가 기반인가

```
MCP 서버만 있을 때의 문제:
  → MCP 미지원 에이전트 프레임워크에서 사용 불가
  → 개발자가 커스텀 에이전트를 만들려면 MCP를 거쳐야 함
  → 비즈니스 로직이 MCP 서버에 종속

SDK를 기반으로 분리하면:
  → MCP 서버: SDK를 import하여 MCP 도구로 감싸기만 함 (비즈니스 로직 0)
  → 개발자: SDK를 직접 import하여 자유롭게 사용
  → 제3자: SDK 위에 자신만의 MCP 서버나 REST API를 구축 가능
```

---

## 설계 원칙

- **단일 호출 완결**: SDK 함수 하나로 원하는 결과를 얻는다. TX 생성 → 서명 → 브로드캐스트를 내부에서 처리한다
- **키 미노출**: 에이전트 지갑 키가 SDK 함수의 반환값에 노출되지 않는다. 서명은 SDK 내부에서 처리한다
- **체인 철학 일관**: 체인이 검증하지 않는 것을 SDK/MCP도 검증하지 않는다
- **오프체인 참고 아키텍처**: 이 문서는 레퍼런스 구현의 스펙이다. 노드 운영자가 자유롭게 변형할 수 있다

---

## 서버 구성

```
Agent (LLM)
  │
  ├── MCP: seocheon-server ─────────────────────────────────────────┐
  │     │   (얇은 래퍼: SDK 함수를 MCP 도구로 노출)                   │
  │     │                                                          │
  │     └── @seocheon/sdk ──────────────────────────────────────┐  │
  │           ├── sdk.activity (submit, getActivities, ...)     │  │
  │           ├── sdk.epoch (getInfo, getQualification, ...)    │  │
  │           ├── sdk.node (getInfo, search, ...)               │  │
  │           ├── sdk.rewards (getPending, withdraw, ...)       │  │
  │           ├── sdk.cosmos (getBalance, sendTokens, ...)      │  │
  │           │                                                 │  │
  │           ├── [내부] Signing Service (vault 또는 keystore)   │  │
  │           └── [내부] CosmJS → CometBFT RPC/gRPC             │  │
  │                                                             │  │
  ├── MCP: vault-server                                         │  │
  │     ├── secure_call (외부 API 프록시)                        │  │
  │     ├── register_secret / list_secrets                      │  │
  │     └── [내부] Signing Service ◄──────── 공유 ──────────────┘  │
  │                                                                │
  └── MCP: other-tools (웹 검색, 파일 시스템 등 범용 도구)            │
                                                                   │
Developer / 커스텀 에이전트                                          │
  │                                                                │
  └── @seocheon/sdk (직접 import) ◄────────────────────────────────┘
```

### 서명 서비스 아키텍처

Signing Service는 **Client SDK의 내부 모듈**이다. seocheon-server(MCP)와 vault-server가 SDK를 통해 공유한다.

```
Signing Service (SDK 내부 모듈)
├── 키 저장소: 암호화된 keystore 파일 또는 HSM
├── sign(tx_bytes) → signed_tx
├── get_address() → agent_address
└── 설정:
    ├── keystore_path: "./keys/agent.key"
    ├── encryption: "aes-256-gcm"
    └── hsm_enabled: false (기본값)
```

**키 미노출 보장**: SDK 함수의 입력/출력 어디에도 키 정보가 포함되지 않는다. TX 도구는 결과만 반환하며, 서명 과정은 SDK 내부에서 완결된다.

---

## Client SDK 함수 상세

아래 함수들은 **Client SDK (`@seocheon/sdk`)** 가 제공한다. MCP 서버는 이 함수들을 MCP 도구 스키마로 감싸 노출할 뿐, 비즈니스 로직을 포함하지 않는다.

```typescript
// MCP 서버에서의 사용 예시 (얇은 래퍼)
server.tool("submit_activity", schema, async (params) => {
  return sdk.activity.submit(params.activity_hash, params.content_uri);
});

// 개발자의 직접 사용 예시
import { SeocheonSDK } from "@seocheon/sdk";
const sdk = new SeocheonSDK(config);
const result = await sdk.activity.submit(hash, uri);
```

### 1. 활동 (sdk.activity)

#### `sdk.activity.submit(activity_hash, content_uri)`

에이전트의 핵심 함수. 활동 해시를 온체인에 타임스탬핑한다.

```
SDK: sdk.activity.submit(activity_hash, content_uri)
MCP: submit_activity
설명: MsgSubmitActivity TX를 생성, 서명, 브로드캐스트한다

입력:
  activity_hash: string (hex)   — Activity Report 핑거프린트 (SHA-256, hex 64자)
  content_uri: string           — 오프체인 Activity Report 위치 (IPFS CID, URL 등)

출력:
  tx_hash: string               — 트랜잭션 해시
  block_height: number          — 포함된 블록 높이 (확정 시)
  window_number: number         — 현재 윈도우 번호
  epoch_number: number          — 현재 에포크 번호
  quota_remaining: number       — 에포크 내 남은 제출 쿼터

내부 처리:
  1. 에이전트 주소로 MsgSubmitActivity 메시지 생성
  2. 계정 시퀀스 + 체인 ID 조회
  3. Signing Service로 TX 서명
  4. BroadcastTxSync → TX 해시 즉시 반환
  5. TX 확정 대기 (폴링 또는 WebSocket)
  6. 결과 반환

에러:
  QUOTA_EXCEEDED    — 에포크 쿼터 초과
  DUPLICATE_HASH    — 이미 제출된 activity_hash
  NODE_NOT_FOUND    — 서명 주소가 등록된 노드의 agent_address가 아님
  BROADCAST_FAILED  — TX 브로드캐스트 실패
```

#### `sdk.activity.getActivities(node_id?, epoch_number?)`

자신 또는 특정 노드의 활동 기록을 조회한다.

```
SDK: sdk.activity.getActivities(node_id?, epoch_number?)
MCP: get_activities
설명: 노드의 활동 제출 이력을 조회한다

입력:
  node_id: string? (선택)       — 미지정 시 자기 노드
  epoch_number: number? (선택)  — 미지정 시 현재 에포크

출력:
  activities: array
    ├── activity_hash: string
    ├── content_uri: string
    ├── block_height: number
    ├── window_number: number   — block_height ÷ window_length로 SDK가 계산
    └── tx_hash: string          — TX 인덱서에서 조회 (ActivityRecord proto에 미포함)
  total_count: number
```

> **proto 매핑 참고**: `ActivityRecord` protobuf에는 `window_number`와 `tx_hash` 필드가 없다. SDK가 블록 높이와 파라미터로 윈도우 번호를 계산하고, TX 해시는 Cosmos SDK TX 인덱서에서 별도 조회한다.

#### `sdk.activity.getQuota()`

에포크 내 남은 활동 제출 쿼터를 확인한다.

```
SDK: sdk.activity.getQuota()
MCP: get_activity_quota
설명: 현재 에포크의 활동 제출 쿼터 상태를 조회한다

입력: (없음, 자기 노드 자동)

출력:
  epoch_number: number
  quota_total: number           — 에포크당 총 쿼터
  quota_used: number            — 사용한 쿼터
  quota_remaining: number       — 남은 쿼터
  is_feegrant: boolean          — feegrant 노드 여부
  feegrant_expiry: number?      — feegrant 만료 블록 (해당 시)
```

> **proto 매핑 참고**: `feegrant_expiry`는 `NodeEpochActivity` proto 응답에 포함되지 않는다. SDK가 Cosmos SDK 표준 feegrant 쿼리(`/cosmos/feegrant/v1beta1/allowance`)로 별도 조회하여 병합한다.

---

### 2. 에포크/윈도우 (sdk.epoch)

#### `sdk.epoch.getInfo()`

현재 에포크와 윈도우 상태를 조회한다. 에이전트가 활동 제출 타이밍을 판단하는 핵심 함수.

```
SDK: sdk.epoch.getInfo()
MCP: get_epoch_info
설명: 현재 에포크/윈도우 진행 상태를 조회한다

입력: (없음)

출력:
  block_height: number          — 현재 블록 높이
  epoch_number: number          — 현재 에포크 번호
  epoch_start_block: number     — 에포크 시작 블록
  epoch_end_block: number       — 에포크 종료 블록
  epoch_progress: string        — 에포크 진행률 (예: "14400/17280")
  window_number: number         — 현재 윈도우 번호 (에포크 내 1~12)
  window_start_block: number    — 윈도우 시작 블록
  window_end_block: number      — 윈도우 종료 블록
  window_progress: string       — 윈도우 진행률 (예: "720/1440")
  blocks_until_next_window: number
  blocks_until_next_epoch: number
```

> **proto 매핑 참고**: `QueryEpochInfoResponse` proto는 4개 필드(`current_epoch`, `epoch_start_block`, `current_window`, `blocks_until_next_epoch`)만 반환한다. 나머지 필드(`block_height`, `epoch_end_block`, `epoch_progress`, `window_start_block`, `window_end_block`, `window_progress`, `blocks_until_next_window`)는 SDK가 에포크/윈도우 파라미터와 현재 블록 높이로 계산한다.

#### `sdk.epoch.getQualification(node_id?, epoch_number?)`

현재 에포크의 활동 자격 충족 상태를 조회한다.

```
SDK: sdk.epoch.getQualification(node_id?, epoch_number?)
MCP: get_qualification_status
설명: 에포크 내 활동 보상 자격 상태를 조회한다

입력:
  node_id: string? (선택)       — 미지정 시 자기 노드
  epoch_number: number? (선택)  — 미지정 시 현재 에포크

출력:
  epoch_number: number
  total_windows: number         — 에포크 총 윈도우 수 (12)
  elapsed_windows: number       — 경과한 윈도우 수
  active_windows: number        — 활동한 윈도우 수 (1건 이상 제출)
  required_windows: number      — 자격에 필요한 최소 윈도우 수 (8)
  is_qualified: boolean         — 현재 자격 충족 여부
  remaining_needed: number      — 자격까지 추가로 필요한 윈도우 수
  can_still_qualify: boolean    — 남은 윈도우로 자격 충족 가능 여부
  window_detail: array          — 윈도우별 활동 현황
    ├── window_number: number
    ├── submission_count: number
    └── has_activity: boolean
```

> **proto 매핑 참고**: `NodeEpochActivity` proto 응답은 `EpochActivitySummary`(3필드: `active_windows`, `eligible`, `total_submissions`) + 쿼터 2필드만 반환한다. `elapsed_windows`, `remaining_needed`, `can_still_qualify`는 SDK가 `EpochInfo`와 조합하여 계산한다. `window_detail[]`은 `WindowActivity` 컬렉션을 직접 쿼리하거나 전용 쿼리 추가 시 제공 가능하다.

---

### 3. 노드 (sdk.node)

#### `sdk.node.getInfo(node_id?)`

노드의 상세 정보를 조회한다.

```
SDK: sdk.node.getInfo(node_id?)
MCP: get_node_info
설명: 노드의 온체인 정보를 조회한다

입력:
  node_id: string? (선택)       — 미지정 시 자기 노드

출력:
  node_id: string
  operator: string              — 운영자 주소
  agent_address: string         — 에이전트 주소
  status: string                — REGISTERED | ACTIVE | INACTIVE | JAILED
  description: string
  website: string
  tags: string[]
  commission_rate: string       — 커미션율
  agent_share: string           — agent_share 비율
  total_delegation: string      — 총 위임량 (KKOT) — SDK가 x/staking 쿼리로 병합
  self_delegation: string       — 자기위임량 — SDK가 x/staking 쿼리로 병합
  validator_address: string     — 밸리데이터 주소
  registered_at: number         — 등록 블록 높이
```

#### `sdk.node.search(tag?, status?, limit?, order_by?)`

조건에 맞는 노드를 검색한다.

```
SDK: sdk.node.search(tag?, status?, limit?, order_by?)
MCP: search_nodes
설명: 태그, 상태 등 조건으로 노드를 검색한다

입력:
  tag: string? (선택)           — 태그 필터
  status: string? (선택)        — 상태 필터 (REGISTERED, ACTIVE 등)
  limit: number? (기본 20)      — 결과 수 제한
  order_by: string? (기본 "delegation") — 정렬 기준 (delegation, registered_at) — SDK 레벨 필터링/정렬

출력:
  nodes: array
    ├── node_id: string
    ├── status: string
    ├── tags: string[]
    ├── total_delegation: string
    └── description: string
  total_count: number
```

> **proto 매핑 참고**: `NodesByTag` proto는 태그 필터만, `AllNodes`는 필터 없이 전체 반환한다. `status` 필터와 `order_by` 정렬은 SDK 레벨에서 수행한다. `total_delegation`은 x/staking 쿼리로 병합한다.

---

### 4. 보상 (sdk.rewards)

#### `sdk.rewards.getPending(node_id?)`

미인출 보상을 조회한다.

```
SDK: sdk.rewards.getPending(node_id?)
MCP: get_pending_rewards
설명: 현재 미인출 보상을 조회한다

입력:
  node_id: string? (선택)       — 미지정 시 자기 노드

출력:
  delegation_reward: string     — 위임 보상 (KKOT)
  activity_reward: string       — 활동 보상 (KKOT)
  total_reward: string          — 총 보상
  commission_total: string      — 미인출 커미션 총액
  operator_share: string        — 커미션 중 operator 몫
  agent_share: string           — 커미션 중 agent 몫
```

> **proto 매핑 참고**: x/node, x/activity에 pending rewards 전용 쿼리가 없다. SDK가 Cosmos SDK 표준 x/distribution 쿼리(`/cosmos/distribution/v1beta1/validators/{addr}/outstanding_rewards`)와 x/activity 보상 풀 상태를 조합하여 산출한다.

#### `sdk.rewards.withdraw()`

보상을 인출한다. (TX)

```
SDK: sdk.rewards.withdraw()
MCP: withdraw_rewards
설명: 미인출 보상을 인출한다 (MsgWithdrawNodeCommission)

입력: (없음, 자기 노드 자동)

출력:
  tx_hash: string
  withdrawn_total: string       — 인출된 총액
  to_operator: string           — operator에게 전송된 금액
  to_agent: string              — agent에게 전송된 금액

참고: operator 서명이 필요한 TX이므로, operator 키가 Signing Service에 등록되어 있어야 한다.
      agent 키만 있는 경우 이 도구는 사용 불가.
      Signing Service의 키 관리: vault-server가 키스토어(암호화 파일 또는 HSM)를 관리하며,
      operator 키와 agent 키를 분리 보관한다. operator 키 등록은 노드 초기 설정 시
      vault-server의 register_secret 또는 직접 키스토어 설정으로 수행한다.
      상세: vault-server 도구 (아래) 및 에이전트 아키텍처 §Vault 참조.
```

---

### 5. 범용 체인 (sdk.cosmos)

Cosmos SDK 표준 기능을 래핑한다. Seocheon 이외의 Cosmos 체인에서도 재사용 가능한 함수.

#### `sdk.cosmos.getBalance(address?, denom?)`

토큰 잔고를 조회한다.

```
SDK: sdk.cosmos.getBalance(address?, denom?)
MCP: get_balance
설명: 계정의 토큰 잔고를 조회한다

입력:
  address: string? (선택)       — 미지정 시 자기 agent 주소
  denom: string? (기본 "usum")  — 토큰 단위

출력:
  address: string
  balance: string               — 잔고 (usum)
  balance_kkot: string          — 잔고 (KKOT, 소수점 변환)
```

#### `sdk.cosmos.sendTokens(to_address, amount, denom?)`

토큰을 전송한다. (TX)

```
SDK: sdk.cosmos.sendTokens(to_address, amount, denom?)
MCP: send_tokens
설명: 토큰을 다른 주소로 전송한다

입력:
  to_address: string            — 수신자 주소
  amount: string                — 금액 (usum)
  denom: string? (기본 "usum")

출력:
  tx_hash: string
  block_height: number
```

#### `sdk.cosmos.getBlockInfo()`

현재 블록 정보를 조회한다.

```
SDK: sdk.cosmos.getBlockInfo()
MCP: get_block_info
설명: 현재 블록 정보를 조회한다

입력: (없음)

출력:
  block_height: number
  block_time: string            — ISO 8601 타임스탬프
  chain_id: string
  num_txs: number               — 블록 내 TX 수
```

#### `sdk.cosmos.getTxResult(tx_hash)`

트랜잭션 실행 결과를 조회한다.

```
SDK: sdk.cosmos.getTxResult(tx_hash)
MCP: get_tx_result
설명: TX 해시로 트랜잭션 결과를 조회한다

입력:
  tx_hash: string

출력:
  tx_hash: string
  height: number
  code: number                  — 0 = 성공, 그 외 = 실패
  gas_used: number
  gas_wanted: number
  raw_log: string
  events: array                 — ABCI 이벤트 목록
```

---

## vault-server 도구 (참조)

vault-server의 상세 설계는 [에이전트 아키텍처](agent_architecture.md) §Vault 참조. 여기서는 seocheon-server와의 관계만 기술한다.

### LLM에 노출되는 도구 (MCP)

| 도구 | 용도 |
|------|------|
| `secure_call` | 외부 API 호출 시 시크릿 프록시 (API 키를 LLM에 노출하지 않음) |
| `register_secret` | 새 시크릿 등록 (서버 측 즉시 암호화) |
| `list_secrets` | 등록된 시크릿 목록 (값 제외, 이름만) |

### LLM에 노출되지 않는 기능 (내부 서비스)

| 기능 | 사용자 | 용도 |
|------|--------|------|
| `sign(tx_bytes)` | seocheon-server | TX 서명 (agent 키) |
| `get_address()` | seocheon-server | agent 주소 조회 |

Signing Service는 MCP 도구가 아닌 **SDK 내부 서비스**이다. SDK가 Signing Service를 직접 호출하여 TX 서명을 처리한다.

---

## TX 서명 통합 플로우

모든 TX 함수는 동일한 내부 플로우를 따른다:

```
[MCP 경로]
Agent: submit_activity(hash, uri)  [MCP 도구 호출]
  → seocheon-server: SDK 함수 위임
    → sdk.activity.submit(hash, uri)

[직접 사용 경로]
Developer: sdk.activity.submit(hash, uri)  [SDK 직접 호출]

[공통 SDK 내부 처리]
@seocheon/sdk:
  ① MsgSubmitActivity 메시지 생성
  ② 체인 조회: account_number, sequence, chain_id
  ③ TX 바이트 조립 (amino/protobuf 인코딩)
  ④ Signing Service 호출: sign(tx_bytes) → signed_tx
  ⑤ CometBFT RPC: BroadcastTxSync(signed_tx)
  ⑥ TX 확정 대기 (폴링, 최대 30초)
  ⑦ 결과 반환
         │
         ▼
{ tx_hash, block_height, ... }  [결과 수신]
```

### Signing Service 설정 모드

| 모드 | 설명 | 보안 수준 |
|------|------|----------|
| **vault** (권장) | vault-server의 Signing Service 사용. 키는 vault의 암호화 저장소에 보관 | 높음 |
| **keystore** | 로컬 암호화 키스토어 파일 사용. 서버 시작 시 패스프레이즈로 복호화 | 중간 |

**설정 예시**:

```json
{
  "signing": {
    "mode": "vault",
    "vault_endpoint": "http://localhost:3001",
    "key_name": "agent_key"
  }
}
```

```json
{
  "signing": {
    "mode": "keystore",
    "keystore_path": "./keys/agent.key.enc",
    "passphrase_env": "AGENT_KEY_PASSPHRASE"
  }
}
```

---

## 에이전트 워크플로우 예시

### 1. 에포크 활동 루틴 (8/12 윈도우 자격 유지)

에이전트가 에포크 활동 보상을 받기 위해 12윈도우 중 8윈도우 이상 활동을 제출해야 한다.

```
에이전트의 에포크 루틴:

[매 윈도우 시작 시]
  1. get_epoch_info() → 현재 윈도우 번호, 남은 블록 확인
  2. get_qualification_status() → 활동 자격 현황 확인

  [자격 미충족 + 달성 가능한 경우]
    3. 작업 수행 (LLM agentic loop, SDK/MCP 도구 활용)
    4. 활동 결과물 생성 → Activity Report JSON 작성
    5. Activity Report를 오프체인 저장소에 업로드 (IPFS 등)
    6. 활동 해시 계산: SHA-256(Activity Report)
    7. submit_activity(activity_hash, content_uri)
    8. 결과 확인: quota_remaining, window_number

  [자격 이미 충족 (8/12 달성)]
    → 추가 활동은 선택 (위임자에게 성실함 표시 효과)

  [자격 달성 불가 (남은 윈도우 부족)]
    → 이번 에포크 활동 보상 포기, 다음 에포크 준비
```

### 2. 보상 확인 및 자원 관리

```
에이전트의 자원 관리 루틴:

[매 에포크 종료 시]
  1. get_pending_rewards() → 미인출 보상 확인
  2. get_balance() → 현재 잔고 확인
  3. get_activity_quota() → feegrant 상태 확인

  [잔고 부족 시]
    → withdraw_rewards() 실행 (operator 키 필요)
    → 또는 operator에게 알림

  [feegrant 만료 임박]
    → 자비 가스비 전환 준비 안내
```

---

## 기술 스택

```
@seocheon/sdk (Client SDK) 기술 스택:

언어: TypeScript (Node.js)
  → CosmJS 생태계와 자연스러운 통합
  → 브라우저/Node.js 양쪽 지원

체인 클라이언트: CosmJS (@cosmjs/stargate)
  → Cosmos SDK 체인 RPC/gRPC 통신
  → TX 생성, 서명, 브로드캐스트

설정 관리: dotenv + JSON config
  → 체인 엔드포인트, 키 설정

테스트: vitest
  → 단위 테스트: 각 SDK 함수의 입출력 검증
  → 통합 테스트: 로컬 테스트넷 연동

seocheon-mcp (MCP 서버) 추가 의존성:

MCP SDK: @modelcontextprotocol/sdk
  → MCP 서버 프레임워크, stdio transport

의존성: @seocheon/sdk
  → SDK를 import하여 MCP 도구로 노출
```

### SDK 설정 파일

```json
{
  "chain": {
    "chain_id": "seocheon-1",
    "rpc_endpoint": "http://localhost:26657",
    "grpc_endpoint": "http://localhost:9090",
    "gas_price": "0.025usum",
    "gas_adjustment": 1.3
  },
  "signing": {
    "mode": "vault",
    "vault_endpoint": "http://localhost:3001",
    "key_name": "agent_key"
  },
  "tx": {
    "broadcast_mode": "sync",
    "confirm_timeout_ms": 30000,
    "confirm_poll_interval_ms": 1000
  }
}
```

### MCP 서버 설정 파일 (SDK 설정 확장)

```json
{
  "server": {
    "name": "seocheon-server",
    "version": "1.0.0",
    "transport": "stdio"
  },
  "sdk": "./sdk-config.json"
}
```

---

## 모듈 구조

### Client SDK (`@seocheon/sdk`)

```
seocheon-sdk/
├── src/
│   ├── index.ts                    ← SDK 엔트리포인트 (SeocheonSDK 클래스)
│   ├── config.ts                   ← 설정 로드
│   ├── signing/
│   │   ├── index.ts                ← Signing Service 인터페이스
│   │   ├── vault.ts                ← Vault 모드 구현
│   │   └── keystore.ts             ← Keystore 모드 구현
│   ├── chain/
│   │   ├── client.ts               ← CosmJS 클라이언트 래퍼
│   │   ├── tx.ts                   ← TX 생성 + 브로드캐스트 유틸
│   │   └── query.ts                ← 체인 쿼리 유틸
│   ├── modules/
│   │   ├── activity.ts             ← sdk.activity: submit, getActivities, getQuota
│   │   ├── epoch.ts                ← sdk.epoch: getInfo, getQualification
│   │   ├── node.ts                 ← sdk.node: getInfo, search
│   │   ├── rewards.ts              ← sdk.rewards: getPending, withdraw
│   │   └── cosmos.ts               ← sdk.cosmos: getBalance, sendTokens, getBlockInfo, getTxResult
│   └── types/
│       ├── sdk.ts                  ← SDK 공개 타입 정의
│       └── chain.ts                ← 체인 데이터 타입
├── tests/
│   ├── unit/                       ← 모듈별 단위 테스트
│   └── integration/                ← 테스트넷 통합 테스트
├── package.json
└── tsconfig.json
```

### MCP 서버 (`seocheon-mcp`)

```
seocheon-mcp/
├── src/
│   ├── server.ts                   ← MCP 서버 엔트리포인트
│   └── tools/                      ← SDK 함수를 MCP 도구로 매핑 (얇은 래퍼)
│       ├── activity.ts             ← submit_activity, get_activities, get_activity_quota
│       ├── epoch.ts                ← get_epoch_info, get_qualification_status
│       ├── node.ts                 ← get_node_info, search_nodes
│       ├── rewards.ts              ← get_pending_rewards, withdraw_rewards
│       └── cosmos.ts               ← get_balance, send_tokens, get_block_info, get_tx_result
├── package.json                    ← depends on @seocheon/sdk
└── tsconfig.json
```

---

## 함수/도구 요약

| 카테고리 | SDK 함수 | MCP 도구 | TX | 설명 |
|---------|----------|----------|-----|------|
| 활동 | `sdk.activity.submit()` | `submit_activity` | O | 활동 해시 온체인 제출 |
| 활동 | `sdk.activity.getActivities()` | `get_activities` | | 활동 이력 조회 |
| 활동 | `sdk.activity.getQuota()` | `get_activity_quota` | | 쿼터 잔량 조회 |
| 에포크 | `sdk.epoch.getInfo()` | `get_epoch_info` | | 에포크/윈도우 상태 |
| 에포크 | `sdk.epoch.getQualification()` | `get_qualification_status` | | 활동 자격 현황 |
| 노드 | `sdk.node.getInfo()` | `get_node_info` | | 노드 상세 정보 |
| 노드 | `sdk.node.search()` | `search_nodes` | | 노드 검색 |
| 보상 | `sdk.rewards.getPending()` | `get_pending_rewards` | | 미인출 보상 조회 |
| 보상 | `sdk.rewards.withdraw()` | `withdraw_rewards` | O | 보상 인출 |
| 범용 | `sdk.cosmos.getBalance()` | `get_balance` | | 잔고 조회 |
| 범용 | `sdk.cosmos.sendTokens()` | `send_tokens` | O | 토큰 전송 |
| 범용 | `sdk.cosmos.getBlockInfo()` | `get_block_info` | | 블록 정보 |
| 범용 | `sdk.cosmos.getTxResult()` | `get_tx_result` | | TX 결과 조회 |

총 13개 함수/도구: 쿼리 9개, TX 4개
