# Seocheon Chain 아키텍처

> 이 디렉토리는 `blockchain_architecture.md`를 담당자별로 분리한 문서 모음이다.
> 각 파일은 독립적으로 읽을 수 있으며, 담당자가 개별적으로 수정할 수 있다.

---

## 문서 구조

| # | 파일 | 담당 | 내용 |
|---|------|------|------|
| 01 | [개요 및 설계 철학](01_overview.md) | 아키텍트 / 프로젝트 리드 | 개요, 설계 철학 4원칙, 설계 결정 근거 |
| 02 | [핵심 개념](02_core_concepts.md) | 체인 코어 개발자 | 에포크/윈도우, 이중 보상 풀, 노드 내 분배 |
| 03 | [x/node 모듈](03_node_module.md) | x/node 모듈 개발자 (Go) | 노드 등록/관리, Registration Pool, x/staking 통합, 0원 참여 |
| 04 | [Activity Protocol](04_activity_protocol.md) | x/activity 모듈 개발자 (Go) | 타임스탬핑, ActivityRecord, 오프체인 Report |
| 07 | [토큰 이코노믹스](07_tokenomics.md) | 이코노미 설계자 | KKOT 토큰, Genesis 배분, 인플레이션, Genesis 1노드(Evangelist) |
| 08 | [스팸/게이밍 방어](08_spam_defense.md) | 보안 / 메커니즘 설계 | 6대 방어 카테고리, 거버넌스 파라미터 종합 |
| 09 | [구현 가이드](09_implementation.md) | 전체 팀 / PM | 거버넌스, API 레퍼런스, 모듈 구조, 로드맵, 테스트 전략 |
| 10 | [체인 업그레이드 전략](10_chain_upgrade.md) | 체인 코어 / DevOps | x/upgrade, cosmovisor, 상태 마이그레이션, 긴급 업그레이드 |
| 11 | [Circuit Breaker](11_circuit_breaker.md) | 체인 코어 / 보안 | 긴급 정지 메커니즘, 수동 트리거 (Guardian), 복구 절차 |
| 12 | [IBC 전략](12_ibc_strategy.md) | 인프라 / 네트워크 | IBC Transfer, KKOT 크로스체인 전송 |
| 13 | [인덱서 아키텍처](13_indexer_architecture.md) | 인프라 / 백엔드 | 이벤트 인덱싱, content_uri 가용성, API/대시보드 |

---

## 담당별 분류

### 아키텍트 / 프로젝트 리드
- [01_overview.md](01_overview.md) — 설계 철학과 핵심 원칙

### 체인 코어 (Go)
- [02_core_concepts.md](02_core_concepts.md) — 에포크, 윈도우, 보상 공식
- [03_node_module.md](03_node_module.md) — x/node 모듈, x/staking 연동
- [04_activity_protocol.md](04_activity_protocol.md) — x/activity 모듈

### 이코노미 / 보안
- [07_tokenomics.md](07_tokenomics.md) — 토큰 이코노믹스
- [08_spam_defense.md](08_spam_defense.md) — 스팸/게이밍 방어

### 체인 코어 / DevOps / 보안
- [10_chain_upgrade.md](10_chain_upgrade.md) — 체인 업그레이드 전략
- [11_circuit_breaker.md](11_circuit_breaker.md) — 긴급 정지 메커니즘

### 인프라 / 네트워크 / 백엔드
- [12_ibc_strategy.md](12_ibc_strategy.md) — IBC 전략
- [13_indexer_architecture.md](13_indexer_architecture.md) — 인덱서 아키텍처

### 전체 팀
- [09_implementation.md](09_implementation.md) — 구현 로드맵, API, 테스트

---

## 관련 문서

- [에이전트 아키텍처](../agent_architecture.md) — 오프체인 에이전트 참고 아키텍처
- [재단 전략](../foundation_strategy.md) — 재단 운영, 탈중앙화 전환
- [아키텍처 리뷰](../architecture_review.md) — 보완 사항 및 미해소 항목

---

## 모듈 의존성

```
x/node
├── 의존 ───► x/staking, x/bank, x/distribution, x/slashing
└── x/activity 가 의존 ◄──

x/circuit (Cosmos SDK 표준)
└── Guardian 권한 관리

x/upgrade
└── 거버넌스(x/gov)를 통한 업그레이드 플랜 관리

ibc-transfer
└── IBC Core (라이트 클라이언트, 채널 관리)

인덱서 (오프체인)
├── 구독 ◄── x/node, x/activity 이벤트
└── content_uri 가용성 모니터링
```
