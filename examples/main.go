package main

import (
	"fmt"
	"log"

	"github.com/chuccp/ai-agent"
	"github.com/chuccp/ai-agent/node"
	"github.com/chuccp/ai-agent/out"
	"github.com/chuccp/ai-agent/value"
)

func main() {
	// 创建函数节点 - 使用 Go text/template 处理数据
	processNode := node.NewFunctionNodeBuilder("process").
		ExecFunc(func(state *node.State) (value.NodeValue, error) {
			rootValue := state.GetRootValue()
			name := rootValue.GetString("name")

			result := value.NewObjectValue()
			result.PutString("greeting", "Hello, "+name+"!")
			result.PutString("processed", "true")
			result.PutNumber("count", 42)

			return result, nil
		}).
		Build()

	// 创建输出节点
	outputNode := node.NewOutputNodeBuilder("output").
		ValuesFrom(value.NewValueFrom("process", "", "")).
		OutFunc(func(nodeValue value.NodeValue) {
			fmt.Println("Output:", nodeValue.String())
		}).
		Build()

	// 创建工作流和代理
	workflow := ai_agent.Of(processNode, outputNode)
	ag := ai_agent.NewAgentBuilder("hello-agent").
		Workflow(workflow).
		Build()

	// 创建执行器
	exec := ai_agent.NewAgentExecutor(ag, nil)

	// 同步执行
	input := value.NewObjectValue()
	input.PutString("name", "World")

	response, err := exec.Exec(input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success:", response.Success)
	fmt.Println("Result:", response.NodeValue.String())

	// 演示 LLMNode 模板功能
	fmt.Println("\n--- LLMNode Template Demo ---")

	// 创建 LLMNode 使用 Go text/template
	llmNode := node.NewLLMNodeBuilder("llm").
		SystemTemplate("You are a helpful assistant.").
		UserTemplate("Hello, {{.name}}! Your count is {{.count}}.").
		LLMFunction(func(nodeState *node.State, resources *value.ResourcesValue, systemPrompt, userPrompt string, format out.OutFormat, stream bool) (value.NodeValue, error) {
			fmt.Println("System Prompt:", systemPrompt)
			fmt.Println("User Prompt:", userPrompt)

			result := value.NewObjectValue()
			result.PutString("response", "This is a mock LLM response")
			return result, nil
		}).
		Build()

	// 在工作流中使用 LLMNode
	llmWorkflow := ai_agent.Of(llmNode)
	llmAgent := ai_agent.NewAgentBuilder("llm-agent").Workflow(llmWorkflow).Build()
	llmExec := ai_agent.NewAgentExecutor(llmAgent, nil)

	llmInput := value.NewObjectValue()
	llmInput.PutString("name", "Alice")
	llmInput.PutNumber("count", 100)

	llmResponse, err := llmExec.Exec(llmInput)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("LLM Result:", llmResponse.NodeValue.String())
}
