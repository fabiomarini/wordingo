# Project Research Summary

**Project:** wordingo (Pure Go .docx Library)
**Domain:** Word document creation/editing with template style preservation
**Researched:** 2026-07-25
**Confidence:** HIGH

## Executive Summary

This project builds a **pure Go library** for reading, creating, and editing .docx files with **template style preservation** as the core differentiator. No existing Go library provides this under a permissive license — unioffice/gooxml is AGPL/commercial, and all MIT-licensed alternatives (godocx, docxgo, nguyenthenguyen/docx) are partial, immature, or lack style resolution entirely. The recommended approach: build from Go standard library (`archive/zip` + `encoding/xml`) with zero external dependencies, targeting Go 1.23+.

**Key recommendation:** Do NOT wrap any existing Go .docx library. Implement the full OPC package abstraction (ZIP + content types + relationships), hand-write the essential ~50 WML schema types, and build a dedicated style resolution engine that follows the ISO 29500 basedOn chain. This is a bounded 2-3 month effort, not a massive rewrite — DocumentFormat.OpenXml-based implementations cover the same subset in ~50KLOC.

**Key risks:** (1) XML namespace handling in `encoding/xml` is fragile — OOXML uses 40+ namespaces with varying prefix conventions across producers. Mitigate: namespace registry with URI-based matching. (2) Style resolution is the hardest non-trivial component — the basedOn inheritance DAG must be correct or all formatting breaks. Mitigate: implement explicit recursive merger and test against real templates. (3) Round-trip fidelity requires hoarding unknown XML elements, or editing a document silently destroys unsupported features. Mitigate: pass-through preservation at the OPC package layer.

## Key Findings

### Recommended Stack

Build a native Go library using only **Go standard library** packages. This is the correct choice after evaluating every major Go .docx library and finding none that meet the permissive-license + style-preservation + round-trip requirements.

**Core technologies:**
- **Go 1.23+**: Target version. `iter` package for clean document traversal; `io/fs` for embedded template support.
- **`archive/zip` (stdlib)**: OPC container read/write. .docx IS a ZIP with OPC conventions. No CGO, no external deps.
- **`encoding/xml` (stdlib)**: XML document parts parsing with namespace-aware token streaming. Must handle 40+ OOXML namespaces.
- **`image` + `image/jpeg`/`image/png` (stdlib)**: Embedded image metadata extraction.

**Key decision: Build new, not borrow.** Rationale:
- Only complete Go libraries (unioffice) use AGPL — incompatible with permissive licensing
- No existing library treats style preservation as a first-class feature
- Implementation scope is bounded (~50-80 initial WML types, ~50KLOC total)
- Zero external deps = no supply-chain risk, no version conflicts, no CGO issues

See [STACK.md](./STACK.md) for full library comparison matrix.

### Expected Features

**Must have (table stakes) — P1 for v1.0:**
- OPC package layer (ZIP + content types + relationships) — foundation of everything
- Read .docx with style preservation — round-trip XML without stripping unknown elements
- Create new .docx from scratch — minimal valid OPC package
- Paragraph + run formatting (bold, italic, underline, font, size, color, alignment, spacing)
- Named styles (Heading1-Normal-custom) with resolution from template
- Template-based creation — clone styles/numbering/theme from reference .docx
- Tables with rows, cells, basic borders, cell shading
- Images (PNG/JPG) with drawingML anchors
- Headers and Footers (basic + first-page/odd-even variants)
- Lists (ordered/bulleted with numbering definitions)
- Hyperlinks on text runs
- Page setup (margins, orientation, paper size)
- `{{placeholder}}` template merge — replace markers in paragraphs and table cells
- Go struct → table mapping (reflection-based)

**Should have (competitive differentiators):**
- Style preservation when editing existing docs — the core differentiator, no Go library does this correctly
- Template-based creation with reference-style inheritance — users bring branded templates
- Content preservation via unknown XML passthrough — edit without losing unsupported features
- Insert/delete paragraphs and runs in existing docs (P2)
- Field codes — PAGE, DATE, NUMPAGES fields (P2)

**Defer (v2+):**
- Comments, footnotes, bookmarks (P3)
- Content controls (SDT), form fields (P3)
- Tracked changes / revisions (P3)
- Watermarks, charts, equations (P3)
- Document validation, RTL/i18n (P3)
- Dump to JSON / batch replay (P3)

See [FEATURES.md](./FEATURES.md) for full feature matrix and gap analysis.

### Architecture Approach

Three-layer architecture: **Public API (docx/)** → **Style Resolution (style/)** → **OPC Package (opc/)**. Schema types (schema/) sit beneath the API layer as value objects with XML tags, wrapped by mutable API objects.

**Major components:**
1. **opc/ (OPC Package Layer)** — ZIP archive I/O, content type manifest, relationship graph management. Format-agnostic (reusable for xlsx/pptx). Wraps `archive/zip` with OPC conventions.
2. **schema/ (WML Schema Types)** — Go structs for OOXML elements (~50 initial, ~200 full). Flat structs with `encoding/xml` tags, no methods. Namespace-separated sub-packages (wordprocessingml/, drawingml/).
3. **style/ (Style Resolution Engine)** — Resolves effective paragraph/run properties through basedOn chain, latent styles, theme colors, numbering definitions. Recursive merge with cycle detection.
4. **docx/ (Public API)** — Document, Paragraph, Run, Table, Section wrappers using builder pattern. Lazy part loading. `X()` escape hatch for raw schema access.
5. **internal/** — XML utilities, ZIP utilities, UUID generation, test helpers.

**Key patterns:** Wrapper-over-Schema (two-layer API), Lazy Part Loading, Immutable Schema with Mutable Wrappers, Round-Trip Preservation via Unknown Element Hoarding. Must NOT pattern-match C# `DocumentFormat.OpenXml` — Go uses composition, builder pattern, and explicit marshaling, not inheritance, constructor nesting, or XML-backed models.

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed data flow diagrams and DocumentFormat.OpenXml→Go mapping.

### Critical Pitfalls

1. **Treating .docx as flat XML instead of OPC package** — Missing [Content_Types].xml, .rels files, or parts produces corrupt files Word rejects. **Prevention:** Implement strict OPC abstraction from day 1 (opc/ package) with relationship graph and content type validation. [Phase 1]

2. **Copying C# DocumentFormat.OpenXml patterns to Go** — Inheritance hierarchies, streaming parts, LINQ-to-XML queries produce overengineered Go code. **Prevention:** Design Go-first API with concrete structs, composition, builder pattern. Decide XML query strategy (struct unmarshal, not token walking) in Phase 0. [Phase 0]

3. **Namespace prefix sensitivity** — `encoding/xml` matches prefixes literally; different producers (Word, LibreOffice, Google Docs) use different prefix conventions causing silent parse failures. **Prevention:** Build URI-based namespace registry, use `xml:"{http://...}localName"` syntax, normalize prefixes on read. [Phase 1]

4. **Style resolution as flat key-value map** — Styles inherit through basedOn DAG (Heading1 ← Heading ← Normal). Copying styles.xml without the chain collapses all formatting to defaults. **Prevention:** Explicit recursive merger following ISO 29500 style hierarchy; copy entire dependency graph (styles + numbering + fontTable + theme) from templates. [Phase 3]

5. **Silent data loss on unknown elements** — Library drops parts it doesn't model (custom XML, mail merge, ActiveX, VBA). **Prevention:** Pass-through preservation — any unparsed part stored as raw bytes and re-emitted identically. Never drop unmodeled parts. [Phase 1]

6. **Transitional vs Strict conformance** — Writing Strict ISO 29500 breaks Office 2007 compatibility. **Prevention:** Default to Transitional for maximum compatibility; detect source conformance on read; match it on write. [Phase 1]

See [PITFALLS.md](./PITFALLS.md) for all 10 critical pitfalls, technical debt patterns, performance traps, and security considerations.

## Implications for Roadmap

### Phase 0: Architecture & API Contracts
**Rationale:** Foundational decisions (XML query strategy, Go API patterns, namespace handling, round-trip policy) must be made before any code. Changing these mid-build forces a rewrite. Addressed by Pitfalls 2 and 10 (C# transliteration, LINQ translation).

**Delivers:** API design document, namespace registry design, round-trip policy, WML type selection list, conformance class decision.

**Addresses:** Architecture decision (build-vs-borrow), Go API patterns.

**Avoids:** Pitfall 2 (C#→Go transliteration), Pitfall 10 (LINQ-to-XML query translation).

**Research flag:** No deeper research needed — patterns are well-established in unioffice's source structure and Go stdlib conventions. **Skip research-phase.**

---

### Phase 1: OPC Core Package
**Rationale:** Every other feature depends on correct ZIP + content type + relationship handling. Must be the first code written and the most battle-tested layer. Addressed by Pitfalls 1, 3, 5, 6, 9.

**Delivers:** `opc/` package — Package, Part, ContentTypes, Relationships, namespaces. Can create a minimal valid .docx that Word opens.

**Addresses:** FEATURES.md — OPC package layer (P1/foundation).

**Avoids:** Pitfall 1 (flat XML), Pitfall 3 (namespace prefixes), Pitfall 5 (unknown elements dropped), Pitfall 6 (conformance class), Pitfall 9 (relationship ID collision).

**Research flag:** **Needs research-phase** — OPC edge cases (external relationships, altChunk, embedded parts) need deeper study. ZIP ordering requirements from ECMA-376 Part 2 need precise implementation.

---

### Phase 2: WML Schema Types (Essential Subset)
**Rationale:** Schema types are value objects needed by every higher layer. Hand-write ~50 essential types covering document body, paragraphs, runs, text, styles, numbering, tables, sections. No library code depends on this — just types + marshal/unmarshal tests.

**Delivers:** `schema/wordprocessingml/` — CT_Document, CT_Body, CT_P, CT_R, CT_Text, CT_PPr, CT_RPr, CT_Style, CT_Styles, CT_Numbering, CT_Tbl, CT_SectPr, etc.

**Addresses:** Foundations for all document features.

**Avoids:** Pitfall 8 (whitespace handling — implemented at XML tag level with `xml:space` attributes).

**Research flag:** No deeper research — WML types are documented in ECMA-376 Part 1 and ISO 29500. Hand-write initial set; code-gen from XSD if time allows later. **Skip research-phase.**

---

### Phase 3: Style Resolution Engine
**Rationale:** Style preservation is the core differentiator. This is the hardest non-trivial component. Requires stable WML types (Phase 2) and OPC package (Phase 1) for loading template parts. Must work before any document creation API is built.

**Delivers:** `style/` package — StyleResolver with basedOn chain resolution, numbering resolution, font table handling, theme color resolution. Template style dependency graph copier.

**Addresses:** FEATURES.md — style preservation (CRITICAL differentiator), template-based creation, named styles.

**Avoids:** Pitfall 4 (flat style model), Pitfall 7 (template style dependency graph).

**Research flag:** **Needs research-phase** — style inheritance edge cases (linked styles, table style banding, latent styles, theme-dependent properties) need deeper study against ECMA-376 Part 4.

---

### Phase 4: Document Model (Read + Write)
**Rationale:** Needs opc/ (Phase 1), schema/ (Phase 2), and style/ (Phase 3). Implements lazy part loading, document body as flat paragraph/table sequence, round-trip serialization. Must pass the gate: open .docx → read → save → binary-identical XML parts.

**Delivers:** `docx/document.go` — Document.Open(), Document.Create(), Document.Save(). DocumentBody with paragraph/table/section content blocks. Lazy part loading infrastructure. Round-trip preservation via unknown element hoarding.

**Addresses:** FEATURES.md — read .docx with style preservation (P1), create new .docx (P1).

**Avoids:** Pitfall 8 (whitespace/run collapse — text handling in document model).

**Research flag:** No deeper research needed — pattern is well-established (unioffice, gooxml). **Skip research-phase.**

---

### Phase 5: Content Creation & Editing API
**Rationale:** The user-facing API. Builds on document model (Phase 4) and style resolution (Phase 3). This is where the library delivers value: adding paragraphs, runs, tables, images, headers, lists, template merge.

**Delivers:**
- Paragraph + Run with fluent builder (`doc.AddParagraph().AddRun().SetText("...").SetBold(true)`)
- Named styles application via resolver
- Tables with rows, cells, basic formatting
- Images (PNG/JPG) with drawingML anchors
- Headers/Footers (basic + variants)
- Lists (ordered/bulleted)
- Hyperlinks on text runs
- Page setup (section properties)
- `{{placeholder}}` template merge
- Go struct → table mapping
- Insert/delete paragraphs and runs (P2)

**Addresses:** FEATURES.md — all P1 features, P2 editing features.

**Avoids:** Pitfall 8 (run boundary preservation, whitespace handling in text).

**Research flag:** **Needs research-phase for template merge** — placeholder detection in complex mixed-format runs, content control alternative strategy. Also verify Go struct mapping patterns against existing Go reflection libraries.

---

### Phase 6: Advanced Features (v2 Ready)
**Rationale:** Deferred features that don't affect the core differentiator. Added after validation that the base library works for real users.

**Delivers:** Field codes (PAGE, DATE, NUMPAGES), multiple sections with independent setup, table row insert/delete, inline formatting in template merge, image sizing/positioning, document properties.

**Addresses:** FEATURES.md — P2 features.

**Avoids:** Not a pitfall phase — these features are additive, not corrective.

**Research flag:** **Needs research-phase for field codes** — fldChar begin/separate/end spanning multiple runs is complex XML structure. Needs deeper OOXML study.

---

### Phase 7: v2+ Polish Features
**Rationale:** Features that require significant OOXML coverage (comments, footnotes, bookmarks, content controls, tracked changes, watermarks, charts, equations).

**Addresses:** FEATURES.md — P3 features.

**Research flag:** **Needs per-feature research-phase** — each is a distinct OOXML sub-spec.

---

### Phase Ordering Rationale

- **Dependency-driven:** OPC → Schema → Style → Document → API is strict bottom-up. No phase can skip ahead (e.g., you can't build template-based creation without style resolution).
- **Risk-first:** Style resolution (Phase 3) is front-loaded because it's the highest-risk differentiator. If it can't be made to work correctly, the project should pivot early.
- **Inline with pitfall prevention:** Phase 0 prevents architectural pitfalls. Phase 1 prevents OPC-level pitfalls. Phase 3 prevents style pitfalls. Each phase explicitly avoids specific pitfalls.
- **MVP in Phase 5:** By Phase 5, the library can create styled documents from templates, include tables/images/headers/lists, and perform placeholder merge — the core MVP. Phase 6-7 are additive.
- **Research phases flagged:** Phase 1 (OPC edge cases), Phase 3 (style inheritance edge cases), Phase 5 (template merge), Phase 6 (field codes), Phase 7 (per-feature) need deeper research during planning. Others can skip.

### Research Flags

Phases needing deeper research during planning:
- **Phase 1:** OPC edge cases — external relationships, altChunk, embedded parts, strict ZIP ordering
- **Phase 3:** Style inheritance edge cases — linked styles, table style banding, latent styles, theme-dependent properties
- **Phase 5:** Template merge — placeholder detection in complex mixed-format runs, content control strategy
- **Phase 6:** Field codes — fldChar begin/separate/end spanning multiple runs
- **Phase 7:** Per-feature — each requires OOXML sub-spec study

Phases with standard patterns (skip research-phase):
- **Phase 0:** Architecture patterns well-documented from unioffice, gooxml
- **Phase 2:** WML types documented in ECMA-376, mechanical translation
- **Phase 4:** Document model pattern established by unioffice and gooxml

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Surveyed 20+ Go docx modules on pkg.go.dev. Verified source code of unioffice, godocx, docxgo, nguyenthenguyen/docx. License analysis confirmed. |
| Features | HIGH | DocumentFormat.OpenXml SDK documentation provides detailed feature reference. OOXML/ISO 29500 specs are complete. unioffice and godocx feature matrices verified against source. |
| Architecture | HIGH | unioffice source structure analyzed as reference. OOXML layered architecture well-documented in ECMA-376. DocumentFormat.OpenXml source structure verified. |
| Pitfalls | HIGH | Real bug reports from unioffice (issues #370, #401) and nguyenthenguyen/docx (issues #27, #30, #40) provide concrete failure patterns. ECMA-376 Part 2 (OPC) specifies ZIP ordering requirements. |

**Overall confidence:** HIGH

### Gaps to Address

- **Template merge edge cases:** How to reliably detect `{{placeholder}}` in runs that split across multiple `<w:r>` elements with mixed formatting. Need to prototype during Phase 5 planning.
- **Namespace handling strategy:** Confirm `encoding/xml` URI-based struct tags (`xml:"{http://...}localName"`) work correctly across all OOXML parts. May need custom `xml.Decoder` wrapper. Test against Word, LibreOffice, Google Docs output during Phase 1.
- **Code generation vs hand-written WML types:** Hand-writing ~50 types is tractable; ~200 types for full coverage is not. Decide in Phase 2 whether to build XSD→Go code generator or stay with hand-written subset.
- **Performance targets:** No baseline established. Test unioffice performance on real documents to set targets during Phase 1.

## Sources

### Primary (HIGH confidence)
- **pkg.go.dev (Go docx search)** — Surveyed 20+ Go docx modules, verified source structures and licenses
- **unioffice/unidoc** — GitHub (4.9k★) — Reference for Go-idiomatic OOXML library architecture; anti-pattern for AGPL licensing (issues #370, #401)
- **DocumentFormat.OpenXml (Microsoft)** — [.NET SDK](https://github.com/dotnet/Open-XML-SDK) — canonical part architecture and style inheritance reference; MIT license
- **ECMA-376 (OOXML) 5th ed.** — OPC packaging standard (Part 2), WordprocessingML definitions (Part 1), Markup Compatibility (Part 3), Transitional/Strict (Part 4)
- **DocumentFormat.OpenXml (Microsoft)** — .NET Open XML SDK — source of truth for part hierarchy patterns (not to be copied 1:1)

### Secondary (MEDIUM confidence)
- **gomutex/godocx** — GitHub (263★) — v0.x MIT library; reference for WML type subset
- **mmonterroca/docxgo** — GitHub (111★) — v2.7.2 MIT library; style preservation claims (unproven)
- **nguyenthenguyen/docx** — GitHub (unknown stars) — Simple text replacement library; issues #27, #30, #40 document real pitfalls
- **Eric White / OpenXmlSdkTs** — Confirms C# Open XML SDK patterns don't port well

### Tertiary (LOW confidence)
- OfficeOpenXML.com — WML reference (community site, not authoritative spec)
- StackOverflow community reports on namespace handling and content type management

---

*Research completed: 2026-07-25*
*Ready for roadmap: yes*
