# Seocheon 아키텍처 리뷰

> 대상 문서: `documents/blockchain/` (12개 문서), `documents/agent_architecture.md`, `documents/mcp_server_architecture.md`, `documents/foundation_strategy.md`

아키텍처 설계는 완료되었다. 아래는 문서 리뷰에서 발견된 보완 사항과 구현 단계 결정 사항이다.

---

## 문서 보완 사항

### 중요도: 하

#### ~~1. Feegrant 자동 부여 메커니즘~~ → ✅ 해소

MsgRegisterNode 핸들러 내부에서 `feegrantKeeper.GrantAllowance()`를 원자적으로 호출. 등록 TX 가스비는 RegistrationFeeDecorator가 `node_feegrant_pool`에서 차감. 상세: [03_node_module.md](blockchain/03_node_module.md) §Feegrant Pool, §RegistrationFeeDecorator, §MsgRegisterNode 통합 플로우 [5]단계.

---

## 구현 시 결정 사항

아키텍처 설계에서 의도적으로 열어둔 항목들. 구현 단계에서 결정하면 충분하다.

### ~~Agent Wallet 권한 세분화~~ → ✅ 해소

AgentPermissionDecorator(AnteHandler)로 agent_address의 메시지 타입을 화이트리스트 제한. 허용: MsgSubmitActivity, MsgSend. 그 외 전부 거부. MsgUpdateAgentAddress로 키 교체 + 비활성화 지원 (쿨다운 1 에포크). 상세: [03_node_module.md](blockchain/03_node_module.md) §AgentPermissionDecorator, §MsgUpdateAgentAddress.

### Vault 구현 상세

Vault는 off-chain 참고 아키텍처로, 노드 운영자의 선택 사항이다. KEK/DEK 키 관리 계층, HSM/TEE 활용, 백업/복구 절차는 레퍼런스 구현 시 가이드라인으로 제공하면 충분하다.

### MaxValidators 설정

CometBFT의 O(n²) 통신 복잡도를 고려하면 Active Validator Set은 150~200이 적절하다. REGISTERED 노드는 무제한 참여 가능하므로 Active Set 크기가 네트워크 규모를 제한하지 않는다.

---

## 구현 로드맵 참고 사항

아키텍처 문서에는 언급되지 않았으나, 구현 단계에서 필요한 기술적 항목들이다.

| 항목 | 중요도 | 비고 |
|------|--------|------|
| 체인 업그레이드 전략 | 상 | ✅ 문서화 완료 ([10_chain_upgrade.md](blockchain/10_chain_upgrade.md)) |
| 긴급 정지 메커니즘 (Circuit Breaker) | 상 | ✅ 문서화 완료 ([11_circuit_breaker.md](blockchain/11_circuit_breaker.md)) |
| IBC 전략 | 중 | ✅ 문서화 완료 ([12_ibc_strategy.md](blockchain/12_ibc_strategy.md)) |
| API rate limiting | 중 | ✅ 인덱서 아키텍처 문서에 포함 ([13_indexer_architecture.md](blockchain/13_indexer_architecture.md) §Rate Limiting 전략) |
| 인덱서 상세 아키텍처 | 중 | ✅ 문서화 완료 ([13_indexer_architecture.md](blockchain/13_indexer_architecture.md)) |
| 클라이언트 SDK 설계 | 중 | 생태계 개발 전략 2순위 (그랜트 프로그램) |
| 스케일링 전략 (IBC 샤딩, L2) | 중 | 노드 수 10만+ 시 TPS 병목 대응, 장기 과제 |
| 에이전트 기억 체계 구현 상세 | 하 | off-chain 참고 아키텍처, 운영자 자유 |
