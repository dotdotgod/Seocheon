# CosmWasm (x/wasm) 통합

## 개요

Seocheon 블록체인에 CosmWasm 스마트 컨트랙트 실행 환경을 통합한다. 이를 통해 AI 에이전트와 노드 운영자가 온체인 스마트 컨트랙트를 배포·실행할 수 있는 기반이 마련된다.

## 호환성 분석

### 의존성 버전 매트릭스

| 컴포넌트 | Seocheon | wasmd v0.61.8 | 호환성 |
|----------|----------|---------------|--------|
| Cosmos SDK | v0.53.6 | v0.53.5 | ✅ 호환 (patch 차이) |
| CometBFT | v0.38.21 | v0.38.21 | ✅ 정확히 일치 |
| IBC-go | v10.4.0→v10.5.0 | v10.5.0 | ✅ wasmd가 v10.5.0 요구, 자동 업그레이드 |
| Go | 1.24.1 | 1.24.0 | ✅ 호환 |
| wasmvm | - | v3.0.3 | ✅ 신규 추가 |

### 자동 업그레이드된 의존성

wasmd v0.61.8 추가 시 다음 의존성이 자동 업그레이드됨:
- `cosmossdk.io/x/evidence`: v0.1.1 → v0.2.0
- `cosmossdk.io/x/feegrant`: v0.1.1 → v0.2.0
- `cosmossdk.io/x/nft`: v0.1.0 → v0.2.0
- `github.com/cosmos/iavl`: v1.2.2 → v1.2.6
- `github.com/cosmos/ibc-go/v10`: v10.4.0 → v10.5.0

## 아키텍처

### 모듈 등록 방식

wasmd는 depinject를 지원하지 않으므로, IBC 모듈과 동일한 **수동 등록 패턴**을 사용한다.

```
app.go (New 함수)
├── depinject.Inject()      ← 기존 모듈 (auth, bank, staking, node, activity, randomness...)
├── registerIBCKeepers()    ← Phase 1: IBC Keeper 생성
├── registerWasmModule()    ← Phase 2: Wasm Keeper 생성 (IBC Keeper 필요)
├── registerIBCRouterAndModules() ← Phase 3: IBC Router 생성 (wasm 라우트 포함) + 모듈 등록
└── appBuilder.Build()
```

### 의존성 순서

```
IBC Keeper 생성
    ↓
Wasm Keeper 생성 (IBCKeeper.ChannelKeeper, ChannelKeeperV2 필요)
    ↓
IBC Router 생성 (wasm IBC 핸들러 라우트 포함)
    ↓
Router Seal
```

### 모듈 실행 순서

**BeginBlockers**: `... → staking → epochs → ibc → node → randomness → wasm`
**EndBlockers**: `... → activity → node → randomness → wasm`
**InitGenesis**: `... → ibc → transfer → ica → wasm → activity → node → randomness`

wasm은:
- BeginBlockers에서 staking 이후 실행
- EndBlockers에서 마지막 실행
- InitGenesis에서 IBC 모듈 이후, 커스텀 체인 모듈 이전에 초기화

### 모듈 계정 권한

| 모듈 계정 | 권한 |
|-----------|------|
| `wasm` | `Burner` |

### Ante Handler 데코레이터

기존 Seocheon ante 데코레이터에 wasm 데코레이터 4개 추가:

1. **RegistrationFeeDecorator** (Seocheon)
2. **AgentPermissionDecorator** (Seocheon)
3. **LimitSimulationGasDecorator** (wasm) — 시뮬레이션 TX 가스 제한
4. **CountTXDecorator** (wasm) — 블록 내 TX 위치 추적
5. **GasRegisterDecorator** (wasm) — 가스 레지스터 컨텍스트 주입
6. **TxContractsDecorator** (wasm) — TX 내 컨트랙트 접근 추적 (캐시 최적화)

### IBC 통합

wasm 모듈은 IBC v1 라우터에 `wasm` 포트로 등록된다:
```
IBC Router
├── transfer    → TransferStack
├── icacontroller → ICAControllerStack
├── icahost     → ICAHostStack
└── wasm        → WasmIBCHandler
```

이를 통해 CosmWasm 컨트랙트가 IBC 패킷을 송수신할 수 있다.

## 파일 구조

```
app/
├── wasm.go     ← Wasm Keeper 생성, AppModule 등록 (신규)
├── ibc.go      ← IBC Keeper 생성 분리 + IBC Router/모듈 등록 분리 (수정)
├── app.go      ← WasmKeeper 필드, 3단계 등록 호출 (수정)
├── app_config.go ← wasm BeginBlockers/EndBlockers/InitGenesis/moduleAccPerms (수정)
└── ante.go     ← wasm ante 데코레이터 추가 (수정)
```

## WasmKeeper 설정

```go
wasmkeeper.NewKeeper(
    codec,
    storeService,
    accountKeeper,
    bankKeeper,
    stakingKeeper,
    distrQuerier,
    ibcChannelKeeper,      // ICS4Wrapper
    ibcChannelKeeper,      // ChannelKeeper
    ibcChannelKeeperV2,    // ChannelKeeperV2 (IBC v2)
    transferKeeper,        // ICS20TransferPortSource
    msgServiceRouter,
    grpcQueryRouter,
    wasmDir,               // ~/.seocheon/wasm
    nodeConfig,            // wasm.ReadNodeConfig(appOpts)
    VMConfig{},
    BuiltInCapabilities(), // iterator, staking, stargate, cosmwasm_2_1, ...
    govModuleAddr,
)
```

## 설정 옵션

`app.toml` 또는 CLI 플래그로 설정 가능:

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `--wasm.memory_cache_size` | 100 (MiB) | Wasm 모듈 인메모리 캐시 크기 |
| `--wasm.query_gas_limit` | 3,000,000 | 컨트랙트 쿼리 최대 가스 |
| `--wasm.simulation_gas_limit` | (없음) | 시뮬레이션 TX 최대 가스 |
| `--wasm.skip_wasmvm_version_check` | false | libwasmvm 버전 검증 건너뛰기 |

## 플랫폼 요구사항

- **CGO 필수**: wasmvm v3은 C 바인딩(libwasmvm)을 사용하므로 `CGO_ENABLED=1` 필요
- **지원 플랫폼**: Linux (amd64, arm64), macOS (amd64, arm64)
- **미지원**: Windows, musl-based Linux (Alpine 등에서는 별도 컴파일 필요)

## 향후 과제

1. **커스텀 컨트랙트 바인딩**: Seocheon 커스텀 모듈(x/node, x/activity)에 대한 CosmWasm 쿼리/메시지 바인딩 개발
2. **거버넌스 통합**: 컨트랙트 업로드/인스턴스화에 대한 거버넌스 제어 파라미터 설정
3. **가스 최적화**: Seocheon 활동 패턴에 맞는 wasm 가스 파라미터 튜닝
4. **E2E 테스트**: 컨트랙트 배포 → 실행 → IBC 전송 E2E 테스트 시나리오 추가
