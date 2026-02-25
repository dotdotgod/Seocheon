# Seocheon Chain 아키텍처

> 이 문서는 담당자별 독립 문서로 분리되었다. 아래 목차에서 해당 문서로 이동할 수 있다.
>
> **전체 목차**: [blockchain/README.md](blockchain/README.md)

---

## 문서 구조

| # | 파일 | 담당 | 내용 |
|---|------|------|------|
| 01 | [개요 및 설계 철학](blockchain/01_overview.md) | 아키텍트 / 프로젝트 리드 | 개요, 설계 철학 4원칙, 설계 결정 근거 |
| 02 | [핵심 개념](blockchain/02_core_concepts.md) | 체인 코어 개발자 | 에포크/윈도우, 이중 보상 풀, 노드 내 분배 |
| 03 | [x/node 모듈](blockchain/03_node_module.md) | x/node 모듈 개발자 (Go) | 노드 등록/관리, Registration Pool, x/staking 통합, 0원 참여 |
| 04 | [Activity Protocol](blockchain/04_activity_protocol.md) | x/activity 모듈 개발자 (Go) | 타임스탬핑, ActivityRecord, 오프체인 Report |
| 07 | [토큰 이코노믹스](blockchain/07_tokenomics.md) | 이코노미 설계자 | KKOT 토큰, Genesis 배분, 인플레이션, Genesis 1노드(Evangelist) |
| 08 | [스팸/게이밍 방어](blockchain/08_spam_defense.md) | 보안 / 메커니즘 설계 | 6대 방어 카테고리, 거버넌스 파라미터 종합 |
| 09 | [구현 가이드](blockchain/09_implementation.md) | 전체 팀 / PM | 거버넌스, API 레퍼런스, 모듈 구조, 로드맵, 테스트 전략 |
| 10 | [체인 업그레이드 전략](blockchain/10_chain_upgrade.md) | 체인 코어 / DevOps | x/upgrade, cosmovisor, 상태 마이그레이션, 긴급 업그레이드 |
| 11 | [Circuit Breaker](blockchain/11_circuit_breaker.md) | 체인 코어 / 보안 | 긴급 정지 메커니즘, 수동 트리거, 복구 절차 |
| 12 | [IBC 전략](blockchain/12_ibc_strategy.md) | 인프라 / 네트워크 | IBC Transfer (토큰 전송) |
| 13 | [인덱서 아키텍처](blockchain/13_indexer_architecture.md) | 인프라 / 백엔드 | 이벤트 인덱싱, content_uri 가용성, API/대시보드 |
