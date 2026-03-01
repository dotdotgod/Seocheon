# 긴급 정지 메커니즘 (Circuit Breaker)

> **담당**: 체인 코어 개발자 / 보안
> **관련 문서**: [개요](01_overview.md) · [노드 모듈](03_node_module.md) · [Activity Protocol](04_activity_protocol.md) · [스팸 방어](08_spam_defense.md) · [구현 가이드](09_implementation.md) · [체인 업그레이드](10_chain_upgrade.md) · [재단 전략](../foundation_strategy.md) · [전체 목차](README.md)

> **구현 상태**: 현재 Cosmos SDK 표준 `x/circuit` 모듈이 적용되어 있다. Phase 2에서 Guardian 권한 설정과 모듈별 선택적 정지 기능을 추가한다.

> **설계 결정**: 자동 트리거는 검토 결과 보류한다. Seocheon은 DeFi가 아니므로 밀리초 단위 자동 대응이 불필요하며, 기존 스팸 방어(쿼터 10건/에포크, feegrant 1 KKOT 제한)가 활동 급증을 이미 차단한다. 밸리데이터 셋 규모(150~200)에서 수동 대응 조율이 충분히 가능하고, 자동 트리거의 오탐 리스크(정상 활동을 공격으로 오판하여 체인 불필요 정지)가 이점보다 크다.

---

## 1. 개요

### Circuit Breaker 개요

보안 취약점 악용, 합의 불안정 시 특정 기능을 선택적으로 정지하는 메커니즘이다.

### 설계 원칙

```
① 최소 범위 정지: 문제가 발생한 모듈/기능만 선택적으로 정지
② 거버넌스 통제: 정지와 해제 모두 거버넌스 투표로 승인
③ 투명성: 정지 상태, 트리거 사유, 복구 조건을 온체인에 기록
④ 탈중앙화 호환: 재단 주도에서 커뮤니티 주도로 자연스럽게 전환
```

---

## 2. 트리거 방식

### 수동 트리거 (유일한 트리거)

재단 또는 커뮤니티가 거버넌스 제안을 통해 Circuit Breaker를 활성화한다.

```
수동 트리거 절차:

  [1] CircuitBreakerProposal 제출
      ├── target_module: 정지 대상 ("activity" | "node" | "all")
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

### Guardian 권한 (x/circuit 활용)

Cosmos SDK 표준 `x/circuit` 모듈의 권한 체계를 활용하여, 거버넌스 투표 없이 즉각 정지할 수 있는 Guardian 역할을 설정한다.

```
Guardian 권한 설계:

  Guardian 주소: 재단 멀티시그 (Phase A/B), Security Council (Phase C)

  권한 수준:
    ├── LEVEL_SUPER_ADMIN: 모든 Msg 타입 비활성화 가능
    ├── LEVEL_ALL_MSGS: 지정된 Msg 타입 비활성화 가능
    └── LEVEL_SOME_MSGS: 특정 Msg 타입만 비활성화 가능

  즉각 대응 플로우:
    [1] 보안 이슈 감지
    [2] Guardian이 x/circuit의 MsgAuthorizeCircuitBreaker 실행
    [3] 대상 Msg 타입 즉시 차단
    [4] 사후 거버넌스 투표로 해제 또는 유지 결정

  이점:
    → 거버넌스 투표 대기(~1일) 없이 수분 내 대응
    → x/circuit 표준이므로 추가 커스텀 모듈 불필요
    → Guardian 권한 자체도 거버넌스로 변경 가능
```

### 자동 트리거 — 보류

```
자동 트리거 보류 사유:

  [1] 기존 방어선이 충분하다
      → 활동 쿼터 10건/에포크 + feegrant 1 KKOT 제한
      → 가스비 체계에 의한 스팸 억제
      → 활동 급증 자체가 쿼터에서 차단됨

  [2] 오탐 리스크가 이점보다 크다
      → 생태계 성장에 따라 임계치를 지속적으로 조정해야 함
      → 정상적인 활동 증가를 공격으로 오판할 가능성
      → 불필요한 체인 정지는 사용자 신뢰를 손상

  [3] 수동 대응이 현실적이다
      → 밸리데이터 150~200: 조율 가능한 규모
      → Guardian 즉각 대응 + 긴급 거버넌스(~1일)
      → 24/7 모니터링으로 이상 징후 사전 감지

  재검토 조건:
      → 밸리데이터 수 1,000+ 초과 시
      → DeFi 기능(대출, 파생상품 등) 추가 시
      → 수동 대응 실패 사례 발생 시
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

### Guardian에 의한 즉시 해제

```
Guardian 해제 절차:

  [1] 문제 해소 확인
  [2] Guardian이 x/circuit의 MsgResetCircuitBreaker 실행
  [3] 대상 Msg 타입 즉시 재활성화
  [4] 사후 거버넌스 보고 (투명성 확보)
```

### 타이머 기반 자동 해제

```
duration_blocks > 0으로 설정된 경우:
  → 현재 블록 - activated_at >= duration_blocks → 자동 해제
  → 이벤트 발행: EventCircuitBreakerExpired
  → 거버넌스 투표 없이 자동 복구
  → 전체 TX 정지(severity = "full") 시 반드시 설정 권장
```

---

## 5. 구현 설계

### x/circuit 표준 활용

Cosmos SDK 표준 `x/circuit` 모듈을 기반으로 구현한다. 커스텀 모듈은 최소화한다.

```
구현 범위 (Phase 2):

  표준 x/circuit 활용:
    ├── Guardian 권한 설정 (재단 멀티시그)
    ├── Msg 타입별 비활성화/재활성화
    └── 권한 변경 거버넌스 연동

  커스텀 확장 (최소):
    ├── CircuitBreakerProposal: 거버넌스 제안 타입 추가
    ├── CircuitBreakerDeactivateProposal: 해제 제안 타입 추가
    ├── duration_blocks 기반 자동 해제 (EndBlocker)
    └── 이력 기록 (CircuitBreakerEvent)
```

### 상태 저장 구조

```protobuf
// Circuit Breaker 상태 (커스텀 확장)
message CircuitBreakerState {
  string module_name = 1;           // 정지 대상 모듈 ("activity", "node", "all")
  bool is_active = 2;               // Circuit Breaker 활성 여부
  string trigger_type = 3;          // "guardian" | "governance"
  string reason = 4;                // 정지 사유
  int64 activated_at = 5;           // 활성화 블록 높이
  int64 duration_blocks = 6;        // 정지 기간 (0 = 거버넌스 해제까지)
  string severity = 7;              // "partial" | "full"
}

// Circuit Breaker 이력
message CircuitBreakerEvent {
  string module_name = 1;
  string action = 2;               // "activated" | "deactivated" | "expired"
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
      ├── x/circuit 상태 조회 (해당 Msg 타입이 비활성화되었는가?)
      ├── 비활성화 상태 → TX 거부 (에러 코드 + 정지 사유 반환)
      └── 정상 상태 → 다음 AnteHandler로 전달
  [3] 나머지 AnteHandler

  예외: 거버넌스 TX는 항상 허용 (정지 해제 투표를 위해)
```

### EndBlocker 통합

타이머 기반 자동 해제만 EndBlocker에서 처리한다.

```
EndBlocker 동작:

  매 블록:
    [1] 타이머 만료 확인
        ├── duration_blocks > 0인 Circuit Breaker 대상
        ├── 현재 블록 - activated_at >= duration_blocks → 자동 해제
        └── 이벤트 발행: EventCircuitBreakerExpired
```

---

## 6. 재단 연계 및 탈중앙화 전환

### Phase A/B/C 연계

재단 전략(foundation_strategy.md)의 탈중앙화 단계에 따라 Circuit Breaker 운영 주체가 전환된다.

```
Phase A: 재단 주도 (Genesis ~ Active Validator Set ≥ 15)
  ├── Guardian 권한: 재단 단일 주소
  ├── 체인 레벨 pause 실행: 재단이 Guardian으로서 직접 실행
  └── 의사결정 속도: 빠름 (재단 투표력 + Guardian 즉각 대응)

Phase B: 혼합 (Active Validator Set ≥ 15 ~ ≥ 30 AND 12개월)
  ├── Guardian 권한: 멀티시그 (재단 + 커뮤니티 대표 2-3인)
  └── 커뮤니티 밸리데이터에게 긴급 대응 역할 인수

Phase C: 커뮤니티 주도 (Active Validator Set ≥ 30 AND 12개월 이후)
  ├── Guardian 권한: Security Council (커뮤니티 선출)
  └── 재단 역할: 긴급 상황 시 거버넌스 제안만 (투표권은 다른 참여자와 동일)
```

---

## 7. 모니터링 및 알림

```
Circuit Breaker 모니터링 항목:

  [A] Circuit Breaker 활성 상태
      ├── 활성 상태 지속 시간
      ├── 정지 대상 모듈 및 Msg 타입
      └── 거버넌스 해제 투표 진행 상황

  [B] 이상 징후 사전 감지 (수동 대응 지원)
      ├── 윈도우당 MsgSubmitActivity 수 추이
      ├── 밸리데이터 서명 참여율 추이
      └── 비정상 패턴 감지 시 Guardian에게 알림

  [C] 이력 분석
      ├── Circuit Breaker 활성화 빈도 (에포크당)
      ├── 평균 정지 기간
      └── Guardian 대응 vs 거버넌스 대응 비율
```

---

## 거버넌스 파라미터

| 파라미터 | 모듈 | 초기값 | 조정 |
|----------|------|--------|------|
| Guardian 주소 | x/circuit | 재단 멀티시그 | 거버넌스 |
| Guardian 권한 수준 | x/circuit | LEVEL_SUPER_ADMIN | 거버넌스 |
| 긴급 투표 기간 | x/gov | 17,280 블록 (~1일) | 거버넌스 |
| 긴급 투표 quorum | x/gov | 67% | 거버넌스 |
| 긴급 투표 threshold | x/gov | 67% | 거버넌스 |
