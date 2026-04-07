<div align="center">

# 🚀 TokMan：AIコーディングアシスタントのToken使用量を60-90%削減

**Token対応CLIプロキシ**

31層の圧縮パイプラインで60-90%のToken節約を達成

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Discord](https://img.shields.io/discord/1470188214710046894?label=Discord&logo=discord&style=flat-square&color=5865F2)](https://discord.gg/HrVA7ePyV)

[🌐 English](README.md) · [🇫🇷 Français](README_fr.md) · [🇨🇳 中文](README_zh.md) · **[🇯🇵 日本語](README_ja.md)** · [🇪🇸 Español](README_es.md) · [🇩🇪 Deutsch](README_de.md) · [🇰🇷 한국어](README_ko.md)

</div>

---

## 💡 TokManとは？

TokManはCLIコマンドをインターセプトし、AIコーディングアシスタントのToken使用量を大幅に削減する**31層の圧縮パイプライン**を適用します。

## ✨ 主な機能

- **60-90% Token削減**
- **31層**の研究ベース圧縮パイプライン
- **RewindStore** - ゼロ損失アーカイブ
- **学習モード** - ノイズパターンの自動発見
- **YAMLフィルター** - 簡単なフィルター定義
- **セッションリカバリ** - クラッシュ後の復旧
- **MCP統合** - Claude Desktop、Cursor対応

## 🚀 クイックスタート

```bash
# インストール
brew install GrayCodeAI/tap/tokman
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# Claude Code用設定
tokman init -g

# 検証
tokman doctor
```

## 📊 実際の節約

| コマンド | 前 | 後 | 節約 |
|----------|-----|-----|------|
| `go test ./...` | 689 tokens | 16 tokens | **97.7%** |
| `git status` | 112 tokens | 16 tokens | **85.7%** |
| `cargo test` | 591 tokens | 5 tokens | **99.2%** |

---

<div align="center">

**⭐ TokManがお役に立ったらスターをお願いします！**

</div>
