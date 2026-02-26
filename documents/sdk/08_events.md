# 부록: 온체인 이벤트 타입

> **담당**: 체인 코어 / SDK 개발자
> **관련 문서**: [메서드](03_methods.md) · [Mock 데이터](07_mock_data.md) · [전체 목차](README.md)

---

SDK는 TX 결과의 이벤트를 파싱하여 파생 필드를 추출한다.

## x/node 이벤트

| 이벤트 타입 | 속성 키 | 발생 시점 |
|------------|---------|----------|
| `node_registered` | `node_id`, `operator`, `agent_address`, `validator_address` | MsgRegisterNode 성공 |
| `node_updated` | `node_id`, `operator` | MsgUpdateNode 성공 |
| `node_deactivated` | `node_id`, `operator` | MsgDeactivateNode 성공 |
| `node_activated` | `node_id`, `validator_address` | Active Set 진입 |
| `agent_address_changed` | `node_id`, `old_agent_address`, `new_agent_address` | MsgUpdateAgentAddress 성공 |
| `agent_share_change_scheduled` | `node_id`, `new_agent_share`, `apply_at_block` | MsgUpdateNodeAgentShare 성공 |
| `agent_share_changed` | `node_id`, `new_agent_share` | 에포크 전환 시 적용 |
| `node_commission_withdrawn` | `node_id`, `operator_amount`, `agent_amount` | MsgWithdrawNodeCommission 성공 |
| `feegrant_grant_failed` | `node_id`, `error` | 자동 feegrant 부여 실패 |
| `node_jailed` | `node_id` | 합의 위반으로 Jailing |
| `undelegate_failed` | `node_id`, `error` | 비활성화 시 언델리게이션 실패 |

## x/activity 이벤트

| 이벤트 타입 | 속성 키 | 발생 시점 |
|------------|---------|----------|
| `activity_submitted` | `node_id`, `submitter`, `activity_hash`, `content_uri`, `epoch`, `window`, `sequence`, `block_height` | MsgSubmitActivity 성공 |
| `activity_pruned` | `pruned_count` | 에포크 전환 EndBlocker |
| `window_completed` | `epoch`, `window` | 윈도우 경계 블록 |
| `epoch_completed` | `epoch`, `eligible_count`, `active_nodes` | 에포크 전환 |
| `epoch_fee_state` | `epoch` | 에포크 전환 시 수수료 상태 캐싱 |
| `fees_collected` | `epoch` | 에포크 전환 시 수수료 정산 |
| `activity_fee_charged` | `node_id`, `submitter`, `quota_type` | 활동 수수료 부과 |
