# Phase 4: Content API - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-07-26
**Phase:** 04-content-api
**Areas discussed:** Run formatting API shape, Style application & resolver, Paragraph formatting API, Error handling in builder chain

---

## Run Formatting API Shape

| Option | Description | Selected |
|--------|-------------|----------|
| Per-attribute setters | SetBold, SetItalic, SetFont — each a method on *Run. Maps 1:1 to CT_RPr. | |
| Properties struct | Single SetFormatting(RunFormat) method | |
| Both | Per-attribute setters + RunFormat struct for bulk | ✓ |

| Option | Description | Selected |
|--------|-------------|----------|
| *Run (builder chain) | SetBold(true).SetFont("Arial") — builder pattern. Errors deferred to Save/Warnings(). | ✓ |
| error (pre-validation) | Each setter validates and returns error immediately | |

| Option | Description | Selected |
|--------|-------------|----------|
| API-01 subset (7 attrs) | Bold, italic, underline, font, size, color, highlight | ✓ |
| Full CT_RPr (17 attrs) | All run property fields | |

| Option | Description | Selected |
|--------|-------------|----------|
| AddRun appends, returns *Run | paragraph.AddRun("text") appends new CT_R | ✓ |
| AddRun + InsertRun/RemoveRun | Granular run manipulation (deferred to Phase 6) | |

**User's choice:** Both per-attribute setters + RunFormat struct. Builder chain. API-01 subset. AddRun appends only.

---

## Style Application & Resolver

| Option | Description | Selected |
|--------|-------------|----------|
| Style reference only (set pStyle/rStyle) | Simple, respects OOXML model | ✓ |
| Reference + pre-populate resolved properties | Resilient but duplicates data | |
| Both — separate methods | SetStyle + ApplyResolvedStyle | |

| Option | Description | Selected |
|--------|-------------|----------|
| Paragraph.SetStyle(name) | On Paragraph, needs Doc back-reference | ✓ |
| Document.SetParagraphStyle() | On Document, no back-reference needed | |

| Option | Description | Selected |
|--------|-------------|----------|
| Internal only | Resolver stays internal to Document | ✓ |
| Exposed via Document | doc.StyleResolver() for advanced users | |

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — run.SetStyle(name) | Character-level named styles | ✓ |
| No — v1 only | Defer run styles | |

**User's choice:** Style reference only. Paragraph.SetStyle(name). Internal resolver. Run.SetStyle(name) included.

---

## Paragraph Formatting API

| Option | Description | Selected |
|--------|-------------|----------|
| Per-category setters | SetAlignment, SetSpacing, SetIndent | ✓ |
| Single ParagraphFormat struct | Fewer methods, bulkier calls | |

| Option | Description | Selected |
|--------|-------------|----------|
| Setters only (no struct) | Minimal API | |
| Both setters + struct | Symmetric with run formatting | ✓ |
| Struct only | Single approach | |

| Option | Description | Selected |
|--------|-------------|----------|
| API-02 subset (3 categories) | Alignment, spacing, indentation | ✓ |
| Full CT_PPr | All paragraph property fields | |

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — *Paragraph (builder chain) | Matches run setters | ✓ |
| No — void | No chaining | |

**User's choice:** Per-category setters + ParFormat struct. API-02 subset. Builder chain (*Paragraph return).

---

## Error Handling in Builder Chain

| Option | Description | Selected |
|--------|-------------|----------|
| Deferred — Warnings() at Save | Consistent with Phase 1/2 | ✓ |
| doc.Errors() accumulator | More explicit | |
| doc.Err() first error | Simple single-error | |
| Hybrid — store + check at Save | Combination | |

| Option | Description | Selected |
|--------|-------------|----------|
| Invalid input → warning, nil state → panic | Matches Go stdlib | ✓ |
| All warnings, never panic | Most defensive | |
| Accumulate → return at Save/WriteTo | Hard error at finalizer | |

| Option | Description | Selected |
|--------|-------------|----------|
| No validation — data holder | Structs are plain data | ✓ |
| Validate on creation | Structs validate returns error | |

**User's choice:** Deferred to Warnings() at Save. Invalid input → warning, nil state → panic. Data holders no validation.

---

## the agent's Discretion

- Exact struct field names for `RunFormat` and `ParFormat`
- Alignment type enum values (e.g., AlignmentLeft, AlignmentCenter)
- Whether ParSpacing and ParIndent are standalone structs or inlined in ParFormat
- Internal file layout within `wordingo/` for formatting code
- SaveFile vs Save naming standardization

## Deferred Ideas

None — discussion stayed within phase scope.
