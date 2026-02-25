# 긴급 정지 메커니즘 (Circuit Breaker)

> **담당**: 체인 코어 개발자 / 보안
> **관련 문서**: [개요](01_overview.md) · [노드 모듈](03_node_module.md) · [Activity Protocol](04_activity_protocol.md) · [스팸 방어](08_spam_defense.md) · [구현 가이드](09_implementation.md) · [체인 업그레이드](10_chain_upgrade.md) · [재단 전략](../foundation_strategy.md) · [전체 목차](README.md)

> **구현 상태**: 현재 Cosmos SDK 표준 `x/circuit` 모듈만 적용되어 있다. 이 문서에 기술된 Seocheon 커스텀 Circuit Breaker는 Phase 2에서 구현 예정이다.

---

## 1. 개요

### Circuit Breaker 개요

이상 트래픽 급증, 모듈 레벨 취약점 악용, 합의 불안정 시 특정 기능을 선택적으로 정지하는 메커니즘이다.

### 설계 원칙

```
① 최소 범위 정지: 문제가 발생한 모듈/기능만 선택적으로 정지
② 거버넌스 통제: 수동 트리거는 반드시 거버넌스 투표로 승인
③ 자동 복구 옵션: 자동 트리거의 경우 조건 해소 시 자동 복구 가능
④ 투명성: 정지 상태, 트리거 사유, 복구 조건을 온체인에 기록
⑤ 탈중앙화 호환: 재단 주도에서 커뮤니티 주도로 자연스럽게 전환
```

---

## 2. 트리거 조건

### 자동 트리거

체인이 사전 정의된 임계치를 감지하여 자동으로 Circuit Breaker를 활성화한다. 사람의 개입 없이 즉각 대응할 수 있는 1차 방어선이다.

```
자동 트리거 조건:

[A] 이상 활동 급증 (x/activity Circuit Breaker)
    조건: 단일 윈도우(1,440 블록) 내 전체 MsgSubmitActivity 수 > activity_surge_threshold
    기본값: activity_surge_threshold = 10,000 (거버넌스 파라미터)
    동작: x/activity 모듈의 MsgSubmitActivity 수용 일시 정지
    설계 근거:
      → 정상 상태: 100 노드 × 에포크당 50건 ÷ 12 윈도우 ≈ 윈도우당 417건
      → 10,000건은 정상 대비 ~24배 이상으로, 명백한 이상 상태
      → 임계치는 노드 수 증가에 따라 거버넌스로 조정

[B] 연속 블록 누락 (합의 불안정 감지)
    조건: 연속 missed_blocks_threshold 블록 동안 2/3+ 밸리데이터 서명 누락
    기본값: missed_blocks_threshold = 100 (거버넌스 파라미터)
    동작: 전체 TX 처리 속도 제한 (블록당 최대 TX 수 축소)
    설계 근거:
      → 100블록 연속 서명 누락은 네트워크 분할 또는 대규모 장애 징후
      → 완전 정지가 아닌 속도 제한으로 합의 부하 경감

[C] 컨트랙트 가스 이상 (CosmWasm 보호)
    조건: 단일 블록 내 CosmWasm 실행 가스 합계 > contract_gas_threshold
    기본값: contract_gas_threshold = 블록 최대 가스의 80%
    동작: MsgExecuteContract 수용 일시 정지 (디렉토리 컨트랙트 포함)
    설계 근거:
      → 컨트랙트가 블록 가스의 대부분을 소비하면 일반 TX 처리 불가
      → 컨트랙트 취약점 악용 또는 DoS 공격 패턴
```

### 수동 트리거

재단 또는 커뮤니티가 거버넌스 제안을 통해 Circuit Breaker를 활성화한다.

```
수동 트리거 절차:

  [1] CircuitBreakerProposal 제출
      ├── target_module: 정지 대상 ("activity" | "directory" | "all")
      ├── reason: 정지 사유 (문자열)
      ├── duration_blocks: 정지 기간 (0 = 거버넌스로 해제할 때까지)
      └── severity: 정지 수준 ("partial" | "full")

  [2] 거버넌스 투표
      ├── 일반: voting_period 172,800 블록 (~10일)
      └── 긴급: ExpeditedProposal 17,280 블록 (~1일)

  [3] 투표 통과 → Circuit Breaker 즉시 활성화
      └── 온체인 이벤트 발행: EventCircuitBreakerActivated

재단의 긴급 제안 시나리오:
  → 보안 취약점 발견 → 재단이 ExpeditedCircuitBreakerProposal 제출
  → 단축된 투표 기간(~1일) + 높은 통과 기준(quorum 67%, threshold 67%)
  → 통과 시 즉시 해당 모듈 정지
```

---

## 3. 정지 범위

### 모듈별 정지

문제가 발생한 모듈만 선택적으로 정지한다. 나머지 체인 기능은 정상 동작을 유지한다.

```
x/activity 정지:
  ├── 차단: MsgSubmitActivity 수용 거부
  ├── 유지: 기존 활동 기록 조회, 보상 분배 (이미 기록된 데이터 기반)
  ├── 유지: 노드 등록/비활성화, 위임/철회, 거버넌스 투표
  └── 영향: 새 활동 제출 불가, 진행 중인 에포크의 활동 자격에 영향

x/node 정지:
  ├── 차단: MsgRegisterNode, MsgDeactivateNode 수용 거부
  ├── 유지: 기존 노드 상태 조회, 위임/철회, 활동 제출
  ├── 유지: StakingHooks에 의한 자동 졸업/탈락은 계속 동작
  └── 영향: 새 노드 등록 불가, 노드 비활성화 불가

x/activity + x/node 동시 정지:
  ├── 차단: MsgSubmitActivity, MsgRegisterNode, MsgDeactivateNode
  ├── 유지: 위임/철회, 거버넌스, 토큰 전송, 합의 참여
  └── 용도: 커스텀 모듈 전체에 대한 긴급 점검
```

### CosmWasm 컨트랙트 정지

CosmWasm 컨트랙트는 컨트랙트 레벨의 pause 메커니즘과 체인 레벨의 Circuit Breaker를 이중으로 적용한다.

```
컨트랙트 레벨 pause (1차):
  ├── 컨트랙트의 admin(재단)이 Pause 실행 메시지 호출
  ├── 컨트랙트 내부 상태: is_paused = true
  ├── 차단: 상태 변경 실행 메시지
  ├── 유지: 쿼리
  └── 복구: admin이 Unpause 실행 메시지 호출

체인 레벨 Circuit Breaker (2차):
  ├── CircuitBreaker keeper가 MsgExecuteContract를 대상 컨트랙트로 필터링
  ├── 차단: 특정 컨트랙트 주소에 대한 모든 MsgExecuteContract
  ├── 유지: 다른 컨트랙트의 MsgExecuteContract는 정상 처리
  └── 복구: 거버넌스 투표 또는 자동 복구 타이머

이중 방어의 의미:
  → 컨트랙트 admin 키가 유출되어도 체인 레벨에서 정지 가능
  → 컨트랙트 자체의 pause 로직에 버그가 있어도 체인에서 차단 가능
```

### 전체 트랜잭션 정지 (극단적 상황)

합의 레이어 취약점이나 체인 전체에 영향을 미치는 상황에서만 사용한다.

```
전체 TX 정지:
  ├── 차단: 모든 사용자 TX (MsgSend, MsgDelegate 등 포함)
  ├── 유지: 합의 (블록 생성, 밸리데이터 서명)
  ├── 유지: 거버넌스 투표 (정지 해제를 위해 반드시 유지)
  └── 적용 조건: CircuitBreakerProposal의 severity = "full"일 때만

주의사항:
  → 전체 정지는 최후의 수단
  → 거버넌스 TX는 항상 허용 (정지 해제 투표를 위해)
  → duration_blocks를 반드시 설정하여 자동 해제 보장
```

---

## 4. 복구 절차

### 거버넌스 투표를 통한 정상화

수동 트리거로 활성화된 Circuit Breaker는 거버넌스 투표로 해제한다.

```
정상화 절차:

  [1] 문제 원인 파악 및 패치 개발/적용

  [2] CircuitBreakerDeactivateProposal 제출
      ├── target_module: 해제 대상
      └── reason: 해제 사유 (패치 완료, 위험 해소 등)

  [3] 거버넌스 투표 (일반 또는 긴급)

  [4] 투표 통과 → Circuit Breaker 즉시 해제
      └── 온체인 이벤트 발행: EventCircuitBreakerDeactivated

  [5] 모니터링
      ├── 해제 후 2 에포크(~2일) 집중 모니터링
      ├── 이상 재발 시 재활성화 준비
      └── 사후 분석 보고서 공개
```

### 자동 복구 조건 (타이머 기반)

자동 트리거로 활성화된 Circuit Breaker는 조건 해소 시 자동 복구된다.

```
자동 복구 조건:

[A] 이상 활동 급증 → 자동 복구
    해제 조건: 연속 3 윈도우(4,320 블록, ~6시간) 동안
               윈도우당 MsgSubmitActivity 수 < activity_surge_threshold × 0.5
    동작: x/activity Circuit Breaker 자동 해제
    안전장치: 최대 자동 정지 기간 = 1 에포크(17,280 블록)
              초과 시 거버넌스 투표 필요

[B] 연속 블록 누락 → 자동 복구
    해제 조건: 연속 100 블록 동안 2/3+ 밸리데이터 정상 서명
    동작: TX 속도 제한 자동 해제
    안전장치: 최대 자동 정지 기간 = 1 에포크

[C] 컨트랙트 가스 이상 → 자동 복구
    해제 조건: 연속 100 블록 동안 CosmWasm 가스 합계 < contract_gas_threshold × 0.5
    동작: MsgExecuteContract 수용 재개
    안전장치: 최대 자동 정지 기간 = 1 에포크

자동 복구 공통 규칙:
  → 자동 복구 횟수 제한: 에포크당 최대 3회 자동 복구
  → 3회 초과 시: Circuit Breaker 유지 + 거버넌스 투표 필요
  → 반복적 트리거는 근본 원인 해결이 필요한 신호
```

---

## 5. 구현 설계

### CircuitBreaker keeper 인터페이스

```go
// CircuitBreaker keeper 인터페이스
type CircuitBreakerKeeper interface {
    // 상태 조회
    IsModulePaused(ctx context.Context, moduleName string) bool
    IsContractPaused(ctx context.Context, contractAddr string) bool
    GetCircuitBreakerStatus(ctx context.Context) CircuitBreakerStatus

    // 자동 트리거
    CheckAndTrigger(ctx context.Context) error

    // 거버넌스 트리거
    ActivateCircuitBreaker(ctx context.Context, proposal CircuitBreakerProposal) error
    DeactivateCircuitBreaker(ctx context.Context, proposal CircuitBreakerDeactivateProposal) error

    // 자동 복구
    CheckAndRecover(ctx context.Context) error
}
```

### 상태 저장 구조

```protobuf
// Circuit Breaker 상태
message CircuitBreakerState {
  string module_name = 1;           // 정지 대상 모듈 ("activity", "node", "directory", "all")
  bool is_active = 2;               // Circuit Breaker 활성 여부
  string trigger_type = 3;          // "auto" | "governance"
  string reason = 4;                // 정지 사유
  int64 activated_at = 5;           // 활성화 블록 높이
  int64 duration_blocks = 6;        // 정지 기간 (0 = 거버넌스 해제까지)
  int64 auto_recovery_count = 7;    // 현재 에포크 내 자동 복구 횟수
  string severity = 8;              // "partial" | "full"
}

// Circuit Breaker 이력
message CircuitBreakerEvent {
  string module_name = 1;
  string action = 2;               // "activated" | "deactivated" | "auto_recovered"
  string trigger_type = 3;
  string reason = 4;
  int64 block_height = 5;
}
```

### AnteHandler 통합

Circuit Breaker는 Cosmos SDK의 AnteHandler 체인에 통합되어, TX 실행 전에 정지 여부를 확인한다.

```
AnteHandler 실행 순서:

  [1] 표준 AnteHandler (서명 검증, 가스 계산 등)
  [2] CircuitBreakerDecorator ← 여기에 삽입
      ├── TX의 메시지 유형 확인
      ├── 해당 모듈의 Circuit Breaker 상태 조회
      ├── 활성 상태 → TX 거부 (에러 코드 + 정지 사유 반환)
      └── 비활성 상태 → 다음 AnteHandler로 전달
  [3] 나머지 AnteHandler
```

```go
// CircuitBreakerDecorator 동작 의사 코드
func (cbd CircuitBreakerDecorator) AnteHandle(ctx Context, tx Tx, simulate bool) error {
    for _, msg := range tx.GetMsgs() {
        moduleName := getModuleForMsg(msg)

        if cbd.keeper.IsModulePaused(ctx, moduleName) {
            // 거버넌스 TX는 항상 허용
            if isGovernanceTx(msg) {
                continue
            }
            return ErrCircuitBreakerActive{
                Module: moduleName,
                Reason: cbd.keeper.GetReason(ctx, moduleName),
            }
        }

        // 컨트랙트별 필터링
        if execMsg, ok := msg.(*wasmtypes.MsgExecuteContract); ok {
            if cbd.keeper.IsContractPaused(ctx, execMsg.Contract) {
                return ErrContractPaused{Contract: execMsg.Contract}
            }
        }
    }
    return next(ctx, tx, simulate)
}

// 메시지 → 모듈 매핑
func getModuleForMsg(msg sdk.Msg) string {
    switch msg.(type) {
    case *activitytypes.MsgSubmitActivity:
        return "activity"
    case *nodetypes.MsgRegisterNode, *nodetypes.MsgDeactivateNode:
        return "node"
    case *wasmtypes.MsgExecuteContract:
        return "wasm"
    default:
        return ""
    }
}
```

### EndBlocker 통합

자동 트리거와 자동 복구는 EndBlocker에서 실행된다.

```
EndBlocker 동작:

  매 블록:
    [1] CheckAndTrigger()
        ├── 윈도우 내 MsgSubmitActivity 카운트 확인 → 임계치 초과 시 활성화
        ├── 연속 미서명 블록 카운트 확인 → 임계치 초과 시 속도 제한
        └── CosmWasm 가스 사용량 확인 → 임계치 초과 시 활성화

    [2] CheckAndRecover()
        ├── 활성 상태인 자동 트리거 Circuit Breaker 대상
        ├── 복구 조건 충족 여부 확인
        ├── 에포크당 자동 복구 횟수 확인 (최대 3회)
        └── 조건 충족 + 횟수 미초과 → 자동 해제

    [3] 타이머 기반 만료 확인
        ├── duration_blocks > 0인 거버넌스 Circuit Breaker 대상
        ├── 현재 블록 - activated_at >= duration_blocks → 자동 해제
        └── 이벤트 발행: EventCircuitBreakerExpired
```

---

## 6. 재단 연계 및 탈중앙화 전환

### Phase A/B/C 연계

재단 전략(foundation_strategy.md)의 탈중앙화 단계에 따라 Circuit Breaker 운영 주체가 전환된다.

```
Phase A: 재단 주도 (Genesis ~ Active Validator Set ≥ 15)
  ├── Circuit Breaker 거버넌스 제안: 재단 주도
  ├── 컨트랙트 pause 실행: 재단이 admin으로서 직접 실행
  ├── 자동 트리거 파라미터 조정: 재단이 거버넌스 제안
  └── 의사결정 속도: 빠름 (재단 투표력)

Phase B: 혼합 (Active Validator Set ≥ 15 ~ ≥ 30 AND 12개월)
  ├── Circuit Breaker 거버넌스 제안: 재단 + 커뮤니티 밸리데이터
  ├── 컨트랙트 admin: 멀티시그로 전환 (재단 + 커뮤니티 대표 2-3인)
  ├── 자동 트리거 파라미터: 커뮤니티 거버넌스로 조정
  └── 커뮤니티 밸리데이터에게 긴급 대응 역할 인수

Phase C: 커뮤니티 주도 (Active Validator Set ≥ 30 AND 12개월 이후)
  ├── Circuit Breaker 거버넌스 제안: 커뮤니티 Security Council
  ├── 컨트랙트 admin: 거버넌스 컨트랙트 (x/gov)로 이전
  │   → admin 권한이 거버넌스 모듈에 있으므로
  │     pause/unpause도 거버넌스 투표로만 실행
  ├── 재단 역할: 긴급 상황 시 거버넌스 제안만 (투표권은 다른 참여자와 동일)
  └── 자동 트리거: 파라미터 조정 포함 모든 결정이 커뮤니티 거버넌스
```

### 컨트랙트 admin 전환 절차

```
컨트랙트 admin 전환:

Phase A:
  admin = 재단 주소 (seocheon1foundation...)

Phase B:
  [1] 멀티시그 주소 생성 (재단 + 커뮤니티 대표)
  [2] UpdateAdmin 거버넌스 제안
      ├── contract: 디렉토리 컨트랙트 주소
      └── new_admin: 멀티시그 주소
  [3] 거버넌스 투표 통과 → admin 변경

Phase C:
  [1] UpdateAdmin 거버넌스 제안
      ├── contract: 디렉토리 컨트랙트 주소
      └── new_admin: x/gov 모듈 주소
  [2] 거버넌스 투표 통과 → admin이 거버넌스 모듈로 이전
  [3] 이후 pause/unpause는 거버넌스 제안으로만 실행
```

---

## 7. 모니터링 및 알림

```
Circuit Breaker 모니터링 항목:

  [A] 자동 트리거 근접도
      ├── 윈도우당 MsgSubmitActivity 수 / activity_surge_threshold
      ├── 연속 미서명 블록 수 / missed_blocks_threshold
      └── 블록당 CosmWasm 가스 / contract_gas_threshold
      → 임계치의 70% 도달 시 경고 알림

  [B] Circuit Breaker 활성 상태
      ├── 활성 상태 지속 시간
      ├── 자동 복구 시도 횟수
      └── 거버넌스 해제 투표 진행 상황

  [C] 이력 분석
      ├── Circuit Breaker 활성화 빈도 (에포크당)
      ├── 평균 정지 기간
      ├── 자동 트리거 vs 수동 트리거 비율
      └── 반복적 트리거 패턴 감지
```

---

## 거버넌스 파라미터

| 파라미터 | 모듈 | 초기값 | 조정 |
|----------|------|--------|------|
| `activity_surge_threshold` | x/circuitbreaker | 10,000 (윈도우당) | 거버넌스 |
| `missed_blocks_threshold` | x/circuitbreaker | 100 (연속 블록) | 거버넌스 |
| `contract_gas_threshold` | x/circuitbreaker | 블록 최대 가스의 80% | 거버넌스 |
| `max_auto_recovery_per_epoch` | x/circuitbreaker | 3 | 거버넌스 |
| `max_auto_pause_duration` | x/circuitbreaker | 17,280 블록 (1 에포크) | 거버넌스 |
| `recovery_observation_windows` | x/circuitbreaker | 3 윈도우 | 거버넌스 |
| `recovery_threshold_ratio` | x/circuitbreaker | 0.5 (임계치의 50% 이하) | 거버넌스 |
