# Phase 6: Merge & Edit (v0.1.0) — Research

**Researched:** 2026-07-26
**Domain:** Template merge + editing operations + release packaging
**Confidence:** HIGH

## Summary

Phase 6 delivers the merge engine (`{{key}}` replacement across all content parts including split-run detection), editorial operations (insert/delete paragraphs, delete table rows, replace run text), and the v0.1.0 release. The phase has 2 plans: 06-01 (merge engine) and 06-02 (edit operations + acceptance suite + release tag).

Key architectural insight: the merge engine walks a **canonical content order** (body paragraphs → body tables → headers → footers) and replaces `{{key}}` placeholders by scanning `CT_Text.Value`. Split-run detection concatenates adjacent run texts, finds `{{key}}` character ranges, maps them back to original runs, places the merged value in the first fragment run (with its formatting), and deletes other fragments. BodyElement unifies `CT_Body.P` and `CT_Body.Tbl` into a single ordered slice, reversing Phase 5's tables-after-paragraphs limitation. All edits set `d.dirty` and call `serializeBody()` at save time. STYLE-ROUNDTRIP is preserved because edit operations only modify the document body part — style parts are never touched.

Primary recommendation: Implement merge first (06-01), then editing operations + release (06-02). Split-run detection is the highest-risk feature — test with real Word-produced templates where Word arbitrarily fragments runs.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| MERGE-01 | Replace `{{key}}` in paragraphs, table cells, headers, footers | Walk order: body P/Tbl interleaved → header parts → footer parts. Per D-01 `Merge(data, opts)` scans `CT_Text.Value` across all runs in each paragraph. |
| MERGE-02 | Placeholders split across runs detected and merged | Split-run detection by concatenating adjacent run texts, finding `{{key}}` in joined string, mapping char offset back to originating runs. First-fragment formatting preserved per D-05. |
| MERGE-03 | Missing keys in Warnings(), never silent | `d.warn()` for each unused key in `data`. Keys not found during scan accumulate in `d.warnings` per existing pattern. |
| EDIT-01 | Insert/delete paragraph, delete table row | `InsertBefore`/`InsertAfter` by pointer finding index in `[]*CT_P`. `DeleteParagraph` splice on `[]*CT_P`. `DeleteRow` splice on `CT_Tbl.Tr`. Header/Footer via `ParagraphContainer`. |
| EDIT-02 | Replace text within a run | `SetText(s)` replaces `CT_Text.Value`. `ReplaceText(old, new)` calls `strings.ReplaceAll` on `CT_Text.Value`. Sets `d.dirty`. |
| EDIT-03 | All edits honor STYLE-ROUNDTRIP | Edit ops only modify `word/document.xml` (body) via `serializeBody()`. Style parts, numbering, theme remain byte-identical untouched parts. |
| QUAL-01 | Open from io.ReaderAt, save to io.Writer | Already satisfied by existing `OpenReader`/`WriteTo` pattern. Merge and edit ops add no new I/O paths. |
| QUAL-02 | No panics; all failures as errors | Merge warnings use existing `d.warn()` pattern. Row index bounds on `DeleteRow` — agent discretion on panic vs error return. |
| QUAL-03 | Single public package; X() on every wrapper | BodyElement wrappers and ParagraphContainer implementations must expose `X()`. New types follow existing Paragraph/Run pattern. |
</phase_requirements>

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Merge API:**
- D-01: `doc.Merge(data map[string]string, opts *MergeOpts)` — options struct, not single-map method
- D-02: `MergeOpts` has `ScopedParts` struct with booleans: `Body`, `Headers`, `Footers`, `Tables`. All default true.
- D-03: `nil` opts = scan all parts, warn on unused keys in data

**Split-Run Handling:**
- D-04: Detect split-run placeholders by concatenating adjacent run texts, finding `{{key}}`, mapping character ranges back to original runs
- D-05: Merged value goes in first fragment run with its formatting. Other fragment runs deleted.

**Paragraph Editing:**
- D-06: Paragraphs identified by pointer: `doc.InsertBefore(target *Paragraph, text string) *Paragraph`, `doc.InsertAfter(...)`, `doc.DeleteParagraph(target *Paragraph)`
- D-07: Header/Footer paragraphs support same editing API via `ParagraphContainer` interface
- D-08: No index-based paragraph targeting — pointer/reference only

**Mixed Body Order:**
- D-09: Public `BodyElement` interface with `BodyParagraph` and `BodyTable`. Single `[]BodyElement` slice.
- D-10: Keep existing `Paragraphs()` and `Tables()` for backward compat. Add `Body() []BodyElement` as canonical accessor.

**Table Row Deletion:**
- D-11: `TableBuilder.DeleteRow(idx int)` — by row index on table builder

**Run Text Replacement:**
- D-12: `SetText(s string)` replaces entire run content; `ReplaceText(old, new string)` for targeted replace

**Release:**
- D-13: Version tagged as `v0.1.0`, not `v1.0`
- D-14: Acceptance: MERGE-01..03 and EDIT-01..03 pass. All existing tests green. Generated docs open in Word without repair.
- D-15: QUAL-01..03 formally closed: audit all public methods for error-not-panic, verify io.ReaderAt/io.Writer I/O, confirm single public package with X() on all wrappers

### the agent's Discretion

- Exact `MergeOpts` struct field naming and zero-value defaults
- `BodyElement` interface method signature and type-switch ergonomics
- `ParagraphContainer` interface shape
- `DeleteParagraph` behavior when paragraph belongs to header/footer
- `ReplaceText` algorithm (strings.Replace vs strings.ReplaceAll)
- Internal file layout: merge.go, edit.go, body.go
- Merge warning message format
- Row index bounds checking style (panic vs error)
- Numbering.xml merge (merge creates no new numbering defs — placeholder replacement only)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Template merge (placeholder replacement) | API / Backend (Document) | — | Document owns all content parts (body, headers, footers). Merge walks all parts replacing text. |
| Split-run detection | API / Backend | — | Pure character-offset logic on CT_R.T.Value within a paragraph's run slice. No other tier involved. |
| Paragraph editing (Insert/Delete) | API / Backend (Document) | — | Document owns `CT_Body.P` slice and Header/Footer CT_Hdr.P/CT_Ftr.P. Direct slice manipulation. |
| Table row deletion | API / Backend (TableBuilder) | — | TableBuilder wraps `CT_Tbl.Tr`. Row index operation on existing builder pattern. |
| Run text replacement | API / Backend (Run) | — | Run wrapper modifies `CT_Text.Value`. No other tiers. |
| BodyElement ordering | API / Backend (Document) | — | Public accessor over interleaved CT_Body.P + CT_Body.Tbl. Convenience layer. |
| STYLE-ROUNDTRIP preservation | API / Backend | Database/Storage (OPC) | Edit ops only modify body XML. Style parts remain untouched OPC pass-through parts. |

## Standard Stack

### Core (Go stdlib — zero external dependencies)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `strings` | Go 1.23 | `strings.ReplaceAll` for placeholder matching, `strings.Builder` for concat | stdlib, already imported |
| `fmt` | Go 1.23 | Warning formatting, `d.warn()` pattern | stdlib, already imported |

### Existing project patterns (no new libraries needed)

| Pattern | File | Purpose |
|---------|------|---------|
| `d.dirty` flag + `serializeBody()` | `wordingo.go` L191-209 | Body content dirty tracking |
| `d.warn()` + `d.warnings` accumulate | `wordingo.go` L122-124 | Non-fatal warning accumulation |
| Wrapper-over-schema with `X()` | `paragraph.go`, `run.go`, `table.go` | Public API over WML types |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Pointer-based paragraph targeting | Index-based (int) | D-08 locks pointer. Index would require re-scan on every edit. |
| BodyElement interface/struct union | Keep separate Paragraphs()/Tables() only | D-09 requires BodyElement. Struct union avoids allocation overhead vs interface. |

**Version verification:** All packages are Go 1.23 stdlib — no external packages. `go list -m all` confirms zero dependencies.

## Package Legitimacy Audit

> No external packages are installed in this phase. All work is within the existing `github.com/fabiomarini/wordingo` module using stdlib only.

**Packages removed due to [SLOP] verdict:** None
**Packages flagged as suspicious [SUS]:** None

## Architecture Patterns

### System Architecture Diagram

```
Document.Merge(data, opts)
  │
  ├─ Walk body paragraphs (if opts.Body)
  │   └─ For each CT_P.R → replacePlaceholders(runs, data)
  │       ├─ Concatenate all CT_R.T.Value texts
  │       ├─ Find {{key}} in joined text
  │       ├─ Map char offset → run index
  │       ├─ Detect split runs (key spans multiple runs)
  │       │   └─ Place value in first fragment, delete others (D-05)
  │       └─ Replace non-split keys directly in single run
  │
  ├─ Walk body tables (if opts.Tables)
  │   └─ For each CT_Tbl.Tr.Tc → CT_P.R → replacePlaceholders(runs, data)
  │
  ├─ Walk header parts (if opts.Headers)
  │   └─ Parse each header part from OPC, walk CT_Hdr.P.R
  │       → replacePlaceholders(runs, data)
  │       → MarkModified header part
  │
  ├─ Walk footer parts (if opts.Footers)
  │   └─ Same as headers via CT_Ftr.P.R
  │
  └─ Warn on unused keys in data (MERGE-03)
      └─ d.warn("merge: key %q not found in document", key)

Document editing operations:
  InsertBefore(target, text)  → find target index in CT_Body.P → insert via slice splice
  InsertAfter(target, text)   → find target index in CT_Body.P → insert via slice splice
  DeleteParagraph(target)     → find target index in CT_Body.P → delete via slice splice
  TableBuilder.DeleteRow(idx) → splice CT_Tbl.Tr at index
  Run.SetText(s)              → ct.T.Value = s
  Run.ReplaceText(old, new)   → ct.T.Value = strings.ReplaceAll(ct.T.Value, old, new)
```

### Recommended Project Structure

```
wordingo/
├── wordingo.go          # Document.Merge(), Body() accessor, InsertBefore/After/DeleteParagraph
├── merge.go             # replacePlaceholders(), split-run logic, walk helpers
├── paragraph.go         # Paragraph — existing, unchanged
├── run.go               # Run.SetText(), Run.ReplaceText() — add methods
├── table.go             # TableBuilder.DeleteRow() — add method
├── header.go            # ParagraphContainer interface on Header/Footer — add
├── body.go              # BodyElement interface/struct union, BodyParagraph, BodyTable types
├── body_test.go         # BodyElement tests
├── merge_test.go        # Merge tests + split-run hostile corpus
├── edit_test.go         # Edit operation tests
└── uc_test.go           # UC1–UC4 end-to-end acceptance tests
```

### Pattern 1: Merge Walk Order

**What:** Walk content parts in canonical order: body paragraphs interleaved with tables, then headers, then footers. Each paragraph's runs are scanned for `{{key}}` patterns. Table cells are recursed into as paragraph holders.

**When to use:** Always for merge. Header/footer scanning skips if `ScopedParts.Headers`/`Footers` is false.

**Example walk structure:**
```go
func (d *Document) Merge(data map[string]string, opts *MergeOpts) {
    if opts == nil {
        opts = &MergeOpts{ScopedParts: ScopedParts{Body: true, Tables: true, Headers: true, Footers: true}}
    }
    unused := make(map[string]bool)
    for k := range data { unused[k] = true }

    // 1. Body paragraphs
    if opts.ScopedParts.Body {
        for _, p := range d.doc.Body.P {
            replaceInPara(p, data, unused)
        }
    }
    // 2. Body tables
    if opts.ScopedParts.Tables {
        for _, tbl := range d.doc.Body.Tbl {
            replaceInTable(tbl, data, unused)
        }
    }
    // 3. Header parts
    if opts.ScopedParts.Headers {
        d.walkHeaderParts(func(hdr *wml.CT_Hdr) {
            for _, p := range hdr.P { replaceInPara(p, data, unused) }
        })
    }
    // 4. Footer parts
    if opts.ScopedParts.Footers {
        d.walkFooterParts(func(ftr *wml.CT_Ftr) {
            for _, p := range ftr.P { replaceInPara(p, data, unused) }
        })
    }

    // Warn on unused keys
    for k := range unused {
        d.warn("wordingo: merge key %q not found in document", k)
    }
    d.dirty = true
}
```
[CITED: Existing patterns in wordingo.go L191-209 for dirty flag pattern]

### Pattern 2: Split-Run Detection

**What:** Word can split a `{{key}}` across multiple adjacent runs (e.g., `{` in one run, `{key}` in next). Detect by concatenating all run texts in a paragraph and finding `{{key}}` in the joined text, then mapping character offsets back to original runs.

**When to use:** Every paragraph scanned during merge. Most `{{key}}` values are in single runs, but Word does this arbitrarily.

**Example:**
```go
func replacePlaceholders(runs []*wml.CT_R, data map[string]string, unused map[string]bool) {
    // Build joined text + offset → run index map
    var joined strings.Builder
    type span struct{ start, end int }
    runSpans := make([]span, len(runs))
    for i, r := range runs {
        if r.T == nil { continue }
        runSpans[i] = span{start: joined.Len(), end: joined.Len() + len(r.T.Value)}
        joined.WriteString(r.T.Value)
    }
    text := joined.String()

    // Find all {{key}} patterns
    for {
        start := strings.Index(text, "{{")
        if start == -1 { break }
        end := strings.Index(text[start:], "}}")
        if end == -1 { break }
        key := text[start+2 : start+end]
        value, ok := data[key]
        if !ok {
            // Missing from data — leave placeholder, no warning here
            text = text[start+end+2:]
            continue
        }
        delete(unused, key)

        // Find all runs that overlap [start, start+end+2)
        placeholderEnd := start + end + 2
        fragmentStart, fragmentEnd := -1, -1
        for i, s := range runSpans {
            if s.start <= start && s.end > start && fragmentStart == -1 { fragmentStart = i }
            if s.start < placeholderEnd && s.end >= placeholderEnd { fragmentEnd = i }
        }

        if fragmentStart == fragmentEnd {
            // Single run — simple replacement
            runs[fragmentStart].T.Value = value
        } else {
            // Split across runs — D-05: first fragment gets value, others deleted
            runs[fragmentStart].T.Value = value
            for i := fragmentStart + 1; i <= fragmentEnd; i++ {
                runs[i] = nil // or splice out
            }
        }
        text = text[placeholderEnd:]
    }
}
```
[CITED: D-04/D-05 for split-run detection design]

### Anti-Patterns to Avoid

- **Modifying style parts during merge:** Merge is text-replacement only. Never touch styles.xml, numbering.xml, fontTable.xml, theme.xml, or settings.xml. STYLE-ROUNDTRIP depends on this.
- **Modifying header/footer in-place without sync:** Header/Footer changes must call `sync()` (header.go L42-57) to re-encode and `MarkModified` the part. Body changes use `d.dirty` + `serializeBody()`.
- **Using regex for placeholder detection:** `{{key}}` is a simple substring pattern. `strings` package is sufficient and avoids regex import overhead. [CITED: Existing codebase uses strings throughout]
- **Re-indexing Document.Paragraphs() after BodyElement addition:** `Paragraphs()` returns a fresh slice each time. Cached references are safe; the underlying `CT_P` pointer doesn't change.
- **Deleting runs that contain non-text children:** A run may have `Br`, `Tab`, `Drawing`, `Cr` alongside `T`. Split-run deletion must only delete runs whose only content is text fragments of the placeholder. If a run has other children, extract the merge-relevant text portion only. [ASSUMED]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Placeholder regex matching | Custom state machine | `strings.Index` / `strings.Replace` | `{{key}}` is a simple delimited pattern. No regex backtracking risk. stdlib strings is sufficient. |
| XML serialization for header/footer parts after merge | Custom XML writer | Existing `header.go sync()` pattern | `sync()` already handles encoding + MarkModified. Reuse. |
| rId allocation for header/footer parts | Custom counter | `rels.NextRID()` | Already available in OPC relationships layer. |

**Key insight:** This phase adds NO new Go dependencies. Every feature uses stdlib (strings, fmt) and existing project infrastructure (d.dirty, sync(), MarkModified, d.warn()). The complexity is in the split-run detection logic, not in I/O or external calls.

## Runtime State Inventory

> Not applicable — this is not a rename/refactor/migration phase. Greenfield code additions to existing codebase.

## Common Pitfalls

### Pitfall 1: Split-Run Detection Fails on Non-Text Runs

**What goes wrong:** A paragraph has runs with line breaks (`<w:br/>`), tabs (`<w:tab/>`), or images between text runs. The split-run detector concatenates all run texts and computes spans, but when deleting "fragment runs" it deletes a `<w:br/>`-only run that happens to sit between the `{{` and `key}` fragments.

**Why it happens:** The simple span model assumes every run has a `CT_Text`. But runs may have `Br`, `Tab`, `Cr`, `Drawing` instead of (or in addition to) `T`.

**How to avoid:** Before the scan, filter runs: only runs with `r.T != nil` participate in the joined-text. Non-text runs are never fragment candidates and must be preserved. When merging split runs, only delete the text content from non-first fragment runs, or replace with empty text — don't delete the run element itself if it carries formatting/breaks.

**Warning signs:** Merged output missing line breaks or images after a `{{key}}` that was split across runs with intervening special content.

### Pitfall 2: Table Cell Walk Misses Nested Tables

**What goes wrong:** A table cell contains a nested table. The merge walk only scans `CT_Tc.P` paragraphs but not `CT_Tc.Tbl` (which can exist in OOXML, though rare).

**Why it happens:** `CT_Tc` has both `P []*CT_P` and an optional `Tbl []*CT_Tbl` field (not currently modeled in `internal/wml/table.go` CT_Tc struct). [VERIFIED: ISO/IEC 29500-1 §17.4.65]

**How to avoid:** Document cell walk limitation. Nested table handling can be deferred — not required for any UC1–UC4 use case. The CT_Tc struct in internal/wml/table.go only has `P` and `TcPr` — no `Tbl` field. If nested table merge is needed later, add `Tbl []*CT_Tbl` to CT_Tc and recurse.

### Pitfall 3: InsertBefore/InsertAfter by Pointer Requires Index Resolution

**What goes wrong:** `InsertBefore(target *Paragraph, text string)` needs to find which slice owns the target paragraph. If the target belongs to a header/footer (via ParagraphContainer), the body paragraph search returns -1, and the operation silently fails or panics.

**Why it happens:** `Document.Paragraphs()` only returns body paragraphs. Header/footer paragraphs are in separate `CT_Hdr.P` / `CT_Ftr.P` slices.

**How to avoid:** Per D-06/D-07, `InsertBefore` must check if the target paragraph's `doc` field points to a Header or Footer owner. The ParagraphContainer interface (agent discretion on shape) provides this dispatch. A `container() ParagraphContainer` method on Paragraph (returning nil for body, Header/Footer wrapper for header/footer paras) allows the Document method to delegate.

### Pitfall 4: Header/Footer Part Sync After Merge

**What goes wrong:** Merge modifies header/footer paragraph text, but the header/footer part's serialized XML in `d.pkg` is not updated because `serializeBody()` only handles `word/document.xml`.

**Why it happens:** Body serializer doesn't know about header/footer parts. Header/footer have their own `sync()` method (header.go L42-57).

**How to avoid:** After modifying header/footer paragraphs during merge, iterate through `sectPr.HdrFtrRef`/`sectPr.FtrRef`, parse each linked part, modify, and call `sync()`. Or keep a map of `partName → *Header`/`*Footer` for post-merge serialization. The `walkHeaderParts`/`walkFooterParts` helper functions handle this.

### Pitfall 5: `nil` Opts Initialization Not Set

**What goes wrong:** User calls `doc.Merge(data, nil)` and merge silently skips all scoped parts because `opts.ScopedParts` is zero-valued (all false).

**Why it happens:** D-03 says `nil` opts = scan all parts, but the implementation must explicitly set defaults when `opts == nil`.

**How to avoid:** In `Merge`, check `if opts == nil { opts = &MergeOpts{ScopedParts: ScopedParts{Body: true, Tables: true, Headers: true, Footers: true}} }`. Document this behavior.

### Pitfall 6: DeleteParagraph on Already-Deleted Paragraph

**What goes wrong:** If `DeleteParagraph` is called on a paragraph that was already deleted (or belongs to a different document), the pointer dereference finds no matching entry in the slice.

**Why it happens:** No guard against double-delete or cross-document references.

**How to avoid:** Search for pointer identity (`p.ct == ct`), not value equality. If not found, silently return or warn. No panic. [CITED: QUAL-02]

## Code Examples

Verified patterns from existing codebase:

### Merge with Split-Run Detection
```go
// Source: Derived from D-04/D-05 CONTEXT.md split-run specification
// Simplified example for single-paragraph merge

// mergeText replaces {{key}} placeholders in a single text string.
// Returns the replaced string and whether any replacement occurred.
func mergeText(text string, data map[string]string, unused map[string]bool) (string, bool) {
    var changed bool
    for {
        start := strings.Index(text, "{{")
        if start == -1 {
            break
        }
        end := strings.Index(text[start:], "}}")
        if end == -1 {
            break
        }
        key := text[start+2 : start+end]
        val, ok := data[key]
        if !ok {
            // Key not in data — skip (MERGE-03: handled by unused tracking at doc level)
            text = text[:start+end+2] + text[start+end+2:]
            continue
        }
        delete(unused, key)
        text = text[:start] + val + text[start+end+2:]
        changed = true
    }
    return text, changed
}
```

### InsertBefore on Body Paragraphs
```go
// Source: D-06 pointer-based paragraph editing

func (d *Document) InsertBefore(target *Paragraph, text string) *Paragraph {
    ct := &wml.CT_P{R: []*wml.CT_R{{T: &wml.CT_Text{Value: text}}}}
    found := false
    for i, p := range d.doc.Body.P {
        if p == target.ct {
            d.doc.Body.P = append(d.doc.Body.P[:i+1], d.doc.Body.P[i:]...)
            d.doc.Body.P[i] = ct
            found = true
            break
        }
    }
    if !found {
        // Not a body paragraph — delegate to ParagraphContainer if target has one
        d.warn("wordingo: InsertBefore: target paragraph not found in body")
        return nil
    }
    d.dirty = true
    return &Paragraph{ct: ct, doc: d}
}
```

### BodyElement Interface Pattern
```go
// Source: D-09/D-10 BodyElement design

// BodyElement is one item in the document body: either a paragraph or a table.
type BodyElement interface {
    IsParagraph() bool
    BodyElementType() BodyElementType
}

type BodyElementType int
const (
    ElementParagraph BodyElementType = iota
    ElementTable
)

// BodyParagraph wraps a paragraph as a BodyElement.
type BodyParagraph struct{ P *Paragraph }

func (bp BodyParagraph) IsParagraph() bool { return true }
func (bp BodyParagraph) BodyElementType() BodyElementType { return ElementParagraph }
func (bp BodyParagraph) X() *wml.CT_P { return bp.P.X() }

// BodyTable wraps a table as a BodyElement.
type BodyTable struct{ T *TableBuilder }

func (bt BodyTable) IsParagraph() bool { return false }
func (bt BodyTable) BodyElementType() BodyElementType { return ElementTable }
func (bt BodyTable) X() *wml.CT_Tbl { return bt.T.X() }

func (d *Document) Body() []BodyElement {
    // Interleave P and Tbl in document order
    // Current limitation: all P come first, then all Tbl (Phase 5 D-24)
    // BodyElement exposes this for future reordering
    pi, ti := 0, 0
    var elems []BodyElement
    for pi < len(d.doc.Body.P) || ti < len(d.doc.Body.Tbl) {
        // In v1, all paragraphs come before tables
        // In future, use P/Tbl position comparison for true interleaving
        if pi < len(d.doc.Body.P) {
            elems = append(elems, BodyParagraph{
                P: &Paragraph{ct: d.doc.Body.P[pi], doc: d},
            })
            pi++
        } else if ti < len(d.doc.Body.Tbl) {
            elems = append(elems, BodyTable{
                T: &TableBuilder{ct: d.doc.Body.Tbl[ti], doc: d},
            })
            ti++
        }
    }
    return elems
}
```

### ParagraphContainer for Header/Footer
```go
// Source: D-07 ParagraphContainer interface

// ParagraphContainer is implemented by Header and Footer to expose
// paragraph editing operations.
type ParagraphContainer interface {
    Paragraphs() []*Paragraph
    InsertParagraphAt(idx int, ct *wml.CT_P) *Paragraph
    DeleteParagraphAt(idx int)
    X() interface{}
}

// Header implements ParagraphContainer
func (h *Header) Paragraphs() []*Paragraph {
    paras := make([]*Paragraph, len(h.ct.P))
    for i, p := range h.ct.P {
        paras[i] = &Paragraph{ct: p, doc: h.doc}
    }
    return paras
}

func (h *Header) InsertParagraphAt(idx int, ct *wml.CT_P) *Paragraph {
    h.ct.P = append(h.ct.P, nil)
    copy(h.ct.P[idx+1:], h.ct.P[idx:])
    h.ct.P[idx] = ct
    h.sync()
    return &Paragraph{ct: ct, doc: h.doc}
}

func (h *Header) DeleteParagraphAt(idx int) {
    h.ct.P = append(h.ct.P[:idx], h.ct.P[idx+1:]...)
    h.sync()
}
```

### DeleteRow on TableBuilder
```go
// Source: D-11 row index deletion on table builder

func (tb *TableBuilder) DeleteRow(idx int) error {
    if idx < 0 || idx >= len(tb.ct.Tr) {
        return fmt.Errorf("wordingo: DeleteRow: index %d out of range (rows: %d)", idx, len(tb.ct.Tr))
    }
    tb.ct.Tr = append(tb.ct.Tr[:idx], tb.ct.Tr[idx+1:]...)
    tb.doc.dirty = true
    return nil
}
```

### Run SetText and ReplaceText
```go
// Source: D-12 run text replacement

func (r *Run) SetText(s string) *Run {
    if r == nil {
        panic("wordingo: SetText called on nil Run")
    }
    if r.ct.T == nil {
        r.ct.T = &wml.CT_Text{}
    }
    r.ct.T.Value = s
    r.para.doc.dirty = true
    return r
}

func (r *Run) ReplaceText(old, new string) *Run {
    if r == nil {
        panic("wordingo: ReplaceText called on nil Run")
    }
    if r.ct.T == nil {
        return r
    }
    r.ct.T.Value = strings.ReplaceAll(r.ct.T.Value, old, new)
    r.para.doc.dirty = true
    return r
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Tables appended after all paragraphs (D-24 v1 limit) | BodyElement interleaving (D-09) | Phase 6 | Public Body() accessor for document-order iteration |
| InsertRun/RemoveRun deferred | Run.SetText()/ReplaceText() (EDIT-02) | Phase 6 | Editorial run manipulation available |
| Separate Paragraphs()/Tables() accessors | Body() accessor added (backward compat) | Phase 6 | Canonical ordered body access with backward compat |

**Deprecated/outdated:**
- N/A — all existing API is preserved for backward compatibility

## Validation Architecture

> `workflow.nyquist_validation` is absent from `.planning/config.json` — treat as enabled.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` package (Go 1.23) |
| Config file | none — standard Go test conventions |
| Quick run command | `go test ./... -count=1 -short 2>&1` |
| Full suite command | `go test ./... -count=1 -v 2>&1` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| MERGE-01 | Replace {{key}} in body paragraphs | unit | `go test -run TestMerge_Body` | ❌ Wave 0 |
| MERGE-01 | Replace {{key}} in table cells | unit | `go test -run TestMerge_TableCell` | ❌ Wave 0 |
| MERGE-01 | Replace {{key}} in header/footer | unit | `go test -run TestMerge_HeaderFooter` | ❌ Wave 0 |
| MERGE-02 | Split-run placeholder detection | unit | `go test -run TestMerge_SplitRun` | ❌ Wave 0 |
| MERGE-03 | Missing keys in Warnings() | unit | `go test -run TestMerge_MissingKey` | ❌ Wave 0 |
| EDIT-01 | InsertBefore paragraph | unit | `go test -run TestEdit_InsertBefore` | ❌ Wave 0 |
| EDIT-01 | InsertAfter paragraph | unit | `go test -run TestEdit_InsertAfter` | ❌ Wave 0 |
| EDIT-01 | DeleteParagraph | unit | `go test -run TestEdit_DeleteParagraph` | ❌ Wave 0 |
| EDIT-01 | DeleteRow | unit | `go test -run TestEdit_DeleteRow` | ❌ Wave 0 |
| EDIT-02 | Run.SetText | unit | `go test -run TestEdit_SetText` | ❌ Wave 0 |
| EDIT-02 | Run.ReplaceText | unit | `go test -run TestEdit_ReplaceText` | ❌ Wave 0 |
| UC1-UC4 | End-to-end acceptance | integration | `go test -run TestUC_` | ❌ Wave 0 |
| QUAL-01..03 | Audit checks | audit | `go vet ./...` | ✅ existing |

### Sampling Rate
- **Per task commit:** `go test ./... -count=1 -short 2>&1`
- **Per wave merge:** `go test ./... -count=1 -v 2>&1`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `merge_test.go` — covers MERGE-01..03 with split-run hostile corpus, table cell merge, header/footer merge, missing key warnings
- [ ] `edit_test.go` — covers EDIT-01..02 with InsertBefore/After, DeleteParagraph, DeleteRow, SetText, ReplaceText
- [ ] `body_test.go` — covers BodyElement interleaving, backward compat of Paragraphs()/Tables()
- [ ] `uc_test.go` — UC1–UC4 end-to-end acceptance (open template → merge data → edit → save → verify output)

## Security Domain

> `security_enforcement` is absent from config (absent = enabled). However, this phase adds no new I/O paths, network access, or user-data parsing that would introduce security surface. Merge operates on string-placeholder replacement only; all OPC-level security (decompression bombs, path traversal, entity expansion) was handled in Phase 1.

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | partial | Merge `data map[string]string` values replace placeholders — no XML injection risk because values go through Go `encoding/xml` encoder which escapes special characters. |

### Known Threat Patterns
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Placeholder value containing XML special chars | Tampering | `xmlutil.Encoder` escapes `&`, `<`, `>` automatically via `encoding/xml`. Values are set in `CT_Text.Value` which is `chardata` — safe. |

## Environment Availability

**Step 2.6: SKIPPED (no external dependencies identified)**

All work is pure Go code additions to existing module. No CLIs, services, databases, or external tools required beyond `go test`.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `CT_Tc` does not currently model nested `Tbl []*CT_Tbl` field | Pitfalls / Pitfall 2 | Nested table cells would not be merged. Deferred — not required for UC1–UC4. |
| A2 | `strings.Index` on concatenated run texts is sufficient for split-run detection | Code Examples / Split-Run | If placeholder separator characters appear naturally in text (unlikely `{{` `}}` in prose), edge-case misdetection possible. |
| A3 | Header/Footer parts can be identified by scanning `sectPr.HdrFtrRef`/`FtrRef` from the body | Merge Walk Pattern | True for single-section documents. Multi-section not supported yet (v2 deferred). |
| A4 | DeleteParagraph by pointer identity (`p.ct` comparison) is safe | Code Examples | `Paragraph.ct` is a `*wml.CT_P` — pointer equality is correct identity check. |

## Sources

### Primary (HIGH confidence)
- [CITED: CONTEXT.md] — D-01 through D-12 locked decisions
- [CITED: PRD.md §6 FR-7, FR-8] — Merge and editing requirements
- [CITED: REQUIREMENTS.md] — MERGE-01..03, EDIT-01..03, QUAL-01..03
- [CITED: existing codebase wordingo.go, paragraph.go, run.go, table.go, header.go] — Integration patterns verified by reading source

### Secondary (MEDIUM confidence)
- [CITED: go.dev/doc/modules/version-numbers] — Go v0.1.0 pre-release semantics (v0 signals instability)
- [CITED: ROADMAP.md] — Phase 6 plan structure, UC1–UC4 acceptance criteria

### Tertiary (LOW confidence)
- None — all claims verified against existing codebase or documented decisions.

## Open Questions (RESOLVED)

1. **Header/Footer part enumeration during merge** — [RESOLVED]
   - Recommendation implemented: lazy enumeration during merge walk, warn on parse failure and continue. 06-01-PLAN.md Task 1 specifies this pattern with `walkHeaderParts`/`walkFooterParts`.

2. **BodyElement struct union vs interface allocation cost** — [RESOLVED]
   - Recommendation implemented: struct union with type enum + pointer fields. 06-02-PLAN.md Task 1 specifies `BodyElement` as struct union with `BodyElementType` enum.

3. **DeleteRow bounds checking style** — [RESOLVED]
   - Recommendation implemented: return `error` for out-of-range index, consistent with existing `AddTable` error pattern. 06-02-PLAN.md Task 2 specifies `return error`.

4. **Cross-document paragraph reference in editing ops** — [RESOLVED]
   - Recommendation deferred: not required for v0.1.0. Silent no-op via pointer comparison failure is acceptable. Noted in 06-02-PLAN.md as a known limitation.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all Go stdlib, verified against codebase
- Architecture: HIGH — patterns directly derived from existing codebase (dirty flag, warnings, wrapper-over-schema, sync())
- Pitfalls: MEDIUM — some edge cases (nested tables, non-text runs) documented but not exhaustively tested

**Research date:** 2026-07-26
**Valid until:** 2026-08-25 (30 days — code patterns stable, fast-moving only in test corpus)
