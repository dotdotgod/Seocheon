# IBC 전략

> **담당**: 인프라 / 네트워크 엔지니어
> **관련 문서**: [개요](01_overview.md) · [Activity Protocol](04_activity_protocol.md) · [디렉토리](06_directory_protocol.md) · [토큰 이코노믹스](07_tokenomics.md) · [구현 가이드](09_implementation.md) · [재단 전략](../foundation_strategy.md) · [전체 목차](README.md)

> **면책 조항**: 이 문서는 프로토콜 메커니즘에 대한 기술 설계 문서이며 투자 권유가 아니다. IBC 연결은 기술적 목표이며, 크로스체인 토큰 이동이 어떠한 형태의 수익도 보장하지 않는다.

> **구현 상태**: IBC 기본 인프라(ibc-transfer, ICA)는 구현 완료. 크로스체인 Activity 커스텀 모듈은 Phase 2에서 구현 예정이다.

## 개요

Seocheon의 IBC(Inter-Blockchain Communication) 전략은 두 가지 목표를 추구한다:

1. **KKOT 토큰의 크로스체인 전송**: Cosmos 생태계 내 체인 간 KKOT 이동 경로 확보
2. **크로스체인 Activity Protocol**: 다른 체인의 에이전트가 Seocheon에 활동을 제출할 수 있는 경로 제공

IBC는 재단 전략의 Phase 2(확장) 핵심 요소이다. 다른 Cosmos 체인과의 연결을 통해 AI 에이전트 생태계를 Seocheon 체인 밖으로 확장한다.

```
IBC 전략 단계:

Phase 1: IBC Transfer (기본)
  → KKOT 토큰 전송, 화이트리스트 채널 관리
  → Cosmos Hub, Osmosis 등 주요 체인 연결

Phase 2: 크로스체인 Activity Protocol
  → IBC 패킷으로 ActivityRecord 전달
  → 소스 체인 신뢰 모델 설계
```

---

## IBC Transfer (토큰 전송)

### KKOT 토큰의 IBC 전송

Cosmos SDK의 `ibc-transfer` 모듈(`x/ibc/applications/transfer`)을 표준 그대로 사용한다. KKOT은 IBC를 통해 다른 Cosmos 체인으로 전송될 수 있으며, 상대 체인에서는 IBC 바우처 토큰(`ibc/HASH...`)으로 표현된다.

```
KKOT 전송 플로우:

Seocheon                          Osmosis
┌──────────┐   IBC Transfer    ┌──────────┐
│ KKOT      │ ───────────────► │ ibc/ABC  │
│ (native) │                   │ (voucher)│
│          │ ◄─────────────── │          │
└──────────┘   IBC Transfer    └──────────┘
                (reverse)
```

### 화이트리스트 기반 채널 관리

무분별한 IBC 채널 개설을 방지하기 위해 **거버넌스 투표**를 통한 화이트리스트 방식을 채택한다.

| 항목 | 정책 |
|------|------|
| 채널 개설 | 거버넌스 제안 + 투표 승인 필요 |
| 채널 폐쇄 | 거버넌스 제안 또는 긴급 시 재단 제안 |
| 파라미터 변경 | 거버넌스 투표 |

**화이트리스트 거버넌스 제안 형식**:

```
IBC 채널 개설 제안:
  상대 체인: cosmoshub-4
  채널 용도: KKOT 토큰 전송 (ics20-1)
  릴레이어 운영자: [지정 또는 공개 모집]
  근거: Cosmos Hub와의 크로스체인 연결 확보
```

### IBC 토큰 수신 정책

Seocheon은 외부 토큰 수신에 대해 **보수적 정책**을 채택한다.

```
외부 토큰 수신 정책:

허용:
  → 화이트리스트 채널을 통한 IBC 토큰 수신
  → 수신된 토큰은 전송에 사용 가능
  → 거버넌스 투표로 허용 토큰 목록 관리

스테이킹/위임/노드 등록:
  → KKOT만 사용 가능 (IBC 토큰 미지원)
```

---

## 크로스체인 Activity Protocol

### 다른 Cosmos 체인에서의 Activity 제출

다른 Cosmos 체인에서 운영되는 AI 에이전트가 Seocheon에 활동을 기록할 수 있는 경로를 제공한다. 이는 Seocheon의 Activity Protocol을 체인 외부로 확장하는 핵심 기능이다.

```
크로스체인 Activity 제출 플로우:

Partner Chain                     Seocheon
┌──────────────┐                 ┌──────────────┐
│ AI Agent     │  IBC Packet    │ x/activity   │
│ (등록된 노드) │ ─────────────► │              │
│              │  ActivityData  │ ActivityRecord│
│              │  {             │ 생성 + 저장   │
│              │   hash,        │              │
│              │   content_uri, │              │
│              │   source_chain │              │
│              │  }             │              │
└──────────────┘                 └──────────────┘
```

### IBC 패킷을 통한 ActivityRecord 전달

크로스체인 Activity를 위해 **커스텀 IBC 애플리케이션 모듈**을 개발한다.

```protobuf
// IBC Activity 패킷 데이터
message ActivityPacketData {
  string submitter = 1;          // 제출자 (소스 체인의 에이전트 주소)
  bytes activity_hash = 2;       // 활동 해시
  string content_uri = 3;        // 오프체인 데이터 위치
  string source_chain_id = 4;    // 소스 체인 식별자
  string seocheon_node_id = 5;   // Seocheon에 등록된 노드 ID (매핑)
}
```

**처리 흐름**:

1. 파트너 체인의 에이전트가 IBC Activity 패킷을 전송
2. Seocheon의 IBC 모듈이 패킷을 수신
3. `seocheon_node_id`로 등록된 노드 확인 (x/node 쿼리)
4. 해당 노드의 크로스체인 Activity 제출 허용 여부 확인
5. ActivityRecord 생성 (`source_chain` 필드 추가)
6. ACK 패킷 반환

### 보안 고려사항 (소스 체인 신뢰성)

크로스체인 Activity의 핵심 리스크는 **소스 체인의 신뢰성**이다.

```
소스 체인 신뢰 모델:

Level 1 — 라이트 클라이언트 검증 (IBC 기본):
  → 소스 체인의 합의가 유효한가? (IBC 프로토콜이 보장)
  → 패킷이 실제로 소스 체인에서 발생했는가? (머클 증명)

Level 2 — 제출자 검증:
  → 소스 체인의 제출자가 Seocheon 노드와 매핑되어 있는가?
  → 매핑은 노드 운영자가 거버넌스/TX로 사전 등록

Level 3 — 쿼터 관리:
  → 크로스체인 Activity도 에포크 쿼터에 포함
  → 로컬 제출 + 크로스체인 제출 합산
  → 크로스체인 전용 추가 제한 가능 (거버넌스 파라미터)

```

**노드-체인 매핑 등록**:

```
노드 운영자가 크로스체인 Activity를 활성화하는 절차:

1. MsgRegisterCrossChainSource {
     operator: "seocheon1...",
     node_id: "node_123",
     source_chain_id: "cosmoshub-4",
     source_address: "cosmos1..."     // 파트너 체인의 에이전트 주소
   }

2. 등록된 매핑은 온체인에 저장
3. IBC Activity 수신 시 매핑 테이블 조회로 노드 식별
```

---

## IBC 채널 구성

### 릴레이어 운영 전략

```
릴레이어 운영 단계:

Phase A (초기 — 재단 주도):
  → 재단이 주요 채널의 릴레이어 직접 운영
  → Hermes 또는 Go Relayer 사용
  → 모니터링 + 알림 인프라 구축

Phase B (혼합):
  → 커뮤니티 릴레이어 운영자 모집
  → 릴레이어 운영 인센티브 (커뮤니티 풀 또는 그랜트)
  → 재단은 백업 릴레이어 유지

Phase C (탈중앙화):
  → 다수의 독립 릴레이어 운영
  → 재단 릴레이어는 최후 방어선
  → 릴레이어 상태 대시보드 공개
```

### 채널 파라미터

| 파라미터 | 값 | 근거 |
|---------|-----|------|
| Port | `transfer` (ics20), `activity` (커스텀) | 표준 + 커스텀 분리 |
| Ordering | `UNORDERED` (transfer), `ORDERED` (activity) | Activity는 순서 보장 필요 |
| Version | `ics20-1` (transfer), `activity-1` (커스텀) | 버전 관리 |
| Timeout Height | 상대 체인 블록 + 1,000 | ~1.5시간 여유 |
| Timeout Timestamp | 현재 + 10분 | 합리적 대기 시간 |

### 초기 연결 대상 체인 선정 기준

```
선정 기준 (우선순위):

1순위 — 크로스체인 인프라:
  → Osmosis: KKOT 크로스체인 전송 경로 확보
  → Cosmos Hub: Cosmos 생태계 중심 체인, 생태계 정통성

2순위 — AI 생태계 시너지:
  → AI 에이전트 관련 체인 (존재 시)
  → 데이터 마켓플레이스 체인

3순위 — 기술 파트너십:
  → CosmWasm 활성 체인 (컨트랙트 상호운용)
  → IBC 미들웨어 선도 체인

```

---

## 보안 고려사항

### 라이트 클라이언트 검증

IBC 보안의 핵심은 라이트 클라이언트이다. Seocheon은 연결된 각 체인의 라이트 클라이언트를 유지하며, 이를 통해 상대 체인의 상태를 검증한다.

```
라이트 클라이언트 보안 정책:

모니터링:
  → 라이트 클라이언트 만료 시간 추적 (trusting_period)
  → 만료 임박 시 자동 알림 → 릴레이어 업데이트 트리거
  → 재단이 라이트 클라이언트 상태 모니터링

Misbehaviour 대응:
  → CometBFT Misbehaviour 증거 수신 시 채널 동결
  → 거버넌스 투표로 동결 해제 또는 채널 폐쇄 결정
  → 동결 중 해당 채널의 모든 패킷 처리 중단

trusting_period 설정:
  → 상대 체인 unbonding_period의 2/3 이하
  → 예시: Cosmos Hub (21일 언본딩) → trusting_period ≤ 14일
```

### 채널 클로징 정책

```
채널 클로징 조건:

자동 클로징:
  → 라이트 클라이언트 Misbehaviour 감지
  → trusting_period 만료 (릴레이어 부재)

거버넌스 클로징:
  → 상대 체인의 보안 이슈 발생
  → 전략적 채널 정리
  → 사용량이 극히 낮은 채널 비용 최적화

긴급 클로징:
  → 재단의 긴급 거버넌스 제안
  → Circuit Breaker 연동 (활성화 시 IBC 채널 일시 중단 가능)
  → 복구 시 거버넌스 투표로 재개
```

### IBC 관련 거버넌스

| 거버넌스 항목 | 내용 |
|--------------|------|
| 채널 개설/폐쇄 | 화이트리스트 추가/제거 |
| 토큰 수신 정책 | 허용 토큰 목록 변경 |
| 크로스체인 Activity 활성화 | 파트너 체인별 Activity 채널 승인 |
| 릴레이어 인센티브 | 커뮤니티 풀에서 릴레이어 보상 |
| 긴급 채널 동결 | Misbehaviour 대응, Circuit Breaker 연동 |
| IBC 파라미터 변경 | 타임아웃, 수수료 정책 |

---

## 로드맵 단계

### Phase 1: IBC Transfer 기본 지원

**시기**: 메인넷 런칭 후 안정화 이후 (구현 로드맵 Phase 4 이후)

```
Phase 1 목표:
  → ibc-transfer 모듈 활성화 (Cosmos SDK 표준)
  → Cosmos Hub, Osmosis와 IBC 채널 개설
  → KKOT 크로스체인 전송 인프라 구축
  → 화이트리스트 거버넌스 프레임워크 구축
  → 릴레이어 인프라 구축 (재단 운영)

완료 조건:
  → KKOT 토큰의 크로스체인 전송 성공
  → IBC Transfer 성공률 ≥ 99.9%
  → 릴레이어 업타임 ≥ 99.5%
```

### Phase 2: 크로스체인 Activity

**시기**: Phase 1 안정화 + 파트너 체인 확보 후

```
Phase 2 목표:
  → 커스텀 IBC Activity 모듈 개발
  → ActivityPacketData 프로토콜 정의
  → 노드-체인 매핑 등록 TX 구현
  → 파일럿 파트너 체인과 Activity 채널 개설
  → 크로스체인 Activity 쿼터 관리 구현

완료 조건:
  → 파트너 체인에서 Seocheon으로 Activity 제출 성공
  → 크로스체인 Activity가 보상 분배에 정상 반영
  → 소스 체인 매핑의 거버넌스 관리 동작
```

