package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ConfigManager struct {
	projectConfig map[string]*ServerConfig
	userConfig    map[string]*ServerConfig
	dynamicConfig map[string]*ServerConfig
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		projectConfig: make(map[string]*ServerConfig),
		userConfig:    make(map[string]*ServerConfig),
		dynamicConfig: make(map[string]*ServerConfig),
	}
}

type McpJsonConfig struct {
	MCPServers map[string]*ServerConfig `json:"mcpServers"`
}

func (cm *ConfigManager) LoadProjectConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var config McpJsonConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	cm.projectConfig = config.MCPServers
	return nil
}

func (cm *ConfigManager) LoadUserConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var config McpJsonConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	cm.userConfig = config.MCPServers
	return nil
}

func (cm *ConfigManager) AddServer(name string, config *ServerConfig, scope ConfigScope) {
	switch scope {
	case ScopeProject:
		cm.projectConfig[name] = config
	case ScopeUser:
		cm.userConfig[name] = config
	case ScopeDynamic:
		cm.dynamicConfig[name] = config
	}
}

func (cm *ConfigManager) RemoveServer(name string, scope ConfigScope) {
	switch scope {
	case ScopeProject:
		delete(cm.projectConfig, name)
	case ScopeUser:
		delete(cm.userConfig, name)
	case ScopeDynamic:
		delete(cm.dynamicConfig, name)
	}
}

func (cm *ConfigManager) GetServer(name string) (*ServerConfig, ConfigScope) {
	if config, ok := cm.dynamicConfig[name]; ok {
		return config, ScopeDynamic
	}
	if config, ok := cm.projectConfig[name]; ok {
		return config, ScopeProject
	}
	if config, ok := cm.userConfig[name]; ok {
		return config, ScopeUser
	}
	return nil, ""
}

func (cm *ConfigManager) GetAllServers() map[string]*ServerConfig {
	result := make(map[string]*ServerConfig)
	for k, v := range cm.userConfig {
		result[k] = v
	}
	for k, v := range cm.projectConfig {
		result[k] = v
	}
	for k, v := range cm.dynamicConfig {
		result[k] = v
	}
	return result
}

func (cm *ConfigManager) SaveUserConfig(path string) error {
	config := McpJsonConfig{
		MCPServers: cm.userConfig,
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func GetDefaultMcpDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".claude", "mcp")
}

func GetDefaultMcpConfigPath() string {
	return filepath.Join(GetDefaultMcpDir(), "servers.json")
}

func GetProjectMcpConfigPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".mcp.json")
}

func ExpandEnvVarsInString(input string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		var varName string
		if strings.HasPrefix(match, "${") {
			varName = match[2 : len(match)-1]
		} else {
			varName = match[1:]
		}
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match
	})
}

func ExpandEnvVarsInConfig(config *ServerConfig) *ServerConfig {
	expanded := &ServerConfig{
		Name:      config.Name,
		Transport: config.Transport,
		Command:   ExpandEnvVarsInString(config.Command),
		URL:       ExpandEnvVarsInString(config.URL),
		Scope:     config.Scope,
	}
	expanded.Args = make([]string, len(config.Args))
	for i, arg := range config.Args {
		expanded.Args[i] = ExpandEnvVarsInString(arg)
	}
	expanded.Env = make(map[string]string)
	for k, v := range config.Env {
		expanded.Env[ExpandEnvVarsInString(k)] = ExpandEnvVarsInString(v)
	}
	expanded.Headers = make(map[string]string)
	for k, v := range config.Headers {
		expanded.Headers[ExpandEnvVarsInString(k)] = ExpandEnvVarsInString(v)
	}
	return expanded
}
