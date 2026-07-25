// Package style — numbering level resolution.
//
// D-13: Resolve numId + ilvl to CT_Lvl (numFmt, lvlText, start, level pPr).
// The level's pPr (indentation) is merged into the paragraph's effective pPr.
// Does NOT include style-based numPr inheritance — bounded scope per D-13.
//
// D-07: Missing numId, ilvl, abstractNum, or numbering.xml part → warning
// + return nil (no numbering contribution to the effective pPr).
// No sentinel errors for resolver misses — problems surface via warnings.
//
// Pitfall 4 (lvlOverride): CT_Num may carry lvlOverride entries that
// override the abstractNum's level values (startOverride or full lvl
// replacement).  The numberingCache.processOverride method handles both.
//
// Cache invalidation: numberingCache.invalidateIfStale checks
// Part.IsModified on word/numbering.xml, following the same pattern as
// 02-01's styles.xml invalidation.
package style

import (
	"fmt"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// numberingCache lazily decodes numbering.xml and provides numId+ilvl
// resolution with lvlOverride support.  Cache invalidation via
// Part.IsModified following the 02-01 styles.xml pattern.
type numberingCache struct {
	numbering *wml.CT_Numbering // lazily decoded tree; nil until first resolve
	parsed    bool              // guards re-parse; set true BEFORE parse to prevent retry loops
	pkg       *opc.Package      // source for lazy read
	warn      func(string)      // warning callback threaded from Resolver.appendWarning
}

// newNumberingCache constructs a numberingCache bound to the given package
// and warning callback.  Does NOT parse eagerly.
func newNumberingCache(pkg *opc.Package, warn func(string)) *numberingCache {
	return &numberingCache{
		numbering: nil,
		parsed:    false,
		pkg:       pkg,
		warn:      warn,
	}
}

// ensureParsed lazily decodes word/numbering.xml into n.numbering.
// Idempotent — second call is a no-op (parsed guard).
// Sets parsed=true BEFORE parse so a failure does not retry indefinitely
// (the cache stays empty and every ResolveLvl warns "not found" — D-07).
func (n *numberingCache) ensureParsed() error {
	if n.parsed {
		return nil
	}
	n.parsed = true // prevent retry loops on parse failure

	part, ok := n.pkg.Parts["word/numbering.xml"]
	if !ok {
		return nil // no numbering part → empty cache; ResolveLvl warns per D-07
	}

	rc, err := part.Open()
	if err != nil {
		return fmt.Errorf("style: open numbering.xml: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)

	var nb wml.CT_Numbering
	if err := dec.Decode(&nb); err != nil {
		return fmt.Errorf("style: decode numbering.xml: %v: %w", err, ErrNumberingParseFailed)
	}

	n.numbering = &nb
	return nil
}

// ResolveLvl resolves a numId + ilvl to a CT_Lvl by walking the
// CT_Num → CT_AbstractNum → CT_Lvl chain, applying lvlOverride (Pitfall 4).
//
// Returns nil on any missing reference (D-07: warning + no numPr contribution).
// The returned lvl is a reference into the cached tree — callers MUST clone
// before mutating (the resolver's mergePPr handles this by cloning composite
// leaves).
func (n *numberingCache) ResolveLvl(numId, ilvl int64) *wml.CT_Lvl {
	if err := n.ensureParsed(); err != nil {
		n.warn(fmt.Sprintf("numbering: parse failed: %v", err))
		return nil
	}

	if n.numbering == nil {
		n.warn(fmt.Sprintf("numbering: no numbering.xml part"))
		return nil
	}

	num := n.findNum(numId)
	if num == nil {
		n.warn(fmt.Sprintf("numbering: numId %d not found in numbering.xml", numId))
		return nil
	}

	if num.AbstractNumID == nil || num.AbstractNumID.Val == nil {
		n.warn(fmt.Sprintf("numbering: num %d has no abstractNumId", numId))
		return nil
	}

	abs := n.findAbstractNum(*num.AbstractNumID.Val)
	if abs == nil {
		n.warn(fmt.Sprintf("numbering: abstractNumId %d not found", *num.AbstractNumID.Val))
		return nil
	}

	lvl := n.findLvl(abs, ilvl)
	if lvl == nil {
		n.warn(fmt.Sprintf("numbering: ilvl %d not found in abstractNum %d", ilvl, *abs.AbstractNumID))
		return nil
	}

	// Apply lvlOverride (Pitfall 4)
	lvl = n.processOverride(num, ilvl, lvl)

	return lvl
}

// processOverride applies lvlOverride from the CT_Num on top of the
// abstractNum's lvl.  Returns the (possibly cloned-and-modified) lvl.
func (n *numberingCache) processOverride(num *wml.CT_Num, ilvl int64, lvl *wml.CT_Lvl) *wml.CT_Lvl {
	o := n.findOverride(num, ilvl)
	if o == nil {
		return lvl
	}

	// Full level replacement
	if o.Lvl != nil {
		return o.Lvl
	}

	// Start override only — clone the level, replace Start
	if o.StartOverride != nil && o.StartOverride.Val != nil {
		cloned := *lvl
		cloned.Start = &wml.CT_Start{Val: o.StartOverride.Val}
		return &cloned
	}

	return lvl
}

// ---------- Lookup helpers ----------

// findNum scans n.numbering.Num for a CT_Num with matching NumID.
func (n *numberingCache) findNum(numId int64) *wml.CT_Num {
	if n.numbering == nil {
		return nil
	}
	for _, nm := range n.numbering.Num {
		if nm.NumID != nil && *nm.NumID == numId {
			return nm
		}
	}
	return nil
}

// findAbstractNum scans n.numbering.AbstractNum for a CT_AbstractNum
// with matching AbstractNumID.
func (n *numberingCache) findAbstractNum(abstractNumId int64) *wml.CT_AbstractNum {
	if n.numbering == nil {
		return nil
	}
	for _, a := range n.numbering.AbstractNum {
		if a.AbstractNumID != nil && *a.AbstractNumID == abstractNumId {
			return a
		}
	}
	return nil
}

// findLvl scans abs.Lvl for a CT_Lvl with matching ILvl.
func (n *numberingCache) findLvl(abs *wml.CT_AbstractNum, ilvl int64) *wml.CT_Lvl {
	for _, l := range abs.Lvl {
		if l.ILvl != nil && *l.ILvl == ilvl {
			return l
		}
	}
	return nil
}

// findOverride scans num.LvlOverride for a CT_LvlOverride with matching ILvl.
func (n *numberingCache) findOverride(num *wml.CT_Num, ilvl int64) *wml.CT_LvlOverride {
	for _, o := range num.LvlOverride {
		if o.ILvl != nil && *o.ILvl == ilvl {
			return o
		}
	}
	return nil
}

// invalidateIfStale drops the parsed numbering tree when
// word/numbering.xml has been modified since the last parse.
// ensureParsed must be called after this to re-decode the new bytes.
//
// Follows the same dirty-flag pattern as 02-01's
// Resolver.invalidateCacheIfStale.
func (n *numberingCache) invalidateIfStale() {
	part, ok := n.pkg.Parts["word/numbering.xml"]
	if !ok {
		return // nothing to invalidate
	}
	if part.IsModified() {
		n.numbering = nil
		n.parsed = false
	}
}
