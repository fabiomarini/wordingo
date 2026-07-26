# Example 06: Lists

## What it demonstrates

Bulleted lists, ordered lists, nested outline lists (with explicit nesting levels), and the `AddListFromSlice` convenience constructor. The full 9-level nesting depth (levels 0-8) of the WML numbering model is supported.

## How to run

```
cd examples/06-lists
go run main.go
```

## Output

- `output.docx` — four list examples: a bulleted shopping list, an ordered recipe list, a 3-level nested outline, and an `AddListFromSlice`-built ordered list.

## Code walkthrough

- `list := doc.AddList(false)` — creates a new bulleted abstract+numbering definition in `word/numbering.xml` (or merges into the existing one from a template). Returns `*ListBuilder` linked to the new `numId`; `true` produces an ordered (decimal) list instead.
- `list.AddItem("Apples", 0)` — appends a paragraph whose `w:pPr/w:numPr` references the list's `numId` at the given `ilvl` (clamped to 0-8). Returns `*ListBuilder` for chaining.
- The nested outline calls `outline.AddItem("Install dependencies", 1)` then `outline.AddItem("Validation", 2)` — same list, different levels, exercising the 9-level depth support and the `%N.`-style level text auto-generated per level.
- `doc.AddListFromSlice([]string{"One","Two","Three","Four"}, true)` — convenience wrapper: builds a new ordered list and adds each slice element as a level-0 item in one call. Returns the same `*ListBuilder` for any further customization.
- `AddNumberingDef("upperRoman", 1)` (not used in this example but supported) lets you specify a custom `numFmt` and start value applied across all 9 levels.