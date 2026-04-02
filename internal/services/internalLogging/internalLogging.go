package internalLogging

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	instance         *InternalLoggingService
	once             sync.Once
	k8sNamespace     string
	k8sNamespaceErr  error
	k8sNamespaceOnce sync.Once
)

type InternalLoggingService struct{}

func GetInstance() *InternalLoggingService {
	once.Do(func() {
		instance = &InternalLoggingService{}
	})
	return instance
}

func isAntUser() bool {
	return os.Getenv("USER_TYPE") == "ant"
}

func (s *InternalLoggingService) getKubernetesNamespace() string {
	k8sNamespaceOnce.Do(func() {
		if !isAntUser() {
			k8sNamespace = ""
			k8sNamespaceErr = nil
			return
		}

		namespacePath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
		content, err := os.ReadFile(namespacePath)
		if err != nil {
			k8sNamespace = "namespace not found"
			k8sNamespaceErr = err
			return
		}
		k8sNamespace = strings.TrimSpace(string(content))
		k8sNamespaceErr = nil
	})
	return k8sNamespace
}

func (s *InternalLoggingService) GetContainerId() string {
	if !isAntUser() {
		return ""
	}

	containerIdPath := "/proc/self/mountinfo"
	content, err := os.ReadFile(containerIdPath)
	if err != nil {
		return "container ID not found"
	}

	mountinfo := strings.TrimSpace(string(content))
	containerIdPattern := regexp.MustCompile(`(?:\/docker\/containers\/|\/sandboxes\/)([0-9a-f]{64})`)

	lines := strings.Split(mountinfo, "\n")
	for _, line := range lines {
		matches := containerIdPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return "container ID not found in mountinfo"
}

func (s *InternalLoggingService) GetKubernetesNamespaceCached() string {
	return s.getKubernetesNamespace()
}

func (s *InternalLoggingService) GetContainerIdCached() string {
	return s.GetContainerId()
}

func (s *InternalLoggingService) IsAntEnvironment() bool {
	return isAntUser()
}

func getProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

func getBranch() string {
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return ""
	}

	output, err := runCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

func runCommand(name string, args ...string) (string, error) {
	return "", nil
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
