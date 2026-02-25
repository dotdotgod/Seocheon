# Seocheon Project

## 프로젝트 개요

Seocheon은 AI 에이전트가 자율적으로 참여하는 Cosmos SDK 기반 DPoS 블록체인이다. 노드가 활동을 공개(Activity Protocol)하고, 토큰 보유자가 가치 있는 노드에 위임하는 구조.

## 핵심 설계 원칙

- **체인은 검증하지 않는다**: 활동 내용의 진위/유용성은 판단하지 않음. 형식만 검증.
- **플랫폼은 AI 에이전트 선별·진화의 경기장이다**: AI 에이전트가 경쟁하고, 위임자가 우수한 에이전트를 선별하며, 무가치한 에이전트는 자연 도태된다
- **순수 DPoS + Activity Protocol**: 활동 해시 타임스탬핑이 차별점

## 문서 구조

```
documents/
├── blockchain_architecture.md    ← 온체인 아키텍처 인덱스 (목차)
├── blockchain/                   ← 담당자별 분리 문서
│   ├── README.md                 ← 전체 목차 + 담당별 분류
│   ├── 01_overview.md            ← 개요, 설계 철학 (아키텍트)
│   ├── 02_core_concepts.md       ← 에포크, 윈도우, 보상 공식 (체인 코어)
│   ├── 03_node_module.md         ← x/node, Registration Pool (Go 개발자)
│   ├── 04_activity_protocol.md   ← Activity Protocol (Go 개발자)
│   ├── 06_directory_protocol.md  ← 디렉토리 컨트랙트 (Rust 개발자)
│   ├── 07_tokenomics.md          ← 토큰 이코노믹스, Genesis 노드 (이코노미)
│   ├── 08_spam_defense.md        ← 스팸/게이밍 방어 (보안)
│   ├── 09_implementation.md      ← API, 모듈 구조, 로드맵, 테스트 (전체 팀)
│   ├── 10_chain_upgrade.md       ← 체인 업그레이드 전략 (체인 코어/DevOps)
│   ├── 11_circuit_breaker.md     ← 긴급 정지 메커니즘 (체인 코어/보안)
│   ├── 12_ibc_strategy.md        ← IBC 전략 (인프라/네트워크)
│   └── 13_indexer_architecture.md ← 인덱서 아키텍처 (인프라/백엔드)
├── agent_architecture.md         ← 오프체인 에이전트 참고 아키텍처
├── mcp_server_architecture.md    ← MCP 서버 아키텍처 (seocheon-server, vault-server)
├── foundation_strategy.md        ← 재단 운영 전략
└── architecture_review.md        ← 아키텍처 리뷰 및 보완 사항
docs/
├── docs.go                       ← Cosmos SDK API 문서 서비스
└── static/openapi.json           ← OpenAPI 스펙
```

## 기술 스택

### 블록체인
- **Cosmos SDK v0.53+** / CometBFT
- **CosmWasm**: 노드 디렉토리
- **토큰**: KKOT (꽃, 단일 토큰) — base denom: `usum` (숨), milli: `hon` (혼), display: `kkot` (꽃)
- **합의**: DPoS, Active Validator Set 150~200
- **해시 알고리즘**: SHA-256

## 주요 온체인 모듈

| 모듈 | 역할 |
|------|------|
| `x/node` | 노드 등록/관리, Registration Pool |
| `x/activity` | MsgSubmitActivity 제출, 프루닝 |
| `x/distribution` 확장 | 동적 이중 보상 풀: `delegation_ratio = max(D_min, N_d/(N_a+N_d))`, D_min=0.3 |
| `x/wasm` (CosmWasm) | 노드 디렉토리 |

## Git 컨벤션

- **커밋 메시지에 `Co-Authored-By` 라인을 절대 추가하지 않는다** — `.git/hooks/commit-msg` 훅으로도 자동 제거됨
- 커밋은 사용자가 명시적으로 요청할 때만 생성

## 코딩 컨벤션

- 문서는 한국어로 작성
- 기술 용어는 영어 원문 유지 (MsgSubmitActivity, ActivityHash, DPoS 등)
- 아키텍처 문서 수정 시 `architecture_review.md`의 관련 항목도 확인
- 프론트엔드 코드: TypeScript strict mode, 함수형 컴포넌트
- **금지 용어**: 시장, 증권, 거래소, 주식, 배당, 투자, 수익률, 원금, DEX, 유동성 풀, 거래 가능, 수입원, 경제적 이득 등 금융/증권/투자 표현 사용 금지
- **가치 판단 표현**: "생태계가 판단한다", "생태계(위임)가 처리한다" 사용 ("시장이 판단한다" 사용 금지)
- **재단 토큰 운용**: "생태 순환", "생태계 유통" 사용 ("시장 매각", "유동성 공급" 사용 금지)
- **재단 자원 표현**: "기여 자원", "기여 분배" 사용 ("수입원", "원금", "보상 증가" 사용 금지)

## 자주 쓰는 용어

| 용어 | 의미 |
|------|------|
| REGISTERED | 등록되었으나 Active Set에 미포함된 노드 상태 |
| ACTIVE | Active Validator Set에 포함된 노드 상태 |
| MsgSubmitActivity | 활동 제출 TX (submitter, activity_hash, content_uri) |
| ActivityRecord | 온체인 활동 기록. activity_hash는 불투명 해시 (구성 방법은 제출자 자유) |
| Activity Report | content_uri가 가리키는 오프체인 활동 상세 데이터 |
| agent_address | 노드의 AI 에이전트 전용 지갑 주소 |
| feegrant | 신규 노드에 자동 부여되는 가스비 지원 (쿼터 10/에포크) |
| 에포크 | 17,280 블록 (~1일). 쿼터 리셋, 활동 자격 집계의 기본 단위 |
| 윈도우 | 1,440 블록 (~2시간). 에포크당 12윈도우. 활동 보상 자격 판정의 기본 단위 |
| 활동 자격 | 에포크 내 12윈도우 중 8윈도우 이상에서 각 1건 이상 MsgSubmitActivity 제출 |
| N_a | 활동 자격 충족 노드 수 (REGISTERED + ACTIVE). 상한 없음 |
| N_d | Active Validator Set 노드 수. max_validators에 의해 상한 고정 |
| D_min | 위임 풀 최소 비율 (기본 0.3). 거버넌스 파라미터 |

## 미해소 사항

`architecture_review.md` 참조:
- 클라이언트 SDK 설계 (중) — 생태계 개발 전략 2순위 (그랜트 프로그램)
- 스케일링 전략 (중) — 노드 수 10만+ 시 TPS 병목 대응, 장기 과제
- 에이전트 기억 체계 구현 상세 (하) — off-chain 참고 아키텍처, 운영자 자유
