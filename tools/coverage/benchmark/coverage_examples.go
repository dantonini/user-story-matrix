package benchmark

import (
	"errors"
	"fmt"
	"strings"
)

// SimpleIfElse demonstrates a basic if/else conditional
// Tests: simple if/else with both branches
func SimpleIfElse(value int) string {
	if value > 10 {
		return "Greater than 10"
	} else {
		return "Less than or equal to 10"
	}
}

// OnlyIfBranch demonstrates a function with only an if, no else
// Tests: if with no else
func OnlyIfBranch(value int) string {
	result := "Default value"
	if value > 10 {
		result = "Greater than 10"
	}
	return result
}

// NestedConditions demonstrates nested if statements
// Tests: nested conditions where some might not be taken
func NestedConditions(a, b int) string {
	if a > 0 {
		if b > 0 {
			return "Both positive"
		} else {
			return "A positive, B non-positive"
		}
	} else {
		if b > 0 {
			return "A non-positive, B positive"
		} else {
			return "Both non-positive"
		}
	}
}

// ComplexCondition demonstrates complex boolean logic
// Tests: complex condition with && and ||
func ComplexCondition(a, b, c int) string {
	if (a > 0 && b > 0) || c > 10 {
		return "Condition met"
	}
	return "Condition not met"
}

// EarlyReturn demonstrates multiple return points
// Tests: early returns and conditions that might not be evaluated
func EarlyReturn(value int, flag bool) (string, error) {
	if value < 0 {
		return "", errors.New("negative value not allowed")
	}
	
	if flag {
		return "Flag is true", nil
	}
	
	if value > 100 {
		return "Value too large", nil
	}
	
	return fmt.Sprintf("Value: %d", value), nil
}

// ShortCircuitCondition demonstrates short-circuit evaluation
// Tests: conditions where second part might not be evaluated due to short-circuiting
func ShortCircuitCondition(a, b int) string {
	// If a <= 0, second part won't be evaluated
	if a > 0 && b/a > 2 {
		return "Condition met"
	}
	return "Condition not met"
}

// SwitchCase demonstrates switch statements
// Tests: switch with multiple cases
func SwitchCase(value int) string {
	switch {
	case value < 0:
		return "Negative"
	case value == 0:
		return "Zero"
	case value < 10:
		return "Small positive"
	case value < 100:
		return "Medium positive"
	default:
		return "Large positive"
	}
}

// MultipleAnds demonstrates a condition with multiple AND parts
// Tests: multi-part conditions where all parts must be true
func MultipleAnds(a, b, c, d int) bool {
	if a > 0 && b > 0 && c > 0 && d > 0 {
		return true
	}
	return false
}

// MultipleOrs demonstrates a condition with multiple OR parts
// Tests: multi-part conditions where only one part needs to be true
func MultipleOrs(a, b, c, d int) bool {
	if a > 10 || b > 10 || c > 10 || d > 10 {
		return true
	}
	return false
}

// NeverExecuted contains code paths that will never be executed
// Tests: unreachable code
func NeverExecuted(value int) string {
	if value < 0 {
		return "Negative"
	}
	
	// This branch will never be executed in tests
	if value > 1000 {
		return "Very large"
	}
	
	return "Normal"
}

// ConditionalLoop demonstrates a conditional inside a loop
// Tests: conditions within loops
func ConditionalLoop(values []int) int {
	count := 0
	for _, v := range values {
		if v > 0 {
			count++
		}
	}
	return count
}

// StringCheckConsistency shows different ways to check strings
// Tests: equivalent string checks that might show different coverage patterns
func StringCheckConsistency(s string) bool {
	// These three checks are effectively the same
	if len(s) == 0 {
		return false
	}
	
	if s == "" {
		return false
	}
	
	if strings.TrimSpace(s) == "" {
		return false
	}
	
	return true
}

// SideEffectInCondition has a function call within the condition
// Tests: conditions with side effects
func SideEffectInCondition(getter func() int) string {
	if val := getter(); val > 10 {
		return "Greater than 10"
	}
	return "Less than or equal to 10"
}

// PanicRecovery contains code paths that might panic
// Tests: panic recovery paths
func PanicRecovery(divisor int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()
	
	if divisor == 0 {
		panic("division by zero")
	}
	
	return 100 / divisor, nil
}

// FunctionWithDefer demonstrates deferred execution
// Tests: deferred function execution
func FunctionWithDefer(value int) (string, error) {
	resource := "resource acquired"
	defer func() {
		// This always runs
		_ = resource // Use the variable to avoid unused variable warning
		resource = "resource released"
	}()
	
	if value < 0 {
		return "", errors.New("invalid value")
	}
	
	if value > 100 {
		return "Value too large", nil
	}
	
	return fmt.Sprintf("Processed value: %d", value), nil
} 