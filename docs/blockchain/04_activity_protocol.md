# Activity Protocol (작업 내역 공개 프로토콜)

> **담당**: x/activity 모듈 개발자 (Go)
> **의존 모듈**: x/node (agent_address 검증)
> **관련 문서**: [핵심 개념](02_core_concepts.md) · [노드 모듈](03_node_module.md) · [스팸 방어](08_spam_defense.md) · [구현 가이드](09_implementation.md) · [Circuit Breaker](11_circuit_breaker.md) · [IBC 전략](12_ibc_strategy.md) · [인덱서 아키텍처](13_indexer_architecture.md) · [전체 목차](README.md)

> **면책 조항**: 이 문서는 프로토콜 메커니즘에 대한 기술 설계 문서이며 투자 권유가 아니다. 보상 분배는 프로토콜 규칙에 따라 자동 실행되며, 어떠한 형태의 수익도 보장하지 않는다.

Seocheon의 차별점. 노드가 자신의 활동을 표준화된 형식으로 공개하는 프로토콜이다.

### 설계 원칙

- **체인은 검증하지 않는다**: 내용의 진위, 유용성, 품질을 판단하지 않음
- **표준 형식만 제공**: 위임자가 노드를 비교할 수 있는 공통 데이터 구조
- **온체인은 해시만**: 작업 내역의 해시와 참조 URI만 온체인에 기록
- **상세 내용은 오프체인**: IPFS, Arweave, 자체 서버 등 운영자가 선택

### 타임스탬핑 모델

Activity Protocol의 핵심은 **타임스탬핑**이다. 활동 해시를 표준 TX로 제출하면, 해당 TX가 블록에 포함되는 순간 "이 활동 해시가 블록 #N에 기록됨"이 증명된다.

```
AI 에이전트 (오프체인)
  → 활동 수행 → 결과 생성
  → 활동 내역 해시 계산
  → MsgSubmitActivity TX 제출
  → 블록에 포함됨 (블록 #N)
  → 누구나 검증 가능: "이 활동이 블록 #N 시점에 기록됨"
```

별도의 Vote Extension이나 블록 레벨 변경 없이, **표준 Cosmos SDK TX가 블록에 포함되는 것 자체가 타임스탬프**이다.

### 온체인 데이터

```protobuf
// 활동 기록 (타임스탬핑)
message ActivityRecord {
  string submitter = 1;          // 제출자 (에이전트 지갑 주소)
  bytes activity_hash = 2;       // 활동 해시 (구성 방법은 제출자 자유)
  string content_uri = 3;        // 오프체인 조회 위치 (IPFS CID, URL 등)
}
```

3개 필드:
- `submitter`: 누가 제출했는가 (Cosmos SDK 서명 검증으로 자동 인증)
- `activity_hash`: 어떤 활동인가 (불투명 해시. 단일 해시든 머클루트든 제출자가 자유롭게 구성)
- `content_uri`: 어디서 확인할 수 있는가

체인이 검증하는 것:
- TX 서명이 유효한가 (Cosmos SDK AnteHandler)
- 제출자가 등록된 노드 운영자인가
- 중복 제출이 아닌가 (같은 activity_hash)

### 해시 구성은 에이전트의 자유

체인은 `activity_hash`의 구성 방법을 강제하지 않는다. 에이전트가 단일 활동의 SHA-256 해시를 제출하든, 여러 활동을 머클트리로 묶어 루트를 제출하든, 자체 방식으로 해시를 계산하든 체인은 관여하지 않는다.

```
예시 — 에이전트의 해싱 전략 (모두 유효):

방식 A: 단일 해시
  activity_hash = SHA-256(활동 데이터)

방식 B: 머클루트
  activity_hash = MerkleRoot([hash_1, hash_2, hash_3])

방식 C: 연결 해시
  activity_hash = SHA-256(hash_1 + hash_2 + hash_3)
```

**설계 원칙**:

- **온체인은 해시만 저장**: 불투명 바이트 배열. 구성 방법 무관
- **쿼터는 TX 기준**: 에포크당 MsgSubmitActivity TX 수로 차감
- **검증은 오프체인**: 위임자/인덱서가 content_uri의 데이터와 activity_hash를 대조

```
오프체인 검증 플로우:
  1. 인덱서/위임자가 content_uri에서 활동 데이터 다운로드
  2. 동일한 해싱 방식으로 해시 재계산
  3. 재계산한 해시 == 온체인 activity_hash 확인
```

### 오프체인 데이터 (Activity Report)

content_uri가 가리키는 오프체인 데이터의 **권장 형식**. 강제는 아니지만 대시보드와 인덱서가 이 형식을 기대한다:

```json
{
  "version": "1.0",
  "node_id": "seocheon1abc...",
  "period": {
    "from": "2026-01-01T00:00:00Z",
    "to": "2026-01-07T00:00:00Z"
  },
  "activities": [
    {
      "title": "Cosmos 생태계 온체인 데이터 분석",
      "description": "주요 Cosmos 체인의 활동 지표 및 네트워크 참여율 비교 분석",
      "tags": ["cosmos", "analysis", "onchain"],
      "artifacts": [
        {
          "type": "report",
          "uri": "ipfs://Qm...",
          "hash": "sha256:abc..."
        }
      ],
      "started_at": "2026-01-03T10:00:00Z",
      "completed_at": "2026-01-03T14:30:00Z"
    }
  ],
  "metadata": {
    "tools_used": ["web_search", "data_analysis"],
    "agent_type": "ai_agent",
    "additional": {}
  }
}
```

- `activities`: 기간 내 수행한 작업 목록
- `artifacts`: 작업 결과물 (보고서, 데이터, 코드 등)의 해시와 참조
- `metadata`: 사용한 도구, 에이전트 유형 등 자유 형식 메타데이터
- 모든 필드는 자기 신고 — 체인은 검증하지 않음

### 콘텐츠 책임 및 면책

체인은 `activity_hash`(SHA-256 해시)와 `content_uri`(포인터)만 저장한다. 해시에서 원본 콘텐츠를 복원할 수 없으며, 실제 콘텐츠는 오프체인(IPFS, Arweave, 자체 서버 등)에 노드 운영자가 호스팅한다.

```
온체인 저장 범위:
  activity_hash  →  불투명 바이트 배열 (SHA-256). 원본 복원 불가
  content_uri    →  오프체인 위치 포인터 (URL, IPFS CID)
  submitter      →  제출자 주소

  ※ 콘텐츠 자체는 온체인에 저장되지 않는다
```

**책임 구분**:

- 체인은 콘텐츠를 호스팅하지 않으며, 기술적으로 콘텐츠를 검증할 수 없다. 이는 설계 의도이다
- `content_uri`가 가리키는 콘텐츠의 합법성, 정확성, 적절성은 전적으로 **제출자(노드 운영자)**의 책임이다
- 노드 운영자는 관할 법률(저작권, 명예훼손, 개인정보보호 등)을 준수하여 콘텐츠를 관리해야 한다
- 불법 콘텐츠 관련 분쟁은 해당 콘텐츠를 제출한 노드 운영자와 호스팅 인프라 사이의 문제이며, 프로토콜과 무관하다

**긴급 대응**: 사법 당국의 요청이나 커뮤니티 신고가 있는 경우, 거버넌스 절차 또는 Circuit Breaker([긴급 정지 메커니즘](11_circuit_breaker.md))를 통해 해당 노드에 대한 제재를 진행할 수 있다. 이는 콘텐츠 검열이 아닌 **노드 수준의 거버넌스 조치**이다
