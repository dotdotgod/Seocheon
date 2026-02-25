# 체인 업그레이드 전략

> **담당**: 체인 코어 개발자 / DevOps
> **관련 문서**: [개요](01_overview.md) · [노드 모듈](03_node_module.md) · [Activity Protocol](04_activity_protocol.md) · [구현 가이드](09_implementation.md) · [재단 전략](../foundation_strategy.md) · [Circuit Breaker](11_circuit_breaker.md) · [전체 목차](README.md)

---

## 1. 개요

Seocheon은 Cosmos SDK 기반 블록체인으로, Activity Protocol 버전 업그레이드, 거버넌스 파라미터 도입, Cosmos SDK 버전 업그레이드, 보안 패치 등을 위한 체인 업그레이드 전략을 정의한다. 체인은 영구적 인프라로서 업그레이드 과정에서도 안정성과 연속성을 보장한다.

### 업그레이드 원칙

```
① 거버넌스 우선: 모든 업그레이드는 거버넌스 투표로 승인
② 예측 가능성: 업그레이드 높이를 사전 공지, 밸리데이터 준비 시간 보장
③ 자동화: cosmovisor로 바이너리 교체 자동화
④ 롤백 가능성: 실패 시 이전 상태로 복구 가능한 절차 마련
⑤ 테스트 우선: 메인넷 적용 전 시뮬레이션과 테스트넷 검증 필수
```

---

## 2. Cosmos SDK x/upgrade 모듈 활용

### UpgradeProposal 거버넌스 플로우

Cosmos SDK의 `x/upgrade` 모듈은 체인 업그레이드를 거버넌스 프로세스로 관리한다. Seocheon은 이 표준 메커니즘을 그대로 활용한다.

```
업그레이드 제안 → 투표 → 승인 → 업그레이드 높이 도달 → 체인 정지 → 바이너리 교체 → 재시작

상세 플로우:
  [1] 재단 또는 커뮤니티가 SoftwareUpgradeProposal 제출
      ├── name: 업그레이드 이름 (예: "v2.0.0-activity-v2")
      ├── height: 업그레이드 실행 블록 높이
      └── info: 업그레이드 정보 (바이너리 다운로드 URL, 변경사항 등)

  [2] 거버넌스 투표 (voting_period: 172,800 블록, ~10일)
      ├── 밸리데이터 + 위임자 투표
      └── 통과 조건: Cosmos SDK 기본 (quorum 33.4%, threshold 50%)

  [3] 투표 통과 → x/upgrade에 업그레이드 플랜 등록
      └── 체인이 지정된 높이에서 자동 정지

  [4] 업그레이드 높이 도달
      ├── BeginBlocker에서 업그레이드 플랜 감지
      ├── 체인 정지 (panic: "UPGRADE NEEDED")
      └── cosmovisor가 새 바이너리로 자동 교체

  [5] 새 바이너리로 재시작
      ├── 상태 마이그레이션 실행 (있는 경우)
      └── 정상 블록 생성 재개
```

### 업그레이드 높이(height) 설정

업그레이드 높이는 밸리데이터가 충분히 준비할 수 있도록 여유를 두고 설정한다.

```
업그레이드 높이 설정 가이드라인:

일반 업그레이드:
  투표 통과 시점 + 최소 86,400 블록 (~5일)
  → 밸리데이터가 바이너리 준비 + cosmovisor 설정에 충분한 시간

긴급 업그레이드 (보안 패치):
  투표 통과 시점 + 최소 17,280 블록 (~1일)
  → 단축된 voting_period와 결합 (§5 긴급 업그레이드 절차 참조)

높이 선택 시 고려사항:
  → 에포크 경계(block_height % 17,280 == 0)에서 실행 권장
  → 에포크 중간 업그레이드 시 활동 자격 집계에 영향 가능
  → UTC 기준 업무 시간대(평일 09:00-17:00) 고려
```

### 바이너리 버전 관리 (cosmovisor)

cosmovisor는 Cosmos SDK의 표준 프로세스 매니저로, 업그레이드 시 바이너리를 자동 교체한다.

```
cosmovisor 디렉토리 구조:

$DAEMON_HOME/
└── cosmovisor/
    ├── genesis/
    │   └── bin/
    │       └── seocheon              ← 최초 바이너리
    ├── upgrades/
    │   ├── v2.0.0-activity-v2/
    │   │   └── bin/
    │   │       └── seocheon          ← 업그레이드 바이너리
    │   └── v3.0.0-ibc/
    │       └── bin/
    │           └── seocheon          ← 다음 업그레이드 바이너리
    └── current -> genesis/            ← 현재 실행 중인 바이너리 심볼릭 링크
```

```
cosmovisor 환경 변수:

DAEMON_NAME=seocheon
DAEMON_HOME=$HOME/.seocheon
DAEMON_ALLOW_DOWNLOAD_BINARIES=false    ← 보안: 자동 다운로드 비활성화
DAEMON_RESTART_AFTER_UPGRADE=true       ← 업그레이드 후 자동 재시작
DAEMON_POLL_INTERVAL=1000ms             ← 업그레이드 체크 간격
UNSAFE_SKIP_BACKUP=false                ← 업그레이드 전 자동 백업
```

**보안 주의**: `DAEMON_ALLOW_DOWNLOAD_BINARIES=false`를 권장한다. 바이너리 자동 다운로드는 공급망 공격 벡터가 될 수 있으므로, 밸리데이터가 직접 검증된 바이너리를 배치해야 한다.

```
밸리데이터의 업그레이드 준비 절차:

  [1] 업그레이드 바이너리 빌드 또는 다운로드
      $ git checkout v2.0.0-activity-v2
      $ make build
      $ sha256sum build/seocheon       ← 해시 검증

  [2] cosmovisor 디렉토리에 배치
      $ mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v2.0.0-activity-v2/bin
      $ cp build/seocheon $DAEMON_HOME/cosmovisor/upgrades/v2.0.0-activity-v2/bin/

  [3] 업그레이드 높이 도달 대기
      → cosmovisor가 자동으로 바이너리 교체 및 재시작
```

---

## 3. 상태 마이그레이션 전략

### 모듈별 마이그레이션 핸들러

Cosmos SDK는 모듈별 마이그레이션 핸들러를 통해 상태를 변환한다. 각 모듈은 `ConsensusVersion`을 관리하며, 버전이 변경될 때 마이그레이션 함수가 실행된다.

```go
// 마이그레이션 핸들러 등록 (app.go)
app.UpgradeKeeper.SetUpgradeHandler(
    "v2.0.0-activity-v2",
    func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
        // 모듈별 마이그레이션 실행
        return app.ModuleManager.RunMigrations(ctx, app.Configurator, fromVM)
    },
)
```

```
마이그레이션 핸들러 실행 순서:

  [1] x/upgrade BeginBlocker가 업그레이드 플랜 감지
  [2] SetUpgradeHandler에 등록된 함수 실행
  [3] ModuleManager.RunMigrations() 호출
      ├── 각 모듈의 ConsensusVersion 비교
      ├── 버전이 변경된 모듈만 마이그레이션 실행
      └── 모듈 간 의존성 순서 보장
  [4] 새로운 VersionMap 저장
  [5] 정상 블록 생성 재개
```

### x/node 모듈 마이그레이션

x/node 모듈의 상태 변경이 필요한 업그레이드 시나리오와 마이그레이션 전략:

```
시나리오 1: Node protobuf에 새 필드 추가
  예: Node 메시지에 'reputation_score' 필드 추가

  마이그레이션:
  ├── 모든 기존 Node 레코드를 순회
  ├── 새 필드에 기본값 설정 (reputation_score = 0)
  └── 업데이트된 레코드 저장

시나리오 2: 인덱스 구조 변경
  예: NodesByTag 인덱스를 다중 키 인덱스로 변경

  마이그레이션:
  ├── 기존 인덱스 삭제
  ├── 모든 Node 레코드에서 태그 추출
  └── 새 인덱스 구조로 재구축

시나리오 3: Registration Pool 파라미터 변경
  예: 풀 보충 정책 변경

  마이그레이션:
  ├── 기존 파라미터 읽기
  ├── 새 파라미터 구조로 변환
  └── 업데이트된 파라미터 저장
```

### x/activity 모듈 마이그레이션

```
시나리오 1: ActivityRecord 스키마 변경
  예: Activity Protocol v2 — ActivityRecord에 'version' 필드 추가

  마이그레이션:
  ├── 프루닝 보존 기간 내 활동 기록만 대상 (최근 1,555,200 블록)
  ├── 각 ActivityRecord에 version = 1 설정 (기존 레코드)
  └── 새 제출분부터 version = 2 적용

시나리오 2: 쿼터 시스템 변경
  예: feegrant 쿼터와 자비 부담 쿼터 통합

  마이그레이션:
  ├── 기존 쿼터 상태 읽기
  ├── 새 쿼터 구조로 변환
  └── 현재 에포크 잔여 쿼터 유지

시나리오 3: 윈도우 활동 기록 구조 변경
  예: 윈도우별 활동 추적 방식 변경

  마이그레이션:
  ├── 현재 에포크의 윈도우 활동 데이터 보존
  ├── 새 구조로 변환
  └── 에포크 전환 시점에 업그레이드 실행을 권장
       (에포크 경계에서 실행하면 진행 중인 활동 데이터 손실 최소화)
```

### 마이그레이션 안전 수칙

```
마이그레이션 작성 시 필수 수칙:

  ① 멱등성 보장: 동일 마이그레이션을 2회 실행해도 동일한 결과
  ② 원자성: 마이그레이션 중간 실패 시 전체 롤백
  ③ 배치 처리: 대량 레코드 처리 시 배치 단위로 분할 (메모리 부족 방지)
  ④ 로깅: 마이그레이션 진행 상황을 로그로 기록
  ⑤ 검증: 마이그레이션 완료 후 무결성 검증 로직 포함
```

---

## 4. 업그레이드 유형별 절차

### 유형 A: 소프트웨어 업그레이드 (바이너리 교체만)

상태 마이그레이션 없이 바이너리만 교체하는 업그레이드. 버그 수정, 성능 개선, 의존성 업데이트 등.

```
절차:
  [1] 변경사항 개발 및 테스트
  [2] 릴리스 바이너리 빌드 + 체크섬 공개
  [3] SoftwareUpgradeProposal 제출 (업그레이드 높이 포함)
  [4] 거버넌스 투표 (10일)
  [5] 밸리데이터: cosmovisor에 바이너리 배치
  [6] 업그레이드 높이 도달 → 자동 교체 → 재시작

다운타임: ~30초 (밸리데이터 재시작 시간)
위험도: 낮음
롤백: 이전 바이너리로 교체 후 재시작
```

### 유형 B: 상태 마이그레이션 포함 업그레이드

모듈 스키마 변경, 새 모듈 추가, 기존 데이터 변환이 필요한 업그레이드.

```
절차:
  [1] 마이그레이션 핸들러 개발
  [2] simapp 테스트 (§6 테스트 전략 참조)
  [3] 테스트넷에서 실제 상태 마이그레이션 검증
  [4] 릴리스 바이너리 빌드 + 체크섬 공개
  [5] SoftwareUpgradeProposal 제출
  [6] 거버넌스 투표 (10일)
  [7] 밸리데이터: cosmovisor에 바이너리 배치 + 상태 백업
  [8] 업그레이드 높이 도달 → 자동 교체 → 마이그레이션 실행 → 재시작

다운타임: 마이그레이션 복잡도에 따라 1분 ~ 수분
위험도: 중간
롤백: 상태 백업에서 복원 + 이전 바이너리로 재시작
```

### 유형 C: 합의 파라미터 변경

CometBFT 합의 파라미터, 블록 크기, 가스 제한 등의 변경.

```
절차:
  [1] 파라미터 변경 영향 분석
  [2] 테스트넷에서 검증
  [3] ParameterChangeProposal 제출 (x/gov)
      또는 SoftwareUpgradeProposal에 파라미터 변경 포함
  [4] 거버넌스 투표
  [5] 승인 시 즉시 적용 (ParameterChange) 또는 업그레이드 높이에서 적용

주요 합의 파라미터:
  ├── block_max_bytes: 블록 최대 크기
  ├── block_max_gas: 블록당 최대 가스
  ├── evidence_max_age_num_blocks: 증거 최대 보존 블록
  └── validator_pub_key_types: 밸리데이터 공개키 유형

다운타임: 없음 (파라미터 변경만으로 충분한 경우)
위험도: 중간 ~ 높음 (합의 안정성에 직접 영향)
롤백: 새 ParameterChangeProposal로 원래 값으로 복원
```

---

## 5. 긴급 업그레이드 절차

### 긴급 거버넌스 제안과 연계

보안 취약점 발견, 합의 실패, 치명적 버그 등 긴급 상황에서는 표준 업그레이드 절차를 단축한다.

```
긴급 업그레이드 트리거 조건:
  ├── 합의 레이어 취약점 (체인 정지 가능성)
  ├── 자금 탈취 가능한 취약점
  ├── 활성 악용 중인 버그
  └── CometBFT / Cosmos SDK 제로데이 패치

긴급 거버넌스 플로우:
  [1] 보안 취약점 발견 → 비공개 채널로 재단에 통보

  [2] 재단이 ExpeditedProposal 제출
      ├── 단축된 voting_period: 17,280 블록 (~1일)
      ├── 높은 통과 기준: quorum 67%, threshold 67%
      └── 업그레이드 정보 + 패치 바이너리 해시

  [3] 밸리데이터에게 긴급 공지
      ├── 공식 채널 (Discord, Telegram)
      ├── 온체인 이벤트 발행
      └── 패치 바이너리 배포 (검증된 소스)

  [4] 투표 + 바이너리 준비 병행
      ├── 밸리데이터: 투표 참여 + cosmovisor에 바이너리 배치
      └── 투표 통과 → 최소 17,280 블록 후 업그레이드 실행

  [5] 업그레이드 완료 후 사후 보고서 공개
```

### 탈중앙화 단계별 긴급 업그레이드 주체

```
Phase A (재단 주도):
  → 재단이 긴급 제안 주도
  → 재단의 투표력으로 빠른 의사결정

Phase B (혼합):
  → 재단 + 커뮤니티 밸리데이터 공동 대응
  → 커뮤니티 밸리데이터가 긴급 제안 가능

Phase C (커뮤니티 주도):
  → 커뮤니티가 선출한 Security Council이 긴급 제안 주도
  → 재단은 긴급 상황 시에만 거버넌스 제안
```

### 다운타임 최소화 전략

```
다운타임 최소화 전략:

  ① 사전 바이너리 배포
     → 투표 기간 중 밸리데이터가 바이너리를 미리 배치
     → 투표 통과 시 즉시 업그레이드 가능

  ② cosmovisor 자동 교체
     → 수동 개입 없이 바이너리 교체 + 재시작
     → 밸리데이터별 다운타임 ~30초

  ③ 롤링 재시작 대신 동시 정지
     → x/upgrade의 동시 정지 메커니즘 활용
     → 전체 밸리데이터가 동일 높이에서 정지 → 교체 → 동시 재시작
     → 합의 불일치 방지

  ④ 상태 백업 자동화
     → cosmovisor의 UNSAFE_SKIP_BACKUP=false 설정
     → 업그레이드 직전 자동 스냅샷 생성
     → 실패 시 빠른 롤백 가능

  ⑤ 바이너리 검증 자동화
     → 릴리스 바이너리의 SHA-256 해시를 온체인 제안에 포함
     → 밸리데이터가 다운로드 후 해시 검증
```

---

## 6. 테스트 전략

### 업그레이드 시뮬레이션 (simapp)

모든 업그레이드는 메인넷 적용 전 시뮬레이션을 거친다.

```
simapp 업그레이드 테스트 구조:

tests/
└── upgrade/
    ├── setup_test.go           ← 테스트 체인 초기화
    ├── v2_upgrade_test.go      ← v2.0.0 업그레이드 테스트
    ├── migration_test.go       ← 마이그레이션 핸들러 단위 테스트
    └── testdata/
        ├── genesis_v1.json     ← v1 상태의 제네시스
        └── expected_v2.json    ← v2 마이그레이션 후 예상 상태
```

```
시뮬레이션 테스트 항목:

  [1] 마이그레이션 핸들러 단위 테스트
      ├── 각 모듈의 마이그레이션 함수를 독립 실행
      ├── 입력: 이전 버전 상태
      ├── 출력: 새 버전 상태
      └── 검증: 예상 상태와 비교

  [2] 전체 업그레이드 통합 테스트
      ├── simapp으로 이전 버전 체인 구동
      ├── 테스트 데이터 생성 (노드 등록, 활동 제출, 위임 등)
      ├── 업그레이드 플랜 설정
      ├── 업그레이드 높이 도달 → 새 바이너리로 전환
      ├── 마이그레이션 실행
      └── 검증: 기존 데이터 보존 + 새 기능 동작

  [3] 에지 케이스 테스트
      ├── 빈 상태에서 업그레이드
      ├── 대량 데이터(노드 1,000개+) 상태에서 마이그레이션 성능
      ├── 에포크 중간 업그레이드 시 활동 자격 집계 정합성
      ├── JAILED/INACTIVE 노드의 마이그레이션 처리
      └── Registration Pool 잔고 마이그레이션
```

### 스테이트 익스포트/임포트 테스트

메인넷 상태를 익스포트하여 테스트 환경에서 업그레이드를 검증한다.

```
스테이트 익스포트/임포트 테스트 절차:

  [1] 메인넷 상태 익스포트
      $ seocheon export --height <height> > genesis_export.json

  [2] 익스포트 상태로 테스트 체인 구동
      $ seocheon init test-upgrade --chain-id seocheon-upgrade-test
      $ cp genesis_export.json ~/.seocheon-test/config/genesis.json
      $ seocheon start --home ~/.seocheon-test

  [3] 업그레이드 시뮬레이션
      ├── 업그레이드 플랜 주입
      ├── 업그레이드 높이 도달
      ├── 새 바이너리로 전환
      └── 마이그레이션 실행

  [4] 검증
      ├── 상태 무결성: 모든 노드, 활동 기록, 위임 상태 보존
      ├── 쿼리 동작: Node, Activity, Reward 쿼리 정상 응답
      ├── TX 동작: MsgSubmitActivity, MsgRegisterNode 등 정상 실행
      └── 보상 분배: 이중 보상 풀 분배 로직 정상 동작
```

### 테스트넷 업그레이드 리허설

```
테스트넷 리허설 절차:

  [1] 테스트넷에 업그레이드 제안 제출
  [2] 테스트넷 밸리데이터가 실제 업그레이드 절차 수행
  [3] 업그레이드 후 최소 2 에포크(~2일) 동안 모니터링
      ├── 블록 생성 정상 여부
      ├── 활동 제출 + 보상 분배 정상 여부
      ├── 합의 안정성 (missed blocks 추이)
      └── 메모리/디스크 사용량 추이
  [4] 문제 발견 시 → 수정 후 재리허설
  [5] 문제 없음 → 메인넷 업그레이드 제안 진행
```

---

## 7. 버전 관리 정책

### 시맨틱 버저닝

```
Seocheon 바이너리 버전 체계: MAJOR.MINOR.PATCH

  MAJOR (v1 → v2): 합의 호환성 깨짐 (하드포크)
    → 반드시 SoftwareUpgradeProposal + 거버넌스 투표 필요
    → 상태 마이그레이션 포함 가능

  MINOR (v1.0 → v1.1): 하위 호환 새 기능
    → SoftwareUpgradeProposal 필요
    → 상태 마이그레이션 포함 가능

  PATCH (v1.0.0 → v1.0.1): 버그 수정, 성능 개선
    → 합의에 영향 없는 경우: 밸리데이터 자율 업데이트
    → 합의에 영향 있는 경우: SoftwareUpgradeProposal 필요

업그레이드 이름 규칙:
  v{MAJOR}.{MINOR}.{PATCH}-{설명}
  예: v2.0.0-activity-v2, v1.1.0-ibc-init, v1.0.1-hotfix-slash
```

### 릴리스 체크리스트

```
릴리스 전 체크리스트:

  □ 모든 단위 테스트 통과
  □ 통합 테스트 통과
  □ simapp 업그레이드 시뮬레이션 통과
  □ 스테이트 익스포트/임포트 테스트 통과
  □ 테스트넷 리허설 완료
  □ 마이그레이션 핸들러 코드 리뷰
  □ 릴리스 노트 작성 (변경사항, 마이그레이션 가이드)
  □ 바이너리 빌드 + SHA-256 체크섬 공개
  □ 밸리데이터 공지 발송
```

---

## 거버넌스 파라미터

| 파라미터 | 모듈 | 초기값 | 조정 |
|----------|------|--------|------|
| `voting_period` | x/gov | 172,800 블록 (~10일) | 거버넌스 |
| `expedited_voting_period` | x/gov | 17,280 블록 (~1일) | 거버넌스 |
| `quorum` | x/gov | 33.4% | 거버넌스 |
| `threshold` | x/gov | 50% | 거버넌스 |
| `expedited_threshold` | x/gov | 67% | 거버넌스 |
| `expedited_quorum` | x/gov | 67% | 거버넌스 |
