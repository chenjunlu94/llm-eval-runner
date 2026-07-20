// Package main 演示用 Go 标准库 net/http 调用 DeepSeek 对话 API 并打印回复。
//
// 这是第一个真正的 Go 程序：不依赖任何第三方 SDK，只用官方标准库，
// 通过 HTTP POST 请求 DeepSeek 的 chat/completions 接口，拿到 AI 回复后打印。
//
// 用法：
//  1. 在 https://platform.deepseek.com/api_keys 申请 API Key
//  2. 设置环境变量：set DEEPSEEK_API_KEY=你的Key   (Windows CMD)
//     $env:DEEPSEEK_API_KEY="你的Key" (PowerShell)
//     export DEEPSEEK_API_KEY="你的Key" (Git Bash)
//  3. 运行：go run main.go

// 一个go文件需要在一个包中
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// DeepSeek 对话补全接口地址（OpenAI 兼容格式，国内可直连无需代理）
	apiURL = "https://api.deepseek.com/chat/completions"
	// 使用的模型：deepseek-v4-flash 是快速通用对话模型
	// （旧的 deepseek-chat 将于 2026/07/24 弃用，故用新模型名）
	modelName = "deepseek-v4-flash"
	// 默认单轮对话内容
	defaultPrompt = "你好，请用一句话介绍你自己。"
)

// ChatRequest 对应 DeepSeek API 的请求体（OpenAI 兼容格式）。
type ChatRequest struct {
	Model    string    `json:"model"`    // 模型名
	Messages []Message `json:"messages"` // 对话消息列表
}

// Message 对应一条对话消息。
type Message struct {
	Role    string `json:"role"`    // 角色：system / user / assistant
	Content string `json:"content"` // 消息内容
}

// ChatResponse 对应 DeepSeek API 的响应体（只定义用到的字段，其余忽略）。
type ChatResponse struct {
	Choices []struct {
		Message      Message `json:"message"`       // 本次回复内容
		FinishReason string  `json:"finish_reason"` // 结束原因：stop / length 等
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`     // 输入 token 数
		CompletionTokens int `json:"completion_tokens"` // 输出 token 数
		TotalTokens      int `json:"total_tokens"`      // 合计 token 数
	} `json:"usage"`
}

func main() {
	// 1. 从环境变量读取 API Key（密钥不硬编码进源码，安全且便于切换）
	//必知语法，“：=”是短变量声明，如“A := B”，右边B的值返回给左边短变量的A
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "错误：未设置环境变量 DEEPSEEK_API_KEY")
		fmt.Fprintln(os.Stderr, "请先在 https://platform.deepseek.com/api_keys 申请 Key，然后：")
		fmt.Fprintln(os.Stderr, "  PowerShell: $env:DEEPSEEK_API_KEY=\"你的Key\"")
		fmt.Fprintln(os.Stderr, "  Git Bash : export DEEPSEEK_API_KEY=\"你的Key\"")
		os.Exit(1)
	}

	// 2. 调用 API 并打印回复；出错则打印错误并退出码 1
	if err := chatWithDeepSeek(apiKey); err != nil {
		fmt.Fprintf(os.Stderr, "调用失败：%v\n", err)
		os.Exit(1)
	}
}

// chatWithDeepSeek 完成一次对话请求：构造请求 → 发送 → 解析 → 打印回复。
// 把核心逻辑拆成独立函数，便于复用和测试（符合 Uber Go Style Guide 的小函数原则）。
func chatWithDeepSeek(apiKey string) error {
	// —— 构造请求体 ——
	reqBody := ChatRequest{
		Model: modelName, // 模型名
		Messages: []Message{ // 对话消息列表,问题内容是个数组，里面每条消息有角色（user）和内容（你好...）
			{Role: "user", Content: defaultPrompt},
		},
	}
	bodyBytes, err := json.Marshal(reqBody) //把填好的规格表翻译成 JSON 字符串
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	// —— 创建 HTTP 请求 ——
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(bodyBytes)) //方法，请求头，请求体
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	// DeepSeek 用 Bearer Token 认证，贴标签

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// —— 发送请求（带 30 秒超时，避免网络卡死一直挂着）——
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	// defer 关闭响应体，确保资源释放
	defer func() { _ = resp.Body.Close() }()

	// —— 读取响应 ——
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// —— 非 2xx 状态码视为错误，把响应体一起返回方便排查 ——
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API 返回 HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	// —— 解析 JSON 响应 ——
	var chatResp ChatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return errors.New("响应中没有 choices 字段")
	}

	// —— 打印结果 ——
	fmt.Println("=== DeepSeek 回复 ===")
	fmt.Println(chatResp.Choices[0].Message.Content)
	fmt.Printf("\n（token 用量：输入 %d + 输出 %d = 合计 %d）\n",
		chatResp.Usage.PromptTokens,
		chatResp.Usage.CompletionTokens,
		chatResp.Usage.TotalTokens)
	return nil
}
