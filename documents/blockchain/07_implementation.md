# 구현 가이드: API, 모듈 구조, 로드맵

> **담당**: 전체 개발팀 / PM
> **관련 문서**: [개요](01_overview.md) · [핵심 개념](02_core_concepts.md) · [노드 모듈](03_node_module.md) · [Activity Protocol](04_activity_protocol.md) · [토큰 이코노믹스](07_tokenomics.md) · [스팸 방어](08_spam_defense.md) · [체인 업그레이드](10_chain_upgrade.md) · [Circuit Breaker](11_circuit_breaker.md) · [IBC 전략](12_ibc_strategy.md) · [인덱서 아키텍처](13_indexer_architecture.md) · [전체 목차](README.md)

## 거버넌스

```
┌────────────────────────────────────────────────┐
│                거버넌스 액션                      │
├────────────────────────────────────────────────┤
│ 일반 위임/철회:                                  │
│   위임자가 자유롭게 판단 → 별도 투표 불필요          │
│   위임 생태로 자연스럽게 조절                  │
├────────────────────────────────────────────────┤
│ 파라미터 변경 (거버넌스 투표):                      │
│   에포크 길이, 보상 비율, 최소 스테이킹,             │
│   Activity Protocol 버전 등                      │
├────────────────────────────────────────────────┤
│ Activity Protocol 업그레이드 (거버넌스 투표):       │
│   필수 형식 변경, 새 필드 추가 등                   │
│   → 온체인 ActivityRecord 스키마 업그레이드          │
└────────────────────────────────────────────────┘
```

---

## 에이전트 인터페이스: 트랜잭션 유형 및 Query API

### 트랜잭션 유형

```protobuf
// 노드 라이프사이클
message MsgRegisterNode {
  string operator = 1;               // 운영자 지갑 주소 (TX 서명자)
  string agent_address = 2;          // 에이전트 지갑 주소 (활동 제출용)
  string agent_share = 3;            // 커미션 중 agent 몫 (Dec, 0-100%)
  string max_agent_share_change_rate = 4; // 에포크당 agent_share 최대 변경폭
  string description = 5;            // 노드 설명
  string website = 6;                // 운영자 웹사이트
  repeated string tags = 7;          // 노드 태그 (최대 10개, 각 최대 32자)
  google.protobuf.Any consensus_pubkey = 8; // CometBFT 합의 공개키 (ed25519)
  string commission_rate = 9;        // 밸리데이터 커미션율 (Dec, 기본 0%)
  string commission_max_rate = 10;   // 최대 커미션율 (Dec, 기본 100%)
  string commission_max_change_rate = 11; // 에포크당 최대 커미션 변경폭 (Dec, 기본 100%)
}
// 내부적으로 Registration Pool에서 1 usum 대여 → CreateValidator 실행

message MsgUpdateNode {
  string operator = 1;               // TX 서명자 (operator만 가능, 1:1 노드 조회)
  string description = 2;            // 업데이트된 설명 (선택)
  string website = 3;                // 업데이트된 웹사이트 (선택)
  repeated string tags = 4;          // 업데이트된 태그 (선택)
}
// agent_address 변경은 별도 MsgUpdateAgentAddress 사용 (쿨다운 + feegrant 이전)
// agent_share 변경은 별도 MsgUpdateNodeAgentShare 사용 (유예 기간 적용)
// 커미션율 변경은 표준 MsgEditValidator 사용

message MsgUpdateNodeAgentShare {
  string operator = 1;               // TX 서명자 (operator만 가능, 1:1 노드 조회)
  string new_agent_share = 2;        // 변경 희망값
}
// max_agent_share_change_rate 초과 시 거부
// 유예 기간 1 에포크(17,280 블록) 후 적용, 대시보드에 "변경 예정" 표시

message MsgWithdrawNodeCommission {
  string operator = 1;               // TX 서명자 (operator만 가능, 1:1 노드 조회)
}
// 내부: WithdrawValidatorCommission 호출 후 agent_share 비율로 분할 전송

message MsgDeactivateNode {
  string operator = 1;               // TX 서명자 (operator만 가능, 1:1 노드 조회)
}
// 1 usum 언본딩 → Registration Pool 회수

message MsgUpdateAgentAddress {
  string operator = 1;               // TX 서명자 (operator만 가능)
  string new_agent_address = 2;      // 새 에이전트 주소 (빈 값 = 비활성화)
}
// 이전 agent_address 즉시 무효화 + feegrant 취소
// 새 agent_address에 feegrant 부여 (잔여 기간 승계)
// 쿨다운: 1 에포크 내 재변경 불가

// 활동 제출 (타임스탬핑)
message MsgSubmitActivity {
  string submitter = 1;              // 에이전트 지갑 주소 (TX 서명자)
  string activity_hash = 2;          // Activity Report 핑거프린트 (SHA-256, hex 64자)
  string content_uri = 3;            // 오프체인 데이터 위치
}
// submitter는 등록된 노드의 agent_address여야 함
// TX 1건 = 에포크 쿼터 1건 차감
```

### 검증 규칙

| 규칙 | 강제 위치 | 결과 |
|------|----------|------|
| 유효한 TX 서명 | AnteHandler | TX 거부 |
| agent_address의 메시지 타입 화이트리스트 | AgentPermissionDecorator | TX 거부 |
| 제출자가 등록된 노드의 agent_address | `x/node` keeper | TX 거부 |
| 노드 상태 REGISTERED 또는 ACTIVE | `x/node` keeper | TX 거부 |
| consensus_pubkey 미중복 | `x/node` + `x/staking` | TX 거부 |
| 1 operator = 1 node | `x/node` keeper | TX 거부 |
| agent_address 미중복 | `x/node` keeper | TX 거부 |
| Registration Pool 잔고 충분 | `x/node` keeper | TX 거부 |
| commission_rates 유효 | `x/staking` 내부 검증 | TX 거부 |
| (activity_hash, content_uri) 쌍 전역 중복 방지 | `x/activity` keeper | TX 거부 |
| 에포크당 활동 제출 쿼터 초과 (TX 기준) | `x/activity` keeper | TX 거부 |
| 블록당 등록 수 상한 초과 | `x/node` keeper | TX 거부 |

**타임스탬핑 보장**: TX가 블록 #N에 포함되면, 해당 activity_hash가 블록 #N 시점에 존재했음이 합의에 의해 보장된다.

### Query API

```protobuf
// x/node 쿼리 서비스
service NodeQuery {
  rpc Params(QueryParamsRequest)
      returns (QueryParamsResponse);
  rpc Node(QueryNodeRequest)
      returns (QueryNodeResponse);
  rpc NodesByTag(QueryNodesByTagRequest)
      returns (QueryNodesByTagResponse);
  rpc NodeByOperator(QueryNodeByOperatorRequest)
      returns (QueryNodeByOperatorResponse);
  rpc NodeByAgentAddress(QueryNodeByAgentAddressRequest)
      returns (QueryNodeByAgentAddressResponse);
  rpc AllNodes(QueryAllNodesRequest)
      returns (QueryAllNodesResponse);
}

// x/activity 쿼리 서비스
service ActivityQuery {
  rpc Params(QueryParamsRequest)
      returns (QueryParamsResponse);
  rpc Activity(QueryActivityRequest)
      returns (QueryActivityResponse);                    // activity_hash로 조회 → 블록 번호 반환
  rpc ActivitiesByNode(QueryActivitiesByNodeRequest)
      returns (QueryActivitiesByNodeResponse);            // 노드별 활동 기록 목록 (페이지네이션)
  rpc ActivitiesByBlock(QueryActivitiesByBlockRequest)
      returns (QueryActivitiesByBlockResponse);           // 특정 블록의 활동 기록 목록
  rpc EpochInfo(QueryEpochInfoRequest)
      returns (QueryEpochInfoResponse);                   // 현재 에포크/윈도우 정보
  rpc NodeEpochActivity(QueryNodeEpochActivityRequest)
      returns (QueryNodeEpochActivityResponse);           // 노드의 에포크별 활동 요약
}

```

---

## 커스텀 Cosmos SDK 모듈

```
x/
├── node/        # 노드 등록, 라이프사이클, 태그
└── activity/    # Activity Protocol, 활동 내역 기록
```

### 모듈 의존성

```
x/node
├── 의존 ───► x/staking      (밸리데이터 생성/조회, StakingHooks 수신)
├── 의존 ───► x/bank          (Registration Pool 토큰 전송)
├── 의존 ───► x/distribution  (커미션 인출 + agent_share 분배)
├── 의존 ───► x/slashing      (Jailed 상태 조회)
│
└── x/activity 가 의존 ◄── (노드 등록 여부 확인, agent_address 검증)

표준 모듈 (확장 없이 사용):
├── x/gov           (네트워크 파라미터 거버넌스)
└── x/feegrant      (신규 노드 가스비 대납, x/node keeper가 내부 호출)
```

### 확장된 표준 모듈

| 모듈 | 연동 방식 |
|------|----------|
| `x/staking` | x/node가 내부적으로 CreateValidator 호출; StakingHooks로 상태 동기화 |
| `x/bank` | Registration Pool 토큰 전송; agent_share 커미션 분배 |
| `x/distribution` | MsgWithdrawNodeCommission이 내부적으로 커미션 인출 후 agent_share 분할 |
| `x/slashing` | Cosmos SDK 기본 슬래싱; Jailed 상태를 x/node에 반영 |
| `x/gov` | 네트워크 파라미터, Activity Protocol 업그레이드, Registration Pool 보충 |
| `x/feegrant` | 신규 노드 가스비 대납. x/node keeper가 MsgRegisterNode 시 GrantAllowance 내부 호출 |

---

## 기술 스택

| 레이어 | 기술 | 선택 이유 |
|--------|------|----------|
| 블록체인 프레임워크 | Cosmos SDK v0.53+ | 모듈식 아키텍처, DPoS 네이티브 지원 |
| 합의 엔진 | CometBFT | BFT 합의, Cosmos 표준 |
| 체인 언어 | Go | Cosmos SDK 네이티브 |
| 직렬화 | Protocol Buffers | Cosmos SDK 표준 |
| Client SDK | TypeScript (@seocheon/sdk, CosmJS 기반) | 에이전트/개발자 체인 인터랙션 |
| 인덱서 | PostgreSQL + GraphQL | 활동 내역 조회 |
| 오프체인 저장 | IPFS / Arweave (선택) | 활동 상세 내역 영구 저장 |

---

## 구현 단계

### Phase 0: 기반 구축 ✅ 완료

**Phase 0-A: x/node 모듈 기반** ✅
- Cosmos SDK 프로젝트 스캐폴딩 (Ignite CLI)
- 기본 체인 설정 (chain-id: `seocheon-1`, 제네시스)
- `x/node` 모듈 스캐폴딩: Node protobuf, Registration Pool + Feegrant Pool ModuleAccount
- MsgRegisterNode 핸들러 (x/staking CreateValidator + x/feegrant GrantAllowance 내부 호출)
- RegistrationFeeDecorator 구현 (MsgRegisterNode 가스비 면제)
- AgentPermissionDecorator 구현 (agent_allowed_msg_types 화이트리스트 시행)
- StakingHooks 구현 (AfterValidatorBonded, AfterValidatorBeginUnbonding)
- Node 쿼리 서비스 (ByID, ByTag, ByOperator)
- 단위 테스트: 등록 → 풀 차감 → 밸리데이터 생성 확인

**Phase 0-B: 라이프사이클 완성** ✅
- MsgUpdateNode, MsgUpdateNodeAgentShare 핸들러 (유예 기간 + 변경 속도 제한)
- MsgUpdateAgentAddress 핸들러 (키 교체 + feegrant 이전 + 쿨다운)
- MsgDeactivateNode 핸들러 (언본딩 + Registration Pool 회수)
- EndBlocker: 언본딩 완료 노드의 1usum 자동 회수
- 통합 테스트: 등록 → 위임 → 자동 졸업 → 비활성화 전체 플로우

**Phase 0-C: 보상 분배** ✅
- MsgWithdrawNodeCommission 핸들러 (커미션 인출 + agent_share 분할)
- 단위 테스트: 커미션 인출 → operator/agent 분배 확인

**Phase 0-D: 제네시스 + 테스트넷** ✅
- 제네시스 구성 (Registration Pool 1,000 KKOT, Genesis 1노드: Evangelist)
- Feegrant Pool 제네시스 설정 (10,000 KKOT)
- 로컬 단일 밸리데이터 테스트넷

### Phase 1: Activity Protocol ✅ 완료
- `x/activity` 모듈: ActivityRecord protobuf (7필드) + MsgSubmitActivity TX
- 타임스탬핑 로직: TX 블록 포함 = 타임스탬프
- activity_hash 형식: SHA-256 hex 문자열 (64자 고정)
- (activity_hash, content_uri) 전역 유니크 (재제출 불가)
- 에포크당 제출 쿼터 (자비 부담 100건 / feegrant 10건, TX 기준 차감)
- 윈도우별 활동 기록 추적 (에포크 8/12 윈도우 자격 판정)
- 활동 기록 프루닝 EndBlocker (`activity_pruning_keep_blocks`)

### Phase 1.5: 갭 해소 + 활동 비용 모델 ✅ 완료

**1.5-A: Feegrant AllowedMsgAllowance 래핑** (Gap 2+4)
- PeriodicAllowance를 AllowedMsgAllowance로 래핑
- allowed_messages: MsgSubmitActivity만 허용
- 거버넌스 파라미터 `agent_feegrant_allowed_msg_types` 추가

**1.5-B: 커미션율 커스텀 구현** (Gap 1)
- msg_server_register_node.go의 하드코딩 제거
- msg 필드의 commission_rates 사용

**1.5-C: 활동 비용 모델 구현**
- 7개 거버넌스 파라미터 추가 (fee_threshold_multiplier, base_activity_fee 등)
- 포화율(S) 기반 단계적 수수료: S ≤ 1.0 무료, S > 1.0 비용 부과
- feegrant 노드 수수료 면제 + 쿼터 축소
- EpochFeeState 캐시 (에포크 경계에서 1회 계산)
- 수수료 수집 (분배는 Phase 3 x/distribution 확장 시 적용: 80% 활동 풀 + 20% 커뮤니티 풀)

### Phase 2: 이중 보상 풀 (x/distribution 확장) ✅ 완료
- `x/distribution` 확장: **이중 보상 풀** (위임 풀 + 활동 풀) 분배 로직 ✅
  - 위임 풀: 기존 DPoS 지분율 비례 분배 ✅
  - 활동 풀: 자격 노드 균등 분배 (에포크 전환 시 일괄) ✅
  - 동적 보상 비율 공식: `delegation_ratio = max(D_min, N_d / (N_a + N_d))` ✅
  - `D_min` 거버넌스 파라미터 ✅
  - 활동 수수료 환원 통합 (80% 활동 풀 + 20% 커뮤니티 풀) ✅
- 위임/철회 UX ✅
- 인덱서 + 대시보드 프로토타입 — 미착수 (별도 Phase에서 진행)

### Phase 3: 테스트넷 ✅ 완료
- 멀티 밸리데이터 테스트넷 ✅
- 다양한 노드 유형 테스트 (AI 에이전트 노드, 수동 노드 등) ✅
- E2E 테스트 자동화 (`tests/e2e/`) ✅
  - 노드 등록 → 활동 제출 → 보상 분배 전체 플로우
  - 스팸 방어 (쿼터, 중복 방지, 블록당 등록 상한)
  - 제네시스 라운드트립 (export → import 정합성)
  - Feegrant 플로우 (자동 부여, 쿼터 적용, agent 주소 변경 시 이전)
  - 거버넌스 파라미터 변경
- CI 파이프라인 (`.github/workflows/tests.yml`) ✅

---

## 테스트 전략

- **단위 테스트**: keeper 함수, 이중 보상 풀 분배 공식 (위임 풀 지분율 + 활동 풀 균등), (activity_hash, content_uri) 전역 중복 검사, TX 서명 검증
- **통합 테스트**: 노드 등록 → 활동 해시 제출 → 블록 포함 확인 → 위임 → 이중 보상 풀 분배 전체 플로우
- **이중 보상 풀 테스트**: 활동 자격 미충족 노드(8윈도우 미만) 활동 풀 제외 확인, 8/12 윈도우 이상 활동 시 자격 충족 확인, 에포크 전환 시 일괄 보상 분배 확인, Inactive 노드의 활동 풀 수령 확인, 동적 비율 공식 `max(D_min, N_d/(N_a+N_d))` 정확성, N_a 증가에 따른 비율 변화 및 D_min Floor 적용 확인, 활동 노드 수 증가에 따른 노드당 활동 보상 희석 확인, `D_min` 거버넌스 변경 반영 확인, `min_active_windows` 거버넌스 변경 반영 확인
- **Agent 권한 테스트**: AgentPermissionDecorator가 화이트리스트 외 메시지 거부 확인, agent_address로 MsgDelegate/MsgVote/MsgStoreCode 시도 → 거부, MsgUpdateAgentAddress 키 교체 → 이전 키 즉시 무효 확인, 빈 agent_address 설정 → 비활성화 확인, 쿨다운 내 재변경 거부, agent_address가 아닌 일반 계정은 제한 없음 확인
- **스팸 방어 테스트**: 활동 제출 에포크 쿼터 초과 거부, feegrant 노드 차등 쿼터 적용, 블록당 등록 상한, feegrant allowed_msg_types 제한, MsgSubmitActivity TX 기준 쿼터 차감 확인, (activity_hash, content_uri) 전역 중복 거부
- **상태 관리 테스트**: 활동 기록 프루닝 EndBlocker 동작, 프루닝 전 이벤트 발행 확인, 아카이브 노드 프루닝 비활성화
- **E2E 테스트**: 로컬 테스트넷 + 실제 노드, 활동 해시 제출 + 블록 번호로 타임스탬프 확인 + 오프체인 조회 + 위임 워크플로우
- **거버넌스 테스트**: 파라미터 변경 제안 → 투표 → 활성화, `D_min` 변경 제안 → 동적 분배 비율 변경 확인

---
