# Content API: paragraphs, runs, formatting, styles, lists, hyperlinks

## Paragraphs

```go
// Add a paragraph with text
p := doc.AddParagraph("Hello, World!")

// Read paragraph content
text := p.Text()     // concatenated run text
style := p.Style()   // paragraph style ID, "" if none

// Get the underlying WML struct
ctP := p.X()
```

### Insert and delete by reference

```go
target := doc.Paragraphs()[2]

// Insert before/after an existing paragraph
p := doc.InsertBefore(target, "New paragraph above")
p := doc.InsertAfter(target,  "New paragraph below")

// Delete a paragraph
doc.DeleteParagraph(target)
```

Insert/Delete use pointer identity — if the target paragraph was removed or not in the body, a warning is recorded instead of panicking.

## Runs

A paragraph contains one or more runs. Each run is a contiguous piece of text with uniform formatting.

```go
p := doc.AddParagraph("Base text")
r := p.AddRun(" appended text")

r.SetText("replacement")   // overwrite the run's text
r.ReplaceText("old", "new") // strings.ReplaceAll on run text
```

## Run formatting

Every setter returns `*Run` for chaining:

```go
r := p.AddRun("styled")
r.SetBold(true)
r.SetItalic(true)
r.SetUnderline("single")       // any Word underline value
r.SetFont("Consolas")
r.SetSize(12)                   // in points (stored as half-points)
r.SetColor("2E75B6")           // 6 hex digits — warns if invalid
r.SetHighlight("yellow")
r.SetStyle("EmphasisIntense")   // run-level character style
```

Bulk set via `RunFormat` struct:

```go
r.SetFormatting(wordingo.RunFormat{
    Bold:   &yes,
    Color:  &red,
    Size:   &size,
})
```

## Paragraph formatting

```go
p.SetAlignment(wordingo.AlignmentLeft)   // Left/Center/Right/Both

p.SetSpacing(&wordingo.ParSpacing{
    Before:   240,      // twips before paragraph
    After:    240,      // twips after paragraph
    Line:     360,      // line spacing in 240ths of a line
    LineRule: "auto",   // auto | exact | atLeast
})

p.SetIndent(&wordingo.ParIndent{
    Left:      720,     // twips from left margin
    Right:     360,     // twips from right margin
    FirstLine: 360,     // twips first-line indent
    Hanging:   360,     // twips hanging indent
})

p.SetPageBreakBefore(true)    // paragraph starts on a new page
```

Bulk set:

```go
p.SetFormatting(wordingo.ParFormat{
    Alignment: &align,
    Spacing:   &spacing,
    Indent:    &indent,
})
```

## Named styles

Blank documents ship with a full set of default style IDs:

| Category | Style IDs |
|----------|-----------|
| Title | `Title` |
| Subtitle | `Subtitle` |
| Headings | `Heading1` through `Heading9` |
| Body | `Normal` |
| Quote | `Quote`, `IntenseQuote` |
| Emphasis | `IntenseEmphasis`, `IntenseReference` |
| List | `ListParagraph` |
| Character | `DefaultParagraphFont`, `TitleChar`, `SubtitleChar`, `Heading1Char`–`Heading9Char`, `QuoteChar`, `IntenseQuoteChar` |

Apply via `SetStyle`:

```go
doc.AddParagraph("Chapter 1").SetStyle("Heading1")
p.AddRun("important term").SetStyle("IntenseEmphasis")
```

Using a style ID not defined in the document's `styles.xml` produces a warning but the file still opens in Word.

When you need custom style names, use a template with `FromTemplate` or `OpenTemplate` — the template's style definitions are cloned into the output.

## Lists

```go
// Convenience: one call, flat list
doc.AddListFromSlice([]string{"Red", "Green", "Blue"}, false)  // bulleted
doc.AddListFromSlice([]string{"First", "Second"}, true)        // ordered

// Builder for nested lists
lb := doc.AddList(true)            // true = ordered, false = bulleted
lb.AddItem("Item 1", 0)           // level 0
lb.AddItem("Sub-item", 1)         // level 1
lb.AddItem("Sub-sub", 2)          // level 2 (up to 8)

// Chainable
lb.AddItem("A", 0).AddItem("B nested", 1)

// Custom numbering format
doc.AddNumberingDef("upperRoman", 1)  // I, II, III ...
```

The library auto-generates numbering definitions and merges them with any numbering from a template. Nine nesting levels (0–8) are supported.

## Hyperlinks

```go
p := doc.AddParagraph("Visit ")
r := p.AddHyperlink("Go docs", "https://go.dev")
r.SetColor("0563C1").SetUnderline("single")
```

The returned `*Run` accepts all formatting methods. Each call creates a fresh relationship — no URI deduplication.
