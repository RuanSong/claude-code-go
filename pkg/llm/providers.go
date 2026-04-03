package llm

import (
	"os"
)

// ProviderConfigBuilder 提供者配置构建器
type ProviderConfigBuilder struct {
	config ProviderConfig
}

// NewProviderConfigBuilder 创建新的配置构建器
func NewProviderConfigBuilder(name string) *ProviderConfigBuilder {
	return &ProviderConfigBuilder{
		config: ProviderConfig{
			Name:  name,
			Extra: make(map[string]any),
		},
	}
}

// WithBaseURL 设置基础URL
func (b *ProviderConfigBuilder) WithBaseURL(url string) *ProviderConfigBuilder {
	b.config.BaseURL = url
	return b
}

// WithAPIKey 直接设置API密钥
func (b *ProviderConfigBuilder) WithAPIKey(key string) *ProviderConfigBuilder {
	b.config.APIKey = key
	return b
}

// WithAPIKeyEnv 通过环境变量名设置API密钥
func (b *ProviderConfigBuilder) WithAPIKeyEnv(envName string) *ProviderConfigBuilder {
	b.config.APIKeyEnv = envName
	b.config.APIKey = os.Getenv(envName)
	return b
}

// WithDefaultModel 设置默认模型
func (b *ProviderConfigBuilder) WithDefaultModel(model string) *ProviderConfigBuilder {
	b.config.DefaultModel = model
	return b
}

// WithCodingModel 设置Coding专用模型
func (b *ProviderConfigBuilder) WithCodingModel(model string) *ProviderConfigBuilder {
	b.config.CodingModel = model
	return b
}

// WithTools 支持工具调用
func (b *ProviderConfigBuilder) WithTools() *ProviderConfigBuilder {
	b.config.Capabilities.SupportsTools = true
	return b
}

// WithVision 支持视觉理解
func (b *ProviderConfigBuilder) WithVision() *ProviderConfigBuilder {
	b.config.Capabilities.SupportsVision = true
	return b
}

// WithThinking 支持思考模式
func (b *ProviderConfigBuilder) WithThinking() *ProviderConfigBuilder {
	b.config.Capabilities.SupportsThinking = true
	return b
}

// WithExtra 设置额外配置
func (b *ProviderConfigBuilder) WithExtra(key string, value any) *ProviderConfigBuilder {
	if b.config.Extra == nil {
		b.config.Extra = make(map[string]any)
	}
	b.config.Extra[key] = value
	return b
}

// Build 构建配置
func (b *ProviderConfigBuilder) Build() ProviderConfig {
	// 如果API密钥环境变量名已设置但API密钥为空，尝试从环境变量获取
	if b.config.APIKey == "" && b.config.APIKeyEnv != "" {
		b.config.APIKey = os.Getenv(b.config.APIKeyEnv)
	}
	return b.config
}

// NewOpenAIProvider 创建 OpenAI Provider
func NewOpenAIProvider() LLMProvider {
	config := NewProviderConfigBuilder("openai").
		WithBaseURL("https://api.openai.com/v1").
		WithAPIKeyEnv("OPENAI_API_KEY").
		WithDefaultModel("gpt-4o").
		WithCodingModel("gpt-4o").
		WithTools().
		WithVision().
		Build()
	return NewOpenAIClient(config)
}

// NewDeepSeekProvider 创建 DeepSeek Provider
func NewDeepSeekProvider() LLMProvider {
	config := NewProviderConfigBuilder("deepseek").
		WithBaseURL("https://api.deepseek.com").
		WithAPIKeyEnv("DEEPSEEK_API_KEY").
		WithDefaultModel("deepseek-chat").
		WithCodingModel("deepseek-coder").
		WithTools().
		Build()
	return NewOpenAIClient(config)
}

// NewKimiProvider 创建 Kimi (Moonshot) Provider
func NewKimiProvider() LLMProvider {
	config := NewProviderConfigBuilder("kimi").
		WithBaseURL("https://api.moonshot.cn/v1").
		WithAPIKeyEnv("MOONSHOT_API_KEY").
		WithDefaultModel("kimi-k2.5").
		WithCodingModel("kimi-k2").
		WithTools().
		WithVision().
		Build()
	return NewOpenAIClient(config)
}

// NewBailianProvider 创建百炼 Provider
func NewBailianProvider() LLMProvider {
	config := NewProviderConfigBuilder("bailian").
		WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1").
		WithAPIKeyEnv("DASHSCOPE_API_KEY").
		WithDefaultModel("qwen-plus").
		WithCodingModel("qwen-coder-plus").
		WithTools().
		WithVision().
		Build()
	return NewOpenAIClient(config)
}

// NewZhipuProvider 创建智谱 Provider
func NewZhipuProvider() LLMProvider {
	config := NewProviderConfigBuilder("zhipu").
		WithBaseURL("https://open.bigmodel.cn/api/pathtowave").
		WithAPIKeyEnv("BIGMODEL_API_KEY").
		WithDefaultModel("glm-4").
		WithCodingModel("glm-4-coder").
		WithTools().
		WithVision().
		Build()
	return NewOpenAIClient(config)
}

// NewMiniMaxProvider 创建 MiniMax Provider
func NewMiniMaxProvider() LLMProvider {
	config := NewProviderConfigBuilder("minimax").
		WithBaseURL("https://api.minimax.chat/v1").
		WithAPIKeyEnv("MINIMAX_API_KEY").
		WithDefaultModel("MiniMax-M2.7").
		WithCodingModel("MiniMax-M2.7").
		WithTools().
		WithVision().
		Build()
	return NewOpenAIClient(config)
}

// NewOpenCodeZenProvider 创建 OpenCode Zen Provider
func NewOpenCodeZenProvider() LLMProvider {
	config := NewProviderConfigBuilder("opencode-zen").
		WithBaseURL("https://opencode.ai/zen/v1").
		WithAPIKeyEnv("OPENCODE_ZEN_API_KEY").
		WithDefaultModel("gpt-4o").
		WithCodingModel("gpt-4o").
		WithTools().
		WithVision().
		Build()
	return NewOpenAIClient(config)
}

// NewOpenCodeGoProvider 创建 OpenCode Go Provider (使用 Anthropic API)
func NewOpenCodeGoProvider() LLMProvider {
	config := NewProviderConfigBuilder("opencode-go").
		WithBaseURL("https://opencode.ai/zen/go/v1").
		WithAPIKeyEnv("OPENCODE_GO_API_KEY").
		WithDefaultModel("opencode-go").
		WithCodingModel("opencode-go").
		WithTools().
		Build()
	return NewAnthropicClient(config)
}

// NewOllamaProvider 创建 Ollama Provider (本地自托管)
func NewOllamaProvider() LLMProvider {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	config := NewProviderConfigBuilder("ollama").
		WithBaseURL(baseURL).
		WithAPIKey(""). // Ollama 不需要 API Key
		WithDefaultModel("llama3").
		WithCodingModel("codellama").
		WithTools().
		Build()
	return NewOpenAIClient(config)
}

// DefaultProviderRegistry 创建默认的提供者注册表
func DefaultProviderRegistry() *ProviderRegistry {
	registry := NewProviderRegistry()

	// 注册所有 Provider
	registry.Register("openai", NewOpenAIProvider())
	registry.Register("deepseek", NewDeepSeekProvider())
	registry.Register("kimi", NewKimiProvider())
	registry.Register("bailian", NewBailianProvider())
	registry.Register("zhipu", NewZhipuProvider())
	registry.Register("minimax", NewMiniMaxProvider())
	registry.Register("opencode-zen", NewOpenCodeZenProvider())
	registry.Register("opencode-go", NewOpenCodeGoProvider())
	registry.Register("ollama", NewOllamaProvider())

	// 设置默认 Provider
	registry.SetDefault("openai")

	return registry
}
