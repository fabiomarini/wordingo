# Text extraction and markdown

## Text extraction

Extract plain text from the document body, tables, headers, and footers.

```go
// All parts, newline-separated
text, _ := doc.ExtractText(nil)
```

With custom options:

```go
text, _ := doc.ExtractText(&wordingo.ExtractOpts{
    ScopedParts: wordingo.ScopedParts{
        Body:    true,
        Tables:  true,
        Headers: false,
        Footers: false,
    },
    Separator: " | ",
})
```

Extraction is read-only — it does not mark the document as modified.

### ExtractOpts

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ScopedParts` | `ScopedParts` | all true | Which parts to extract |
| `Separator` | `string` | `"\n"` | Separator between extracted text blocks |

## Markdown export

Convert any .docx to GitHub-flavored Markdown.

```go
md, _ := doc.ToMarkdown(nil)
os.WriteFile("output.md", []byte(md), 0644)
```

With scope control:

```go
md, _ := doc.ToMarkdown(&wordingo.ExtractOpts{
    ScopedParts: wordingo.ScopedParts{
        Body:    true,
        Tables:  true,
        Headers: false,
        Footers: false,
    },
})
```

### Markdown format

| .docx element | Markdown output |
|---------------|-----------------|
| Heading1–Heading6 | `#`–`######` ATX headings |
| Bold runs | `**bold**` |
| Italic runs | `*italic*` |
| Strike runs | `~~strike~~` |
| Tables | Pipe tables with separator row |
| Hyperlinks | `[text](url)` |
| Images | `![alt](data:image/png;base64,...)` data URIs |
| Lists | `-` bullet or `1.` numbered markers |
| Code blocks | Fenced code blocks for Consolas/Courier/Courier New runs |

## Markdown import

### Create from markdown

```go
doc, err := wordingo.CreateFromMarkdown(sourceMD)
defer doc.Close()
doc.Save("output.docx")
```

### Append markdown to existing document

```go
doc.ImportMarkdown("\n## More content\n\nAppended paragraph.")
```

### Supported markdown syntax

| Markdown | Result |
|----------|--------|
| `# Heading` | Heading1 paragraph style |
| `## Heading` | Heading2 paragraph style |
| `Paragraph text` | Normal paragraph |
| `- bullet item` | Bulleted list |
| `1. numbered item` | Ordered list (with nesting) |
| `\| Col1 \| Col2 \|` | Pipe table |
| `` `code` `` | Courier New run |
| `**bold**` | Bold run |
| `*italic*` | Italic run |
| `[text](url)` | Hyperlink |
| `![alt](data:...)` | Embedded image |
| `` ``` `` fenced blocks | Consolas paragraphs |

Parse errors are non-fatal — they surface in `Warnings()` and never panic.

## Round-trip example

```go
// docx → markdown
md, _ := doc.ToMarkdown(nil)

// markdown → new docx (round-trip)
imported, _ := wordingo.CreateFromMarkdown(md)

// Append more content
imported.ImportMarkdown("\n## Appendix\n\nAdditional notes.")
imported.Save("roundtrip.docx")
```
