package runtime

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/aymerick/raymond"
	exprpkg "github.com/expr-lang/expr"
)

var registerOnce sync.Once

// TemplateCacheEntry represents a cached template evaluation result
type TemplateCacheEntry struct {
	Result    string
	Timestamp time.Time
	HitCount  int
}

// TemplateExpressionCache provides caching for template expression evaluations
type TemplateExpressionCache struct {
	cache map[string]*TemplateCacheEntry
	mutex sync.RWMutex
	ttl   time.Duration // Time to live for cache entries
}

// Global cache instance
var templateCache = &TemplateExpressionCache{
	cache: make(map[string]*TemplateCacheEntry),
	ttl:   5 * time.Minute, // 5 minute TTL by default
}

// Template function map for enhanced expression evaluation
var templateFunctions = template.FuncMap{
	// String manipulation functions
	"replace": func(old, new, s string) string {
		return strings.ReplaceAll(s, old, new)
	},
	"trim":      strings.TrimSpace,
	"trimSpace": strings.TrimSpace,
	"trimStart": func(s string) string {
		return strings.TrimLeft(s, " \t\n\r")
	},
	"trimEnd": func(s string) string {
		return strings.TrimRight(s, " \t\n\r")
	},
	"upper":     strings.ToUpper,
	"lower":     strings.ToLower,
	"title":     strings.Title,
	"contains":  strings.Contains,
	"hasPrefix": strings.HasPrefix,
	"hasSuffix": strings.HasSuffix,

	// Regex functions
	"replaceRegex": func(pattern, replacement, input string) string {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return input // Return original if regex is invalid
		}
		return re.ReplaceAllString(input, replacement)
	},
	"matchRegex": func(pattern, input string) bool {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(input)
	},
	"findRegex": func(pattern, input string) []string {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return []string{}
		}
		return re.FindAllString(input, -1)
	},

	// Array/Object functions
	"len": func(v interface{}) int {
		// Handle various types
		switch val := v.(type) {
		case string:
			return len(val)
		case []interface{}:
			return len(val)
		case map[string]interface{}:
			return len(val)
		default:
			return 0
		}
	},
	"first": func(v interface{}) interface{} {
		switch val := v.(type) {
		case []interface{}:
			if len(val) > 0 {
				return val[0]
			}
		}
		return nil
	},
	"last": func(v interface{}) interface{} {
		switch val := v.(type) {
		case []interface{}:
			if len(val) > 0 {
				return val[len(val)-1]
			}
		}
		return nil
	},
}

// Global template engine with functions
var templateEngine = template.New("poml").Funcs(templateFunctions)

// generateCacheKey creates a unique key for template + environment combination
func generateCacheKey(template string, env map[string]interface{}) string {
	// Create a deterministic string representation of the environment
	envStr := ""
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	// Sort keys for consistent ordering
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, k := range keys {
		envStr += k + "=" + fmt.Sprintf("%v", env[k]) + ";"
	}

	// Create MD5 hash of template + environment
	hash := md5.Sum([]byte(template + "|" + envStr))
	return fmt.Sprintf("%x", hash)
}

// Get retrieves a cached result if available and not expired
func (c *TemplateExpressionCache) Get(key string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return "", false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > c.ttl {
		return "", false
	}

	// Update hit count
	entry.HitCount++
	return entry.Result, true
}

// Put stores a result in the cache
func (c *TemplateExpressionCache) Put(key, result string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &TemplateCacheEntry{
		Result:    result,
		Timestamp: time.Now(),
		HitCount:  0,
	}
}

// Clear removes all expired entries from the cache
func (c *TemplateExpressionCache) ClearExpired() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expired := 0
	for key, entry := range c.cache {
		if time.Since(entry.Timestamp) > c.ttl {
			delete(c.cache, key)
			expired++
		}
	}
	return expired
}

// GetStats returns cache statistics
func (c *TemplateExpressionCache) GetStats() (totalEntries, totalHits int) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	totalEntries = len(c.cache)
	for _, entry := range c.cache {
		totalHits += entry.HitCount
	}
	return totalEntries, totalHits
}

// InvalidateByKey removes a specific cache entry by key
func (c *TemplateExpressionCache) InvalidateByKey(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.cache[key]; exists {
		delete(c.cache, key)
		return true
	}
	return false
}

// InvalidateByTemplate removes all cache entries for a specific template
func (c *TemplateExpressionCache) InvalidateByTemplate(template string) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	invalidated := 0
	for key := range c.cache {
		// Extract template from cache key (key format: hash(template|env))
		if strings.Contains(key, template) {
			delete(c.cache, key)
			invalidated++
		}
	}
	return invalidated
}

// InvalidateByVariable removes cache entries that depend on a specific variable
func (c *TemplateExpressionCache) InvalidateByVariable(varName string) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	invalidated := 0
	for key := range c.cache {
		// Check if the cache key contains the variable name
		// This is a simple heuristic - in production you might want more sophisticated dependency tracking
		if strings.Contains(key, varName+"=") {
			delete(c.cache, key)
			invalidated++
		}
	}
	return invalidated
}

// InvalidateAll clears the entire cache
func (c *TemplateExpressionCache) InvalidateAll() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	totalInvalidated := len(c.cache)
	c.cache = make(map[string]*TemplateCacheEntry)
	return totalInvalidated
}

// SetTTL updates the time-to-live for cache entries
func (c *TemplateExpressionCache) SetTTL(ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.ttl = ttl
}

// GetTTL returns the current time-to-live setting
func (c *TemplateExpressionCache) GetTTL() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.ttl
}

func registerHelpers() {
	// String helpers
	raymond.RegisterHelper("upper", func(options *raymond.Options) string {
		params := options.Params()
		if len(params) == 0 {
			return ""
		}
		if s, ok := params[0].(string); ok {
			return strings.ToUpper(s)
		}
		return ""
	})

	raymond.RegisterHelper("lower", func(options *raymond.Options) string {
		params := options.Params()
		if len(params) == 0 {
			return ""
		}
		if s, ok := params[0].(string); ok {
			return strings.ToLower(s)
		}
		return ""
	})

	raymond.RegisterHelper("trim", func(options *raymond.Options) string {
		params := options.Params()
		if len(params) == 0 {
			return ""
		}
		if s, ok := params[0].(string); ok {
			return strings.TrimSpace(s)
		}
		return ""
	})

	raymond.RegisterHelper("replace", func(options *raymond.Options) string {
		params := options.Params()
		if len(params) < 3 {
			return ""
		}
		if s, ok := params[0].(string); ok {
			if pattern, ok := params[1].(string); ok {
				if replacement, ok := params[2].(string); ok {
					re, err := regexp.Compile(pattern)
					if err != nil {
						return s
					}
					return re.ReplaceAllString(s, replacement)
				}
			}
		}
		return ""
	})

	// Array helpers
	raymond.RegisterHelper("length", func(options *raymond.Options) int {
		params := options.Params()
		if len(params) == 0 {
			return 0
		}
		switch v := params[0].(type) {
		case []interface{}:
			return len(v)
		case []string:
			return len(v)
		case string:
			return len(v)
		default:
			return 0
		}
	})

	raymond.RegisterHelper("join", func(options *raymond.Options) string {
		params := options.Params()
		if len(params) < 2 {
			return ""
		}
		sep, ok := params[1].(string)
		if !ok {
			sep = ","
		}
		switch v := params[0].(type) {
		case []interface{}:
			strs := make([]string, len(v))
			for i, item := range v {
				strs[i] = fmt.Sprintf("%v", item)
			}
			return strings.Join(strs, sep)
		case []string:
			return strings.Join(v, sep)
		default:
			return fmt.Sprintf("%v", v)
		}
	})

	// Math helpers
	raymond.RegisterHelper("add", func(options *raymond.Options) interface{} {
		params := options.Params()
		if len(params) < 2 {
			return 0
		}
		a, ok1 := params[0].(float64)
		b, ok2 := params[1].(float64)
		if ok1 && ok2 {
			return a + b
		}
		// Try int conversion
		if ai, ok := params[0].(int); ok {
			if bi, ok := params[1].(int); ok {
				return ai + bi
			}
		}
		return 0
	})
}

// RenderTemplate renders a Handlebars-like template using raymond with the provided context.
func RenderTemplate(tpl string, ctx map[string]interface{}) (string, error) {
	// register helpers once
	registerOnce.Do(registerHelpers)
	return raymond.Render(tpl, ctx)
}

// EvaluateExpression evaluates a simple expression using expr in a safe environment.
func EvaluateExpression(expression string, env map[string]interface{}) (interface{}, error) {
	program, err := exprpkg.Compile(expression, exprpkg.Env(env))
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression '%s': %w", expression, err)
	}
	result, err := exprpkg.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression '%s': %w", expression, err)
	}
	return result, nil
}

// SafeString converts a value to string safely, handling various types.
func SafeString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// SafeInt converts a value to int safely.
func SafeInt(value interface{}) int {
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		return 0
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// SafeFloat converts a value to float64 safely.
func SafeFloat(value interface{}) float64 {
	if value == nil {
		return 0.0
	}
	switch v := value.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return 0.0
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	default:
		return 0.0
	}
}

// SafeBool converts a value to bool safely.
func SafeBool(value interface{}) bool {
	if value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(strings.TrimSpace(v)) == "true"
	case int, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0.0
	default:
		return false
	}
}

// EvaluateTemplateExpression evaluates expressions within template strings like {{expr}}.
func EvaluateTemplateExpression(template string, env map[string]interface{}) (string, error) {
	// Check cache first for exact template + environment match
	cacheKey := generateCacheKey(template, env)
	if cachedResult, found := templateCache.Get(cacheKey); found {
		return cachedResult, nil
	}

	// Try Go template evaluation for expressions with template syntax
	if strings.Contains(template, "{{") && strings.Contains(template, "}}") {
		tmpl, err := templateEngine.Parse(template)
		if err == nil {
			var result strings.Builder
			err = tmpl.Execute(&result, env)
			if err == nil {
				output := result.String()
				templateCache.Put(cacheKey, output)
				return output, nil
			}
			// If template execution fails, log and fall back to existing implementation
			// This allows for graceful degradation when template functions aren't available
		}
		// If template parsing fails, fall back to existing implementation
	}

	// Fallback: Use existing expr-lang evaluation for backward compatibility
	// Simple regex to find {{expression}} patterns
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract the expression inside {{ }}
		expr := strings.TrimSpace(match[2 : len(match)-2])
		if expr == "" {
			return match
		}

		// Evaluate the expression with method chaining support
		value, err := EvaluateComplexExpression(expr, env)
		if err != nil {
			// Return the original match if evaluation fails
			return match
		}

		return SafeString(value)
	})

	// Cache the result for future use
	templateCache.Put(cacheKey, result)

	return result, nil
}

// Public cache management functions

// InvalidateTemplateCache clears cache entries for a specific template
func InvalidateTemplateCache(template string) int {
	return templateCache.InvalidateByTemplate(template)
}

// InvalidateVariableCache clears cache entries that depend on a specific variable
func InvalidateVariableCache(varName string) int {
	return templateCache.InvalidateByVariable(varName)
}

// ClearTemplateCache clears all cached template results
func ClearTemplateCache() int {
	return templateCache.InvalidateAll()
}

// GetTemplateCacheStats returns cache statistics
func GetTemplateCacheStats() (totalEntries, totalHits int) {
	return templateCache.GetStats()
}

// SetTemplateCacheTTL updates the cache time-to-live
func SetTemplateCacheTTL(ttl time.Duration) {
	templateCache.SetTTL(ttl)
}

// GetTemplateCacheTTL returns the current cache time-to-live
func GetTemplateCacheTTL() time.Duration {
	return templateCache.GetTTL()
}

// EvaluateComplexExpression handles method chaining and complex expressions
func EvaluateComplexExpression(expr string, env map[string]interface{}) (interface{}, error) {
	// Split by dot notation for method chaining
	parts := strings.Split(expr, ".")
	if len(parts) == 1 {
		// Simple expression, no method chaining
		return EvaluateExpression(expr, env)
	}

	// Start with the base variable
	baseValue, err := EvaluateExpression(parts[0], env)
	if err != nil {
		return nil, err
	}

	// Apply method chaining
	currentValue := baseValue
	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])

		// Handle method calls like replace(pattern, replacement) or trimStart()
		if strings.Contains(part, "(") && strings.HasSuffix(part, ")") {
			// Extract method name and arguments
			methodCall := part[:len(part)-1] // Remove closing )
			openParen := strings.Index(methodCall, "(")
			if openParen == -1 {
				return nil, fmt.Errorf("invalid method call syntax: %s", part)
			}

			methodName := methodCall[:openParen]
			argsStr := methodCall[openParen+1:]

			// Parse arguments (simple comma-separated for now)
			args := parseMethodArgs(argsStr, env)

			// Apply the method
			result, err := applyMethod(currentValue, methodName, args...)
			if err != nil {
				return nil, fmt.Errorf("method %s failed: %w", methodName, err)
			}
			currentValue = result
		} else {
			// Property access (not implemented yet)
			return nil, fmt.Errorf("property access not supported: %s", part)
		}
	}

	return currentValue, nil
}

// parseMethodArgs parses method arguments, handling regex literals and variables
func parseMethodArgs(argsStr string, env map[string]interface{}) []interface{} {
	if argsStr == "" {
		return []interface{}{}
	}

	args := []interface{}{}
	// Simple parsing - split by comma but handle regex literals
	parts := strings.Split(argsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Handle regex literals like /^/gm
		if strings.HasPrefix(part, "/") && strings.Contains(part, "/") {
			lastSlash := strings.LastIndex(part, "/")
			if lastSlash > 0 {
				regexPattern := part[1:lastSlash]
				flags := part[lastSlash+1:]
				// For now, just return the pattern as a string
				// In a full implementation, you'd create a regex object
				args = append(args, regexPattern)
				if flags != "" {
					args = append(args, flags)
				}
				continue
			}
		}

		// Handle string literals
		if (strings.HasPrefix(part, "'") && strings.HasSuffix(part, "'")) ||
			(strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"")) {
			args = append(args, part[1:len(part)-1])
		} else {
			// Try to evaluate as expression
			if val, err := EvaluateExpression(part, env); err == nil {
				args = append(args, val)
			} else {
				args = append(args, part)
			}
		}
	}

	return args
}

// applyMethod applies a method to a value
func applyMethod(value interface{}, method string, args ...interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("method %s can only be called on strings", method)
	}

	switch method {
	case "replace":
		if len(args) < 2 {
			return nil, fmt.Errorf("replace method requires 2 arguments")
		}
		pattern, ok1 := args[0].(string)
		replacement, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("replace arguments must be strings")
		}

		// Handle regex flags (basic support)
		if len(args) > 2 {
			if flags, ok := args[2].(string); ok && strings.Contains(flags, "g") {
				// Global replacement
				re, err := regexp.Compile(pattern)
				if err != nil {
					return nil, fmt.Errorf("invalid regex pattern: %w", err)
				}
				return re.ReplaceAllString(str, replacement), nil
			}
		}

		// Simple string replacement
		return strings.ReplaceAll(str, pattern, replacement), nil

	case "trimStart", "trimLeft":
		return strings.TrimLeft(str, " \t\n\r"), nil

	case "trimEnd", "trimRight":
		return strings.TrimRight(str, " \t\n\r"), nil

	case "trim":
		return strings.TrimSpace(str), nil

	case "toUpper", "upper":
		return strings.ToUpper(str), nil

	case "toLower", "lower":
		return strings.ToLower(str), nil

	case "length", "len":
		return len(str), nil

	case "substring", "substr":
		if len(args) < 1 {
			return nil, fmt.Errorf("substring method requires at least 1 argument")
		}
		start, ok1 := args[0].(int)
		if !ok1 {
			return nil, fmt.Errorf("substring start must be integer")
		}

		if len(args) >= 2 {
			end, ok2 := args[1].(int)
			if !ok2 {
				return nil, fmt.Errorf("substring end must be integer")
			}
			if start < 0 {
				start = len(str) + start
			}
			if end < 0 {
				end = len(str) + end
			}
			if start < 0 {
				start = 0
			}
			if end > len(str) {
				end = len(str)
			}
			if start > end {
				return "", nil
			}
			return str[start:end], nil
		} else {
			if start < 0 {
				start = len(str) + start
			}
			if start < 0 {
				start = 0
			}
			if start > len(str) {
				return "", nil
			}
			return str[start:], nil
		}

	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}
