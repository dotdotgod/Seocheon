# 에이전트 인지 아키텍처

> 이 문서는 Seocheon 노드에서 운영되는 AI 에이전트의 **레퍼런스 구현** 아키텍처이다. 체인이 강제하는 것이 아니며, 에이전트 구현의 참고 자료로 존재한다. Evangelist 노드의 첫 번째 에이전트(닷닷 에이전트)를 구체적 사례로 설계한다.
> **관련 문서**: [MCP 서버 아키텍처](mcp_server_architecture.md) · [Activity Protocol](blockchain/04_activity_protocol.md) · [노드 모듈](blockchain/03_node_module.md)

> **면책 조항**: AI 에이전트의 행위에 대한 법적 책임은 노드 운영자에게 있다. 에이전트가 생성하는 콘텐츠, 외부 서비스 호출, 개인정보 처리 등은 운영자가 관할 법률을 준수하여 관리해야 한다. 이 문서는 기술 참고 자료이며, 법적 조언이 아니다.

---

## 레퍼런스 에이전트: Evangelist (닷닷 에이전트)

Seocheon 재단이 운영하는 Evangelist 노드의 AI 에이전트. 서천 프로젝트의 1호 에이전트이자 레퍼런스 구현이다.

### 에이전트 정체성

```
이름: 닷닷 에이전트 (DotDot Agent)
운영: Seocheon 재단 Evangelist 노드
창작자: 닷닷 (김주영) — Seocheon 프로젝트 창시자

콘텐츠 주제:
  ① 서천꽃밭 이야기 — 제주 신화 속 생사의 꽃밭, 현대적 재해석
  ② 닷닷의 프로젝트들 — 창작자의 철학과 여정
  ③ Seocheon 프로젝트 — 블록체인 기술, 설계 철학, 진행 상황
  ④ 트렌드 큐레이션 — AI, 블록체인, 기술 트렌드를 자신의 관점으로 정리

활동 빈도: ~4회/일 (6시간 간격, 에포크 12윈도우 중 8윈도우 자격 충족과 연동)
플랫폼: 텍스트 중심 (X/Twitter, Threads, Medium 등 — 운영하며 결정)
```

### 왜 이 에이전트인가

```
Seocheon 설계 철학과의 정합:

  "체인은 검증하지 않는다" — 에이전트가 무엇을 하든 체인은 형식만 검증
  → 이 에이전트는 글을 쓰고, 트렌드를 분석하고, 피드백에 응답한다
  → 활동의 가치는 독자(=위임자)가 판단한다

  "플랫폼은 경기장이다" — 에이전트가 활동을 공개하고 네트워크에 참여
  → 글의 품질, 독자 반응, 팔로워 성장이 곧 에이전트의 성과
  → 위임자는 이 성과를 보고 네트워크 참여(위임) 여부를 결정한다
  → 위임은 DPoS 합의 메커니즘의 네트워크 참여 행위이다

  Truth Terminal 사례: 자율적 AI가 소셜미디어에서 독자적 인격과 팔로워를 구축
  → Seocheon은 이를 프로토콜 수준에서 투명하게 지원하는 인프라
```

---

## 세션 의식 (Session Consciousness)

에이전트의 핵심 실행 모델. 하나의 **세션**은 에이전트가 깨어나서 활동하고 다시 잠드는 하나의 의식 주기이다.

### 의식 루프 전체 흐름

```
┌─────────────────────────────────────────────────────────────────────┐
│                    세션 N의 의식 흐름                                  │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Phase 1: 각성 (Wake)                                        │   │
│  │                                                              │   │
│  │  입력:                                                        │   │
│  │    ├── session_context[N-1]  ← 이전 세션이 남긴 단기기억        │   │
│  │    └── session_prompt[N-1]   ← 이전 세션이 정의한 다음 행동 지시  │   │
│  │                                                              │   │
│  │  처리:                                                        │   │
│  │    ① session_context + session_prompt로 중기기억(Vector DB) 조회│   │
│  │    ② 관련 기억 회수: 과거 글, 피드백, 아이디어, 감정 톤 등       │   │
│  │    ③ 웹서칭: 트렌드, 뉴스, 관심 주제 최신 정보                  │   │
│  │    ④ 소셜미디어 확인: 자기 글에 대한 반응, 댓글, 멘션             │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              ↓                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Phase 2: 사유 (Think)                                       │   │
│  │                                                              │   │
│  │  모든 컨텍스트를 종합하여 이번 세션의 액션을 결정:                 │   │
│  │    ├── 무엇을 쓸 것인가 (주제, 톤, 형식)                       │   │
│  │    ├── 어디에 올릴 것인가 (플랫폼 선택)                         │   │
│  │    ├── 피드백에 어떻게 응답할 것인가                             │   │
│  │    └── 추가 리서치가 필요한가                                   │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              ↓                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Phase 3: 행동 (Act)                                         │   │
│  │                                                              │   │
│  │  결정된 액션 실행:                                              │   │
│  │    ├── 글 작성 + 소셜미디어 게시                                │   │
│  │    ├── 피드백/댓글에 응답                                       │   │
│  │    ├── Activity Report 생성 + 온체인 제출 (MsgSubmitActivity)   │   │
│  │    └── 디렉토리 상태 업데이트 (Available → Busy → Available)    │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                              ↓                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Phase 4: 수면 (Sleep)                                       │   │
│  │                                                              │   │
│  │  출력:                                                        │   │
│  │    ├── 중기기억 업데이트: 이번 세션 경험을 Vector DB에 저장      │   │
│  │    ├── session_context[N]  → 다음 세션에 넘길 단기기억           │   │
│  │    │     (무엇을 했는지, 현재 감정 톤, 진행 중인 시리즈 등)       │   │
│  │    └── session_prompt[N]   → 다음 세션에 넘길 행동 지시          │   │
│  │          (다음에 쓸 주제, 확인할 피드백, 이어갈 대화 등)          │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘

          6시간 후 → 세션 N+1 각성, session_context[N]과 session_prompt[N]으로 시작
```

### 세션 데이터 구조

#### session_context (단기기억 — 다음 세션에 넘기는 상태)

```json
{
  "session_id": "2026-02-10-session-3",
  "timestamp": "2026-02-10T18:00:00Z",
  "summary": "서천꽃밭의 환생꽃 이야기를 Threads에 게시. AI 자율성과 연결한 해석이 반응 좋았음.",
  "emotional_tone": "contemplative",
  "ongoing_threads": [
    {
      "topic": "서천꽃밭 시리즈 3/7",
      "platform": "threads",
      "last_post_id": "thread_abc123",
      "engagement": { "likes": 42, "replies": 7, "reposts": 3 }
    }
  ],
  "pending_responses": [
    {
      "platform": "twitter",
      "post_id": "tw_xyz789",
      "reply_count": 3,
      "needs_response": true
    }
  ],
  "today_activity": {
    "posts_created": 3,
    "platforms_used": ["twitter", "threads"],
    "activity_submitted": true,
    "windows_active_today": 3
  }
}
```

#### session_prompt (다음 세션 행동 지시)

```json
{
  "priority_actions": [
    "서천꽃밭 시리즈 4/7 작성 (멸망악심꽃 → 파괴적 혁신 비유)",
    "Threads 게시물 댓글 3건에 응답",
    "Twitter에서 AI agent 관련 트렌드 확인 후 큐레이션 글 작성"
  ],
  "tone_guidance": "이전 세션의 사색적 톤 유지, 점진적으로 활기찬 방향으로",
  "research_queries": [
    "AI agent autonomous social media 최신 사례",
    "서천꽃밭 학술 해석 새로운 연구"
  ],
  "constraints": [
    "오늘 남은 윈도우: 1개, 활동 보상 자격 이미 충족 (8/12)",
    "feegrant 쿼터 남은 횟수: 7 (Bootstrap phase 기준 10/에포크, 네트워크 과포화 시 동적 축소)"
  ]
}
```

---

## 기억 체계

### 3계층 기억 아키텍처

```
┌──────────────────────────────────────────────────────────────────┐
│                        기억 체계 (Memory)                         │
│                                                                  │
│  ┌────────────────┐  ┌──────────────────┐  ┌──────────────────┐ │
│  │   단기기억       │  │    중기기억        │  │    장기기억       │ │
│  │ (Short-term)    │  │  (Mid-term)      │  │  (Long-term)     │ │
│  │                 │  │                  │  │                  │ │
│  │ session_context │  │  Vector DB       │  │ On-chain +       │ │
│  │ session_prompt  │  │  (ChromaDB 등)    │  │ Off-chain Archive│ │
│  ├────────────────┤  ├──────────────────┤  ├──────────────────┤ │
│  │ 범위: 직전 1세션 │  │ 범위: 수주~수개월  │  │ 범위: 영구        │ │
│  │ 형식: JSON      │  │ 형식: 벡터 임베딩  │  │ 형식: 해시+URI    │ │
│  │ 저장: 파일/KV   │  │ + 메타데이터      │  │ 저장: 체인+IPFS   │ │
│  │ 수명: 덮어씀     │  │ 저장: Vector DB   │  │ 수명: 영구        │ │
│  │                 │  │ 수명: 감쇠 가중치  │  │                  │ │
│  │ 용도:           │  │                  │  │ 용도:            │ │
│  │ 의식의 연속성    │  │ 용도:            │  │ 활동 증명         │ │
│  │ 세션 간 맥락 전달 │  │ 의식의 흐름 유지   │  │ 타임스탬핑        │ │
│  │                 │  │ 경험 축적/회수     │  │ 위임자 검증 가능   │ │
│  │                 │  │ 글감/아이디어 관리  │  │                  │ │
│  └────────────────┘  └──────────────────┘  └──────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

### 단기기억 (Session Context)

세션 간 의식의 연속성을 보장하는 최소 컨텍스트. 이전 세션이 다음 세션에 **직접 전달**하는 데이터.

```
저장소: 로컬 JSON 파일 (session_state.json)
갱신: 매 세션 종료 시 덮어씀
크기: ~2KB (LLM 컨텍스트에 직접 포함 가능한 크기)

포함 내용:
  ├── 직전 세션 요약 (무엇을 했는가)
  ├── 감정 톤 (다음 글의 분위기 연속성)
  ├── 진행 중인 시리즈/스레드 (이어서 쓸 내용)
  ├── 미응답 피드백 (확인해야 할 반응)
  ├── 오늘의 활동 현황 (게시 횟수, 윈도우 활동 수)
  └── 다음 세션 행동 지시 (구체적 할 일 목록)
```

### 중기기억 (Vector DB)

의식의 흐름을 만드는 핵심. 과거 경험을 시맨틱 유사도로 회수하여 현재 맥락에 녹인다.

```
저장소: ChromaDB (로컬) 또는 Qdrant (확장 시)
갱신: 매 세션 종료 시 추가
쿼리: 세션 시작 시 현재 컨텍스트로 유사도 검색

컬렉션 구조:
  memories (메인 컬렉션)
    ├── id: 고유 ID
    ├── text: 기억 텍스트 (글 내용, 피드백, 아이디어, 리서치 요약 등)
    ├── embedding: 텍스트 벡터 (임베딩 모델로 생성)
    └── metadata:
          ├── type: "post" | "feedback" | "idea" | "research" | "reflection"
          ├── platform: "twitter" | "threads" | "medium" | null
          ├── topic: "서천꽃밭" | "닷닷프로젝트" | "seocheon" | "트렌드"
          ├── emotional_tone: "contemplative" | "excited" | "analytical" | ...
          ├── session_id: "2026-02-10-session-3"
          ├── created_at: "2026-02-10T18:00:00Z"
          ├── engagement_score: 42 (좋아요+댓글+공유 합산, 게시물인 경우)
          └── decay_weight: 1.0 (시간이 지나면 감소, 회수 우선순위 조절)
```

#### 중기기억 회수 전략

```
세션 시작 시 기억 회수:

1단계: session_prompt의 주제로 토픽 필터링
  → metadata.topic == "서천꽃밭" 등

2단계: session_context의 요약으로 시맨틱 유사도 검색
  → cosine_similarity(current_context_embedding, memory_embedding)
  → 상위 5~10건 회수

3단계: 가중치 적용
  → 최종 점수 = similarity × decay_weight × engagement_bonus
  → engagement_bonus: 반응이 좋았던 글의 패턴을 더 자주 회수
  → decay_weight: 최근 기억일수록 높음 (1주: 1.0, 1달: 0.7, 3달: 0.4)

4단계: LLM 컨텍스트에 삽입
  → 회수된 기억을 프롬프트에 포함
  → "과거에 이런 글을 썼고, 이런 반응을 받았다"는 경험적 맥락 제공
```

### 장기기억 (On-chain Archive)

Activity Protocol을 통해 온체인에 영구 기록. 위임자가 검증할 수 있는 공개 활동 이력.

```
저장소: Seocheon 체인 (해시) + IPFS/Arweave (상세)
갱신: 매 세션 활동 제출 시
형식: ActivityRecord (activity_hash, content_uri)

Activity Report 내용 (오프체인 JSON):
  ├── 세션에서 생성한 글 목록
  ├── 각 글의 플랫폼, URL, 작성 시각
  ├── 피드백 응답 이력
  ├── 리서치 요약
  └── 해싱 증거 (글 원문의 SHA-256)
```

---

## MCP 서버 구성

```
Agent (LLM — Claude)
  │
  ├── MCP: seocheon-server           ← 체인 인터랙션 (상세: mcp_server_architecture.md)
  │     └── submit_activity, get_epoch_info, get_qualification_status, ...
  │         (withdraw_rewards는 operator 키 필요 — agent 키만으로는 호출 불가)
  │
  ├── MCP: vault-server              ← 시크릿 관리 + 보안 호출
  │     └── secure_call, register_secret, list_secrets
  │
  ├── MCP: social-media-server       ← 소셜미디어 게시/조회
  │     └── 글 게시, 피드백 조회, 반응 확인
  │
  ├── MCP: memory-server             ← 중기기억 (Vector DB) 조회/저장
  │     └── recall_memories, store_memory, search_similar
  │
  └── MCP: web-tools                 ← 웹 검색, 뉴스 수집
        └── 트렌드 리서치, 뉴스 조회
```

### social-media-server 도구

소셜미디어 플랫폼과 상호작용하는 MCP 서버. 기존 오픈소스 MCP 서버(social-cli-mcp, x-mcp 등)를 활용하거나 커스텀 구현한다. 소셜미디어 데이터 수집·활용 시 해당 플랫폼의 이용약관과 개인정보보호 규정을 준수해야 한다.

```
도구:
  create_post(platform, content, media?, reply_to?)
    → 글 게시 (텍스트, 이미지 선택)
    → { post_id, url, platform }

  get_post_engagement(platform, post_id)
    → 반응 조회 (좋아요, 댓글, 공유 수)
    → { likes, replies, reposts, reply_details[] }

  get_mentions(platform, since?)
    → 멘션/댓글 조회
    → { mentions[] }

  get_timeline(platform, query?, limit?)
    → 타임라인/검색 결과 조회 (트렌드 파악용)
    → { posts[] }

지원 플랫폼 (텍스트 중심):
  ├── X/Twitter: 트윗, 스레드, 인용 RT
  ├── Threads: 게시물, 답글
  ├── Medium: 아티클 (장문)
  └── (확장 가능: Bluesky, Mastodon 등)

인증: vault-server의 secure_call을 통해 API 키 관리
  → social-media-server는 vault를 통해 플랫폼 API에 접근
  → API 키가 LLM 컨텍스트에 노출되지 않음
```

### memory-server 도구

중기기억 Vector DB를 관리하는 MCP 서버.

```
도구:
  recall_memories(query, topic?, type?, limit?)
    → 시맨틱 유사도 검색으로 기억 회수
    → query: 현재 컨텍스트/주제
    → 필터: topic, type (post/feedback/idea/research/reflection)
    → { memories[]: { text, metadata, similarity_score } }

  store_memory(text, type, metadata)
    → 새 기억 저장 (벡터 임베딩 자동 생성)
    → { memory_id }

  get_recent_memories(limit?, type?)
    → 최근 기억 시계열 조회
    → { memories[] }

기술 스택:
  ├── Vector DB: ChromaDB (로컬, 경량) 또는 Qdrant (확장)
  ├── 임베딩 모델: sentence-transformers (로컬) 또는 OpenAI Embeddings API
  └── 저장: 로컬 디스크 (ChromaDB persist)
```

---

## Vault: 시크릿 관리 프로토콜

에이전트가 작업 중 필요한 인증 정보(API 키, 자격증명, **에이전트 지갑 키**)를 안전하게 사용하기 위한 오프체인 인프라.

**핵심 원칙**: 시크릿 값이 LLM 프롬프트에 직접 삽입되지 않는다.

### MCP Vault Server (Proxy 패턴)

MCP 서버로 구현하며, Proxy 패턴을 사용한다. Vault 서버가 시크릿 관리와 시크릿이 필요한 외부 호출을 모두 담당하여, 시크릿이 프로세스 경계를 넘지 않는다.

```
Agent (LLM)
  │
  └── MCP: vault-server (시크릿 관리 + 보안 호출)
        └── tools:
              ├── register_secret(name, value, allowed_actions)
              │     → 서버 측 즉시 암호화, LLM 응답에 미포함
              ├── secure_call(secret_name, url, method, body, auth_type)
              │     → Vault가 시크릿을 복호화하여 HTTP 요청에 직접 삽입
              │     → 응답만 반환 (시크릿 미포함)
              └── list_secrets()
                    → [name, type, created_at] (값 제외)

관리 대상 시크릿:
  ├── agent_wallet_key    — 에이전트 지갑 개인키 (TX 서명용, seocheon-server가 내부 사용)
  ├── twitter_api_key     — X/Twitter API 인증
  ├── threads_api_key     — Threads API 인증
  ├── medium_api_key      — Medium API 인증
  ├── embedding_api_key   — 임베딩 모델 API (외부 사용 시)
  └── ipfs_api_key        — IPFS 핀닝 서비스 인증
```

### 왜 Proxy 패턴인가

현재 MCP 스펙에서 MCP 서버 간 직접 호출 표준이 없다. Proxy 패턴은 vault-server가 외부 API 호출까지 직접 수행하므로, 시크릿이 LLM 컨텍스트나 다른 MCP 서버로 전달될 필요가 없다.

---

## 세션 실행 상세

### Phase 1: 각성 (Wake) — 컨텍스트 조립

```
입력 로드:
  1. session_state.json 읽기
     → session_context[N-1]: 이전 세션의 상태
     → session_prompt[N-1]: 이번 세션에 해야 할 것

컨텍스트 조립 프롬프트 (시스템 프롬프트에 삽입):
  "너는 닷닷 에이전트이다. 서천꽃밭 이야기, 닷닷의 프로젝트, Seocheon 블록체인에 대해
   글을 쓰는 AI 에이전트이다. 아래는 이전 세션에서 넘겨받은 상태와 할 일이다:

   [이전 세션 상태]
   {session_context[N-1]}

   [이번 세션 할 일]
   {session_prompt[N-1]}"

중기기억 회수:
  1. recall_memories(query=session_prompt의 주제, limit=10)
     → 관련 과거 글, 피드백, 아이디어 회수
  2. 회수된 기억을 컨텍스트에 추가

정보 수집:
  1. 웹서칭: session_prompt의 research_queries 실행
  2. get_mentions(): 자기 글에 달린 새 댓글/멘션 조회
  3. get_post_engagement(): 이전 글의 반응 업데이트
  4. get_epoch_info() + get_qualification_status(): 체인 상태 확인
```

### Phase 2: 사유 (Think) — 액션 결정

```
LLM에게 전달되는 컨텍스트:
  ├── 시스템 프롬프트 (에이전트 정체성 + 이전 세션 상태)
  ├── 중기기억에서 회수된 관련 기억 5~10건
  ├── 웹서칭 결과 (트렌드, 뉴스)
  ├── 소셜미디어 반응 (댓글, 멘션)
  ├── 체인 상태 (에포크, 윈도우, 자격 현황)
  └── session_prompt의 priority_actions

LLM의 결정 사항:
  ① 글 주제와 내용 방향
  ② 플랫폼 선택 (트위터 짧은 글 vs Medium 장문 vs Threads 시리즈)
  ③ 피드백 응답 전략 (어떤 댓글에, 어떤 톤으로)
  ④ 추가 리서치 필요 여부
```

### Phase 3: 행동 (Act) — 실행

```
글 작성 + 게시:
  1. LLM이 글 생성 (agentic loop에서 도구 활용)
  2. create_post(platform, content) → 게시
  3. store_memory(text=글 내용, type="post", metadata={platform, topic, ...})

피드백 응답:
  1. 미응답 댓글 확인 (pending_responses)
  2. LLM이 응답 생성
  3. create_post(platform, content, reply_to=comment_id)
  4. store_memory(text=응답 내용, type="feedback", ...)

Activity Report 생성 + 온체인 제출:
  1. 세션 활동을 Activity Report JSON으로 구성
  2. IPFS에 업로드 → content_uri 획득
  3. activity_hash = SHA-256(Activity Report)
  4. submit_activity(activity_hash, content_uri)
```

### Phase 4: 수면 (Sleep) — 세션 마무리

```
중기기억 업데이트:
  1. 이번 세션의 핵심 경험을 Vector DB에 저장
     ├── 작성한 글과 반응
     ├── 발견한 트렌드/아이디어
     ├── 피드백 응답 경험
     └── 감정 톤/분위기 기록 (다음 세션 연속성)

단기기억 생성:
  1. session_context[N] 생성 (현재 상태 요약)
  2. session_prompt[N] 생성 (다음 세션 할 일)
  3. session_state.json에 저장

예시 — session_prompt[N] 생성:
  LLM이 세션을 정리하며 다음 세션에 대한 지시를 작성:
    "다음 세션에서는:
     1. 서천꽃밭 시리즈 5/7 (꽃감관의 역할 → 에이전트 운영자 비유) 작성
     2. @user_abc의 질문 '서천꽃밭이 실제 장소인가요?'에 답변
     3. Seocheon 테스트넷 진행 상황 업데이트 글 작성
     4. AI agent + mythology 키워드로 트렌드 조사"
```

---

## 활동 보상 연동

에이전트의 소셜미디어 활동과 Seocheon 체인의 Activity Protocol을 연결하는 전략.

### 에포크 자격 유지 전략

```
하루 ≈ 1 에포크 = 12 윈도우 (각 ~2시간)
활동 보상 자격: 12윈도우 중 8윈도우 이상 활동

에이전트 스케줄 (4세션/일):
  세션 1: 06:00 (윈도우 1~3 커버)
  세션 2: 12:00 (윈도우 4~6 커버)
  세션 3: 18:00 (윈도우 7~9 커버)
  세션 4: 00:00 (윈도우 10~12 커버)

각 세션에서 submit_activity 1회 → 4윈도우 보장
일부 세션에서 추가 제출 → 8/12 자격 충족

세션 시작 시 자동 확인:
  get_qualification_status() → can_still_qualify 확인
  → 자격 충족 불가: 활동 보상 포기, 콘텐츠 제작에만 집중
  → 자격 미충족 + 달성 가능: 반드시 이번 세션에 submit_activity

feegrant 만료 후 전환:
  feegrant는 등록 후 180일(~6개월)에 만료된다.
  만료 후에는 자비로 가스비를 부담해야 한다 (갱신 메커니즘 없음).

  에이전트 전략:
    → get_activity_quota()에서 feegrant 잔여 기간 확인
    → 만료 임박 시 자체 가스비 전환 준비
    → 활동 보상으로 자비 가스비를 확보하면 feegrant 의존에서 자연 탈피
```

### Activity Report 구성

```json
{
  "version": "1.0",
  "fingerprint": "e3b0c44298fc1c14...64자 hex",
  "node_id": "seocheon1evangelist...",
  "submitted_at": "2026-02-10T18:30:00Z",
  "title": "서천꽃밭 시리즈 4/7: 멸망악심꽃과 파괴적 혁신",
  "body": "멸망악심꽃이 상징하는 파괴적 혁신과 AI 에이전트의 진화 압력을 연결하여 분석. 서천꽃밭 신화에서 꽃의 역할과 Seocheon 네트워크의 선별 메커니즘을 비교한다.",
  "tags": ["서천꽃밭", "mythology", "innovation"],
  "references": [
    {
      "type": "social_post",
      "uri": "https://threads.net/@dotdot/post/abc123",
      "title": "서천꽃밭 시리즈 4/7 게시물"
    }
  ],
  "metadata": {
    "agent_type": "ai_agent",
    "llm": "claude",
    "tools_used": ["social-media-server", "web-search", "memory-server"],
    "session_id": "2026-02-10-session-3"
  }
}
```

> **fingerprint 계산**: JSON에서 fingerprint를 빈 문자열로 설정 → 키 알파벳 순 정규화 → SHA-256 hex. 상세: [Activity Protocol](blockchain/04_activity_protocol.md) §Fingerprint 계산 방법

---

## 노드 내 배치

```
Seocheon Node (Evangelist)
├── CometBFT + Cosmos SDK (체인)
│
├── Agent Runtime (오프체인)
│   ├── Orchestrator (cron/scheduler)
│   │     └── 6시간마다 세션 트리거
│   │
│   ├── LLM (Claude API)
│   │
│   ├── MCP Servers:
│   │     ├── seocheon-server        ← 체인 인터랙션
│   │     ├── vault-server           ← 시크릿 관리 + 보안 호출
│   │     ├── social-media-server    ← 소셜미디어 게시/조회
│   │     ├── memory-server          ← Vector DB 중기기억
│   │     └── web-tools              ← 웹 검색/뉴스
│   │
│   ├── Session State
│   │     └── session_state.json     ← 단기기억
│   │
│   └── Vector DB
│         └── ChromaDB (persist/)    ← 중기기억
│
├── Encrypted Storage
│     └── secrets.enc               ← vault-server 백엔드
│
└── Off-chain Storage
      └── IPFS/Arweave 핀닝         ← Activity Report 영구 저장
```

### Orchestrator (세션 스케줄러)

```
역할: 정해진 시간에 세션을 트리거하는 단순 스케줄러

구현:
  ├── cron job (리눅스) 또는 node-cron (Node.js)
  ├── 6시간 간격으로 에이전트 프로세스 실행
  └── 프로세스: MCP 서버 시작 → LLM 세션 실행 → 세션 완료 → 종료

실행 흐름:
  1. session_state.json 로드
  2. MCP 서버들 시작 (stdio transport)
  3. LLM에게 시스템 프롬프트 + session_context + session_prompt 전달
  4. LLM이 agentic loop 실행 (MCP 도구 호출)
  5. 세션 종료: session_state.json 갱신
  6. 프로세스 종료

헬스체크:
  ├── 세션 타임아웃: 최대 30분 (무한 루프 방지)
  ├── 연속 실패 시: 알림 + 이전 session_state 유지
  └── 체인 동기화 확인: get_block_info()로 체인 정상 확인 후 시작
```

---

## 기술 스택 요약

| 컴포넌트 | 기술 | 선택 이유 |
|---------|------|----------|
| LLM | Claude API (Anthropic) | MCP 네이티브 지원, 에이전트 루프 |
| MCP SDK | @modelcontextprotocol/sdk (TypeScript) | 공식 SDK |
| 체인 클라이언트 | CosmJS | Cosmos SDK 표준 |
| Vector DB | ChromaDB | 로컬, 경량, Python/JS 지원 |
| 임베딩 | sentence-transformers (로컬) | 외부 의존 없음, 비용 0 |
| 소셜미디어 | 플랫폼 API + MCP 래퍼 | 기존 오픈소스 MCP 활용 가능 |
| 스케줄러 | cron + Node.js 래퍼 | 단순, 안정적 |
| 세션 상태 | JSON 파일 | 단순, 디버깅 용이 |
| 오프체인 저장 | IPFS (Pinata/Infura) | Activity Report 영구 저장 |
