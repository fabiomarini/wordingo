// Package style resolves effective paragraph and run properties through
// the OOXML style inheritance chain: docDefaults → latentStyles fallback
// → basedOn chain → direct formatting.
//
// Phase boundary (plan 02-01 vs 02-02)
//   Plan 02-01: ThemeColor and NumPr are passed through UNCHANGED on the
//   resolved clone.  The chain walker, memo cache, merge and clone helpers
//   are untouched by 02-02.
//
//   Plan 02-02 added theme color concretization (ResolveRun →
//   theme.ResolveColor) and numbering level pPr merge (ResolveParagraph →
//   numbering.ResolveLvl + mergePPr) as post-processing steps.  CT_Shd
//   theme colors are NOT concretized (out of v1 scope).
//
// Decision traceability
//   D-01 clone output (ResolveParagraph / ResolveRun return fresh
//        clones, never the cached pointer)
//   D-03 memo by styleId + cache invalidation via Part.IsModified
//   D-04 single entry point (caller never touches cache directly)
//   D-05 cycle detection (visited-set, last-good, warning appended)
//   D-06 theme colors concretized at resolve-time (read-only over theme1.xml)
//   D-07 missing ref → warning + sensible default
//   D-12 latentStyles consulted ONLY when explicit styleId missing from
//        styles.xml; unstyled paragraphs skip it entirely
//   D-13 numbering resolution depth: numFmt, lvlText, start, level pPr
package style

import (
	"fmt"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// Resolver walks the OOXML basedOn chain for paragraph and run
// properties, returning deep-merged clones.
//
// Plan 02-02 added the theme and numbering caches (lazy-parse, warn
// callback threaded from appendWarning).  They are post-processors:
// ResolveRun → theme.ResolveColor (D-06); ResolveParagraph →
// numbering.ResolveLvl + mergePPr (D-13).
type Resolver struct {
	pkg *opc.Package

	styles       *wml.CT_Styles // lazily decoded styles tree; nil until first resolve
	stylesParsed bool           // guards re-parse when styles.xml has no styles element

	memoPPr map[string]*wml.CT_PPr // merged chain pPr per styleId (cached, NOT cloned)
	memoRPr map[string]*wml.CT_RPr // merged chain rPr per styleId

	warnings []string       // accumulator for D-05/D-07 warnings
	warn     func(string)   // bound to appendWarning; threaded into collectChain

	theme     *themeCache     // lazy theme1.xml color map (plan 02-02)
	numbering *numberingCache // lazy numbering.xml resolver (plan 02-02)
}

// NewResolver constructs a Resolver with empty memo maps and lazy
// caches for theme color and numbering resolution.
// styles.xml, theme1.xml, and numbering.xml are NOT parsed eagerly —
// parse is lazy on the first ResolveParagraph / ResolveRun call.
func NewResolver(pkg *opc.Package) *Resolver {
	r := &Resolver{
		pkg:     pkg,
		memoPPr: make(map[string]*wml.CT_PPr),
		memoRPr: make(map[string]*wml.CT_RPr),
	}
	r.warn = r.appendWarning
	r.theme = newThemeCache(pkg, r.appendWarning)
	r.numbering = newNumberingCache(pkg, r.appendWarning)
	return r
}

// Warnings returns every warning appended during the lifetime of the
// resolver (D-05 cycle, D-07 missing ref, D-12 latentStyles fallback).
func (r *Resolver) Warnings() []string { return r.warnings }

// ---------- Paragraph resolution ----------

// ResolveParagraph returns a deep-merged *wml.CT_PPr clone for p.
//
// Merge order: docDefaults.PPrDefault.PPr → basedOn chain pPr (root
// first, child last) → direct pPr on p (overrides chain).  ThemeColor
// and NumPr are passed through UNCHANGED — plan 02-02 post-processes
// theme color concretization and numbering level pPr merge.  Do not
// add that logic here.
//
// If p is nil or p.PPr is absent, returns nil, nil.
func (r *Resolver) ResolveParagraph(p *wml.CT_P) (*wml.CT_PPr, error) {
	if p == nil {
		return nil, nil
	}

	if err := r.ensureStylesParsed(); err != nil {
		return nil, fmt.Errorf("style: resolve paragraph: %w", err)
	}

	r.invalidateCacheIfStale()
	r.ensureStylesParsed() // re-parse if cache was just invalidated (one re-parse per entry)

	// Extract styleId — nil PStyle → unstyled path (D-12).
	var styleId string
	if p.PPr != nil && p.PPr.PStyle != nil && p.PPr.PStyle.Val != nil {
		styleId = *p.PPr.PStyle.Val
	}

	// Compute chain-merged base.
	var base *wml.CT_PPr
	if styleId == "" {
		// D-12: unstyled paragraph — docDefaults only; no latentStyles.
		base = &wml.CT_PPr{}
		if r.styles != nil && r.styles.DocDefaults != nil && r.styles.DocDefaults.PPrDefault != nil && r.styles.DocDefaults.PPrDefault.PPr != nil {
			base = clonePPr(r.styles.DocDefaults.PPrDefault.PPr)
		}
	} else {
		base = r.resolveChainPPr(styleId)
	}

	// Clone the base so the caller cannot corrupt the memo cache.
	result := clonePPr(base)

	// Merge direct pPr on top (D-04: direct formatting overrides chain).
	if p.PPr != nil {
		result = mergePPr(result, p.PPr)
	}

	// D-13: merge numbering level pPr into effective pPr (plan 02-02 post-process).
	if result.NumPr != nil && result.NumPr.NumId != nil && result.NumPr.NumId.Val != nil {
		numId := *result.NumPr.NumId.Val
		ilvl := int64(0)
		if result.NumPr.ILvl != nil && result.NumPr.ILvl.Val != nil {
			ilvl = *result.NumPr.ILvl.Val
		}

		r.numbering.invalidateIfStale()

		lvl := r.numbering.ResolveLvl(numId, ilvl)
		if lvl != nil && lvl.PPr != nil {
			result = mergePPr(result, lvl.PPr)
		}
		// D-07: if lvl is nil, the warning was already appended by
		// ResolveLvl.  NumPr is left intact on result.
	}

	return result, nil
}

// ---------- Run resolution ----------

// ResolveRun returns a deep-merged *wml.CT_RPr clone for run r inside
// paragraph p.
//
// Merge order: docDefaults.RPrDefault.RPr → paragraph-style-chain rPr
// (each style's RPr merged in chain order, root first, child last) →
// rStyle-chain rPr (if run has an explicit rStyle) → run direct rPr
// (overrides everything).  ThemeColor fields are passed through
// UNCHANGED — plan 02-02 concretizes.
//
// If r is nil, returns nil, nil.
func (r *Resolver) ResolveRun(p *wml.CT_P, run *wml.CT_R) (*wml.CT_RPr, error) {
	if run == nil {
		return nil, nil
	}

	if err := r.ensureStylesParsed(); err != nil {
		return nil, fmt.Errorf("style: resolve run: %w", err)
	}

	r.invalidateCacheIfStale()
	r.ensureStylesParsed()

	// Extract paragraph styleId.
	var styleId string
	if p != nil && p.PPr != nil && p.PPr.PStyle != nil && p.PPr.PStyle.Val != nil {
		styleId = *p.PPr.PStyle.Val
	}

	// Base: docDefaults rPr.
	base := &wml.CT_RPr{}
	if r.styles != nil && r.styles.DocDefaults != nil && r.styles.DocDefaults.RPrDefault != nil && r.styles.DocDefaults.RPrDefault.RPr != nil {
		base = cloneRPr(r.styles.DocDefaults.RPrDefault.RPr)
	}

	// Merge paragraph-style-chain rPr (if the paragraph has a style).
	if styleId != "" {
		chainRPr := r.resolveChainRPr(styleId)
		base = mergeRPr(base, chainRPr)
	}

	// Clone so far to separate from memo cache.
	result := cloneRPr(base)

	// Merge rStyle-chain rPr (run's explicit character style).
	if run.RPr != nil && run.RPr.RStyle != nil && run.RPr.RStyle.Val != nil {
		rStyleRPr := r.resolveChainRPr(*run.RPr.RStyle.Val)
		result = mergeRPr(result, rStyleRPr)
	}

	// Merge run direct rPr on top (highest priority).
	if run.RPr != nil {
		result = mergeRPr(result, run.RPr)
	}

	// D-06: concretize theme color at resolve-time (plan 02-02 post-process).
	if result.Color != nil {
		result.Color = r.theme.ResolveColor(result.Color)
	}

	return result, nil
}

// ---------- Chain walkers (unexported) ----------

// resolveChainPPr returns the memo-cached (or freshly computed)
// merged pPr for styleId.  Caller MUST clone before returning.
func (r *Resolver) resolveChainPPr(styleId string) *wml.CT_PPr {
	if cached, ok := r.memoPPr[styleId]; ok {
		return cached
	}

	result := &wml.CT_PPr{}

	// Always start with docDefaults as the base layer (D-12).
	if r.styles != nil && r.styles.DocDefaults != nil && r.styles.DocDefaults.PPrDefault != nil && r.styles.DocDefaults.PPrDefault.PPr != nil {
		result = mergePPr(result, r.styles.DocDefaults.PPrDefault.PPr)
	}

	// Walk chain child-first: [styleId, parent, grandparent, ...].
	chain := r.collectChain(styleId)
	// Apply in reverse (root first, child last) so child overrides.
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].PPr != nil {
			result = mergePPr(result, chain[i].PPr)
		}
	}

	r.memoPPr[styleId] = result
	return result
}

// resolveChainRPr returns the memo-cached (or freshly computed)
// merged rPr for styleId.  Caller MUST clone before returning.
func (r *Resolver) resolveChainRPr(styleId string) *wml.CT_RPr {
	if cached, ok := r.memoRPr[styleId]; ok {
		return cached
	}

	result := &wml.CT_RPr{}

	if r.styles != nil && r.styles.DocDefaults != nil && r.styles.DocDefaults.RPrDefault != nil && r.styles.DocDefaults.RPrDefault.RPr != nil {
		result = mergeRPr(result, r.styles.DocDefaults.RPrDefault.RPr)
	}

	chain := r.collectChain(styleId)
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].RPr != nil {
			result = mergeRPr(result, chain[i].RPr)
		}
	}

	r.memoRPr[styleId] = result
	return result
}

// collectChain walks the basedOn chain from styleId, returning styles
// child-first: [styleId, parent, grandparent, ...].  NEVER reads Next
// (Pitfall 1) or Link (deferred to 02-02 per phase_boundary).
func (r *Resolver) collectChain(styleId string) []*wml.CT_Style {
	visited := make(map[string]bool)
	var chain []*wml.CT_Style

	cursor := styleId
	for cursor != "" {
		if visited[cursor] {
			// D-05: circular basedOn — warn and break; keep last-good.
			r.warn(fmt.Sprintf("style: circular basedOn chain at %q", cursor))
			break
		}

		s := r.findStyle(cursor)
		if s == nil {
			// D-12: latentStyles fallback only when styleId missing from styles.xml.
			if ls := r.findLatent(cursor); ls != nil {
				r.warn(fmt.Sprintf("style: %q not instantiated; using latentStyles fallback", cursor))
				// Latent lsdException carries no pPr/rPr; docDefaults
				// base already applied.  Break.
			} else {
				// D-07: missing styleId — not in styles.xml or latentStyles.
				r.warn(fmt.Sprintf("style: %q not found in styles.xml or latentStyles", cursor))
			}
			break
		}

		visited[cursor] = true
		chain = append(chain, s)

		if s.BasedOn != nil && s.BasedOn.Val != nil {
			cursor = *s.BasedOn.Val
		} else {
			cursor = ""
		}
	}

	return chain
}

// ---------- Lookup helpers ----------

// findStyle linearly scans r.styles.Style for a style with matching
// StyleID.  Returns nil if r.styles is nil or not found.
func (r *Resolver) findStyle(styleId string) *wml.CT_Style {
	if r.styles == nil {
		return nil
	}
	for _, s := range r.styles.Style {
		if s.StyleID != nil && *s.StyleID == styleId {
			return s
		}
	}
	return nil
}

// findLatent scans r.styles.LatentStyles.LsdException for a matching
// Name.  Returns nil if r.styles, LatentStyles, or no match.
func (r *Resolver) findLatent(name string) *wml.CT_LsdException {
	if r.styles == nil || r.styles.LatentStyles == nil {
		return nil
	}
	for _, e := range r.styles.LatentStyles.LsdException {
		if e.Name != nil && *e.Name == name {
			return e
		}
	}
	return nil
}

// ---------- Lazy parse and cache invalidation ----------

// ensureStylesParsed lazily decodes word/styles.xml into r.styles.
// If the part is absent, r.styles is set to an empty CT_Styles so
// every subsequent basedOn lookup warns "not found" (D-07 resilient).
func (r *Resolver) ensureStylesParsed() error {
	if r.stylesParsed {
		return nil
	}

	part, ok := r.pkg.Parts["word/styles.xml"]
	if !ok {
		// No styles part → empty tree; every resolve will warn and
		// use docDefaults (D-07 resilient).
		r.styles = &wml.CT_Styles{}
		r.stylesParsed = true
		return nil
	}

	rc, err := part.Open()
	if err != nil {
		return fmt.Errorf("style: open styles.xml: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)
	var s wml.CT_Styles
	if err := dec.Decode(&s); err != nil {
		return fmt.Errorf("style: decode styles.xml: %v: %w", err, ErrStylesParseFailed)
	}

	r.styles = &s
	r.stylesParsed = true
	return nil
}

// invalidateCacheIfStale drops both memo maps when word/styles.xml has
// been modified since the last parse.  ensureStylesParsed must be
// called after this to re-decode the new bytes.
func (r *Resolver) invalidateCacheIfStale() {
	part, ok := r.pkg.Parts["word/styles.xml"]
	if !ok {
		return // nothing to invalidate
	}
	if part.IsModified() {
		r.memoPPr = nil
		r.memoRPr = nil
		r.memoPPr = make(map[string]*wml.CT_PPr)
		r.memoRPr = make(map[string]*wml.CT_RPr)
		r.stylesParsed = false
		r.styles = nil
	}
}

// appendWarning adds a message to the warnings accumulator.
func (r *Resolver) appendWarning(msg string) {
	r.warnings = append(r.warnings, msg)
}

// ---------- Merge helpers (pure functions) ----------

// mergePPr merges src paragraph properties into dst (shallow override
// by default, deep-merge for Spacing and Ind).
//
// Returns dst for chaining.  dst may be nil (initialised to empty).
// src may be nil (returns dst unchanged).
func mergePPr(dst, src *wml.CT_PPr) *wml.CT_PPr {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &wml.CT_PPr{}
	}

	if src.PStyle != nil {
		v := *src.PStyle.Val
		dst.PStyle = &wml.CT_PStyle{Val: &v}
	}
	if src.KeepNext != nil {
		dst.KeepNext = cloneOnOff(src.KeepNext)
	}
	if src.KeepLines != nil {
		dst.KeepLines = cloneOnOff(src.KeepLines)
	}
	if src.PageBreakBefore != nil {
		dst.PageBreakBefore = cloneOnOff(src.PageBreakBefore)
	}
	if src.WidowControl != nil {
		dst.WidowControl = cloneOnOff(src.WidowControl)
	}
	if src.NumPr != nil {
		dst.NumPr = cloneNumPr(src.NumPr)
	}
	if src.Spacing != nil {
		dst.Spacing = mergeSpacing(dst.Spacing, src.Spacing)
	}
	if src.Ind != nil {
		dst.Ind = mergeInd(dst.Ind, src.Ind)
	}
	if src.Jc != nil {
		dst.Jc = &wml.CT_Jc{Val: src.Jc.Val}
	}
	if src.Tabs != nil {
		dst.Tabs = cloneTabs(src.Tabs)
	}
	if src.SectPr != nil {
		// SectPr is a complex struct; only pointer-copy at top level
		// (shallow override — caller responsible for not sharing).
		s := *src.SectPr
		dst.SectPr = &s
	}
	if src.Shd != nil {
		s := *src.Shd
		dst.Shd = &s
	}
	if src.OutlineLvl != nil {
		dst.OutlineLvl = &wml.CT_OutlineLvl{Val: src.OutlineLvl.Val}
	}
	return dst
}

// mergeRPr merges src run properties into dst (shallow override by
// default, deep-merge for RFonts and Color).
//
// Returns dst for chaining.  dst may be nil.  src may be nil.
func mergeRPr(dst, src *wml.CT_RPr) *wml.CT_RPr {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &wml.CT_RPr{}
	}

	if src.RStyle != nil {
		dst.RStyle = &wml.CT_RStyle{Val: src.RStyle.Val}
	}
	if src.RFonts != nil {
		dst.RFonts = mergeRFonts(dst.RFonts, src.RFonts)
	}
	if src.B != nil {
		dst.B = cloneOnOff(src.B)
	}
	if src.I != nil {
		dst.I = cloneOnOff(src.I)
	}
	if src.U != nil {
		dst.U = &wml.CT_U{Val: src.U.Val, Color: src.U.Color}
	}
	if src.Sz != nil {
		dst.Sz = &wml.CT_Sz{Val: src.Sz.Val}
	}
	if src.SzCs != nil {
		dst.SzCs = &wml.CT_Sz{Val: src.SzCs.Val}
	}
	if src.Color != nil {
		dst.Color = mergeColor(dst.Color, src.Color)
	}
	if src.Highlight != nil {
		dst.Highlight = &wml.CT_Highlight{Val: src.Highlight.Val}
	}
	if src.VertAlign != nil {
		dst.VertAlign = &wml.CT_VertAlign{Val: src.VertAlign.Val}
	}
	if src.Lang != nil {
		dst.Lang = &wml.CT_Lang{Val: src.Lang.Val, EastAsia: src.Lang.EastAsia}
	}
	if src.Strike != nil {
		dst.Strike = cloneOnOff(src.Strike)
	}
	if src.DStrike != nil {
		dst.DStrike = cloneOnOff(src.DStrike)
	}
	if src.Vanish != nil {
		dst.Vanish = cloneOnOff(src.Vanish)
	}
	if src.Kern != nil {
		dst.Kern = &wml.CT_Kern{Val: src.Kern.Val}
	}
	if src.Shd != nil {
		s := *src.Shd
		dst.Shd = &s
	}
	if src.SmallCaps != nil {
		dst.SmallCaps = cloneOnOff(src.SmallCaps)
	}
	if src.Caps != nil {
		dst.Caps = cloneOnOff(src.Caps)
	}
	return dst
}

// mergeSpacing deep-merges src spacing into dst (per-attribute
// independence per ISO §17.7.2).
func mergeSpacing(dst, src *wml.CT_Spacing) *wml.CT_Spacing {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &wml.CT_Spacing{}
	}
	if src.Before != nil {
		dst.Before = src.Before
	}
	if src.After != nil {
		dst.After = src.After
	}
	if src.Line != nil {
		dst.Line = src.Line
	}
	if src.LineRule != nil {
		dst.LineRule = src.LineRule
	}
	if src.BeforeLines != nil {
		dst.BeforeLines = src.BeforeLines
	}
	if src.AfterLines != nil {
		dst.AfterLines = src.AfterLines
	}
	return dst
}

// mergeInd deep-merges src indentation into dst.
func mergeInd(dst, src *wml.CT_Ind) *wml.CT_Ind {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &wml.CT_Ind{}
	}
	if src.Left != nil {
		dst.Left = src.Left
	}
	if src.Right != nil {
		dst.Right = src.Right
	}
	if src.FirstLine != nil {
		dst.FirstLine = src.FirstLine
	}
	if src.Hanging != nil {
		dst.Hanging = src.Hanging
	}
	return dst
}

// mergeRFonts deep-merges src run fonts into dst.
func mergeRFonts(dst, src *wml.CT_RFonts) *wml.CT_RFonts {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &wml.CT_RFonts{}
	}
	if src.Ascii != nil {
		dst.Ascii = src.Ascii
	}
	if src.HAnsi != nil {
		dst.HAnsi = src.HAnsi
	}
	if src.EastAsia != nil {
		dst.EastAsia = src.EastAsia
	}
	if src.CS != nil {
		dst.CS = src.CS
	}
	if src.AsciiTheme != nil {
		dst.AsciiTheme = src.AsciiTheme
	}
	if src.HAnsiTheme != nil {
		dst.HAnsiTheme = src.HAnsiTheme
	}
	return dst
}

// mergeColor deep-merges src color into dst (per-attribute
// independence).  ThemeColor is NOT cleared — plan 02-02
// concretizes.  This ensures child.Val overrides parent.Val while
// parent.ThemeColor survives if child does not set it.
func mergeColor(dst, src *wml.CT_Color) *wml.CT_Color {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &wml.CT_Color{}
	}
	if src.Val != nil {
		dst.Val = src.Val
	}
	if src.ThemeColor != nil {
		dst.ThemeColor = src.ThemeColor
	}
	if src.ThemeShade != nil {
		dst.ThemeShade = src.ThemeShade
	}
	if src.ThemeTint != nil {
		dst.ThemeTint = src.ThemeTint
	}
	return dst
}

// ---------- Clone helpers ----------

// clonePPr returns a deep copy of p via mergePPr into an empty CT_PPr.
func clonePPr(p *wml.CT_PPr) *wml.CT_PPr {
	return mergePPr(&wml.CT_PPr{}, p)
}

// cloneRPr returns a deep copy of r via mergeRPr into an empty CT_RPr.
func cloneRPr(r *wml.CT_RPr) *wml.CT_RPr {
	return mergeRPr(&wml.CT_RPr{}, r)
}

// cloneOnOff deep-copies a CT_OnOff value (including the *bool).
func cloneOnOff(src *wml.CT_OnOff) *wml.CT_OnOff {
	dst := &wml.CT_OnOff{}
	if src.Val != nil {
		v := *src.Val
		dst.Val = &v
	}
	return dst
}

// cloneNumPr returns a shallow copy of src NumPr.
func cloneNumPr(src *wml.CT_NumPr) *wml.CT_NumPr {
	if src == nil {
		return nil
	}
	dst := &wml.CT_NumPr{}
	if src.ILvl != nil {
		dst.ILvl = &wml.CT_ILvl{Val: src.ILvl.Val}
	}
	if src.NumId != nil {
		dst.NumId = &wml.CT_NumId{Val: src.NumId.Val}
	}
	return dst
}

// cloneTabs returns a shallow copy of src Tabs.
func cloneTabs(src *wml.CT_Tabs) *wml.CT_Tabs {
	if src == nil {
		return nil
	}
	dst := &wml.CT_Tabs{Tab: make([]*wml.CT_TabStop, len(src.Tab))}
	for i, t := range src.Tab {
		tt := *t
		dst.Tab[i] = &tt
	}
	return dst
}


