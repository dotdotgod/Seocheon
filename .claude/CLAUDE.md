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
│   ├── 07_tokenomics.md          ← 토큰 이코노믹스, Genesis 노드 (이코노미)
│   ├── 08_spam_defense.md        ← 스팸/게이밍 방어 (보안)
│   ├── 09_implementation.md      ← API, 모듈 구조, 로드맵, 테스트 (전체 팀)
│   ├── 10_chain_upgrade.md       ← 체인 업그레이드 전략 (체인 코어/DevOps)
│   ├── 11_circuit_breaker.md     ← 긴급 정지 메커니즘 (체인 코어/보안)
│   ├── 12_ibc_strategy.md        ← IBC 전략 (인프라/네트워크)
│   └── 13_indexer_architecture.md ← 인덱서 아키텍처 (인프라/백엔드)
├── sdk_specification.md          ← Client SDK 설계 스펙 인덱스 (목차)
├── sdk/                          ← SDK 설계 스펙 분리 문서
│   ├── README.md                 ← 전체 목차 + 담당별 분류
│   ├── 01_architecture.md        ← 계층 구조, 5모듈 구조, 설계 원칙 (SDK 아키텍트)
│   ├── 02_interfaces.md          ← 클래스/인터페이스/데이터 타입 (SDK 개발자)
│   ├── 03_methods.md             ← 13개 메서드 시그니처 (SDK 개발자)
│   ├── 04_constants.md           ← 전역 상수, 에러 코드, JSON Schema (SDK/체인 코어)
│   ├── 05_communication.md       ← gRPC/REST 엔드포인트, TX 플로우 (SDK/인프라)
│   ├── 06_testing.md             ← 테스트 시나리오, E2E (QA/SDK)
│   ├── 07_mock_data.md           ← Proto 응답 Mock, 경계값 Mock (QA/SDK)
│   └── 08_events.md              ← 온체인 이벤트 타입 (체인 코어/SDK)
├── agent_architecture.md         ← 오프체인 에이전트 참고 아키텍처
├── mcp_server_architecture.md    ← MCP 서버 아키텍처 (seocheon-server, vault-server)
├── foundation_strategy.md        ← 재단 운영 전략
└── architecture_review.md        ← 아키텍처 리뷰 및 보완 사항
docs/
├── docs.go                       ← Cosmos SDK API 문서 서비스
└── static/openapi.json           ← OpenAPI 스펙
```

## 소스코드 구조

```
x/
├── node/                                ← x/node 모듈 (노드 등록/관리)
│   ├── ante/                            ← AnteHandler 데코레이터
│   │   ├── agent_permission_decorator.go    ← Agent 주소 권한 검증
│   │   └── registration_fee_decorator.go    ← 등록비 처리
│   ├── client/cli/                      ← 커스텀 CLI (register-node --pubkey)
│   ├── keeper/                          ← 비즈니스 로직
│   │   ├── msg_server_*.go              ← Msg 핸들러 (메시지당 1파일)
│   │   ├── query_*.go                   ← Query 핸들러
│   │   ├── feegrant.go                  ← Feegrant 자동 부여
│   │   ├── hooks.go                     ← Staking Hooks (Active 상태 전환)
│   │   ├── abci.go                      ← EndBlocker (에포크 전환)
│   │   └── invariants.go               ← Registration Pool 불변식
│   ├── module/                          ← AppModule, AutoCLI, Depinject
│   └── types/                           ← 에러, 이벤트, 파라미터, Store Key
├── activity/                            ← x/activity 모듈 (Activity Protocol)
│   ├── keeper/                          ← 비즈니스 로직
│   │   ├── msg_server_submit_activity.go    ← MsgSubmitActivity 핸들러
│   │   ├── reward_distribution.go       ← 활동 보상 일괄 분배 (All-or-nothing)
│   │   ├── reward_split.go              ← 노드/에이전트 수수료 분할 (80/20)
│   │   ├── fee_model.go                 ← Feegrant 쿼터 우선 적용
│   │   ├── epoch_utils.go               ← 에포크/윈도우 유틸
│   │   └── abci.go                      ← EndBlocker (윈도우/에포크 전환)
│   ├── module/                          ← AppModule, AutoCLI, Depinject
│   └── types/                           ← 에러, 이벤트, 파라미터, Store Key
app/                                     ← 앱 초기화
├── app.go                               ← App 구조체, 모듈 등록
├── ante.go                              ← AnteHandler 체인 구성
├── testnet_fixture.go                   ← E2E 테스트용 TestNetworkFixture
cmd/seocheon/cmd/                        ← CLI 명령어 (아래 "CLI 도구" 참조)
proto/seocheon/                          ← Protobuf 정의
├── node/v1/                             ← x/node 메시지, 쿼리, Genesis
└── activity/v1/                         ← x/activity 메시지, 쿼리, Genesis
tests/e2e/                               ← E2E 테스트 (6개, ~5.5분)
testutil/                                ← 샘플 데이터, 테스트넷 프로필
docker/                                  ← Dockerfile, docker-compose.testnet.yml
scripts/                                 ← genesis_build.sh
```

## 기술 스택

### 블록체인
- **Cosmos SDK v0.53+** / CometBFT
- **Go**: 1.24.1
- **토큰**: KKOT (꽃, 단일 토큰) — base denom: `usum` (숨), milli: `hon` (혼), display: `kkot` (꽃)
- **합의**: DPoS, Active Validator Set 150~200
- **해시 알고리즘**: SHA-256

### 개발 도구
- **빌드**: Makefile, Ignite CLI (protobuf 코드 생성)
- **Protobuf**: buf v2 (Cosmos SDK, IBC, gogo-proto 의존성)
- **CI/CD**: GitHub Actions (unit + e2e 분리 파이프라인)
- **컨테이너**: Docker, Docker Compose (testnet 배포)

## 주요 온체인 모듈

| 모듈 | 역할 |
|------|------|
| `x/node` | 노드 등록/관리, Registration Pool |
| `x/activity` | MsgSubmitActivity 제출, 프루닝 |
| `x/distribution` 확장 | 동적 이중 보상 풀: `delegation_ratio = max(D_min, N_d/(N_a+N_d))`, D_min=0.3 |

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

### Go (Cosmos SDK)
- **모듈 디렉토리 구조**: `keeper/`, `types/`, `module/`, `ante/`, `client/cli/` 패턴
- **Msg 핸들러**: 메시지당 1파일 분리 (`msg_server_{action}.go` + `msg_server_{action}_test.go`)
- **Query 핸들러**: 쿼리당 1파일 분리 (`query_{subject}.go` + `query_{subject}_test.go`)
- **에러 정의**: `types/errors.go`에 `errors.Register(ModuleName, code, msg)`. 코드 범위: x/node 1100~, x/activity 1200~
- **이벤트 정의**: `types/events.go`에 `EventType*`, `AttributeKey*` 상수
- **Ante 데코레이터**: `ante/` 디렉토리, `AnteHandle(ctx, tx, simulate, next)` 시그니처
- **Collections**: `collections.Map[K,V]`, `collections.Item[T]`, `collections.KeySet[K]` (Keeper struct)
- **ABCI**: `keeper/abci.go`에 `EndBlocker(ctx) error`. 에포크/윈도우 경계 로직
- **Depinject**: `module/depinject.go`에 `ProvideModule(ModuleInputs) ModuleOutputs` 패턴
- **AutoCLI**: `module/autocli.go`. 복잡한 플래그는 `Skip: true` + `client/cli/` custom CLI
- **Proto 파일**: `proto/seocheon/{module}/v1/*.proto`, `option go_package = "seocheon/x/{module}/types"`
- **테스트**: `keeper_test.go`에 mock keeper struct + 팩토리 함수 (`newMockAuthKeeper()` 등)

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

## 빌드 및 테스트

### 빌드
```bash
make install          # 바이너리 빌드 및 설치 ($GOPATH/bin/seocheon)
make proto-gen        # Protobuf 코드 생성 (Ignite)
make lint             # golangci-lint 실행
```

### 테스트
```bash
make test             # 전체 테스트 (govet, govulncheck, unit)
make test-unit        # 유닛 테스트만 (30분 타임아웃)
make test-race        # Race condition 검증
make test-cover       # 커버리지 리포트
```

### E2E 테스트
```bash
go test -v -timeout 20m -count=1 ./tests/e2e/...
```
`app.TestNetworkFixture` 기반 인메모리 테스트넷. CI에서 unit (15분), e2e (30분) 별도 실행.

## CLI 도구

기본 Cosmos SDK CLI 외 커스텀 도구:

| 명령어 | 파일 | 용도 |
|--------|------|------|
| `seocheon testnet` | testnet.go | 단일 노드 로컬 테스트넷 초기화 |
| `seocheon multi-node` | testnet_multi_node.go | 멀티 밸리데이터 테스트넷 (Docker Compose) |
| `seocheon genesis-build` | genesis_builder.go | 프로덕션 Genesis 파일 생성 |
| `seocheon genesis-airdrop` | genesis_airdrop.go | Genesis 에어드랍 CSV → Bank balances 추가 |
| `seocheon airdrop-snapshot` | airdrop_snapshot.go | 활동 노드 조회 → 균등 분배 스냅샷 |
| `seocheon simulate-activity` | simulate_activity.go | 테스트넷용 합성 MsgSubmitActivity 제출 |
| `seocheon verify-rewards` | verify_rewards.go | 에포크 활동 보상 분배 검증 |
| `seocheon tx node register-node` | x/node/client/cli/ | 노드 등록 TX (--pubkey 플래그) |

## 미해소 사항

`architecture_review.md` 참조:
- Client SDK 구현 (상) — 설계 스펙 완료 (`sdk/`), MCP 서버 기반 계층. 구현 대기
- 스케일링 전략 (중) — 노드 수 10만+ 시 TPS 병목 대응 (IBC 샤딩, L2), 장기 과제
- 에이전트 기억 체계 구현 상세 (하) — off-chain 참고 아키텍처, 운영자 자유
