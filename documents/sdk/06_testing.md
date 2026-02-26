# 테스트 시나리오

> **담당**: QA / SDK 개발자
> **관련 문서**: [메서드](03_methods.md) · [Mock 데이터](07_mock_data.md) · [상수](04_constants.md) · [전체 목차](README.md)

---

## 6.1 메서드별 테스트 케이스

### `activity.submit`

| # | 구분 | 입력 | 기대 결과 |
|---|------|------|----------|
| 1 | 정상 | `("a1b2c3...64chars", "ipfs://QmXyz...")` | tx_hash 반환, block_height > 0, quota_remaining = 9 |
| 2 | 정상 (자비) | `("d4e5f6...64chars", "https://example.com/report.json")` | quota_remaining = 99 (self-funded) |
| 3 | 에러 | `("invalid_hash", "ipfs://...")` | ERR_INVALID_ACTIVITY_HASH (1204) |
| 4 | 에러 | `("a1b2c3...64chars", "")` | ERR_INVALID_CONTENT_URI (1205) |
| 5 | 에러 | 쿼터 소진 후 호출 | ERR_QUOTA_EXCEEDED (1203) |
| 6 | 에러 | 동일 (hash, uri) 재제출 | ERR_DUPLICATE_ACTIVITY_HASH (1202) |
| 7 | 에러 | 미등록 agent 주소 | ERR_SUBMITTER_NOT_REGISTERED (1200) |
| 8 | 경계 | 에포크 마지막 블록에서 제출 | 정상, 해당 에포크의 마지막 윈도우 번호(11) 반환 |
| 9 | 경계 | 에포크 전환 직후 제출 | 정상, 새 에포크 번호, 윈도우 0, 쿼터 리셋 |

### `activity.getActivities`

| # | 구분 | 입력 | 기대 결과 |
|---|------|------|----------|
| 1 | 정상 | `(null, null)` | 자기 노드 현재 에포크 활동 목록 |
| 2 | 정상 | `("node-abc", 5)` | 지정 노드의 에포크 5 활동 목록 |
| 3 | 정상 | `("node-abc", 5)` 활동 없음 | activities: [], total_count: 0 |
| 4 | 에러 | `("nonexistent-node", null)` | NODE_NOT_FOUND |
| 5 | 경계 | 현재 에포크, 활동 100건 | 페이지네이션 정상 처리, total_count: 100 |

### `activity.getQuota`

| # | 구분 | 기대 결과 |
|---|------|----------|
| 1 | 정상 (feegrant) | quota_total: 10, is_feegrant: true, feegrant_expiry 존재 |
| 2 | 정상 (self-funded) | quota_total: 100, is_feegrant: false, feegrant_expiry: null |
| 3 | 정상 (쿼터 사용 후) | quota_used: 3, quota_remaining: 7 |
| 4 | 에러 | 미등록 에이전트 → NODE_NOT_FOUND |

### `epoch.getInfo`

| # | 구분 | 기대 결과 |
|---|------|----------|
| 1 | 정상 | 모든 11필드 정상 반환 |
| 2 | 경계 (에포크 시작) | block_height == epoch_start_block, window_number: 0, epoch_progress: "1/17280" |
| 3 | 경계 (윈도우 전환) | 직전: window_progress "1440/1440", 직후: 다음 윈도우, "1/1440" |
| 4 | 경계 (에포크 끝) | blocks_until_next_epoch: 0, blocks_until_next_window: 0 |

### `epoch.getQualification`

| # | 구분 | 입력 | 기대 결과 |
|---|------|------|----------|
| 1 | 정상 | `(null, null)` 8윈도우 활동 | is_qualified: true, remaining_needed: 0 |
| 2 | 정상 | `(null, null)` 5윈도우 활동, 7윈도우 경과 | is_qualified: false, can_still_qualify: true, remaining_needed: 3 |
| 3 | 정상 | `(null, null)` 3윈도우 활동, 8윈도우 경과 | is_qualified: false, can_still_qualify: false |
| 4 | 정상 | 과거 에포크 | elapsed_windows: 12, window_detail 12건 |
| 5 | 에러 | 미래 에포크 번호 | INVALID_EPOCH |

### `node.getInfo`

| # | 구분 | 입력 | 기대 결과 |
|---|------|------|----------|
| 1 | 정상 | `(null)` | 자기 노드 정보, total_delegation 포함 |
| 2 | 정상 | `("node-abc")` | 지정 노드 정보 |
| 3 | 에러 | `("nonexistent")` | ERR_NODE_NOT_FOUND (1101) |

### `node.search`

| # | 구분 | 입력 | 기대 결과 |
|---|------|------|----------|
| 1 | 정상 | `("ai", null, 10, "delegation")` | 태그 "ai" 노드, 위임량 내림차순 상위 10건 |
| 2 | 정상 | `(null, "ACTIVE", 20, null)` | ACTIVE 노드 전체, 위임량 정렬 |
| 3 | 정상 | `("nonexistent-tag")` | nodes: [], total_count: 0 |
| 4 | 정상 | `(null, null, 5, "registered_at")` | 등록일 내림차순 상위 5건 |

### `rewards.getPending`

| # | 구분 | 입력 | 기대 결과 |
|---|------|------|----------|
| 1 | 정상 | `(null)` | 보상 금액, 커미션 분배 포함 |
| 2 | 정상 | 보상 없음 | 모든 금액 "0.000000" |
| 3 | 에러 | `("nonexistent")` | NODE_NOT_FOUND |

### `rewards.withdraw`

| # | 구분 | 기대 결과 |
|---|------|----------|
| 1 | 정상 | tx_hash, withdrawn_total > 0, to_operator + to_agent == withdrawn_total |
| 2 | 에러 | agent 키로 서명 시 ERR_UNAUTHORIZED_OPERATOR (1108) |
| 3 | 에러 | 보상 0일 때 | 정상 처리, withdrawn_total: "0" |

### `cosmos.getBalance`

| # | 구분 | 입력 | 기대 결과 |
|---|------|------|----------|
| 1 | 정상 | `(null, null)` | 자기 주소 usum 잔고 |
| 2 | 정상 | `("seocheon1abc...", "usum")` | 지정 주소 잔고 |
| 3 | 경계 | 존재하지 않는 주소 | balance: "0", balance_kkot: "0.000000" |
| 4 | 정상 | 잔고 1,500,000 usum | balance: "1500000", balance_kkot: "1.500000" |

### `cosmos.sendTokens`

| # | 구분 | 입력 | 기대 결과 |
|---|------|------|----------|
| 1 | 정상 | `("seocheon1abc...", "1000000", "usum")` | tx_hash, block_height > 0 |
| 2 | 에러 | 잔고 부족 | INSUFFICIENT_FUNDS |
| 3 | 에러 | 잘못된 주소 | INVALID_ADDRESS |

### `cosmos.getBlockInfo`

| # | 구분 | 기대 결과 |
|---|------|----------|
| 1 | 정상 | block_height > 0, block_time ISO 8601, chain_id 일치 |

### `cosmos.getTxResult`

| # | 구분 | 입력 | 기대 결과 |
|---|------|------|----------|
| 1 | 정상 (성공 TX) | 유효 tx_hash | code: 0, events 존재 |
| 2 | 정상 (실패 TX) | 실패한 tx_hash | code > 0, raw_log에 에러 메시지 |
| 3 | 에러 | 존재하지 않는 해시 | TX_NOT_FOUND |

## 6.2 E2E 시나리오

### 시나리오 1: 노드 등록 → 활동 제출 → 자격 확인 → 보상 조회

```pseudocode
// Setup: operator 키로 SDK 초기화
sdk_op = new SeocheonSDK(operator_config)
sdk_op.connect()

// Step 1: 노드 정보 확인 (등록은 별도 TX)
node_info = sdk_op.node.getInfo()
ASSERT node_info.status == "REGISTERED" OR node_info.status == "ACTIVE"

// Step 2: agent 키로 SDK 전환
sdk_agent = new SeocheonSDK(agent_config)
sdk_agent.connect()

// Step 3: 쿼터 확인
quota = sdk_agent.activity.getQuota()
ASSERT quota.quota_remaining > 0

// Step 4: 에포크 정보 확인
epoch_info = sdk_agent.epoch.getInfo()
ASSERT epoch_info.epoch_number >= 0

// Step 5: 활동 제출
result = sdk_agent.activity.submit(
  "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
  "ipfs://QmTestContentHash123"
)
ASSERT result.tx_hash != ""
ASSERT result.block_height > 0
ASSERT result.quota_remaining == quota.quota_remaining - 1

// Step 6: 활동 이력 확인
activities = sdk_agent.activity.getActivities()
ASSERT activities.total_count >= 1
ASSERT activities.activities[0].activity_hash == "abcdef01..."

// Step 7: 자격 상태 확인
qual = sdk_agent.epoch.getQualification()
ASSERT qual.active_windows >= 1

// Step 8: 보상 조회
rewards = sdk_agent.rewards.getPending()
// 보상은 에포크 전환 후 발생하므로 첫 에포크에서는 0일 수 있음
ASSERT rewards.total_reward != null
```

### 시나리오 2: feegrant 노드 쿼터 소진 시나리오

```pseudocode
// feegrant 노드의 쿼터 테스트
sdk = new SeocheonSDK(feegrant_agent_config)
sdk.connect()

// 쿼터 확인
quota = sdk.activity.getQuota()
ASSERT quota.is_feegrant == true
ASSERT quota.quota_total == 10

// 10회 제출하여 쿼터 소진
FOR i = 0 TO 9:
  hash = sha256("test_content_" + i)
  result = sdk.activity.submit(hash, "ipfs://QmTest" + i)
  ASSERT result.quota_remaining == 10 - i - 1

// 11번째 시도 → 쿼터 초과
TRY:
  sdk.activity.submit(sha256("test_content_10"), "ipfs://QmTest10")
  FAIL("Expected QUOTA_EXCEEDED error")
CATCH SDKError as e:
  ASSERT e.code == 1203
```
