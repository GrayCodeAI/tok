# tok

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **写更少，得更多。** 发送前压缩你的提示词。过滤嘈杂的输出以提高可读性。

---

## tok 是什么？

tok 是一个统一的令牌优化 CLI 工具，专为 AI 编码助手设计。它可以：

1. **压缩提示词** — 减少发送到 AI 的输入令牌数量
2. **过滤终端输出** — 只显示重要信息
3. **设置 AI 代理语气** — 让代理简洁回复，节省输出令牌

## 安装

```bash
go install github.com/GrayCodeAI/tok/cmd/tok@latest
```

或通过 Homebrew：

```bash
brew install GrayCodeAI/tap/tok
```

## 快速开始

```bash
# 压缩提示词
$ tok compress -mode ultra -input "请帮我实现一个带JWT令牌的用户认证系统"
实现用户认证，使用JWT。

# 过滤命令输出
$ tok npm test
# 200行测试结果 → 3行：通过/失败 + 失败项

# 设置代理语气
$ tok on ultra       # 最大简洁度
$ tok status         # 查看当前模式
```

## 压缩模式

| 模式 | 风格 | 输入节省 |
|------|------|---------|
| `lite` | 去除冗余，保留语法 | ~20% |
| `full` | 去除冠词，片段可接受 | ~40% _(默认)_ |
| `ultra` | 电报式，缩写 | ~60% |

## 更多文档

请参阅 [英文 README](README.md) 获取完整文档。

## 许可证

[MIT](LICENSE)
