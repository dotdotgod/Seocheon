# 14. 스케일링 전략

> **담당**: 체인 코어 / 아키텍트
> **관련 모듈**: `x/activity`, `x/node`
> **상태**: 병목 분석 완료 | Phase 1은 N_a 실제 증가 시점에 대응

---

## 1. 배경 및 설계 철학

Seocheon의 핵심 구조적 특징은 **N_a(활동 참여 노드)와 N_d(Active Validator Set)의 분리**다.

- **N_d (Active Validator Set)**: `max_validators` 거버넌스 파라미터로 상한 고정. 기본 150~200.
- **N_a (활동 자격 노드)**: REGISTERED + ACTIVE 통합. **상한 없음.**

이 구조는 의도적 설계다. AI 에이전트가 자유롭게 참여하고, 생태계가 가치 있는 에이전트를 위임으로 선별하는 메커니즘은 N_a에 진입 장벽을 두지 않아야 한다. 그러나 N_a의 무제한 성장은 온체인 연산 비용 증가로 이어진다.

**자연 경제 균형점**: N_a가 무한정 증가하면 에포크 보상 1/N_a로 희석 → 활동 자격 유지 비용 > 보상 → 자연 이탈. 이론적 균형점에서 N_a는 경제적으로 안정된다. 하지만 이 균형점에 도달하기 전에 블록 타임 초과가 발생하면 안 된다.

---

## 2. 확인된 병목 지점

### 🔴 병목 1: `getEligibleNodeIDs()` 전체 순회 (치명적)

**파일**: `x/activity/keeper/reward_distribution.go:126`

```go
iter, err := k.EpochSummary.Iterate(ctx, nil)  // nil = 전체 순회
// K1=node_id, K2=epoch 구조라 에포크 prefix 불가
// N_a 노드 × E 에포크 저장분 모두 스캔
```

`EpochSummary`는 `Map[Pair[node_id, epoch], Summary]`로 정의되어 있어 K1이 node_id, K2가 epoch다. 특정 에포크만 prefix 조회가 불가능하므로 전체 순회 후 `kv.Key.K2() != epoch`로 필터링한다.

| N_a | 저장 에포크 수 (prune 전) | 순회 항목 수 |
|-----|--------------------------|-------------|
| 100 | ~73 | ~7,300 |
| 1,000 | ~73 | ~73,000 |
| 10,000 | ~73 | **~730,000** |
| 100,000 | ~73 | **~7,300,000** |

에포크 경계 블록마다 이 순회가 실행된다. N_a > 5,000 시 Cosmos SDK 기본 블록 가스 한도(10억) 초과 위험.

### 🟠 병목 2: `getNextBlockSeq()` O(N_block) 반복 (중간)

**파일**: `x/activity/keeper/msg_server_submit_activity.go`

매 `MsgSubmitActivity` 처리 시 현재 블록 내 모든 기존 활동을 순회하여 max seq를 계산한다. 블록당 TX가 많을수록 누적 비용이 증가한다.

| 블록당 TX | 순회 비용 |
|-----------|-----------|
| 10 | 1+2+...+10 = 55 iter |
| 100 | 1+2+...+100 = 5,050 iter |
| 1,000 | ~500,000 iter |

### 🟠 병목 3: `pruneOldActivities()` 중첩 구조 (낮음)

**파일**: `x/activity/keeper/abci.go:82`

`BlockActivities`를 순회하여 node_id를 획득한 뒤, `Activities`에서 해당 node_id의 활동을 다시 순회하여 block_height를 비교한다. 현재 `maxPrunePerEpoch = 1000` 캡으로 에포크당 가스가 제한되나, N_a 증가 시 프루닝 백로그가 누적된다.

### 🟡 에포크 경계 블록 gas 집중

17,280블록 중 단 1개 블록에 아래 연산이 집중된다:

**x/node EndBlocker**:
- `processExpiredDelegations` — O(N_active_delegations)
- `distributeBoostPool` — O(N_validators)
- `applyPendingAgentShareChanges` — O(pending)

**x/activity EndBlocker**:
- `getEligibleNodeIDs` — O(N_a × E_stored) ← 주요 병목
- `DistributeActivityRewards` — O(N_eligible)
- `pruneOldActivities` — O(maxPrunePerEpoch)

---

## 3. N_a 규모별 지원 범위

| N_a 규모 | 현재 구조 | 예상 에포크 경계 블록 시간 | 권고 대응 |
|----------|-----------|--------------------------|-----------|
| ≤ 1,000 | ✅ 완전 지원 | < 1초 | 불필요 |
| 1,000~5,000 | ⚠️ EndBlocker 가스 증가 | 1~5초 | Phase 1 코드 최적화 권고 |
| 5,000~10,000 | ❌ 에포크 경계 블록 위험 | 5초+ (블록 타임 초과) | Phase 1 필수 |
| 10,000~100,000 | ❌ 단일 블록 연산 불가 | 불가 | Phase 2 필요 |
| 100,000+ | ❌ 단일 L1 한계 | 불가 | Phase 3 필요 |

> **현재 결정**: Phase 1 코드 최적화는 N_a가 실제로 1,000을 넘기는 시점에 대응한다.
> 제네시스 직후 N_a는 수십~수백 수준이므로 즉각 조치 불필요.

---

## 4. Phase별 대응 방안

### Phase 1 — 온체인 코드 최적화 (N_a ≤ 10,000 지원)

**1-A. EpochSummary 키 순서 교체**

현재 `EpochSummary: Map[Pair[node_id, epoch], Summary]`의 키 순서를 교체한다.

```go
// 변경 전: K1=node_id, K2=epoch → epoch prefix 불가
EpochSummary: Map[Pair[node_id, epoch], Summary]

// 변경 후: K1=epoch, K2=node_id → epoch prefix 스캔 가능
EpochSummary: Map[Pair[epoch, node_id], Summary]
```

- 제네시스 전이므로 기존 상태 마이그레이션 부담 없음
- 컬렉션 1개 유지 → 스토리지 추가 없음
- `IsNodeEligibleForEpoch(nodeID, epoch)` → `Get(Join(epoch, nodeID))` 인수 순서만 교체, O(1) 유지
- `getEligibleNodeIDs()` → `Iterate(prefix=epoch)` 로 교체
- 복잡도: **O(N_a × E_stored) → O(N_submitted_this_epoch)**
- 변경 파일: `x/activity/keeper/keeper.go`, `reward_distribution.go`, `abci.go`

**1-B. `getNextBlockSeq()` 카운터 전환**

현재 `BlockActivities`를 순회하여 max seq를 계산하는 방식을 카운터로 교체한다.

```go
// 추가: 블록별 활동 카운터
BlockActivityCount: Map[int64, uint64]  // block_height → count
```

- 매 TX마다 O(N_block) → O(1)
- 변경 파일: `x/activity/keeper/keeper.go`, `msg_server_submit_activity.go`

**1-C. BlockActivities value 타입 확장**

현재 value가 `string(node_id)`인 것을 `struct{node_id, epoch}`로 확장한다.

- 프루닝 시 epoch 기반 직접 삭제 가능 → 중첩 순회 제거
- 변경 파일: `x/activity/keeper/keeper.go`, `abci.go`, 관련 proto

### Phase 2 — 에포크 보상 배치 분산 (N_a 10,000~100,000 지원)

**문제**: 에포크 경계 1개 블록에서 N_eligible × 2번의 bank.Send(operator + agent)가 동시 실행된다.
bank.Send 1번 = KV write 2건(잔액 차감·증가) + 이벤트 emit으로, N_a=10,000이면 단일 블록에서 20,000 KV write가 발생한다.

**해결**: 에포크 경계 블록에서는 **계산만 수행하고 대기열에 기록**, 실제 bank.Send는 이후 블록에 분산한다.

#### 신규 Collections

```go
// x/activity/keeper/keeper.go에 추가

// 배치 대기 보상: 순차 인덱스 → RewardEntry
// 인덱스 키 사용 이유: prefix 스캔으로 [processed, processed+batch_size) 범위 처리 가능
PendingRewardEntries   collections.Map[uint64, types.RewardEntry]

// 전체 대기 수 (에포크 경계에서 기록)
PendingRewardCount     collections.Item[uint64]

// 처리 완료 수 (매 배치 블록에서 증가)
PendingRewardProcessed collections.Item[uint64]
```

```protobuf
// proto/seocheon/activity/v1/types.proto에 추가
message RewardEntry {
  string operator_address = 1;
  string agent_address    = 2;
  string operator_amount  = 3;  // uppyeo, math.Int 직렬화
  string agent_amount     = 4;
}
```

#### 상태 전이

```
에포크 E 마지막 블록 (EndBlocker, IsEpochBoundary=true):
  1. getEligibleNodeIDs(epoch) → eligibleNodes       ← Phase 1-A 이후 O(N_eligible)
  2. perNodeReward = activityPool / len(eligibleNodes)
  3. for seq, nodeID := range eligibleNodes:
       operatorAmt, agentAmt = split(perNodeReward, agentShare)
       PendingRewardEntries[seq] = RewardEntry{...}  ← bank.Send 없음
  4. PendingRewardCount     = len(eligibleNodes)
  5. PendingRewardProcessed = 0
  emit: EventRewardDistributionScheduled{epoch, eligible_count}

에포크 E+1 이후, 매 블록 EndBlocker (비경계 블록):
  processed = PendingRewardProcessed (없으면 skip)
  total     = PendingRewardCount
  if processed >= total: skip (완료)

  end = min(processed + batch_size, total)
  for seq in [processed, end):
    entry = PendingRewardEntries[seq]
    bank.SendCoinsFromModuleToAccount(activity_pool → operator, entry.OperatorAmount)
    bank.SendCoinsFromModuleToAccount(activity_pool → agent,    entry.AgentAmount)
    delete PendingRewardEntries[seq]
  PendingRewardProcessed = end
  emit: EventRewardDistributionBatch{batch_start: processed, batch_end: end}
```

#### 가스 예산 분석

| batch_size | 배치 블록당 bank.Send | 배치 블록 추가 시간 | N_eligible=10,000 완료 블록 수 | 분산 소요 시간 |
|-----------|----------------------|-------------------|-------------------------------|--------------|
| 250       | 500                  | ~1.5초             | 40 블록                        | ~4분          |
| 500       | 1,000                | ~3초               | 20 블록                        | ~2분          |
| 1,000     | 2,000                | ~6초               | 10 블록                        | ~1분          |

> bank.Send 1번 ≈ KV write 2건 + 이벤트 ≈ 약 100,000 gas.
> 블록 가스 한도 기본값 10억 기준: batch_size=500일 때 배치 블록 부하 ≈ 10%.
> 기본값 `batch_size=500` 권고.

#### 주의: 에포크 간 배치 겹침

N_eligible이 매우 크면 다음 에포크 경계가 오기 전에 이전 에포크 배치가 완료되지 않을 수 있다.
이 경우 에포크 경계 블록에서 PendingRewardProcessed < PendingRewardCount이면:
- 새 집계를 덮어쓰지 않고 **에러 로그 emit + 신규 분배 건너뜀**
- 운영자가 batch_size를 늘리거나 Phase 3(샤딩)으로 전환해야 함을 알리는 거버넌스 이벤트 발생

#### 변경 파일

- `x/activity/types/` — `RewardEntry` proto 추가, `params.go`에 `RewardBatchSize` 파라미터
- `x/activity/keeper/keeper.go` — 3개 컬렉션 추가
- `x/activity/keeper/reward_distribution.go` — `DistributeActivityRewards()`: bank.Send → PendingRewardEntries 기록으로 교체
- `x/activity/keeper/abci.go` — `EndBlocker()`: 배치 처리 분기 추가

### Phase 3 — Activity Shard Chain (N_a 100,000+)

단일 L1으로 처리 불가능한 규모에서는 **IBC 연결 전용 샤드 체인**으로 분리한다.

```
Seocheon Main Chain
├── x/node (등록/관리)
├── x/distribution (보상 분배)
└── IBC Channel ←──────────────────────────────┐
                                                │
Activity Shard Chain (별도 L1)                  │
├── x/activity (MsgSubmitActivity 처리)          │
├── 에포크 집계                                  │
└── KKOT 앵커링: 보상 요청 → IBC → Main Chain ──┘
```

- KKOT는 Main Chain에서만 발행/관리, Shard는 IBC로 앵커링
- Shard는 활동 특화 파라미터 (짧은 블록 시간, 높은 TPS) 사용
- 여러 Shard 병렬 운영 가능 (지역별, 역할별)

### 참고: ZK Rollup

ZK-SNARK/STARK 기반 활동 증명 롤업. Cosmos SDK와의 통합 복잡도가 매우 높아 현실적이지 않다. 장기 참고 방향으로만 유지.

---

## 5. 거버넌스 파라미터 연계

Phase 1~2 구현 시 아래 파라미터를 추가한다.

| 파라미터 | 기본값 | 설명 |
|---------|--------|------|
| `reward_batch_size` | 500 | 에포크 경계 이후 블록당 처리할 bank.Send 쌍 수 (operator+agent 각 1회 = 1쌍). 값이 클수록 분산 기간 단축, 배치 블록 가스 부하 증가. N_eligible=10,000 기준 권장: 500 (20블록, ~2분). |
| `prune_batch_size` | 1,000 | 에포크당 최대 프루닝 수 (현재 `maxPrunePerEpoch` 상수를 파라미터화) |

---

## 6. 관련 문서

- [02_core_concepts.md](02_core_concepts.md) — 에포크/윈도우, 이중 보상 풀 수식
- [04_activity_protocol.md](04_activity_protocol.md) — MsgSubmitActivity, ActivityRecord 구조
- [10_ibc_strategy.md](10_ibc_strategy.md) — IBC Transfer 전략 (Phase 3 기반)
- [architecture_review.md](../architecture_review.md) — 보완 사항 추적
