// Package style resolves effective paragraph and run properties through
// the OOXML style inheritance chain, and clones the style dependency
// graph from a source package into a fresh-empty target.
//
// Cloner (this file)
//
//   CloneStyles byte-copies the 5 style dependency-graph parts (styles.xml,
//   numbering.xml, fontTable.xml, theme/theme1.xml, settings.xml) from a
//   source *opc.Package into a fresh-empty target *opc.Package, allocates
//   fresh relationship ids, and registers content-type Overrides.
//
//   D-08  Fresh-empty-target only.  If the target already has any of the 5
//         style parts, CloneStyles returns ErrCloneTargetNotEmpty.  Merge-
//         by-styleId is deferred to a future phase — no partial-merge story
//         in v1.
//
//   D-09  All 5 parts are treated as raw bytes via opc.MarkModified
//         pass-through — no re-parse, no wml/xmlutil decode.  Invalid
//         source XML surfaces at resolve time (D-07), not at clone time.
//
//   The 5 OOXML relationship type URIs and content-type MIMEs below are
//   defined by ISO/IEC 29500-1 §15 (Package Relationships).
//
//   This file reuses Phase 1 opc primitives only (MarkModified, NextRID,
//   ContentTypes.Overrides) — per Don't Hand-Roll principle.  It does
//   NOT import internal/wml or internal/xmlutil.
//
//   Path-traversal safety: all part names are compile-time constants in
//   the cloneParts table below; no user-supplied names reach MarkModified,
//   ContentTypes.Overrides, or Relationship.Target.  Path traversal is
//   structurally impossible (inherited from opc.validatePartName on Open).
//
//   Wave-1 independence: this file imports only fmt, io, and internal/opc.
//   It does NOT call the resolver (02-01), theme.go (02-02), or
//   numbering.go (02-02) — those layers are post-clone.  Corpus
//   validation that exercises the resolver over cloned packages is
//   plan 02-04 (Wave 3).
package style

import (
	"fmt"
	"io"
	"strings"

	"github.com/fabiomarini/wordingo/internal/opc"
)

// cloneParts is the constant table of the 5 style dependency-graph parts.
//
// Each entry carries the package-relative part name (no leading slash),
// the canonical OOXML relationship type URI (ISO/IEC 29500-1 §15), and
// the content-type MIME string for the [Content_Types].xml Override entry.
//
// Order: styles, numbering, fontTable, theme, settings — all present parts
// are processed in this order.
var cloneParts = []struct {
	name    string // package-relative part name, e.g. "word/styles.xml"
	relType string // relationship type URI for word/_rels/document.xml.rels
	ct      string // content-type MIME for Override
}{
	{
		name:    "word/styles.xml",
		relType: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles",
		ct:      "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml",
	},
	{
		name:    "word/numbering.xml",
		relType: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering",
		ct:      "application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml",
	},
	{
		name:    "word/fontTable.xml",
		relType: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable",
		ct:      "application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml",
	},
	{
		name:    "word/theme/theme1.xml",
		relType: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme",
		ct:      "application/vnd.openxmlformats-officedocument.theme+xml",
	},
	{
		name:    "word/settings.xml",
		relType: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings",
		ct:      "application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml",
	},
}

// CloneStyles copies the 5 style dependency-graph parts from src into dst.
//
// The algorithm has three steps:
//
// Step 1 — Precondition scan (atomicity, D-08):
//
//	Iterate cloneParts and check dst.Parts[p.name] for each.  If any of
//	the 5 parts already exists in dst, return ErrCloneTargetNotEmpty
//	IMMEDIATELY — no byte copy, no relationship added, no content-type
//	set.  The full scan completes before any write to dst.
//
// Step 2 — Byte copy + relationships + content types:
//
//	Iterate cloneParts again.  For each part:
//	  - If the source part is absent (legitimate — a template without
//	    numbering.xml or fontTable.xml is valid), skip silently.
//	  - Read source bytes via Part.Open + io.ReadAll (capped by opc's
//	    internal capReader at MaxPartBytes).
//	  - dst.MarkModified(p.name, bytes) — byte pass-through (D-09).
//	  - dst.ContentTypes.Overrides["/"+p.name] = p.ct — the leading
//	    slash is critical (PATTERNS critical flag; TypeFor reads
//	    c.Overrides["/"+partName] at contenttypes.go:61).
//	  - dst.Rels["word/document.xml"].NextRID() allocates a fresh rId
//	    (OPC-06 high-water, no reuse); append a Relationship with the
//	    canonical type URI and package-relative Target.
//
// Step 3 — Return nil on success.
//
// Nil-package defensive: if src or dst is nil, CloneStyles returns a
// clear error wrapping ErrCloneTargetNotEmpty (no panic).
func CloneStyles(src, dst *opc.Package) error {
	if src == nil {
		return fmt.Errorf("style: clone: src is nil: %w", ErrCloneTargetNotEmpty)
	}
	if dst == nil {
		return fmt.Errorf("style: clone: dst is nil: %w", ErrCloneTargetNotEmpty)
	}

	// Step 1 — Precondition scan (atomicity, D-08).
	// Check ALL 5 parts before any byte copy.
	for _, p := range cloneParts {
		if _, exists := dst.Parts[p.name]; exists {
			return fmt.Errorf("style: clone %s: %w", p.name, ErrCloneTargetNotEmpty)
		}
	}

	// Step 2 — Byte copy, relationships, content types.
	for _, p := range cloneParts {
		srcPart, ok := src.Parts[p.name]
		if !ok {
			// Absent source part — skip silently (D-09).
			// A template without numbering.xml or fontTable.xml
			// is a valid Word document.
			continue
		}

		// Read source bytes (capped by opc.Part.Open).
		rc, err := srcPart.Open()
		if err != nil {
			return fmt.Errorf("style: open source %s: %w", p.name, err)
		}
		bytes, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return fmt.Errorf("style: read source %s: %w", p.name, err)
		}

		// Byte pass-through via MarkModified (D-09).
		dst.MarkModified(p.name, bytes)

		// Content-type Override with leading slash (PATTERNS critical flag).
		dst.ContentTypes.Overrides["/"+p.name] = p.ct

		// Relationship wiring in word/_rels/document.xml.rels (OPC-06).
		wordRels := dst.Rels["word/document.xml"]
		if wordRels == nil {
			// Defensive: fresh-empty target created by Create() may
			// not yet have a word/document.xml rels set.
			wordRels = &opc.Relationships{}
			dst.Rels["word/document.xml"] = wordRels
		}
		rid := wordRels.NextRID()
		// Target is relative to word/document.xml's directory ("word/").
		// OPC resolveTarget joins it with the source part's directory;
		// "word/styles.xml" would resolve to "word/word/styles.xml" —
		// wrong.  Strip "word/" prefix for the relationship target.
		relTarget := strings.TrimPrefix(p.name, "word/")
		wordRels.Rels = append(wordRels.Rels, opc.Relationship{
			ID:     rid,
			Type:   p.relType,
			Target: relTarget,
		})
	}

	return nil
}
