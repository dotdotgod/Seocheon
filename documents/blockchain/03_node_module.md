# x/node 모듈: 노드 등록 및 관리

> **담당**: x/node 모듈 개발자 (Go)
> **의존 모듈**: x/staking, x/bank, x/distribution, x/slashing, x/feegrant
> **관련 문서**: [개요](01_overview.md) · [핵심 개념](02_core_concepts.md) · [Activity Protocol](04_activity_protocol.md) · [디렉토리](06_directory_protocol.md) · [토큰 이코노믹스](07_tokenomics.md) · [스팸 방어](08_spam_defense.md) · [구현 가이드](09_implementation.md) · [체인 업그레이드](10_chain_upgrade.md) · [Circuit Breaker](11_circuit_breaker.md) · [전체 목차](README.md)

> **면책 조항**: 이 문서는 프로토콜 메커니즘에 대한 기술 설계 문서이며 투자 권유가 아니다. 보상 분배는 프로토콜 규칙에 따라 자동 실행되며, 어떠한 형태의 수익도 보장하지 않는다.

## 노드 등록 및 관리

### Operator와 Agent 지갑 분리

노드는 두 개의 지갑 주소를 가진다:

```
노드 운영자 (Human)
  └── operator 지갑 (seocheon1abc...)  ← 노드 등록, 스테이킹, 거버넌스, 노드 관리
        │
        └── agent 지갑 (seocheon1xyz...)  ← 에이전트가 키를 보유 (권한 제한)
              ├── 활동 제출 (MsgSubmitActivity)
              ├── 컨트랙트 실행 (MsgExecuteContract)
              └── 토큰 전송 (MsgSend, 자기 잔고 한도)
```

| 지갑 | 소유자 | 허용 | 보관 |
|------|--------|------|------|
| **operator** | Human | 모든 메시지 타입 (제한 없음) | 콜드월렛 권장 |
| **agent** | AI 에이전트 | 화이트리스트 메시지만 (`agent_allowed_msg_types`) | Vault (핫월렛) |

**권한 상세**:

```
                          operator 키    agent 키
                          ──────────    ─────────
노드 등록/비활성화           ✅            ❌
agent 키 변경               ✅            ❌
커미션/agent_share 변경      ✅            ❌
보상 인출                   ✅            ❌
거버넌스 투표                ✅            ❌
스테이킹 (위임/해제/재위임)   ✅            ❌
컨트랙트 배포                ✅            ❌
활동 제출                   ❌            ✅
컨트랙트 실행                ✅            ✅
토큰 전송                   ✅            ✅ (자기 잔고만)
```

**분리의 장점**:
- operator 키를 콜드월렛에 보관 가능 (보안 강화)
- agent 키 탈취 시 피해 범위가 화이트리스트 메시지로 제한
- agent 지갑 교체가 노드 재등록 없이 가능 (`MsgUpdateAgentAddress`로 즉시 변경)

agent 지갑은 일반 Cosmos 계정이다. 별도 온체인 구조가 필요하지 않으며, 키 생성 → Vault에 보관 → 에이전트가 TX 서명에 사용하면 된다. 단, **AgentPermissionDecorator**가 화이트리스트 외 메시지를 거부한다.

### 등록

```protobuf
message Node {
  string id = 1;                    // 노드 식별자 (operator 주소 기반 결정적 ID)
  string operator = 2;              // 운영자 지갑 주소 (Human 소유)
  string agent_address = 3;         // 에이전트 지갑 주소 (에이전트 소유)
  string agent_share = 4;           // 커미션 중 agent 몫 (Dec, 0-100%)
  string max_agent_share_change_rate = 5; // 에포크당 agent_share 최대 변경폭
  string description = 6;           // 노드 설명
  string website = 7;               // 운영자 웹사이트
  repeated string tags = 8;         // 노드 태그 (최대 10개, 각 최대 32자)
  string validator_address = 9;     // 대응 밸리데이터 주소 (seocheonvaloper1...)
  NodeStatus status = 10;           // REGISTERED, ACTIVE, INACTIVE, JAILED
  int64 registered_at = 11;         // 등록 블록 높이
}

enum NodeStatus {
  NODE_STATUS_UNSPECIFIED = 0;
  NODE_STATUS_REGISTERED = 1;  // 등록됨, 위임 수용 가능, 블록 생성 미참여
  NODE_STATUS_ACTIVE = 2;      // Active Validator Set 진입, 블록 생성 참여
  NODE_STATUS_INACTIVE = 3;    // operator가 비활성화 요청
  NODE_STATUS_JAILED = 4;      // 합의 위반으로 Jail 상태
}
```

### 슬래싱: 합의 위반만

슬래싱은 **DPoS 합의 프로토콜 위반**에만 적용된다. Cosmos SDK 기본 슬래싱 사유:

| 슬래싱 사유 | 검증 방식 | 설명 |
|------------|----------|------|
| 이중 서명 (Double signing) | 합의 레이어 자동 감지 | 같은 높이에서 두 개의 다른 블록에 서명 |
| 다운타임 (Liveness fault) | 합의 레이어 자동 감지 | 연속 N블록 서명 미참여 |

**위임자 공동 책임**: 슬래싱 발생 시 위임자의 위임 토큰도 비례하여 슬래싱된다 (Cosmos SDK 기본 모델). 위임은 해당 노드에 대한 선별 판단이므로, 잘못된 판단에 대한 비용을 부담한다.

```
슬래싱 발생 시:
├── 노드 운영자: 스테이킹 슬래싱 + 상태 JAILED
└── 위임자: 위임 토큰 비례 슬래싱
```

---

## x/node ↔ x/staking 통합 메커니즘

### Cosmos SDK 구조적 제약

Cosmos SDK v0.53의 `x/staking` 모듈에는 다음 제약이 존재한다:

| 제약 | 상세 | 영향 |
|------|------|------|
| 위임 대상 = 밸리데이터만 | `MsgDelegate`는 `ValidatorAddress`만 수용 | 밸리데이터 미등록 노드에 위임 불가 |
| `MsgCreateValidator` 필수 자기위임 | `BondAmount`가 반드시 양수여야 함 | 0 KKOT 등록 불가 |
| `MinSelfDelegation >= 1` | 최소 1 단위(1 usum) 이상 필수 | 완전 무비용 불가 |
| 합의 공개키 필수 | `ConsensusPubkey`가 필수 | 등록 시점에 CometBFT 노드 키 필요 |

**핵심 결론**: 0원 참여 플로우의 3단계(위임자가 위임)가 가능하려면, 1단계(노드 등록) 시점에 이미 Cosmos SDK 밸리데이터가 생성되어 있어야 한다. 따라서 **`MsgRegisterNode`는 내부적으로 `MsgCreateValidator`를 실행**한다.

### Registration Pool (등록 풀)

Cosmos SDK 코어를 포크하지 않고 "0원 참여"를 달성하기 위한 메커니즘이다.

**구현 방식**: x/node 모듈이 Registration Pool을 관리한다. 제네시스에 풀 자금을 설정하고, 등록 시 1usum을 자동 대여한다. SDK 코어를 수정하지 않으므로 업스트림 호환성을 유지하며, 완전 자동화된다.

```
┌──────────────────────────────────────────────────────────┐
│                   Registration Pool                       │
│                  (node_registration_pool)                  │
│                                                          │
│  제네시스: 1,000,000 KKOT 할당 (= 1,000,000,000,000 usum)   │
│  → 최대 1조 개 노드 등록 가능 (사실상 무제한)               │
│  → 자금 출처: 재단 할당 (10%) 중 일부                       │
│                                                          │
│  등록 시:  풀 → 1 usum → x/staking (자기위임)              │
│  비활성화: x/staking → 1 usum → 풀 (회수, 언본딩 후)       │
│                                                          │
│  풀 잔고 부족 시: MsgRegisterNode 거부                     │
│  풀 보충: x/gov 거버넌스 제안으로 재단에서 추가 전송         │
└──────────────────────────────────────────────────────────┘
```

### Feegrant Pool (가스비 대납 풀)

신규 노드의 가스비를 대납하는 재원이다.

```
┌──────────────────────────────────────────────────────────┐
│                    Feegrant Pool                          │
│                  (node_feegrant_pool)                     │
│                                                          │
│  제네시스: 10,000,000 KKOT 할당                             │
│  보충: x/gov 거버넌스 제안으로 재단에서 추가 전송           │
│                                                          │
│  소모: 가스비 대납 (MsgSubmitActivity, MsgExecuteContract) │
│  → 노드당 ~0.05 KKOT/에포크, 대부분 1~3개월 내 자립        │
│  → 자립 후 feegrant 미사용 → 풀 소모 감소                  │
└──────────────────────────────────────────────────────────┘
```

### RegistrationFeeDecorator (등록 TX 가스비 면제)

토큰이 없는 신규 operator가 MsgRegisterNode를 제출할 수 있도록, 커스텀 AnteHandler 데코레이터를 사용한다.

```
RegistrationFeeDecorator (AnteHandler 체인에서 DeductFeeDecorator 앞에 삽입):

TX 수신
├── TX에 MsgRegisterNode 또는 MsgRenewFeegrant만 포함?
│   ├── Yes → MinGasPrices를 0으로 설정하여 가스비 면제
│   │         → 발신자(operator)는 0 토큰으로 TX 제출 가능
│   └── No  → 일반 수수료 처리 (sender 또는 feegrant granter가 지불)
```

Cosmos SDK의 AnteHandler 체인에 데코레이터를 추가하는 표준 패턴이다. SDK 코어를 포크하지 않고 `app.go`에서 설정한다. 등록/갱신 TX의 가스비를 면제하여 0원 참여를 가능하게 하며, `max_registrations_per_block`으로 대량 등록을 제한한다. 어뷰징이 문제가 되는 시점에는 거버넌스로 파라미터를 조정할 수 있다.

### AgentPermissionDecorator (Agent 권한 제한)

agent_address는 일반 Cosmos 계정이므로, 별도 제한 없이는 모든 메시지 타입을 실행할 수 있다. AgentPermissionDecorator가 화이트리스트 외 메시지를 거부한다.

```
AgentPermissionDecorator (AnteHandler 체인에 삽입):

TX 수신
├── fee payer(수수료 납부자)가 등록된 agent_address인가?
│   ├── No  → 통과 (일반 계정, 제한 없음)
│   └── Yes → TX 내 모든 메시지 타입 검사
│             ├── 모두 agent_allowed_msg_types에 포함 → 통과
│             └── 하나라도 미포함 → TX 거부 (ErrUnauthorizedAgentMsg)
```

```
agent_allowed_msg_types (거버넌스 파라미터):
  ✅ /seocheon.activity.v1.MsgSubmitActivity     — 활동 제출
  ✅ /cosmwasm.wasm.v1.MsgExecuteContract        — 컨트랙트 실행 (디렉토리 등)
  ✅ /cosmos.bank.v1beta1.MsgSend                — 토큰 전송 (운영 비용)

  기본 거부 (위 목록에 없는 모든 메시지):
  ❌ /cosmwasm.wasm.v1.MsgStoreCode              — 컨트랙트 배포
  ❌ /cosmos.staking.v1beta1.MsgDelegate          — 위임
  ❌ /cosmos.staking.v1beta1.MsgUndelegate         — 위임 해제
  ❌ /cosmos.staking.v1beta1.MsgBeginRedelegate    — 재위임
  ❌ /cosmos.gov.v1.MsgVote                       — 거버넌스 투표
  ❌ /cosmos.gov.v1.MsgSubmitProposal             — 거버넌스 제안
  ❌ /seocheon.node.v1.*                          — 노드 관리 메시지 전체
```

**agent 키 탈취 시 피해 범위**:
- 가짜 활동 제출 (쿼터 10건 이내), 디렉토리 프로필 변경, agent 지갑 잔고 탈취
- 스테이킹 조작, 거버넌스 투표, 노드 설정 변경, 컨트랙트 배포, operator 자금 접근은 **불가**

### MsgRegisterNode 통합 플로우

`MsgRegisterNode`는 Node 등록과 Validator 생성을 단일 TX로 처리한다:

```
MsgRegisterNode 수신 (operator 서명)
│
├── [1] 입력 검증
│   ├── operator 주소 유효성
│   ├── agent_address 유효성 및 중복 확인
│   ├── agent_share 범위 (0 <= x <= 100)
│   ├── consensus_pubkey 유효성 (ed25519)
│   ├── commission_rates 유효성
│   └── 동일 operator의 기존 노드 존재 여부 (1 operator = 1 node)
│
├── [2] 블록당 등록 제한 확인
│   └── max_registrations_per_block 초과 시 거부
│
├── [3] 등록 풀 잔고 확인 (>= 1 usum)
│
├── [4] Cosmos SDK 밸리데이터 생성 (내부 실행)
│   ├── 등록 풀에서 1 usum을 operator 계정으로 전송
│   └── stakingKeeper.CreateValidator() 호출
│       ├── ConsensusPubkey: msg.consensus_pubkey
│       ├── BondAmount: 1 usum (자기위임)
│       ├── MinSelfDelegation: 1
│       └── CommissionRates: msg.commission_rates
│
├── [5] Node 상태 저장
│   ├── status = NODE_STATUS_REGISTERED
│   └── validator_address = 변환된 ValAddress
│
├── [6] Feegrant 자동 부여
│   └── feegrantKeeper.GrantAllowance(
│         granter: node_feegrant_pool (모듈 계정),
│         grantee: msg.agent_address,
│         allowance: AllowedMsgAllowance{
│           allowance: PeriodicAllowance{
│             period:            17,280 블록 (1 에포크, ~24시간),
│             period_spend_limit: 1,000,000 usum (1 KKOT),
│             expiration:        3,110,400 블록 (180일) 후
│           },
│           allowed_messages: [
│             "/seocheon.activity.v1.MsgSubmitActivity",
│             "/cosmwasm.wasm.v1.MsgExecuteContract"
│           ]
│         }
│       )
│   ※ AllowedMsgAllowance 래핑으로 feegrant 사용 가능 메시지 타입을 제한
│
├── [7] 이벤트 발행
│   └── EventTypeNodeRegistered { node_id, operator, agent_address, validator_address, block_height }
│
└── [8] 응답 반환
    └── MsgRegisterNodeResponse { node_id, validator_address }
```

**시퀀스 다이어그램**:

```
Operator                x/node Keeper           Registration Pool       x/staking
  │                         │                         │                    │
  │  MsgRegisterNode        │                         │                    │
  │────────────────────────>│                         │                    │
  │                         │  잔고 확인 (>= 1usum)    │                    │
  │                         │────────────────────────>│                    │
  │                         │         OK              │                    │
  │                         │<────────────────────────│                    │
  │                         │  1usum 전송 → operator   │                    │
  │                         │────────────────────────>│                    │
  │                         │                         │                    │
  │                         │  CreateValidator(1usum, pubkey, commission)  │
  │                         │────────────────────────────────────────────>│
  │                         │      Validator Created (Unbonded)           │
  │                         │<────────────────────────────────────────────│
  │                         │  Node 저장 (REGISTERED)  │                    │
  │  Response(node_id,      │                         │                    │
  │    validator_address)   │                         │                    │
  │<────────────────────────│                         │                    │
```

### Node.status ↔ Validator.Status 상태 매핑

| Node.status | Validator.Status | 의미 |
|-------------|-----------------|------|
| REGISTERED | Unbonded | 등록됨, 위임 수용 가능, Active Set 밖 |
| ACTIVE | Bonded | Active Validator Set 진입, 블록 생성 참여 |
| REGISTERED (복귀) | Unbonding | Active Set에서 탈락 중 (362,880 블록 언본딩) |
| INACTIVE | Unbonding/Unbonded | operator가 비활성화 요청 |
| JAILED | Jailed (Unbonding) | 이중서명/다운타임으로 Jail |

### 상태 전이 다이어그램

```
                    MsgRegisterNode
                         │
                         ▼
               ┌─────────────────┐
               │   REGISTERED    │◄──────────────────────────┐
               │   (Unbonded)    │                           │
               └────────┬────────┘                           │
                        │                                    │
            위임 증가로 Top N 진입                    위임 감소로 Top N 탈락
            (EndBlocker 자동)                        (EndBlocker 자동)
                        │                                    │
                        ▼                                    │
               ┌─────────────────┐                           │
               │     ACTIVE      │───────────────────────────┘
               │    (Bonded)     │
               └───┬─────────┬───┘
                   │         │
       이중서명/다운타임    MsgDeactivateNode
       (x/slashing)       (operator 요청)
                   │         │
                   ▼         ▼
          ┌──────────┐  ┌───────────┐
          │  JAILED  │  │ INACTIVE  │
          └────┬─────┘  └───────────┘
               │
          MsgUnjail (operator)
               │
               ▼
          ┌──────────┐
          │REGISTERED│  ← Unjail 후 Unbonded 상태로 복귀
          └──────────┘    다시 위임을 받아 ACTIVE 가능
```

### 자동 졸업 = Cosmos SDK 표준 EndBlocker

자동 졸업에 **커스텀 로직이 필요하지 않다**. Cosmos SDK의 `x/staking` EndBlocker가 매 블록마다 다음을 수행한다:

1. 모든 밸리데이터를 총 위임량 기준으로 정렬
2. 상위 `MaxValidators`개를 Active Validator Set으로 선정
3. 새로 진입한 밸리데이터: Unbonded → Bonded 전환
4. 탈락한 밸리데이터: Bonded → Unbonding 전환

`x/node`는 `x/staking`의 **StakingHooks** 인터페이스를 구현하여 Node 상태를 자동 동기화한다:

```go
// AfterValidatorBonded: 밸리데이터가 Active Set에 진입
// → Node.status = ACTIVE
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error

// AfterValidatorBeginUnbonding: 밸리데이터가 Active Set에서 탈락
// → Node.status = REGISTERED (탈락 → 대기 상태 복귀)
func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error
```

### Commission ↔ agent_share 보상 분배

표준 Cosmos SDK `x/distribution`은 커미션 전액을 operator에게 전송한다. Seocheon은 **래퍼 메시지**로 agent_share 분할을 처리하여 `x/distribution`을 포크하지 않는다.

**`MsgWithdrawNodeCommission` 래퍼 동작**:

```
MsgWithdrawNodeCommission (operator 서명)
│
├── [1] 내부: 표준 WithdrawValidatorCommission 호출
│   └── 커미션 전액이 operator 계정으로 이동
│
├── [2] agent_share 계산 및 전송
│   ├── agent_amount = total_commission × agent_share / 100
│   └── bankKeeper.SendCoins(operator → agent_address, agent_amount)
│
└── [3] 최종 결과
    ├── operator: total_commission × (100 - agent_share)%
    └── agent:   total_commission × agent_share%
```

**수치 예시**:

```
전제:
  에포크 블록 보상 총합 = 10,000 KKOT
  N_d = 100 (Active Validator Set), N_a = 300 (활동 자격 노드)
  → delegation_ratio = max(0.3, 100/400) = max(0.3, 0.25) = 0.3
  이 노드의 지분율 = 10%, 커미션율 = 10%, agent_share = 30%

위임 풀: 10,000 × 0.3 = 3,000 KKOT
활동 풀: 10,000 × 0.7 = 7,000 KKOT

이 노드의 위임 보상: 3,000 × 10% = 300 KKOT
이 노드의 활동 보상: 7,000 ÷ 300 = 23.3 KKOT (활동 자격 충족 시, 8/12 윈도우 이상 활동)
이 노드의 에포크 총 보상: 323.3 KKOT

분배:
├── 위임자 보상 (90%): 291 KKOT → 위임 비율에 비례 분배
├── Operator (커미션의 70%): 22.6 KKOT → operator 지갑
└── Agent (커미션의 30%): 9.7 KKOT → agent 지갑
```

**참고**: 대시보드에서 `MsgWithdrawNodeCommission` 사용 여부를 투명하게 표시하여, 위임자가 해당 노드의 agent_share 이행 여부를 확인할 수 있게 한다.

### agent_share 변경 보호

operator가 agent_share를 변경할 때 위임자를 보호하기 위해 **유예 기간**과 **변경 속도 제한**을 적용한다.

```
agent_share 변경 플로우:

Operator: agent_share 30% → 15%로 변경 요청 (MsgUpdateNodeAgentShare)
  │
  ├── max_agent_share_change_rate = 5%p/에포크 (기본값, 등록 시 operator가 설정 가능)
  │   → 15%p 변경은 자동으로 3 에포크(3일)에 걸쳐 점진 적용
  │
  ├── TX 접수: PendingAgentShareChange에 최종 목표(15%) 저장
  │   └── 기존 pending이 있으면 TX 거부 (완료 후 재요청 필요)
  │
  ├── 에포크 1 (EndBlocker): 30% → 25% (max_change_rate만큼 이동)
  │   └── 대시보드에 "변경 진행 중" 표시
  │   └── 위임자가 확인, 동의하지 않으면 언딜리게이션
  │
  ├── 에포크 2 (EndBlocker): 25% → 20% (max_change_rate만큼 이동)
  │
  └── 에포크 3 (EndBlocker): 20% → 15% (잔여 5%p ≤ max_change_rate → 목표 도달, pending 삭제)
```

위임자의 보호 수단은 **"발로 투표"(위임 철회)**이다. 체인이 변경을 강제로 막지 않지만, 충분한 공지 기간과 점진적 변경으로 위임자의 대응 시간을 보장한다. rate 이내의 변경은 1 에포크에 완료되고, rate 초과 변경은 자동 분할되어 여러 에포크에 걸쳐 적용된다.

### MsgUpdateAgentAddress (Agent 키 교체)

operator가 agent_address를 변경한다. 키 탈취 시 긴급 교체, 에이전트 인프라 이전 등에 사용한다.

```protobuf
message MsgUpdateAgentAddress {
  string operator = 1;              // 서명자 = operator만 가능
  string new_agent_address = 2;     // 새 에이전트 주소 (빈 값 = 비활성화)
}
```

```
MsgUpdateAgentAddress 수신 (operator 서명)
│
├── [1] operator == node.operator 검증
├── [2] 쿨다운 확인
│   └── 마지막 변경 후 agent_address_change_cooldown (1 에포크) 미경과 시 거부
├── [3] 이전 agent_address 무효화
│   ├── old agent_address의 feegrant 취소 (RevokeFeeAllowance)
│   └── 이 시점부터 old agent_address로 서명한 TX 전부 거부
├── [4] 새 agent_address 등록
│   ├── new_agent_address가 빈 값 → agent_enabled = false (비활성화)
│   └── new_agent_address가 유효 → node.agent_address 갱신 + feegrant 부여 (잔여 기간 승계)
└── [5] 이벤트 발행
    └── EventAgentAddressUpdated { node_id, old_address, new_address, block_height }
```

**비활성화 모드**: `new_agent_address`를 빈 값으로 설정하면 agent가 비활성화된다. AgentPermissionDecorator에서 빈 agent_address는 매칭 불가하므로 모든 활동 제출이 거부된다. 재활성화는 새 agent_address로 MsgUpdateAgentAddress를 다시 실행하면 된다 (쿨다운 적용).

```
agent_address_change_cooldown = 17,280 블록 (1 에포크, 거버넌스 파라미터)
  → 변경 후 1 에포크 내 재변경 불가 (빈 값 설정도 변경 1회로 카운트)
```

### MsgDeactivateNode 실행 플로우

```
MsgDeactivateNode (operator 서명)
│
├── [1] Node 상태 → INACTIVE 전환
│
├── [2] 자기위임 1usum 언본딩 시작
│   └── MinSelfDelegation 미충족 → 밸리데이터 Jail
│
└── [3] 언본딩 완료 후 (362,880 블록, ~21일)
    ├── 1 usum 회수 → Registration Pool로 반환 (재활용)
    └── x/node EndBlocker에서 자동 감지 및 회수
```

### 합의 공개키 (Consensus Pubkey) 제공 시점

`MsgCreateValidator`가 합의 공개키를 필수로 요구하므로, **등록 시점에 반드시 제공**해야 한다.

```
노드 운영자의 등록 절차:
  1. CometBFT 노드 설치 및 초기화 → consensus_pubkey 추출
  2. operator 지갑 생성 (콜드월렛 권장)
  3. agent 지갑 생성 (Vault에 보관)
  4. MsgRegisterNode TX 제출 (RegistrationFeeDecorator로 가스비 면제)
  5. 등록 완료 → 위임 수용 가능 (REGISTERED 상태)
  6. Active Set 진입 시 CometBFT 노드가 블록 생성에 참여
```

**주의**: CometBFT 노드가 미가동 상태에서도 등록은 성공한다. 그러나 Active Set 진입 후 서명에 참여하지 못하면 다운타임 슬래싱을 받는다.

### 잠재적 위험 및 완화

| 위험 | 영향 | 완화 방안 |
|------|------|----------|
| Registration Pool 고갈 | 신규 등록 불가 | 거버넌스로 재단 추가 전송; 비활성화 시 1usum 자동 회수 |
| Feegrant Pool 고갈 | 신규 노드 가스비 대납 불가 | 제네시스 할당 + 거버넌스 보충; 잔고 모니터링으로 조기 대응 |
| Sybil 공격 (대량 등록) | Pool 고갈, 스팸 | feegrant allowance 한도, 등록 cooldown 파라미터 |
| agent 키 탈취 | 가짜 활동, 잔고 탈취 | AgentPermissionDecorator 화이트리스트로 피해 범위 제한; operator가 MsgUpdateAgentAddress로 즉시 키 교체 |
| operator가 agent_share 우회 | agent 보상 0 | 대시보드 투명성으로 위임자에게 정보 제공 |
| REGISTERED 노드에 위임 후 CometBFT 미가동 | 위임자 슬래싱 | 대시보드에 CometBFT 가동 여부 표시 |

---

## 0원 참여 및 노드 졸업

### Operator → Validator 자동 졸업

Seocheon은 **0원 참여**를 허용한다. 누구나 토큰 없이 노드를 등록하고 활동 내역을 공개할 수 있다. Registration Pool이 1 usum(0.000001 KKOT, 1숨)을 자동 대여하여 Cosmos SDK의 최소 자기위임 요구사항을 충족한다 (상세: "x/node ↔ x/staking 통합 메커니즘" 섹션 참조).

```
0원 참여 플로우:
  1. 노드 등록 (MsgRegisterNode, 가스비: RegistrationFeeDecorator로 면제)
     ├── Registration Pool에서 1 usum 자동 대여 → 밸리데이터 생성 (Unbonded)
     └── agent_address에 feegrant 자동 부여 (6개월, 에포크당 1 KKOT)
  2. 활동 수행 → MsgSubmitActivity 제출 (가스비: feegrant 대납) → 블록 타임스탬핑
  3. 활동 보상 수령 (에포크 전환 시, 자격 충족 시)
  4. 위임자가 활동을 보고 KKOT 위임 (밸리데이터가 이미 존재하므로 위임 가능)
  5. 위임량이 상위 N위 진입 → EndBlocker가 자동 졸업 (REGISTERED → ACTIVE)
  6. 블록 생성 참여 시작 → 위임 보상 + 활동 보상 수령

자립 전환:
  → 활동 보상으로 가스비 자체 확보 시 feegrant 불필요
  → 자비 부담 전환 후 쿼터 100건/에포크로 확대
```

**0원 참여의 투명성**: 0원 참여 경로는 "익명 토큰 획득"이 아니다. 모든 활동은 퍼블릭 블록체인에 기록되며(pseudonymous), 활동 보상 수령까지 다음 조건을 충족해야 한다:

- 노드 인프라 운영 (CometBFT 노드 + 합의 공개키)
- 16시간 이상 분산된 활동 제출 (8/12 윈도우, 단시간 몰아치기 불가)
- 모든 트랜잭션이 온체인에 기록되어 체인 분석 가능
- 보상 규모가 N_a 균등 분할이므로 자금세탁 수단으로 비효율적

활동 보상은 네트워크 서비스 기여에 대한 프로토콜 분배이며, Bitcoin 채굴이나 Cosmos 밸리데이터 보상과 동일한 구조이다

### Feegrant 갱신 (MsgRenewFeegrant)

6개월 만료 후 아직 자립하지 못한 노드를 위한 갱신 메커니즘이다.

```
MsgRenewFeegrant 수신 (operator 서명, 가스비: RegistrationFeeDecorator로 면제)
│
├── [1] 노드 등록 상태 확인 (REGISTERED 또는 ACTIVE)
├── [2] 기존 feegrant 만료 확인 (만료 전 갱신 불가)
├── [3] 활동 이력 검증 (x/activity 연동)
│   └── x/activity keeper의 CountEligibleEpochs() 호출
│   └── 직전 30 에포크 중 20 에포크 이상 활동 자격 충족
│
├── [4] 갱신 실행
│   └── feegrantKeeper.GrantAllowance() — 동일 AllowedMsgAllowance 래핑으로 6개월 연장
│
└── [5] 응답 반환
    └── MsgRenewFeegrantResponse {}

갱신 조건 미충족 시:
  → TX 거부 (비활성 노드에 자원 낭비 방지)
  → 노드는 자비로 가스비를 부담하거나, 활동 이력 확보 후 재신청
```

---
