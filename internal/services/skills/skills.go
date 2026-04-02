package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Skill 技能定义
// 对应 TypeScript: src/services/skills/ 技能结构
// 表示一个可执行的技能或命令
type Skill struct {
	Name        string                 `json:"name"`                  // 技能名称
	Description string                 `json:"description,omitempty"` // 技能描述
	Source      string                 `json:"source"`                // 来源 (local/command/builtin)
	Type        SkillType              `json:"type"`                  // 类型 (prompt/agent/builtin)
	Content     string                 `json:"content,omitempty"`     // 技能内容
	Prompt      string                 `json:"prompt,omitempty"`      // 提示词模板
	Tools       []string               `json:"tools,omitempty"`       // 需要的工具列表
	Model       string                 `json:"model,omitempty"`       // 使用的模型
	Effort      int                    `json:"effort,omitempty"`      // 估计的工作量
	Metadata    map[string]interface{} `json:"metadata,omitempty"`    // 元数据
	Aliases     []string               `json:"aliases,omitempty"`     // 别名
	LoadedFrom  string                 `json:"loadedFrom,omitempty"`  // 加载来源路径
	LoadedAt    time.Time              `json:"loadedAt"`              // 加载时间
}

// SkillType 技能类型
// 对应 TypeScript: 技能类型枚举
type SkillType string

const (
	SkillTypePrompt  SkillType = "prompt"  // 提示词类型技能
	SkillTypeAgent   SkillType = "agent"   // Agent类型技能
	SkillTypeBuiltin SkillType = "builtin" // 内置技能
)

// Command 命令结构
// 对应 TypeScript: 命令格式
// 包装Skill并添加路径和参数信息
type Command struct {
	Skill
	Path          string   `json:"path"`          // 命令路径
	ArgumentNames []string `json:"argumentNames"` // 参数名称列表
}

// SkillManager 技能管理器
// 对应 TypeScript: SkillManager
// 负责技能的加载、搜索和管理
type SkillManager struct {
	mu     sync.RWMutex
	skills map[string]*Skill
	paths  []string // 搜索路径
	loaded bool     // 是否已加载
}

// NewSkillManager 创建技能管理器
func NewSkillManager() *SkillManager {
	return &SkillManager{
		skills: make(map[string]*Skill),
		paths:  make([]string, 0),
	}
}

// AddSearchPath 添加技能搜索路径
// 对应 TypeScript: 添加技能目录
// 扫描此路径下的SKILL.md文件来发现技能
func (m *SkillManager) AddSearchPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paths = append(m.paths, path)
}

// LoadSkills 加载所有技能
// 对应 TypeScript: 加载技能目录
// 遍历搜索路径，解析SKILL.md和command.md文件
func (m *SkillManager) LoadSkills(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loaded {
		return nil
	}

	for _, basePath := range m.paths {
		if err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				skillFile := filepath.Join(path, "SKILL.md")
				if data, err := os.ReadFile(skillFile); err == nil {
					skill := m.parseSkillFile(path, string(data))
					m.skills[skill.Name] = skill
				}

				commandFile := filepath.Join(path, "command.md")
				if data, err := os.ReadFile(commandFile); err == nil {
					skill := m.parseCommandFile(path, string(data))
					m.skills[skill.Name] = skill
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	m.loaded = true
	return nil
}

// parseSkillFile 解析SKILL.md文件
// 对应 TypeScript: 解析技能文件
// 提取frontmatter元数据和正文内容
func (m *SkillManager) parseSkillFile(dir string, content string) *Skill {
	name := filepath.Base(dir)

	skill := &Skill{
		Name:       name,
		Source:     "local",
		Type:       SkillTypePrompt,
		Content:    content,
		LoadedAt:   time.Now(),
		LoadedFrom: dir,
	}

	lines := strings.Split(content, "\n")
	frontmatter := false
	frontmatterLines := []string{}
	bodyLines := []string{}

	// 解析YAML frontmatter (--- ... ---)
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if !frontmatter {
				frontmatter = true
				continue
			} else {
				frontmatter = false
				continue
			}
		}

		if frontmatter {
			frontmatterLines = append(frontmatterLines, line)
		} else {
			bodyLines = append(bodyLines, line)
		}
	}

	// 解析frontmatter字段
	for _, fml := range frontmatterLines {
		parts := strings.SplitN(fml, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			skill.Name = value
		case "description":
			skill.Description = value
		case "type":
			skill.Type = SkillType(value)
		case "tools":
			skill.Tools = strings.Split(value, ",")
		case "model":
			skill.Model = value
		}
	}

	// 正文作为prompt
	skill.Prompt = strings.Join(bodyLines, "\n")

	return skill
}

// parseCommandFile 解析command.md文件
// 对应 TypeScript: 解析命令文件
// 与parseSkillFile类似但解析别名等信息
func (m *SkillManager) parseCommandFile(dir string, content string) *Skill {
	name := filepath.Base(dir)

	skill := &Skill{
		Name:       name,
		Source:     "command",
		Type:       SkillTypePrompt,
		Content:    content,
		LoadedAt:   time.Now(),
		LoadedFrom: dir,
	}

	lines := strings.Split(content, "\n")
	frontmatter := false
	frontmatterLines := []string{}
	bodyLines := []string{}

	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if !frontmatter {
				frontmatter = true
				continue
			} else {
				frontmatter = false
				continue
			}
		}

		if frontmatter {
			frontmatterLines = append(frontmatterLines, line)
		} else {
			bodyLines = append(bodyLines, line)
		}
	}

	for _, fml := range frontmatterLines {
		parts := strings.SplitN(fml, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			skill.Name = value
		case "description":
			skill.Description = value
		case "aliases":
			skill.Aliases = strings.Split(value, ",")
		}
	}

	skill.Prompt = strings.Join(bodyLines, "\n")

	return skill
}

// GetSkill 获取指定名称的技能
// 对应 TypeScript: 获取技能
func (m *SkillManager) GetSkill(name string) (*Skill, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	skill, ok := m.skills[name]
	return skill, ok
}

// ListSkills 列出所有技能
// 对应 TypeScript: 列出所有技能
func (m *SkillManager) ListSkills() []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	skills := make([]*Skill, 0, len(m.skills))
	for _, skill := range m.skills {
		skills = append(skills, skill)
	}
	return skills
}

// SearchSkills 搜索技能
// 对应 TypeScript: 搜索技能
// 在名称和描述中搜索匹配的关键字
func (m *SkillManager) SearchSkills(query string) []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	query = strings.ToLower(query)
	results := make([]*Skill, 0)

	for _, skill := range m.skills {
		if strings.Contains(strings.ToLower(skill.Name), query) {
			results = append(results, skill)
			continue
		}
		if strings.Contains(strings.ToLower(skill.Description), query) {
			results = append(results, skill)
			continue
		}
	}

	return results
}

// Reload 重新加载所有技能
// 对应 TypeScript: 重新加载技能
// 清空现有技能并重新扫描
func (m *SkillManager) Reload(ctx context.Context) error {
	m.mu.Lock()
	m.skills = make(map[string]*Skill)
	m.loaded = false
	m.mu.Unlock()

	return m.LoadSkills(ctx)
}

// Registry 命令注册表
// 对应 TypeScript: 命令注册表
// 管理可通过slash命令调用的技能
type Registry struct {
	mu       sync.RWMutex
	commands map[string]*Command
}

// NewRegistry 创建新的命令注册表
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*Command),
	}
}

// Register 注册命令
// 对应 TypeScript: 注册slash命令
func (r *Registry) Register(cmd *Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[cmd.Name]; exists {
		return fmt.Errorf("command already registered: %s", cmd.Name)
	}

	r.commands[cmd.Name] = cmd

	// 同时注册别名
	for _, alias := range cmd.Aliases {
		if _, exists := r.commands[alias]; exists {
			continue
		}
		r.commands[alias] = cmd
	}

	return nil
}

// Unregister 取消注册命令
// 对应 TypeScript: 卸载命令
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cmd, exists := r.commands[name]
	if !exists {
		return fmt.Errorf("command not registered: %s", name)
	}

	delete(r.commands, name)

	// 移除别名
	for _, alias := range cmd.Aliases {
		delete(r.commands, alias)
	}

	return nil
}

// Get 获取命令
func (r *Registry) Get(name string) (*Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, ok := r.commands[name]
	return cmd, ok
}

// List 列出所有命令
func (r *Registry) List() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]*Command, 0, len(r.commands))
	seen := make(map[string]bool)

	// 去重，只返回主命令
	for _, cmd := range r.commands {
		if seen[cmd.Name] {
			continue
		}
		seen[cmd.Name] = true
		commands = append(commands, cmd)
	}

	return commands
}

// FindByPrefix 按前缀查找命令
// 对应 TypeScript: 模糊匹配命令
func (r *Registry) FindByPrefix(prefix string) (*Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, cmd := range r.commands {
		if strings.HasPrefix(name, prefix) {
			return cmd, true
		}
	}

	return nil, false
}

// LoadSkillsDir 加载指定目录下的所有技能
// 对应 TypeScript: 加载技能目录
// 搜索并加载目录下所有SKILL.md文件
func LoadSkillsDir(ctx context.Context, dir string) ([]*Skill, error) {
	skills := make([]*Skill, 0)

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			skillFile := filepath.Join(path, "SKILL.md")
			if data, err := os.ReadFile(skillFile); err == nil {
				name := filepath.Base(path)
				skill := &Skill{
					Name:       name,
					Source:     "local",
					Type:       SkillTypePrompt,
					Content:    string(data),
					LoadedAt:   time.Now(),
					LoadedFrom: path,
				}
				skills = append(skills, skill)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return skills, nil
}

// ToJSON 将技能序列化为JSON
func (s *Skill) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// SkillFromJSON 从JSON反序列化技能
func SkillFromJSON(data []byte) (*Skill, error) {
	var skill Skill
	if err := json.Unmarshal(data, &skill); err != nil {
		return nil, err
	}
	return &skill, nil
}
