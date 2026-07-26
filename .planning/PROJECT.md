# wordingo — Pure Go Word Document Library

## What This Is

A pure Go, MIT-licensed, zero-dependency library for creating Microsoft Word (.docx) documents from scratch or from existing templates — blank or pre-populated — while preserving styles for every document object: titles, headings, paragraphs, runs, tables, lists, headers, and footers. Also supports template merge with `{{placeholder}}` replacement, edit operations, text extraction, and GFM markdown conversion. Imported by Go applications as a library, not run as a CLI.

See `.planning/PRD.md` for the full product requirements document (authoritative source).

## Core Value

Create styled .docx documents in Go — from a template or from scratch — that open in Word looking exactly as designed. Style fidelity is the product; everything else is plumbing.

## Business Context

- **Customer**: Go backend developers building document generation into services, CI pipelines, and internal tools
- **Revenue model**: None — MIT open source; adoption is the metric
- **Success metric**: A Go developer generates a branded Word document from a template in under 10 lines of code, with zero dependencies, and it opens in Word without repair
- **Strategy notes**: The gap is licensing + fidelity. unioffice is AGPL/commercial; MIT alternatives lack style preservation. We win by being the only MIT library with a real style engine

## Requirements

### Validated

- [x] OPC package layer (ZIP + content types + relationships + namespace registry) — Phase 1 — v0.1.0
- [x] WML schema types (~60 essential WordprocessingML structs) — Phase 1 — v0.1.0
- [x] Blank document creation with professional defaults — Phase 1 — v0.1.0
- [x] STYLE-CLONE: template style dependency graph copy — Phase 2 — v0.1.0
- [x] STYLE-RESOLVE: effective property resolution — Phase 2 — v0.1.0
- [x] STYLE-ROUNDTRIP: edit + save with zero unintended changes — Phase 3 — v0.1.0
- [x] Content API: paragraphs, runs, formatting, named styles — Phase 4 — v0.1.0
- [x] Rich content: tables, images, headers/footers, lists, hyperlinks, page setup — Phase 5 — v0.1.0
- [x] Template merge with `{{placeholder}}` replacement (including split-run placeholders) — Phase 6 — v0.1.0
- [x] Edit operations: insert/delete paragraphs and rows, replace run text — Phase 6 — v0.1.0
- [x] Text extraction and GFM markdown export — Phase 06.1 — v0.1.0
- [x] Markdown import (CreateFromMarkdown, ImportMarkdown) — Phase 06.1 — v0.1.0

### Active

*(None — v0.1.0 shipped all planned requirements)*

### Out of Scope

- PDF conversion — separate product (pandoc, LibreOffice)
- .doc / .odt / .rtf / .xlsx / .pptx — focus on .docx only
- CLI tool, watch mode, HTML rendering, MCP server — not a library's job
- Word COM automation / CGO — defeats the purpose
- Porting any existing tool — ISO/IEC 29500 is authoritative; DocumentFormat.OpenXml is a conceptual reference only; no code ported
- Full OOXML spec coverage — ~60 types modeled; unknown elements preserved, not parsed
- Field codes, comments, bookmarks, content controls, tracked changes, charts, equations — v2, deferred behind validation

## Current State (v0.1.0)

Shipped 2026-07-26. 19,870 LOC Go, 7 phases, 18 plans, 37 requirements.

**What shipped:**
- OPC package, WML types, blank document generator
- Full style engine (clone + resolve + round-trip)
- Document model (open/read/save, FromTemplate)
- Content API (paragraph/run builders, named styles)
- Rich content (tables, images, headers/footers, lists, hyperlinks, page setup)
- Merge engine + edit operations with acceptance tests
- Text extraction + GFM markdown export/import

**Tech debt:**
- Test fixture corpus not committed (integration tests skip)
- Tables render after paragraphs only (v1 limitation)
- P2 resolver not wired into public API (Word resolves at open time)

## Context

Researched 2026-07-25 (see `.planning/research/`): 20+ Go docx modules surveyed. Complete ones are AGPL/commercial; MIT ones are partial or immature. The missing capability is style fidelity — documents that render like the template because styles.xml, numbering.xml, fontTable.xml, and theme.xml were handled as a dependency graph, not copied as files. Microsoft's DocumentFormat.OpenXml (.NET) provides the architectural reference for part structure and style inheritance. Implementation is from-scratch Go stdlib code (~bounded scope, not a transliteration).

## Next Milestone Goals

v0.2.0 / v1.1 — TBD. Potential candidates:
- Field codes (PAGE, DATE, NUMPAGES, TOC)
- Multiple sections with independent page setup
- Image positioning and text wrapping
- Comments, bookmarks, footnotes, endnotes
- Content controls (SDT), form fields

## Constraints

- **Tech stack**: Go 1.23+, standard library only — zero external dependencies, MIT license
- **Compatibility**: Output opens in Word 2016/2019/2021/M365, LibreOffice 7+, Google Docs — no repair dialogs
- **Conformance**: Write ISO 29500 Transitional; read Transitional + Strict
- **API**: Single public package; fluent builders; errors not panics; `io.ReaderAt`/`io.Writer` I/O
- **Binary size**: < 5 MB contribution (wordprocessingML only, no xlsx/pptx codegen)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Build from scratch, not wrap existing Go libs | AGPL on complete libs; MIT libs lack style engine | ✓ Good — v0.1.0 validated |
| Stdlib only (`archive/zip` + `encoding/xml`) | Zero deps = zero license/supply-chain review for adopters; trade-off is ~500 LOC namespace registry | ✓ Good — v0.1.0 validated |
| Library, not CLI; from-scratch, not a port | CLI infrastructure does not belong in a Go library; ISO 29500 is the spec | ✓ Good — v0.1.0 validated |
| Three separate style operations (clone/resolve/roundtrip) | Conflating them produces one mechanism that does none well | ✓ Good — Phase 2 validated |
| MVP proves style thesis before rich content | If template-styled output doesn't render correctly, tables/images don't matter | ✓ Good — thesis proven |
| Transitional conformance on write | Maximum compatibility (Office 2007 through M365) | ✓ Good — validated |
| Tables render after paragraphs only | Simpler implementation for v1 | ⚠️ v1 limitation, revisit for v2 |
| Merge by pointer identity (not index) | Avoids O(n²) index tracking during edits | ✓ Good — validated in UC1–UC4 |
| BodyElement struct union (not interface) | Allocation efficiency | ✓ Good — validated |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-07-26 after v0.1.0 milestone*
