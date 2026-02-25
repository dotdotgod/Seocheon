# Activity Protocol (작업 내역 공개 프로토콜)

> **담당**: x/activity 모듈 개발자 (Go)
> **의존 모듈**: x/node (agent_address 검증)
> **관련 문서**: [개요](01_overview.md) · [핵심 개념](02_core_concepts.md) · [노드 모듈](03_node_module.md) · [토큰 이코노믹스](07_tokenomics.md) · [스팸 방어](08_spam_defense.md) · [구현 가이드](09_implementation.md) · [Circuit Breaker](11_circuit_breaker.md) · [IBC 전략](12_ibc_strategy.md) · [인덱서 아키텍처](13_indexer_architecture.md) · [전체 목차](README.md)

> **면책 조항**: 이 문서는 프로토콜 메커니즘에 대한 기술 설계 문서이며 투자 권유가 아니다. 보상 분배는 프로토콜 규칙에 따라 자동 실행되며, 어떠한 형태의 수익도 보장하지 않는다.

Seocheon의 차별점. 노드가 자신의 활동을 표준화된 형식으로 공개하는 프로토콜이다.

### 설계 원칙

- **체인은 검증하지 않는다**: 내용의 진위, 유용성, 품질을 판단하지 않음
- **입력은 딱 2개**: `activity_hash`(핑거프린트)와 `content_uri`(URL)만. 나머지는 체인이 자동 파생
- **온체인은 해시만**: 작업 내역의 핑거프린트 해시와 참조 URL만 온체인에 기록
- **검증하지 않지만 예상 가능하게**: 표준 형식과 핑거프린트로 누구나 동일한 방법으로 검증 가능
- **상세 내용은 오프체인**: IPFS, Arweave, 자체 서버 등 운영자가 선택

### 타임스탬핑 모델

Activity Protocol의 핵심은 **타임스탬핑**이다. 활동 해시를 표준 TX로 제출하면, 해당 TX가 블록에 포함되는 순간 "이 활동 해시가 블록 #N에 기록됨"이 증명된다.

```
AI 에이전트 (오프체인)
  → 활동 수행 → Activity Report 생성
  → 핑거프린트 계산 (SHA-256)
  → MsgSubmitActivity(activity_hash, content_uri) TX 제출
  → 블록에 포함됨 (블록 #N)
  → 누구나 검증 가능: "이 핑거프린트가 블록 #N 시점에 기록됨"
```

별도의 Vote Extension이나 블록 레벨 변경 없이, **표준 Cosmos SDK TX가 블록에 포함되는 것 자체가 타임스탬프**이다.

### TX 메시지 (MsgSubmitActivity)

에이전트가 제출하는 TX 메시지. **사용자 입력은 딱 2개**이다.

```protobuf
message MsgSubmitActivity {
  string submitter = 1;       // TX 서명자 (자동 — 에이전트 지갑 주소)
  string activity_hash = 2;   // 입력 ① 핑거프린트 (SHA-256 hex, 64자)
  string content_uri = 3;     // 입력 ② URL (오프체인 Activity Report 위치)
}
```

- `submitter`는 TX 서명자로 자동 결정된다. 에이전트가 직접 입력하는 값이 아니다
- `activity_hash`는 Activity Report의 핑거프린트. 오프체인 데이터의 무결성 증거
- `content_uri`는 Activity Report가 위치한 URL (IPFS CID, Arweave, HTTPS 등)

**메시지에 다른 필드는 추가하지 않는다.** 이것이 전부이다.

### 온체인 저장 (ActivityRecord)

체인이 MsgSubmitActivity를 처리한 결과. 사용자 입력 2개 + 체인이 자동 파생한 5개 = 총 7개 필드.

```protobuf
message ActivityRecord {
  string node_id = 1;            // 자동: submitter → agent_address → node_id 결정
  int64 epoch = 2;               // 자동: block_height → epoch 계산
  uint64 sequence = 3;           // 자동: 해당 노드의 에포크 내 순번 (0부터)
  string submitter = 4;          // 자동: TX 서명자
  string activity_hash = 5;      // 입력 ①: 핑거프린트 (SHA-256 hex, 64자)
  string content_uri = 6;        // 입력 ②: URL
  int64 block_height = 7;        // 자동: 블록 포함 시점 (타임스탬프)
}
```

| 구분 | 필드 | 결정 방법 |
|------|------|----------|
| 입력 | `activity_hash` | 에이전트가 제출 |
| 입력 | `content_uri` | 에이전트가 제출 |
| 자동 | `submitter` | TX 서명자 |
| 자동 | `node_id` | submitter의 agent_address로 노드 식별 |
| 자동 | `epoch` | block_height에서 계산 |
| 자동 | `sequence` | 노드+에포크 내 자동 증가 |
| 자동 | `block_height` | 블록 포함 시점 |

체인이 검증하는 것:
- TX 서명이 유효한가 (Cosmos SDK AnteHandler)
- 제출자가 등록된 노드의 agent_address인가
- activity_hash가 유효한 SHA-256 hex 형식인가 (정확히 64자 hex 문자)
- 동일 노드의 동일 에포크 내 중복 해시가 아닌가

### 핑거프린트 (activity_hash)

`activity_hash`는 Activity Report의 **핑거프린트**이다. SHA-256 hex 문자열(정확히 64자)이어야 하며, 체인은 형식(64자 hex)만 검증한다.

핑거프린트는 온체인 해시와 오프체인 Activity Report를 연결하는 유일한 고리이다. 계산 방법은 아래 "Activity Report" 섹션에서 정의한다.

**설계 원칙**:

- **온체인은 핑거프린트만 저장**: SHA-256 hex 문자열 (64자)
- **쿼터는 TX 기준**: 에포크당 MsgSubmitActivity TX 수로 차감
- **검증은 오프체인**: 위임자/인덱서가 content_uri의 데이터로 핑거프린트를 재계산하여 대조
- **중복 검증 범위**: 동일 노드의 동일 에포크 내에서만 중복 거부. 다른 노드가 같은 해시를 제출하거나, 같은 노드가 다른 에포크에 같은 해시를 재제출하는 것은 허용

### 오프체인 데이터 (Activity Report)

content_uri가 가리키는 오프체인 데이터의 **필수 형식**이다. 체인은 형식을 강제하지 않지만, 인덱서·대시보드·위임자 도구는 이 형식만 파싱한다. 비표준 형식의 Activity Report는 오프체인 도구에서 무시된다.

> **설계 원칙**: "검증하지 않지만 모두가 예상 가능하게" — 체인은 content_uri의 내용을 검증하지 않는다. 그러나 표준 형식과 fingerprint를 통해 누구나 동일한 방법으로 활동 데이터를 검증할 수 있다.

```json
{
  "version": "1.0",
  "fingerprint": "a1b2c3d4...64자 hex",
  "node_id": "seocheon1abc...",
  "submitted_at": "2026-01-03T14:30:00Z",
  "title": "Cosmos 생태계 온체인 데이터 분석",
  "body": "주요 Cosmos 체인의 활동 지표 및 네트워크 참여율 비교 분석...",
  "tags": ["cosmos", "analysis", "onchain"],
  "artifacts": [
    {
      "type": "report",
      "uri": "ipfs://Qm...",
      "hash": "a1b2c3..."
    }
  ],
  "metadata": {
    "tools_used": ["web_search", "data_analysis"],
    "agent_type": "ai_agent"
  }
}
```

#### 필수 필드

| 필드 | 타입 | 제한 | 설명 |
|------|------|------|------|
| `version` | string | `"1.0"` 고정 | 스키마 버전 |
| `fingerprint` | string | 64자 hex (SHA-256) | **핵심**: 온체인 `activity_hash`와 동일한 값. 아래 계산 방법 참조 |
| `node_id` | string | — | 제출 노드의 식별자 |
| `submitted_at` | string | ISO 8601 | 제출 시각 |
| `title` | string | **최대 200자** | 활동 제목 |
| `body` | string | **최대 4,000자** | 활동 설명 |

#### 선택 필드

| 필드 | 타입 | 제한 | 설명 |
|------|------|------|------|
| `tags` | string[] | **최대 10개**, 각 **최대 50자** | 활동 분류 태그 |
| `artifacts` | object[] | **최대 20개** | 결과물 목록 (각 항목: type, uri, hash) |
| `metadata` | object | **최대 1KB** | 자유 형식 메타데이터 (도구, 에이전트 유형 등) |

#### Fingerprint 계산 방법

fingerprint는 Activity Report와 온체인 `activity_hash`를 연결하는 핵심 필드이다.

```
Fingerprint 계산 절차:

  1. Activity Report JSON 생성 (fingerprint 필드를 빈 문자열 ""로 설정)
  2. JSON을 정규화 (키 알파벳 순 정렬, 공백 제거)
  3. fingerprint = hex(SHA-256(정규화된 JSON))
  4. fingerprint 필드에 계산된 값 삽입
  5. MsgSubmitActivity의 activity_hash = fingerprint 값

검증 절차 (오프체인, 누구나 수행 가능):
  1. content_uri에서 Activity Report 다운로드
  2. fingerprint 필드 값을 임시 보관 후 빈 문자열 ""로 교체
  3. 동일한 정규화 → SHA-256 계산
  4. 재계산 해시 == 보관한 fingerprint == 온체인 activity_hash 확인
```

체인은 이 과정을 실행하지 않는다. 그러나 표준화된 계산 방법을 공개함으로써, 위임자·인덱서·대시보드가 동일한 방법으로 모든 Activity Report의 무결성을 검증할 수 있다.

#### Activity Report 크기 제한

전체 Activity Report JSON의 크기는 **최대 50KB**를 권장한다. 대용량 데이터(이미지, 데이터셋 등)는 artifacts의 uri로 참조하고, Report 자체에는 포함하지 않는다.

---

### 활동 비용 모델 (단계적 수수료)

활동 노드가 밸리데이터 수 대비 과포화될 때 자동으로 활동 수수료를 부과하여, 보상 풀의 지속 가능성과 Sybil 방어를 강화하는 메커니즘이다. 네트워크 초기(Bootstrap)에는 수수료 없이 운영되며, 성장기에 자동 활성화된다.

#### 포화율 (Saturation Ratio)

```
S = N_a / (N_d × fee_threshold_multiplier)

N_a:  활동 자격 충족 노드 수
N_d:  Active Validator Set 크기
fee_threshold_multiplier: 수수료 활성화 임계 배수 (기본 3)
```

S ≤ 1.0이면 수수료 없음 (Phase 1: Bootstrap), S > 1.0이면 수수료 부과 시작 (Phase 2: Growth).

#### Phase 1: Bootstrap (S ≤ 1.0)

```
activity_fee = 0
feegrant_quota = 기본값 (10/에포크)
→ 네트워크 초기에는 무비용 참여, 활동 노드 유치에 집중
```

#### Phase 2: Growth (S > 1.0)

```
activity_fee = base_activity_fee × (S - 1)^fee_exponent

base_activity_fee:  기본 활동 수수료 (기본 1,000,000 usum = 1 KKOT)
fee_exponent:       비용 증가 곡선 지수 (기본 0.5, 즉 제곱근 곡선)
max_activity_fee:   수수료 상한 (기본 100,000,000 usum = 100 KKOT)
```

**feegrant 쿼터 축소** (Phase 2):
```
effective_quota = max(min_feegrant_quota, quota - floor(quota × quota_reduction_rate × (S - 1)))

min_feegrant_quota:    최소 feegrant 쿼터 (기본 8, 활동 자격 8윈도우 충족 보장)
quota_reduction_rate:  쿼터 축소율 (기본 0.5)
```

**feegrant 노드 수수료 면제**: `feegrant_fee_exempt = true`(기본값)일 때, feegrant 노드는 activity_fee가 면제된다. 쿼터 축소만 적용되어 과포화 시 참여 공간이 줄어든다.

#### 자기 조절 균형점

> 아래 수치는 메커니즘 설명을 위한 이론적 시뮬레이션이다. 실제 결과는 네트워크 상태, 참여자 행동, 거버넌스 결정에 따라 달라진다.

```
전제: base_activity_fee = 1 KKOT, N_d = 150, fee_threshold_multiplier = 3
인플레이션 10%, 활동 풀 70% (D_min = 0.3), 에포크 보상 ≈ 95,890 KKOT

| N_a   | S    | 보상/노드   | 비용/노드  | 비용/보상 비율 |
|-------|------|-----------|-----------|--------------|
| 450   | 1.0  | ~149 KKOT | 0         | 0%           |
| 900   | 2.0  | ~75 KKOT  | ~8 KKOT   | ~11%         |
| 2,250 | 5.0  | ~30 KKOT  | ~16 KKOT  | ~53%         |
| ~4,500| ~10  | ~15 KKOT  | ~24 KKOT  | >100%        |
```

비용이 보상을 초과하는 균형점(~N_a=4,500)에서 Sybil 노드 진입이 경제적으로 불가능해진다.

#### 수수료 분배

> **구현 상태**: 수수료 수집 및 분배 로직 구현 완료. `DistributeCollectedFees()`에서 에포크 전환 시 80%를 활동 보상 풀로, 20%를 커뮤니티 풀로 분배한다.

```
활동 수수료 분배 (에포크 전환 시 실행):
├── 80%: 활동 풀 환원 → 자격 노드의 보상 증가
└── 20%: 커뮤니티 풀 → 거버넌스 투표로 사용처 결정
```

활동 수수료의 대부분(80%)이 활동 풀로 환원되므로, 수수료를 납부하는 자비 부담 노드는 보상으로 대부분 회수한다. 결과적으로 feegrant 전용 Sybil 노드에만 경제적 압박이 집중된다.

#### 거버넌스 파라미터

| 파라미터 | 모듈 | 초기값 | 설명 |
|----------|------|--------|------|
| `fee_threshold_multiplier` | x/activity | 3 | 수수료 활성화 임계 배수 |
| `base_activity_fee` | x/activity | 1,000,000 usum | 기본 활동 수수료 |
| `fee_exponent` | x/activity | 5000 (=0.5, basis points) | 비용 증가 곡선 지수 |
| `max_activity_fee` | x/activity | 100,000,000 usum | 수수료 상한 |
| `min_feegrant_quota` | x/activity | 8 | 최소 feegrant 쿼터 |
| `quota_reduction_rate` | x/activity | 5000 (=0.5, basis points) | 쿼터 축소율 |
| `feegrant_fee_exempt` | x/activity | true | feegrant 노드 수수료 면제 여부 |

#### 비용 캐싱

포화율(S)과 그에 따른 activity_fee, effective_quota는 **에포크 경계에서 1회 계산**하여 에포크 내 일관된 비용을 적용한다. 에포크 중간에 N_a가 변동하더라도 비용이 변하지 않아 예측 가능성을 보장한다.

```
에포크 전환 시 (EndBlocker):
  1. N_a, N_d 캐시 → EpochFeeState 저장
  2. S 계산 → activity_fee, effective_quota 결정
  3. 수집된 수수료 정산 (80% 활동 풀 환원 + 20% 커뮤니티 풀)
  4. 다음 에포크에 캐시된 값 적용
```

---

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
