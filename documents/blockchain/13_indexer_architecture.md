# 인덱서 상세 아키텍처

> **담당**: 인프라 / 백엔드 개발자
> **관련 문서**: [핵심 개념](02_core_concepts.md) · [노드 모듈](03_node_module.md) · [Activity Protocol](04_activity_protocol.md) · [구현 가이드](09_implementation.md) · [전체 목차](README.md)

> **구현 상태**: 인덱서는 오프체인 인프라로 Phase 3 이후에 구현 예정이다.

## 개요

Seocheon 인덱서는 온체인 이벤트를 수집·가공하여 위임자, 노드 운영자, 대시보드에 구조화된 데이터를 제공하는 오프체인 인프라이다.

```
인덱서의 역할:

1. 온체인 이벤트 수집: x/node, x/activity 모듈의 ABCI Events
2. 데이터 집계: 에포크/윈도우별 활동 통계, 보상 분배 이력
3. content_uri 가용성 모니터링: 오프체인 Activity Report 접근성 추적
4. API 제공: RESTful, GraphQL, WebSocket으로 클라이언트에 데이터 서빙
5. 대시보드 지원: 노드 활동 현황, 네트워크 통계 시각화
```

**설계 원칙**: 인덱서는 체인의 합의에 관여하지 않는 **순수 읽기 인프라**이다. 인덱서가 다운되어도 체인 운영에 영향 없다. 인덱서가 제공하는 데이터(content_uri 가용성 점수 등)는 위임자의 판단을 돕는 **정보성 지표**이지, 체인이 강제하는 기준이 아니다.

---

## 이벤트 기반 인덱싱

### ABCI Events 수집

인덱서는 CometBFT의 Event 시스템을 통해 실시간으로 온체인 이벤트를 수집한다.

```
이벤트 수집 아키텍처:

CometBFT Node
  ├── WebSocket (/websocket)
  │   └── Subscribe: tm.event='Tx' AND message.module='node'
  │   └── Subscribe: tm.event='Tx' AND message.module='activity'
  │   └── Subscribe: tm.event='NewBlock'
  │
  └── RPC (/block_results)
      └── 블록별 이벤트 배치 수집 (catch-up용)

인덱서
  ├── Event Consumer: 실시간 이벤트 수신
  ├── Block Scanner: 누락 블록 복구 (catch-up)
  ├── Event Processor: 이벤트 파싱 + 데이터 변환
  └── Data Writer: PostgreSQL 저장
```

### 이벤트 스키마 정의

각 모듈이 발행하는 ABCI Events와 인덱서의 수집 대상:

**x/node 모듈 이벤트**:

| 이벤트 타입 | 어트리뷰트 | 설명 |
|------------|-----------|------|
| `node_registered` | `node_id`, `operator`, `agent_address`, `tags` | 노드 등록 |
| `node_updated` | `node_id`, `fields_changed` | 노드 정보 변경 |
| `node_deactivated` | `node_id`, `operator` | 노드 비활성화 |
| `node_status_changed` | `node_id`, `old_status`, `new_status` | 상태 전이 (REGISTERED → ACTIVE 등) |
| `agent_share_updated` | `node_id`, `old_share`, `new_share`, `effective_epoch` | agent_share 변경 예약 |

**x/activity 모듈 이벤트**:

| 이벤트 타입 | 어트리뷰트 | 설명 |
|------------|-----------|------|
| `activity_submitted` | `submitter`, `activity_hash`, `content_uri`, `block_height`, `window_number` | 활동 제출 |
| `activity_pruned` | `pruned_count`, `oldest_block_kept` | 프루닝 실행 |
| `epoch_transition` | `epoch_number`, `qualified_nodes`, `total_activity_count` | 에포크 전환 |
| `window_transition` | `window_number`, `active_nodes` | 윈도우 전환 |

### 실시간 vs 배치 인덱싱

```
인덱싱 전략:

실시간 (Streaming):
  → CometBFT WebSocket 구독
  → 블록 확정 즉시 이벤트 수신 + 처리
  → 대시보드, WebSocket 구독자에게 즉시 전달
  → 지연: < 2초 (블록 확정 후)

배치 (Batch):
  → 에포크/윈도우 경계에서 집계 데이터 계산
  → 보상 분배 내역 집계
  → content_uri 가용성 점수 산출
  → 통계 데이터 갱신

복구 (Catch-up):
  → 인덱서 다운타임 후 누락 블록 복구
  → /block_results RPC로 블록별 이벤트 순차 수집
  → 마지막 처리 블록 높이 기록 → 재시작 시 해당 높이부터 복구
```

---

## 데이터 모델

### 노드 활동 이력 테이블

```sql
-- 노드 기본 정보
CREATE TABLE nodes (
    node_id         TEXT PRIMARY KEY,
    operator        TEXT NOT NULL,
    agent_address   TEXT NOT NULL,
    status          TEXT NOT NULL,          -- REGISTERED, ACTIVE, INACTIVE, JAILED
    agent_share     DECIMAL(5,4),
    tags            TEXT[],
    registered_at   BIGINT NOT NULL,        -- 등록 블록 높이
    updated_at      BIGINT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 활동 제출 이력
CREATE TABLE activities (
    id              BIGSERIAL PRIMARY KEY,
    submitter       TEXT NOT NULL,
    node_id         TEXT NOT NULL REFERENCES nodes(node_id),
    activity_hash   BYTEA NOT NULL UNIQUE,
    content_uri     TEXT NOT NULL,
    block_height    BIGINT NOT NULL,
    window_number   BIGINT NOT NULL,
    epoch_number    BIGINT NOT NULL,
    tx_hash         TEXT NOT NULL,
    indexed_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_activities_node_epoch ON activities(node_id, epoch_number);
CREATE INDEX idx_activities_window ON activities(window_number);
CREATE INDEX idx_activities_block ON activities(block_height);
```

### 에포크/윈도우별 집계 데이터

```sql
-- 윈도우별 노드 활동 집계
CREATE TABLE window_activity_summary (
    node_id         TEXT NOT NULL,
    window_number   BIGINT NOT NULL,
    epoch_number    BIGINT NOT NULL,
    submission_count BIGINT DEFAULT 0,
    PRIMARY KEY (node_id, window_number)
);

-- 에포크별 노드 종합 집계
CREATE TABLE epoch_node_summary (
    node_id             TEXT NOT NULL,
    epoch_number        BIGINT NOT NULL,
    total_submissions   BIGINT DEFAULT 0,
    active_windows      INT DEFAULT 0,        -- 활동한 윈도우 수 (8이상이면 자격 충족)
    is_qualified        BOOLEAN DEFAULT FALSE, -- 활동 자격 충족 여부
    delegation_reward   BIGINT DEFAULT 0,      -- 위임 보상 (usum)
    activity_reward     BIGINT DEFAULT 0,      -- 활동 보상 (usum)
    total_reward        BIGINT DEFAULT 0,      -- 총 보상
    PRIMARY KEY (node_id, epoch_number)
);

CREATE INDEX idx_epoch_summary_qualified ON epoch_node_summary(epoch_number, is_qualified);

-- 에포크별 네트워크 집계
CREATE TABLE epoch_network_summary (
    epoch_number        BIGINT PRIMARY KEY,
    total_nodes         INT NOT NULL,
    active_nodes        INT NOT NULL,          -- ACTIVE 상태
    qualified_nodes     INT NOT NULL,          -- 활동 자격 충족
    total_submissions   BIGINT NOT NULL,
    total_delegation    BIGINT NOT NULL,       -- 총 위임량 (usum)
    delegation_pool     BIGINT NOT NULL,       -- 위임 풀 보상 총액
    activity_pool       BIGINT NOT NULL,       -- 활동 풀 보상 총액
    inflation_rate      DECIMAL(5,4)
);
```

---

## content_uri 가용성 모니터링

> **핵심**: 이 섹션의 모든 데이터는 **정보성**이다. 체인이 content_uri의 가용성을 강제하지 않는다. 인덱서가 수집한 가용성 데이터는 위임자가 노드를 평가할 때 참고하는 오프체인 지표이다.

### 헬스체크 주기 및 방법

```
content_uri 헬스체크 전략:

대상: 최근 90일(activity_pruning_keep_blocks) 이내의 ActivityRecord content_uri

방법:
  1. HTTP HEAD 요청 (응답 코드 확인, 본문 미다운로드)
  2. IPFS 게이트웨이를 통한 CID 해석 가능 여부
  3. Arweave TX 존재 여부

주기:
  → 최근 1 에포크(~1일): 매 윈도우(~2시간)마다 체크
  → 1~7일: 1일 1회
  → 7~30일: 1주 1회
  → 30~90일: 1월 1회

타임아웃: 10초 (연결), 30초 (응답)
재시도: 3회 (exponential backoff: 5s, 15s, 45s)
```

### 가용성 점수 산출

```
가용성 점수 산출 공식:

노드 가용성 점수 = 성공 응답 수 / 전체 체크 수 × 100

가중 가용성 점수 (최근 활동에 높은 가중치):
  weighted_score = Σ(check_result × weight) / Σ(weight)

  가중치:
    → 최근 1일: weight = 4
    → 1~7일: weight = 2
    → 7~30일: weight = 1
    → 30일 이후: weight = 0.5

점수 분류:
  95~100%: 우수 (Excellent)
  80~94%:  양호 (Good)
  50~79%:  주의 (Fair)
  0~49%:   불량 (Poor)
```

```sql
-- content_uri 헬스체크 결과
CREATE TABLE content_uri_checks (
    id              BIGSERIAL PRIMARY KEY,
    node_id         TEXT NOT NULL,
    activity_hash   BYTEA NOT NULL,
    content_uri     TEXT NOT NULL,
    check_time      TIMESTAMPTZ NOT NULL,
    status_code     INT,                    -- HTTP 상태 코드 (NULL = 연결 실패)
    response_time_ms INT,                   -- 응답 시간 (ms)
    is_available    BOOLEAN NOT NULL,
    error_message   TEXT                    -- 실패 시 에러 메시지
);

CREATE INDEX idx_uri_checks_node ON content_uri_checks(node_id, check_time);

-- 노드별 가용성 점수 (집계)
CREATE TABLE node_availability_score (
    node_id             TEXT PRIMARY KEY,
    total_checks        BIGINT NOT NULL,
    successful_checks   BIGINT NOT NULL,
    raw_score           DECIMAL(5,2) NOT NULL,  -- 단순 비율
    weighted_score      DECIMAL(5,2) NOT NULL,  -- 가중 점수
    last_checked_at     TIMESTAMPTZ NOT NULL,
    last_calculated_at  TIMESTAMPTZ NOT NULL
);
```

### 가용성 데이터 공개

인덱서는 가용성 데이터를 API와 대시보드를 통해 공개한다. 위임자는 이 데이터를 참고하여 노드의 신뢰성을 판단할 수 있다.

```
가용성 데이터 공개 채널:

1. API: GET /api/v1/nodes/{node_id}/availability
   → 가용성 점수, 최근 체크 이력, 추세

2. 대시보드: 노드 상세 페이지
   → 가용성 점수 배지 (Excellent/Good/Fair/Poor)
   → 가용성 추이 그래프 (30일)
   → 개별 content_uri 상태 목록

3. 비교 뷰: 노드 목록 페이지
   → 가용성 점수 컬럼 (정렬 가능)
   → 위임자가 노드 간 비교 시 참고

주의사항:
  → "체인이 강제하지 않음" 명시
  → 가용성 점수가 낮아도 온체인 보상에 영향 없음 표시
  → IPFS/Arweave의 특성 (전파 지연 등) 안내
```

---

## API 설계

### RESTful API 엔드포인트

```
Base URL: /api/v1

노드:
  GET /nodes                              → 노드 목록 (페이지네이션, 필터, 정렬)
  GET /nodes/{node_id}                    → 노드 상세
  GET /nodes/{node_id}/activities         → 노드 활동 이력
  GET /nodes/{node_id}/rewards            → 노드 보상 이력
  GET /nodes/{node_id}/availability       → content_uri 가용성
  GET /nodes/{node_id}/stats              → 노드 종합 통계

활동:
  GET /activities                         → 전체 활동 목록 (최신순)
  GET /activities/{activity_hash}         → 활동 상세
  GET /activities/epoch/{epoch_number}    → 에포크별 활동

에포크/윈도우:
  GET /epochs                             → 에포크 목록
  GET /epochs/{epoch_number}              → 에포크 상세 (네트워크 집계)
  GET /epochs/{epoch_number}/rewards      → 에포크 보상 분배 내역
  GET /epochs/current                     → 현재 에포크 정보
  GET /windows/current                    → 현재 윈도우 정보

네트워크:
  GET /network/stats                      → 네트워크 종합 통계
  GET /network/staking                    → 스테이킹 현황
  GET /network/inflation                  → 인플레이션 데이터
```

### GraphQL 지원

RESTful API 외에 GraphQL 엔드포인트를 제공한다. 복잡한 관계형 쿼리와 클라이언트 주도의 데이터 선택이 필요한 경우에 활용한다.

```graphql
type Query {
  node(id: ID!): Node
  nodes(filter: NodeFilter, pagination: Pagination): NodeConnection
  activity(hash: String!): Activity
  activitiesByNode(nodeId: ID!, epoch: Int): [Activity!]!
  epoch(number: Int!): Epoch
  currentEpoch: Epoch
  networkStats: NetworkStats
}

type Node {
  id: ID!
  operator: String!
  agentAddress: String!
  status: NodeStatus!
  tags: [String!]!
  activities(epoch: Int, limit: Int): [Activity!]!
  rewards(epoch: Int): [RewardEntry!]!
  availability: AvailabilityScore
  epochSummaries(limit: Int): [EpochNodeSummary!]!
}

type Activity {
  activityHash: String!
  submitter: String!
  node: Node!
  contentUri: String!
  blockHeight: Int!
  windowNumber: Int!
  epochNumber: Int!
  contentAvailable: Boolean
}
```

**GraphQL 도입 시기**: Phase 3(위임 및 보상) 이후 대시보드 프로토타입과 함께 도입. 초기에는 RESTful API만으로 충분하며, 대시보드 요구사항이 복잡해지면 GraphQL을 추가한다.

### WebSocket 실시간 구독

```
WebSocket 엔드포인트: /ws/v1

구독 채널:
  blocks          → 새 블록 알림 (높이, 시간, TX 수)
  activities      → 새 활동 제출 알림
  activities:{node_id}  → 특정 노드의 활동 알림
  epochs          → 에포크/윈도우 전환 알림
  network         → 네트워크 통계 업데이트 (1분 간격)

메시지 포맷:
{
  "channel": "activities:node_123",
  "event": "activity_submitted",
  "data": {
    "activity_hash": "abc...",
    "content_uri": "ipfs://...",
    "block_height": 123456,
    "window_number": 85
  },
  "timestamp": "2026-01-01T00:00:05Z"
}
```

### Rate Limiting 전략

```
Rate Limiting:

Tier 1 — 공개 (인증 없음):
  → 60 req/min (IP 기준)
  → WebSocket 구독 채널: 최대 5개
  → 용도: 일반 사용자, 탐색기

Tier 2 — 등록 (API 키):
  → 300 req/min
  → WebSocket 구독 채널: 최대 20개
  → 용도: 대시보드, 봇, 통합 서비스

Tier 3 — 프리미엄 (노드 운영자):
  → 1,000 req/min
  → WebSocket 구독 채널: 무제한
  → 용도: 노드 운영자, 릴레이어, 인덱서 미러

구현:
  → Redis 기반 Sliding Window Counter
  → X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset 헤더
  → 429 Too Many Requests 응답 + Retry-After 헤더
```

---

## 대시보드 지원

### 노드 활동 현황

```
대시보드 — 노드 활동 현황 페이지:

[노드 목록 뷰]
┌──────────────────────────────────────────────────────────────┐
│ 노드 ID  │ 상태   │ 태그        │ 활동(에포크) │ 가용성  │ 위임량   │
├──────────┼────────┼────────────┼────────────┼────────┼────────┤
│ node_001 │ ACTIVE │ defi, ai   │ 12/12 ✅   │ 98.5%  │ 5.2M   │
│ node_002 │ ACTIVE │ research   │ 9/12 ✅    │ 85.3%  │ 3.1M   │
│ node_003 │ REG    │ analytics  │ 6/12 ❌    │ 72.1%  │ 0      │
└──────────────────────────────────────────────────────────────┘

[노드 상세 뷰]
├── 기본 정보: operator, agent_address, tags, 커미션율, agent_share
├── 활동 타임라인: 윈도우별 활동 제출 현황 (히트맵)
├── 보상 이력: 에포크별 위임 보상 + 활동 보상 그래프
├── 가용성: content_uri 가용성 점수 + 추이 그래프
└── 위임자 목록: 위임량 상위 (거버넌스 참고)
```

### 에포크별 보상 분배

```
대시보드 — 에포크 보상 분배 페이지:

[에포크 개요]
├── 에포크 번호, 시작/종료 블록, 시간 범위
├── 총 블록 보상
├── 위임 풀 (동적 비율, D_min=30%): 총액 + 지분율 상위 노드
├── 활동 풀 (동적 비율): 총액 + 자격 노드 수(N_a) + 노드당 보상
└── 인플레이션율

[보상 분배 상세]
├── 위임 풀 분배: 노드별 지분율 × 위임 풀 총액
├── 활동 풀 분배: 자격 노드 균등 분배 (에포크 전환 시 일괄)
├── 노드 내 분배: 커미션 → operator/agent 분할 → 위임자
└── 자격 미충족 노드 목록 (활동 윈도우 < 8)
```

### 네트워크 통계

```
대시보드 — 네트워크 통계 페이지:

[실시간 지표]
├── Active Validator Set: 현재 수 / MaxValidators
├── REGISTERED 노드 수
├── 총 스테이킹량 / 스테이킹 비율
├── 현재 인플레이션율
├── 현재 에포크 / 윈도우 진행률

[추이 차트 (30일)]
├── 일별 활동 제출 건수
├── 일별 신규 노드 등록 수
├── 스테이킹 비율 변화
├── 활동 자격 충족 노드 비율
└── 평균 content_uri 가용성 점수 추이
```

---

## 운영 고려사항

### 기술 스택

```
인덱서 기술 스택:

런타임: Go 또는 TypeScript (Node.js)
데이터베이스: PostgreSQL 15+
캐시: Redis 7+
메시지 큐: NATS 또는 Redis Streams (이벤트 버퍼링)
API 프레임워크: Echo (Go) 또는 Fastify (Node.js)
GraphQL: gqlgen (Go) 또는 graphql-yoga (Node.js)
모니터링: Prometheus + Grafana
```

### 스케일링 전략

```
스케일링 전략:

수직 (Scale-Up):
  → 초기 단계에서 단일 인스턴스로 운영
  → PostgreSQL: 적절한 인덱스, 쿼리 최적화, 커넥션 풀링
  → 예상 처리량: ~100 TPS (활동 제출 이벤트)

수평 (Scale-Out):
  → 읽기 분리: PostgreSQL Read Replica
  → API 서버: 스테이트리스, 수평 확장 (로드밸런서)
  → 이벤트 처리: 소비자 그룹 (파티셔닝)
  → WebSocket: Sticky Session 또는 Redis Pub/Sub 백플레인

단계별:
  Phase 1 (노드 < 100): 단일 인스턴스, PostgreSQL + Redis
  Phase 2 (노드 100~1,000): Read Replica, API 2~3 인스턴스
  Phase 3 (노드 1,000+): 샤딩 검토, 전용 WebSocket 서버
```

### 캐싱

```
Redis 캐싱 전략:

캐시 대상 및 TTL:
  → 노드 목록 (필터/정렬): TTL 30초
  → 노드 상세: TTL 1분
  → 현재 에포크/윈도우 정보: TTL 10초
  → 에포크 집계 (완료된 에포크): TTL 1시간 (불변 데이터)
  → 네트워크 통계: TTL 1분
  → content_uri 가용성 점수: TTL 5분

캐시 무효화:
  → 이벤트 기반 무효화 (노드 등록/변경 시 관련 캐시 삭제)
  → 에포크/윈도우 전환 시 관련 집계 캐시 갱신
  → TTL 만료에 의한 자연 갱신

캐시 키 패턴:
  node:{node_id}
  nodes:list:{filter_hash}
  epoch:{number}:summary
  epoch:current
  network:stats
  availability:{node_id}
```

### 데이터 보관 정책

```
데이터 보관 정책 (온체인 프루닝과 연계):

온체인 프루닝:
  → activity_pruning_keep_blocks: 1,555,200 (~90일)
  → 프루닝된 ActivityRecord는 온체인에서 삭제

인덱서 보관:
  → 상세 데이터 (activities): 180일 보관
  → 집계 데이터 (epoch/window summary): 영구 보관
  → content_uri 체크 이력: 90일 보관
  → 가용성 점수 (집계): 영구 보관

아카이브:
  → 180일 초과 상세 데이터: 콜드 스토리지 이관 (선택)
  → 아카이브 API: 별도 엔드포인트 제공 (응답 지연 허용)
  → 풀 아카이브 노드 연동: 프루닝되지 않은 온체인 데이터 접근

데이터 크기 추정 (노드 100개, 에포크당):
  activities: ~100 노드 × ~10 제출/에포크 = ~1,000 행/일
  집계: ~100 노드 × 12 윈도우 = ~1,200 행/일
  content_uri 체크: ~1,000 URI × 12 체크/일 = ~12,000 행/일

  → 연간 약 5.4M 행 (압축 시 ~2GB)
  → 노드 1,000개 시 10배 = ~20GB/년
```

---

## 장애 대응

```
장애 시나리오 및 대응:

인덱서 다운:
  → 체인 운영에 영향 없음 (순수 읽기 인프라)
  → 대시보드/API 서비스 중단
  → 복구 시 catch-up 모드로 누락 블록 자동 복구
  → 마지막 처리 블록 높이를 DB에 기록하여 정확한 복구 시점 보장

데이터베이스 장애:
  → Read Replica 자동 페일오버 (Phase 2 이후)
  → WAL 기반 PITR(Point-In-Time Recovery) 설정
  → 일간 전체 백업 + 연속 WAL 아카이빙

이벤트 누락:
  → WebSocket 연결 끊김 시 자동 재연결 + 누락 블록 스캔
  → 이벤트 처리 실패 시 Dead Letter Queue에 보관 → 재처리
  → 주기적 일관성 검증: 온체인 상태와 인덱서 데이터 대조

모니터링 알림:
  → 인덱싱 지연 > 10 블록: 경고
  → 인덱싱 지연 > 100 블록: 긴급
  → API 응답 시간 > 500ms: 경고
  → content_uri 체크 실패율 > 50%: 알림
  → 디스크 사용량 > 80%: 경고
```
