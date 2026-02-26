# Mock 데이터

> **담당**: QA / SDK 개발자
> **관련 문서**: [메서드](03_methods.md) · [테스트](06_testing.md) · [상수](04_constants.md) · [전체 목차](README.md)

---

## 7.1 Proto 쿼리 응답 Mock

### `QueryEpochInfoResponse`

```json
{
  "current_epoch": 42,
  "current_window": 7,
  "epoch_start_block": 725761,
  "blocks_until_next_epoch": 6040
}
```

SDK 파생 후 `EpochInfoResponse`:

```json
{
  "block_height": 737000,
  "epoch_number": 42,
  "epoch_start_block": 725761,
  "epoch_end_block": 743040,
  "epoch_progress": "11240/17280",
  "window_number": 7,
  "window_start_block": 735841,
  "window_end_block": 737280,
  "window_progress": "1160/1440",
  "blocks_until_next_window": 280,
  "blocks_until_next_epoch": 6040
}
```

### `QueryNodeEpochActivityResponse`

```json
{
  "summary": {
    "total_activities": 12,
    "active_windows": 6,
    "eligible": false
  },
  "quota_used": 12,
  "quota_limit": 100
}
```

### `QueryNodeResponse`

```json
{
  "node": {
    "id": "node-seocheon1abc123",
    "operator": "seocheon1operator456",
    "agent_address": "seocheon1agent789",
    "agent_share": "30.000000000000000000",
    "max_agent_share_change_rate": "5.000000000000000000",
    "description": "AI-powered content curation agent",
    "website": "https://example.com",
    "tags": ["ai", "curation", "content"],
    "validator_address": "seocheonvaloper1val012",
    "status": 2,
    "registered_at": 17281
  }
}
```

### `QueryActivitiesByNodeResponse`

```json
{
  "activities": [
    {
      "node_id": "node-seocheon1abc123",
      "epoch": 42,
      "sequence": 1,
      "submitter": "seocheon1agent789",
      "activity_hash": "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90",
      "content_uri": "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG",
      "block_height": 726500
    },
    {
      "node_id": "node-seocheon1abc123",
      "epoch": 42,
      "sequence": 2,
      "submitter": "seocheon1agent789",
      "activity_hash": "b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90a1",
      "content_uri": "ipfs://QmZwBPKzv5DZtnB726t4Yg3nemtZhQqIeXFz80pkXoQcea",
      "block_height": 728200
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

### `cosmos.bank.v1beta1.QueryBalanceResponse`

```json
{
  "balance": {
    "denom": "usum",
    "amount": "1500000000"
  }
}
```

### `cosmos.feegrant.v1beta1.QueryAllowanceResponse`

```json
{
  "allowance": {
    "@type": "/cosmos.feegrant.v1beta1.AllowedMsgAllowance",
    "allowed_messages": [
      "/seocheon.activity.v1.MsgSubmitActivity"
    ],
    "allowance": {
      "@type": "/cosmos.feegrant.v1beta1.PeriodicAllowance",
      "basic": {
        "spend_limit": [],
        "expiration": "2026-08-10T00:00:00Z"
      },
      "period": "86400s",
      "period_spend_limit": [{ "denom": "usum", "amount": "1000000" }],
      "period_can_spend": [{ "denom": "usum", "amount": "800000" }],
      "period_reset": "2026-02-27T00:00:00Z"
    }
  }
}
```

## 7.2 에포크/윈도우 경계값 Mock

### 에포크 0 시작 (블록 1)

```json
{
  "block_height": 1,
  "epoch_number": 0,
  "epoch_start_block": 1,
  "epoch_end_block": 17280,
  "window_number": 0,
  "window_start_block": 1,
  "window_end_block": 1440,
  "blocks_until_next_window": 1439,
  "blocks_until_next_epoch": 17279
}
```

### 윈도우 전환 (블록 1441, 에포크 0 윈도우 0→1)

```json
{
  "block_height": 1441,
  "epoch_number": 0,
  "epoch_start_block": 1,
  "epoch_end_block": 17280,
  "window_number": 1,
  "window_start_block": 1441,
  "window_end_block": 2880,
  "blocks_until_next_window": 1439,
  "blocks_until_next_epoch": 15839
}
```

### 에포크 전환 (블록 17281, 에포크 0→1)

```json
{
  "block_height": 17281,
  "epoch_number": 1,
  "epoch_start_block": 17281,
  "epoch_end_block": 34560,
  "window_number": 0,
  "window_start_block": 17281,
  "window_end_block": 18720,
  "blocks_until_next_window": 1439,
  "blocks_until_next_epoch": 17279
}
```

### 에포크 마지막 블록 (블록 17280)

```json
{
  "block_height": 17280,
  "epoch_number": 0,
  "epoch_start_block": 1,
  "epoch_end_block": 17280,
  "window_number": 11,
  "window_start_block": 15841,
  "window_end_block": 17280,
  "blocks_until_next_window": 0,
  "blocks_until_next_epoch": 0
}
```

## 7.3 TX 결과 Mock

### 성공: MsgSubmitActivity

```json
{
  "tx_hash": "A1B2C3D4E5F60718293A4B5C6D7E8F90A1B2C3D4E5F60718293A4B5C6D7E8F90",
  "height": 726500,
  "code": 0,
  "gas_used": 85000,
  "gas_wanted": 120000,
  "raw_log": "[]",
  "events": [
    {
      "type": "activity_submitted",
      "attributes": [
        { "key": "node_id", "value": "node-seocheon1abc123" },
        { "key": "submitter", "value": "seocheon1agent789" },
        { "key": "activity_hash", "value": "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90" },
        { "key": "content_uri", "value": "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG" },
        { "key": "epoch", "value": "42" },
        { "key": "window", "value": "0" },
        { "key": "sequence", "value": "1" },
        { "key": "block_height", "value": "726500" }
      ]
    }
  ]
}
```

### 실패: 쿼터 초과

```json
{
  "tx_hash": "F0E1D2C3B4A5968778695A4B3C2D1E0FF0E1D2C3B4A5968778695A4B3C2D1E0F",
  "height": 726510,
  "code": 1203,
  "gas_used": 45000,
  "gas_wanted": 120000,
  "raw_log": "failed to execute message; message index: 0: activity quota exceeded for this epoch: invalid request",
  "events": []
}
```

### 성공: MsgWithdrawNodeCommission

```json
{
  "tx_hash": "C3D4E5F60718293A4B5C6D7E8F90A1B2C3D4E5F60718293A4B5C6D7E8F90A1B2",
  "height": 743100,
  "code": 0,
  "gas_used": 95000,
  "gas_wanted": 150000,
  "raw_log": "[]",
  "events": [
    {
      "type": "node_commission_withdrawn",
      "attributes": [
        { "key": "node_id", "value": "node-seocheon1abc123" },
        { "key": "operator_amount", "value": "700000" },
        { "key": "agent_amount", "value": "300000" }
      ]
    }
  ]
}
```

### 성공: MsgSend

```json
{
  "tx_hash": "D4E5F60718293A4B5C6D7E8F90A1B2C3D4E5F60718293A4B5C6D7E8F90A1B2C3",
  "height": 726550,
  "code": 0,
  "gas_used": 65000,
  "gas_wanted": 100000,
  "raw_log": "[]",
  "events": [
    {
      "type": "transfer",
      "attributes": [
        { "key": "recipient", "value": "seocheon1recipient..." },
        { "key": "sender", "value": "seocheon1agent789" },
        { "key": "amount", "value": "1000000usum" }
      ]
    },
    {
      "type": "message",
      "attributes": [
        { "key": "action", "value": "/cosmos.bank.v1beta1.MsgSend" },
        { "key": "sender", "value": "seocheon1agent789" },
        { "key": "module", "value": "bank" }
      ]
    }
  ]
}
```
