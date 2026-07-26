# Example 11: Text Extraction & Markdown Conversion

Phase 06.1 features: ExtractText, ToMarkdown, CreateFromMarkdown, ImportMarkdown.

## Features Demonstrated

### ExtractText — `extract.go`
| Feature | Usage |
|---------|-------|
| Full document text | `doc.ExtractText(nil)` extracts body, tables, headers, footers |
| Scoped parts | `ExtractOpts{ScopedParts: {Body, Tables, Headers, Footers}}` per-part toggle |
| Custom separator | `ExtractOpts{Separator: " | "}` custom delimiter |
| Header/footer exclusion | `ScopedParts{Headers: false}` skips header content |
| Read-only | No `d.dirty` mutation on source document |

### ToMarkdown — `markdown.go`
| Feature | Usage |
|---------|-------|
| GFM export | `doc.ToMarkdown(nil)` converts to GitHub Flavored Markdown |
| Heading detection | `Heading1`-`Heading6` paragraph styles → `#`-`######` |
| Inline formatting | `**bold**`, `*italic*`, `***bold+italic***` |
| Pipe tables | `\| Feature \| Status \|` with `\| --- \|` separator |
| Hyperlinks | `[text](url)` format from relationships |
| Images | `![alt](data:image/png;base64,...)` data URIs |
| List detection | Ordered (`1.`) and bullet (`-`) markers |
| Scoped parts | Same `ExtractOpts` control as ExtractText |

### CreateFromMarkdown — `markdown_import.go`
| Feature | Usage |
|---------|-------|
| Document from markdown | `CreateFromMarkdown(markdown)` returns `*Document` |
| Full block support | Headings, paragraphs, lists, tables, code blocks |
| Inline formatting | Bold, italic, code spans, links, images via data URI |
| Headings | `Heading1`-`Heading6` style mapping |

### ImportMarkdown — `markdown_import.go`
| Feature | Usage |
|---------|-------|
| Append to existing doc | `doc.ImportMarkdown(markdown)` adds content after body |
| Warning surface | Parse errors via `doc.Warnings()` |
| No panic on bad input | Malformed markdown produces warnings, continues |

## Internal GFM Parser — `internal/markdown/parser.go`

All import/export builds on a standalone GFM parser:
- `BlockType`: Paragraph, Heading, CodeBlock, OrderedList, BulletList, Table
- `Block`: Type, Level, Content, Lines, Cells, Language, ListItems, Inlines
- `InlineSpan`: Text, Bold, Italic, Code, LinkURL, LinkText, ImageURL, ImageAlt
- `Parse(input string) ([]Block, error)` — line-based state machine

## Validation

Run the example:

```bash
cd examples/11-text-extraction-markdown
go run .
```

Expected: All 18 verification checks pass.
