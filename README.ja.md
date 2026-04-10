# AI Agent Workflow

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **他の言語:** [English](README.md) | [中文简体](README.zh.md) | [中文繁體](README.zh-TW.md)

---

Go 言語製の軽量で宣言的な AI エージェントワークフローエンジン。ノードの依存関係を宣言するだけで、エンジンが実行順序を自動的に解決します。依存関係のないノードは goroutine によって並列実行されます。

> **宣言するだけで並列化** — ノードの依存関係を宣言するだけで、エンジンが自動的に実行レイヤーを構築し、並列実行します。

---

## このプロジェクトの特徴

多くの Go 製 DAG フレームワークは、手動でエッジを定義してグラフを構築する必要があります。本プロジェクトは異なるアプローチを採用しています：

- **`ValueFrom` で依存関係を宣言** — エンジンが自動的に依存グラフを発見
- **グラフの手動構築不要** — ノードをリストするだけで、エンジンが実行順序を計算
- **自動レイヤー並列化** — 同じ依存レベルのノードが並行実行
- **外部依存ゼロ** — Redis、データベース、メッセージキュー不要

## 比較

| 機能 | ai-agent | [CloudWeGo Eino](https://github.com/cloudwego/eino) | [Dagu](https://github.com/dagucloud/dagu) |
|------|----------|------|-------|
| 依存関係の宣言 | `ValueFrom` 自動発見 | 手動でエッジを定義 | YAML で定義 |
| 並列実行 | 自動レイヤー並列化 | DAG スケジューラ | YAML で定義 |
| AI/LLM 標準搭載 | LLMNode, ImageGenerationNode | 対応 | 無し |
| 外部依存 | 無し | 複数 | SQLite |
| API スタイル | Go コード（Builder パターン） | Go コード | YAML |
| 位置付け | 軽量な埋め込みライブラリ | エンタープライズ向けフレームワーク | ローカルワークフローエンジン |

---

## クイックスタート

```bash
go get github.com/chuccp/ai-agent
```

### Hello World

```go
package main

import (
    "fmt"
    ai_agent "github.com/chuccp/ai-agent"
    "github.com/chuccp/ai-agent/executor"
    "github.com/chuccp/ai-agent/node"
    "github.com/chuccp/ai-agent/value"
)

func main() {
    processNode := node.NewFunctionNodeBuilder("process").
        ExecFunc(func(state *node.State) (value.NodeValue, error) {
            name := state.GetRootValue().GetString("name")
            result := value.NewObjectValue()
            result.PutString("greeting", "Hello, "+name+"!")
            return result, nil
        }).Build()

    workflow := ai_agent.Of(processNode)
    ag := ai_agent.NewAgentBuilder("hello-agent").Workflow(workflow).Build()
    exec := ai_agent.NewAgentExecutor(ag, &executor.Config{MaxConcurrency: 2})

    input := value.NewObjectValue()
    input.PutString("name", "World")

    response, _ := exec.Exec(input)
    fmt.Println(response.NodeValue.(*value.ObjectValue).GetString("greeting"))
    // 出力: Hello, World!
}
```

---

## アーキテクチャ

```
Agent → Workflow → NodeExecutor → Nodes（自動レイヤー並列実行）
```

1. **Agent** — 設定付きで Workflow をラップ
2. **Workflow** — ノードのシーケンスを保持
3. **NodeExecutor** — `ValueFrom` の依存関係を分析して実行レイヤーを構築
4. **同一レイヤーのノードは goroutine で並列実行**

### 実行レイヤーの構築

```
NodeA ──┐
        ├──→ NodeC（A と B に依存、両方完了後に実行）
NodeB ──┘

レイヤー 1: [NodeA, NodeB]  ← 並列実行
レイヤー 2: [NodeC]          ← レイヤー 1 完了後に実行
```

---

## ノードの種類

| ノード | 説明 |
|--------|------|
| **FunctionNode** | `ExecFunc` でカスタムロジックを実装 |
| **IFNode** | Then/Else ワークフローによる条件分岐 |
| **IterationNode** | 配列入力に対する並列バッチ処理 |
| **OrderIterationNode** | 配列入力に対する逐次バッチ処理 |
| **LLMNode** | テンプレートベースの LLM 呼び出し（キャッシュ対応） |
| **ImageGenerationNode** | テンプレートプロンプト付き画像生成 |
| **InputNode** | エントリポイント、ルートパラメータの解析 |
| **OutputNode** | 出口ノード、出力変換対応 |

---

## コア機能

### 1. 宣言的な依存関係 → 自動並列化

ノードが必要なデータを宣言するだけで、エンジンが DAG を構築し、独立したノードを並列実行します。

```go
// nodeA と nodeB は依存関係なし → 並列実行
nodeA := node.NewFunctionNodeBuilder("nodeA").
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        res := value.NewObjectValue()
        res.PutString("data", "from A")
        return res, nil
    }).Build()

nodeB := node.NewFunctionNodeBuilder("nodeB").
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        res := value.NewObjectValue()
        res.PutString("data", "from B")
        return res, nil
    }).Build()

// nodeC は nodeA と nodeB の両方に依存 → 両方完了後に実行
nodeC := node.NewFunctionNodeBuilder("nodeC").
    ValuesFrom(
        value.NewValueFrom("nodeA", "", ""),
        value.NewValueFrom("nodeB", "", ""),
    ).
    ExecFunc(func(state *node.State) (value.NodeValue, error) {
        ctx := state.GetWorkflowContext()
        dataA := ctx.GetNodeValue("nodeA").(*value.ObjectValue).GetString("data")
        dataB := ctx.GetNodeValue("nodeB").(*value.ObjectValue).GetString("data")

        res := value.NewObjectValue()
        res.PutString("merged", dataA + " + " + dataB)
        return res, nil
    }).Build()

workflow := ai_agent.Of(nodeA, nodeB, nodeC)
```

### 2. 条件分岐（IFNode）

```go
ifNode, _ := node.NewIFNodeBuilder("check").
    Condition(func(ctx node.WorkflowContext) bool {
        return ctx.GetRootValue().GetInt("score") >= 60
    }).
    Then(ai_agent.Of( /* ... */ )).
    Else(ai_agent.Of( /* ... */ )).
    Build()
```

### 3. 並列反復処理（IterationNode）

```go
iterationNode := node.NewIterationNodeBuilder("iterate").
    IterationFrom(value.NewValueFrom("", "items", "")).
    Workflow(ai_agent.Of(
        node.NewFunctionNodeBuilder("processItem").
            ExecFunc(func(state *node.State) (value.NodeValue, error) {
                item := state.GetRootValue().GetInt("item")
                res := value.NewObjectValue()
                res.PutNumber("squared", float64(item*item))
                return res, nil
            }).Build(),
    )).Build()
```

### 4. LLMNode テンプレートエンジン

```go
llmNode := node.NewLLMNodeBuilder("llm").
    SystemTemplate("あなたは有能なアシスタントです。").
    UserTemplate("こんにちは、${name}さん！注文数は ${count} です。").
    LLMFunction(func(state *node.State, urls *value.UrlsValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error) {
        // systemPrompt: "あなたは有能なアシスタントです。"
        // userPrompt:   "こんにちは、Aliceさん！注文数は 100 です。"
        return value.NewTextValue("モックレスポンス"), nil
    }).
    Build()
```

`${variable}` と `{{.variable}}` の両方のテンプレート構文をサポートしています。

### 5. 非同期実行

```go
result := exec.ExecAsync(input)
if result.Error != nil {
    log.Fatal(result.Error)
}
fmt.Println(result.Response.Success)
```

---

## 値システム

パスベースの検索をサポートする豊富な多相値型：

```go
obj := value.NewObjectValue()
obj.PutString("name", "Alice").
    PutNumber("age", 30).
    PutBool("active", true)

// メソッドチェーン
obj.PutObject("address", value.NewObjectValue().
    PutString("city", "東京"),
)

// パスベースの検索
obj.FindValue("address.city")  // TextValue("東京")
obj.FindValue("$.name")        // TextValue("Alice")
```

型：`ObjectValue` | `ArrayValue` | `TextValue` | `BoolValue` | `NumberValue` | `NullValue` | `UrlsValue` | `FilesValue` | `StreamNodeValue`

---

## テスト

```bash
# 全テストを実行
go test ./...

# 特定のテストを実行
go test -v -run TestIFNode ./...

# カバレッジ
go test -cover ./...
```

---

## License

MIT
