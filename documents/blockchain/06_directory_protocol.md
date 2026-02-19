# Node Directory Protocol (서비스 디스커버리)

> **담당**: CosmWasm 컨트랙트 개발자 (Rust)
> **의존**: x/node (Stargate Query), x/wasm (CosmWasm 런타임)
> **관련 문서**: [노드 모듈](03_node_module.md) · [구현 가이드](09_implementation.md) · [IBC 전략](12_ibc_strategy.md) · [전체 목차](README.md)

노드가 자신의 능력(capabilities)과 인터페이스를 공개하여, 다른 에이전트가 프로그래밍적으로 협업 대상을 탐색할 수 있는 CosmWasm 컨트랙트이다.

### 설계 원칙

- **별도 CosmWasm 컨트랙트**: 독립적으로 배포·업그레이드 가능
- **에이전트용 디스커버리**: `x/node`의 `tags`가 인간 대시보드용 라벨이라면, 디렉토리의 `capabilities`는 에이전트가 자동으로 탐색하는 구조화된 데이터
- **Stargate Query**: `x/node` 모듈에 노드 등록 여부를 확인

### x/node tags와의 역할 구분

```
x/node (커스텀 모듈):
  tags: ["ai-agent", "data-analysis", "korean"]
  → 인간용 라벨, 위임자 대시보드에서 필터링
  → 자유 텍스트, 비구조적

Node Directory (CosmWasm 컨트랙트):
  capabilities: ["text_analysis", "image_generation", "translation"]
  interfaces: [{ protocol: "mcp", version: "1.0", schema_uri: "ipfs://..." }]
  → 에이전트용 디스커버리, 프로그래밍적 탐색
  → 구조화된 데이터, 인터페이스 스펙 포함
```

### 컨트랙트 메시지

```rust
// 실행 메시지
enum ExecuteMsg {
    // 프로필 업데이트 (capabilities, 상태 URI 등)
    UpdateProfile {
        capabilities: Option<Vec<String>>,  // 에이전트 능력 목록
        profile_uri: Option<String>,        // 상세 프로필 (오프체인)
    },
    // 인터페이스 스펙 등록
    RegisterInterface {
        interface_id: String,               // 고유 식별자 (예: "mcp-data-analysis-v1")
        protocol: String,                   // 프로토콜 타입 (예: "mcp", "http", "grpc")
        version: String,                    // 버전
        schema_uri: String,                 // 인터페이스 스키마 (오프체인 JSON)
        accepted_topics: Vec<String>,       // 수신 가능한 메시지 topic 목록
    },
    // 인터페이스 제거
    RemoveInterface {
        interface_id: String,
    },
    // 노드 상태 설정
    SetStatus {
        status: NodeStatus,                 // Available | Busy | Offline
    },
}

enum NodeStatus {
    Available,      // 서비스 수용 가능
    Busy,           // 현재 작업 중 (새 요청 제한)
    Offline,        // 오프라인 (요청 거부)
}

// 조회 메시지
enum QueryMsg {
    // 특정 노드의 프로필
    ProfileByNode { node: String },
    // 특정 능력을 가진 노드 목록
    NodesByCapability { capability: String, status: Option<NodeStatus>, limit: u32 },
    // 특정 인터페이스를 지원하는 노드 목록
    NodesByInterface { protocol: String, limit: u32 },
    // 특정 인터페이스 스펙 조회
    InterfaceSpec { node: String, interface_id: String },
    // 전체 등록된 능력 목록 (인덱싱용)
    AllCapabilities { limit: u32, start_after: Option<String> },
}
```

### 저장 모델

```rust
struct NodeProfile {
    node_address: Addr,                // 노드 agent_address
    capabilities: Vec<String>,         // 능력 목록
    status: NodeStatus,                // 현재 상태
    profile_uri: String,               // 상세 프로필 오프체인 위치
    updated_block: u64,                // 마지막 업데이트 블록
}

struct InterfaceSpec {
    interface_id: String,              // 고유 식별자
    protocol: String,                  // 프로토콜 타입
    version: String,                   // 버전
    schema_uri: String,                // 스키마 오프체인 위치
    accepted_topics: Vec<String>,      // 수신 가능한 topic
}
```

### 인터페이스 스펙 표준 (오프체인 JSON)

`schema_uri`가 가리키는 오프체인 JSON 형식의 권장 표준:

```json
{
  "interface_id": "mcp-data-analysis-v1",
  "protocol": "mcp",
  "version": "1.0",
  "description": "온체인 데이터 분석 서비스",
  "endpoint": "https://node-b.example.com/mcp",
  "methods": [
    {
      "name": "analyze_chain",
      "description": "Cosmos 체인 활동 데이터 분석",
      "params": {
        "chain_id": { "type": "string", "required": true },
        "timeframe": { "type": "string", "enum": ["24h", "7d", "30d"] }
      },
      "returns": { "type": "object", "schema_uri": "ipfs://Qm..." }
    }
  ],
  "pricing": {
    "model": "per_request",
    "base_fee": "1000000usum"
  }
}
```

**표준화 참고**: 인터페이스 스펙은 체인이 강제하지 않는다. off-chain 권장 표준으로만 존재하며, 에이전트 프레임워크에서 채택하는 형태이다.

### 디스커버리 워크플로우 예시

```
Node A가 데이터 분석 서비스를 찾는 플로우:

1. QueryMsg::NodesByCapability { capability: "data_analysis", status: Available }
   → [Node B, Node D, Node F]

2. QueryMsg::InterfaceSpec { node: "nodeB", interface_id: "mcp-data-analysis-v1" }
   → { protocol: "mcp", schema_uri: "ipfs://...", accepted_topics: ["data_analysis"] }

3. Node A가 schema_uri에서 인터페이스 상세 확인
   → 호환 가능 판단

4. Node A가 오프체인에서 Node B와 직접 통신하여 서비스 요청 (MCP 프로토콜 등)
```
