package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Plugin 插件结构
// 对应 TypeScript: src/services/plugins/ 插件定义
// 存储插件的所有元信息和功能
type Plugin struct {
	Name        string                 `json:"name"`                 // 插件唯一名称
	Version     string                 `json:"version"`              // 插件版本
	Description string                 `json:"description"`          // 插件描述
	Author      string                 `json:"author"`               // 插件作者
	Repository  string                 `json:"repository,omitempty"` // 插件仓库URL
	Enabled     bool                   `json:"enabled"`              // 是否启用
	LoadedAt    time.Time              `json:"loaded_at"`            // 加载时间
	Commands    []PluginCommand        `json:"commands,omitempty"`   // 插件提供的命令
	Tools       []PluginTool           `json:"tools,omitempty"`      // 插件提供的工具
	Metadata    map[string]interface{} `json:"metadata,omitempty"`   // 自定义元数据
}

// PluginCommand 插件命令定义
// 对应 TypeScript: 插件命令格式
type PluginCommand struct {
	Name        string `json:"name"`        // 命令名称
	Description string `json:"description"` // 命令描述
	Type        string `json:"type"`        // 命令类型 (prompt/custom)
}

// PluginTool 插件工具定义
// 对应 TypeScript: 插件工具格式
type PluginTool struct {
	Name        string                 `json:"name"`                   // 工具名称
	Description string                 `json:"description"`            // 工具描述
	InputSchema map[string]interface{} `json:"input_schema,omitempty"` // 输入参数模式
}

// PluginManager 插件管理器
// 对应 TypeScript: PluginManager
// 负责插件的加载、卸载和状态管理
type PluginManager struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
	paths   []string // 插件搜索路径
}

// NewPluginManager 创建新的插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*Plugin),
		paths:   make([]string, 0),
	}
}

// AddSearchPath 添加插件搜索路径
// 对应 TypeScript: 添加插件目录
// 扫描此路径下的plugin.json文件来发现插件
func (m *PluginManager) AddSearchPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paths = append(m.paths, path)
}

// LoadPlugin 加载插件
// 对应 TypeScript: installPluginOp()
// 将插件注册到管理器并标记为已启用
func (m *PluginManager) LoadPlugin(name string, plugin *Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin already loaded: %s", name)
	}

	plugin.LoadedAt = time.Now()
	plugin.Enabled = true
	m.plugins[name] = plugin
	return nil
}

// UnloadPlugin 卸载插件
// 对应 TypeScript: uninstallPluginOp()
// 从管理器中移除插件
func (m *PluginManager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; !exists {
		return fmt.Errorf("plugin not loaded: %s", name)
	}

	delete(m.plugins, name)
	return nil
}

// GetPlugin 获取指定名称的插件
// 对应 TypeScript: 获取插件信息
func (m *PluginManager) GetPlugin(name string) (*Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	plugin, ok := m.plugins[name]
	return plugin, ok
}

// ListPlugins 列出所有已加载的插件
// 对应 TypeScript: 获取已安装插件列表
func (m *PluginManager) ListPlugins() []*Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// EnablePlugin 启用插件
// 对应 TypeScript: enablePluginOp()
func (m *PluginManager) EnablePlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not loaded: %s", name)
	}

	plugin.Enabled = true
	return nil
}

// DisablePlugin 禁用插件
// 对应 TypeScript: disablePluginOp()
func (m *PluginManager) DisablePlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not loaded: %s", name)
	}

	plugin.Enabled = false
	return nil
}

// ScanPlugins 扫描搜索路径下的所有插件
// 对应 TypeScript: 发现和扫描插件
// 遍历搜索路径查找plugin.json文件并解析
func (m *PluginManager) ScanPlugins(ctx context.Context) ([]*Plugin, error) {
	m.mu.RLock()
	paths := make([]string, len(m.paths))
	copy(paths, m.paths)
	m.mu.RUnlock()

	plugins := make([]*Plugin, 0)

	for _, basePath := range paths {
		if err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if !info.IsDir() && info.Name() == "plugin.json" {
				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}

				var plugin Plugin
				if err := json.Unmarshal(data, &plugin); err != nil {
					return nil
				}

				plugins = append(plugins, &plugin)
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return plugins, nil
}

// HookFunc 插件钩子函数类型
// 对应 TypeScript: 插件生命周期钩子
// 在插件加载、卸载等事件触发时调用
type HookFunc func(ctx context.Context, plugin *Plugin) error

// CommandHandler 插件命令处理器类型
// 对应 TypeScript: 插件命令处理函数
type CommandHandler func(ctx context.Context, args []string) (string, error)

// PluginRegistry 插件注册表
// 对应 TypeScript: 插件注册中心
// 管理插件注册、钩子和命令处理器
type PluginRegistry struct {
	mu              sync.RWMutex
	plugins         map[string]*Plugin
	hooks           map[string][]HookFunc     // 事件 -> 钩子列表
	commandHandlers map[string]CommandHandler // 命令名 -> 处理器
}

// NewPluginRegistry 创建新的插件注册表
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins:         make(map[string]*Plugin),
		hooks:           make(map[string][]HookFunc),
		commandHandlers: make(map[string]CommandHandler),
	}
}

// Register 注册插件
// 对应 TypeScript: 注册插件
func (r *PluginRegistry) Register(plugin *Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[plugin.Name]; exists {
		return fmt.Errorf("plugin already registered: %s", plugin.Name)
	}

	r.plugins[plugin.Name] = plugin
	return nil
}

// Unregister 取消注册插件
// 对应 TypeScript: 卸载插件
func (r *PluginRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin not registered: %s", name)
	}

	delete(r.plugins, name)
	return nil
}

// Get 获取已注册的插件
func (r *PluginRegistry) Get(name string) (*Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plugin, ok := r.plugins[name]
	return plugin, ok
}

// List 列出所有已注册的插件
func (r *PluginRegistry) List() []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// RegisterHook 注册插件钩子
// 对应 TypeScript: 注册生命周期钩子
// 事件包括: onLoad, onUnload, onEnable, onDisable 等
func (r *PluginRegistry) RegisterHook(event string, hook HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[event] = append(r.hooks[event], hook)
}

// EmitHook 触发插件钩子
// 对应 TypeScript: 调用生命周期钩子
// 执行所有注册到该事件的钩子函数
func (r *PluginRegistry) EmitHook(ctx context.Context, event string, plugin *Plugin) error {
	r.mu.RLock()
	hooks := make([]HookFunc, len(r.hooks[event]))
	copy(hooks, r.hooks[event])
	r.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook(ctx, plugin); err != nil {
			return err
		}
	}

	return nil
}

// RegisterCommand 注册插件命令处理器
// 对应 TypeScript: 注册slash命令
func (r *PluginRegistry) RegisterCommand(name string, handler CommandHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commandHandlers[name] = handler
}

// HandleCommand 处理插件命令
// 对应 TypeScript: 执行插件命令
func (r *PluginRegistry) HandleCommand(ctx context.Context, name string, args []string) (string, error) {
	r.mu.RLock()
	handler, exists := r.commandHandlers[name]
	r.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("command handler not found: %s", name)
	}

	return handler(ctx, args)
}

// GetCommandNames 获取所有已注册的命令名称
func (r *PluginRegistry) GetCommandNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.commandHandlers))
	for name := range r.commandHandlers {
		names = append(names, name)
	}
	return names
}

// PluginAPI 插件API接口
// 对应 TypeScript: 暴露给插件的API
// 提供插件与主程序交互的接口
type PluginAPI struct {
	manager  *PluginManager
	registry *PluginRegistry
}

// NewPluginAPI 创建插件API接口
func NewPluginAPI(manager *PluginManager, registry *PluginRegistry) *PluginAPI {
	return &PluginAPI{
		manager:  manager,
		registry: registry,
	}
}

// LoadBuiltInPlugins 加载内置插件
// 对应 TypeScript: 内置命令插件
// 注册内置的help、model、cost等命令
func (api *PluginAPI) LoadBuiltInPlugins() error {
	builtInPlugins := []*Plugin{
		{
			Name:        "built-in-commands",
			Version:     "1.0.0",
			Description: "Built-in slash commands",
			Author:      "claude-code-go",
			Commands: []PluginCommand{
				{Name: "help", Description: "Show help information", Type: "prompt"},
				{Name: "model", Description: "Get or set the model", Type: "custom"},
				{Name: "cost", Description: "Show usage and cost", Type: "custom"},
			},
		},
	}

	for _, plugin := range builtInPlugins {
		if err := api.registry.Register(plugin); err != nil {
			return err
		}
	}

	return nil
}
