package services

import (
	"testing"
)

func TestMockRateLimiter(t *testing.T) {
	// 测试初始状态
	if ShouldProcessMockLimits() {
		t.Error("Mock rate limiter should not be enabled initially")
	}

	// 测试设置场景
	SetMockRateLimitScenario(MockScenarioNormal)

	if !ShouldProcessMockLimits() {
		t.Error("Mock rate limiter should be enabled after setting scenario")
	}

	headers := GetMockHeaders()
	if headers == nil {
		t.Error("Expected mock headers after setting scenario")
	}

	if headers["anthropic-ratelimit-unified-status"] != "allowed" {
		t.Errorf("Expected status 'allowed', got %q", headers["anthropic-ratelimit-unified-status"])
	}

	// 测试清除
	SetMockRateLimitScenario(MockScenarioClear)

	if ShouldProcessMockLimits() {
		t.Error("Mock rate limiter should not be enabled after clear")
	}
}

func TestMockRateLimitScenarios(t *testing.T) {
	scenarios := []MockRateLimitScenario{
		MockScenarioNormal,
		MockScenarioSessionLimitReached,
		MockScenarioApproachingWeeklyLimit,
		MockScenarioWeeklyLimitReached,
		MockScenarioOverageActive,
		MockScenarioOverageWarning,
		MockScenarioOverageExhausted,
	}

	for _, scenario := range scenarios {
		t.Run(string(scenario), func(t *testing.T) {
			SetMockRateLimitScenario(MockScenarioClear)
			SetMockRateLimitScenario(scenario)

			if !ShouldProcessMockLimits() {
				t.Errorf("Scenario %s: should be enabled", scenario)
			}

			current := GetCurrentMockScenario()
			if current != scenario {
				t.Errorf("Scenario %s: expected current %s", scenario, current)
			}
		})
	}
}

func TestGetCurrentMockScenario(t *testing.T) {
	SetMockRateLimitScenario(MockScenarioClear)

	current := GetCurrentMockScenario()
	if current != "" {
		t.Errorf("Expected empty scenario after clear, got %s", current)
	}

	SetMockRateLimitScenario(MockScenarioNormal)
	current = GetCurrentMockScenario()
	if current != MockScenarioNormal {
		t.Errorf("Expected %s, got %s", MockScenarioNormal, current)
	}

	SetMockRateLimitScenario(MockScenarioSessionLimitReached)
	current = GetCurrentMockScenario()
	if current != MockScenarioSessionLimitReached {
		t.Errorf("Expected %s, got %s", MockScenarioSessionLimitReached, current)
	}
}

func TestClearMockHeaders(t *testing.T) {
	SetMockRateLimitScenario(MockScenarioOverageActive)

	if !ShouldProcessMockLimits() {
		t.Error("Should be enabled after setting scenario")
	}

	ClearMockHeaders()

	if ShouldProcessMockLimits() {
		t.Error("Should not be enabled after clear")
	}

	headers := GetMockHeaders()
	if headers != nil {
		t.Error("Headers should be nil after clear")
	}
}

func TestIsMockFastModeRateLimitScenario(t *testing.T) {
	SetMockRateLimitScenario(MockScenarioClear)

	if IsMockFastModeRateLimitScenario() {
		t.Error("Should not be fast mode scenario initially")
	}

	SetMockRateLimitScenario(MockScenarioFastModeLimit)

	if !IsMockFastModeRateLimitScenario() {
		t.Error("Should be fast mode scenario after setting")
	}

	SetMockRateLimitScenario(MockScenarioClear)

	if IsMockFastModeRateLimitScenario() {
		t.Error("Should not be fast mode scenario after clear")
	}
}
