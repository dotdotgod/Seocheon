# Randomness 모듈 (x/randomness)

## 1. 개요

x/randomness 모듈은 [drand](https://drand.love/) 분산 랜덤 비콘 네트워크의 검증 가능한 난수를 온체인에 저장하고, CosmWasm 컨트랙트에 공정한 랜덤 소스를 제공하는 인프라 모듈이다. 카드덱 셔플, 랜덤 선택 등 공정성이 요구되는 오프체인 에이전트 기능을 지원한다.

**합의 미사용 원칙**: x/randomness는 합의 과정에 사용하지 않는다. 비콘이 없어도 블록 생성은 정상 진행되며, 컨트랙트에서만 소비하는 순수 인프라 모듈이다.

**설계 원칙**:

1. **검증 가능성**: drand 비콘의 BLS12-381 서명을 온체인에서 검증 (quicknet, `bls-unchained-g1-rfc9380`)
2. **무신뢰**: 특정 참여자를 신뢰할 필요 없이 drand 네트워크의 분산 비콘 사용
3. **유저 1회 서명**: 전체 흐름에서 유저는 `MsgRequestRandomness` 1회만 서명. 이행은 프로토콜(EndBlocker)이 자동 처리
4. **Front-run 불가**: 커밋 시점에 비콘이 존재하지 않으므로 결과 예측 불가
5. **최소 의존성**: 다른 커스텀 모듈(x/node, x/activity)에 의존하지 않음

**3중 리스크 완화 메커니즘**:

| # | 메커니즘 | 완화 대상 |
|---|---------|----------|
| 1 | BLS 서명 검증 | 가짜 비콘 제출 차단 |
| 2 | Commit-Reveal + Future Beacon | Front-running / MEV 방지 |
| 3 | Multi-source Mixing | 단일 소스 조작 불가 |

---

## 2. 핵심 설계: 2단계 "유저 1회 서명 + EndBlocker 자동 이행"

기존 3단계(Request → SubmitBeacon → FulfillRandomness) 설계에서 발견된 문제점을 해소한 **2단계 설계**로 단순화한다.

### 2.1 전체 흐름

```
Block N    : User → MsgRequestRandomness(commit_hash, request_fee)  ← 유저 1회 서명
Block N+K  : Relayer → MsgSubmitBeacon(round=target_round)           ← 릴레이어 (비콘만 저장)
Block N+K  : EndBlocker → 자동 fulfill                               ← 프로토콜이 처리 (가스 무관)
             result = SHA256(beacon.randomness || AppHash || commit_hash || i)
```

### 2.2 흐름 다이어그램

```
                     유저 (1회 서명)              Relayer Bot
                          │                          │
Block N    ───────────────┼──────────────────────────┤
                          │                          │
                   MsgRequestRandomness              │
                   (commit_hash, num_words,           │
                    request_fee)                      │
                          │                          │
                   target_round 계산 (block_time 기반)│
                   request_fee → 모듈 계정 에스크로    │
                   status = PENDING                   │
                          │                          │
Block N+K  ───────────────┼──────────────────────────┤
                          │                          │
                          │               MsgSubmitBeacon
                          │            (round=target_round)
                          │              BLS 검증 통과
                          │                          │
                          │            ┌─ EndBlocker ─┐
                          │            │ 자동 fulfill  │
                          │            │ Multi-source  │
                          │            │ Mixing:       │
                          │            │ SHA256(       │
                          │            │  beacon ||    │
                          │            │  AppHash ||   │
                          │            │  commit_hash  │
                          │            │  || i)        │
                          │            │              │
                          │            │ status =     │
                          │            │ FULFILLED    │
                          │            │              │
                          │            │ request_fee  │
                          │            │ → relayer    │
                          │            └──────────────┘
                          │                          │
                   컨트랙트가 쿼리로 결과 조회          │
                   RandomnessRequest(id)             │
```

### 2.3 기존 대비 핵심 변경

| 항목 | 기존 (3단계) | 변경 (2단계) | 이유 |
|------|------------|------------|------|
| Fulfill 방식 | MsgFulfillRandomness (별도 TX) | **EndBlocker 자동** | Fulfiller Bot 불필요, Gas Griefing 방지 |
| Mixing의 block 소스 | fulfill 블록 해시 | **비콘 제출 블록 AppHash** | Last-Revealer 문제 해소 |
| target_round 계산 | latest_round 기반 | **block_time 기반 drand round 계산** | 릴레이어 지연 시 타이밍 공격 해소 |
| min_future_rounds | 4 (2분, 30초 기준) | **30 (90초, quicknet 3초 기준)** | quicknet 실제 주기 반영 |
| per-requester 제한 | 없음 | **max_requests_per_requester = 10** | DoS 방지 |
| 만료 메커니즘 | 향후 과제 | **request_timeout_blocks = 5760** | PENDING 고착 방지 |
| drand_period_seconds | 30 | **3** | quicknet 실제 주기 |
| 릴레이어 인센티브 | 없음 | **Request Fee → 릴레이어 보상** | 가스비 보상 + 제출 동기 |

### 2.4 설계 근거: MsgFulfillRandomness 삭제

**문제점 (3단계)**:

1. **Fulfiller Bot 필수**: 별도의 bot이 pending 요청을 모니터링하고 `MsgFulfillRandomness`를 제출해야 함. bot 부재 시 PENDING 영구 대기
2. **Gas Griefing**: fulfiller가 가스비를 부담하지만 보상 없음. 악의적 요청자가 대량 요청으로 fulfiller에게 가스비 소모 공격 가능
3. **Last-Revealer 문제**: fulfiller가 fulfill TX의 `block_hash`를 mixing에 포함하므로, 블록 제안자가 fulfill TX를 선택적으로 포함/제외하여 결과에 영향 가능

**해결 (2단계)**:

- **EndBlocker 자동 이행**: 비콘 도착 시 프로토콜이 자동으로 pending 요청을 fulfill. Fulfiller Bot 불필요
- **릴레이어 가스 고정**: `MsgSubmitBeacon`은 비콘 저장만 수행. pending 요청 수와 무관하게 가스비 일정
- **AppHash 사용**: `ctx.HeaderInfo().AppHash`는 이전 블록의 AppHash이므로 현재 블록 제안자가 조작 불가
- **Request Fee**: 유저가 요청 시 수수료 납부 → 비콘 제출 릴레이어에게 보상 지급

---

## 3. 용어 설명

| 용어 | 설명 |
|------|------|
| `commit_hash` | 유저가 랜덤 요청 시 제출하는 32바이트 해시. `SHA256(user_seed)` 형태. 비콘이 아직 없는 시점에 제출되므로 결과 예측 불가. Mixing의 "유저 엔트로피" 역할 |
| `MsgSubmitBeacon` | 릴레이어 봇이 drand 네트워크에서 받은 비콘을 온체인에 제출하는 트랜잭션. round, randomness, signature 3개 필드 포함. BLS 서명 검증 통과 시만 저장 |
| `beacon.randomness` | drand 분산 네트워크가 생성한 32바이트 랜덤값. BLS 서명으로 검증 가능. `SHA256(signature)` 관계. 누구도 예측/조작 불가 |
| `AppHash` | Cosmos SDK에서 각 블록 실행 후 계산되는 앱 상태 해시. 이전 블록의 모든 상태 변경을 반영. 메시지 실행 중 `ctx.HeaderInfo().AppHash`로 접근 — **이전 블록의 AppHash**이므로 현재 블록 제안자가 조작 불가 |
| `target_round` | 요청 시 계산되는 미래 drand 라운드 번호. `current_drand_round + min_future_rounds`. 이 라운드의 비콘이 도착해야 fulfill 가능 |
| `request_fee` | 유저가 랜덤 요청 시 납부하는 수수료. 비콘 제출 릴레이어에 대한 가스비 보상 역할. 모듈 계정에 에스크로 후 fulfill 시 릴레이어에게 지급, 만료 시 요청자에게 환불 |
| `Relayer Bot` | drand 네트워크에서 비콘을 수신하여 `MsgSubmitBeacon`으로 온체인에 제출하는 외부 프로세스. Permissionless — 누구나 운영 가능 |

---

## 4. 리스크 완화 3계층

### 4.1 BLS 서명 검증 (가짜 비콘 차단)

**리스크**: `beacon_verification_enabled = false` (stub 모드)에서는 누구나 임의의 randomness 값을 제출 가능. 공격자가 원하는 랜덤값을 주입할 수 있음.

**완화**: drand quicknet의 BLS12-381 서명을 온체인에서 실제 검증.

```
Input: round, randomness (hex), signature (hex), public_key (G2 hex)

1. sig_bytes = hex_decode(signature)
2. expected_randomness = SHA256(sig_bytes)
3. assert randomness == hex(expected_randomness)  // 랜덤값 일관성 검증

4. msg = SHA256(big_endian_uint64(round))          // Unchained: 이전 서명 불필요
5. scheme = bls-unchained-g1-rfc9380
6. verify(public_key_G2, sig_G1, msg)              // BLS12-381 pairing check
```

**의존성**: `github.com/drand/kyber`, `github.com/drand/kyber-bls12381`

**운영 전략**:
- stub 모드(`beacon_verification_enabled = false`)는 테스트넷까지만 사용
- 메인넷 전 `beacon_verification_enabled = true` 전환 필수
- drand 공개키를 거버넌스 파라미터(`drand_public_key`)로 관리하여 키 교체 대응

### 4.2 Commit-Reveal + Future Beacon (Front-running 방지)

**리스크**: 비콘을 직접 사용하면 블록 제안자가 TX 순서를 조작하거나, mempool에서 비콘 TX를 관찰하여 결과를 예측할 수 있음.

**완화**: 요청 시점에 아직 존재하지 않는 미래 비콘을 할당하여 결과 예측 불가.

```
Block N:   유저가 MsgRequestRandomness(commit_hash) 제출
           target_round = current_drand_round + min_future_rounds (30)
           → 이 시점에 target_round 비콘은 아직 생성되지 않음
           → 결과 예측 불가

Block N+K: Relayer가 target_round 비콘 제출
           → 이 시점에 commit_hash는 이미 확정
           → 비콘을 보고 선택을 바꿀 수 없음
```

**target_round 계산 (block_time 기반)**:

```
current_drand_round = (block_time - drand_genesis_time) / drand_period_seconds
target_round = current_drand_round + min_future_rounds
```

기존 `latest_round` 기반 계산의 문제: 릴레이어 지연으로 `latest_round`가 실제보다 뒤처지면 "미래"라고 할당된 라운드가 이미 drand 네트워크에 존재할 수 있음. `block_time` 기반 계산으로 이 문제를 해소한다.

### 4.3 Multi-source Mixing (단일 소스 조작 불가)

**리스크**: 단일 랜덤 소스만 사용하면 해당 소스를 통제하는 주체가 결과를 조작할 수 있음.

**완화**: 3개 독립 소스를 SHA256으로 혼합.

```
result[i] = SHA256(beacon.randomness || AppHash || commit_hash || uint8(i))
```

| 입력 | 확정 시점 | 제공자 | 조작 가능성 |
|------|----------|--------|-----------|
| `beacon.randomness` | 비콘 제출 시 | drand 분산 네트워크 | BLS 검증으로 조작 불가 |
| `AppHash` | 비콘 제출 블록의 이전 블록 | CometBFT 합의 | 이미 확정된 과거 블록값 |
| `commit_hash` | 유저 요청 시 | 요청자 | 이미 확정된 값 |

3개 입력 중 어느 하나라도 독립적으로 통제하면 결과를 조작할 수 있어야 하지만, 각 입력이 서로 다른 시점에 확정되므로 어떤 참여자도 3개를 동시에 통제할 수 없다.

---

## 5. Mixing 보안 속성 분석

### 5.1 시간순 보안 분석

| 시점 | 확정된 입력 | 미확정 입력 | 공격 가능성 |
|------|-----------|-----------|-----------|
| 유저 요청 (Block N) | commit_hash | beacon.randomness, AppHash | beacon이 존재하지 않으므로 결과 예측 불가 |
| 비콘 제출 (Block N+K) | commit_hash, beacon.randomness, AppHash | (없음) | commit_hash 이미 확정. 비콘을 보고 선택 변경 불가 |
| EndBlocker 이행 (Block N+K) | 3개 모두 | (없음) | 결과 결정론적, 조작 불가 |

### 5.2 공격 시나리오별 분석

**공격 1: 블록 제안자 TX 순서 조작**
- 기존(직접 비콘 사용): 비콘 TX와 컨트랙트 TX 순서를 조작하여 유리한 결과 유도 가능
- 현재(Commit-Reveal): commit_hash가 비콘 도착 전에 확정. TX 순서와 무관하게 결과 동일. **무력화**

**공격 2: Mempool 관찰 (Front-running)**
- 기존: 비콘 TX가 mempool에 노출되면 결과 예측 가능
- 현재: target_round 비콘이 아직 drand 네트워크에서 생성되지 않은 상태. **무력화**

**공격 3: 릴레이어 선택적 비콘 제출**
- 기존: 릴레이어가 유리한 라운드만 제출 가능
- 현재: commit_hash가 이미 확정되어 있으므로, 특정 라운드를 건너뛰어도 다음 라운드에 비콘 제출 시 결과가 다를 뿐, 릴레이어가 "유리한" 결과를 선택할 수 없음. **완화**

**공격 4: 블록 제안자 AppHash 조작 (Last-Revealer)**
- 기존(block_hash 사용): fulfill TX 포함 여부로 block_hash 변경 가능
- 현재(AppHash 사용): AppHash는 이전 블록의 상태 해시. 현재 블록 제안자가 변경 불가. **무력화**

### 5.3 한계 및 잔여 리스크

| 잔여 리스크 | 심각도 | 설명 | 완화 |
|------------|--------|------|------|
| drand 네트워크 장애 | 낮음 | drand 네트워크 전체 장애 시 비콘 생성 중단 | drand는 글로벌 분산 네트워크, 역사적으로 안정적 |
| 릴레이어 전원 다운 | 중간 | 비콘 미제출 → PENDING 영구 대기 | request_timeout_blocks로 만료 + fee 환불, 복수 독립 릴레이어 운영 |
| commit_hash 엔트로피 부족 | 낮음 | 유저가 예측 가능한 user_seed 사용 | beacon.randomness가 주요 엔트로피 소스이므로 영향 제한적 |

---

## 6. 메시지 정의

### 6.1 MsgSubmitBeacon (기존 + BLS 검증 추가)

drand 비콘을 온체인에 제출한다. BLS 검증 활성화 시 서명 검증을 통과해야 저장된다.

```protobuf
message MsgSubmitBeacon {
  string submitter = 1;   // 제출자 주소 (릴레이어)
  uint64 round = 2;       // drand 라운드 번호
  string randomness = 3;  // 32바이트 랜덤값 (hex)
  string signature = 4;   // BLS 서명 (hex)
}
```

**처리 흐름**:

```
1. 라운드 유효성 검증 (round > 0)
2. 랜덤값 형식 검증 (64자 hex = 32바이트)
3. 서명 형식 검증 (hex 문자열)
4. 중복 라운드 검사
5. 라운드 시간 검증:
   expected_time = drand_genesis_time + (round * drand_period_seconds)
   assert expected_time ≤ block_time
6. BLS 서명 검증 (beacon_verification_enabled = true 일 때):
   a. randomness == SHA256(signature) 검증
   b. message = SHA256(round_to_big_endian_bytes)
   c. drand quicknet 스킴으로 BLS12-381 G1 서명 검증
   d. 공개키: params.drand_public_key (G2 포인트)
7. Beacon 저장 (submitter 필드에 릴레이어 주소 기록)
8. LatestRound 업데이트 (새 라운드가 더 큰 경우)
9. 이벤트 발행 (beacon_submitted)
```

### 6.2 MsgRequestRandomness (신규)

유저가 랜덤 결과를 요청한다. 1회 서명으로 commit_hash를 제출하고, 비콘 도착 후 EndBlocker가 자동으로 결과를 이행한다.

```protobuf
message MsgRequestRandomness {
  string requester = 1;                   // 요청자 주소
  string commit_hash = 2;                 // SHA256(user_seed), 32바이트 hex
  uint32 num_words = 3;                   // 요청 랜덤 워드 수 (1-10)
  string callback_data = 4;              // 임의 콜백 데이터 (max 256 bytes hex)
  cosmos.base.v1beta1.Coin request_fee = 5; // 요청 수수료 (>= min_request_fee)
}

message MsgRequestRandomnessResponse {
  uint64 request_id = 1;
  uint64 target_round = 2;
}
```

**처리 흐름**:

```
 1. commit_reveal_enabled 파라미터 확인 (비활성 시 ErrCommitRevealDisabled)
 2. commit_hash 형식 검증 (64 hex chars = 32바이트)
 3. num_words 범위 검증 (1 ≤ num_words ≤ 10)
 4. callback_data 길이 검증 (max 512 hex chars = 256 bytes)
 5. request_fee 검증: fee >= min_request_fee (ErrInsufficientRequestFee)
 6. 전체 pending < max_pending_requests 확인 (ErrTooManyPendingRequests)
 7. 주소당 pending < max_requests_per_requester 확인 (ErrRequesterLimitExceeded)
 8. target_round 계산 (block_time 기반):
    current_drand_round = (block_time - drand_genesis_time) / drand_period_seconds
    target_round = current_drand_round + min_future_rounds
    assert target_round > latest_round
 9. request_fee를 requester → randomness 모듈 계정으로 전송
    (bank.SendCoinsFromAccountToModule)
10. request_id 할당 (RequestSequence auto-increment)
11. RandomnessRequest 저장 (status = PENDING, request_fee 포함)
12. 인덱스 업데이트:
    - PendingByRound[(target_round, request_id)] 추가
    - RequestsByRequester[(requester, request_id)] 추가
    - PendingRequestCount 증가
13. 이벤트 발행 (randomness_requested, fee 포함)
14. 응답에 request_id, target_round 반환
```

### 6.3 MsgUpdateParams (기존)

거버넌스를 통한 파라미터 변경. signer = authority (x/gov).

---

## 7. Query 정의

### 비콘 쿼리

| 쿼리 | REST 엔드포인트 | 설명 |
|------|---------------|------|
| `Params` | `/seocheon/randomness/v1/params` | 모듈 파라미터 조회 |
| `LatestBeacon` | `/seocheon/randomness/v1/beacons/latest` | 최신 비콘 조회 |
| `Beacon(round)` | `/seocheon/randomness/v1/beacons/{round}` | 특정 라운드 비콘 조회 |
| `Beacons(pagination)` | `/seocheon/randomness/v1/beacons` | 비콘 목록 (페이지네이션) |

### Commit-Reveal 쿼리

| 쿼리 | REST 엔드포인트 | 설명 |
|------|---------------|------|
| `RandomnessRequest(request_id)` | `/seocheon/randomness/v1/requests/{request_id}` | 특정 요청 조회 (상태, 결과 포함) |
| `PendingRequests(pagination)` | `/seocheon/randomness/v1/requests/pending` | PENDING 요청 목록 (릴레이어 수익성 판단용) |
| `RequestsByRequester(requester)` | `/seocheon/randomness/v1/requests/by_requester/{requester}` | 특정 주소의 요청 목록 |

**PendingByRound 활용**: 릴레이어는 `PendingRequests` 쿼리로 수익성 있는 라운드를 판단하여 선택적으로 비콘을 제출할 수 있다. 1개 비콘이 N개 요청을 fulfill하므로, fee 합산이 가스비를 초과하는 라운드만 제출하는 전략이 가능하다.

---

## 8. 데이터 구조

### 8.1 Beacon (기존)

```protobuf
message Beacon {
  uint64 round = 1;        // drand 라운드 번호
  string randomness = 2;   // 32바이트 랜덤값 (hex)
  string signature = 3;    // BLS 서명 (hex)
  int64  submitted_at = 4; // 저장 블록 높이
  string submitter = 5;    // 제출자 주소 (릴레이어)
}
```

`submitter` 필드는 Request Fee 메커니즘에서 릴레이어 보상 수신 주소로 사용된다.

### 8.2 RandomnessRequest (신규)

```protobuf
message RandomnessRequest {
  uint64 request_id = 1;                          // 고유 요청 ID (auto-increment)
  string requester = 2;                           // 요청자 주소
  string commit_hash = 3;                         // 유저 시드 해시 (32바이트 hex)
  uint32 num_words = 4;                           // 요청 랜덤 워드 수 (1-10)
  string callback_data = 5;                      // 임의 콜백 데이터 (max 256 bytes hex)
  uint64 target_round = 6;                       // 할당된 미래 drand 라운드
  RandomnessRequestStatus status = 7;             // 요청 상태
  int64  created_at = 8;                          // 생성 블록 높이
  int64  fulfilled_at = 9;                        // 이행 블록 높이
  string result = 10;                             // 최종 랜덤 (hex, num_words * 32바이트)
  string beacon_app_hash = 11;                   // mixing에 사용된 AppHash (hex)
  cosmos.base.v1beta1.Coin request_fee = 12;     // 에스크로된 요청 수수료
}
```

### 8.3 RandomnessRequestStatus (신규)

```protobuf
enum RandomnessRequestStatus {
  RANDOMNESS_REQUEST_STATUS_UNSPECIFIED = 0;
  RANDOMNESS_REQUEST_STATUS_PENDING = 1;     // 비콘 대기 중
  RANDOMNESS_REQUEST_STATUS_FULFILLED = 2;   // 이행 완료
  RANDOMNESS_REQUEST_STATUS_EXPIRED = 3;     // 타임아웃 만료
}
```

상태 전이: `PENDING → FULFILLED` (비콘 도착 시 EndBlocker) 또는 `PENDING → EXPIRED` (타임아웃 도달 시 EndBlocker)

---

## 9. 파라미터 (확장)

### 기존 파라미터 (필드 1~5)

| # | 파라미터 | 타입 | 기본값 | 설명 |
|---|---------|------|--------|------|
| 1 | `beacon_verification_enabled` | bool | false | BLS 서명 검증 활성화. 메인넷 전 true 전환 필수 |
| 2 | `max_beacon_age_blocks` | uint64 | 17280 | 비콘 최대 수용 연령 (블록) |
| 3 | `drand_genesis_time` | uint64 | 1692803367 | drand quicknet genesis 시간 (Unix) |
| 4 | `drand_period_seconds` | uint64 | 3 | drand quicknet 비콘 발행 주기 (초) |
| 5 | `drand_chain_hash` | string | (quicknet hash) | drand 체인 해시 |

### 신규 파라미터 (필드 6~15)

| # | 파라미터 | 타입 | 기본값 | 설명 |
|---|---------|------|--------|------|
| 6 | `drand_public_key` | string | (quicknet pubkey hex) | drand 공개키 (BLS12-381 G2, hex) |
| 7 | `drand_scheme_id` | string | `bls-unchained-g1-rfc9380` | drand 서명 스킴 ID |
| 8 | `commit_reveal_enabled` | bool | false | Commit-Reveal 기능 활성화 |
| 9 | `min_future_rounds` | uint64 | 30 | 최소 미래 라운드 수 (30 × 3초 = 90초) |
| 10 | `request_timeout_blocks` | uint64 | 5760 | PENDING 만료 블록 수 (~8시간, 5초/블록) |
| 11 | `request_pruning_blocks` | uint64 | 17280 | 이행/만료 후 프루닝 블록 수 (~1에포크) |
| 12 | `max_pending_requests` | uint64 | 1000 | 전체 최대 동시 PENDING 요청 수 |
| 13 | `max_requests_per_requester` | uint64 | 10 | 주소당 최대 PENDING 요청 수 |
| 14 | `max_fulfills_per_block` | uint64 | 50 | EndBlocker 블록당 최대 fulfill 수 |
| 15 | `min_request_fee` | Coin | 10000000uppyeo | 최소 요청 수수료 (≈ 0.001 KKOT) |

**기본값 변경 사항**:
- `drand_genesis_time`: 1595431050 → **1692803367** (quicknet genesis)
- `drand_period_seconds`: 30 → **3** (quicknet 실제 주기)

---

## 10. 스토어 구조 (확장)

### 비콘 스토어 (기존)

| 프리픽스 | 컬렉션 | 타입 | 설명 |
|---------|--------|------|------|
| 0x01 | Beacons | Map[uint64, Beacon] | 라운드 → Beacon |
| 0x02 | LatestRound | Item[uint64] | 최근 라운드 번호 |
| 0x03 | Params | Item[Params] | 모듈 파라미터 |

### Commit-Reveal 스토어 (신규)

| 프리픽스 | 컬렉션 | 타입 | 설명 |
|---------|--------|------|------|
| 0x1E (30) | RequestSequence | Sequence | auto-increment 요청 ID |
| 0x1F (31) | Requests | Map[uint64, RandomnessRequest] | 요청 ID → 요청 데이터 |
| 0x20 (32) | PendingByRound | KeySet[(target_round, request_id)] | 비콘 라운드별 pending 인덱스 |
| 0x21 (33) | RequestsByRequester | KeySet[(requester, request_id)] | 요청자별 인덱스 |
| 0x22 (34) | PendingRequestCount | Item[uint64] | 현재 PENDING 총 수 |
| 0x23 (35) | FulfilledBlockIndex | KeySet[(block_height, request_id)] | 이행/만료 블록별 인덱스 (프루닝용) |

---

## 11. EndBlocker 처리 흐름

EndBlocker는 3개 페이즈를 순차 실행한다.

### 11.1 Phase 1: 자동 Fulfill + Fee 지급

비콘이 존재하는 pending 요청을 자동으로 fulfill하고, 릴레이어에게 fee를 지급한다.

```
EndBlocker Phase 1: 자동 Fulfill

fulfilled_count = 0

// PendingByRound의 가장 작은 target_round부터 순회
for target_round in PendingByRound (ascending order):
  beacon = Beacons[target_round]
  if beacon not found:
    break  // 라운드는 순차적이므로 이후 라운드도 없음

  relayer_reward = sdk.Coins{}  // 이 비콘으로 fulfill된 모든 fee 합산

  for (target_round, request_id) in PendingByRound[target_round]:
    if fulfilled_count >= max_fulfills_per_block:
      break

    request = Requests[request_id]
    app_hash = ctx.HeaderInfo().AppHash  // 이전 블록의 AppHash

    // Multi-source Mixing
    result = []
    for i in 0..request.num_words:
      result[i] = SHA256(
        beacon.randomness ||  // drand 비콘 (BLS 검증됨)
        app_hash           ||  // 이전 블록 AppHash (조작 불가)
        request.commit_hash||  // 유저 시드 해시 (확정됨)
        uint8(i)               // 워드 인덱스
      )

    // 상태 업데이트
    request.status = FULFILLED
    request.result = hex(concat(result))
    request.fulfilled_at = current_height
    request.beacon_app_hash = hex(app_hash)

    // 인덱스 업데이트
    PendingByRound[(target_round, request_id)] 제거
    FulfilledBlockIndex[(current_height, request_id)] 추가
    PendingRequestCount 감소

    relayer_reward += request.request_fee
    fulfilled_count++

    이벤트 발행 (randomness_fulfilled)

  // 릴레이어 보상 일괄 전송
  if relayer_reward.IsAllPositive():
    bank.SendCoinsFromModuleToAccount(
      ctx, "randomness", beacon.submitter, relayer_reward
    )
```

**성능 고려**: `max_fulfills_per_block = 50`으로 블록당 fulfill 수를 제한하여 EndBlocker 처리 시간을 일정하게 유지한다. 미처리된 요청은 다음 블록의 EndBlocker에서 계속 처리된다.

### 11.2 Phase 2: PENDING 만료 + Fee 환불

타임아웃에 도달한 pending 요청을 만료 처리하고, 수수료를 환불한다.

```
EndBlocker Phase 2: 만료 처리

for each pending request:
  if request.created_at + request_timeout_blocks < current_height:
    request.status = EXPIRED
    request.fulfilled_at = current_height  // 만료 처리 시점 기록

    // 수수료 환불
    if request.request_fee.IsPositive():
      bank.SendCoinsFromModuleToAccount(
        ctx, "randomness", request.requester, request.request_fee
      )

    // 인덱스 업데이트
    PendingByRound[(target_round, request_id)] 제거
    FulfilledBlockIndex[(current_height, request_id)] 추가
    PendingRequestCount 감소

    이벤트 발행 (randomness_expired, refund_amount)
```

### 11.3 Phase 3: 프루닝

이행/만료 완료 후 일정 블록이 경과한 요청 데이터를 삭제하여 스토어 팽창을 방지한다.

```
EndBlocker Phase 3: 프루닝

for (block_height, request_id) in FulfilledBlockIndex:
  if block_height + request_pruning_blocks < current_height:
    request = Requests[request_id]
    Requests[request_id] 삭제
    RequestsByRequester[(request.requester, request_id)] 삭제
    FulfilledBlockIndex[(block_height, request_id)] 삭제
```

---

## 12. Request Fee 메커니즘 (릴레이어 가스비 보상)

릴레이어가 `MsgSubmitBeacon`을 제출할 때 가스비를 부담한다. Request Fee는 이 가스비를 보상하고 비콘 제출 동기를 부여하는 메커니즘이다.

### 12.1 자금 흐름

```
유저 요청 시: requester ──[request_fee]──► randomness 모듈 계정 (에스크로)
비콘 이행 시: randomness 모듈 계정 ──[request_fee]──► beacon.submitter (릴레이어)
요청 만료 시: randomness 모듈 계정 ──[request_fee]──► requester (환불)
```

| 시점 | 행위 | 자금 흐름 |
|------|------|----------|
| `MsgRequestRandomness` | 유저가 fee 납부 | requester → module account |
| EndBlocker fulfill | 비콘 제출자에게 보상 | module account → beacon.submitter |
| EndBlocker expire | 만료 시 환불 | module account → requester |

### 12.2 설계 포인트

- **`min_request_fee` 거버넌스 파라미터**: 최소 수수료를 거버넌스로 제어 (기본: 10000000uppyeo ≈ 0.001 KKOT)
- **수익성 판단**: 릴레이어는 `PendingRequests` 쿼리로 각 라운드의 pending 요청 수와 fee 합계를 확인하여 수익성 있는 라운드만 선택적으로 제출 가능
- **N:1 관계**: 1개 비콘이 N개 요청을 fulfill → 릴레이어는 N개 fee를 합산 수령
- **모듈 계정**: randomness 모듈 계정이 에스크로 역할 (`app_config.go`의 `moduleAccPerms`에 등록 필요)

### 12.3 경제성 분석

```
릴레이어 수익 = Σ(fulfill된 요청의 request_fee) - MsgSubmitBeacon 가스비

예시:
- 1개 비콘으로 5개 요청 fulfill
- 각 요청 fee: 10000000uppyeo
- 릴레이어 수령: 50000000uppyeo (≈ 0.005 KKOT)
- MsgSubmitBeacon 가스비: ~5000000uppyeo (추정)
- 순이익: ~45000000uppyeo
```

pending 요청이 없는 라운드는 릴레이어가 비콘을 제출할 인센티브가 없다. 이는 의도된 동작으로, 불필요한 비콘 저장을 줄여 스토어 효율을 높인다.

---

## 13. 에러 코드

### 비콘 에러 (1300~1309)

| 코드 | 이름 | 설명 |
|------|------|------|
| 1300 | ErrInvalidSigner | UpdateParams 권한 없음 |
| 1301 | ErrBeaconNotFound | 비콘 없음 |
| 1302 | ErrDuplicateBeacon | 중복 라운드 |
| 1303 | ErrInvalidRound | 잘못된 라운드 번호 |
| 1304 | ErrInvalidRandomness | 잘못된 랜덤값 형식 |
| 1305 | ErrInvalidSignature | 잘못된 서명 형식 |
| 1306 | ErrBeaconTooOld | 너무 오래된 비콘 |
| 1307 | ErrVerificationDisabled | 검증 비활성화 상태 |
| 1308 | ErrVerificationFailed | BLS 서명 검증 실패 |
| 1309 | ErrNoBeaconAvailable | 저장된 비콘 없음 |

### Commit-Reveal 에러 (1310~1320)

| 코드 | 이름 | 설명 |
|------|------|------|
| 1310 | ErrCommitRevealDisabled | commit-reveal 기능 비활성화 |
| 1311 | ErrInvalidCommitHash | commit_hash 형식 불량 (64 hex chars 아님) |
| 1312 | ErrInvalidNumWords | num_words 범위 초과 (1~10 아님) |
| 1313 | ErrCallbackDataTooLong | callback_data 길이 초과 (>256 bytes) |
| 1314 | ErrInvalidCallbackData | callback_data hex 형식 불량 |
| 1315 | ErrTooManyPendingRequests | 전체 pending 요청 수 상한 도달 |
| 1316 | ErrRequestNotFound | 요청 ID 없음 |
| 1317 | ErrRequestNotPending | 요청이 PENDING 상태가 아님 |
| 1318 | ErrBeaconNotYetAvailable | target_round 비콘 미도착 |
| 1319 | ErrRequesterLimitExceeded | 주소당 pending 요청 수 상한 도달 |
| 1320 | ErrInsufficientRequestFee | 요청 수수료 부족 (< min_request_fee) |

---

## 14. 이벤트

| 이벤트 | 속성 | 설명 |
|--------|------|------|
| `beacon_submitted` | round, randomness, submitter, block_height | 비콘 저장 성공 |
| `randomness_requested` | request_id, requester, commit_hash, target_round, num_words, request_fee | 랜덤 요청 커밋 |
| `randomness_fulfilled` | request_id, target_round, result_hash, beacon_submitter, fee_paid | 랜덤 이행 완료 (EndBlocker) |
| `randomness_expired` | request_id, target_round, refund_amount | PENDING 요청 만료 + 수수료 환불 |

---

## 15. CosmWasm 통합 가이드

### 15.1 제약사항

- CosmWasm 컨트랙트는 `MsgRequestRandomness`를 직접 발행할 수 없다 (Stargate Message 제한)
- 대신 **Stargate Query**로 결과를 조회하는 패턴을 사용
- 랜덤 요청은 EOA(외부 소유 계정) 또는 중개 컨트랙트를 통해 제출

### 15.2 사용 패턴

**패턴 A: 고보안 (Commit-Reveal)**

외부 트리거가 `MsgRequestRandomness`를 제출하고, 컨트랙트는 결과를 쿼리한다.

```rust
// 1. 외부에서 MsgRequestRandomness 제출 (commit_hash = SHA256(game_id + salt))
//    → request_id 획득

// 2. 컨트랙트에서 결과 조회
let request: QueryRandomnessRequestResponse = deps.querier.query(
    &QueryRequest::Stargate {
        path: "/seocheon.randomness.v1.Query/RandomnessRequest".to_string(),
        data: QueryRandomnessRequestRequest { request_id }.encode_to_vec().into(),
    }
)?;

if request.status == FULFILLED {
    let random_bytes = hex::decode(request.result)?;
    // random_bytes를 카드 셔플, 추첨 등에 사용
}
```

**패턴 B: 저보안 (직접 비콘 사용)**

단순한 용도에서 비콘을 직접 조회한다. Front-running 리스크가 허용되는 경우에만 사용.

```rust
// 최신 비콘 직접 조회
let beacon: QueryLatestBeaconResponse = deps.querier.query(
    &QueryRequest::Stargate {
        path: "/seocheon.randomness.v1.Query/LatestBeacon".to_string(),
        data: QueryLatestBeaconRequest {}.encode_to_vec().into(),
    }
)?;

// submitted_at 검증 (stale 비콘 거부)
let current_height = env.block.height;
if current_height - beacon.submitted_at > MAX_BEACON_AGE {
    return Err(StdError::generic_err("beacon too old"));
}
```

**권장 컨트랙트 흐름 (고보안)**:

```
1. ExecuteMsg::StartGame { user_seed }
   → commit_hash = SHA256(user_seed)
   → 외부 서비스가 MsgRequestRandomness 발행
   → request_id를 게임 상태에 매핑하여 저장

2. QueryMsg::GetGameResult { game_id }
   → 저장된 request_id로 RandomnessRequest 쿼리
   → status == FULFILLED이면 result로 게임 결과 계산
   → status == PENDING이면 "대기 중" 반환
   → status == EXPIRED이면 "만료" 반환 (재요청 필요)
```

---

## 16. 수정 전후 설계 변경 요약

| # | 항목 | 수정 전 | 수정 후 | 근거 |
|---|------|---------|---------|------|
| 1 | Fulfill 방식 | MsgFulfillRandomness (별도 TX) | EndBlocker 자동 | Fulfiller Bot 불필요, Gas Griefing 방지 |
| 2 | Mixing block 소스 | fulfill 블록 해시 | 비콘 제출 블록 AppHash | Last-Revealer 문제 해소 |
| 3 | target_round 계산 | latest_round + min_future_rounds | block_time 기반 drand round + min_future_rounds | 릴레이어 지연 시 타이밍 공격 해소 |
| 4 | min_future_rounds | 4 (2분, 30초 기준) | 30 (90초, quicknet 3초 기준) | quicknet 실제 주기 반영 |
| 5 | drand_period_seconds | 30 | 3 | quicknet 실제 주기 |
| 6 | drand_genesis_time | 1595431050 | 1692803367 | quicknet genesis 시간 |
| 7 | per-requester 제한 | 없음 | max_requests_per_requester = 10 | DoS 방지 |
| 8 | 만료 메커니즘 | 향후 과제 | request_timeout_blocks = 5760 (~8시간) | PENDING 고착 방지 |
| 9 | 릴레이어 인센티브 | 없음 | Request Fee → 릴레이어 보상 | 가스비 보상 + 제출 동기 |
| 10 | max_fulfills_per_block | - | 50 | EndBlocker 처리 시간 제한 |
| 11 | RandomnessRequest.fulfiller | 존재 | 삭제 | EndBlocker 처리로 fulfiller 개념 불필요 |
| 12 | RandomnessRequest.beacon_round_block_hash | fulfill 블록 해시 | beacon_app_hash (AppHash) | Last-Revealer 해소 |
| 13 | RandomnessRequest.request_fee | - | 추가 | 에스크로 기록 |

---

## 17. 구현 대상 파일 목록 (코드 구현 시 참조)

### Proto 파일 (5파일)

| 파일 | 변경 | 내용 |
|------|------|------|
| `params.proto` | 수정 | 필드 6~15 추가 (drand_public_key ~ min_request_fee) |
| `randomness.proto` | 수정 | RandomnessRequestStatus enum, RandomnessRequest message 추가 |
| `tx.proto` | 수정 | MsgRequestRandomness 추가 (MsgFulfillRandomness 삭제) |
| `query.proto` | 수정 | RandomnessRequest, PendingRequests, RequestsByRequester 쿼리 추가 |
| `genesis.proto` | 수정 | randomness_requests, next_request_id 필드 추가 |

### Types 파일 (5파일)

| 파일 | 변경 | 내용 |
|------|------|------|
| `types/keys.go` | 수정 | 프리픽스 30~35 추가 |
| `types/errors.go` | 수정 | 에러 코드 1319~1320 추가 |
| `types/events.go` | 수정 | randomness_requested, randomness_fulfilled, randomness_expired 이벤트 추가 |
| `types/codec.go` | 수정 | MsgRequestRandomness 등록 (MsgFulfillRandomness 삭제) |
| `types/params.go` | 수정 | 기본값 업데이트 (drand_genesis_time, drand_period_seconds 등) |

### Keeper 파일 (5파일)

| 파일 | 변경 | 내용 |
|------|------|------|
| `keeper/keeper.go` | 수정 | 6개 컬렉션 추가, BankKeeper 의존성 추가 |
| `keeper/msg_server_submit_beacon.go` | 수정 | BLS 검증 로직 + 라운드 시간 검증 추가 |
| `keeper/beacon_verification.go` | **신규** | VerifyBeacon() BLS 검증 함수 |
| `keeper/msg_server_request_randomness.go` | **신규** | MsgRequestRandomness 핸들러 |
| `keeper/abci.go` | **신규** | EndBlocker (자동 fulfill + 만료 + 프루닝) |

### Query/Module 파일 (3파일)

| 파일 | 변경 | 내용 |
|------|------|------|
| `keeper/query_request.go` | **신규** | RandomnessRequest, PendingRequests, RequestsByRequester 쿼리 핸들러 |
| `keeper/genesis.go` | 수정 | InitGenesis/ExportGenesis 확장 |
| `module/module.go` | 수정 | ConsensusVersion 1→2, EndBlocker 연결 |

### 기타 (3파일)

| 파일 | 변경 | 내용 |
|------|------|------|
| `go.mod` | 수정 | drand BLS 라이브러리 추가 (drand/kyber, drand/kyber-bls12381) |
| `app/app_config.go` | 수정 | randomness 모듈 계정 권한 추가 (moduleAccPerms) |
| `module/autocli.go` | 수정 | Query CLI + Tx CLI 업데이트 |

**총계**: 신규 4파일, 수정 16파일
