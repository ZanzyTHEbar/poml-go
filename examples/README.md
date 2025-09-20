# POML Example Library

This directory contains example POML files demonstrating a variety of use cases. Examples are organized by difficulty:

- **Beginner:** Filenames start with `1XX_`
- **Intermediate:** Filenames start with `2XX_`
- **Advanced:** Filenames start with `3XX_`

Each example highlights different POML features, such as structured prompting, data handling, and templating.

Non-POML examples (e.g., Python or JavaScript scripts) are prefixed with `4XX_` and use appropriate file extensions. Inline comments explain their usage.

## 🆕 Advanced Features Examples

### 🎭 Speaker Mode Examples
- **301_speaker_mode_basic.poml**: Basic speaker assignment with system, human, and AI roles
- **302_speaker_detection.poml**: Automatic speaker detection based on content patterns
- **303_speaker_conversation.poml**: Multi-turn conversation structure

### 🖼️ Multimedia Support Examples
- **304_multimedia_images.poml**: Image embedding with file references and base64
- **305_multimedia_rich_content.poml**: Mixed text and multimedia content arrays
- **306_multimedia_charts.poml**: Data visualization with embedded charts

### ⚡ Advanced Expression Engine
- **307_expressions_basic.poml**: String manipulation methods (length, substring, toUpper)
- **308_expressions_regex.poml**: Regular expression support with replace operations
- **309_expressions_chaining.poml**: Method chaining for complex transformations
- **310_expressions_loops.poml**: Loop variables and dynamic content generation

### 🔗 Complex Integration Examples
- **311_complex_document_processing.poml**: Full document analysis with multiple features
- **312_business_dashboard.poml**: Business intelligence with charts and data processing
- **313_multimodal_content.poml**: Text, images, and structured data integration

### 🛡️ Error Handling Examples
- **314_error_handling_files.poml**: Graceful handling of missing files
- **315_error_handling_expressions.poml**: Safe expression evaluation with fallbacks
- **316_error_recovery.poml**: Production-ready error recovery patterns

### 🚀 Performance Optimization Examples
- **317_performance_caching.poml**: Template expression result caching
- **318_performance_loops.poml**: Efficient loop processing patterns
- **319_performance_batch_processing.poml**: Large dataset processing optimization

## 📋 Usage Patterns & Best Practices

### Speaker Mode Guidelines
```xml
<!-- ✅ Recommended: Explicit speaker assignment -->
<p speaker="system">You are a helpful assistant.</p>
<p speaker="human">Please help me with this task.</p>
<p speaker="ai">I'll assist you with that.</p>

<!-- ✅ Auto-detection: Content-based assignment -->
<p>System: Initialize as a coding assistant.</p>
<p>User: Help me write a function.</p>
<p>Assistant: I'll help you write that function.</p>
```

### Multimedia Best Practices
```xml
<!-- ✅ File references for large images -->
<img src="large_diagram.png" alt="System architecture diagram" />

<!-- ✅ Base64 for small, frequent assets -->
<img base64="iVBORw0KGgoAAAANS..." alt="Small icon" />

<!-- ✅ Rich content arrays for mixed media -->
<poml>
<p>Analysis results:</p>
<img src="chart.png" alt="Data visualization" />
<p>Key insights from the chart above...</p>
<img src="details.png" alt="Detailed breakdown" />
</poml>
```

### Expression Engine Patterns
```xml
<!-- ✅ Safe expression evaluation -->
<p>Length: {{content && content.length || 0}}</p>
<p>Preview: {{content?.substring(0, 100) || "No content"}}</p>

<!-- ✅ Method chaining for complex operations -->
<p>Processed: {{content.replace(/\s+/g, ' ').trim().substring(0, 200)}}</p>

<!-- ✅ Loop variables for dynamic content -->
<for var="item" in="['doc1', 'doc2', 'doc3']">
<p>Document {{loop.index + 1}}: {{item}} ({{loop.total}} total)</p>
</for>
```

### Error Handling Strategies
```xml
<!-- ✅ Fallback values for missing files -->
<let src="data.txt" name="content" />
<p>Content: {{content || "Default content"}}</p>

<!-- ✅ Safe method calls -->
<p>Stats: {{content?.length || 0}} characters</p>

<!-- ✅ Conditional processing -->
<p>{{content && "Content loaded" || "Content unavailable"}}</p>
```

## 🔧 Implementation Examples

### Go Implementation
```go
// Basic usage
result, err := poml.Write(`<poml><p>Hello World</p></poml>`)

// With speaker mode
opts := &poml.WriteOptions{SpeakerMode: true}
result, err := poml.Write(pomlContent, opts)
```

### JavaScript Implementation
```javascript
// Basic usage
const { MarkdownWriter } = require('poml/writer');
const writer = new MarkdownWriter();
const result = writer.write(pomlContent);

// With source mapping
const result = writer.writeWithSourceMap(pomlContent);
```

## Contributing

Contributions are welcome! To add or improve an example:

- Follow the naming conventions above.
- Include clear explanations and comments in your examples.
- Place any assets in an `assets/` subdirectory within the example's folder.
- If your example includes expected output, place it in an `expects/` subdirectory.

Submit your changes via pull request.
