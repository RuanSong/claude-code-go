package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// VoiceService 语音服务
// 对应 TypeScript: src/services/voice.ts
// 提供语音输入输出功能
type VoiceService struct {
	mu         sync.RWMutex
	enabled    bool         // 是否启用
	sampleRate int          // 采样率 (默认16000)
	channels   int          // 声道数 (默认1)
	backend    VoiceBackend // 语音后端
	listening  bool         // 是否正在监听
	muted      bool         // 是否静音
}

// VoiceBackend 语音后端接口
// 对应 TypeScript: 录音后端实现
// 支持不同平台的语音输入输出
type VoiceBackend interface {
	Start() error                      // 启动后端
	Stop() error                       // 停止后端
	Listen() (<-chan AudioData, error) // 开始监听音频
	Speak(text string) error           // 文本转语音
}

// AudioData 音频数据
// 对应 TypeScript: 原始音频数据格式
type AudioData struct {
	SampleRate int           `json:"sampleRate"` // 采样率
	Channels   int           `json:"channels"`   // 声道数
	Data       []byte        `json:"data"`       // 音频数据
	Duration   time.Duration `json:"duration"`   // 音频时长
}

// Transcription 转录结果
// 对应 TypeScript: 语音转文字结果
type Transcription struct {
	Text       string  `json:"text"`       // 转录文本
	Confidence float64 `json:"confidence"` // 置信度
	Language   string  `json:"language"`   // 语言
	StartTime  float64 `json:"startTime"`  // 开始时间
	EndTime    float64 `json:"endTime"`    // 结束时间
}

// VoiceConfig 语音配置
// 对应 TypeScript: 语音配置选项
type VoiceConfig struct {
	Enabled    bool   `json:"enabled"`    // 是否启用
	SampleRate int    `json:"sampleRate"` // 采样率
	Channels   int    `json:"channels"`   // 声道数
	Backend    string `json:"backend"`    // 后端类型
	Language   string `json:"language"`   // 语言
	Model      string `json:"model"`      // 模型
}

// NewVoiceService 创建新的语音服务
// 对应 TypeScript: 初始化语音服务
func NewVoiceService() *VoiceService {
	return &VoiceService{
		enabled:    false,
		sampleRate: 16000, // 默认16kHz采样率
		channels:   1,     // 默认单声道
		listening:  false,
		muted:      false,
	}
}

// Enable 启用语音服务
// 对应 TypeScript: 启用语音
func (v *VoiceService) Enable() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.enabled {
		return nil
	}

	v.enabled = true
	return nil
}

// Disable 禁用语音服务
// 对应 TypeScript: 禁用语音
func (v *VoiceService) Disable() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.enabled {
		return nil
	}

	if v.listening {
		if err := v.StopListening(); err != nil {
			return err
		}
	}

	v.enabled = false
	return nil
}

// IsEnabled 检查语音服务是否启用
func (v *VoiceService) IsEnabled() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.enabled
}

// IsListening 检查是否正在监听
func (v *VoiceService) IsListening() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.listening
}

// IsMuted 检查是否静音
func (v *VoiceService) IsMuted() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.muted
}

// SetMuted 设置静音状态
func (v *VoiceService) SetMuted(muted bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.muted = muted
}

// StartListening 开始监听音频输入
// 对应 TypeScript: startRecording()
// 需要先启用服务并配置后端
func (v *VoiceService) StartListening() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.enabled {
		return fmt.Errorf("voice service not enabled")
	}

	if v.listening {
		return fmt.Errorf("already listening")
	}

	if v.backend == nil {
		return fmt.Errorf("no voice backend configured")
	}

	if err := v.backend.Start(); err != nil {
		return fmt.Errorf("start backend: %w", err)
	}

	v.listening = true
	return nil
}

// StopListening 停止监听
// 对应 TypeScript: stopRecording()
func (v *VoiceService) StopListening() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.listening {
		return nil
	}

	if v.backend != nil {
		if err := v.backend.Stop(); err != nil {
			return fmt.Errorf("stop backend: %w", err)
		}
	}

	v.listening = false
	return nil
}

// Listen 开始语音识别循环
// 对应 TypeScript: 连接语音流
// 返回转录文本的通道
func (v *VoiceService) Listen(ctx context.Context) (<-chan Transcription, error) {
	if err := v.StartListening(); err != nil {
		return nil, err
	}

	transcriptions := make(chan Transcription)

	go func() {
		defer close(transcriptions)

		for {
			select {
			case <-ctx.Done():
				v.StopListening()
				return
			default:
				if v.IsMuted() {
					continue
				}
			}
		}
	}()

	return transcriptions, nil
}

// Speak 文本转语音
// 对应 TypeScript: 语音合成
func (v *VoiceService) Speak(text string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if !v.enabled {
		return fmt.Errorf("voice service not enabled")
	}

	if v.muted {
		return nil // 静音时不输出但也不报错
	}

	if v.backend == nil {
		return fmt.Errorf("no voice backend configured")
	}

	return v.backend.Speak(text)
}

// SetBackend 设置语音后端
// 对应 TypeScript: 设置录音后端
func (v *VoiceService) SetBackend(backend VoiceBackend) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.backend = backend
}

// GetConfig 获取语音配置
func (v *VoiceService) GetConfig() VoiceConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return VoiceConfig{
		Enabled:    v.enabled,
		SampleRate: v.sampleRate,
		Channels:   v.channels,
	}
}

// SetSampleRate 设置采样率
func (v *VoiceService) SetSampleRate(rate int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.sampleRate = rate
}

// SetChannels 设置声道数
func (v *VoiceService) SetChannels(channels int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.channels = channels
}

// MockVoiceBackend 模拟语音后端
// 用于测试环境
type MockVoiceBackend struct {
	mu        sync.RWMutex
	running   bool
	listeners []chan AudioData
}

// NewMockVoiceBackend 创建模拟后端
func NewMockVoiceBackend() *MockVoiceBackend {
	return &MockVoiceBackend{
		listeners: make([]chan AudioData, 0),
	}
}

func (b *MockVoiceBackend) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.running = true
	return nil
}

func (b *MockVoiceBackend) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.running = false

	for _, ch := range b.listeners {
		close(ch)
	}
	b.listeners = make([]chan AudioData, 0)
	return nil
}

func (b *MockVoiceBackend) Listen() (<-chan AudioData, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan AudioData, 100)
	b.listeners = append(b.listeners, ch)
	return ch, nil
}

func (b *MockVoiceBackend) Speak(text string) error {
	return nil
}

func (b *MockVoiceBackend) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// AudioDevice 音频设备
// 对应 TypeScript: 音频设备信息
type AudioDevice struct {
	Name     string `json:"name"`     // 设备名称
	ID       string `json:"id"`       // 设备ID
	IsInput  bool   `json:"isInput"`  // 是否为输入设备
	IsOutput bool   `json:"isOutput"` // 是否为输出设备
}

// DeviceManager 音频设备管理器
// 对应 TypeScript: 设备管理
type DeviceManager struct {
	mu         sync.RWMutex
	devices    []AudioDevice
	defaultIn  string // 默认输入设备
	defaultOut string // 默认输出设备
}

// NewDeviceManager 创建设备管理器
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices: make([]AudioDevice, 0),
	}
}

// ListDevices 列出所有音频设备
func (m *DeviceManager) ListDevices() []AudioDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]AudioDevice, len(m.devices))
	copy(devices, m.devices)
	return devices
}

// ListInputDevices 列出输入设备（麦克风）
func (m *DeviceManager) ListInputDevices() []AudioDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]AudioDevice, 0)
	for _, d := range m.devices {
		if d.IsInput {
			devices = append(devices, d)
		}
	}
	return devices
}

// ListOutputDevices 列出输出设备（扬声器）
func (m *DeviceManager) ListOutputDevices() []AudioDevice {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]AudioDevice, 0)
	for _, d := range m.devices {
		if d.IsOutput {
			devices = append(devices, d)
		}
	}
	return devices
}

// SetDefaultInput 设置默认输入设备
func (m *DeviceManager) SetDefaultInput(deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, d := range m.devices {
		if d.ID == deviceID && d.IsInput {
			m.defaultIn = deviceID
			return nil
		}
	}

	return fmt.Errorf("input device not found: %s", deviceID)
}

// SetDefaultOutput 设置默认输出设备
func (m *DeviceManager) SetDefaultOutput(deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, d := range m.devices {
		if d.ID == deviceID && d.IsOutput {
			m.defaultOut = deviceID
			return nil
		}
	}

	return fmt.Errorf("output device not found: %s", deviceID)
}

// GetDefaultInput 获取默认输入设备ID
func (m *DeviceManager) GetDefaultInput() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultIn
}

// GetDefaultOutput 获取默认输出设备ID
func (m *DeviceManager) GetDefaultOutput() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultOut
}

// AddDevice 添加音频设备
func (m *DeviceManager) AddDevice(device AudioDevice) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.devices = append(m.devices, device)
}

// VoiceStreamSTT 语音流识别服务
// 对应 TypeScript: src/services/voiceStreamSTT.ts
// 处理语音流并识别关键词
type VoiceStreamSTT struct {
	mu                sync.RWMutex
	enabled           bool
	keywords          []string // 监听关键词
	lastTranscription string   // 上次转录文本
}

// NewVoiceStreamSTT 创建语音流识别服务
func NewVoiceStreamSTT() *VoiceStreamSTT {
	return &VoiceStreamSTT{
		enabled:  false,
		keywords: make([]string, 0),
	}
}

func (s *VoiceStreamSTT) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
}

func (s *VoiceStreamSTT) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
}

func (s *VoiceStreamSTT) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// AddKeyword 添加监听关键词
// 对应 TypeScript: 添加关键词
func (s *VoiceStreamSTT) AddKeyword(keyword string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keywords = append(s.keywords, keyword)
}

// RemoveKeyword 移除关键词
func (s *VoiceStreamSTT) RemoveKeyword(keyword string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, k := range s.keywords {
		if k == keyword {
			s.keywords = append(s.keywords[:i], s.keywords[i+1:]...)
			return
		}
	}
}

// GetKeywords 获取所有关键词
func (s *VoiceStreamSTT) GetKeywords() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keywords := make([]string, len(s.keywords))
	copy(keywords, s.keywords)
	return keywords
}

// Process 处理转录文本
// 对应 TypeScript: 检查关键词匹配
// 检查文本是否包含任何关键词
func (s *VoiceStreamSTT) Process(transcription string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastTranscription = transcription

	if len(s.keywords) == 0 {
		return true // 无关键词时默认匹配
	}

	lowerTranscription := strings.ToLower(transcription)
	for _, keyword := range s.keywords {
		if strings.Contains(lowerTranscription, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// GetLastTranscription 获取上次转录文本
func (s *VoiceStreamSTT) GetLastTranscription() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastTranscription
}

// VoiceConfigStore 语音配置存储
// 对应 TypeScript: 配置持久化
type VoiceConfigStore struct {
	config VoiceConfig
	mu     sync.RWMutex
}

// NewVoiceConfigStore 创建配置存储
func NewVoiceConfigStore() *VoiceConfigStore {
	return &VoiceConfigStore{
		config: VoiceConfig{
			Enabled:    false,
			SampleRate: 16000,
			Channels:   1,
			Backend:    "auto",
			Language:   "en",
			Model:      "default",
		},
	}
}

// Get 获取当前配置
func (s *VoiceConfigStore) Get() VoiceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// Set 设置配置
func (s *VoiceConfigStore) Set(config VoiceConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// LoadFromFile 从文件加载配置
func (s *VoiceConfigStore) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var config VoiceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	s.Set(config)
	return nil
}

// SaveToFile 保存配置到文件
func (s *VoiceConfigStore) SaveToFile(path string) error {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
