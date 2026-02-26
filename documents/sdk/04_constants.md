# 전역 상수 및 설정 스키마

> **담당**: SDK 개발자 / 체인 코어
> **관련 문서**: [아키텍처](01_architecture.md) · [인터페이스](02_interfaces.md) · [메서드](03_methods.md) · [통신](05_communication.md) · [전체 목차](README.md)

---

## 4.1 체인 파라미터 기본값

```
// 에포크/윈도우 (x/activity params)
EPOCH_LENGTH                  = 17280         // 블록/에포크 (~24시간, 5초/블록)
WINDOWS_PER_EPOCH             = 12            // 윈도우/에포크
MIN_ACTIVE_WINDOWS            = 8             // 활동 보상 자격 최소 윈도우
WINDOW_LENGTH                 = 1440          // 블록/윈도우 (EPOCH_LENGTH / WINDOWS_PER_EPOCH)

// 쿼터 (x/activity params)
SELF_FUNDED_QUOTA             = 100           // 자비 부담 노드 에포크당 쿼터
FEEGRANT_QUOTA                = 10            // feegrant 노드 에포크당 쿼터

// 프루닝 (x/activity params)
ACTIVITY_PRUNING_KEEP_BLOCKS  = 6307200       // 활동 기록 보존 블록 수 (~1년)

// 활동 비용 모델 (x/activity params)
FEE_THRESHOLD_MULTIPLIER      = 3             // 수수료 활성화 임계치 배수
BASE_ACTIVITY_FEE             = 1000000       // 기본 활동 수수료 (usum, 1 KKOT)
FEE_EXPONENT                  = 5000          // 비용 곡선 지수 (basis points, 0.5)
MAX_ACTIVITY_FEE              = 100000000     // 최대 활동 수수료 (usum, 100 KKOT)
MIN_FEEGRANT_QUOTA            = 8             // 포화 시 feegrant 최소 쿼터
QUOTA_REDUCTION_RATE          = 5000          // 쿼터 축소율 (basis points, 0.5)
FEEGRANT_FEE_EXEMPT           = true          // feegrant 노드 수수료 면제

// 이중 보상 풀 (x/activity params)
D_MIN                         = 3000          // 위임 풀 최소 비율 (basis points, 0.3)
FEE_TO_ACTIVITY_POOL_RATIO    = 8000          // 수수료→활동풀 비율 (basis points, 0.8)

// 노드 등록 (x/node params)
MAX_REGISTRATIONS_PER_BLOCK   = 5
REGISTRATION_COOLDOWN_BLOCKS  = 100
REGISTRATION_DEPOSIT          = "0"           // usum (초기값 0)
MAX_TAGS                      = 10
MAX_TAG_LENGTH                = 32

// Agent 권한 (x/node params)
AGENT_ALLOWED_MSG_TYPES       = ["/seocheon.activity.v1.MsgSubmitActivity", "/cosmos.bank.v1beta1.MsgSend"]
AGENT_FEEGRANT_ALLOWED_MSG_TYPES = ["/seocheon.activity.v1.MsgSubmitActivity"]
AGENT_ADDRESS_CHANGE_COOLDOWN = 17280         // 블록 (1 에포크)

// 시간-블록 변환 (5초/블록 기준)
BLOCKS_PER_HOUR               = 720
BLOCKS_PER_DAY                = 17280
BLOCKS_PER_YEAR               = 6307200
UNBONDING_PERIOD_BLOCKS       = 362880        // ~21일
FEEGRANT_EXPIRY_BLOCKS        = 3110400       // ~180일 (6개월)
```

## 4.2 토큰 Denomination 상수

```
// Denomination 정의
TOKEN_BASE_DENOM              = "usum"        // base denomination (10^0)
TOKEN_MILLI_DENOM             = "hon"         // milli denomination (10^3)
TOKEN_DISPLAY_DENOM           = "kkot"        // display denomination (10^6)

// 변환 계수
USUM_PER_HON                  = 1000          // 1 hon = 1,000 usum
USUM_PER_KKOT                 = 1000000       // 1 KKOT = 1,000,000 usum
HON_PER_KKOT                  = 1000          // 1 KKOT = 1,000 hon

// 유래: 서천꽃밭 신화(이공본풀이)
// 숨(usum) = 호흡 → 혼(hon) = 영혼 → 꽃(kkot) = 만개
```

**변환 함수**:
```pseudocode
function usum_to_kkot(usum: uint64) → string:
  integer_part = usum / USUM_PER_KKOT
  decimal_part = usum % USUM_PER_KKOT
  RETURN format("{}.{:06d}", integer_part, decimal_part)

function kkot_to_usum(kkot: string) → uint64:
  RETURN parse_decimal(kkot) * USUM_PER_KKOT
```

## 4.3 에러 코드 상수

### x/node 에러 (1100번대)

| 상수명 | 코드 | 메시지 |
|--------|------|--------|
| `ERR_INVALID_SIGNER` | 1100 | expected gov account as only signer for proposal message |
| `ERR_NODE_NOT_FOUND` | 1101 | node not found |
| `ERR_NODE_ALREADY_EXISTS` | 1102 | node already exists for this operator |
| `ERR_AGENT_ADDRESS_ALREADY_USED` | 1103 | agent address already registered to another node |
| `ERR_REGISTRATION_POOL_DEPLETED` | 1104 | registration pool has insufficient funds |
| `ERR_MAX_REGISTRATIONS_PER_BLOCK` | 1105 | maximum registrations per block exceeded |
| `ERR_INVALID_AGENT_SHARE` | 1106 | agent share must be between 0 and 100 |
| `ERR_INVALID_TAGS` | 1107 | invalid tags |
| `ERR_UNAUTHORIZED_OPERATOR` | 1108 | unauthorized: signer is not the node operator |
| `ERR_UNAUTHORIZED_AGENT_MSG` | 1109 | agent address is not authorized for this message type |
| `ERR_AGENT_ADDRESS_CHANGE_COOLDOWN` | 1110 | agent address change cooldown not elapsed |
| `ERR_NODE_NOT_ACTIVE` | 1111 | node is not in an active state |
| `ERR_INVALID_CONSENSUS_PUBKEY` | 1112 | invalid consensus public key |
| `ERR_AGENT_SHARE_CHANGE_EXCEEDS_MAX` | 1113 | agent share change exceeds max change rate |
| `ERR_NODE_INACTIVE` | 1116 | node is inactive or jailed |
| `ERR_VALIDATOR_ALREADY_EXISTS` | 1117 | validator with this consensus pubkey already exists |
| `ERR_AGENT_SHARE_CHANGE_PENDING` | 1118 | a pending agent share change already exists |

### x/activity 에러 (1200번대)

| 상수명 | 코드 | 메시지 |
|--------|------|--------|
| `ERR_SUBMITTER_NOT_REGISTERED` | 1200 | submitter agent address is not registered to any node |
| `ERR_NODE_NOT_ELIGIBLE` | 1201 | node is not in an eligible status (REGISTERED or ACTIVE) |
| `ERR_DUPLICATE_ACTIVITY_HASH` | 1202 | duplicate (activity_hash, content_uri) pair already exists |
| `ERR_QUOTA_EXCEEDED` | 1203 | activity quota exceeded for this epoch |
| `ERR_INVALID_ACTIVITY_HASH` | 1204 | activity hash must be exactly 64 hex characters (32 bytes) |
| `ERR_INVALID_CONTENT_URI` | 1205 | content URI must not be empty |
| `ERR_INVALID_AUTHORITY` | 1206 | expected gov account as only signer for proposal message |
| `ERR_INVALID_PARAMS` | 1207 | invalid module parameters |
| `ERR_ACTIVITY_NOT_FOUND` | 1208 | activity record not found |
| `ERR_NODE_NOT_FOUND_ACTIVITY` | 1209 | node not found |

### SDK 자체 에러 (9000번대)

| 상수명 | 코드 | 메시지 |
|--------|------|--------|
| `ERR_NOT_CONNECTED` | 9000 | SDK is not connected to chain |
| `ERR_BROADCAST_FAILED` | 9001 | transaction broadcast failed |
| `ERR_TX_TIMEOUT` | 9002 | transaction confirmation timeout |
| `ERR_TX_NOT_FOUND` | 9003 | transaction not found |
| `ERR_SIGNING_FAILED` | 9004 | transaction signing failed |
| `ERR_INVALID_CONFIG` | 9005 | invalid SDK configuration |
| `ERR_QUERY_FAILED` | 9006 | chain query failed |
| `ERR_INVALID_ADDRESS` | 9007 | invalid address format |

## 4.4 SDKConfig JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "SeocheonSDKConfig",
  "type": "object",
  "required": ["chain", "signing"],
  "properties": {
    "chain": {
      "type": "object",
      "required": ["chain_id", "rpc_endpoint", "grpc_endpoint"],
      "properties": {
        "chain_id": { "type": "string", "description": "Chain ID", "examples": ["seocheon-1"] },
        "rpc_endpoint": { "type": "string", "format": "uri", "description": "CometBFT RPC endpoint" },
        "grpc_endpoint": { "type": "string", "format": "uri", "description": "Cosmos gRPC endpoint" },
        "gas_price": { "type": "string", "default": "0.025usum", "pattern": "^[0-9.]+[a-z]+$" },
        "gas_adjustment": { "type": "number", "default": 1.3, "minimum": 1.0 }
      }
    },
    "signing": {
      "type": "object",
      "required": ["mode"],
      "properties": {
        "mode": { "type": "string", "enum": ["vault", "keystore", "direct"] },
        "vault_endpoint": { "type": "string", "format": "uri" },
        "key_name": { "type": "string" },
        "keystore_path": { "type": "string" },
        "passphrase_env": { "type": "string" },
        "mnemonic": { "type": "string" }
      },
      "allOf": [
        { "if": { "properties": { "mode": { "const": "vault" } } },
          "then": { "required": ["vault_endpoint", "key_name"] } },
        { "if": { "properties": { "mode": { "const": "keystore" } } },
          "then": { "required": ["keystore_path", "passphrase_env"] } },
        { "if": { "properties": { "mode": { "const": "direct" } } },
          "then": { "required": ["mnemonic"] } }
      ]
    },
    "tx": {
      "type": "object",
      "properties": {
        "broadcast_mode": { "type": "string", "enum": ["sync", "async"], "default": "sync" },
        "confirm_timeout_ms": { "type": "integer", "default": 30000, "minimum": 1000 },
        "confirm_poll_interval_ms": { "type": "integer", "default": 1000, "minimum": 100 }
      }
    }
  }
}
```
