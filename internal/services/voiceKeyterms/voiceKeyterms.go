package voiceKeyterms

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const MAX_KEYTERMS = 50

var GLOBAL_KEYTERMS = []string{
	"MCP",
	"symlink",
	"grep",
	"regex",
	"localhost",
	"codebase",
	"TypeScript",
	"JSON",
	"OAuth",
	"webhook",
	"gRPC",
	"dotfiles",
	"subagent",
	"worktree",
}

type VoiceKeytermsService struct{}

var (
	instance *VoiceKeytermsService
	once     sync.Once
)

func GetInstance() *VoiceKeytermsService {
	once.Do(func() {
		instance = &VoiceKeytermsService{}
	})
	return instance
}

func (s *VoiceKeytermsService) splitIdentifier(name string) []string {
	re1 := regexp.MustCompile(`([a-z])([A-Z])`)
	name = re1.ReplaceAllString(name, "$1 $2")

	re2 := regexp.MustCompile(`[-_./\s]+`)
	parts := re2.Split(name, -1)

	result := make([]string, 0)
	for _, w := range parts {
		w = strings.TrimSpace(w)
		if len(w) > 2 && len(w) <= 20 {
			result = append(result, w)
		}
	}
	return result
}

func (s *VoiceKeytermsService) fileNameWords(filePath string) []string {
	stem := filepath.Base(filePath)
	ext := filepath.Ext(stem)
	if ext != "" {
		stem = stem[:len(stem)-len(ext)]
	}
	return s.splitIdentifier(stem)
}

func (s *VoiceKeytermsService) getProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

func (s *VoiceKeytermsService) getBranch() string {
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return ""
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (s *VoiceKeytermsService) GetVoiceKeyterms(recentFiles []string) []string {
	terms := make(map[string]bool)

	for _, term := range GLOBAL_KEYTERMS {
		terms[term] = true
	}

	if projectRoot := s.getProjectRoot(); projectRoot != "" {
		name := filepath.Base(projectRoot)
		if len(name) > 2 && len(name) <= 50 {
			terms[name] = true
		}
	}

	if branch := s.getBranch(); branch != "" {
		for _, word := range s.splitIdentifier(branch) {
			terms[word] = true
		}
	}

	if recentFiles != nil {
		for _, filePath := range recentFiles {
			if len(terms) >= MAX_KEYTERMS {
				break
			}
			for _, word := range s.fileNameWords(filePath) {
				terms[word] = true
			}
		}
	}

	result := make([]string, 0, len(terms))
	for term := range terms {
		result = append(result, term)
	}

	if len(result) > MAX_KEYTERMS {
		result = result[:MAX_KEYTERMS]
	}

	return result
}

func (s *VoiceKeytermsService) GetGlobalKeyterms() []string {
	return GLOBAL_KEYTERMS
}
