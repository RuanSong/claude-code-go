package commands

import (
	"github.com/claude-code-go/claude/internal/engine"
)

type BaseCommand = engine.BaseCommand

func newCommand(name string, description string) *BaseCommand {
	return engine.NewBaseCommand(engine.CommandTypeCustom, name, description)
}

func newPromptCommand(name string, description string) *BaseCommand {
	return engine.NewBaseCommand(engine.CommandTypePrompt, name, description)
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
	registry.Register(NewMCPCommand())
	registry.Register(NewSkillsCommand())
	registry.Register(NewCostCommand())
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

	return registry
}
