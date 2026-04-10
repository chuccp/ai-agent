# AI Agent Workflow

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **其他語言：** [English](README.md) | [中文简体](README.zh.md) | [日本語](README.ja.md)

---

一個輕量級、聲明式的 Go 語言 AI Agent 工作流引擎。宣告節點依賴，引擎自動解析執行順序——無依賴的節點透過 goroutine 並行執行。

> **宣告即並行** — 宣告節點依賴，引擎自動構建執行層並並行執行。

---

## 專案特色

大多數 Go DAG 框架需要你手動定義邊和構建圖。本專案採用不同的思路：

- **透過 `ValueFrom` 宣告依賴** — 引擎自動發現依賴圖
- **無需手動構建圖** — 只需列出節點，引擎自動計算執行順序
- **自動分層並行** — 同一依賴層級的節點並行執行
- **零外部依賴** — 無需 Redis、資料庫或訊息佇列

## 對比

| 特性 | ai-agent | [CloudWeGo Eino](https://github.com/cloudwego/eino) | [Dagu](https://github.com/dagucloud/dagu) |
|------|----------|------|-------|
| 依賴宣告 | `ValueFrom` 自動發現 | 手動定義邊 | YAML 定義 |
| 並行執行 | 自動分層並行 | DAG 調度器 | YAML 定義 |
| AI/LLM 內建 | LLMNode, ImageGenerationNode | 支援 | 無 |
| 外部依賴 | 無 | 多個 | SQLite |
| API 風格 | Go 程式碼（Builder 模式） | Go 程式碼 | YAML |
| 定位 | 輕量級嵌入式函式庫 | 企業級框架 | 本機工作流引擎 |

---

## 快速開始

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
    // 輸出: Hello, World!
}
```

---

## 架構

```
Agent → Workflow → NodeExecutor → Nodes（自動分層並行執行）
```

1. **Agent** 用設定包裝 Workflow
2. **Workflow** 持有節點序列
3. **NodeExecutor** 透過分析 `ValueFrom` 依賴構建執行層
4. **同一層的節點透過 goroutine 並行執行**

### 執行層構建

```
NodeA ──┐
        ├──→ NodeC（依賴 A + B，等兩者完成後執行）
NodeB ──┘

第 1 層: [NodeA, NodeB]  ← 並行執行
第 2 層: [NodeC]          ← 第 1 層完成後執行
```

---

## 節點類型

| 節點 | 說明 |
|------|------|
| **FunctionNode** | 透過 `ExecFunc` 實現自定義邏輯 |
| **IFNode** | 條件分支，支援 Then/Else 工作流 |
| **IterationNode** | 對陣列輸入進行並行批次處理 |
| **OrderIterationNode** | 對陣列輸入進行順序批次處理 |
| **LLMNode** | 基於模板的 LLM 呼叫，支援快取 |
| **ImageGenerationNode** | 圖片生成，支援模板提示詞 |
| **InputNode** | 入口節點，解析根參數 |
| **OutputNode** | 出口節點，支援輸出轉換 |

---

## 核心功能

### 1. 宣告式依賴 → 自動並行

```go
// nodeA 和 nodeB 沒有依賴 → 並行執行
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

// nodeC 依賴 nodeA 和 nodeB → 等兩者完成後執行
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

### 2. 條件分支（IFNode）

```go
ifNode, _ := node.NewIFNodeBuilder("check").
    Condition(func(ctx node.WorkflowContext) bool {
        return ctx.GetRootValue().GetInt("score") >= 60
    }).
    Then(ai_agent.Of( /* ... */ )).
    Else(ai_agent.Of( /* ... */ )).
    Build()
```

### 3. 並行迭代（IterationNode）

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

### 4. LLMNode 模板引擎

```go
llmNode := node.NewLLMNodeBuilder("llm").
    SystemTemplate("你是一個有用的助手。").
    UserTemplate("你好，${name}！你的訂單數是 ${count}。").
    LLMFunction(func(state *node.State, urls *value.UrlsValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error) {
        return value.NewTextValue("模擬回覆"), nil
    }).
    Build()
```

支援 `${variable}` 和 `{{.variable}}` 兩種模板語法。

### 5. 非同步執行

```go
result := exec.ExecAsync(input)
if result.Error != nil {
    log.Fatal(result.Error)
}
fmt.Println(result.Response.Success)
```

---

## 值系統

豐富的多型值型別，支援路徑查詢：

```go
obj := value.NewObjectValue()
obj.PutString("name", "Alice").
    PutNumber("age", 30).
    PutBool("active", true)

// 鏈式呼叫
obj.PutObject("address", value.NewObjectValue().
    PutString("city", "台北"),
)

// 路徑查詢
obj.FindValue("address.city")  // TextValue("台北")
```

型別：`ObjectValue` | `ArrayValue` | `TextValue` | `BoolValue` | `NumberValue` | `NullValue` | `UrlsValue` | `FilesValue` | `StreamNodeValue`

---

## 測試

```bash
# 執行所有測試
go test ./...

# 執行指定測試
go test -v -run TestIFNode ./...

# 覆蓋率
go test -cover ./...
```

---

## License

MIT
