package toolUseSummary

import (
	"sync"
	"time"
)

type ToolUseSummary struct {
	ToolName      string
	CallCount     int
	TotalDuration time.Duration
	LastUsed      time.Time
	SuccessRate   float64
}

type ToolUseSummaryService struct {
	mu        sync.RWMutex
	summaries map[string]*ToolUseSummary
}

var (
	instance     *ToolUseSummaryService
	instanceOnce sync.Once
)

func GetToolUseSummaryService() *ToolUseSummaryService {
	instanceOnce.Do(func() {
		instance = &ToolUseSummaryService{
			summaries: make(map[string]*ToolUseSummary),
		}
	})
	return instance
}

func (s *ToolUseSummaryService) RecordToolUse(toolName string, duration time.Duration, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	summary, exists := s.summaries[toolName]
	if !exists {
		summary = &ToolUseSummary{
			ToolName: toolName,
		}
		s.summaries[toolName] = summary
	}

	summary.CallCount++
	summary.TotalDuration += duration
	summary.LastUsed = time.Now()

	if success {
		currentSuccesses := summary.SuccessRate * float64(summary.CallCount-1)
		summary.SuccessRate = (currentSuccesses + 1) / float64(summary.CallCount)
	} else {
		currentSuccesses := summary.SuccessRate * float64(summary.CallCount-1)
		summary.SuccessRate = currentSuccesses / float64(summary.CallCount)
	}
}

func (s *ToolUseSummaryService) GetSummary(toolName string) (*ToolUseSummary, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary, exists := s.summaries[toolName]
	return summary, exists
}

func (s *ToolUseSummaryService) GetAllSummaries() []*ToolUseSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]*ToolUseSummary, 0, len(s.summaries))
	for _, summary := range s.summaries {
		summaries = append(summaries, summary)
	}
	return summaries
}

func (s *ToolUseSummaryService) GetMostUsedTools(limit int) []*ToolUseSummary {
	summaries := s.GetAllSummaries()

	for i := 0; i < len(summaries)-1; i++ {
		for j := i + 1; j < len(summaries); j++ {
			if summaries[j].CallCount > summaries[i].CallCount {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}

	if limit > 0 && limit < len(summaries) {
		summaries = summaries[:limit]
	}
	return summaries
}

func (s *ToolUseSummaryService) GetTotalCalls() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, summary := range s.summaries {
		total += summary.CallCount
	}
	return total
}

func (s *ToolUseSummaryService) GetAverageDuration(toolName string) time.Duration {
	summary, exists := s.GetSummary(toolName)
	if !exists || summary.CallCount == 0 {
		return 0
	}
	return summary.TotalDuration / time.Duration(summary.CallCount)
}

func (s *ToolUseSummaryService) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.summaries = make(map[string]*ToolUseSummary)
}

func (s *ToolUseSummaryService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total_tools": len(s.summaries),
		"total_calls": s.GetTotalCalls(),
	}

	var totalDuration time.Duration
	for _, summary := range s.summaries {
		totalDuration += summary.TotalDuration
	}
	stats["total_duration"] = totalDuration.String()

	return stats
}
