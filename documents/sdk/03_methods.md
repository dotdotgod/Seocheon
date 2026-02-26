# 메서드 시그니처

> **담당**: SDK 개발자
> **관련 문서**: [인터페이스](02_interfaces.md) · [상수](04_constants.md) · [통신](05_communication.md) · [테스트](06_testing.md) · [전체 목차](README.md)

---

## 3.1 `activity.submit(activity_hash, content_uri)`

활동 해시를 온체인에 타임스탬핑한다. 에이전트의 핵심 함수.

**시그니처**:
```
submit(
  activity_hash: string,              // 필수. SHA-256 hex, 64자
  content_uri: string                 // 필수. 오프체인 Activity Report 위치
) → SubmitActivityResponse
```

**Proto RPC 매핑**:
- TX: `seocheon.activity.v1.Msg/SubmitActivity`
- 메시지: `MsgSubmitActivity { submitter, activity_hash, content_uri }`
- 응답: `MsgSubmitActivityResponse { epoch, sequence }`

**파생 필드 계산**:
```pseudocode
// 1. TX 확정 후 block_height 획득
block_height = tx_result.height

// 2. 에포크/윈도우 번호 계산 (§3.3 공식 참조)
params = query_activity_params()
epoch_number = (block_height - 1) / params.epoch_length
window_length = params.epoch_length / params.windows_per_epoch
window_number = ((block_height - 1) % params.epoch_length) / window_length

// 3. 남은 쿼터 조회
node_epoch = query_node_epoch_activity(node_id, epoch_number)
quota_remaining = node_epoch.quota_limit - node_epoch.quota_used
```

**에러 코드**:

| 에러 | ABCI 코드 | 조건 |
|------|----------|------|
| `QUOTA_EXCEEDED` | 1203 | 에포크 쿼터 초과 |
| `DUPLICATE_HASH` | 1202 | 동일 (activity_hash, content_uri) 쌍 이미 존재 |
| `NODE_NOT_FOUND` | 1200 | submitter가 등록된 노드의 agent_address가 아님 |
| `NODE_NOT_ELIGIBLE` | 1201 | 노드 상태가 REGISTERED/ACTIVE가 아님 |
| `INVALID_HASH` | 1204 | activity_hash가 hex 64자가 아님 |
| `INVALID_URI` | 1205 | content_uri가 비어 있음 |
| `BROADCAST_FAILED` | — | TX 브로드캐스트 네트워크 오류 |
| `TIMEOUT` | — | TX 확정 대기 타임아웃 |

**처리 플로우**:
```pseudocode
function submit(activity_hash, content_uri):
  // 1. 형식 검증
  ASSERT len(activity_hash) == 64 AND is_hex(activity_hash)
  ASSERT len(content_uri) > 0

  // 2. 메시지 생성
  msg = MsgSubmitActivity {
    submitter: signing_service.getAddress(),
    activity_hash: activity_hash,
    content_uri: content_uri
  }

  // 3. TX 조립
  account = chain_client.getAccountInfo(msg.submitter)
  tx_bytes = build_tx(msg, account.sequence, account.account_number, config.chain.chain_id)

  // 4. 서명
  signed_tx = signing_service.sign(tx_bytes)

  // 5. 브로드캐스트
  broadcast_result = chain_client.broadcastTx(signed_tx, config.tx.broadcast_mode)
  ASSERT broadcast_result.code == 0, raise BROADCAST_FAILED

  // 6. TX 확정 대기
  tx_result = poll_tx_confirmation(broadcast_result.tx_hash, config.tx.confirm_timeout_ms)

  // 7. 파생 필드 계산
  params = query_activity_params()
  epoch_number = compute_epoch(tx_result.height, params.epoch_length)
  window_number = compute_window(tx_result.height, params.epoch_length, params.windows_per_epoch)
  node_epoch = query_node_epoch_activity(resolve_node_id(), epoch_number)

  RETURN SubmitActivityResponse {
    tx_hash: broadcast_result.tx_hash,
    block_height: tx_result.height,
    window_number: window_number,
    epoch_number: epoch_number,
    quota_remaining: node_epoch.quota_limit - node_epoch.quota_used
  }
```

---

## 3.2 `activity.getActivities(node_id?, epoch_number?)`

노드의 활동 제출 이력을 조회한다.

**시그니처**:
```
getActivities(
  node_id?: string,                   // 선택. 미지정 시 자기 노드
  epoch_number?: int64                // 선택. 미지정 시 현재 에포크
) → GetActivitiesResponse
```

**Proto RPC 매핑**:
- Query: `seocheon.activity.v1.Query/ActivitiesByNode`
- 요청: `QueryActivitiesByNodeRequest { node_id, epoch, pagination }`
- 응답: `QueryActivitiesByNodeResponse { activities[], pagination }`

**파생 필드 계산**:
```pseudocode
// node_id 해석
effective_node_id = node_id ?? resolve_own_node_id()

// epoch_number 해석
effective_epoch = epoch_number ?? compute_current_epoch()

// 각 ActivityRecord에 대해
FOR each record IN response.activities:
  params = get_cached_activity_params()
  window_length = params.epoch_length / params.windows_per_epoch
  record.window_number = ((record.block_height - 1) % params.epoch_length) / window_length

  // tx_hash는 Cosmos TX 인덱서에서 별도 조회
  // activity_hash + submitter로 TX 검색
  record.tx_hash = search_tx_by_events("activity_submitted", {"activity_hash": record.activity_hash})
```

**에러 코드**:

| 에러 | 조건 |
|------|------|
| `NODE_NOT_FOUND` | 지정된 node_id가 존재하지 않음 |
| `INVALID_EPOCH` | epoch_number가 음수이거나 현재 에포크보다 큼 |

**처리 플로우**:
```pseudocode
function getActivities(node_id, epoch_number):
  effective_node_id = node_id ?? resolve_own_node_id()
  effective_epoch = epoch_number ?? compute_current_epoch()

  // 페이지네이션으로 전체 조회
  all_activities = []
  pagination_key = null
  LOOP:
    response = chain_client.queryGrpc(
      "seocheon.activity.v1.Query", "ActivitiesByNode",
      { node_id: effective_node_id, epoch: effective_epoch, pagination: { key: pagination_key, limit: 100 } }
    )
    all_activities.append(response.activities)
    IF response.pagination.next_key == null: BREAK
    pagination_key = response.pagination.next_key

  // 파생 필드 계산
  params = get_cached_activity_params()
  FOR each item IN all_activities:
    item.window_number = compute_window(item.block_height, params.epoch_length, params.windows_per_epoch)
    item.tx_hash = search_tx_by_events("activity_submitted", {"activity_hash": item.activity_hash})

  RETURN GetActivitiesResponse {
    activities: all_activities,
    total_count: len(all_activities)
  }
```

---

## 3.3 `activity.getQuota()`

에포크 내 남은 활동 제출 쿼터를 확인한다.

**시그니처**:
```
getQuota() → GetQuotaResponse         // 파라미터 없음 (자기 노드 자동)
```

**Proto RPC 매핑**:
- Query 1: `seocheon.activity.v1.Query/NodeEpochActivity`
  - 요청: `QueryNodeEpochActivityRequest { node_id, epoch }`
  - 응답: `QueryNodeEpochActivityResponse { summary, quota_used, quota_limit }`
- Query 2 (파생): `cosmos.feegrant.v1beta1.Query/Allowance`
  - 요청: `QueryAllowanceRequest { granter: feegrant_pool, grantee: agent_address }`

**파생 필드 계산**:
```pseudocode
// feegrant 여부 판별
feegrant_allowance = query_feegrant_allowance(FEEGRANT_POOL_ADDRESS, agent_address)
is_feegrant = feegrant_allowance != null

// feegrant 만료 블록 계산
IF is_feegrant AND feegrant_allowance.expiration != null:
  feegrant_expiry = estimate_block_from_time(feegrant_allowance.expiration)
ELSE:
  feegrant_expiry = null
```

**에러 코드**:

| 에러 | 조건 |
|------|------|
| `NODE_NOT_FOUND` | 자기 agent_address로 등록된 노드가 없음 |

**처리 플로우**:
```pseudocode
function getQuota():
  node_id = resolve_own_node_id()
  epoch_number = compute_current_epoch()

  // 쿼터 정보 조회
  response = chain_client.queryGrpc(
    "seocheon.activity.v1.Query", "NodeEpochActivity",
    { node_id: node_id, epoch: epoch_number }
  )

  // feegrant 정보 조회
  agent_address = signing_service.getAddress()
  feegrant = query_feegrant_allowance(FEEGRANT_POOL_ADDRESS, agent_address)

  RETURN GetQuotaResponse {
    epoch_number: epoch_number,
    quota_total: response.quota_limit,
    quota_used: response.quota_used,
    quota_remaining: response.quota_limit - response.quota_used,
    is_feegrant: feegrant != null,
    feegrant_expiry: feegrant?.expiration != null ? estimate_block(feegrant.expiration) : null
  }
```

---

## 3.4 `epoch.getInfo()`

현재 에포크와 윈도우 상태를 조회한다. 에이전트가 활동 제출 타이밍을 판단하는 핵심 함수.

**시그니처**:
```
getInfo() → EpochInfoResponse          // 파라미터 없음
```

**Proto RPC 매핑**:
- Query: `seocheon.activity.v1.Query/EpochInfo`
  - 요청: `QueryEpochInfoRequest {}`
  - 응답: `QueryEpochInfoResponse { current_epoch, current_window, epoch_start_block, blocks_until_next_epoch }`
  - **proto 반환 4필드만**. 나머지 7필드는 SDK 계산.

**파생 필드 계산**:
```pseudocode
// proto 응답 4필드
proto = query_epoch_info()

// 파라미터 조회
params = get_cached_activity_params()
window_length = params.epoch_length / params.windows_per_epoch

// 현재 블록 높이
block_height = get_latest_block_height()

// 에포크 관련 파생
epoch_end_block = proto.epoch_start_block + params.epoch_length - 1
epoch_elapsed = block_height - proto.epoch_start_block + 1
epoch_progress = "{epoch_elapsed}/{params.epoch_length}"

// 윈도우 관련 파생
epoch_offset = (block_height - 1) % params.epoch_length
window_start_block = proto.epoch_start_block + (proto.current_window * window_length)
window_end_block = window_start_block + window_length - 1
window_elapsed = block_height - window_start_block + 1
window_progress = "{window_elapsed}/{window_length}"
blocks_until_next_window = window_end_block - block_height
```

**에포크/윈도우 핵심 공식** (모든 파생 필드의 기반):

```
epoch_length = 17,280 블록 (거버넌스 파라미터)
windows_per_epoch = 12 (거버넌스 파라미터)
window_length = epoch_length / windows_per_epoch = 1,440 블록

epoch_number = (block_height - 1) / epoch_length
window_index = ((block_height - 1) % epoch_length) / window_length

에포크 시작 블록 = epoch_number * epoch_length + 1
에포크 종료 블록 = (epoch_number + 1) * epoch_length
윈도우 시작 블록 = epoch_start + window_index * window_length
윈도우 종료 블록 = window_start + window_length - 1
```

**에러 코드**: 없음 (쿼리 전용, 항상 성공)

**처리 플로우**:
```pseudocode
function getInfo():
  proto = chain_client.queryGrpc("seocheon.activity.v1.Query", "EpochInfo", {})
  params = get_cached_activity_params()
  block_height = chain_client.getLatestBlock().height
  window_length = params.epoch_length / params.windows_per_epoch

  epoch_end_block = proto.epoch_start_block + params.epoch_length - 1
  window_start_block = proto.epoch_start_block + (proto.current_window * window_length)
  window_end_block = window_start_block + window_length - 1

  RETURN EpochInfoResponse {
    block_height: block_height,
    epoch_number: proto.current_epoch,
    epoch_start_block: proto.epoch_start_block,
    epoch_end_block: epoch_end_block,
    epoch_progress: format_progress(block_height - proto.epoch_start_block + 1, params.epoch_length),
    window_number: proto.current_window,
    window_start_block: window_start_block,
    window_end_block: window_end_block,
    window_progress: format_progress(block_height - window_start_block + 1, window_length),
    blocks_until_next_window: window_end_block - block_height,
    blocks_until_next_epoch: proto.blocks_until_next_epoch
  }
```

---

## 3.5 `epoch.getQualification(node_id?, epoch_number?)`

활동 보상 자격 상태를 조회한다.

**시그니처**:
```
getQualification(
  node_id?: string,                   // 선택. 미지정 시 자기 노드
  epoch_number?: int64                // 선택. 미지정 시 현재 에포크
) → QualificationResponse
```

**Proto RPC 매핑**:
- Query 1: `seocheon.activity.v1.Query/NodeEpochActivity`
  - 응답: `QueryNodeEpochActivityResponse { summary { total_activities, active_windows, eligible }, quota_used, quota_limit }`
- Query 2: `seocheon.activity.v1.Query/EpochInfo` (파생 필드 계산용)
- Query 3: `seocheon.activity.v1.Query/ActivitiesByNode` (window_detail 구성용)

**파생 필드 계산**:
```pseudocode
// proto 응답에서 기본 데이터
summary = node_epoch_response.summary

// 파라미터
params = get_cached_activity_params()
window_length = params.epoch_length / params.windows_per_epoch

// elapsed_windows 계산
epoch_info = query_epoch_info()
IF effective_epoch == epoch_info.current_epoch:
  elapsed_windows = epoch_info.current_window + 1   // 0-indexed → 1-indexed
ELSE:
  elapsed_windows = params.windows_per_epoch         // 과거 에포크는 12 고정

// remaining_needed
remaining_needed = max(0, params.min_active_windows - summary.active_windows)

// can_still_qualify
remaining_windows = params.windows_per_epoch - elapsed_windows
can_still_qualify = (summary.active_windows + remaining_windows) >= params.min_active_windows

// window_detail 구성
activities = query_activities_by_node(effective_node_id, effective_epoch)
window_counts = map<int64, uint64>{}                   // window_number → count
FOR each record IN activities:
  wn = compute_window(record.block_height, params.epoch_length, params.windows_per_epoch)
  window_counts[wn] = (window_counts[wn] ?? 0) + 1

window_detail = []
FOR w = 0 TO params.windows_per_epoch - 1:
  count = window_counts[w] ?? 0
  window_detail.append(WindowActivity { window_number: w, submission_count: count, has_activity: count > 0 })
```

**에러 코드**:

| 에러 | 조건 |
|------|------|
| `NODE_NOT_FOUND` | 지정된 node_id가 존재하지 않음 |
| `INVALID_EPOCH` | epoch_number가 음수이거나 현재 에포크보다 큼 |

**처리 플로우**:
```pseudocode
function getQualification(node_id, epoch_number):
  effective_node_id = node_id ?? resolve_own_node_id()
  effective_epoch = epoch_number ?? compute_current_epoch()
  params = get_cached_activity_params()

  // 기본 데이터
  node_epoch = chain_client.queryGrpc(
    "seocheon.activity.v1.Query", "NodeEpochActivity",
    { node_id: effective_node_id, epoch: effective_epoch }
  )

  // elapsed_windows
  epoch_info = chain_client.queryGrpc("seocheon.activity.v1.Query", "EpochInfo", {})
  elapsed_windows = effective_epoch == epoch_info.current_epoch
    ? epoch_info.current_window + 1
    : params.windows_per_epoch

  // window_detail
  activities = query_all_activities(effective_node_id, effective_epoch)
  window_detail = build_window_detail(activities, params)

  remaining_needed = max(0, params.min_active_windows - node_epoch.summary.active_windows)
  remaining_windows = params.windows_per_epoch - elapsed_windows
  can_still_qualify = (node_epoch.summary.active_windows + remaining_windows) >= params.min_active_windows

  RETURN QualificationResponse {
    epoch_number: effective_epoch,
    total_windows: params.windows_per_epoch,
    elapsed_windows: elapsed_windows,
    active_windows: node_epoch.summary.active_windows,
    required_windows: params.min_active_windows,
    is_qualified: node_epoch.summary.eligible,
    remaining_needed: remaining_needed,
    can_still_qualify: can_still_qualify,
    window_detail: window_detail
  }
```

---

## 3.6 `node.getInfo(node_id?)`

노드의 상세 정보를 조회한다.

**시그니처**:
```
getInfo(
  node_id?: string                    // 선택. 미지정 시 자기 노드
) → NodeInfoResponse
```

**Proto RPC 매핑**:
- Query 1: `seocheon.node.v1.Query/Node`
  - 요청: `QueryNodeRequest { node_id }`
  - 응답: `QueryNodeResponse { node }`
- Query 2 (파생): `cosmos.staking.v1beta1.Query/Validator`
  - 요청: `QueryValidatorRequest { validator_addr }`
- Query 3 (파생): `cosmos.staking.v1beta1.Query/DelegatorDelegations`

**파생 필드 계산**:
```pseudocode
// total_delegation: 밸리데이터의 총 위임량
validator = query_validator(node.validator_address)
total_delegation = validator.tokens  // usum → KKOT 변환

// self_delegation: operator 자기위임량
self_del = query_delegation(node.operator, node.validator_address)
self_delegation = self_del.balance.amount

// commission_rate: 밸리데이터 커미션율
commission_rate = validator.commission.commission_rates.rate

// status 문자열 변환
status_map = { 0: "UNSPECIFIED", 1: "REGISTERED", 2: "ACTIVE", 3: "INACTIVE", 4: "JAILED" }
status = status_map[node.status]
```

**에러 코드**:

| 에러 | ABCI 코드 | 조건 |
|------|----------|------|
| `NODE_NOT_FOUND` | 1101 | 지정된 node_id가 존재하지 않음 |

**처리 플로우**:
```pseudocode
function getInfo(node_id):
  effective_node_id = node_id ?? resolve_own_node_id()

  node = chain_client.queryGrpc("seocheon.node.v1.Query", "Node", { node_id: effective_node_id })

  // 스테이킹 정보 병합
  validator = query_validator(node.node.validator_address)
  self_del = query_delegation(node.node.operator, node.node.validator_address)

  RETURN NodeInfoResponse {
    node_id: node.node.id,
    operator: node.node.operator,
    agent_address: node.node.agent_address,
    status: enum_to_string(node.node.status),
    description: node.node.description,
    website: node.node.website,
    tags: node.node.tags,
    commission_rate: validator.commission.commission_rates.rate,
    agent_share: node.node.agent_share,
    total_delegation: format_kkot(validator.tokens),
    self_delegation: format_kkot(self_del.balance.amount),
    validator_address: node.node.validator_address,
    registered_at: node.node.registered_at
  }
```

---

## 3.7 `node.search(tag?, status?, limit?, order_by?)`

조건에 맞는 노드를 검색한다.

**시그니처**:
```
search(
  tag?: string,                       // 선택. 태그 필터
  status?: string,                    // 선택. 상태 필터 ("REGISTERED", "ACTIVE" 등)
  limit?: uint32,                     // 선택. 기본 20
  order_by?: string                   // 선택. 기본 "delegation". ("delegation" | "registered_at")
) → NodeSearchResponse
```

**Proto RPC 매핑**:
- tag 지정 시: `seocheon.node.v1.Query/NodesByTag`
  - 요청: `QueryNodesByTagRequest { tag, pagination }`
- tag 미지정 시: `seocheon.node.v1.Query/AllNodes`
  - 요청: `QueryAllNodesRequest { pagination }`

**파생 필드 및 SDK 레벨 처리**:
```pseudocode
// 1. proto 쿼리 (태그 여부에 따라 분기)
IF tag != null:
  nodes = paginate_all("seocheon.node.v1.Query/NodesByTag", { tag: tag })
ELSE:
  nodes = paginate_all("seocheon.node.v1.Query/AllNodes", {})

// 2. status 필터 (SDK 레벨)
IF status != null:
  target_status = string_to_enum(status)
  nodes = nodes.filter(n => n.status == target_status)

// 3. total_delegation 병합 (배치 조회)
FOR each node IN nodes:
  validator = query_validator(node.validator_address)
  node.total_delegation = validator.tokens

// 4. order_by 정렬 (SDK 레벨)
IF order_by == "delegation":
  nodes.sort_desc_by(n => n.total_delegation)
ELSE IF order_by == "registered_at":
  nodes.sort_desc_by(n => n.registered_at)

// 5. limit 적용
nodes = nodes.take(limit ?? 20)
```

**에러 코드**: 없음 (빈 결과 반환 가능)

**처리 플로우**:
```pseudocode
function search(tag, status, limit, order_by):
  effective_limit = limit ?? 20
  effective_order = order_by ?? "delegation"

  // proto 쿼리
  IF tag != null:
    raw_nodes = paginate_all("NodesByTag", { tag: tag })
  ELSE:
    raw_nodes = paginate_all("AllNodes", {})

  // SDK 레벨 필터링
  IF status != null:
    raw_nodes = raw_nodes.filter(n => enum_to_string(n.status) == status)

  // delegation 병합
  FOR each n IN raw_nodes:
    validator = query_validator_cached(n.validator_address)
    n._delegation = parse_int(validator.tokens)

  // 정렬
  IF effective_order == "delegation":
    raw_nodes.sort_desc_by(n => n._delegation)
  ELSE:
    raw_nodes.sort_desc_by(n => n.registered_at)

  // limit
  result = raw_nodes.take(effective_limit)

  RETURN NodeSearchResponse {
    nodes: result.map(n => NodeSummary {
      node_id: n.id, status: enum_to_string(n.status),
      tags: n.tags, total_delegation: format_kkot(n._delegation),
      description: n.description
    }),
    total_count: len(raw_nodes)    // 필터 후 전체 수
  }
```

---

## 3.8 `rewards.getPending(node_id?)`

미인출 보상을 조회한다.

**시그니처**:
```
getPending(
  node_id?: string                    // 선택. 미지정 시 자기 노드
) → PendingRewardsResponse
```

**Proto RPC 매핑**:
- Query 1: `cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards`
  - 요청: `QueryValidatorOutstandingRewardsRequest { validator_address }`
- Query 2: `cosmos.distribution.v1beta1.Query/ValidatorCommission`
  - 요청: `QueryValidatorCommissionRequest { validator_address }`
- Query 3: `seocheon.node.v1.Query/Node` (agent_share 조회)

**파생 필드 계산**:
```pseudocode
// 위임 보상
outstanding = query_validator_outstanding_rewards(validator_address)
delegation_reward = outstanding.rewards.filter(r => r.denom == "usum").amount

// 커미션 (위임 보상에서 분리)
commission = query_validator_commission(validator_address)
commission_total = commission.commission.filter(r => r.denom == "usum").amount

// 활동 보상: 현재 에포크의 활동 풀 분배 예상치
// (에포크 완료 전이므로 예상치)
activity_reward = estimate_activity_reward(node_id)

// 총 보상
total_reward = delegation_reward + activity_reward

// 커미션 분배
node = query_node(effective_node_id)
agent_share_ratio = parse_dec(node.agent_share) / 100
operator_share = commission_total * (1 - agent_share_ratio)
agent_share_amount = commission_total * agent_share_ratio
```

**에러 코드**:

| 에러 | 조건 |
|------|------|
| `NODE_NOT_FOUND` | 지정된 node_id가 존재하지 않음 |

**처리 플로우**:
```pseudocode
function getPending(node_id):
  effective_node_id = node_id ?? resolve_own_node_id()
  node = query_node(effective_node_id)

  // 위임 보상 + 커미션
  outstanding = query_outstanding_rewards(node.validator_address)
  commission = query_commission(node.validator_address)

  delegation_reward = sum_denom(outstanding.rewards, "usum")
  commission_total = sum_denom(commission.commission, "usum")
  activity_reward = estimate_activity_reward(effective_node_id)
  total_reward = delegation_reward + activity_reward

  agent_ratio = parse_dec(node.agent_share) / 100

  RETURN PendingRewardsResponse {
    delegation_reward: format_kkot(delegation_reward),
    activity_reward: format_kkot(activity_reward),
    total_reward: format_kkot(total_reward),
    commission_total: format_kkot(commission_total),
    operator_share: format_kkot(commission_total * (1 - agent_ratio)),
    agent_share: format_kkot(commission_total * agent_ratio)
  }
```

---

## 3.9 `rewards.withdraw()`

보상을 인출한다. (TX)

**시그니처**:
```
withdraw() → WithdrawRewardsResponse   // 파라미터 없음 (자기 노드 자동)
```

**Proto RPC 매핑**:
- TX: `seocheon.node.v1.Msg/WithdrawNodeCommission`
- 메시지: `MsgWithdrawNodeCommission { operator }`
- 응답: `MsgWithdrawNodeCommissionResponse { operator_amount, agent_amount }`

**에러 코드**:

| 에러 | ABCI 코드 | 조건 |
|------|----------|------|
| `UNAUTHORIZED_OPERATOR` | 1108 | 서명자가 노드의 operator가 아님 |
| `NODE_NOT_FOUND` | 1101 | operator에 연결된 노드가 없음 |
| `BROADCAST_FAILED` | — | TX 브로드캐스트 실패 |

**참고**: operator 서명이 필요한 TX이다. agent 키만 있는 경우 이 함수는 사용 불가.

**처리 플로우**:
```pseudocode
function withdraw():
  msg = MsgWithdrawNodeCommission {
    operator: signing_service.getAddress()
  }

  // TX 조립 → 서명 → 브로드캐스트 → 확정 대기 (공통 플로우)
  tx_result = execute_tx(msg)

  // 응답에서 금액 추출
  operator_amount = tx_result.msg_responses[0].operator_amount
  agent_amount = tx_result.msg_responses[0].agent_amount
  withdrawn_total = parse_int(operator_amount) + parse_int(agent_amount)

  RETURN WithdrawRewardsResponse {
    tx_hash: tx_result.tx_hash,
    withdrawn_total: format_kkot(withdrawn_total),
    to_operator: format_kkot(operator_amount),
    to_agent: format_kkot(agent_amount)
  }
```

---

## 3.10 `cosmos.getBalance(address?, denom?)`

토큰 잔고를 조회한다.

**시그니처**:
```
getBalance(
  address?: string,                   // 선택. 미지정 시 자기 agent 주소
  denom?: string                      // 선택. 기본 "usum"
) → BalanceResponse
```

**Proto RPC 매핑**:
- Query: `cosmos.bank.v1beta1.Query/Balance`
- 요청: `QueryBalanceRequest { address, denom }`
- 응답: `QueryBalanceResponse { balance { denom, amount } }`

**파생 필드 계산**:
```pseudocode
// usum → KKOT 변환
balance_kkot = format_decimal(balance_usum / 1_000_000, 6)
```

**에러 코드**: 없음 (존재하지 않는 주소는 잔고 0 반환)

**처리 플로우**:
```pseudocode
function getBalance(address, denom):
  effective_address = address ?? signing_service.getAddress()
  effective_denom = denom ?? "usum"

  response = chain_client.queryGrpc(
    "cosmos.bank.v1beta1.Query", "Balance",
    { address: effective_address, denom: effective_denom }
  )

  balance_usum = parse_int(response.balance.amount)

  RETURN BalanceResponse {
    address: effective_address,
    balance: response.balance.amount,
    balance_kkot: format_decimal(balance_usum, 6)  // ÷ 1,000,000
  }
```

---

## 3.11 `cosmos.sendTokens(to_address, amount, denom?)`

토큰을 전송한다. (TX)

**시그니처**:
```
sendTokens(
  to_address: string,                 // 필수. 수신자 주소
  amount: string,                     // 필수. 금액 (usum)
  denom?: string                      // 선택. 기본 "usum"
) → SendTokensResponse
```

**Proto RPC 매핑**:
- TX: `cosmos.bank.v1beta1.Msg/Send`
- 메시지: `MsgSend { from_address, to_address, amount[] }`

**에러 코드**:

| 에러 | 조건 |
|------|------|
| `INSUFFICIENT_FUNDS` | 잔고 부족 |
| `INVALID_ADDRESS` | 수신자 주소 형식 오류 |
| `UNAUTHORIZED_AGENT_MSG` | agent_address가 MsgSend 권한 없음 (agent_allowed_msg_types 미포함 시) |
| `BROADCAST_FAILED` | TX 브로드캐스트 실패 |

**처리 플로우**:
```pseudocode
function sendTokens(to_address, amount, denom):
  effective_denom = denom ?? "usum"

  msg = MsgSend {
    from_address: signing_service.getAddress(),
    to_address: to_address,
    amount: [{ denom: effective_denom, amount: amount }]
  }

  tx_result = execute_tx(msg)

  RETURN SendTokensResponse {
    tx_hash: tx_result.tx_hash,
    block_height: tx_result.height
  }
```

---

## 3.12 `cosmos.getBlockInfo()`

현재 블록 정보를 조회한다.

**시그니처**:
```
getBlockInfo() → BlockInfoResponse     // 파라미터 없음
```

**Proto RPC 매핑**:
- CometBFT RPC: `GET /block` (최신 블록)
- 또는 gRPC: `cosmos.base.tendermint.v1beta1.Service/GetLatestBlock`

**에러 코드**: 없음

**처리 플로우**:
```pseudocode
function getBlockInfo():
  block = chain_client.getLatestBlock()

  RETURN BlockInfoResponse {
    block_height: block.header.height,
    block_time: block.header.time,     // ISO 8601
    chain_id: block.header.chain_id,
    num_txs: len(block.data.txs)
  }
```

---

## 3.13 `cosmos.getTxResult(tx_hash)`

트랜잭션 실행 결과를 조회한다.

**시그니처**:
```
getTxResult(
  tx_hash: string                     // 필수. TX 해시
) → TxResultResponse
```

**Proto RPC 매핑**:
- CometBFT RPC: `GET /tx?hash=0x{tx_hash}`
- 또는 gRPC: `cosmos.tx.v1beta1.Service/GetTx`

**에러 코드**:

| 에러 | 조건 |
|------|------|
| `TX_NOT_FOUND` | 해당 해시의 TX가 없음 (미확정 또는 존재하지 않음) |

**처리 플로우**:
```pseudocode
function getTxResult(tx_hash):
  tx = chain_client.getTx(tx_hash)
  IF tx == null: raise TX_NOT_FOUND

  RETURN TxResultResponse {
    tx_hash: tx_hash,
    height: tx.height,
    code: tx.tx_result.code,
    gas_used: tx.tx_result.gas_used,
    gas_wanted: tx.tx_result.gas_wanted,
    raw_log: tx.tx_result.log,
    events: tx.tx_result.events.map(e => TxEvent {
      type: e.type,
      attributes: e.attributes.map(a => EventAttribute { key: a.key, value: a.value })
    })
  }
```
