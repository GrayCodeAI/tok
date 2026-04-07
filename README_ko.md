<div align="center">

# 🚀 TokMan: AI 코딩 어시스턴트 Token 사용량 60-90% 절감

**Token 인식 CLI 프록시**

31층 압축 파이프라인으로 60-90% Token 절약 달성

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Discord](https://img.shields.io/discord/1470188214710046894?label=Discord&logo=discord&style=flat-square&color=5865F2)](https://discord.gg/HrVA7ePyV)

[🌐 English](README.md) · [🇫🇷 Français](README_fr.md) · [🇨🇳 中文](README_zh.md) · [🇯🇵 日本語](README_ja.md) · [🇪🇸 Español](README_es.md) · [🇩🇪 Deutsch](README_de.md) · **[🇰🇷 한국어](README_ko.md)**

</div>

---

## 💡 TokMan이란?

TokMan은 CLI 명령을 가로채 AI 코딩 어시스턴트의 Token 사용량을 획기적으로 줄이는 **31층 압축 파이프라인**을 적용합니다.

## ✨ 주요 기능

- **60-90% Token 절감**
- **31층** 연구 기반 압축
- **RewindStore** - 무손실 아카이브
- **학습 모드** - 자동 패턴 발견
- **YAML 필터** - 간편한 필터 정의
- **세션 복구** - 충돌 후 복구
- **MCP 통합** - Claude Desktop, Cursor 호환

## 🚀 빠른 시작

```bash
# 설치
brew install GrayCodeAI/tap/tokman

# 또는 Go로
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Claude Code 설정
tokman init -g

# 확인
tokman doctor
```

## 📊 실제 절약

| 명령 | 이전 | 이후 | 절약 |
|------|------|------|------|
| `go test ./...` | 689 tokens | 16 tokens | **97.7%** |
| `git status` | 112 tokens | 16 tokens | **85.7%** |
| `cargo test` | 591 tokens | 5 tokens | **99.2%** |

---

<div align="center">

**⭐ TokMan이 도움이 되었다면 Star를 눌러주세요!**

</div>
