# Roadmap: wordingo — Pure Go Word Document Library

## Overview

Six phases, each with a hard gate that proves the phase's thesis before proceeding. The first three phases prove the risky parts (OPC correctness, style fidelity, round-trip safety) before any rich content work begins. PRD is authoritative: `.planning/PRD.md`.

## Phases

- [ ] **Phase 1: Foundation** — OPC package, WML types, blank document
- [ ] **Phase 2: Style Engine** — STYLE-CLONE + STYLE-RESOLVE
- [ ] **Phase 3: Document Model** — open/read/save, STYLE-ROUNDTRIP, FromTemplate
- [ ] **Phase 4: Content API** — paragraphs, runs, formatting, named styles
- [ ] **Phase 5: Rich Content** — tables, images, headers/footers, lists, hyperlinks, page setup
- [ ] **Phase 6: Merge & Edit** — {{placeholder}} merge, insert/delete, v1.0

## Phase Details

### Phase 1: Foundation

**Goal**: Library opens any .docx as an OPC package, models essential WML types, and generates a blank .docx Word opens without repair
**Depends on**: Nothing
**Requirements**: OPC-01..07, WML-01..04, CREATE-01, CREATE-02
**Success Criteria**:

  1. Open a real .docx (from Word, LibreOffice, Google Docs) — all parts enumerated, namespaces resolve by URI regardless of prefix
  2. Unmodeled parts (customXml, glossary, VBA) survive open→save byte-identically
  3. Create() produces blank .docx that Word 2016/2019/2021/M365 and LibreOffice open with no repair dialog
  4. ~60 WML types marshal/unmarshal bidirectionally; whitespace and xml:space survive
  5. Safety limits enforced (zip bomb, entity expansion, path traversal)

**Plans**: 1/3 plans executed

Plans:

- [x] 01-01-PLAN.md
- [ ] 01-02-PLAN.md
- [ ] 01-03-PLAN.md

**Wave 1**

- [x] 01-01: opc package — ZIP I/O, content types, relationship graph, canonical entry ordering, safety limits
- [ ] 01-02: xmlutil namespace registry + wml essential types (~60 structs) with round-trip tests

**Wave 2** *(blocked on Wave 1 completion)*

- [ ] 01-03: Blank document generator — default styles, theme, fontTable, settings, sectPr; Word-open gate

### Phase 2: Style Engine

**Goal**: Given any template .docx, library clones its complete style dependency graph and resolves effective properties through the full inheritance chain
**Depends on**: Phase 1
**Requirements**: STYLE-CLONE-01..02, STYLE-RESOLVE-01..03
**Success Criteria**:

  1. Cloning a template copies styles + numbering + fontTable + theme + settings with valid relationships/content types
  2. Effective paragraph/run properties resolve correctly through docDefaults → latentStyles → basedOn chain → direct formatting
  3. Theme colors and numbering definitions resolve; circular basedOn detected without infinite recursion
  4. Test corpus: paragraph with "Heading2" (basedOn Heading1 basedOn Normal) resolves identical effective properties to Word's own rendering

**Plans**: 3 plans

Plans:

- [ ] 02-01: Style resolver — recursive basedOn merger, docDefaults, latentStyles, cycle detection
- [ ] 02-02: Theme color + numbering resolution; font table handling
- [ ] 02-03: Dependency graph cloner + real-template test corpus validation

### Phase 3: Document Model

**Goal**: Library opens, reads, and saves existing documents with zero unintended diffs; creates documents from templates (blank or pre-populated)
**Depends on**: Phase 2
**Requirements**: STYLE-ROUNDTRIP-01..02, CREATE-03, CREATE-04
**Success Criteria**:

  1. Open existing .docx → read body content → save: unmodified parts byte-identical, style parts untouched
  2. FromTemplate() produces new document whose body is empty but styles/theme/numbering are the template's
  3. Pre-populated template opens, keeps existing content, accepts insertion
  4. Lazy part loading: unopened parts never parsed

**Plans**: 2 plans

Plans:

- [ ] 03-01: Document open/read/save — lazy loading, body model, round-trip diff harness in CI
- [ ] 03-02: FromTemplate — clone package, clear or keep body, ready for content

### Phase 4: Content API

**Goal**: Users create fully formatted text documents programmatically
**Depends on**: Phase 3
**Requirements**: API-01..03, QUAL-01..03
**Success Criteria**:

  1. Paragraphs and runs with full inline formatting (bold, italic, underline, font, size, color, highlight)
  2. Paragraph formatting (alignment, spacing, indentation)
  3. Named styles applied by name and resolved through the style engine — output renders like template
  4. io.ReaderAt/io.Writer I/O; errors not panics; Warnings() exposed

**Plans**: 2 plans

Plans:

- [ ] 04-01: Paragraph/Run builders + inline + paragraph formatting
- [ ] 04-02: Named style application + public API polish (single package, X() escape hatch)

### Phase 5: Rich Content

**Goal**: Complete business documents: tables, images, headers/footers, lists, hyperlinks, page setup
**Depends on**: Phase 4
**Requirements**: API-04..09
**Success Criteria**:

  1. Tables with named styles, borders, shading, hMerge/vMerge render correctly in Word
  2. PNG/JPEG images embedded with explicit sizing
  3. Headers/footers (default, first-page, odd/even) with content
  4. Multi-level ordered/bulleted lists backed by numbering definitions
  5. Hyperlinks, page margins/orientation/size, page breaks

**Plans**: 3 plans

Plans:

- [ ] 05-01: Tables + images
- [ ] 05-02: Headers/footers + sections + page setup
- [ ] 05-03: Lists + hyperlinks

### Phase 6: Merge & Edit (v1.0)

**Goal**: Template merge with hostile-input correctness, editing operations, v1.0 release
**Depends on**: Phase 5
**Requirements**: MERGE-01..03, EDIT-01..03
**Success Criteria**:

  1. {{key}} replaced in paragraphs, table cells, headers, footers — including placeholders split across runs
  2. Missing keys in Warnings(), never silent
  3. Insert/delete paragraphs and rows; replace run text; all edits round-trip-safe
  4. All four PRD use cases (UC1–UC4) pass end-to-end; v1.0.0 tagged

**Plans**: 2 plans

Plans:

- [ ] 06-01: Merge engine — placeholder detection across split runs, all part types
- [ ] 06-02: Edit operations + UC1–UC4 acceptance suite + v1.0 release

## Progress

| Phase | Plans Complete | Status | Gate |
|-------|----------------|--------|------|
| 1. Foundation | 1/3 | In Progress|  |
| 2. Style Engine | 0/3 | Not started | Template styles render correctly |
| 3. Document Model | 0/2 | Not started | Zero unintended diffs on round-trip |
| 4. Content API | 0/2 | Not started | Formatted text doc programmatically |
| 5. Rich Content | 0/3 | Not started | Complete business document |
| 6. Merge & Edit | 0/2 | Not started | UC1–UC4 green, v1.0 tagged |

Total: 15 plans, 33 v1 requirements, 0 complete.
