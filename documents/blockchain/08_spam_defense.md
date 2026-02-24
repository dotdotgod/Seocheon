# 스팸/게이밍 방어

> **담당**: 보안 / 메커니즘 설계 담당
> **관련 모듈**: x/node, x/activity
> **관련 문서**: [핵심 개념](02_core_concepts.md) · [노드 모듈](03_node_module.md) · [Activity Protocol](04_activity_protocol.md) · [토큰 이코노믹스](07_tokenomics.md) · [구현 가이드](09_implementation.md) · [Circuit Breaker](11_circuit_breaker.md) · [전체 목차](README.md)

> **면책 조항**: 이 문서는 프로토콜 메커니즘에 대한 기술 설계 문서이며 투자 권유가 아니다. 보상 분배는 프로토콜 규칙에 따라 자동 실행되며, 어떠한 형태의 수익도 보장하지 않는다.

### 방어 원칙

Seocheon의 핵심 철학("체인이 측정할 수 없는 것은 합의에서 판단하지 않는다")을 유지하면서, **구조적·형식적 규칙**으로만 스팸을 방어한다. 콘텐츠 품질, 활동의 진위, 유용성 등은 여전히 생태계(위임)가 판단한다.

```
체인의 구조적 방어 범위:
├── 활동 대량 제출 (에포크 쿼터)
├── 노드 대량 등록 (블록당 등록 상한)
├── feegrant 남용 (allowance 정책)
└── 상태 무한 증가 (프루닝 전략 + 머클트리)
```

### A. 활동 제출 제한

에포크당 노드별 제출 상한을 두어 스팸 대량 제출을 방지한다.

```
활동 제출 쿼터:

자비 부담 노드:
  max_activity_submissions_per_epoch = 100    ← 거버넌스 파라미터

feegrant 노드:
  max_feegrant_submissions_per_epoch = 10     ← 거버넌스 파라미터

쿼터 초과 시: MsgSubmitActivity TX 거부
쿼터 리셋: 매 에포크 시작 시
쿼터 차감: MsgSubmitActivity 1건 = 쿼터 1건 차감
```

**feegrant 노드 판별**: `x/node` keeper가 제출자의 operator 주소에 대해 `x/feegrant`의 `Allowance` 쿼리를 실행하여 feegrant 수혜 여부를 확인한다.

**설계 근거**: feegrant 노드는 가스비가 무료이므로 스팸 제출의 경제적 비용이 0이다. 낮은 쿼터(10건)로 제한하되, 활동 보상 자격(8윈도우 × 1건 = 8건)을 충족할 수 있는 수준을 보장한다. 잔여 2건은 기타 TX에 사용할 수 있으나 여유가 적으므로 Sybil 노드의 활동 외 행동이 극히 제한된다. 자비 부담 노드는 가스비 자체가 스팸 억제 역할을 하므로 더 높은 쿼터를 부여한다.

**Sybil 방어 계층**:

```
Layer 1: 시간 분산 요구 (8/12 윈도우 = 최소 16시간 분산)
  → 각 Sybil 노드마다 2시간 간격 스케줄러 필요
  → 운영 복잡도와 인프라 비용 증가

Layer 2: feegrant 쿼터 제한 (10건/에포크)
  → 활동 자격 8건 + 잔여 2건만 가능
  → 활동 보상만 노리는 빈 껍데기 노드의 보상 효율 낮음

Layer 3: registration_deposit (거버넌스 활성화 시)
  → 초기값 0, Active ≥ 100 시점에 거버넌스로 활성화 권장
  → Sybil 노드 등록의 경제적 비용 부과
```

### B. 등록 스팸 방어

Registration Pool과 feegrant를 이용한 무비용 대량 등록을 방지한다.

```
등록 속도 제한:

max_registrations_per_block = 5              ← 거버넌스 파라미터
registration_cooldown_blocks = 100           ← 동일 operator의 재등록 대기 블록 수

선택적 등록 보증금:
  registration_deposit = 0 usum              ← 초기값 0 (거버넌스로 활성화 가능)
  → 활성화 시: 등록 시 보증금 예치, 비활성화 시 반환
  → Sybil 공격 비용을 높이는 옵션
```

**설계 근거**: 0원 참여 원칙을 유지하면서(`registration_deposit` 초기값 0), Sybil 공격이 실제 문제가 될 때 거버넌스로 보증금을 활성화할 수 있는 여지를 남긴다. `max_registrations_per_block`은 단일 블록에 등록 TX가 집중되는 것을 방지한다.

### C. Feegrant 정책 표준화

Foundation이 신규 노드에 설정하는 feegrant의 구체적 파라미터를 정의한다.

```
Foundation → 신규 노드 feegrant 설정:

x/feegrant AllowedMsgAllowance (래핑):
├── allowed_messages:
│   ├── /seocheon.activity.v1.MsgSubmitActivity
│   └── /cosmwasm.wasm.v1.MsgExecuteContract
│   └── (MsgSend, MsgStoreCode 등은 제외)
│
└── allowance: PeriodicAllowance
    ├── period: 24시간 (~1 에포크)
    ├── period_spend_limit: 1,000,000 usum (1 KKOT/에포크)
    ├── period_can_spend: 매 에포크 리셋
    └── expiration: 등록 후 ~180일 (6개월)
```

**AllowedMsgAllowance 래핑**: `PeriodicAllowance`를 `AllowedMsgAllowance`로 래핑하여, feegrant가 `MsgSubmitActivity`와 `MsgExecuteContract`에만 사용되도록 한다. 토큰 전송(`MsgSend`)이나 컨트랙트 배포(`MsgStoreCode`)에는 feegrant를 사용할 수 없다. `allowed_messages`는 거버넌스 파라미터(`agent_feegrant_allowed_msg_types`)로 관리한다.

**만료**: 약 180일(6개월) 후 자동 만료. 이 기간 내에 Activity Airdrop을 통해 자체 가스비를 확보하는 것이 기대 시나리오이다. 필요 시 MsgRenewFeegrant로 재신청 가능.

### C-1. 활동 비용 모델의 스팸 방어 효과

활동 비용 모델([Activity Protocol](04_activity_protocol.md) 참조)은 네트워크 성장기(S > 1.0)에 자동으로 활성화되어 추가적인 스팸 방어 계층을 제공한다.

```
활동 비용 모델의 스팸 방어 기여:

Layer 4: 경제적 비용 (자비 부담 노드)
  → S > 1.0일 때 activity_fee 자동 부과
  → 비용/보상 비율이 100% 초과 시 Sybil 노드 경제적 불가능
  → 수수료의 80%가 활동 풀 환원 → 정직한 노드에게 보상 증가

Layer 5: feegrant 쿼터 자동 축소 (feegrant 노드)
  → S > 1.0일 때 effective_quota 감소 (최소 8까지)
  → 과포화 시 feegrant 노드의 활동 여유분 0에 수렴
  → feegrant 노드는 수수료 면제이나 쿼터 축소로 압박
```

쿼터 제한(Layer 2)과 활동 비용 모델(Layer 4-5)이 결합되어, 네트워크 규모에 따라 방어 강도가 자동 조절된다.

### D. 상태 관리 전략

장기적인 온체인 상태 증가를 관리한다.

#### 활동 기록 프루닝

활동 기록(ActivityRecord)은 제출 빈도가 높아 가장 빠르게 누적된다. `x/activity` 모듈 레벨에서 프루닝한다.

```
활동 기록 상태 관리:

activity_pruning_keep_blocks = 1,555,200     ← 90일 보존 (17,280 × 90, 거버넌스 파라미터)

프루닝 동작 (EndBlocker):
├── 현재 블록 - activity_pruning_keep_blocks 이전의 ActivityRecord 삭제
├── 삭제 전 이벤트 발행 → 인덱서가 수집하여 영구 보존
└── 블록당 최대 프루닝 건수 제한 (성능 영향 방지)

풀노드: 최근 1,555,200 블록(~90일) 데이터만 보유
인덱서: 전체 이력 보유 (PostgreSQL + GraphQL로 조회)
아카이브 노드: 선택적 운영 (프루닝 비활성화)
```

#### 상태 크기 추정

```
시나리오: 100 노드, 1,555,200 블록(~90일) 보존

ActivityRecord:
  100 노드 × 에포크당 50건 × 90 에포크(90일) = 450,000건
  × 약 200 bytes/건 = ~90 MB

총 온체인 상태: ~90 MB (1,555,200 블록 ≈ 90일 기준, 프루닝 적용 시)
→ 풀노드 운영에 부담 없는 수준
```

### E. 거버넌스 파라미터 종합

스팸/게이밍 방어 관련 모든 거버넌스 파라미터를 정리한다.

| 파라미터 | 모듈/컨트랙트 | 초기값 | 조정 |
|----------|--------------|--------|------|
| **에포크 / 윈도우** | | | |
| `epoch_length` | x/activity | 17,280 블록 (~24시간) | 거버넌스 |
| `windows_per_epoch` | x/activity | 12 | 거버넌스 |
| `min_active_windows` | x/activity | 8 | 거버넌스 |
| **보상 분배** | | | |
| `D_min` | x/distribution | 0.3 (위임 풀 최소 30%) | 거버넌스 |
| **Activity Protocol** | | | |
| `max_activity_submissions_per_epoch` | x/activity | 100 | 거버넌스 |
| `max_feegrant_submissions_per_epoch` | x/activity | 10 | 거버넌스 |
| `activity_pruning_keep_blocks` | x/activity | 1,555,200 (90일) | 거버넌스 |
| **노드 등록** | | | |
| `min_self_delegation` | x/staking | 1 usum | CreateValidator 시 자동 설정 |
| `max_registrations_per_block` | x/node | 5 | 거버넌스 |
| `registration_cooldown_blocks` | x/node | 100 | 거버넌스 |
| `registration_deposit` | x/node | 0 usum | 거버넌스 |
| **Agent 권한** | | | |
| `agent_allowed_msg_types` | x/node | [MsgSubmitActivity, MsgExecuteContract, MsgSend] | 거버넌스 |
| `agent_address_change_cooldown` | x/node | 17,280 블록 (1 에포크) | 거버넌스 |
| **활동 비용 모델** | | | |
| `fee_threshold_multiplier` | x/activity | 3 | 거버넌스 |
| `base_activity_fee` | x/activity | 1,000,000 usum (1 KKOT) | 거버넌스 |
| `fee_exponent` | x/activity | 5000 (=0.5) | 거버넌스 |
| `max_activity_fee` | x/activity | 100,000,000 usum (100 KKOT) | 거버넌스 |
| `min_feegrant_quota` | x/activity | 8 | 거버넌스 |
| `quota_reduction_rate` | x/activity | 5000 (=0.5) | 거버넌스 |
| `feegrant_fee_exempt` | x/activity | true | 거버넌스 |
