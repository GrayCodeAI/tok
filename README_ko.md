# tok

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **적게 쓰고, 많이 얻으세요.** 전송 전에 프롬프트를 압축하세요. 가독성을 위해 노이즈가 많은 출력을 필터링하세요.

---

## tok 이란?

tok 은 AI 코딩 어시스턴트를 위한 통합 토큰 최적화 CLI 도구입니다:

1. **프롬프트 압축** — AI 로 전송되는 입력 토큰 수를 줄입니다
2. **터미널 출력 필터링** — 중요한 정보만 표시합니다
3. **AI 에이전트 톤 설정** — 에이전트가 간결하게 응답하도록 하여 출력 토큰을 절약합니다

## 설치

```bash
go install github.com/GrayCodeAI/tok/cmd/tok@latest
```

또는 Homebrew 로:

```bash
brew install GrayCodeAI/tap/tok
```

## 빠른 시작

```bash
# 프롬프트 압축
$ tok compress -mode ultra -input "JWT 토큰을 사용한 사용자 인증 시스템을 구현해주세요"
JWT로 사용자 인증 구현.

# 명령어 출력 필터링
$ tok npm test
# 200 줄 테스트 결과 → 3 줄: 합격/불합격 + 실패 항목

# 에이전트 톤 설정
$ tok on ultra       # 최대 간결성
$ tok status         # 현재 모드 확인
```

## 압축 모드

| 모드 | 스타일 | 입력 절약 |
|------|-------|---------|
| `lite` | 불필요한 단어 제거, 문법 유지 | ~20% |
| `full` | 관사 제거, 단편 허용 | ~40% _(기본값)_ |
| `ultra` | 전보체, 약어 | ~60% |

## 문서

전체 문서는 [영문 README](README.md)를 참조하세요.

## 라이선스

[MIT](LICENSE)
