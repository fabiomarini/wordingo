---
phase: 06-merge-edit-v1-0
plan: 01
subsystem: merge-engine
tags: [merge, placeholder, split-run, header-footer, template]
requires:
  - phase: 05-rich-content
    provides: Document body API, header/footer parts, table API, opc package
provides:
  - Merge() — replace {{key}} in body paragraphs, table cells, headers, footers
  - Split-run placeholder detection and merge across adjacent runs
  - Missing-key warnings surfaced in WarnIngs()
  - ScopedParts option to control which part types scanned
  - walkHeaderParts/walkFooterParts helpers for OPC part traversal
affects: [06-02 edit operations]
tech-stack:
  added: []
  patterns:
    - In-place text replacement on CT_Text.Value
    - Split-run detection via char-offset span mapping
    - Header/footer part read-decode-encode-write cycle via xmlutil.SafeDecoder + xmlutil.NewEncoder + MarkModified
    - Warnings() accumulator for unused/missing keys
key-files:
  created:
    - merge.go
    - merge_test.go
  modified: []
key-decisions:
  - "nil opts defaults to all parts (Body, Tables, Headers, Footers)"
  - "Split-run merge preserves first-fragment formatting, empties subsequent runs"
  - "Non-text runs (br, tab, drawing) preserved during split-run merge — never deleted"
requirements-completed:
  - MERGE-01
  - MERGE-02
  - MERGE-03
---

## Summary

Built merge engine for template placeholder replacement across all content parts.

### Changes

- **merge.go** (NEW, 260 lines): `Merge()` walks body paragraphs, table cells, header parts, and footer parts. `replacePlaceholders()` detects `{{key}}` patterns — including placeholders split across adjacent runs — and replaces text while preserving first-fragment formatting. Split-run detection uses char-offset span mapping. Missing data keys that are present in the data but not found in any part surface as warnings via `d.warn()`. Non-text runs (`<w:br>`, `<w:tab>`, `<w:drawing>`) are preserved during split-run merge.

- **merge_test.go** (NEW, 274 lines): 9 tests covering body merge, table cell merge, header/footer merge, split-run merge (with and without non-text runs), missing-key warnings, nil opts default, scoped parts filtering, and key-not-found-in-document warnings.

### Deviations

None.
