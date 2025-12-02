// Full path: pkg/scenarios/scenarios_test.go

package scenarios

import (
	"testing"
	"time"
)

func TestGetInfo_ValidScenarios(t *testing.T) {
	for _, info := range Registry {
		result := GetInfo(info.ID)
		if result == nil {
			t.Errorf("expected info for scenario %d, got nil", info.ID)
		}
		if result != nil && result.ID != info.ID {
			t.Errorf("expected ID %d, got %d", info.ID, result.ID)
		}
	}
}

func TestGetInfo_InvalidScenario(t *testing.T) {
	invalidIDs := []int{0, 4, 5, 9, 99, 100, -1}

	for _, id := range invalidIDs {
		info := GetInfo(id)
		if info != nil {
			t.Errorf("expected nil for scenario %d, got %v", id, info)
		}
	}
}

func TestIsValid(t *testing.T) {
	if len(Registry) > 0 && !IsValid(Registry[0].ID) {
		t.Errorf("scenario %d should be valid", Registry[0].ID)
	}

	if IsValid(0) {
		t.Error("scenario 0 should be invalid")
	}
	if IsValid(999) {
		t.Error("scenario 999 should be invalid")
	}
}

func TestRequiresBurst(t *testing.T) {
	for _, info := range Registry {
		result := RequiresBurst(info.ID)
		if result != info.RequiresBurst {
			t.Errorf("scenario %d: expected RequiresBurst=%v, got %v",
				info.ID, info.RequiresBurst, result)
		}
	}

	if RequiresBurst(999) {
		t.Error("invalid scenario should return false")
	}
}

func TestScenarioCategories(t *testing.T) {
	for _, info := range Registry {
		result := GetInfo(info.ID)
		if result == nil {
			t.Errorf("scenario %d not found", info.ID)
			continue
		}
		if result.Category != info.Category {
			t.Errorf("scenario %d: category mismatch", info.ID)
		}
	}
}

func TestMultiSymbols(t *testing.T) {
	if len(MultiSymbols) == 0 {
		t.Error("MultiSymbols should not be empty")
	}

	for _, sym := range MultiSymbols {
		if len(sym) == 0 {
			t.Error("empty symbol in MultiSymbols")
		}
	}
}

func TestRegistryCompleteness(t *testing.T) {
	if len(Registry) == 0 {
		t.Fatal("registry is empty")
	}

	for _, info := range Registry {
		if info.ID == 0 {
			t.Error("scenario with ID 0 found")
		}
		if info.Name == "" {
			t.Errorf("scenario %d has empty name", info.ID)
		}
		if info.Description == "" {
			t.Errorf("scenario %d has empty description", info.ID)
		}
	}
}

func TestResultFinalize(t *testing.T) {
	r := &Result{
		OrdersSent:        1000,
		ResponsesReceived: 1000,
		StartTime:         time.Now(),
	}
	r.EndTime = r.StartTime.Add(100 * time.Millisecond)

	r.Finalize()

	if r.Duration != 100*time.Millisecond {
		t.Errorf("expected duration 100ms, got %v", r.Duration)
	}
	if r.OrdersPerSec == 0 {
		t.Error("orders per sec should be calculated")
	}
}

func TestResultFinalizeZeroDuration(t *testing.T) {
	r := &Result{
		OrdersSent: 100,
		StartTime:  time.Now(),
	}
	r.EndTime = r.StartTime

	r.Finalize()

	if r.Duration != 0 {
		t.Errorf("expected zero duration, got %v", r.Duration)
	}
}

func TestNewRunner(t *testing.T) {
	runner := NewRunner(nil, 123, true)

	if runner == nil {
		t.Fatal("expected runner, got nil")
	}
	if runner.userID != 123 {
		t.Errorf("expected userID 123, got %d", runner.userID)
	}
	if !runner.verbose {
		t.Error("expected verbose to be true")
	}
	if runner.nextOrderID != 1 {
		t.Errorf("expected nextOrderID 1, got %d", runner.nextOrderID)
	}
}

func TestRunInvalidScenario(t *testing.T) {
	runner := NewRunner(nil, 1, false)

	_, err := runner.Run(999, false)
	if err == nil {
		t.Error("expected error for invalid scenario")
	}
}
