package benchmark

import (
	"testing"
)

func TestSimpleIfElse(t *testing.T) {
	// Test both branches
	if SimpleIfElse(15) != "Greater than 10" {
		t.Error("Expected 'Greater than 10' for value 15")
	}
	
	if SimpleIfElse(5) != "Less than or equal to 10" {
		t.Error("Expected 'Less than or equal to 10' for value 5")
	}
}

func TestOnlyIfBranch(t *testing.T) {
	// Test both paths
	if OnlyIfBranch(15) != "Greater than 10" {
		t.Error("Expected 'Greater than 10' for value 15")
	}
	
	if OnlyIfBranch(5) != "Default value" {
		t.Error("Expected 'Default value' for value 5")
	}
}

func TestNestedConditions(t *testing.T) {
	// Test only some paths
	if NestedConditions(5, 5) != "Both positive" {
		t.Error("Expected 'Both positive' for (5, 5)")
	}
	
	if NestedConditions(-5, 5) != "A non-positive, B positive" {
		t.Error("Expected 'A non-positive, B positive' for (-5, 5)")
	}
	
	// Intentionally skip testing "A positive, B non-positive" to show partial coverage
}

func TestComplexCondition(t *testing.T) {
	// Test some branches but not all
	if ComplexCondition(5, 5, 0) != "Condition met" {
		t.Error("Expected 'Condition met' for (5, 5, 0)")
	}
	
	if ComplexCondition(0, 0, 15) != "Condition met" {
		t.Error("Expected 'Condition met' for (0, 0, 15)")
	}
	
	if ComplexCondition(0, 0, 5) != "Condition not met" {
		t.Error("Expected 'Condition not met' for (0, 0, 5)")
	}
}

func TestEarlyReturn(t *testing.T) {
	// Test only some of the early return paths
	_, err := EarlyReturn(-5, false)
	if err == nil {
		t.Error("Expected error for negative value")
	}
	
	result, err := EarlyReturn(50, true)
	if err != nil || result != "Flag is true" {
		t.Error("Expected 'Flag is true' without error for (50, true)")
	}
	
	// Skip testing value > 100 to demonstrate partial coverage
	
	result, err = EarlyReturn(50, false)
	if err != nil || result != "Value: 50" {
		t.Error("Expected 'Value: 50' without error for (50, false)")
	}
}

func TestShortCircuitCondition(t *testing.T) {
	// Test both outcomes but not all paths
	if ShortCircuitCondition(5, 20) != "Condition met" {
		t.Error("Expected 'Condition met' for (5, 20)")
	}
	
	if ShortCircuitCondition(0, 0) != "Condition not met" {
		t.Error("Expected 'Condition not met' for (0, 0)")
	}
	
	// Note: We intentionally avoid testing cases where a > 0 but b/a <= 2
}

func TestSwitchCase(t *testing.T) {
	// Test only some of the cases
	if SwitchCase(-5) != "Negative" {
		t.Error("Expected 'Negative' for -5")
	}
	
	if SwitchCase(0) != "Zero" {
		t.Error("Expected 'Zero' for 0")
	}
	
	if SwitchCase(5) != "Small positive" {
		t.Error("Expected 'Small positive' for 5")
	}
	
	// Skip testing the large positive case
}

func TestMultipleAnds(t *testing.T) {
	// Test only some scenarios
	if !MultipleAnds(1, 1, 1, 1) {
		t.Error("Expected true for all positive values")
	}
	
	if MultipleAnds(0, 1, 1, 1) {
		t.Error("Expected false when one value is non-positive")
	}
	
	// Skip testing other combinations
}

func TestMultipleOrs(t *testing.T) {
	// Test true for different OR parts
	if !MultipleOrs(20, 0, 0, 0) {
		t.Error("Expected true when a > 10")
	}
	
	if !MultipleOrs(0, 20, 0, 0) {
		t.Error("Expected true when b > 10")
	}
	
	// Test false case
	if MultipleOrs(5, 5, 5, 5) {
		t.Error("Expected false when all values <= 10")
	}
	
	// Skip testing other combinations
}

func TestNeverExecuted(t *testing.T) {
	// Never test value > 1000
	if NeverExecuted(-5) != "Negative" {
		t.Error("Expected 'Negative' for -5")
	}
	
	if NeverExecuted(50) != "Normal" {
		t.Error("Expected 'Normal' for 50")
	}
}

func TestConditionalLoop(t *testing.T) {
	// Test mix of positive and non-positive values
	count := ConditionalLoop([]int{1, -2, 3, 0, 5})
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
	
	// Test all positive
	count = ConditionalLoop([]int{1, 2, 3})
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
	
	// Test all non-positive
	count = ConditionalLoop([]int{-1, -2, 0})
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestStringCheckConsistency(t *testing.T) {
	// Test only some paths
	if StringCheckConsistency("") {
		t.Error("Expected false for empty string")
	}
	
	if !StringCheckConsistency("hello") {
		t.Error("Expected true for non-empty string")
	}
	
	// Notice we're only testing the first condition,
	// the others will appear as unreachable/uncovered
}

func TestSideEffectInCondition(t *testing.T) {
	// Test both outcomes
	getter := func() int { return 20 }
	if SideEffectInCondition(getter) != "Greater than 10" {
		t.Error("Expected 'Greater than 10'")
	}
	
	getter = func() int { return 5 }
	if SideEffectInCondition(getter) != "Less than or equal to 10" {
		t.Error("Expected 'Less than or equal to 10'")
	}
}

func TestPanicRecovery(t *testing.T) {
	// Test normal case
	result, err := PanicRecovery(4)
	if err != nil || result != 25 {
		t.Errorf("Expected result 25, got %d with error: %v", result, err)
	}
	
	// Test panic case
	_, err = PanicRecovery(0)
	if err == nil {
		t.Error("Expected error from panic recovery, got nil")
	}
}

func TestFunctionWithDefer(t *testing.T) {
	// Test normal case
	result, err := FunctionWithDefer(50)
	if err != nil || result != "Processed value: 50" {
		t.Errorf("Expected successful processing, got: %s with error: %v", result, err)
	}
	
	// Test error case
	_, err = FunctionWithDefer(-10)
	if err == nil {
		t.Error("Expected error for negative value")
	}
	
	// Skip testing value > 100
} 