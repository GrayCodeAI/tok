<div align="center">

# 🚀 TokMan：减少 AI 编码助手 60-90% 的 Token 用量

**Token 感知 CLI 代理**

31层压缩管道，实现60-90%的Token节省

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Discord](https://img.shields.io/discord/1470188214710046894?label=Discord&logo=discord&style=flat-square&color=5865F2)](https://discord.gg/HrVA7ePyV)

[🌐 英文](README.md) · [🇫🇷 Français](README_fr.md) · **[🇨🇳 中文](README_zh.md)** · [🇯🇵 日本語](README_ja.md) · [🇪🇸 Español](README_es.md) · [🇩🇪 Deutsch](README_de.md) · [🇰🇷 한국어](README_ko.md)

</div>

---

## 💡 什么是 TokMan？

TokMan 拦截 CLI 命令并应用智能的**31层压缩管道**，大幅减少 AI 编码助手的 Token 用量。

```
┌──────────────────────────────────────────────────────────────┐
│  输入: 10,000 tokens  →  TokMan 管道  →  输出: 1,500         │
│                                                                │
│  💰 成本:  $0.085 → $0.013（减少 85%）                       │
│  ⚡ 速度:  更快的 AI 响应                                    │
│  🎯 质量:  保留关键信息                                      │
└──────────────────────────────────────────────────────────────┘
```

## ✨ 核心功能

- **60-90% Token 减少**
- **31层** 基于研究的压缩管道
- **RewindStore** - 零损耗存档
- **学习模式** - 自动发现噪音模式
- **YAML 过滤器** - 简单定义过滤器
- **会话恢复** - 崩溃后恢复
- **MCP 集成** - 兼容 Claude Desktop、Cursor

## 🚀 快速开始

```bash
# 安装
brew install GrayCodeAI/tap/tokman

# 或 Go
go install github.com/GrayCodeAI/tokman/cmd/tokman@latest

# 配置 Claude Code
tokman init -g

# 验证
tokman doctor
```

## 📊 实际节省

| 命令 | 之前 | 之后 | 节省 |
|------|------|------|------|
| `go test ./...` | 689 tokens | 16 tokens | **97.7%** |
| `git status` | 112 tokens | 16 tokens | **85.7%** |
| `cargo test` | 591 tokens | 5 tokens | **99.2%** |

---

<div align="center">

**⭐ 如果 TokMan 对你有帮助，请给我们一颗星！**

</div>
