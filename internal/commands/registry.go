package commands

import (
	"time"

	"github.com/claude-code-go/claude/internal/engine"
	"github.com/claude-code-go/claude/internal/services/cost"
)

type BaseCommand = engine.BaseCommand

func newCommand(name string, description string) *BaseCommand {
	return engine.NewBaseCommand(engine.CommandTypeCustom, name, description)
}

func newPromptCommand(name string, description string) *BaseCommand {
	return engine.NewBaseCommand(engine.CommandTypePrompt, name, description)
}

// 全局成本追踪器
var globalCostTracker *cost.CostTracker

// InitGlobalCostTracker 初始化全局成本追踪器
func InitGlobalCostTracker(sessionID string) {
	globalCostTracker = cost.NewCostTracker(sessionID)
}

// GetGlobalCostTracker 获取全局成本追踪器
func GetGlobalCostTracker() *cost.CostTracker {
	return globalCostTracker
}

// Registry manages all available commands
type Registry struct {
	commands map[string]engine.Command
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]engine.Command),
	}
}

func (r *Registry) Register(cmd engine.Command) error {
	name := cmd.Name()
	if name == "" {
		return &RegistryError{Message: "command name cannot be empty"}
	}
	if _, exists := r.commands[name]; exists {
		return &RegistryError{Message: "command already registered: " + name}
	}
	r.commands[name] = cmd
	return nil
}

func (r *Registry) Get(name string) (engine.Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

func (r *Registry) List() []engine.Command {
	cmds := make([]engine.Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	return names
}

// RegistryError represents a registry error
type RegistryError struct {
	Message string
}

func (e *RegistryError) Error() string {
	return e.Message
}

// DefaultRegistry creates the default command registry with built-in commands
func DefaultRegistry() *Registry {
	registry := NewRegistry()

	// 初始化全局成本追踪器
	InitGlobalCostTracker(time.Now().Format("20060102150405"))

	// Register built-in commands
	registry.Register(NewCommitCommand())
	registry.Register(NewDiffCommand())
	registry.Register(NewCompactCommand())
	registry.Register(NewDoctorCommand())
	registry.Register(NewReviewCommand())
	registry.Register(NewUltraReviewCommand())
	registry.Register(NewModelCommand())
	registry.Register(NewConfigCommand())
	registry.Register(NewHelpCommand())
	registry.Register(NewVersionCommand())
	registry.Register(NewInitCommand())
	registry.Register(NewLoginCommand())
	registry.Register(NewLogoutCommand())
	registry.Register(NewMCPCommand(nil))
	registry.Register(NewSkillsCommand())
	registry.Register(NewCostCommand(GetGlobalCostTracker()))
	registry.Register(NewMemoryCommand())
	registry.Register(NewResumeCommand())
	registry.Register(NewShareCommand())
	registry.Register(NewExitCommand())
	registry.Register(NewTasksCommand())
	registry.Register(NewThemeCommand())
	registry.Register(NewKeybindingsCommand())
	registry.Register(NewVimCommand())
	registry.Register(NewContextCommand())
	registry.Register(NewPermissionsCommand())
	registry.Register(NewPlanCommand())
	registry.Register(NewStatusCommand())
	registry.Register(NewClearCommand())
	registry.Register(NewEnvCommand())
	registry.Register(NewBranchCommand())
	registry.Register(NewLogsCommand())
	registry.Register(NewStatsCommand())
	registry.Register(NewHistoryCommand())
	registry.Register(NewHooksCommand())
	registry.Register(NewFilesCommand())
	registry.Register(NewCopyCommand())
	registry.Register(NewExportCommand())
	registry.Register(NewSessionCommand())
	registry.Register(NewRewindCommand())
	registry.Register(NewTeleportCommand())
	registry.Register(NewTagCommand())
	registry.Register(NewEffortCommand())
	registry.Register(NewFeedbackCommand())
	registry.Register(NewRenameCommand())
	registry.Register(NewReleaseNotesCommand())
	registry.Register(NewUsageCommand())
	registry.Register(NewDesktopCommand())
	registry.Register(NewMobileCommand())
	registry.Register(NewIDECommand())
	registry.Register(NewHealthCommand())

	// 注册额外的命令
	registry.Register(NewAgentsCommand())
	registry.Register(NewPluginCommand())
	registry.Register(NewReloadPluginsCommand())
	registry.Register(NewOnboardingCommand())
	registry.Register(NewUpgradeCommand())
	registry.Register(NewVoiceCommand())
	registry.Register(NewBtwCommand())
	registry.Register(NewThinkbackCommand())
	registry.Register(NewSandboxCommand())
	registry.Register(NewPrivacyCommand())
	registry.Register(NewAutoFixCommand())
	registry.Register(NewHeapdumpCommand())
	registry.Register(NewChromeCommand())
	registry.Register(NewSummaryCommand())
	registry.Register(NewRateLimitCommand())

	// 注册缺失的命令
	registry.Register(NewAddDirCommand())
	registry.Register(NewColorCommand())
	registry.Register(NewFastCommand())
	registry.Register(NewInsightsCommand())
	registry.Register(NewInitVerifiersCommand())
	registry.Register(NewPassesCommand())
	registry.Register(NewStickersCommand())
	registry.Register(NewExtraUsageCommand())
	registry.Register(NewRemoteEnvCommand())
	registry.Register(NewCommitPushPRCommand())
	registry.Register(NewOutputStyleCommand())
	registry.Register(NewTerminalSetupCommand())
	registry.Register(NewSecurityReviewCommand())
	registry.Register(NewAdvisorCommand())
	registry.Register(NewThinkbackPlayCommand())
	registry.Register(NewRemoteSetupCommand())
	registry.Register(NewBridgeCommand())
	registry.Register(NewBridgeKickCommand())
	registry.Register(NewInstallGitHubAppCommand())
	registry.Register(NewInstallSlackAppCommand())
	registry.Register(NewStatuslineCommand())
	registry.Register(NewBriefCommand())
	registry.Register(NewInstallCommand())
	registry.Register(NewUltraplanCommand())
	registry.Register(NewIssueCommand())
	registry.Register(NewPRCommentsCommand())
	registry.Register(NewBugHunterCommand())
	registry.Register(NewAntTraceCommand())
	registry.Register(NewBreakCacheCommand())
	registry.Register(NewGoodClaudeCommand())
	registry.Register(NewResetLimitsCommand())
	registry.Register(NewDebugToolCallCommand())
	registry.Register(NewPerfIssueCommand())
	registry.Register(NewOAuthRefreshCommand())

	return registry
}
