package autoDream

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	lockFilePath     = ""
	lockFilePathOnce sync.Once
)

const (
	DefaultMinHours    = 24
	DefaultMinSessions = 5
	LockFileName       = ".auto-dream.lock"
)

type Config struct {
	MinHours    int
	MinSessions int
}

var defaultConfig = Config{
	MinHours:    DefaultMinHours,
	MinSessions: DefaultMinSessions,
}

type AutoDreamStats struct {
	TotalSessions    int
	HoursSince       float64
	LastConsolidated time.Time
	ConsolidatedAt   time.Time
}

type ConsolidationLock struct {
	mu        sync.Mutex
	filePath  string
	isHolding bool
}

func NewConsolidationLock() *ConsolidationLock {
	cl := &ConsolidationLock{}
	lockFilePathOnce.Do(func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/tmp"
		}
		lockFilePath = filepath.Join(homeDir, ".claude", LockFileName)
		if err := os.MkdirAll(filepath.Dir(lockFilePath), 0755); err != nil {
			lockFilePath = filepath.Join("/tmp", LockFileName)
		}
		cl.filePath = lockFilePath
	})
	return cl
}

func (l *ConsolidationLock) TryAcquire() (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isHolding {
		return true, nil
	}

	info, err := os.Stat(l.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(l.filePath)
			if err != nil {
				return false, fmt.Errorf("failed to create lock file: %w", err)
			}
			file.Close()
			l.isHolding = true
			return true, nil
		}
		return false, fmt.Errorf("failed to stat lock file: %w", err)
	}

	lockAge := time.Since(info.ModTime())
	if lockAge > 24*time.Hour {
		if err := os.Remove(l.filePath); err != nil {
			return false, fmt.Errorf("failed to remove stale lock: %w", err)
		}
		file, err := os.Create(l.filePath)
		if err != nil {
			return false, fmt.Errorf("failed to create lock file after removing stale: %w", err)
		}
		file.Close()
		l.isHolding = true
		return true, nil
	}

	return false, nil
}

func (l *ConsolidationLock) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.isHolding {
		return nil
	}

	if err := os.Remove(l.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}
	l.isHolding = false
	return nil
}

func (l *ConsolidationLock) IsHolding() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.isHolding
}

func getLockFilePath() string {
	lockFilePathOnce.Do(func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/tmp"
		}
		lockFilePath = filepath.Join(homeDir, ".claude", LockFileName)
		if err := os.MkdirAll(filepath.Dir(lockFilePath), 0755); err != nil {
			lockFilePath = filepath.Join("/tmp", LockFileName)
		}
	})
	return lockFilePath
}

func ReadLastConsolidatedAt() (time.Time, error) {
	lockPath := getLockFilePath()
	info, err := os.Stat(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("failed to stat lock file: %w", err)
	}
	return info.ModTime(), nil
}

func ListSessionsTouchedSince(since time.Time) ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	sessionsDir := filepath.Join(homeDir, ".claude", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions dir: %w", err)
	}

	var sessions []string
	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(since) {
				sessions = append(sessions, entry.Name())
			}
		}
	}
	return sessions, nil
}

func GetStats() (*AutoDreamStats, error) {
	lastConsolidated, err := ReadLastConsolidatedAt()
	if err != nil {
		return nil, err
	}

	sessions, err := ListSessionsTouchedSince(lastConsolidated)
	if err != nil {
		return nil, err
	}

	hoursSince := 0.0
	if !lastConsolidated.IsZero() {
		hoursSince = time.Since(lastConsolidated).Hours()
	}

	return &AutoDreamStats{
		TotalSessions:    len(sessions),
		HoursSince:       hoursSince,
		LastConsolidated: lastConsolidated,
		ConsolidatedAt:   time.Now(),
	}, nil
}

func IsEnabled() bool {
	return true
}

func GetConfig() Config {
	return defaultConfig
}

func SetConfig(cfg Config) {
	if cfg.MinHours <= 0 {
		cfg.MinHours = DefaultMinHours
	}
	if cfg.MinSessions <= 0 {
		cfg.MinSessions = DefaultMinSessions
	}
	defaultConfig = cfg
}
