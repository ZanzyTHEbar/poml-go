package runtime

import (
	"encoding/json"
	"fmt"
)

// EvaluateIf evaluates a boolean expression using expr and returns a bool.
func EvaluateIf(exprStr string, ctx map[string]interface{}) (bool, error) {
	v, err := EvaluateExpression(exprStr, ctx)
	if err != nil {
		return false, err
	}
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		if val == "true" {
			return true, nil
		}
		if val == "false" {
			return false, nil
		}
		// try to unmarshal JSON booleans
		var b bool
		if err := json.Unmarshal([]byte(val), &b); err == nil {
			return b, nil
		}
		return false, fmt.Errorf("expression did not evaluate to bool: %v", v)
	default:
		return false, fmt.Errorf("expression did not evaluate to bool: %v", v)
	}
}

// EvaluateFor evaluates an iterable expression and returns a slice of interface{}.
func EvaluateFor(exprStr string, ctx map[string]interface{}) ([]interface{}, error) {
	v, err := EvaluateExpression(exprStr, ctx)
	if err != nil {
		return nil, err
	}
	switch it := v.(type) {
	case []interface{}:
		return it, nil
	case string:
		// attempt to parse JSON array in string
		var arr []interface{}
		if err := json.Unmarshal([]byte(it), &arr); err == nil {
			return arr, nil
		}
		return nil, fmt.Errorf("expression did not evaluate to iterable: %T", v)
	default:
		return nil, fmt.Errorf("expression did not evaluate to iterable: %T", v)
	}
}
