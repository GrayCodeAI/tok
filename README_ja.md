# tok

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **少なく書いて、多くを得る。** 送信前にプロンプトを圧縮。可読性のためにノイズの多い出力をフィルタリング。

---

## tok とは？

tok は AI コーディングアシスタント向けの統合トークン最適化 CLI ツールです：

1. **プロンプト圧縮** — AI に送信される入力トークンを削減
2. **ターミナル出力フィルタリング** — 重要な情報のみを表示
3. **AI エージェントのトーン設定** — エージェントに簡潔に返信させ、出力トークンを節約

## インストール

```bash
go install github.com/GrayCodeAI/tok/cmd/tok@latest
```

または Homebrew で：

```bash
brew install GrayCodeAI/tap/tok
```

## クイックスタート

```bash
# プロンプトを圧縮
$ tok compress -mode ultra -input "JWTトークンを使ったユーザー認証システムを実装してください"
JWTでユーザー認証を実装。

# コマンド出力をフィルタリング
$ tok npm test
# 200行のテスト結果 → 3行：合否 + 失敗項目

# エージェントのトーンを設定
$ tok on ultra       # 最大限の簡潔さ
$ tok status         # 現在のモードを確認
```

## 圧縮モード

| モード | スタイル | 入力節約 |
|------|---------|---------|
| `lite` | フィラーを削除、構文を保持 | ~20% |
| `full` | 冠詞を削除、断片はOK | ~40% _(デフォルト)_ |
| `ultra` | 電信式、省略形 | ~60% |

## ドキュメント

完全なドキュメントは [英語の README](README.md) を参照してください。

## ライセンス

[MIT](LICENSE)
