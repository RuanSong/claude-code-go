package plugins

import (
	"context"
	"testing"
	"time"
)

func TestNewPluginManager(t *testing.T) {
	manager := NewPluginManager()

	if manager == nil {
		t.Fatal("NewPluginManager() returned nil")
	}

	if manager.plugins == nil {
		t.Error("NewPluginManager() did not initialize plugins map")
	}

	if manager.paths == nil {
		t.Error("NewPluginManager() did not initialize paths slice")
	}
}

func TestPluginManager_AddSearchPath(t *testing.T) {
	manager := NewPluginManager()

	manager.AddSearchPath("/path/to/plugins")
	manager.AddSearchPath("/another/path")

	if len(manager.paths) != 2 {
		t.Errorf("AddSearchPath() added %d paths, want 2", len(manager.paths))
	}
}

func TestPluginManager_LoadPlugin(t *testing.T) {
	manager := NewPluginManager()

	plugin := &Plugin{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Author:      "Test Author",
	}

	err := manager.LoadPlugin("test-plugin", plugin)
	if err != nil {
		t.Fatalf("LoadPlugin() error = %v", err)
	}

	// Verify plugin was loaded
	retrieved, ok := manager.GetPlugin("test-plugin")
	if !ok {
		t.Fatal("LoadPlugin() did not register plugin")
	}

	if retrieved.Name != "test-plugin" {
		t.Error("LoadPlugin() stored wrong plugin")
	}
}

func TestPluginManager_LoadPlugin_AlreadyLoaded(t *testing.T) {
	manager := NewPluginManager()

	plugin := &Plugin{Name: "test-plugin"}
	manager.LoadPlugin("test-plugin", plugin)

	// Try to load again
	err := manager.LoadPlugin("test-plugin", plugin)
	if err == nil {
		t.Error("LoadPlugin() should return error for already loaded plugin")
	}
}

func TestPluginManager_UnloadPlugin(t *testing.T) {
	manager := NewPluginManager()

	plugin := &Plugin{Name: "test-plugin"}
	manager.LoadPlugin("test-plugin", plugin)

	err := manager.UnloadPlugin("test-plugin")
	if err != nil {
		t.Fatalf("UnloadPlugin() error = %v", err)
	}

	// Verify plugin was unloaded
	_, ok := manager.GetPlugin("test-plugin")
	if ok {
		t.Error("UnloadPlugin() did not remove plugin")
	}
}

func TestPluginManager_UnloadPlugin_NotLoaded(t *testing.T) {
	manager := NewPluginManager()

	err := manager.UnloadPlugin("non-existent")
	if err == nil {
		t.Error("UnloadPlugin() should return error for non-loaded plugin")
	}
}

func TestPluginManager_GetPlugin(t *testing.T) {
	manager := NewPluginManager()

	plugin := &Plugin{
		Name:    "test-plugin",
		Version: "1.0.0",
	}
	manager.LoadPlugin("test-plugin", plugin)

	retrieved, ok := manager.GetPlugin("test-plugin")
	if !ok {
		t.Fatal("GetPlugin() returned ok = false for loaded plugin")
	}

	if retrieved.Version != "1.0.0" {
		t.Error("GetPlugin() returned wrong plugin")
	}
}

func TestPluginManager_GetPlugin_NotFound(t *testing.T) {
	manager := NewPluginManager()

	_, ok := manager.GetPlugin("non-existent")
	if ok {
		t.Error("GetPlugin() returned ok = true for non-existent plugin")
	}
}

func TestPluginManager_ListPlugins(t *testing.T) {
	manager := NewPluginManager()

	manager.LoadPlugin("plugin1", &Plugin{Name: "plugin1"})
	manager.LoadPlugin("plugin2", &Plugin{Name: "plugin2"})

	plugins := manager.ListPlugins()

	if len(plugins) != 2 {
		t.Errorf("ListPlugins() returned %d plugins, want 2", len(plugins))
	}
}

func TestPluginManager_EnablePlugin(t *testing.T) {
	manager := NewPluginManager()

	plugin := &Plugin{Name: "test-plugin", Enabled: false}
	manager.LoadPlugin("test-plugin", plugin)

	err := manager.EnablePlugin("test-plugin")
	if err != nil {
		t.Fatalf("EnablePlugin() error = %v", err)
	}

	retrieved, _ := manager.GetPlugin("test-plugin")
	if !retrieved.Enabled {
		t.Error("EnablePlugin() did not set Enabled = true")
	}
}

func TestPluginManager_EnablePlugin_NotLoaded(t *testing.T) {
	manager := NewPluginManager()

	err := manager.EnablePlugin("non-existent")
	if err == nil {
		t.Error("EnablePlugin() should return error for non-loaded plugin")
	}
}

func TestPluginManager_DisablePlugin(t *testing.T) {
	manager := NewPluginManager()

	plugin := &Plugin{Name: "test-plugin", Enabled: true}
	manager.LoadPlugin("test-plugin", plugin)

	err := manager.DisablePlugin("test-plugin")
	if err != nil {
		t.Fatalf("DisablePlugin() error = %v", err)
	}

	retrieved, _ := manager.GetPlugin("test-plugin")
	if retrieved.Enabled {
		t.Error("DisablePlugin() did not set Enabled = false")
	}
}

func TestNewPluginRegistry(t *testing.T) {
	registry := NewPluginRegistry()

	if registry == nil {
		t.Fatal("NewPluginRegistry() returned nil")
	}

	if registry.plugins == nil {
		t.Error("NewPluginRegistry() did not initialize plugins map")
	}

	if registry.hooks == nil {
		t.Error("NewPluginRegistry() did not initialize hooks map")
	}

	if registry.commandHandlers == nil {
		t.Error("NewPluginRegistry() did not initialize commandHandlers map")
	}
}

func TestPluginRegistry_Register(t *testing.T) {
	registry := NewPluginRegistry()

	plugin := &Plugin{Name: "test-plugin"}
	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	retrieved, ok := registry.Get("test-plugin")
	if !ok {
		t.Fatal("Register() did not register plugin")
	}

	if retrieved.Name != "test-plugin" {
		t.Error("Register() stored wrong plugin")
	}
}

func TestPluginRegistry_Register_AlreadyRegistered(t *testing.T) {
	registry := NewPluginRegistry()

	plugin := &Plugin{Name: "test-plugin"}
	registry.Register(plugin)

	err := registry.Register(plugin)
	if err == nil {
		t.Error("Register() should return error for already registered plugin")
	}
}

func TestPluginRegistry_Unregister(t *testing.T) {
	registry := NewPluginRegistry()

	plugin := &Plugin{Name: "test-plugin"}
	registry.Register(plugin)

	err := registry.Unregister("test-plugin")
	if err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}

	_, ok := registry.Get("test-plugin")
	if ok {
		t.Error("Unregister() did not remove plugin")
	}
}

func TestPluginRegistry_Unregister_NotRegistered(t *testing.T) {
	registry := NewPluginRegistry()

	err := registry.Unregister("non-existent")
	if err == nil {
		t.Error("Unregister() should return error for non-registered plugin")
	}
}

func TestPluginRegistry_List(t *testing.T) {
	registry := NewPluginRegistry()

	registry.Register(&Plugin{Name: "plugin1"})
	registry.Register(&Plugin{Name: "plugin2"})

	plugins := registry.List()

	if len(plugins) != 2 {
		t.Errorf("List() returned %d plugins, want 2", len(plugins))
	}
}

func TestPluginRegistry_RegisterHook(t *testing.T) {
	registry := NewPluginRegistry()

	hookCalled := false
	hook := func(ctx context.Context, plugin *Plugin) error {
		hookCalled = true
		return nil
	}

	registry.RegisterHook("onLoad", hook)

	if len(registry.hooks["onLoad"]) != 1 {
		t.Error("RegisterHook() did not register hook")
	}

	_ = hookCalled // silence unused variable warning
}

func TestPluginRegistry_EmitHook(t *testing.T) {
	registry := NewPluginRegistry()

	hookCalled := false
	hook := func(ctx context.Context, plugin *Plugin) error {
		hookCalled = true
		return nil
	}

	registry.RegisterHook("onLoad", hook)

	plugin := &Plugin{Name: "test-plugin"}
	err := registry.EmitHook(context.Background(), "onLoad", plugin)
	if err != nil {
		t.Fatalf("EmitHook() error = %v", err)
	}

	if !hookCalled {
		t.Error("Hook was not called")
	}
}

func TestPluginRegistry_EmitHook_Error(t *testing.T) {
	registry := NewPluginRegistry()

	expectedErr := context.DeadlineExceeded
	hook := func(ctx context.Context, plugin *Plugin) error {
		return expectedErr
	}

	registry.RegisterHook("onLoad", hook)

	plugin := &Plugin{Name: "test-plugin"}
	err := registry.EmitHook(context.Background(), "onLoad", plugin)
	if err != expectedErr {
		t.Errorf("EmitHook() error = %v, want %v", err, expectedErr)
	}
}

func TestPluginRegistry_RegisterCommand(t *testing.T) {
	registry := NewPluginRegistry()

	handler := func(ctx context.Context, args []string) (string, error) {
		return "result", nil
	}

	registry.RegisterCommand("test-cmd", handler)

	if len(registry.commandHandlers) != 1 {
		t.Error("RegisterCommand() did not register handler")
	}
}

func TestPluginRegistry_HandleCommand(t *testing.T) {
	registry := NewPluginRegistry()

	expectedResult := "command result"
	handler := func(ctx context.Context, args []string) (string, error) {
		return expectedResult, nil
	}

	registry.RegisterCommand("test-cmd", handler)

	result, err := registry.HandleCommand(context.Background(), "test-cmd", nil)
	if err != nil {
		t.Fatalf("HandleCommand() error = %v", err)
	}

	if result != expectedResult {
		t.Errorf("HandleCommand() = %v, want %v", result, expectedResult)
	}
}

func TestPluginRegistry_HandleCommand_NotFound(t *testing.T) {
	registry := NewPluginRegistry()

	_, err := registry.HandleCommand(context.Background(), "non-existent", nil)
	if err == nil {
		t.Error("HandleCommand() should return error for non-existent command")
	}
}

func TestPluginRegistry_GetCommandNames(t *testing.T) {
	registry := NewPluginRegistry()

	registry.RegisterCommand("cmd1", func(ctx context.Context, args []string) (string, error) { return "", nil })
	registry.RegisterCommand("cmd2", func(ctx context.Context, args []string) (string, error) { return "", nil })

	names := registry.GetCommandNames()

	if len(names) != 2 {
		t.Errorf("GetCommandNames() returned %d names, want 2", len(names))
	}
}

func TestNewPluginAPI(t *testing.T) {
	manager := NewPluginManager()
	registry := NewPluginRegistry()

	api := NewPluginAPI(manager, registry)

	if api == nil {
		t.Fatal("NewPluginAPI() returned nil")
	}

	if api.manager != manager {
		t.Error("NewPluginAPI() did not set manager correctly")
	}

	if api.registry != registry {
		t.Error("NewPluginAPI() did not set registry correctly")
	}
}

func TestPluginAPI_LoadBuiltInPlugins(t *testing.T) {
	manager := NewPluginManager()
	registry := NewPluginRegistry()
	api := NewPluginAPI(manager, registry)

	err := api.LoadBuiltInPlugins()
	if err != nil {
		t.Fatalf("LoadBuiltInPlugins() error = %v", err)
	}

	plugins := registry.List()
	if len(plugins) == 0 {
		t.Error("LoadBuiltInPlugins() did not load any plugins")
	}
}

func TestPlugin_Structure(t *testing.T) {
	now := time.Now()
	plugin := Plugin{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Author:      "Test Author",
		Repository:  "https://github.com/test/plugin",
		Enabled:     true,
		LoadedAt:    now,
		Commands: []PluginCommand{
			{Name: "cmd1", Description: "Command 1", Type: "prompt"},
		},
		Tools: []PluginTool{
			{Name: "tool1", Description: "Tool 1"},
		},
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	if plugin.Name == "" {
		t.Error("Plugin.Name is empty")
	}

	if plugin.LoadedAt != now {
		t.Error("Plugin.LoadedAt not set correctly")
	}

	if len(plugin.Commands) != 1 {
		t.Error("Plugin.Commands not set correctly")
	}

	if len(plugin.Tools) != 1 {
		t.Error("Plugin.Tools not set correctly")
	}
}

func TestPluginCommand_Structure(t *testing.T) {
	cmd := PluginCommand{
		Name:        "test-cmd",
		Description: "A test command",
		Type:        "prompt",
	}

	if cmd.Name == "" {
		t.Error("PluginCommand.Name is empty")
	}

	if cmd.Type == "" {
		t.Error("PluginCommand.Type is empty")
	}
}

func TestPluginTool_Structure(t *testing.T) {
	tool := PluginTool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
	}

	if tool.Name == "" {
		t.Error("PluginTool.Name is empty")
	}

	if tool.InputSchema == nil {
		t.Error("PluginTool.InputSchema is nil")
	}
}

func TestHookFunc_Structure(t *testing.T) {
	// Just verify the type exists and can be used
	var hook HookFunc = func(ctx context.Context, plugin *Plugin) error {
		return nil
	}

	if hook == nil {
		t.Error("HookFunc is nil")
	}
}

func TestCommandHandler_Structure(t *testing.T) {
	// Just verify the type exists and can be used
	var handler CommandHandler = func(ctx context.Context, args []string) (string, error) {
		return "", nil
	}

	if handler == nil {
		t.Error("CommandHandler is nil")
	}
}

func TestPluginManager_ConcurrentAccess(t *testing.T) {
	manager := NewPluginManager()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			manager.LoadPlugin(string(rune('0'+id)), &Plugin{Name: string(rune('0' + id))})
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	plugins := manager.ListPlugins()
	if len(plugins) != 10 {
		t.Errorf("ListPlugins() after concurrent loads returned %d, want 10", len(plugins))
	}
}
