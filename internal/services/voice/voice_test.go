package voice

import (
	"testing"
	"time"
)

func TestNewVoiceService(t *testing.T) {
	service := NewVoiceService()

	if service == nil {
		t.Fatal("NewVoiceService() returned nil")
	}

	if service.enabled {
		t.Error("NewVoiceService() should not enable service initially")
	}

	if service.listening {
		t.Error("NewVoiceService() should not set listening = true initially")
	}

	if service.muted {
		t.Error("NewVoiceService() should not set muted = true initially")
	}

	if service.sampleRate != 16000 {
		t.Errorf("NewVoiceService() default sampleRate = %d, want 16000", service.sampleRate)
	}

	if service.channels != 1 {
		t.Errorf("NewVoiceService() default channels = %d, want 1", service.channels)
	}
}

func TestVoiceService_Enable(t *testing.T) {
	service := NewVoiceService()

	err := service.Enable()
	if err != nil {
		t.Fatalf("Enable() error = %v", err)
	}

	if !service.enabled {
		t.Error("Enable() did not set enabled = true")
	}

	// Enabling again should not error
	err = service.Enable()
	if err != nil {
		t.Errorf("Enable() second call error = %v", err)
	}
}

func TestVoiceService_Disable(t *testing.T) {
	service := NewVoiceService()
	service.Enable()

	err := service.Disable()
	if err != nil {
		t.Fatalf("Disable() error = %v", err)
	}

	if service.enabled {
		t.Error("Disable() did not set enabled = false")
	}
}

func TestVoiceService_IsEnabled(t *testing.T) {
	service := NewVoiceService()

	if service.IsEnabled() {
		t.Error("IsEnabled() should return false initially")
	}

	service.enabled = true

	if !service.IsEnabled() {
		t.Error("IsEnabled() should return true when enabled")
	}
}

func TestVoiceService_IsListening(t *testing.T) {
	service := NewVoiceService()

	if service.IsListening() {
		t.Error("IsListening() should return false initially")
	}

	service.listening = true

	if !service.IsListening() {
		t.Error("IsListening() should return true when listening")
	}
}

func TestVoiceService_IsMuted(t *testing.T) {
	service := NewVoiceService()

	if service.IsMuted() {
		t.Error("IsMuted() should return false initially")
	}

	service.SetMuted(true)

	if !service.IsMuted() {
		t.Error("IsMuted() should return true when muted")
	}
}

func TestVoiceService_SetMuted(t *testing.T) {
	service := NewVoiceService()

	service.SetMuted(true)
	if !service.muted {
		t.Error("SetMuted(true) did not set muted")
	}

	service.SetMuted(false)
	if service.muted {
		t.Error("SetMuted(false) did not clear muted")
	}
}

func TestVoiceService_StartListening_NoBackend(t *testing.T) {
	service := NewVoiceService()
	service.enabled = true

	err := service.StartListening()
	if err == nil {
		t.Error("StartListening() should error when no backend configured")
	}
}

func TestVoiceService_StartListening_NotEnabled(t *testing.T) {
	service := NewVoiceService()

	err := service.StartListening()
	if err == nil {
		t.Error("StartListening() should error when not enabled")
	}
}

func TestVoiceService_StopListening(t *testing.T) {
	service := NewVoiceService()
	service.enabled = true
	service.listening = true

	err := service.StopListening()
	if err != nil {
		t.Errorf("StopListening() error = %v", err)
	}

	if service.listening {
		t.Error("StopListening() did not set listening = false")
	}
}

func TestVoiceService_Speak_NotEnabled(t *testing.T) {
	service := NewVoiceService()

	err := service.Speak("Hello")
	if err == nil {
		t.Error("Speak() should error when not enabled")
	}
}

func TestVoiceService_Speak_Muted(t *testing.T) {
	service := NewVoiceService()
	service.enabled = true
	service.muted = true

	err := service.Speak("Hello")
	if err != nil {
		t.Errorf("Speak() should not error when muted (just no-ops)")
	}
}

func TestVoiceService_Speak_NoBackend(t *testing.T) {
	service := NewVoiceService()
	service.enabled = true

	err := service.Speak("Hello")
	if err == nil {
		t.Error("Speak() should error when no backend configured")
	}
}

func TestVoiceService_SetBackend(t *testing.T) {
	service := NewVoiceService()
	backend := NewMockVoiceBackend()

	service.SetBackend(backend)

	if service.backend != backend {
		t.Error("SetBackend() did not set backend correctly")
	}
}

func TestVoiceService_GetConfig(t *testing.T) {
	service := NewVoiceService()
	service.enabled = true
	service.sampleRate = 44100
	service.channels = 2

	config := service.GetConfig()

	if !config.Enabled {
		t.Error("GetConfig() Enabled not set correctly")
	}

	if config.SampleRate != 44100 {
		t.Error("GetConfig() SampleRate not set correctly")
	}

	if config.Channels != 2 {
		t.Error("GetConfig() Channels not set correctly")
	}
}

func TestVoiceService_SetSampleRate(t *testing.T) {
	service := NewVoiceService()

	service.SetSampleRate(44100)

	if service.sampleRate != 44100 {
		t.Error("SetSampleRate() did not set sampleRate correctly")
	}
}

func TestVoiceService_SetChannels(t *testing.T) {
	service := NewVoiceService()

	service.SetChannels(2)

	if service.channels != 2 {
		t.Error("SetChannels() did not set channels correctly")
	}
}

func TestNewMockVoiceBackend(t *testing.T) {
	backend := NewMockVoiceBackend()

	if backend == nil {
		t.Fatal("NewMockVoiceBackend() returned nil")
	}

	if backend.running {
		t.Error("NewMockVoiceBackend() should not set running = true initially")
	}

	if backend.listeners == nil {
		t.Error("NewMockVoiceBackend() did not initialize listeners slice")
	}
}

func TestMockVoiceBackend_Start(t *testing.T) {
	backend := NewMockVoiceBackend()

	err := backend.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if !backend.running {
		t.Error("Start() did not set running = true")
	}
}

func TestMockVoiceBackend_Stop(t *testing.T) {
	backend := NewMockVoiceBackend()
	backend.Start()

	err := backend.Stop()
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	if backend.running {
		t.Error("Stop() did not set running = false")
	}
}

func TestMockVoiceBackend_IsRunning(t *testing.T) {
	backend := NewMockVoiceBackend()

	if backend.IsRunning() {
		t.Error("IsRunning() should return false initially")
	}

	backend.running = true

	if !backend.IsRunning() {
		t.Error("IsRunning() should return true when running")
	}
}

func TestMockVoiceBackend_Listen(t *testing.T) {
	backend := NewMockVoiceBackend()

	ch, err := backend.Listen()
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	if ch == nil {
		t.Error("Listen() returned nil channel")
	}
}

func TestMockVoiceBackend_Speak(t *testing.T) {
	backend := NewMockVoiceBackend()

	err := backend.Speak("Hello world")
	if err != nil {
		t.Errorf("Speak() error = %v", err)
	}
}

func TestNewDeviceManager(t *testing.T) {
	manager := NewDeviceManager()

	if manager == nil {
		t.Fatal("NewDeviceManager() returned nil")
	}

	if manager.devices == nil {
		t.Error("NewDeviceManager() did not initialize devices slice")
	}
}

func TestDeviceManager_ListDevices(t *testing.T) {
	manager := NewDeviceManager()

	manager.devices = []AudioDevice{
		{Name: "Device 1", ID: "dev1"},
		{Name: "Device 2", ID: "dev2"},
	}

	devices := manager.ListDevices()

	if len(devices) != 2 {
		t.Errorf("ListDevices() returned %d devices, want 2", len(devices))
	}
}

func TestDeviceManager_ListInputDevices(t *testing.T) {
	manager := NewDeviceManager()

	manager.devices = []AudioDevice{
		{Name: "Mic 1", ID: "mic1", IsInput: true},
		{Name: "Speaker 1", ID: "spk1", IsOutput: true},
		{Name: "Headset", ID: "headset", IsInput: true, IsOutput: true},
	}

	inputDevices := manager.ListInputDevices()

	if len(inputDevices) != 2 {
		t.Errorf("ListInputDevices() returned %d devices, want 2", len(inputDevices))
	}
}

func TestDeviceManager_ListOutputDevices(t *testing.T) {
	manager := NewDeviceManager()

	manager.devices = []AudioDevice{
		{Name: "Mic 1", ID: "mic1", IsInput: true},
		{Name: "Speaker 1", ID: "spk1", IsOutput: true},
		{Name: "Headset", ID: "headset", IsInput: true, IsOutput: true},
	}

	outputDevices := manager.ListOutputDevices()

	if len(outputDevices) != 2 {
		t.Errorf("ListOutputDevices() returned %d devices, want 2", len(outputDevices))
	}
}

func TestDeviceManager_SetDefaultInput(t *testing.T) {
	manager := NewDeviceManager()

	manager.devices = []AudioDevice{
		{Name: "Mic 1", ID: "mic1", IsInput: true},
	}

	err := manager.SetDefaultInput("mic1")
	if err != nil {
		t.Fatalf("SetDefaultInput() error = %v", err)
	}

	if manager.defaultIn != "mic1" {
		t.Error("SetDefaultInput() did not set defaultIn correctly")
	}
}

func TestDeviceManager_SetDefaultInput_NotFound(t *testing.T) {
	manager := NewDeviceManager()

	err := manager.SetDefaultInput("non-existent")
	if err == nil {
		t.Error("SetDefaultInput() should error for non-existent device")
	}
}

func TestDeviceManager_SetDefaultOutput(t *testing.T) {
	manager := NewDeviceManager()

	manager.devices = []AudioDevice{
		{Name: "Speaker 1", ID: "spk1", IsOutput: true},
	}

	err := manager.SetDefaultOutput("spk1")
	if err != nil {
		t.Fatalf("SetDefaultOutput() error = %v", err)
	}

	if manager.defaultOut != "spk1" {
		t.Error("SetDefaultOutput() did not set defaultOut correctly")
	}
}

func TestDeviceManager_GetDefaultInput(t *testing.T) {
	manager := NewDeviceManager()
	manager.defaultIn = "mic1"

	if manager.GetDefaultInput() != "mic1" {
		t.Error("GetDefaultInput() not returning correct value")
	}
}

func TestDeviceManager_GetDefaultOutput(t *testing.T) {
	manager := NewDeviceManager()
	manager.defaultOut = "spk1"

	if manager.GetDefaultOutput() != "spk1" {
		t.Error("GetDefaultOutput() not returning correct value")
	}
}

func TestDeviceManager_AddDevice(t *testing.T) {
	manager := NewDeviceManager()

	device := AudioDevice{Name: "New Device", ID: "new", IsInput: true}
	manager.AddDevice(device)

	if len(manager.devices) != 1 {
		t.Errorf("AddDevice() added %d devices, want 1", len(manager.devices))
	}
}

func TestNewVoiceStreamSTT(t *testing.T) {
	stt := NewVoiceStreamSTT()

	if stt == nil {
		t.Fatal("NewVoiceStreamSTT() returned nil")
	}

	if stt.enabled {
		t.Error("NewVoiceStreamSTT() should not set enabled = true initially")
	}

	if stt.keywords == nil {
		t.Error("NewVoiceStreamSTT() did not initialize keywords slice")
	}
}

func TestVoiceStreamSTT_Enable(t *testing.T) {
	stt := NewVoiceStreamSTT()

	stt.Enable()

	if !stt.enabled {
		t.Error("Enable() did not set enabled = true")
	}
}

func TestVoiceStreamSTT_Disable(t *testing.T) {
	stt := NewVoiceStreamSTT()
	stt.enabled = true

	stt.Disable()

	if stt.enabled {
		t.Error("Disable() did not set enabled = false")
	}
}

func TestVoiceStreamSTT_IsEnabled(t *testing.T) {
	stt := NewVoiceStreamSTT()

	if stt.IsEnabled() {
		t.Error("IsEnabled() should return false initially")
	}

	stt.enabled = true

	if !stt.IsEnabled() {
		t.Error("IsEnabled() should return true when enabled")
	}
}

func TestVoiceStreamSTT_AddKeyword(t *testing.T) {
	stt := NewVoiceStreamSTT()

	stt.AddKeyword("test")

	if len(stt.keywords) != 1 {
		t.Errorf("AddKeyword() added %d keywords, want 1", len(stt.keywords))
	}
}

func TestVoiceStreamSTT_RemoveKeyword(t *testing.T) {
	stt := NewVoiceStreamSTT()
	stt.keywords = []string{"test", "hello"}

	stt.RemoveKeyword("test")

	if len(stt.keywords) != 1 {
		t.Errorf("RemoveKeyword() left %d keywords, want 1", len(stt.keywords))
	}
}

func TestVoiceStreamSTT_GetKeywords(t *testing.T) {
	stt := NewVoiceStreamSTT()
	stt.keywords = []string{"test", "hello"}

	keywords := stt.GetKeywords()

	if len(keywords) != 2 {
		t.Errorf("GetKeywords() returned %d keywords, want 2", len(keywords))
	}
}

func TestVoiceStreamSTT_Process_NoKeywords(t *testing.T) {
	stt := NewVoiceStreamSTT()

	result := stt.Process("Hello world")

	if !result {
		t.Error("Process() should return true when no keywords configured")
	}
}

func TestVoiceStreamSTT_Process_WithMatchingKeyword(t *testing.T) {
	stt := NewVoiceStreamSTT()
	stt.keywords = []string{"hello"}

	result := stt.Process("Hello world")

	if !result {
		t.Error("Process() should return true when keyword matches")
	}
}

func TestVoiceStreamSTT_Process_CaseInsensitive(t *testing.T) {
	stt := NewVoiceStreamSTT()
	stt.keywords = []string{"hello"}

	result := stt.Process("HELLO world")

	if !result {
		t.Error("Process() should be case insensitive")
	}
}

func TestVoiceStreamSTT_Process_NoMatch(t *testing.T) {
	stt := NewVoiceStreamSTT()
	stt.keywords = []string{"goodbye"}

	result := stt.Process("Hello world")

	if result {
		t.Error("Process() should return false when no keyword matches")
	}
}

func TestVoiceStreamSTT_GetLastTranscription(t *testing.T) {
	stt := NewVoiceStreamSTT()
	stt.lastTranscription = "Hello world"

	if stt.GetLastTranscription() != "Hello world" {
		t.Error("GetLastTranscription() not returning correct value")
	}
}

func TestNewVoiceConfigStore(t *testing.T) {
	store := NewVoiceConfigStore()

	if store == nil {
		t.Fatal("NewVoiceConfigStore() returned nil")
	}

	if store.config.Enabled {
		t.Error("NewVoiceConfigStore() default config.Enabled should be false")
	}

	if store.config.SampleRate != 16000 {
		t.Error("NewVoiceConfigStore() default config.SampleRate should be 16000")
	}
}

func TestVoiceConfigStore_Get(t *testing.T) {
	store := NewVoiceConfigStore()
	store.config.Enabled = true

	config := store.Get()

	if !config.Enabled {
		t.Error("Get() not returning correct config")
	}
}

func TestVoiceConfigStore_Set(t *testing.T) {
	store := NewVoiceConfigStore()

	config := VoiceConfig{
		Enabled:    true,
		SampleRate: 44100,
		Channels:   2,
		Backend:    "auto",
		Language:   "en",
		Model:      "default",
	}

	store.Set(config)

	if !store.config.Enabled {
		t.Error("Set() did not update config.Enabled")
	}

	if store.config.SampleRate != 44100 {
		t.Error("Set() did not update config.SampleRate")
	}
}

func TestAudioData_Structure(t *testing.T) {
	data := AudioData{
		SampleRate: 16000,
		Channels:   1,
		Data:       []byte{0x00, 0x01, 0x02},
		Duration:   100 * time.Millisecond,
	}

	if data.SampleRate != 16000 {
		t.Error("AudioData.SampleRate not set correctly")
	}

	if len(data.Data) != 3 {
		t.Error("AudioData.Data not set correctly")
	}
}

func TestTranscription_Structure(t *testing.T) {
	transcription := Transcription{
		Text:       "Hello world",
		Confidence: 0.95,
		Language:   "en",
		StartTime:  0.0,
		EndTime:    1.5,
	}

	if transcription.Text == "" {
		t.Error("Transcription.Text is empty")
	}

	if transcription.Confidence != 0.95 {
		t.Error("Transcription.Confidence not set correctly")
	}
}

func TestVoiceConfig_Structure(t *testing.T) {
	config := VoiceConfig{
		Enabled:    true,
		SampleRate: 44100,
		Channels:   2,
		Backend:    "auto",
		Language:   "en",
		Model:      "default",
	}

	if !config.Enabled {
		t.Error("VoiceConfig.Enabled not set correctly")
	}

	if config.Backend != "auto" {
		t.Error("VoiceConfig.Backend not set correctly")
	}
}

func TestAudioDevice_Structure(t *testing.T) {
	device := AudioDevice{
		Name:     "Test Microphone",
		ID:       "test-mic",
		IsInput:  true,
		IsOutput: false,
	}

	if device.Name == "" {
		t.Error("AudioDevice.Name is empty")
	}

	if device.ID == "" {
		t.Error("AudioDevice.ID is empty")
	}

	if !device.IsInput {
		t.Error("AudioDevice.IsInput not set correctly")
	}
}

func TestVoiceService_Disable_WhileListening(t *testing.T) {
	service := NewVoiceService()
	service.enabled = true
	service.listening = true
	service.backend = NewMockVoiceBackend()

	// This should stop listening first
	err := service.Disable()
	if err != nil {
		t.Errorf("Disable() error = %v", err)
	}

	if service.enabled {
		t.Error("Disable() should disable the service")
	}
}
