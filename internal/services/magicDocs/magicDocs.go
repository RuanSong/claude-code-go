package magicDocs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	MagicDocHeaderPattern  = `^#\s*MAGIC\s+DOC:\s*(.+)$`
	MagicDocUpdateInterval = 5 * time.Minute
)

type MagicDocInfo struct {
	Path         string
	Title        string
	Instructions string
	LastUpdated  time.Time
	Content      string
}

type MagicDocsService struct {
	mu          sync.RWMutex
	trackedDocs map[string]*MagicDocInfo
	updateQueue chan string
	stopChan    chan struct{}
	isRunning   bool
}

var (
	magicDocRegex = regexp.MustCompile(`(?mi)^#\s*MAGIC\s+DOC:\s*(.+)$`)
	italicsRegex  = regexp.MustCompile(`(?m)^[_\*](.+?)[_\*]\s*$`)
	instance      *MagicDocsService
	instanceOnce  sync.Once
)

func GetMagicDocsService() *MagicDocsService {
	instanceOnce.Do(func() {
		instance = &MagicDocsService{
			trackedDocs: make(map[string]*MagicDocInfo),
			updateQueue: make(chan string, 100),
			stopChan:    make(chan struct{}),
		}
	})
	return instance
}

func DetectMagicDocHeader(content string) (string, string, bool) {
	match := magicDocRegex.FindStringSubmatch(content)
	if match == nil || len(match) < 2 {
		return "", "", false
	}

	title := strings.TrimSpace(match[1])

	titleEndIndex := magicDocRegex.FindStringIndex(content)
	if titleEndIndex == nil {
		return title, "", true
	}

	afterHeader := content[titleEndIndex[1]:]
	lines := strings.SplitN(afterHeader, "\n", 3)
	if len(lines) >= 2 {
		nextLine := strings.TrimSpace(lines[1])
		italicsMatch := italicsRegex.FindStringSubmatch(nextLine)
		if italicsMatch != nil && len(italicsMatch) >= 2 {
			instructions := strings.TrimSpace(italicsMatch[1])
			return title, instructions, true
		}
	}

	return title, "", true
}

func (s *MagicDocsService) RegisterMagicDoc(filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, exists := s.trackedDocs[absPath]; exists {
		return nil
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	title, instructions, ok := DetectMagicDocHeader(string(content))
	if !ok {
		return fmt.Errorf("file does not contain magic doc header")
	}

	s.trackedDocs[absPath] = &MagicDocInfo{
		Path:         absPath,
		Title:        title,
		Instructions: instructions,
		LastUpdated:  time.Now(),
		Content:      string(content),
	}

	return nil
}

func (s *MagicDocsService) UnregisterMagicDoc(filePath string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return
	}

	delete(s.trackedDocs, absPath)
}

func (s *MagicDocsService) GetMagicDoc(filePath string) (*MagicDocInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, false
	}

	info, exists := s.trackedDocs[absPath]
	return info, exists
}

func (s *MagicDocsService) ListMagicDocs() []*MagicDocInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docs := make([]*MagicDocInfo, 0, len(s.trackedDocs))
	for _, doc := range s.trackedDocs {
		docs = append(docs, doc)
	}
	return docs
}

func (s *MagicDocsService) Start() {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	go s.updateLoop()
}

func (s *MagicDocsService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}
	s.isRunning = false
	close(s.stopChan)
}

func (s *MagicDocsService) updateLoop() {
	ticker := time.NewTicker(MagicDocUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case filePath := <-s.updateQueue:
			s.processUpdate(filePath)
		case <-ticker.C:
			s.processAllUpdates()
		}
	}
}

func (s *MagicDocsService) QueueUpdate(filePath string) {
	select {
	case s.updateQueue <- filePath:
	default:
	}
}

func (s *MagicDocsService) processUpdate(filePath string) {
	s.mu.RLock()
	info, exists := s.trackedDocs[filePath]
	s.mu.RUnlock()

	if !exists {
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	s.mu.Lock()
	info.Content = string(content)
	info.LastUpdated = time.Now()
	s.mu.Unlock()
}

func (s *MagicDocsService) processAllUpdates() {
	s.mu.RLock()
	docs := make([]string, 0, len(s.trackedDocs))
	for path := range s.trackedDocs {
		docs = append(docs, path)
	}
	s.mu.RUnlock()

	for _, path := range docs {
		s.processUpdate(path)
	}
}

func (s *MagicDocsService) UpdateMagicDocContent(filePath, newContent string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	info, exists := s.trackedDocs[absPath]
	if !exists {
		return fmt.Errorf("magic doc not registered")
	}

	if err := os.WriteFile(absPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	info.Content = newContent
	info.LastUpdated = time.Now()

	return nil
}

func (s *MagicDocsService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total_tracked": len(s.trackedDocs),
		"is_running":    s.isRunning,
		"queue_size":    len(s.updateQueue),
		"docs":          make([]map[string]interface{}, 0),
	}

	for _, doc := range s.trackedDocs {
		docStats := map[string]interface{}{
			"path":         doc.Path,
			"title":        doc.Title,
			"last_updated": doc.LastUpdated.Format(time.RFC3339),
		}
		stats["docs"] = append(stats["docs"].([]map[string]interface{}), docStats)
	}

	return stats
}

func ClearTrackedMagicDocs() {
	svc := GetMagicDocsService()
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.trackedDocs = make(map[string]*MagicDocInfo)
}

func IsMagicDocFile(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	_, _, ok := DetectMagicDocHeader(string(content))
	return ok
}

func BuildMagicDocsUpdatePrompt(previousSummary string, instructions string) string {
	var prompt strings.Builder
	prompt.WriteString("# Magic Docs Update\n\n")
	prompt.WriteString("You are updating a MAGIC DOC based on recent conversation activity.\n\n")

	if instructions != "" {
		prompt.WriteString("## Instructions\n")
		prompt.WriteString(instructions + "\n\n")
	}

	prompt.WriteString("## Task\n")
	prompt.WriteString("Review the recent conversation and update the magic doc with new information, insights, or changes.\n")
	prompt.WriteString("Focus on:\n")
	prompt.WriteString("- New learnings or discoveries\n")
	prompt.WriteString("- Important decisions made\n")
	prompt.WriteString("- Changes to the project structure or design\n")
	prompt.WriteString("- Open questions or areas needing attention\n\n")

	if previousSummary != "" {
		prompt.WriteString("## Previous Summary\n")
		prompt.WriteString(previousSummary + "\n\n")
	}

	prompt.WriteString("## Guidelines\n")
	prompt.WriteString("- Keep the # MAGIC DOC header intact\n")
	prompt.WriteString("- Preserve any italics instructions on the line after the header\n")
	prompt.WriteString("- Update only the content below the header, keeping important historical information\n")
	prompt.WriteString("- Use clear, concise language\n")

	return prompt.String()
}
