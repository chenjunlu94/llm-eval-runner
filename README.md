# GoLang 学习笔记

Go 语言新手项目的练习仓库，包含一个调用 DeepSeek API 的示例和一个最简单的 hello world。

## 环境要求

- Go 1.26.5 或更高版本
- 模块模式（GO111MODULE=on）
- 已配置国内代理（goproxy.cn），依赖下载更快

## 目录结构

```
GoLang/
├── deepseek-demo/      # 用 net/http 调用 DeepSeek API 的示例
│   ├── main.go         # 主要代码：构造请求 → 发送 → 打印回复
│   ├── go.mod
│   └── go.sum
├── test-hello/         # 最简单的 Go 项目，验证环境是否配置正确
│   ├── main.go
│   ├── go.mod
│   └── go.sum
├── .gitignore
├── LICENSE
└── README.md
```

## 快速开始

### 1. 验证环境

```bash
cd test-hello
go run main.go
```

输出 `Hello, world.` 说明环境正常。

### 2. 调用 DeepSeek API

```bash
# 先设置 API Key（从 https://platform.deepseek.com/api_keys 申请）
export DEEPSEEK_API_KEY="你的Key"

# 运行
cd deepseek-demo
go run main.go
```

程序会向 DeepSeek 发送一条消息并打印回复，同时显示 token 用量。

## 许可证

[MIT](LICENSE) © 2026 大小丽
