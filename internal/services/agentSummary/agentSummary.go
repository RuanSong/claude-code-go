package agentSummary

import (
	"strings"
	"sync"
	"time"
)

const (
	SUMMARY_INTERVAL_MS = 30_000
)

type AgentProgress struct {
	AgentID   string
	Summary   string
	UpdatedAt time.Time
}

type AgentSummaryService struct {
	mu        sync.RWMutex
	progress  map[string]*AgentProgress
	stopChan  chan struct{}
	isRunning bool
	tasks     map[string]*agentTask
}

type agentTask struct {
	stopFunc func()
}

var (
	instance     *AgentSummaryService
	instanceOnce sync.Once
)

func GetAgentSummaryService() *AgentSummaryService {
	instanceOnce.Do(func() {
		instance = &AgentSummaryService{
			progress: make(map[string]*AgentProgress),
			stopChan: make(chan struct{}),
			tasks:    make(map[string]*agentTask),
		}
	})
	return instance
}

func (s *AgentSummaryService) StartAgentSummarization(
	taskId string,
	agentId string,
	getTranscriptFunc func(agentId string) ([]*Message, error),
	updateProgressFunc func(summary string),
) func() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[taskId]; exists {
		return func() {}
	}

	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(SUMMARY_INTERVAL_MS * time.Millisecond)
		defer ticker.Stop()

		var previousSummary string

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				messages, err := getTranscriptFunc(agentId)
				if err != nil || len(messages) < 3 {
					continue
				}

				summary := s.generateSummary(messages, previousSummary)
				previousSummary = summary

				s.mu.Lock()
				s.progress[agentId] = &AgentProgress{
					AgentID:   agentId,
					Summary:   summary,
					UpdatedAt: time.Now(),
				}
				s.mu.Unlock()

				updateProgressFunc(summary)
			}
		}
	}()

	stopFunc := func() {
		close(stopChan)
		wg.Wait()
	}

	s.tasks[taskId] = &agentTask{
		stopFunc: stopFunc,
	}

	return stopFunc
}

func (s *AgentSummaryService) StopAgentSummarization(taskId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[taskId]; exists {
		task.stopFunc()
		delete(s.tasks, taskId)
	}
}

func (s *AgentSummaryService) GetProgress(agentId string) (*AgentProgress, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	progress, exists := s.progress[agentId]
	return progress, exists
}

func (s *AgentSummaryService) GetAllProgress() map[string]*AgentProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*AgentProgress)
	for k, v := range s.progress {
		result[k] = v
	}
	return result
}

func (s *AgentSummaryService) generateSummary(messages []*Message, previousSummary string) string {
	if len(messages) == 0 {
		return "No activity yet"
	}

	var recentActions []string
	for i := len(messages) - 1; i >= 0 && len(recentActions) < 5; i-- {
		msg := messages[i]
		if msg.Role == "assistant" && len(msg.Content) > 0 {
			action := extractAction(msg.Content)
			if action != "" {
				recentActions = append(recentActions, action)
			}
		}
	}

	if len(recentActions) == 0 {
		if previousSummary != "" {
			return previousSummary
		}
		return "Processing..."
	}

	var summary strings.Builder
	summary.WriteString("Recent: ")
	for i, action := range recentActions {
		if i > 0 {
			summary.WriteString(" → ")
		}
		summary.WriteString(action)
	}

	return summary.String()
}

func extractAction(content string) string {
	content = strings.TrimSpace(content)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "```") {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if len(line) > 100 {
			line = line[:100] + "..."
		}

		return line
	}

	if len(content) > 80 {
		return content[:80] + "..."
	}
	return content
}

type Message struct {
	Role    string
	Content string
	Time    time.Time
}

func (s *AgentSummaryService) ClearProgress(agentId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.progress, agentId)
}

func (s *AgentSummaryService) ClearAllProgress() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.progress = make(map[string]*AgentProgress)
}

func (s *AgentSummaryService) StopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		task.stopFunc()
	}
	s.tasks = make(map[string]*agentTask)
}

func (s *AgentSummaryService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"active_tasks":   len(s.tasks),
		"tracked_agents": len(s.progress),
	}
}

func BuildSummaryPrompt(previousSummary string) string {
	var prompt strings.Builder
	prompt.WriteString("Describe your most recent action in 3-5 words using present tense (-ing).\n")
	prompt.WriteString("Name the file or function, not the branch.\n")
	prompt.WriteString("Do not use tools.\n\n")
	prompt.WriteString("Good: \"Reading runAgent.ts\"\n")
	prompt.WriteString("Good: \"Fixing null check in validate.ts\"\n")
	prompt.WriteString("Good: \"Running auth module tests\"\n")
	prompt.WriteString("Good: \"Adding retry logic to fetchUser\"\n\n")
	prompt.WriteString("Bad (past tense): \"Analyzed the branch diff\"\n")
	prompt.WriteString("Bad (too vague): \"Investigating the issue\"\n")
	prompt.WriteString("Bad (too long): \"Reviewing full branch diff and AgentTool.tsx integration\"\n")
	prompt.WriteString("Bad (branch name): \"Analyzed adam/background-summary branch diff\"\n\n")

	if previousSummary != "" {
		prompt.WriteString("Previous: \"")
		prompt.WriteString(previousSummary)
		prompt.WriteString("\" — say something NEW.\n")
	}

	return prompt.String()
}
