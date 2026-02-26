# Seocheon Client SDK 설계 스펙

> 이 디렉토리는 `sdk_specification.md`를 담당자별로 분리한 문서 모음이다.
> 각 파일은 독립적으로 읽을 수 있으며, 담당자가 개별적으로 수정할 수 있다.

---

## 문서 구조

| # | 파일 | 담당 | 내용 |
|---|------|------|------|
| 01 | [아키텍처 개요](01_architecture.md) | SDK 아키텍트 | 계층 구조, 5모듈 구조, 설계 원칙 |
| 02 | [클래스/인터페이스](02_interfaces.md) | SDK 개발자 | SeocheonSDK, SDKConfig, 모듈 클래스, SigningService, ChainClient, Response 구조체, Proto↔SDK 매핑 |
| 03 | [메서드 시그니처](03_methods.md) | SDK 개발자 | 13개 메서드 시그니처 (§3.1~3.13) |
| 04 | [전역 상수](04_constants.md) | SDK 개발자 / 체인 코어 | 체인 파라미터, 토큰 denomination, 에러 코드 상수, JSON Schema |
| 05 | [통신 스펙](05_communication.md) | SDK 개발자 / 인프라 | gRPC/REST 엔드포인트, TX 플로우, 폴링, Pagination |
| 06 | [테스트 시나리오](06_testing.md) | QA / SDK 개발자 | 메서드별 테스트 케이스, E2E 시나리오 |
| 07 | [Mock 데이터](07_mock_data.md) | QA / SDK 개발자 | Proto 응답 Mock, 경계값 Mock, TX 결과 Mock |
| 08 | [온체인 이벤트](08_events.md) | 체인 코어 / SDK 개발자 | x/node 이벤트, x/activity 이벤트 |

---

## 담당별 분류

### SDK 아키텍트
- [01_architecture.md](01_architecture.md) — 계층 구조와 설계 원칙

### SDK 개발자
- [02_interfaces.md](02_interfaces.md) — 클래스, 인터페이스, 데이터 타입
- [03_methods.md](03_methods.md) — 13개 메서드 시그니처 전체
- [04_constants.md](04_constants.md) — 전역 상수 및 설정 스키마
- [05_communication.md](05_communication.md) — 통신 스펙, TX 플로우

### QA / 테스트
- [06_testing.md](06_testing.md) — 메서드별 테스트 케이스, E2E 시나리오
- [07_mock_data.md](07_mock_data.md) — Mock 데이터

### 체인 코어
- [08_events.md](08_events.md) — 온체인 이벤트 타입

---

## 관련 문서

- [MCP 서버 아키텍처](../mcp_server_architecture.md) — SDK를 감싸는 MCP 서버 (seocheon-server)
- [핵심 개념](../blockchain/02_core_concepts.md) — 에포크, 윈도우, 보상 공식
- [Activity Protocol](../blockchain/04_activity_protocol.md) — MsgSubmitActivity 온체인 처리
- [노드 모듈](../blockchain/03_node_module.md) — x/node 모듈
- [스팸 방어](../blockchain/08_spam_defense.md) — 쿼터, 수수료, 게이밍 방어
- [아키텍처 리뷰](../architecture_review.md) — 보완 사항 및 미해소 항목
