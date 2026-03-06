#!/usr/bin/env bash
# sdk_e2e.sh — SDK E2E 통합 테스트 오케스트레이션 스크립트
#
# 흐름:
#   1. 단일 밸리데이터 로컬 테스트넷 초기화 (seocheon init)
#   2. 에이전트 키 생성 및 genesis 자금 지원
#   3. 테스트넷 구동 (백그라운드)
#   4. 블록 생성 대기
#   5. 테스트 노드 등록 (CLI)
#   6. 각 SDK E2E 테스트 실행
#   7. 결과 집계 및 정리
#
# 사용법:
#   make test-sdk-e2e
#   또는: bash scripts/sdk_e2e.sh
#
# 사전 조건:
#   - seocheon 바이너리 빌드 완료 (make install)
#   - Go, Node.js, Python(uv), Kotlin(gradle), Swift, .NET 설치

set -euo pipefail

###############################################################################
# 설정
###############################################################################

TESTNET_HOME="${TESTNET_HOME:-.e2e-testnet}"
CHAIN_ID="${CHAIN_ID:-seocheon-e2e}"
DENOM="uppyeo"
MONIKER="validator0"

RPC_ENDPOINT="http://localhost:26657"
GRPC_ENDPOINT="http://localhost:1317"

SEOCHEON_BIN="${SEOCHEON_BIN:-seocheon}"
KEYRING="test"

# 테스트 타임아웃 (초)
BLOCK_WAIT_TIMEOUT=60
TESTNET_PID=""

###############################################################################
# 색상 출력 헬퍼
###############################################################################

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

###############################################################################
# 정리 함수
###############################################################################

cleanup() {
    info "테스트넷 정리 중..."
    if [[ -n "$TESTNET_PID" ]] && kill -0 "$TESTNET_PID" 2>/dev/null; then
        kill "$TESTNET_PID" 2>/dev/null || true
        wait "$TESTNET_PID" 2>/dev/null || true
    fi
    # 남아 있는 seocheon 프로세스 종료
    pkill -f "seocheon start" 2>/dev/null || true
    rm -rf "$TESTNET_HOME"
    info "정리 완료"
}

trap cleanup EXIT INT TERM

###############################################################################
# 1. 테스트넷 초기화
###############################################################################

info "=== 1. 테스트넷 초기화 (chain-id: $CHAIN_ID) ==="

rm -rf "$TESTNET_HOME"

$SEOCHEON_BIN init "$MONIKER" \
    --chain-id "$CHAIN_ID" \
    --home "$TESTNET_HOME" \
    --default-denom "$DENOM" \
    > /dev/null 2>&1

###############################################################################
# 2. 키 생성
###############################################################################

info "=== 2. 밸리데이터 키 생성 ==="

VALIDATOR_JSON=$(
    $SEOCHEON_BIN keys add validator \
        --keyring-backend "$KEYRING" \
        --home "$TESTNET_HOME" \
        --output json 2>&1
)
VALIDATOR_ADDR=$(echo "$VALIDATOR_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin)['address'])")
info "밸리데이터 주소: $VALIDATOR_ADDR"

info "=== 에이전트 키 생성 ==="

AGENT_JSON=$(
    $SEOCHEON_BIN keys add agent \
        --keyring-backend "$KEYRING" \
        --home "$TESTNET_HOME" \
        --output json 2>&1
)
AGENT_MNEMONIC=$(echo "$AGENT_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin)['mnemonic'])")
AGENT_ADDR=$(echo "$AGENT_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin)['address'])")
info "에이전트 주소: $AGENT_ADDR"

info "=== 노드 오퍼레이터 키 생성 ==="

NODE_OPERATOR_JSON=$(
    $SEOCHEON_BIN keys add node-operator \
        --keyring-backend "$KEYRING" \
        --home "$TESTNET_HOME" \
        --output json 2>&1
)
NODE_OPERATOR_ADDR=$(echo "$NODE_OPERATOR_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin)['address'])")
info "노드 오퍼레이터 주소: $NODE_OPERATOR_ADDR"

###############################################################################
# 3. Genesis 패치 (bond denom → uppyeo)
###############################################################################

info "=== 3. Genesis 패치 (bond_denom → $DENOM) ==="

GENESIS="$TESTNET_HOME/config/genesis.json"

python3 - "$GENESIS" "$DENOM" << 'PYEOF'
import sys, json
genesis_path, denom = sys.argv[1], sys.argv[2]
with open(genesis_path) as f:
    g = json.load(f)

# 스테이킹 bond denom
g["app_state"]["staking"]["params"]["bond_denom"] = denom
# crisis fee denom
if "crisis" in g["app_state"]:
    g["app_state"]["crisis"]["constant_fee"]["denom"] = denom
# mint denom
if "mint" in g["app_state"]:
    g["app_state"]["mint"]["params"]["mint_denom"] = denom
# gov min deposit denom
if "gov" in g["app_state"]:
    for d in g["app_state"]["gov"].get("params", {}).get("min_deposit", []):
        d["denom"] = denom

with open(genesis_path, "w") as f:
    json.dump(g, f, indent=2)
PYEOF

###############################################################################
# 4. Genesis 계정 추가 및 gentx
###############################################################################

info "=== 4. Genesis 계정 추가 ==="

$SEOCHEON_BIN genesis add-genesis-account "$VALIDATOR_ADDR" \
    "100000000000000${DENOM}" \
    --keyring-backend "$KEYRING" \
    --home "$TESTNET_HOME"

$SEOCHEON_BIN genesis add-genesis-account "$AGENT_ADDR" \
    "10000000000000${DENOM}" \
    --keyring-backend "$KEYRING" \
    --home "$TESTNET_HOME"

$SEOCHEON_BIN genesis add-genesis-account "$NODE_OPERATOR_ADDR" \
    "10000000000000${DENOM}" \
    --keyring-backend "$KEYRING" \
    --home "$TESTNET_HOME"

info "=== gentx 생성 ==="

$SEOCHEON_BIN genesis gentx validator \
    "10000000000000${DENOM}" \
    --chain-id "$CHAIN_ID" \
    --keyring-backend "$KEYRING" \
    --home "$TESTNET_HOME" \
    > /dev/null 2>&1

$SEOCHEON_BIN genesis collect-gentxs \
    --home "$TESTNET_HOME" \
    > /dev/null 2>&1

###############################################################################
# 5. 테스트넷 구동
###############################################################################

info "=== 5. 테스트넷 구동 (백그라운드) ==="

# app.toml: REST API 서버 활성화 (포트 1317)
APP_TOML="$TESTNET_HOME/config/app.toml"
python3 - "$APP_TOML" << 'PYEOF'
import re, sys
path = sys.argv[1]
with open(path) as f:
    content = f.read()
content = re.sub(
    r'(\[api\](?:[^\[]*?\n)enable = )false',
    r'\1true',
    content,
    flags=re.DOTALL
)
with open(path, 'w') as f:
    f.write(content)
PYEOF
info "app.toml 패치 완료: REST API 서버(포트 1317) 활성화"

$SEOCHEON_BIN start \
    --home "$TESTNET_HOME" \
    --minimum-gas-prices "250${DENOM}" \
    > /tmp/seocheon_e2e.log 2>&1 &
TESTNET_PID=$!

info "seocheon 프로세스 PID: $TESTNET_PID"

###############################################################################
# 6. 블록 생성 대기
###############################################################################

info "=== 6. 첫 블록 대기 (최대 ${BLOCK_WAIT_TIMEOUT}초) ==="

wait_for_block() {
    local timeout=$1
    local start
    start=$(date +%s)
    while true; do
        local now
        now=$(date +%s)
        if (( now - start > timeout )); then
            error "블록 생성 타임아웃 (${timeout}초)"
            return 1
        fi
        local height
        height=$(curl -s "${RPC_ENDPOINT}/status" 2>/dev/null \
            | python3 -c "import sys,json; print(json.load(sys.stdin)['result']['sync_info']['latest_block_height'])" 2>/dev/null || echo "0")
        if [[ "$height" -gt 2 ]]; then
            info "블록 생성 확인: height=$height"
            return 0
        fi
        sleep 2
    done
}

wait_for_block "$BLOCK_WAIT_TIMEOUT"

###############################################################################
# 7. 테스트 노드 등록
###############################################################################

info "=== 7. 테스트 노드 등록 ==="

# 새 ed25519 pubkey 생성 (validator consensus pubkey와 충돌 방지)
NEW_NODE_PUBKEY=$(python3 -c "
import secrets, base64, json
raw = secrets.token_bytes(32)
print(json.dumps({'@type': '/cosmos.crypto.ed25519.PubKey', 'key': base64.b64encode(raw).decode()}))
")
info "새 노드 pubkey 생성 완료"

# 변경 4: 등록 전 node-operator 잔액 사전 확인
info "node-operator 잔액 확인..."
$SEOCHEON_BIN query bank balances "$NODE_OPERATOR_ADDR" \
    --home "$TESTNET_HOME" 2>&1 || true

# 변경 1: TX 전체 출력 + TX hash 캡처 (--output json 필수)
REGISTER_TX_OUTPUT=$(
    $SEOCHEON_BIN tx node register-node \
        --pubkey "$NEW_NODE_PUBKEY" \
        --agent-address "$AGENT_ADDR" \
        --agent-share "0.200000000000000000" \
        --max-agent-share-change-rate "0.010000000000000000" \
        --commission-rate "0.100000000000000000" \
        --commission-max-rate "0.200000000000000000" \
        --commission-max-change-rate "0.010000000000000000" \
        --description "e2e-test-node" \
        --from node-operator \
        --keyring-backend "$KEYRING" \
        --home "$TESTNET_HOME" \
        --chain-id "$CHAIN_ID" \
        --fees "25000${DENOM}" \
        --gas 300000 \
        --output json \
        --yes 2>&1
)
echo "$REGISTER_TX_OUTPUT"
REGISTER_TXHASH=$(echo "$REGISTER_TX_OUTPUT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(data.get('txhash', ''))
except Exception:
    print('')
" 2>/dev/null || echo "")
info "TX hash: ${REGISTER_TXHASH:-<캡처 실패>}"

# 변경 2: sleep 후 TX 결과 쿼리 (DeliverTx 실제 결과 확인)
info "노드 등록 TX 전송 완료. 블록 포함 대기..."
sleep 10

if [[ -n "$REGISTER_TXHASH" ]]; then
    info "TX 결과 쿼리: $REGISTER_TXHASH"
    TX_RESULT=$($SEOCHEON_BIN query tx "$REGISTER_TXHASH" \
        --home "$TESTNET_HOME" \
        --node "$RPC_ENDPOINT" \
        --output json 2>&1 || echo '{}')
    TX_CODE=$(echo "$TX_RESULT" | python3 -c "
import sys, json
try:
    print(json.load(sys.stdin).get('code', -1))
except Exception:
    print(-1)
" 2>/dev/null || echo "-1")
    TX_LOG=$(echo "$TX_RESULT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('raw_log', '') or d.get('logs', ''))
except Exception:
    print('')
" 2>/dev/null || echo "")
    info "TX code: $TX_CODE"
    if [[ "$TX_CODE" != "0" ]]; then
        error "TX 실패! code=$TX_CODE"
        error "raw_log: $TX_LOG"
    fi
fi

# 검증: AGENT_ADDR이 AgentIndex에 등록됐는지 확인
NODE_ID=$(curl -s "${GRPC_ENDPOINT}/seocheon/node/v1/nodes/by-agent/${AGENT_ADDR}" 2>/dev/null \
    | python3 -c "import sys,json; print(json.load(sys.stdin).get('node',{}).get('id',''))" 2>/dev/null || echo "")

# 변경 3: 실패 시 테스트넷 로그 + 잔액 출력
if [[ -z "$NODE_ID" ]]; then
    error "노드 에이전트 등록 확인 실패: AGENT_ADDR=${AGENT_ADDR}"
    error "=== 테스트넷 로그 (마지막 80줄) ==="
    tail -80 /tmp/seocheon_e2e.log >&2 || true
    error "=== node-operator 잔액 ==="
    $SEOCHEON_BIN query bank balances "$NODE_OPERATOR_ADDR" \
        --home "$TESTNET_HOME" 2>&1 >&2 || true
    exit 1
fi
info "노드 에이전트 등록 확인: nodeId=${NODE_ID}"

###############################################################################
# 8. 환경변수 설정
###############################################################################

export SEOCHEON_RPC="$RPC_ENDPOINT"
export SEOCHEON_GRPC="$GRPC_ENDPOINT"
export SEOCHEON_CHAIN_ID="$CHAIN_ID"
export SEOCHEON_MNEMONIC="$AGENT_MNEMONIC"

info "=== E2E 환경변수 설정 완료 ==="
info "  SEOCHEON_RPC=$SEOCHEON_RPC"
info "  SEOCHEON_GRPC=$SEOCHEON_GRPC"
info "  SEOCHEON_CHAIN_ID=$SEOCHEON_CHAIN_ID"
info "  SEOCHEON_MNEMONIC=[설정됨]"

###############################################################################
# 9. SDK E2E 테스트 실행
###############################################################################

PASS=0
FAIL=0

run_test() {
    local name=$1
    shift
    info "--- $name ---"
    if "$@"; then
        info "✅ $name PASS"
        PASS=$((PASS + 1))
    else
        error "❌ $name FAIL"
        FAIL=$((FAIL + 1))
    fi
}

info "=== 9. SDK E2E 테스트 실행 ==="

# Go SDK
run_test "Go SDK" bash -c "
    cd sdk/go && \
    go test -tags e2e -v -timeout 120s ./e2e/... 2>&1
"

# TypeScript SDK
run_test "TypeScript SDK" bash -c "
    cd sdk/typescript && \
    npx vitest run --config e2e/vitest.config.e2e.ts 2>&1
"

# Python SDK
run_test "Python SDK" bash -c "
    cd sdk/python && \
    uv run pytest e2e/ -m e2e -v 2>&1
"

# Kotlin SDK
run_test "Kotlin SDK" bash -c "
    cd sdk/kotlin && \
    ./gradlew e2eTest 2>&1
"

# Swift SDK
run_test "Swift SDK" bash -c "
    cd sdk/swift && \
    swift test --filter E2E 2>&1
"

# C# SDK
run_test "C# SDK" bash -c "
    cd sdk/csharp && \
    dotnet test --filter 'Category=e2e' --verbosity minimal 2>&1
"

###############################################################################
# 10. 결과 요약
###############################################################################

echo ""
echo "============================================"
echo " SDK E2E 테스트 결과"
echo "============================================"
echo " ✅ PASS: $PASS"
echo " ❌ FAIL: $FAIL"
echo "============================================"

if [[ "$FAIL" -gt 0 ]]; then
    exit 1
fi
