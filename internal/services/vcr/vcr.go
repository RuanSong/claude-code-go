package vcr

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	vcrInstance *VCRService
	vcrOnce     sync.Once
)

func GetInstance() *VCRService {
	vcrOnce.Do(func() {
		vcrInstance = &VCRService{
			isRecording: os.Getenv("VCR_RECORD") == "1",
		}
	})
	return vcrInstance
}

type VCRService struct {
	mu          sync.RWMutex
	isRecording bool
}

func (v *VCRService) ShouldUseVCR() bool {
	if os.Getenv("NODE_ENV") == "test" {
		return true
	}
	if os.Getenv("USER_TYPE") == "ant" && os.Getenv("FORCE_VCR") == "1" {
		return true
	}
	return false
}

func (v *VCRService) IsRecording() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.isRecording
}

func (v *VCRService) SetRecording(recording bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.isRecording = recording
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getcwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

func getConfigHome() string {
	if home := os.Getenv("CLAUDE_CONFIG_DIR"); home != "" {
		return home
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", "claude")
	}
	return ""
}

func isCI() bool {
	return os.Getenv("CI") == "true" || os.Getenv("isCI") == "true"
}

func isEnvTruthy(val string) bool {
	return val == "1" || val == "true" || val == "yes"
}

func jsonStringify(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

func jsonParse(data string) interface{} {
	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil
	}
	return result
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeFile(path string, data string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(data), 0644)
}

func (v *VCRService) WithFixture(
	input interface{},
	fixtureName string,
	f func() (interface{}, error),
) (interface{}, error) {
	if !v.ShouldUseVCR() {
		return f()
	}

	hash := sha1.New()
	hash.Write([]byte(jsonStringify(input)))
	hashStr := fmt.Sprintf("%x", hash.Sum(nil))[:12]

	fixturesRoot := getEnv("CLAUDE_CODE_TEST_FIXTURES_ROOT", getcwd())
	filename := filepath.Join(fixturesRoot, "fixtures", fmt.Sprintf("%s-%s.json", fixtureName, hashStr))

	if fileExists(filename) {
		data, err := readFile(filename)
		if err != nil {
			return nil, err
		}
		parsed := jsonParse(data)
		return parsed, nil
	}

	if isCI() && !v.IsRecording() {
		return nil, fmt.Errorf("fixture missing: %s. Re-run tests with VCR_RECORD=1, then commit the result", filename)
	}

	result, err := f()
	if err != nil {
		return nil, err
	}

	if err := writeFile(filename, jsonStringify(result)); err != nil {
		return nil, err
	}

	return result, nil
}

func (v *VCRService) WithVCR(
	messages []VCRMessage,
	f func() ([]VCRResult, error),
) ([]VCRResult, error) {
	if !v.ShouldUseVCR() {
		return f()
	}

	filtered := filterMetaMessages(messages)
	dehydratedInput := v.dehydrateMessages(filtered)

	hashes := make([]string, len(dehydratedInput))
	for i, msg := range dehydratedInput {
		h := sha1.New()
		h.Write([]byte(jsonStringify(msg)))
		hashes[i] = fmt.Sprintf("%x", h.Sum(nil))[:6]
	}

	fixturesRoot := getEnv("CLAUDE_CODE_TEST_FIXTURES_ROOT", getcwd())
	filename := filepath.Join(fixturesRoot, "fixtures", strings.Join(hashes, "-")+".json")

	if fileExists(filename) {
		data, err := readFile(filename)
		if err != nil {
			return nil, err
		}
		var cached struct {
			Output []VCRResult `json:"output"`
		}
		if err := json.Unmarshal([]byte(data), &cached); err != nil {
			return nil, err
		}
		for _, msg := range cached.Output {
			v.addCachedCost(msg)
		}
		result := make([]VCRResult, len(cached.Output))
		for i, msg := range cached.Output {
			result[i] = v.hydrateMessage(msg, i)
		}
		return result, nil
	}

	if isCI() && !v.IsRecording() {
		return nil, fmt.Errorf("Anthropic API fixture missing: %s. Re-run tests with VCR_RECORD=1, then commit the result", filename)
	}

	results, err := f()
	if err != nil {
		return nil, err
	}

	if isCI() && !v.IsRecording() {
		return results, nil
	}

	dehydratedResults := make([]VCRResult, len(results))
	for i, msg := range results {
		dehydratedResults[i] = v.dehydrateMessage(msg, i)
	}

	if err := writeFile(filename, jsonStringify(struct {
		Input  []interface{} `json:"input"`
		Output []VCRResult   `json:"output"`
	}{
		Input:  dehydratedInput,
		Output: dehydratedResults,
	})); err != nil {
		return nil, err
	}

	return results, nil
}

func (v *VCRService) WithStreamingVCR(
	messages []VCRMessage,
	f func() (<-chan VCRResult, error),
) (<-chan VCRResult, error) {
	if !v.ShouldUseVCR() {
		return f()
	}

	buffer := make([]VCRResult, 0)
	resultChan := make(chan VCRResult, 100)

	cached, err := v.WithVCR(messages, func() ([]VCRResult, error) {
		stream, err := f()
		if err != nil {
			return nil, err
		}
		for msg := range stream {
			buffer = append(buffer, msg)
		}
		return buffer, nil
	})

	if err != nil {
		return nil, err
	}

	go func() {
		defer close(resultChan)
		if len(cached) > 0 {
			for _, msg := range cached {
				resultChan <- msg
			}
			return
		}
		for _, msg := range buffer {
			resultChan <- msg
		}
	}()

	return resultChan, nil
}

func (v *VCRService) WithTokenCountVCR(
	messages interface{},
	tools interface{},
	f func() (int, error),
) (int, error) {
	cwdSlug := strings.ReplaceAll(getcwd(), "/", "-")
	cwdSlug = regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(cwdSlug, "-")

	dehydratedStr := jsonStringify(struct {
		Messages interface{} `json:"messages"`
		Tools    interface{} `json:"tools"`
	}{
		Messages: messages,
		Tools:    tools,
	})

	dehydrated := v.dehydrateValue(dehydratedStr)
	if dehydratedStr, ok := dehydrated.(string); ok {
		dehydrated = strings.ReplaceAll(dehydratedStr, cwdSlug, "[CWD_SLUG]")
		dehydrated = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).ReplaceAllString(dehydratedStr, "[UUID]")
		dehydrated = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z?`).ReplaceAllString(dehydratedStr, "[TIMESTAMP]")
	}

	result, err := v.WithFixture(dehydrated, "token-count", func() (interface{}, error) {
		count, err := f()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"tokenCount": count}, nil
	})

	if err != nil {
		return 0, err
	}

	if m, ok := result.(map[string]interface{}); ok {
		if count, ok := m["tokenCount"].(float64); ok {
			return int(count), nil
		}
	}

	return 0, nil
}

func filterMetaMessages(messages []VCRMessage) []VCRMessage {
	filtered := make([]VCRMessage, 0)
	for _, msg := range messages {
		if msg.Type != "user" {
			filtered = append(filtered, msg)
			continue
		}
		if !msg.IsMeta {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

func (v *VCRService) dehydrateMessages(messages []VCRMessage) []interface{} {
	result := make([]interface{}, len(messages))
	for i, msg := range messages {
		result[i] = v.dehydrateMessageContent(msg.Content)
	}
	return result
}

func (v *VCRService) dehydrateMessageContent(content interface{}) interface{} {
	switch c := content.(type) {
	case string:
		return v.dehydrateValue(c)
	case []interface{}:
		dehydrated := make([]interface{}, len(c))
		for i, item := range c {
			dehydrated[i] = v.dehydrateContentBlock(item)
		}
		return dehydrated
	}
	return content
}

func (v *VCRService) dehydrateContentBlock(block interface{}) interface{} {
	if m, ok := block.(map[string]interface{}); ok {
		blockType, _ := m["type"].(string)
		switch blockType {
		case "tool_result":
			content := m["content"]
			if str, ok := content.(string); ok {
				return map[string]interface{}{
					"type":    "tool_result",
					"id":      m["id"],
					"content": v.dehydrateValue(str),
				}
			}
			if arr, ok := content.([]interface{}); ok {
				dehydratedContent := make([]interface{}, len(arr))
				for i, item := range arr {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if textType, _ := itemMap["type"].(string); textType == "text" {
							if text, ok := itemMap["text"].(string); ok {
								dehydratedContent[i] = map[string]interface{}{
									"type": "text",
									"text": v.dehydrateValue(text),
								}
								continue
							}
						}
					}
					dehydratedContent[i] = item
				}
				return map[string]interface{}{
					"type":    "tool_result",
					"id":      m["id"],
					"content": dehydratedContent,
				}
			}
			return block
		case "text":
			if text, ok := m["text"].(string); ok {
				return map[string]interface{}{
					"type": "text",
					"text": v.dehydrateValue(text),
				}
			}
			return block
		case "tool_use":
			return map[string]interface{}{
				"type":  "tool_use",
				"id":    m["id"],
				"name":  m["name"],
				"input": v.dehydrateValue(m["input"]),
			}
		}
	}
	return block
}

func (v *VCRService) dehydrateMessage(msg VCRResult, index int) VCRResult {
	return msg
}

func (v *VCRService) hydrateMessage(msg VCRResult, index int) VCRResult {
	return msg
}

func (v *VCRService) dehydrateValue(s interface{}) interface{} {
	if str, ok := s.(string); ok {
		cwd := getcwd()
		configHome := getConfigHome()

		result := str
		result = regexp.MustCompile(`num_files="\d+"`).ReplaceAllString(result, `num_files="[NUM]"`)
		result = regexp.MustCompile(`duration_ms="\d+"`).ReplaceAllString(result, `duration_ms="[DURATION]"`)
		result = regexp.MustCompile(`cost_usd="\d+"`).ReplaceAllString(result, `cost_usd="[COST]"`)

		if configHome != "" {
			result = strings.ReplaceAll(result, configHome, "[CONFIG_HOME]")
		}
		result = strings.ReplaceAll(result, cwd, "[CWD]")

		if strings.Contains(result, "Available commands:") {
			result = regexp.MustCompile(`Available commands:.+`).ReplaceAllString(result, "Available commands: [COMMANDS]")
		}

		if strings.Contains(result, "Files modified by user:") {
			return "Files modified by user: [FILES]"
		}

		return result
	}
	return s
}

func (v *VCRService) hydrateValue(s interface{}) interface{} {
	if str, ok := s.(string); ok {
		result := str
		result = strings.ReplaceAll(result, "[NUM]", "1")
		result = strings.ReplaceAll(result, "[DURATION]", "100")
		result = strings.ReplaceAll(result, "[CONFIG_HOME]", getConfigHome())
		result = strings.ReplaceAll(result, "[CWD]", getcwd())
		return result
	}
	return s
}

func (v *VCRService) addCachedCost(msg VCRResult) {
}

type VCRMessage struct {
	Type    string
	Role    string
	Content interface{}
	IsMeta  bool
}

type VCRResult struct {
	Type      string
	UUID      string
	RequestID string
	Timestamp int64
	Message   VCRAssistantMessage
}

type VCRAssistantMessage struct {
	Role    string            `json:"role"`
	Content []VCRContentBlock `json:"content"`
}

type VCRContentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	Content   interface{}            `json:"content,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	IsError   bool                   `json:"is_error,omitempty"`
}
