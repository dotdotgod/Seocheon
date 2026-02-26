# Seocheon Client SDK 설계 스펙

> 이 문서는 담당자별 독립 문서로 분리되었다. 아래 목차에서 해당 문서로 이동할 수 있다.
>
> **전체 목차**: [sdk/README.md](sdk/README.md)

> **목적**: 7개 언어(TypeScript, Go, Python, Kotlin, Java, Swift, C#) 구현의 **단일 진실 원천(Single Source of Truth)**
> **관련 문서**: [MCP 서버 아키텍처](mcp_server_architecture.md) · [핵심 개념](blockchain/02_core_concepts.md) · [Activity Protocol](blockchain/04_activity_protocol.md) · [노드 모듈](blockchain/03_node_module.md) · [스팸 방어](blockchain/08_spam_defense.md)

> **면책 조항**: 이 문서는 프로토콜 메커니즘에 대한 기술 설계 문서이다. 보상 분배는 프로토콜 규칙에 따라 자동 실행되며, 어떠한 형태의 수익도 보장하지 않는다.

---

## 문서 구조

| # | 파일 | 담당 | 내용 |
|---|------|------|------|
| 01 | [아키텍처 개요](sdk/01_architecture.md) | SDK 아키텍트 | 계층 구조, 5모듈 구조, 설계 원칙 |
| 02 | [클래스/인터페이스](sdk/02_interfaces.md) | SDK 개발자 | SeocheonSDK, SDKConfig, 모듈 클래스, SigningService, ChainClient, Response 구조체, Proto↔SDK 매핑 |
| 03 | [메서드 시그니처](sdk/03_methods.md) | SDK 개발자 | 13개 메서드 시그니처 (§3.1~3.13) |
| 04 | [전역 상수](sdk/04_constants.md) | SDK 개발자 / 체인 코어 | 체인 파라미터, 토큰 denomination, 에러 코드 상수, JSON Schema |
| 05 | [통신 스펙](sdk/05_communication.md) | SDK 개발자 / 인프라 | gRPC/REST 엔드포인트, TX 플로우, 폴링, Pagination |
| 06 | [테스트 시나리오](sdk/06_testing.md) | QA / SDK 개발자 | 메서드별 테스트 케이스, E2E 시나리오 |
| 07 | [Mock 데이터](sdk/07_mock_data.md) | QA / SDK 개발자 | Proto 응답 Mock, 경계값 Mock, TX 결과 Mock |
| 08 | [온체인 이벤트](sdk/08_events.md) | 체인 코어 / SDK 개발자 | x/node 이벤트, x/activity 이벤트 |
