// Package style — theme color concretization.
//
// D-06: Theme colors resolve to concrete hex at resolve-time.  The
// returned CT_Color carries explicit hex, not the theme placeholder.
// The raw theme1.xml part is never mutated — resolution is a read-only
// view over the cached color map.
//
// D-07: Missing theme color ref (unknown enum, missing element, missing
// theme1.xml part, malformed XML) → warning + return color as-is.
// No sentinel errors for resolver misses — problems surface via warnings.
//
// Pitfall 3 (enum ≠ element name): themeColor enum values (dark1, light1,
// hyperlink, followedHyperlink) do NOT match the theme1.xml element names
// (dk1, lt1, hlink, folHlink).  The themeEnumToElement map handles this.
//
// Pitfall 5 (tint/shade order): when both themeTint and themeShade are
// present, shade is applied first, then tint (Word's observed order).
//
// sysClr lastClr handling: dark1/light1 use <a:sysClr val="windowText"
// lastClr="000000"/> — the actual hex is in lastClr, not val (val is a
// system color name like "windowText").  srgbClr elements use val="HHHHHH".
//
// Boundary: CT_Shd theme colors are NOT concretized in this plan
// (out of v1 scope).
package style

import (
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// themeEnumToElement maps OOXML themeColor enum values to theme1.xml
// a:clrScheme child element local names (ISO §17.18.96 ST_ThemeColor).
//
// The enum values differ from the element names for dark1/light1/dark2/light2
// (dk1/lt1/dk2/lt2) and hyperlink/followedHyperlink (hlink/folHlink).
// The accent1-6 names match directly.
//
// Source: defaults/theme1.xml element names verified below (Assumption A8).
var themeEnumToElement = map[string]string{
	"dark1":             "dk1",
	"light1":            "lt1",
	"dark2":             "dk2",
	"light2":            "lt2",
	"accent1":           "accent1",
	"accent2":           "accent2",
	"accent3":           "accent3",
	"accent4":           "accent4",
	"accent5":           "accent5",
	"accent6":           "accent6",
	"hyperlink":         "hlink",
	"followedHyperlink": "folHlink",
}

// themeCache lazily parses theme1.xml a:clrScheme children and provides
// theme color resolution.  Read-only over the source part (D-06).
type themeCache struct {
	colors  map[string]string // key = element local name ("dk1"), value = 6-char hex ("000000")
	parsed  bool              // guards re-parse; set true BEFORE parse to prevent retry loops
	pkg     *opc.Package      // source for lazy read
	warn    func(string)      // warning callback threaded from Resolver.appendWarning
}

// newThemeCache constructs a themeCache bound to the given package and
// warning callback.  Does NOT parse eagerly — lazy on first ResolveColor call.
func newThemeCache(pkg *opc.Package, warn func(string)) *themeCache {
	return &themeCache{
		colors: nil,
		parsed: false,
		pkg:    pkg,
		warn:   warn,
	}
}

// ensureParsed lazily decodes theme1.xml a:clrScheme children into
// t.colors.  Idempotent — second call is a no-op (parsed guard).
// Sets parsed=true BEFORE parse so a failure does not retry indefinitely
// (the cache stays empty and every ResolveColor warns "not found" — D-07).
func (t *themeCache) ensureParsed() error {
	if t.parsed {
		return nil
	}
	t.parsed = true // prevent retry loops on parse failure

	part, ok := t.pkg.Parts["word/theme/theme1.xml"]
	if !ok {
		return nil // no theme part → empty cache; ResolveColor will warn per D-07
	}

	rc, err := part.Open()
	if err != nil {
		return fmt.Errorf("style: open theme1.xml: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)

	t.colors = make(map[string]string, 12) // exactly 12 clrScheme children per ISO

	// Token-scan through the theme1.xml stream looking for a:clrScheme children.
	// We do NOT decode into a full CT_Theme (DrawingML — out of scope; anti-pattern).
	inClrScheme := false
	for {
		tok, err := dec.Token()
		if err != nil {
			break // end of stream or error — either way, we have what we have
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "clrScheme" && tok.Name.Space == "http://schemas.openxmlformats.org/drawingml/2006/main" {
				inClrScheme = true
				continue
			}
			if !inClrScheme {
				continue
			}
			elemLocal := tok.Name.Local
			// We only care about the 12 scheme color elements
			if _, ok := themeEnumToElement[reverseLookup(elemLocal)]; !ok && !isSchemeColorElem(elemLocal) {
				continue
			}

			// Read the next token — should be srgbClr or sysClr
			valTok, err := dec.Token()
			if err != nil {
				break
			}
			childStart, ok := valTok.(xml.StartElement)
			if !ok {
				continue
			}

			var hex string
			if childStart.Name.Local == "srgbClr" {
				// <a:srgbClr val="HHHHHH"/>
				for _, attr := range childStart.Attr {
					if attr.Name.Local == "val" {
						hex = normalizeHex(attr.Value)
						break
					}
				}
			} else if childStart.Name.Local == "sysClr" {
				// <a:sysClr val="windowText" lastClr="000000"/>
				for _, attr := range childStart.Attr {
					if attr.Name.Local == "lastClr" {
						hex = normalizeHex(attr.Value)
						break
					}
				}
			}

			if hex != "" {
				t.colors[elemLocal] = hex
			}

		case xml.EndElement:
			if inClrScheme && tok.Name.Local == "clrScheme" {
				inClrScheme = false
			}
		}
	}

	return nil
}

// ResolveColor concretizes a CT_Color's ThemeColor reference to concrete
// hex, returning a fresh clone.  Returns nil on nil input (safe).
//
// Algorithm:
//  1. Shallow-copy input
//  2. If ThemeColor is nil or empty → return copy as-is
//  3. ensureParsed() — if error → warn, return copy as-is (D-07)
//  4. Lookup ThemeColor in themeEnumToElement map — if unknown → warn, return (D-07)
//  5. Lookup element name in colors map — if not found → warn, return (D-07)
//  6. parseHexRGB → apply Shade (Pitfall 5: shade FIRST) → apply Tint (after shade)
//  7. Set Val to formatted hex, nil-out ThemeColor/ThemeShade/ThemeTint
//  8. Return the clone
func (t *themeCache) ResolveColor(c *wml.CT_Color) *wml.CT_Color {
	if c == nil {
		return nil
	}

	out := *c // shallow copy — caller gets a fresh CT_Color, not the input pointer

	if out.ThemeColor == nil || *out.ThemeColor == "" {
		return &out
	}

	if err := t.ensureParsed(); err != nil {
		t.warn(fmt.Sprintf("theme: parse failed: %v", err))
		return &out
	}

	// Pitfall 3: map enum to element name
	elemName, ok := themeEnumToElement[*out.ThemeColor]
	if !ok {
		t.warn(fmt.Sprintf("theme: unknown themeColor %q", *out.ThemeColor))
		return &out
	}

	hex, ok := t.colors[elemName]
	if !ok {
		// Two cases: no theme part (colors is nil) or element missing from clrScheme
		if t.colors == nil {
			t.warn(fmt.Sprintf("theme: no theme1.xml part"))
		} else {
			t.warn(fmt.Sprintf("theme: %q not found in theme1.xml", elemName))
		}
		return &out
	}

	r, g, b := parseHexRGB(hex)

	// Pitfall 5: shade FIRST, then tint
	if out.ThemeShade != nil && *out.ThemeShade != "" {
		r, g, b = applyShade(r, g, b, *out.ThemeShade)
	}
	if out.ThemeTint != nil && *out.ThemeTint != "" {
		r, g, b = applyTint(r, g, b, *out.ThemeTint)
	}

	v := fmt.Sprintf("%02X%02X%02X", r, g, b)
	out.Val = &v
	out.ThemeColor = nil
	out.ThemeShade = nil
	out.ThemeTint = nil

	return &out
}

// ---------- Color math helpers ----------

// parseHexRGB parses a hex color string (6 hex digits, optionally prefixed
// with "#" or "0x") into per-channel ints.  Returns 0,0,0 on parse error.
func parseHexRGB(hex string) (r, g, b int) {
	h := strings.TrimPrefix(hex, "#")
	h = strings.TrimPrefix(h, "0x")
	h = strings.TrimPrefix(h, "0X")
	for len(h) < 6 {
		h = "0" + h
	}
	if len(h) > 6 {
		h = h[:6]
	}

	rv, err := strconv.ParseInt(h[0:2], 16, 0)
	if err != nil {
		return 0, 0, 0
	}
	gv, err := strconv.ParseInt(h[2:4], 16, 0)
	if err != nil {
		return 0, 0, 0
	}
	bv, err := strconv.ParseInt(h[4:6], 16, 0)
	if err != nil {
		return 0, 0, 0
	}
	return int(rv), int(gv), int(bv)
}

// applyTint lightens each channel toward white.
// Formula: new = orig + (255 - orig) * (tintVal / 255)
func applyTint(r, g, b int, tintHex string) (int, int, int) {
	tv, err := strconv.ParseInt(tintHex, 16, 0)
	if err != nil {
		return r, g, b
	}
	factor := float64(tv) / 255.0
	nr := clampByte(float64(r) + (255.0-float64(r))*factor)
	ng := clampByte(float64(g) + (255.0-float64(g))*factor)
	nb := clampByte(float64(b) + (255.0-float64(b))*factor)
	return nr, ng, nb
}

// applyShade darkens each channel toward black.
// Formula: new = orig * (shadeVal / 255)
func applyShade(r, g, b int, shadeHex string) (int, int, int) {
	sv, err := strconv.ParseInt(shadeHex, 16, 0)
	if err != nil {
		return r, g, b
	}
	factor := float64(sv) / 255.0
	nr := clampByte(float64(r) * factor)
	ng := clampByte(float64(g) * factor)
	nb := clampByte(float64(b) * factor)
	return nr, ng, nb
}

// clampByte rounds to nearest and clamps to 0-255.
func clampByte(v float64) int {
	iv := int(math.Round(v))
	if iv < 0 {
		return 0
	}
	if iv > 255 {
		return 255
	}
	return iv
}

// normalizeHex normalizes a hex color string: uppercase, no prefix, 6 digits.
func normalizeHex(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimPrefix(s, "0X")
	for len(s) < 6 {
		s = "0" + s
	}
	return s
}

// reverseLookup is a helper for the token-scan: given an element local name,
// find the themeColor enum that maps to it.
func reverseLookup(elemLocal string) string {
	for k, v := range themeEnumToElement {
		if v == elemLocal {
			return k
		}
	}
	return ""
}

// isSchemeColorElem returns true if local is one of the 12 a:clrScheme
// child element names.
func isSchemeColorElem(local string) bool {
	switch local {
	case "dk1", "lt1", "dk2", "lt2",
		"accent1", "accent2", "accent3", "accent4", "accent5", "accent6",
		"hlink", "folHlink":
		return true
	}
	return false
}
