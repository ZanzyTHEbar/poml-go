# POML API Reference

This document provides comprehensive API documentation for the advanced features in POML, including speaker mode and multimedia support.

## Table of Contents

- [Speaker Mode API](#speaker-mode-api)
- [Multimedia Support API](#multimedia-support-api)
- [Rich Content Types](#rich-content-types)
- [Go Implementation](#go-implementation)
- [JavaScript Implementation](#javascript-implementation)
- [Error Handling](#error-handling)

---

## 🎭 Speaker Mode API

Speaker mode enables structured conversation flows with automatic role assignment.

### Speaker Types

| Speaker  | Description                   | Use Case                         |
| -------- | ----------------------------- | -------------------------------- |
| `system` | System instructions and setup | Initial context, role definition |
| `human`  | User input and queries        | User messages, questions         |
| `ai`     | AI responses and actions      | Assistant responses, completions |

### Basic Usage

#### Go Implementation
```go
// Basic speaker mode
opts := &poml.WriteOptions{
    SpeakerMode: true,
}
result, err := poml.Write(pomlContent, opts)

// result is []poml.Message
```

#### JavaScript Implementation
```javascript
const { MarkdownWriter } = require('poml/writer');
const writer = new MarkdownWriter();

// With source mapping (enables speaker mode)
const result = writer.writeWithSourceMap(pomlContent);
// result contains speaker information
```

### POML Syntax

```xml
<poml>
<p speaker="system">You are a helpful AI assistant.</p>
<p speaker="human">Please help me with this task.</p>
<p speaker="ai">I'll assist you with that request.</p>
</poml>
```

### Automatic Speaker Detection

When speaker mode is enabled without explicit `speaker` attributes, POML automatically detects speakers based on content patterns:

```xml
<poml>
<p>System: Initialize as a coding assistant.</p>
<p>User: Help me write a function.</p>
<p>Assistant: I'll help you write that function.</p>
</poml>
```

### Output Format

#### Go Output Structure
```go
type Message struct {
    Speaker Speaker     `json:"speaker"`
    Content interface{} `json:"content"`
}

type Speaker string
const (
    SpeakerSystem Speaker = "system"
    SpeakerHuman  Speaker = "human"
    SpeakerAI     Speaker = "ai"
)
```

#### JavaScript Output Structure
```javascript
[
  {
    "speaker": "system",
    "content": "You are a helpful assistant"
  },
  {
    "speaker": "human",
    "content": "Please help me"
  },
  {
    "speaker": "ai",
    "content": "I'll assist you"
  }
]
```

---

## 🖼️ Multimedia Support API

POML supports embedding images and other media content directly in prompts.

### Image Embedding Methods

#### 1. File Reference
```xml
<img src="diagram.png" alt="System architecture diagram" />
```

#### 2. Base64 Encoding
```xml
<img base64="iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAI9jzyr5wAAAABJRU5ErkJggg=="
     alt="Small icon" />
```

### MIME Type Support

| Format   | MIME Type       | Support |
| -------- | --------------- | ------- |
| PNG      | `image/png`     | ✅ Full  |
| JPEG/JPG | `image/jpeg`    | ✅ Full  |
| GIF      | `image/gif`     | ✅ Full  |
| SVG      | `image/svg+xml` | ✅ Full  |
| WebP     | `image/webp`    | ✅ Full  |

### Multimedia Attributes

| Attribute | Description                  | Required    | Example             |
| --------- | ---------------------------- | ----------- | ------------------- |
| `src`     | File path to image           | No*         | `src="chart.png"`   |
| `base64`  | Base64 encoded image data    | No*         | `base64="iVBOR..."` |
| `alt`     | Alternative text description | Recommended | `alt="Sales chart"` |

*Either `src` or `base64` must be provided

### Processing Behavior

#### File Reference Processing
1. Reads file from specified path
2. Detects MIME type from file extension
3. Encodes file content to base64
4. Creates multimedia object

#### Base64 Processing
1. Validates base64 format
2. Extracts MIME type from data URL (if present)
3. Uses provided MIME type or detects from content
4. Creates multimedia object

### Error Handling

#### Missing Files
```xml
<!-- Graceful fallback for missing files -->
<img src="missing.png" alt="Fallback description" />
<!-- Result: Placeholder multimedia object -->
```

#### Invalid Base64
```xml
<!-- Invalid base64 handled gracefully -->
<img base64="invalid-data" alt="Error image" />
<!-- Result: Error multimedia object -->
```

---

## 📦 Rich Content Types

POML supports mixed content types through RichContent arrays.

### Content Type Detection

| Content Type        | Description               | Example                        |
| ------------------- | ------------------------- | ------------------------------ |
| `string`            | Plain text content        | `"Hello world"`                |
| `ContentMultiMedia` | Image/audio/video content | Multimedia object              |
| `[]interface{}`     | Mixed content array       | `[string, multimedia, string]` |

### RichContent Structure

#### Go Implementation
```go
type RichContent = []interface{}

type ContentMultiMedia struct {
    Type   string `json:"type"`     // MIME type
    Base64 string `json:"base64"`   // Base64 encoded data
    Alt    string `json:"alt,omitempty"` // Alternative text
}
```

#### JavaScript Implementation
```javascript
// Rich content array
[
  "Text content here",
  {
    "type": "image/png",
    "base64": "iVBORw0KGgoAAAANS...",
    "alt": "Image description"
  },
  "More text content"
]
```

### Mixed Content Example

```xml
<poml>
<p>Analysis results:</p>
<img src="chart.png" alt="Data visualization" />
<p>The chart above shows our performance metrics.</p>
<img src="details.png" alt="Detailed breakdown" />
<p>Key insights from the visualizations...</p>
</poml>
```

**Output:**
```json
[
  "Analysis results:",
  {
    "type": "image/png",
    "base64": "...",
    "alt": "Data visualization"
  },
  "The chart above shows our performance metrics.",
  {
    "type": "image/png",
    "base64": "...",
    "alt": "Detailed breakdown"
  },
  "Key insights from the visualizations..."
]
```

---

## 🔧 Go Implementation

### Core API

#### Write Function
```go
func Write(body string, opts *WriteOptions) (interface{}, error)
```

**Parameters:**
- `body`: POML content as string
- `opts`: Write options (optional)

**Returns:**
- `interface{}`: Result (string, []Message, or RichContent)
- `error`: Error if processing fails

#### WriteOptions
```go
type WriteOptions struct {
    SpeakerMode bool     // Enable speaker mode output
    BaseDir     string   // Base directory for file references
    Context     map[string]interface{} // Template context variables
}
```

#### Message Structure
```go
type Message struct {
    Speaker Speaker     `json:"speaker"`
    Content interface{} `json:"content"`
}

type Speaker string
const (
    SpeakerSystem Speaker = "system"
    SpeakerHuman  Speaker = "human"
    SpeakerAI     Speaker = "ai"
)
```

### Advanced Usage

```go
// Speaker mode with custom context
opts := &poml.WriteOptions{
    SpeakerMode: true,
    BaseDir:     "/path/to/files",
    Context: map[string]interface{}{
        "user_name": "Alice",
        "api_key":   "secret",
    },
}

result, err := poml.Write(pomlContent, opts)
if err != nil {
    log.Fatal(err)
}

// Type assertion for speaker mode
messages, ok := result.([]poml.Message)
if !ok {
    log.Fatal("Expected speaker messages")
}

for _, msg := range messages {
    fmt.Printf("%s: %v\n", msg.Speaker, msg.Content)
}
```

---

## 📜 JavaScript Implementation

### Core API

#### MarkdownWriter Class
```javascript
const { MarkdownWriter } = require('poml/writer');

const writer = new MarkdownWriter();
```

#### Basic Write Method
```javascript
const result = writer.write(pomlContent);
// Returns: string (regular mode)
```

#### Source Map Write Method
```javascript
const result = writer.writeWithSourceMap(pomlContent);
// Returns: object with speaker information and rich content
```

### Advanced Usage

```javascript
const { MarkdownWriter } = require('poml/writer');
const writer = new MarkdownWriter();

// Regular mode
const regularResult = writer.write(pomlContent);
console.log('Regular output:', regularResult);

// Speaker mode with source mapping
const speakerResult = writer.writeWithSourceMap(pomlContent);
console.log('Speaker output:', speakerResult);

// Access rich content
if (speakerResult.messages) {
    speakerResult.messages.forEach(msg => {
        console.log(`${msg.speaker}: ${msg.content}`);
    });
}

// Handle multimedia content
if (speakerResult.richContent) {
    speakerResult.richContent.forEach(item => {
        if (typeof item === 'string') {
            console.log('Text:', item);
        } else if (item.type && item.base64) {
            console.log('Multimedia:', item.alt, item.type);
        }
    });
}
```

### Configuration Options

```javascript
const writer = new MarkdownWriter({
    // Configuration options (if supported)
    enableRichContent: true,
    speakerMode: true,
    base64Encoding: true
});
```

---

## 🛡️ Error Handling

### Go Error Handling

```go
result, err := poml.Write(pomlContent, opts)
if err != nil {
    // Handle specific error types
    switch err.(type) {
    case *ParseError:
        log.Printf("Parse error: %v", err)
    case *TemplateError:
        log.Printf("Template error: %v", err)
    case *FileError:
        log.Printf("File error: %v", err)
    default:
        log.Printf("Unknown error: %v", err)
    }
    return
}
```

### JavaScript Error Handling

```javascript
try {
    const result = writer.write(pomlContent);
    // Process result
} catch (error) {
    if (error.name === 'ParseError') {
        console.error('POML parsing failed:', error.message);
    } else if (error.name === 'FileError') {
        console.error('File loading failed:', error.message);
    } else {
        console.error('Unknown error:', error);
    }
}
```

### Error Recovery Patterns

#### Safe Template Expressions
```xml
<!-- Go-style safe evaluation -->
<p>Length: {{content && content.length || 0}}</p>

<!-- JavaScript-style optional chaining -->
<p>Preview: {{content?.substring(0, 100) || "No content"}}</p>
```

#### File Fallbacks
```xml
<!-- Graceful file handling -->
<let src="data.txt" name="content" />
<p>Content: {{content || "Default content"}}</p>
```

---

## 📊 Performance Considerations

### Go Performance Characteristics
- **Startup**: Near-instantaneous
- **Simple parsing**: ~200µs
- **Complex expressions**: ~400µs
- **Speaker mode overhead**: +2-5%
- **Memory usage**: Minimal

### JavaScript Performance Characteristics
- **Startup**: Node.js initialization (~50-100ms)
- **Simple parsing**: ~165µs
- **Complex expressions**: ~165µs
- **Speaker mode overhead**: -22% (faster)
- **Memory usage**: V8 runtime overhead

### Optimization Strategies

#### Template Caching
```xml
<!-- Cache expensive operations -->
<let name="processed_content" value="{{content.replace(/\s+/g, ' ').trim()}}" />
<p>Processed: {{processed_content}}</p>
```

#### Batch Processing
```xml
<!-- Efficient batch operations -->
<let name="items" value="{{data.split('\n')}}" />
<for var="item" in="{{items}}">
<p>Item {{loop.index + 1}}: {{item}}</p>
</for>
```

---

## 🔍 Troubleshooting

### Common Issues

#### Speaker Mode Not Working
**Symptom:** Speaker assignments ignored
**Solution:**
```xml
<!-- ✅ Correct -->
<p speaker="system">Content here</p>

<!-- ❌ Incorrect -->
<p>system: Content here</p>
```

#### Multimedia Files Not Loading
**Symptom:** Placeholder content instead of images
**Solutions:**
```xml
<!-- ✅ Check file path -->
<img src="./images/chart.png" alt="Chart" />

<!-- ✅ Use absolute paths -->
<img src="/full/path/to/image.png" alt="Image" />

<!-- ✅ Verify file exists -->
<!-- Check file permissions and format -->
```

#### Template Expression Errors
**Symptom:** Expression evaluation fails
**Solutions:**
```xml
<!-- ✅ Safe evaluation -->
<p>{{content && content.length || 0}}</p>

<!-- ✅ Method existence check -->
<p>{{content?.substring && content.substring(0, 100) || "Fallback"}}</p>
```

### Debug Mode

#### Go Debug Mode
```go
opts := &poml.WriteOptions{
    Debug: true,  // Enable debug output
}
```

#### JavaScript Debug Mode
```javascript
const writer = new MarkdownWriter({
    debug: true  // Enable debug logging
});
```

---

## 📚 Related Documentation

- [README.md](../README.md) - Main project documentation
- [examples/README.md](../examples/README.md) - Usage examples
- [PERFORMANCE_ANALYSIS.md](../PERFORMANCE_ANALYSIS.md) - Performance benchmarks
- [INTEGRATION_COMPARISON.md](../INTEGRATION_COMPARISON.md) - Implementation comparison

---

## 🤝 Contributing

To contribute to the API:

1. **Add new features** following the established patterns
2. **Update documentation** when changing APIs
3. **Add tests** for new functionality
4. **Follow semantic versioning** for breaking changes

### API Evolution Guidelines

- **Backward compatibility**: Maintain existing API contracts
- **Deprecation warnings**: Warn before removing features
- **Migration guides**: Provide upgrade paths
- **Versioning**: Use semantic versioning for releases

---

## 📞 Support

For API-related questions:

- **GitHub Issues**: Report bugs and request features
- **Discord**: Join community discussions
- **Documentation**: Check examples and guides

---

*Last updated: 2024*
*API Version: 1.0.0*
*Go Implementation: v1.0.0*
*JavaScript Implementation: v1.0.0*
