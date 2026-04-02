package diagnosticTracking

import (
	"encoding/json"
	"strings"
	"sync"
)

const (
	// MAX_DIAGNOSTICS_SUMMARY_CHARS 诊断摘要最大字符数
	MAX_DIAGNOSTICS_SUMMARY_CHARS = 4000
)

// Diagnostic 诊断信息结构
type Diagnostic struct {
	Message  string             // 诊断消息
	Severity DiagnosticSeverity // 严重程度: Error, Warning, Info, Hint
	Range    DiagnosticRange    // 位置范围
	Source   string             // 来源(可选)
	Code     string             // 错误代码(可选)
}

// DiagnosticSeverity 诊断严重程度
type DiagnosticSeverity string

const (
	SeverityError   DiagnosticSeverity = "Error"
	SeverityWarning DiagnosticSeverity = "Warning"
	SeverityInfo    DiagnosticSeverity = "Info"
	SeverityHint    DiagnosticSeverity = "Hint"
)

// DiagnosticRange 诊断位置范围
type DiagnosticRange struct {
	Start Position `json:"start"` // 起始位置
	End   Position `json:"end"`   // 结束位置
}

// Position 文本位置
type Position struct {
	Line      int `json:"line"`      // 行号(0-indexed)
	Character int `json:"character"` // 列号(0-indexed)
}

// DiagnosticFile 包含诊断信息的文件
type DiagnosticFile struct {
	URI         string       `json:"uri"`         // 文件URI
	Diagnostics []Diagnostic `json:"diagnostics"` // 诊断列表
}

// DiagnosticTrackingService IDE诊断追踪服务
// 在文件编辑前后追踪诊断变化,用于检测新引入的错误和警告
type DiagnosticTrackingService struct {
	mu      sync.RWMutex
	enabled bool

	// 基准诊断映射: normalizedPath -> diagnostics
	baseline map[string][]Diagnostic

	// _claude_fs_right文件诊断状态
	rightFileDiagnosticsState map[string][]Diagnostic

	// 文件最后处理时间戳
	lastProcessedTimestamps map[string]int64
}

var (
	diagnosticInstance *DiagnosticTrackingService
	diagnosticOnce     sync.Once
)

// GetInstance 获取单例实例
func GetInstance() *DiagnosticTrackingService {
	diagnosticOnce.Do(func() {
		diagnosticInstance = &DiagnosticTrackingService{
			enabled:                   true,
			baseline:                  make(map[string][]Diagnostic),
			rightFileDiagnosticsState: make(map[string][]Diagnostic),
			lastProcessedTimestamps:   make(map[string]int64),
		}
	})
	return diagnosticInstance
}

// Initialize 初始化服务
func (s *DiagnosticTrackingService) Initialize() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
}

// Shutdown 关闭服务
func (s *DiagnosticTrackingService) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
	s.baseline = make(map[string][]Diagnostic)
	s.rightFileDiagnosticsState = make(map[string][]Diagnostic)
	s.lastProcessedTimestamps = make(map[string]int64)
}

// Reset 重置追踪状态
func (s *DiagnosticTrackingService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baseline = make(map[string][]Diagnostic)
	s.rightFileDiagnosticsState = make(map[string][]Diagnostic)
	s.lastProcessedTimestamps = make(map[string]int64)
}

// normalizeFileUri 标准化文件URI
// 移除协议前缀(file://, _claude_fs_right:, _claude_fs_left:)
func (s *DiagnosticTrackingService) NormalizeFileUri(fileUri string) string {
	protocolPrefixes := []string{
		"file://",
		"_claude_fs_right:",
		"_claude_fs_left:",
	}

	normalized := fileUri
	for _, prefix := range protocolPrefixes {
		if len(fileUri) > len(prefix) && fileUri[:len(prefix)] == prefix {
			normalized = fileUri[len(prefix):]
			break
		}
	}

	return normalized
}

// BeforeFileEdited 在文件编辑前捕获基准诊断
func (s *DiagnosticTrackingService) BeforeFileEdited(filePath string, diagnostics []Diagnostic) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return
	}

	timestamp := int64(0)
	normalizedPath := s.NormalizeFileUri(filePath)

	// 如果有诊断结果,存储基准
	if len(diagnostics) > 0 {
		s.baseline[normalizedPath] = diagnostics
	} else {
		// 无诊断时存储空切片
		s.baseline[normalizedPath] = []Diagnostic{}
	}
	s.lastProcessedTimestamps[normalizedPath] = timestamp
}

// GetNewDiagnostics 获取新增的诊断(排除基准)
func (s *DiagnosticTrackingService) GetNewDiagnostics(
	allDiagnostics []DiagnosticFile,
	baselineDiagnostics map[string][]Diagnostic,
) []DiagnosticFile {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return nil
	}

	var newDiagnosticFiles []DiagnosticFile

	// 处理file://协议诊断
	for _, file := range allDiagnostics {
		if !strings.HasPrefix(file.URI, "file://") {
			continue
		}

		normalizedPath := s.NormalizeFileUri(file.URI)
		if _, hasBaseline := baselineDiagnostics[normalizedPath]; !hasBaseline {
			continue
		}

		baselineDiags := baselineDiagnostics[normalizedPath]
		newDiags := s.FilterNewDiagnostics(file.Diagnostics, baselineDiags)

		if len(newDiags) > 0 {
			newDiagnosticFiles = append(newDiagnosticFiles, DiagnosticFile{
				URI:         file.URI,
				Diagnostics: newDiags,
			})
		}
	}

	return newDiagnosticFiles
}

// filterNewDiagnostics 过滤出新增的诊断
func (s *DiagnosticTrackingService) FilterNewDiagnostics(
	allDiagnostics []Diagnostic,
	baselineDiagnostics []Diagnostic,
) []Diagnostic {
	var newDiags []Diagnostic

	for _, diag := range allDiagnostics {
		if !s.ContainsDiagnostic(baselineDiagnostics, diag) {
			newDiags = append(newDiags, diag)
		}
	}

	return newDiags
}

// containsDiagnostic 检查诊断是否在列表中
func (s *DiagnosticTrackingService) ContainsDiagnostic(diags []Diagnostic, diag Diagnostic) bool {
	for _, d := range diags {
		if s.AreDiagnosticsEqual(d, diag) {
			return true
		}
	}
	return false
}

// areDiagnosticsEqual 判断两个诊断是否相等
func (s *DiagnosticTrackingService) AreDiagnosticsEqual(a, b Diagnostic) bool {
	return a.Message == b.Message &&
		a.Severity == b.Severity &&
		a.Source == b.Source &&
		a.Code == b.Code &&
		a.Range.Start.Line == b.Range.Start.Line &&
		a.Range.Start.Character == b.Range.Start.Character &&
		a.Range.End.Line == b.Range.End.Line &&
		a.Range.End.Character == b.Range.End.Character
}

// ParseDiagnosticResult 解析诊断结果
func (s *DiagnosticTrackingService) ParseDiagnosticResult(result json.RawMessage) []DiagnosticFile {
	var diagnosticFiles []DiagnosticFile

	if err := json.Unmarshal(result, &diagnosticFiles); err != nil {
		return nil
	}

	return diagnosticFiles
}

// FormatDiagnosticsSummary 格式化诊断摘要
func (s *DiagnosticTrackingService) FormatDiagnosticsSummary(files []DiagnosticFile) string {
	if len(files) == 0 {
		return ""
	}

	var result string
	for i, file := range files {
		if i > 0 {
			result += "\n\n"
		}

		// 提取文件名
		uri := file.URI
		for j := len(uri) - 1; j >= 0; j-- {
			if uri[j] == '/' {
				uri = uri[j+1:]
				break
			}
		}

		result += uri + ":\n"
		for _, d := range file.Diagnostics {
			severitySymbol := s.GetSeveritySymbol(d.Severity)
			location := ""
			if d.Range.Start.Line > 0 || d.Range.Start.Character > 0 {
				location = " [Line " + intToString(d.Range.Start.Line+1) + ":" + intToString(d.Range.Start.Character+1) + "]"
			}
			codeStr := ""
			if d.Code != "" {
				codeStr = " [" + d.Code + "]"
			}
			sourceStr := ""
			if d.Source != "" {
				sourceStr = " (" + d.Source + ")"
			}
			result += "  " + severitySymbol + location + " " + d.Message + codeStr + sourceStr + "\n"
		}
	}

	// 截断超长结果
	if len(result) > MAX_DIAGNOSTICS_SUMMARY_CHARS {
		truncationMarker := "…[truncated]"
		result = result[:MAX_DIAGNOSTICS_SUMMARY_CHARS-len(truncationMarker)] + truncationMarker
	}

	return result
}

// GetSeveritySymbol 获取严重程度符号
func (s *DiagnosticTrackingService) GetSeveritySymbol(severity DiagnosticSeverity) string {
	switch severity {
	case SeverityError:
		return "✗"
	case SeverityWarning:
		return "⚠"
	case SeverityInfo:
		return "ℹ"
	case SeverityHint:
		return "•"
	default:
		return "•"
	}
}

// SetEnabled 设置是否启用
func (s *DiagnosticTrackingService) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}

// IsEnabled 检查是否启用
func (s *DiagnosticTrackingService) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// Helper function to convert int to string
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// String helper for hasPrefix-like functionality
func (s *DiagnosticTrackingService) startsWithPrefix(uri, prefix string) bool {
	return len(uri) >= len(prefix) && uri[:len(prefix)] == prefix
}
