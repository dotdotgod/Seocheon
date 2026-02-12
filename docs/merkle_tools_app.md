# Seocheon Merkle Tools — 프론트엔드 앱

> Seocheon Activity Protocol의 해시 생성, 머클트리 구성, 머클 증명 검증을 위한 클라이언트 도구. 백엔드 없이 브라우저에서 동작한다.

---

## 개요

3개의 페이지로 구성된 SPA(Single Page Application)이다. 모든 연산은 브라우저에서 수행되며, 서버로 데이터를 전송하지 않는다.

```
Merkle Tools App
├── /hash      — SHA-256 해시 생성기
├── /tree      — 머클트리 & 머클루트 생성기
└── /verify    — 머클 증명 검증기
```

---

## 페이지 1: Hash Generator (`/hash`)

임의의 입력에서 SHA-256 해시를 생성한다.

### 기능

- 텍스트 입력 → SHA-256 해시 출력
- 파일 입력 → SHA-256 해시 출력
- 복사 버튼 (해시 클립보드 복사)

### 입출력

```
입력:
  - 텍스트 (textarea)
  - 또는 파일 (File input)

출력:
  - SHA-256 해시 (hex string, 64자)
  - 예: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
```

### 구현

```
브라우저 Web Crypto API 사용:
  const hash = await crypto.subtle.digest('SHA-256', data)
  → ArrayBuffer → hex string 변환

외부 라이브러리 불필요
```

### Seocheon 연동

이 페이지에서 생성한 해시는 MsgSubmitActivity의 `activity_hash`로 직접 사용할 수 있다. 노드 운영자가 오프체인 Activity Report의 해시를 계산하는 데 사용한다.

---

## 페이지 2: Merkle Tree Builder (`/tree`)

해시 목록을 입력받아 머클트리와 머클루트를 생성한다.

### 기능

- 해시 목록 입력 (줄바꿈 구분)
- 머클트리 시각화 (트리 구조)
- 머클루트 출력
- 각 리프별 머클 증명(Proof) 생성 및 다운로드
- 전체 트리 데이터 JSON 다운로드

### 입출력

```
입력:
  해시 목록 (줄바꿈 구분, 각 줄은 64자 hex string)
  예:
    a1b2c3...  (hash_1)
    d4e5f6...  (hash_2)
    g7h8i9...  (hash_3)

출력:
  ① 머클루트 (hex string)
  ② 머클트리 시각화
  ③ 각 리프별 머클 증명 (JSON)
```

### 머클트리 구성 규칙

```
해시 함수: SHA-256
결합 방식: SHA-256(left_hash + right_hash)
  → 두 해시를 바이트 단위로 연결(concatenate)한 후 SHA-256

홀수 리프 처리: 마지막 리프를 복제하여 짝수로 만듦
  예: [A, B, C] → [A, B, C, C]

리프 정렬: 입력 순서 유지 (정렬하지 않음)
```

### 머클 증명 데이터 구조

각 리프에 대해 생성되는 증명:

```json
{
  "leaf": "a1b2c3...",
  "leaf_index": 0,
  "merkle_root": "f9e8d7...",
  "proof": [
    {
      "hash": "d4e5f6...",
      "position": "right"
    },
    {
      "hash": "ab12cd...",
      "position": "left"
    }
  ]
}
```

| 필드 | 설명 |
|------|------|
| `leaf` | 검증 대상 리프 해시 |
| `leaf_index` | 리프의 트리 내 인덱스 (0부터) |
| `merkle_root` | 트리의 머클루트 |
| `proof` | 루트까지 경로의 형제 해시 배열 |
| `proof[].hash` | 해당 레벨의 형제 노드 해시 |
| `proof[].position` | 형제가 왼쪽(`left`)인지 오른쪽(`right`)인지 |

### 전체 트리 JSON

```json
{
  "merkle_root": "f9e8d7...",
  "leaf_count": 4,
  "tree": {
    "levels": [
      ["a1b2c3...", "d4e5f6...", "g7h8i9...", "j0k1l2..."],
      ["ab12cd...", "ef34gh..."],
      ["f9e8d7..."]
    ]
  },
  "proofs": [
    { "leaf_index": 0, "proof": [...] },
    { "leaf_index": 1, "proof": [...] },
    { "leaf_index": 2, "proof": [...] },
    { "leaf_index": 3, "proof": [...] }
  ]
}
```

### Seocheon 연동

이 페이지의 출력은 `MsgSubmitActivity`와 연동된다:

```
머클루트        → MsgSubmitActivity.activity_hash (머클루트를 activity_hash로 제출)
전체 트리 JSON  → content_uri가 가리키는 오프체인 데이터
```

**참고**: 체인은 activity_hash의 구성 방법을 강제하지 않는다. 이 앱이 사용하는 머클트리 방식은 여러 가능한 해싱 전략 중 하나이다.

---

## 페이지 3: Merkle Verifier (`/verify`)

머클루트, 리프 해시, 머클 증명을 입력받아 해당 해시가 트리에 포함되어 있는지 검증한다.

### 기능

- 머클루트 입력
- 검증할 리프 해시 입력
- 머클 증명(Proof) 입력 (JSON)
- 검증 결과 표시 (성공/실패)
- 검증 과정 단계별 시각화

### 입출력

```
입력:
  ① 머클루트 (hex string)
  ② 리프 해시 (hex string)
  ③ 머클 증명 (JSON 또는 파일 업로드)

출력:
  ① 검증 결과: 성공 ✓ / 실패 ✗
  ② 검증 과정 (단계별):
     Step 1: leaf_hash = "a1b2c3..."
     Step 2: SHA-256(leaf_hash + proof[0].hash) = "ab12cd..."
     Step 3: SHA-256(proof[1].hash + step2_result) = "f9e8d7..."
     Step 4: "f9e8d7..." == merkle_root? → ✓ 검증 성공
```

### 검증 알고리즘

```
function verify(leaf, proof, merkle_root):
  current = leaf
  for each step in proof:
    if step.position == "left":
      current = SHA-256(step.hash + current)
    else:
      current = SHA-256(current + step.hash)
  return current == merkle_root
```

### Seocheon 연동

위임자 또는 인덱서가 `content_uri`에서 개별 해시와 증명을 다운로드한 후, 이 페이지에서 온체인 `activity_hash`와 대조하여 특정 활동이 포함되었는지 검증할 수 있다.

---

## 기술 스택

```
프레임워크: React + TypeScript
빌드: Vite
스타일: Tailwind CSS
해시: Web Crypto API (crypto.subtle.digest)
라우팅: React Router
상태 관리: React 내장 (useState/useReducer)
배포: 정적 호스팅 (GitHub Pages, Vercel 등)

외부 의존성: 최소화 (암호화 라이브러리 불필요)
```

### 왜 Web Crypto API인가?

| 방식 | 장단점 |
|------|--------|
| **Web Crypto API** | **브라우저 네이티브, 추가 번들 0KB, 하드웨어 가속 가능** |
| js-sha256 등 라이브러리 | 번들 크기 증가, 순수 JS 구현이라 느림 |
| Node.js crypto | 서버사이드 전용, 브라우저 불가 |

Web Crypto API는 모든 모던 브라우저에서 지원된다 (Chrome 37+, Firefox 34+, Safari 11+).

---

## 프로젝트 구조

```
apps/merkle-tools/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.ts
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── pages/
│   │   ├── HashGenerator.tsx        ← /hash
│   │   ├── MerkleTreeBuilder.tsx    ← /tree
│   │   └── MerkleVerifier.tsx       ← /verify
│   ├── lib/
│   │   ├── hash.ts                  ← SHA-256 래퍼 (Web Crypto API)
│   │   ├── merkle.ts                ← 머클트리 구성 + 증명 생성
│   │   └── verify.ts                ← 머클 증명 검증
│   ├── components/
│   │   ├── TreeVisualization.tsx     ← 머클트리 시각화 컴포넌트
│   │   ├── ProofSteps.tsx           ← 검증 단계 시각화
│   │   ├── HashOutput.tsx           ← 해시 출력 + 복사
│   │   └── Layout.tsx               ← 공통 레이아웃 + 네비게이션
│   └── types/
│       └── merkle.ts                ← 타입 정의
└── tests/
    ├── hash.test.ts
    ├── merkle.test.ts
    └── verify.test.ts
```

---

## 핵심 라이브러리 인터페이스

```typescript
// lib/hash.ts
async function sha256(data: Uint8Array): Promise<string>
async function sha256FromText(text: string): Promise<string>
async function sha256FromFile(file: File): Promise<string>

// lib/merkle.ts
interface MerkleTree {
  root: string
  leafCount: number
  levels: string[][]
}

interface MerkleProof {
  leaf: string
  leafIndex: number
  merkleRoot: string
  proof: Array<{ hash: string; position: 'left' | 'right' }>
}

async function buildMerkleTree(leaves: string[]): Promise<MerkleTree>
async function generateProof(tree: MerkleTree, leafIndex: number): MerkleProof

// lib/verify.ts
async function verifyProof(
  leaf: string,
  proof: MerkleProof['proof'],
  merkleRoot: string
): Promise<{ valid: boolean; steps: VerificationStep[] }>
```
