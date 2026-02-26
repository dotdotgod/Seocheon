# 통신 스펙

> **담당**: SDK 개발자 / 인프라
> **관련 문서**: [인터페이스](02_interfaces.md) · [메서드](03_methods.md) · [상수](04_constants.md) · [전체 목차](README.md)

---

## 5.1 gRPC/REST 쿼리 엔드포인트

### Seocheon 커스텀 엔드포인트

| 서비스 | RPC | gRPC Path | REST Path | HTTP |
|--------|-----|-----------|-----------|------|
| `activity.Query` | `Params` | `seocheon.activity.v1.Query/Params` | `/seocheon/activity/v1/params` | GET |
| `activity.Query` | `Activity` | `seocheon.activity.v1.Query/Activity` | `/seocheon/activity/v1/activities/{activity_hash}` | GET |
| `activity.Query` | `ActivitiesByNode` | `seocheon.activity.v1.Query/ActivitiesByNode` | `/seocheon/activity/v1/nodes/{node_id}/activities` | GET |
| `activity.Query` | `ActivitiesByBlock` | `seocheon.activity.v1.Query/ActivitiesByBlock` | `/seocheon/activity/v1/blocks/{block_height}/activities` | GET |
| `activity.Query` | `EpochInfo` | `seocheon.activity.v1.Query/EpochInfo` | `/seocheon/activity/v1/epoch-info` | GET |
| `activity.Query` | `NodeEpochActivity` | `seocheon.activity.v1.Query/NodeEpochActivity` | `/seocheon/activity/v1/nodes/{node_id}/epochs/{epoch}` | GET |
| `node.Query` | `Params` | `seocheon.node.v1.Query/Params` | `/seocheon/node/v1/params` | GET |
| `node.Query` | `Node` | `seocheon.node.v1.Query/Node` | `/seocheon/node/v1/nodes/{node_id}` | GET |
| `node.Query` | `NodeByOperator` | `seocheon.node.v1.Query/NodeByOperator` | `/seocheon/node/v1/nodes/by-operator/{operator}` | GET |
| `node.Query` | `NodeByAgentAddress` | `seocheon.node.v1.Query/NodeByAgentAddress` | `/seocheon/node/v1/nodes/by-agent/{agent_address}` | GET |
| `node.Query` | `NodesByTag` | `seocheon.node.v1.Query/NodesByTag` | `/seocheon/node/v1/nodes/by-tag/{tag}` | GET |
| `node.Query` | `AllNodes` | `seocheon.node.v1.Query/AllNodes` | `/seocheon/node/v1/nodes` | GET |

### Cosmos 표준 엔드포인트 (SDK 사용분)

| 용도 | REST Path | 사용 함수 |
|------|-----------|----------|
| 잔고 조회 | `/cosmos/bank/v1beta1/balances/{address}/by_denom?denom={denom}` | `cosmos.getBalance` |
| 밸리데이터 조회 | `/cosmos/staking/v1beta1/validators/{validator_addr}` | `node.getInfo`, `node.search` |
| 위임 조회 | `/cosmos/staking/v1beta1/validators/{validator_addr}/delegations/{delegator_addr}` | `node.getInfo` |
| 보상 조회 | `/cosmos/distribution/v1beta1/validators/{validator_addr}/outstanding_rewards` | `rewards.getPending` |
| 커미션 조회 | `/cosmos/distribution/v1beta1/validators/{validator_addr}/commission` | `rewards.getPending` |
| feegrant 조회 | `/cosmos/feegrant/v1beta1/allowance/{granter}/{grantee}` | `activity.getQuota` |
| 계정 정보 | `/cosmos/auth/v1beta1/accounts/{address}` | TX 시퀀스 조회 |
| 최신 블록 | `/cosmos/base/tendermint/v1beta1/blocks/latest` | `cosmos.getBlockInfo`, `epoch.getInfo` |

### TX 엔드포인트

| 서비스 | RPC | 사용 함수 |
|--------|-----|----------|
| `seocheon.activity.v1.Msg` | `SubmitActivity` | `activity.submit` |
| `seocheon.node.v1.Msg` | `WithdrawNodeCommission` | `rewards.withdraw` |
| `cosmos.bank.v1beta1.Msg` | `Send` | `cosmos.sendTokens` |

## 5.2 TX 브로드캐스트 플로우

모든 TX 함수(`activity.submit`, `rewards.withdraw`, `cosmos.sendTokens`)가 공유하는 내부 처리 플로우.

```pseudocode
function execute_tx(msg):
  // Phase 1: TX 조립
  signer_address = signing_service.getAddress()
  account = chain_client.getAccountInfo(signer_address)
    // → { account_number, sequence }

  tx_body = TxBody {
    messages: [Any.pack(msg)],
    memo: "",
    timeout_height: 0,
    extension_options: []
  }

  auth_info = AuthInfo {
    signer_infos: [{
      public_key: Any.pack(signing_service.getPubKey()),
      mode_info: { single: { mode: SIGN_MODE_DIRECT } },
      sequence: account.sequence
    }],
    fee: Fee {
      amount: [{ denom: "usum", amount: estimate_fee(msg) }],
      gas_limit: estimate_gas(msg) * config.chain.gas_adjustment,
      payer: "",
      granter: ""  // feegrant 사용 시 granter 주소 설정
    }
  }

  sign_doc = SignDoc {
    body_bytes: encode(tx_body),
    auth_info_bytes: encode(auth_info),
    chain_id: config.chain.chain_id,
    account_number: account.account_number
  }

  // Phase 2: 서명
  signature = signing_service.sign(encode(sign_doc))

  tx_raw = TxRaw {
    body_bytes: sign_doc.body_bytes,
    auth_info_bytes: sign_doc.auth_info_bytes,
    signatures: [signature]
  }

  // Phase 3: 브로드캐스트
  broadcast_response = chain_client.broadcastTx(encode(tx_raw), config.tx.broadcast_mode)
  IF broadcast_response.code != 0:
    raise SDKError(ERR_BROADCAST_FAILED, broadcast_response.raw_log)

  // Phase 4: TX 확정 대기
  tx_result = poll_tx_confirmation(broadcast_response.tx_hash)

  RETURN tx_result
```

## 5.3 TX 확정 대기 폴링 로직

```pseudocode
function poll_tx_confirmation(tx_hash):
  deadline = now() + config.tx.confirm_timeout_ms
  interval = config.tx.confirm_poll_interval_ms

  LOOP:
    IF now() > deadline:
      raise SDKError(ERR_TX_TIMEOUT, "TX confirmation timeout: " + tx_hash)

    TRY:
      result = chain_client.getTx(tx_hash)
      IF result != null:
        IF result.tx_result.code != 0:
          raise SDKError(result.tx_result.code, result.tx_result.log)
        RETURN result
    CATCH NotFoundError:
      // TX 아직 미확정, 계속 폴링
      PASS

    sleep(interval)
```

## 5.4 Pagination 처리

Proto `cosmos.base.query.v1beta1.PageRequest/PageResponse`를 사용하는 모든 쿼리에 적용.

```pseudocode
function paginate_all(query_path, params):
  all_items = []
  next_key = null

  LOOP:
    request = params + {
      pagination: {
        key: next_key,
        limit: 100,        // 페이지당 최대 100
        count_total: (next_key == null)  // 첫 요청에서만 total 카운트
      }
    }

    response = chain_client.queryGrpc(query_path, request)
    all_items.extend(response.items)

    next_key = response.pagination.next_key
    IF next_key == null OR len(next_key) == 0:
      BREAK

  RETURN all_items
```
