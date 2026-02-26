# 아키텍처 개요

> **담당**: SDK 아키텍트
> **관련 문서**: [인터페이스](02_interfaces.md) · [메서드](03_methods.md) · [상수](04_constants.md) · [통신](05_communication.md) · [전체 목차](README.md)

---

## 1.1 계층 구조

```
┌─────────────────────────────────────────────────────────────┐
│  Application Layer                                          │
│    AI 에이전트, 커스텀 앱, MCP 서버                            │
├─────────────────────────────────────────────────────────────┤
│  Seocheon Client SDK                                        │
│    5 모듈: activity, epoch, node, rewards, cosmos            │
│    인프라: SigningService, ChainClient                        │
├─────────────────────────────────────────────────────────────┤
│  Cosmos Client (언어별)                                      │
│    TypeScript: CosmJS · Go: Cosmos SDK client                │
│    Python: CosmPy · Kotlin/Java: CosmosJ                     │
│    Swift: CosmosSwift · C#: Cosmos.NET                       │
├─────────────────────────────────────────────────────────────┤
│  CometBFT RPC / Cosmos gRPC                                 │
└─────────────────────────────────────────────────────────────┘
```

## 1.2 5 모듈 구조

| 모듈 | 네임스페이스 | 함수 수 | 역할 |
|------|-------------|---------|------|
| `activity` | `sdk.activity` | 3 (TX 1, Query 2) | 활동 제출·조회·쿼터 |
| `epoch` | `sdk.epoch` | 2 (Query 2) | 에포크/윈도우 상태·자격 조회 |
| `node` | `sdk.node` | 2 (Query 2) | 노드 정보·검색 |
| `rewards` | `sdk.rewards` | 2 (TX 1, Query 1) | 보상 조회·인출 |
| `cosmos` | `sdk.cosmos` | 4 (TX 2, Query 2) | 잔고·전송·블록·TX 결과 |

**합계**: 13개 함수 (Query 9, TX 4)

## 1.3 설계 원칙

1. **단일 호출 완결**: SDK 함수 하나로 원하는 결과를 얻는다. TX 생성 → 서명 → 브로드캐스트를 내부에서 처리한다.
2. **키 미노출**: 에이전트 지갑 키가 SDK 함수의 입출력에 노출되지 않는다. 서명은 SDK 내부에서 처리한다.
3. **체인 철학 일관**: 체인이 검증하지 않는 것을 SDK도 검증하지 않는다. 형식 검증만 수행.
