---
phase: 01-foundation
plan: 01
subsystem: infra
tags: [go, opc, zip, ecma-376, docx, stdlib]

requires: []
provides:
  - internal/opc package: Open/Save with safety limits, canonical ordering, byte-exact pass-through
  - Sentinel error taxonomy (ErrInvalidPackage, ErrUnsafePath, ErrDecompressionLimit, ErrTooManyParts, ErrXMLDepth)
  - Module identity github.com/fabiomarini/wordingo (go 1.23)
  - testdata/ fixture directory layout (corpus pending user commit)
affects: [01-02 (xmlutil/wml built on opc), 01-03 (blank generator assembles via opc.Save)]

tech-stack:
  added: [archive/zip, encoding/xml]
  patterns: [sentinel errors + %w wrapping, lazy size-capped part reads, raw zip.Writer.Copy pass-through, monotonic rId high-water allocator, deterministic zip headers]

key-files:
  created:
    - go.mod
    - internal/opc/package.go
    - internal/opc/contenttypes.go
    - internal/opc/relationships.go
    - internal/opc/zipio.go
    - internal/opc/errors.go
    - internal/opc/opc_test.go
    - testdata/word/.gitkeep (+ libreoffice, googledocs, hostile)
  modified: []

key-decisions:
  - "NextRID uses a high-water mark established at parse time, not a rescan — deleted rIds are never reused (OPC-06)"
  - "DiffParts harness excludes modeled manifests ([Content_Types].xml, .rels) from byte comparison — they re-serialize canonically; OPC-04 byte-identity scope is unmodeled parts"
  - "Compression ratio >100:1 on small compressed entries is a Warnings() entry, not an error — declared-size caps carry the hard limit"

patterns-established:
  - "Sentinel errors wrapped with fmt.Errorf(\"opc: ...: %w\", err); errors.Is for callers (D-11)"
  - "Part payloads lazy via Part.Open with io.LimitReader-style capReader (QUAL-01)"
  - "Save validates then buffers: no partial output on failure"

requirements-completed: [OPC-01, OPC-02, OPC-04, OPC-05, OPC-06, OPC-07]

coverage:
  - id: D1
    description: "Open synthetic .docx: parts, content types, rels enumerated; Transitional + Strict conformance recorded"
    requirement: OPC-01
    verification:
      - kind: unit
        ref: "internal/opc/opc_test.go#TestOpen, #TestOpenStrictConformance"
        status: pass
    human_judgment: false
  - id: D2
    description: "Safety limits: path traversal, >4096 parts, oversized part fail with sentinel errors, no panics"
    requirement: OPC-07
    verification:
      - kind: unit
        ref: "internal/opc/opc_test.go#TestSafety"
        status: pass
    human_judgment: false
  - id: D3
    description: "Canonical write order: [Content_Types].xml then _rels/.rels first entries"
    requirement: OPC-02
    verification:
      - kind: unit
        ref: "internal/opc/opc_test.go#TestCanonicalOrder, #TestRoundTrip"
        status: pass
    human_judgment: false
  - id: D4
    description: "Unmodeled part (customXml blob) byte-identical after open→save; deterministic repeated saves"
    requirement: OPC-04
    verification:
      - kind: unit
        ref: "internal/opc/opc_test.go#TestRoundTrip, #TestSave/deterministic, #TestSave/modified_part_serializes"
        status: pass
    human_judgment: false
  - id: D5
    description: "Relationship integrity: monotonic rId allocation, dangling-target save fails closed, external rels warned never fetched"
    requirement: OPC-06
    verification:
      - kind: unit
        ref: "internal/opc/opc_test.go#TestRelationshipsNextRID, #TestSave/dangling_target_fails_closed, #TestOpen"
        status: pass
    human_judgment: false
  - id: D6
    description: "Real-producer fixture corpus (Word/LO/GDocs + hostile) round-trips byte-identical"
    requirement: OPC-04
    verification: []
    human_judgment: true
    rationale: "User-authored fixture corpus (D-02/D-03) not yet committed under testdata/{word,libreoffice,googledocs,hostile}/ — TestRoundTripHostileFixture auto-skips via HasFixtures until fixtures land"

duration: 12 min
completed: 2026-07-25
status: complete
---

# Phase 01 Plan 01: OPC Package Layer Summary

**OPC open/save layer on archive/zip with zip-bomb/traversal safety caps, canonical §9.1.4.2 entry ordering, raw pass-through byte-identity for unmodeled parts, and monotonic rId allocation — stdlib only, zero dependencies.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-07-25T17:19Z
- **Completed:** 2026-07-25T17:31Z
- **Tasks:** 3 (incl. Task 0 scaffold/checkpoint)
- **Files created:** 8 source + 4 .gitkeep

## Accomplishments

- `Package.Open` enumerates parts, content types, and relationship sets over `io.ReaderAt` with all OPC-07 limits enforced before decompression (4096 parts, 128 MiB/part, 512 MiB total, 100:1 ratio warning, strict part-name grammar)
- `Package.Save` emits canonically ordered, fully deterministic packages (two consecutive saves byte-identical); untouched parts raw-copied via `zip.Writer.Copy` — byte-identity by construction, not by recompression luck
- Relationship graph: monotonic rId high-water allocator (no reuse after delete), save-time validation that fails closed (dangling target → wrapped `ErrInvalidPackage`, zero bytes written), external rels preserved + warned, never fetched (SSRF guard)
- Conformance detection: Transitional vs Strict recorded on open from `word/document.xml` namespace (OPC-05)
- Module bootstrapped: `github.com/fabiomarini/wordingo`, go 1.23 (D-09/D-10)

## Task Commits

1. **Task 0: scaffold module + testdata dirs** — `98ba7f0` (chore)
2. **Task 1: OPC open path** — `72aa8a9` (feat)
3. **Task 2: OPC save path** — `848ef01` (feat)

## Files Created/Modified

- `go.mod` — module identity, go 1.23
- `internal/opc/errors.go` — sentinel taxonomy + limit constants
- `internal/opc/package.go` — Package, Part, Open, Save, name validation, conformance detection
- `internal/opc/contenttypes.go` — [Content_Types].xml model (Defaults/Overrides)
- `internal/opc/relationships.go` — Relationship sets, NextRID, validate, target resolution
- `internal/opc/zipio.go` — CanonicalOrder, NewDeterministicHeader
- `internal/opc/opc_test.go` — synthetic fixtures, safety tests, round-trip harness (DiffParts, HasFixtures)
- `testdata/{word,libreoffice,googledocs,hostile}/.gitkeep` — fixture layout

## Decisions Made

- NextRID high-water mark set at parse time (rescan once), then increments — plan wording said "scan existing ids," but the acceptance test requires deleted ids to never reissue, which only a persistent high-water mark guarantees.
- DiffParts excludes modeled manifests from byte diff: [Content_Types].xml and .rels are re-serialized canonically on save (sorted, compact), so their bytes legitimately differ from producer formatting. OPC-04's byte-identity scope is unmodeled parts, which the test proves exactly.
- Compression-ratio trips are warnings, not errors; hard stops come from declared-size caps.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] NextRID reissued deleted rId3**
- **Found during:** Task 1 verification (TestRelationshipsNextRID)
- **Issue:** Plan's "scan existing ids, issue max+1" reissues rId3 after rId3 is deleted — violating the plan's own acceptance criterion
- **Fix:** Added persistent high-water mark on Relationships, established by rescan at parse; NextRID increments from it
- **Files modified:** internal/opc/relationships.go, internal/opc/opc_test.go
- **Verification:** 100 allocations after deleting rId3, no reuse
- **Committed in:** `72aa8a9`

**2. [Rule 1 - Bug] TestRoundTrip diffed re-serialized manifests**
- **Found during:** Task 2 verification
- **Issue:** DiffParts compared [Content_Types].xml/.rels payloads, which change under canonical re-serialization — false failure, not an OPC-04 violation
- **Fix:** DiffParts excludes modeled manifests; unmodeled-part comparison remains strict per-byte
- **Files modified:** internal/opc/opc_test.go
- **Verification:** full suite green
- **Committed in:** `848ef01`

---

**Total deviations:** 2 auto-fixed (both Rule 1 - bugs against plan's own acceptance criteria)
**Impact on plan:** No scope creep; both fixes required for plan's stated criteria to be self-consistent.

## Pending Checkpoint (Task 0 — graceful skip active)

**checkpoint:human-action — fixture corpus (D-02/D-03) PENDING.**
`testdata/word|libreoffice|googledocs|hostile/` layout exists (`.gitkeep`), but the 6–8 user-authored .docx fixtures are not yet committed. (Five pre-existing files sit flat at `testdata/01-blank.docx`…`05-…`; they are not in the D-03 layout and were left untouched.)
Unit tests run against synthetic in-memory fixtures; `TestRoundTripHostileFixture` and per-producer tests auto-skip via `HasFixtures(t)` until fixtures land.
**To resolve:** author and commit fixtures per D-03 (blank.docx + styled.docx per producer + hostile/customxml-glossary.docx), or reply "skip fixtures" to accept the synthetic-only path.

## Issues Encountered

None beyond the auto-fixed deviations above.

## User Setup Required

None — no external services.

## Next Phase Readiness

- Ready for 01-02 (xmlutil namespace registry + WML types) — opc layer exposes everything it needs (Part.Open capped reads, sentinel `ErrXMLDepth` reserved, Conformance recorded)
- Blocker: none; fixture corpus pending does not gate 01-02 unit tests (same HasFixtures pattern applies)

## Self-Check: PASSED

- `go.mod`, `internal/opc/{package,contenttypes,relationships,zipio,errors}.go`, `internal/opc/opc_test.go`, `testdata/*/.gitkeep` — all exist on disk
- Commits `98ba7f0`, `72aa8a9`, `848ef01` — verified in `git log`
- Plan-level verification: `go build ./... && go vet ./... && go test ./internal/opc/ -count=1` → exit 0 (13 tests, 1 auto-skip)

---
*Phase: 01-foundation*
*Completed: 2026-07-25*
