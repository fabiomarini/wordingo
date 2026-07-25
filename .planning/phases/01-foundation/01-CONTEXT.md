# Phase 1: Foundation - Context

**Gathered:** 2026-07-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver the OPC package layer (ZIP I/O, [Content_Types].xml, relationship graph, canonical entry ordering, safety limits), the xmlutil URI-based namespace registry, ~60 essential WML struct types with round-trip fidelity, and a blank-document generator whose output opens in Word 2016/2019/2021/M365 and LibreOffice without repair. Requirements: OPC-01..07, WML-01..04, CREATE-01, CREATE-02.

</domain>

<decisions>
## Implementation Decisions

### Word-Compat Verification
- **D-01:** Verification strategy is a committed fixture corpus + golden files in CI; a manual Word-open check gates phase sign-off. No Open XML SDK validator container in CI.
- **D-02:** User authors real .docx fixtures now (produced by Word, LibreOffice, Google Docs) and commits them under `testdata/` — this unblocks OPC-03 (producer-identical parsing) immediately.
- **D-03:** Fixture set is minimal, 6–8 files: `blank.docx` + `styled.docx` (headings, table, list) per producer (Word, LibreOffice, Google Docs), plus one hostile file containing customXml and a glossary part for pass-through (OPC-04) tests.
- **D-04:** Round-trip fidelity (OPC-04) is asserted as a per-part byte diff (unzip both, diff each part) — not whole-file ZIP bytes, since ZIP metadata varies.

### XML Layer Strategy
- **D-05:** xmlutil is a token-stream wrapper over `encoding/xml` Decoder/Encoder: normalize prefixes→URIs on read, emit canonical prefixes on write. WML structs keep `encoding/xml` struct tags. No fully custom parser.
- **D-06:** WML-04 unknown-child hoarding uses raw token subtree blobs (an `xmlutil.RawXML`-type field per struct) re-emitted verbatim on save — not a generic element tree.
- **D-07:** WML struct fields use pointers for optional attributes/children (nil = absent). Value structs, no presence flags.
- **D-08:** Write policy is canonical OOXML prefixes (`w`, `r`, `a`, `wp`, …). Read accepts any prefix, resolved by namespace URI (OPC-03).

### Module & Repo Identity
- **D-09:** Module path is `github.com/fabiomarini/wordingo` (matches PRD import example).
- **D-10:** Go version floor is 1.23 in go.mod; CI tests 1.23 + latest stable.

### Error/Warning API Shape
- **D-11:** Error taxonomy is sentinel errors (e.g., `ErrInvalidPackage`, `ErrUnsafePath`) wrapped with `fmt.Errorf("...: %w", err)` plus context; callers use `errors.Is`. No typed error structs.
- **D-12:** `Warnings()` returns `[]string` of human-readable messages (per PRD sketch) — no typed Warning struct.

### the agent's Discretion
Internal package organization details within `internal/opc`, `internal/wml`, `internal/xmlutil`, the exact list of ~60 WML types (guided by WML-01 + RESEARCH.md), test naming/layout conventions, and safety-limit numeric thresholds (OPC-07) — propose values in RESEARCH.md.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product & Requirements
- `.planning/PRD.md` — Authoritative product spec: FR-1 (OPC), FR-2 (WML), FR-4 (blank creation), FR-9 (I/O), NFR table, architecture layout, API sketch
- `.planning/REQUIREMENTS.md` — OPC-01..07, WML-01..04, CREATE-01..02 normative requirement text
- `.planning/PROJECT.md` — Constraints (stdlib only, MIT, Go 1.23+, <5MB), Key Decisions table

### External Specifications (no local copies — cite in RESEARCH.md as needed)
- ISO/IEC 29500 / ECMA-376 Part 2 (OPC) §9.1.4.2 — canonical ZIP entry ordering (OPC-02)
- ISO/IEC 29500 Part 1 (WordprocessingML) — type definitions for the ~60 WML structs

</canonical_refs>

<code_context>
## Existing Code Insights

New project — repository contains only `LICENSE` and `.planning/`. No existing code to reuse, no patterns established. All architecture comes from `.planning/PRD.md` §9:

### Established Patterns (from PRD, locked)
- Layout: `wordingo/` public root package + `internal/opc`, `internal/wml`, `internal/style`, `internal/xmlutil`; strict layering (opc ⊥ wml; style depends on both)
- Wrapper-over-schema with `X()` escape hatch; lazy part loading; URI-based namespace registry; unknown-element hoarding

</code_context>

<specifics>
## Specific Ideas

- Fixture corpus lands under `testdata/` and must include one "hostile" file (customXml + glossary part) for pass-through tests.
- Phase sign-off includes a manual "opens in Word without repair" check on the generated blank document.

</specifics>

<deferred>
## Deferred Ideas

- Open XML SDK validator in CI (.NET container) — considered and declined for Phase 1; revisit if golden-file approach proves insufficient.
- V2 features per REQUIREMENTS.md (field codes, comments, bookmarks, SDT, tracked changes, charts, equations) — behind v1 validation.

</deferred>

---

*Phase: 01-foundation*
*Context gathered: 2026-07-25*
