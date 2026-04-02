package extractMemories

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MemoryFileName   = "memory"
	AutoMemDirName   = ".auto-memory"
	MemoryExtension  = ".md"
	MaxMemoryAgeDays = 14
)

type MemoryEntry struct {
	Content   string
	Timestamp time.Time
	Source    string
}

type MemoryStats struct {
	TotalEntries    int
	TotalSize       int64
	OldestEntry     time.Time
	NewestEntry     time.Time
	EntriesBySource map[string]int
}

func GetAutoMemPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".claude", AutoMemDirName), nil
}

func EnsureMemoryDir() (string, error) {
	memPath, err := GetAutoMemPath()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(memPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create memory directory: %w", err)
	}
	return memPath, nil
}

func ListMemoryEntries() ([]MemoryEntry, error) {
	memPath, err := GetAutoMemPath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(memPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []MemoryEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read memory directory: %w", err)
	}

	var memories []MemoryEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), MemoryExtension) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(memPath, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		source := strings.TrimSuffix(entry.Name(), MemoryExtension)
		memories = append(memories, MemoryEntry{
			Content:   string(content),
			Timestamp: info.ModTime(),
			Source:    source,
		})
	}
	return memories, nil
}

func ReadMemoryFile(source string) (string, error) {
	memPath, err := GetAutoMemPath()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(memPath, source+MemoryExtension)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read memory file: %w", err)
	}
	return string(content), nil
}

func WriteMemoryFile(source, content string) error {
	memPath, err := EnsureMemoryDir()
	if err != nil {
		return err
	}

	filePath := filepath.Join(memPath, source+MemoryExtension)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}
	return nil
}

func DeleteMemoryFile(source string) error {
	memPath, err := GetAutoMemPath()
	if err != nil {
		return err
	}

	filePath := filepath.Join(memPath, source+MemoryExtension)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete memory file: %w", err)
	}
	return nil
}

func GetMemoryStats() (*MemoryStats, error) {
	memPath, err := GetAutoMemPath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(memPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &MemoryStats{
				EntriesBySource: make(map[string]int),
			}, nil
		}
		return nil, fmt.Errorf("failed to read memory directory: %w", err)
	}

	stats := &MemoryStats{
		EntriesBySource: make(map[string]int),
	}

	var oldest, newest time.Time
	hasTime := false

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), MemoryExtension) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		stats.TotalEntries++
		stats.TotalSize += info.Size()

		if !hasTime {
			oldest = info.ModTime()
			newest = info.ModTime()
			hasTime = true
		} else {
			if info.ModTime().Before(oldest) {
				oldest = info.ModTime()
			}
			if info.ModTime().After(newest) {
				newest = info.ModTime()
			}
		}

		source := strings.TrimSuffix(entry.Name(), MemoryExtension)
		stats.EntriesBySource[source]++
	}

	if hasTime {
		stats.OldestEntry = oldest
		stats.NewestEntry = newest
	}

	return stats, nil
}

func CleanOldMemories() (int, error) {
	memPath, err := GetAutoMemPath()
	if err != nil {
		return 0, err
	}

	entries, err := os.ReadDir(memPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read memory directory: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -MaxMemoryAgeDays)
	removed := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), MemoryExtension) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filePath := filepath.Join(memPath, entry.Name())
			if err := os.Remove(filePath); err == nil {
				removed++
			}
		}
	}

	return removed, nil
}

func ExtractFromSession(sessionContent string) (string, error) {
	var important []string
	lines := strings.Split(sessionContent, "\n")

	inCodeBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "##") ||
			strings.HasPrefix(line, "###") {
			important = append(important, line)
		} else if strings.Contains(line, "important:") ||
			strings.Contains(line, "remember:") ||
			strings.Contains(line, "note:") {
			important = append(important, line)
		}
	}

	return strings.Join(important, "\n"), nil
}

func IsMemoryEnabled() bool {
	return true
}
