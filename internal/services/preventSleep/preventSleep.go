package preventSleep

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	// CAFFEINATE_TIMEOUT_SECONDS caffeinate超时秒数,进程在此时间后自动退出
	CAFFEINATE_TIMEOUT_SECONDS = 300 // 5分钟
	// RESTART_INTERVAL_MS 重启间隔,在到期前重启caffeinate
	RESTART_INTERVAL_MS = 4 * 60 * 1000 // 4分钟
)

// PreventSleepService 防止macOS进入空闲睡眠的服务
// 使用内置的caffeinate命令创建电源断言防止空闲睡眠
// 这确保长时间运行的API请求和工具执行不会被中断
type PreventSleepService struct {
	mu                sync.RWMutex
	caffeinateProcess *os.Process
	restartInterval   *time.Timer
	refCount          int
	cleanupRegistered bool
	running           bool
	stopChan          chan struct{}
}

var (
	instance *PreventSleepService
	once     sync.Once
)

// GetInstance 获取单例实例
func GetInstance() *PreventSleepService {
	once.Do(func() {
		instance = &PreventSleepService{
			refCount: 0,
			stopChan: make(chan struct{}),
		}
	})
	return instance
}

// StartPreventSleep 增加引用计数并在需要时开始阻止睡眠
// 在开始应该保持Mac唤醒的工作时调用
func (s *PreventSleepService) StartPreventSleep() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refCount++

	if s.refCount == 1 {
		s.spawnCaffeinateLocked()
		s.startRestartIntervalLocked()
	}
}

// StopPreventSleep 减少引用计数并在没有更多待处理工作时允许睡眠
// 在工作完成时调用
func (s *PreventSleepService) StopPreventSleep() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.refCount > 0 {
		s.refCount--
	}

	if s.refCount == 0 {
		s.stopRestartIntervalLocked()
		s.killCaffeinateLocked()
	}
}

// ForceStopPreventSleep 强制停止阻止睡眠,无视引用计数
// 用于退出时的清理
func (s *PreventSleepService) ForceStopPreventSleep() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refCount = 0
	s.stopRestartIntervalLocked()
	s.killCaffeinateLocked()
}

// startRestartIntervalLocked 在内部启动重启间隔(需要持有锁)
func (s *PreventSleepService) startRestartIntervalLocked() {
	// 仅在macOS上运行
	if runtime.GOOS != "darwin" {
		return
	}

	// 已在运行
	if s.running {
		return
	}

	s.running = true
	s.stopChan = make(chan struct{})

	go func() {
		ticker := time.NewTicker(RESTART_INTERVAL_MS)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				if s.refCount > 0 {
					s.killCaffeinateLocked()
					s.spawnCaffeinateLocked()
				}
				s.mu.Unlock()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// stopRestartIntervalLocked 停止重启间隔(需要持有锁)
func (s *PreventSleepService) stopRestartIntervalLocked() {
	if s.stopChan != nil {
		close(s.stopChan)
		s.stopChan = make(chan struct{})
	}
	s.running = false
}

// spawnCaffeinateLocked 生成caffeinate进程(需要持有锁)
func (s *PreventSleepService) spawnCaffeinateLocked() {
	// 仅在macOS上运行
	if runtime.GOOS != "darwin" {
		return
	}

	// 已在运行
	if s.caffeinateProcess != nil {
		return
	}

	// 检查caffeinate是否可用
	caffeinatePath, err := exec.LookPath("caffeinate")
	if err != nil {
		// caffeinate不可用,静默失败
		return
	}

	// 注册清理
	if !s.cleanupRegistered {
		s.cleanupRegistered = true
		// 在Go中我们需要在进程退出时清理
	}

	// -i: 创建防止空闲睡眠的断言
	//     这是最不激进的选择 - 显示器仍然可以睡眠
	// -t: 超时秒数 - caffeinate在此时间后自动退出
	//     这提供了自我修复能力,如果Node被SIGKILL杀死
	cmd := exec.Command(caffeinatePath, "-i", "-t", strconv.Itoa(CAFFEINATE_TIMEOUT_SECONDS))
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	err = cmd.Start()
	if err != nil {
		s.caffeinateProcess = nil
		return
	}

	s.caffeinateProcess = cmd.Process

	// 不让caffeinate保持Go进程存活
	// cmd.Process.Release() - 实际上我们应该在后台运行

	// 在后台等待进程退出
	go func() {
		cmd.Wait()
		s.mu.Lock()
		if s.caffeinateProcess != nil && s.caffeinateProcess.Pid == cmd.Process.Pid {
			s.caffeinateProcess = nil
		}
		s.mu.Unlock()
	}()
}

// killCaffeinateLocked 杀死caffeinate进程(需要持有锁)
func (s *PreventSleepService) killCaffeinateLocked() {
	if s.caffeinateProcess != nil {
		proc := s.caffeinateProcess
		s.caffeinateProcess = nil

		// SIGKILL立即终止 - SIGTERM可能有延迟
		proc.Kill()
	}
}

// GetRefCount 获取当前引用计数
func (s *PreventSleepService) GetRefCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refCount
}

// IsRunning 检查caffeinate是否正在运行
func (s *PreventSleepService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.caffeinateProcess != nil
}
