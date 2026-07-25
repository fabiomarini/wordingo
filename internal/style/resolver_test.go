package style

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
)

// ---------- Synthetic fixture builders ----------

const stylesNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

// buildStylesXML wraps styleBody in a standard styles.xml envelope.
func buildStylesXML(t *testing.T, styleBody string) []byte {
	t.Helper()
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="%s">
%s
</w:styles>`, stylesNS, styleBody)
	return []byte(xml)
}

// buildPkgWithStyles constructs an in-memory *opc.Package containing the
// given styles XML and a minimal document.  Reuses the buildSyntheticZip
// pattern from opc_test.go (copy, not export).
func buildPkgWithStyles(t *testing.T, stylesXML []byte) *opc.Package {
	t.Helper()

	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`

	rootRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	docRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

	document := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="%s"><w:body><w:p/></w:body></w:document>`
	document = fmt.Sprintf(document, stylesNS)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name string, payload []byte) {
		t.Helper()
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.Modified = time.Date(2020, 3, 4, 5, 6, 7, 0, time.UTC)
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	add("[Content_Types].xml", []byte(contentTypes))
	add("_rels/.rels", []byte(rootRels))
	add("word/document.xml", []byte(document))
	add("word/_rels/document.xml.rels", []byte(docRels))
	add("word/styles.xml", stylesXML)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	pkg, err := opc.Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return pkg
}

// ---------- Assertion helpers ----------

// assertPPrSpacing asserts a spacing field value on a CT_PPr.
func assertPPrSpacing(t *testing.T, ppr *wml.CT_PPr, field string, want int64) {
	t.Helper()
	if ppr == nil || ppr.Spacing == nil {
		t.Fatalf("ppr.Spacing is nil")
	}
	var got *int64
	switch field {
	case "Before":
		got = ppr.Spacing.Before
	case "After":
		got = ppr.Spacing.After
	case "Line":
		got = ppr.Spacing.Line
	default:
		t.Fatalf("unknown spacing field: %s", field)
	}
	if got == nil {
		t.Fatalf("ppr.Spacing.%s is nil, want %d", field, want)
	}
	if *got != want {
		t.Fatalf("ppr.Spacing.%s = %d, want %d", field, *got, want)
	}
}

// assertPPrSpacingNil asserts a spacing field is nil on a CT_PPr.
func assertPPrSpacingNil(t *testing.T, ppr *wml.CT_PPr, field string) {
	t.Helper()
	if ppr == nil || ppr.Spacing == nil {
		return // nil spacing means all fields are nil
	}
	var got *int64
	switch field {
	case "Before":
		got = ppr.Spacing.Before
	case "After":
		got = ppr.Spacing.After
	default:
		t.Fatalf("unknown spacing field: %s", field)
	}
	if got != nil {
		t.Fatalf("ppr.Spacing.%s = %d, want nil", field, *got)
	}
}

// assertRPrSz asserts the sz.val on a CT_RPr.
func assertRPrSz(t *testing.T, rpr *wml.CT_RPr, want int64) {
	t.Helper()
	if rpr == nil || rpr.Sz == nil || rpr.Sz.Val == nil {
		t.Fatalf("rpr.Sz.Val is nil, want %d", want)
	}
	if *rpr.Sz.Val != want {
		t.Fatalf("rpr.Sz.Val = %d, want %d", *rpr.Sz.Val, want)
	}
}

// assertBoolField asserts a *CT_OnOff field is non-nil on a CT_PPr or CT_RPr.
// For CT_PPr pass the whole ppr; for CT_RPr pass the whole rpr.
func assertBoolField(t *testing.T, name string, field *wml.CT_OnOff) {
	t.Helper()
	if field == nil {
		t.Fatalf("%s: field is nil, want non-nil (present)", name)
	}
}

// assertStringFieldVal asserts a *string field pointer is non-nil and equals want.
func assertStringFieldVal(t *testing.T, name string, field *string, want string) {
	t.Helper()
	if field == nil {
		t.Fatalf("%s: field is nil, want %q", name, want)
	}
	if *field != want {
		t.Fatalf("%s = %q, want %q", name, *field, want)
	}
}

// assertWarningContains checks that at least one warning contains substr.
func assertWarningContains(t *testing.T, warnings []string, substr string) {
	t.Helper()
	for _, w := range warnings {
		if strings.Contains(w, substr) {
			return
		}
	}
	t.Fatalf("no warning contains %q; got %v", substr, warnings)
}

// assertNoWarning checks that no warning contains substr.
func assertNoWarning(t *testing.T, warnings []string, substr string) {
	t.Helper()
	for _, w := range warnings {
		if strings.Contains(w, substr) {
			t.Fatalf("unexpected warning %q contains %q", w, substr)
		}
	}
}

// buildEmptyDocDefaults returns a styles body with docDefaults only (no explicit styles).
func buildEmptyDocDefaults(ddBody string) string {
	return fmt.Sprintf(`<w:docDefaults>%s</w:docDefaults>`, ddBody)
}

// ---------- Test: docDefaults-only paragraph (unstyled) ----------

func TestResolveParagraph_DocDefaultsOnly(t *testing.T) {
	stylesBody := buildEmptyDocDefaults(`<w:pPrDefault><w:pPr><w:spacing w:after="160"/></w:pPr></w:pPrDefault>`)
	stylesXML := buildStylesXML(t, stylesBody)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{}} // no PStyle → docDefaults only
	ppr, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if ppr == nil {
		t.Fatal("ppr is nil")
	}
	assertPPrSpacing(t, ppr, "After", 160)
	if len(resolver.Warnings()) != 0 {
		t.Fatalf("expected no warnings, got %v", resolver.Warnings())
	}
}

// ---------- Test: Heading2 chain ----------

const headingChainStyles = `
<w:docDefaults>
  <w:pPrDefault><w:pPr><w:spacing w:after="160"/></w:pPr></w:pPrDefault>
  <w:rPrDefault><w:rPr><w:rFonts w:ascii="Aptos"/><w:sz w:val="22"/></w:rPr></w:rPrDefault>
</w:docDefaults>
<w:style w:type="paragraph" w:styleId="Normal">
  <w:name w:val="Normal"/>
  <w:rPr><w:rFonts w:asciiTheme="minorHAnsi"/><w:sz w:val="22"/></w:rPr>
</w:style>
<w:style w:type="paragraph" w:styleId="Heading1">
  <w:name w:val="heading 1"/>
  <w:basedOn w:val="Normal"/>
  <w:pPr><w:keepNext/><w:spacing w:before="480" w:after="240"/></w:pPr>
  <w:rPr><w:b/><w:sz w:val="32"/><w:color w:themeColor="accent1"/></w:rPr>
</w:style>
<w:style w:type="paragraph" w:styleId="Heading2">
  <w:name w:val="heading 2"/>
  <w:basedOn w:val="Heading1"/>
  <w:pPr><w:spacing w:before="240" w:after="120"/><w:outlineLvl w:val="1"/></w:pPr>
  <w:rPr><w:sz w:val="28"/></w:rPr>
</w:style>`

func TestResolveParagraph_Heading2Chain(t *testing.T) {
	stylesXML := buildStylesXML(t, headingChainStyles)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("Heading2")},
	}}
	ppr, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if ppr == nil {
		t.Fatal("ppr is nil")
	}

	// Heading2's own before overrides Heading1 and docDefaults.
	assertPPrSpacing(t, ppr, "Before", 240)
	// Heading2's own after overrides Heading1; docDefaults after=160 is
	// overridden by Heading1's after=240, then Heading2's after=120.
	assertPPrSpacing(t, ppr, "After", 120)
	// KeepNext inherited from Heading1 (Heading2 does not set it).
	assertBoolField(t, "KeepNext", ppr.KeepNext)
	// OutlineLvl from Heading2.
	if ppr.OutlineLvl == nil || ppr.OutlineLvl.Val == nil || *ppr.OutlineLvl.Val != 1 {
		t.Fatalf("OutlineLvl = %+v, want Val=1", ppr.OutlineLvl)
	}

	if len(resolver.Warnings()) != 0 {
		t.Fatalf("expected no warnings, got %v", resolver.Warnings())
	}
}

func TestResolveRun_Heading2Chain(t *testing.T) {
	stylesXML := buildStylesXML(t, headingChainStyles)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("Heading2")},
	}}
	run := &wml.CT_R{} // no explicit rPr

	rpr, err := resolver.ResolveRun(p, run)
	if err != nil {
		t.Fatalf("ResolveRun: %v", err)
	}
	if rpr == nil {
		t.Fatal("rpr is nil")
	}

	// Sz=28 from Heading2.
	assertRPrSz(t, rpr, 28)
	// B=true inherited from Heading1.
	assertBoolField(t, "B", rpr.B)
	// RFonts.Ascii="Aptos" from docDefaults (not overridden in chain).
	assertStringFieldVal(t, "rFonts.Ascii", rpr.RFonts.Ascii, "Aptos")
	// RFonts.AsciiTheme="minorHAnsi" from Normal (inherited through chain).
	assertStringFieldVal(t, "rFonts.AsciiTheme", rpr.RFonts.AsciiTheme, "minorHAnsi")

	if len(resolver.Warnings()) != 0 {
		t.Fatalf("expected no warnings, got %v", resolver.Warnings())
	}
}

// ---------- Test: circular basedOn ----------

func TestResolveParagraph_CircularBasedOn(t *testing.T) {
	stylesBody := `
<w:style w:type="paragraph" w:styleId="CycleA">
  <w:name w:val="CycleA"/>
  <w:basedOn w:val="CycleB"/>
  <w:pPr><w:spacing w:before="100"/></w:pPr>
</w:style>
<w:style w:type="paragraph" w:styleId="CycleB">
  <w:name w:val="CycleB"/>
  <w:basedOn w:val="CycleA"/>
  <w:pPr><w:spacing w:before="200"/></w:pPr>
</w:style>`
	stylesXML := buildStylesXML(t, stylesBody)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("CycleA")},
	}}

	var ppr *wml.CT_PPr
	var err error

	// No-panic guard (circular must not stack-overflow).
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("ResolveParagraph panicked on circular basedOn: %v", r)
			}
		}()
		ppr, err = resolver.ResolveParagraph(p)
	}()

	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if ppr == nil {
		t.Fatal("ppr is nil")
	}

	// CycleA's spacing.before=100 should be in the result (it was
	// visited before the cycle was detected).  Since there are no
	// docDefaults, this is the only value.
	assertPPrSpacing(t, ppr, "Before", 100)

	// Warning must be emitted.
	warnings := resolver.Warnings()
	assertWarningContains(t, warnings, "circular basedOn chain")
}

// ---------- Test: dangling basedOn (missing styleId) ----------

func TestResolveParagraph_DanglingBasedOn(t *testing.T) {
	stylesBody := `
<w:style w:type="paragraph" w:styleId="Ghost">
  <w:name w:val="Ghost"/>
  <w:basedOn w:val="Missing"/>
  <w:pPr><w:spacing w:before="50"/></w:pPr>
</w:style>`
	stylesXML := buildStylesXML(t, stylesBody)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("Ghost")},
	}}
	ppr, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if ppr == nil {
		t.Fatal("ppr is nil")
	}

	// Ghost's own spacing.before=50 should be present.
	assertPPrSpacing(t, ppr, "Before", 50)
	// Warning for missing styleId.
	assertWarningContains(t, resolver.Warnings(), "not found in styles.xml or latentStyles")
}

// ---------- Test: latentStyles fallback ----------

func TestResolveParagraph_LatentStylesFallback(t *testing.T) {
	stylesBody := `
<w:latentStyles>
  <w:lsdException w:name="LatentOnly"/>
</w:latentStyles>`
	stylesXML := buildStylesXML(t, stylesBody)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("LatentOnly")},
	}}
	ppr, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if ppr == nil {
		t.Fatal("ppr is nil")
	}

	// No docDefaults in this fixture, so ppr should be empty CT_PPr{}.
	// Just check spacing is nil (no docDefaults).
	if ppr.Spacing != nil {
		t.Fatalf("ppr.Spacing = %+v, want nil (no docDefaults)", ppr.Spacing)
	}
	assertWarningContains(t, resolver.Warnings(), "latentStyles fallback")
}

// ---------- Test: unstyled paragraph skips latentStyles ----------

func TestResolveParagraph_UnstyledSkipsLatent(t *testing.T) {
	stylesBody := `
<w:latentStyles>
  <w:lsdException w:name="Heading1"/>
</w:latentStyles>`
	stylesXML := buildStylesXML(t, stylesBody)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	// Unstyled paragraph — no PStyle.
	p := &wml.CT_P{PPr: &wml.CT_PPr{}}
	ppr, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if ppr == nil {
		t.Fatal("ppr is nil")
	}

	// No warnings — latentStyles NEVER consulted for unstyled paragraphs.
	if len(resolver.Warnings()) != 0 {
		t.Fatalf("expected no warnings for unstyled paragraph, got %v", resolver.Warnings())
	}
}

// ---------- Test: memo hit ----------

func TestResolveParagraph_MemoHit(t *testing.T) {
	stylesBody := headingChainStyles
	stylesXML := buildStylesXML(t, stylesBody)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("Heading2")},
	}}

	// First call — walks chain and caches.
	ppr1, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("first ResolveParagraph: %v", err)
	}

	// Mutate the styles tree directly (white-box) so we can detect
	// whether the second call re-walks or hits the cache.
	if resolver.styles != nil && len(resolver.styles.Style) > 0 {
		for _, s := range resolver.styles.Style {
			if s.StyleID != nil && *s.StyleID == "Heading2" && s.PPr != nil && s.PPr.Spacing != nil {
				newBefore := int64(9999)
				s.PPr.Spacing.Before = &newBefore
				break
			}
		}
	}

	// Second call — should hit the memo cache (return pre-mutation value).
	ppr2, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("second ResolveParagraph: %v", err)
	}

	// Same Before value as first call (memo hit, not re-walked).
	assertPPrSpacing(t, ppr2, "Before", 240)

	// Now clear memo cache (white-box) and re-resolve.
	resolver.memoPPr = make(map[string]*wml.CT_PPr)
	resolver.memoRPr = make(map[string]*wml.CT_RPr)

	ppr3, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("third ResolveParagraph: %v", err)
	}

	// After cache clear and re-walk, the mutated value (9999) should appear.
	assertPPrSpacing(t, ppr3, "Before", 9999)

	// First and second results should be equal (memo hit).
	if (ppr1.Spacing == nil || ppr1.Spacing.Before == nil) ||
		(ppr2.Spacing == nil || ppr2.Spacing.Before == nil) ||
		*ppr1.Spacing.Before != *ppr2.Spacing.Before {
		t.Fatalf("memo hit not working: ppr1.Before=%v, ppr2.Before=%v",
			safeInt64(ppr1), safeInt64(ppr2))
	}
}

func safeInt64(ppr *wml.CT_PPr) string {
	if ppr == nil || ppr.Spacing == nil || ppr.Spacing.Before == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *ppr.Spacing.Before)
}

// ---------- Test: clone independence ----------

func TestResolveParagraph_CloneIndependence(t *testing.T) {
	stylesXML := buildStylesXML(t, headingChainStyles)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("Heading2")},
	}}

	ppr1, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("first ResolveParagraph: %v", err)
	}
	before1 := *ppr1.Spacing.Before

	// Mutate the returned clone.
	newBefore := int64(99999)
	ppr1.Spacing.Before = &newBefore

	// Second call must return the original value, not the mutated one.
	ppr2, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("second ResolveParagraph: %v", err)
	}
	if *ppr2.Spacing.Before != before1 {
		t.Fatalf("clone independence violated: got %d, want %d",
			*ppr2.Spacing.Before, before1)
	}
}

// ---------- Test: cache invalidation via Part.IsModified ----------

func TestResolveParagraph_CacheInvalidation(t *testing.T) {
	// Build initial styles with Heading2 spacing.before=240.
	initialBody := `
<w:style w:type="paragraph" w:styleId="Heading2">
  <w:name w:val="heading 2"/>
  <w:pPr><w:spacing w:before="240"/></w:pPr>
</w:style>`
	initialXML := buildStylesXML(t, initialBody)
	pkg := buildPkgWithStyles(t, initialXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("Heading2")},
	}}

	// First call — establishes cache.
	ppr1, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("first ResolveParagraph: %v", err)
	}
	assertPPrSpacing(t, ppr1, "Before", 240)

	// Replace styles.xml with new content (different spacing).
	newBody := `
<w:style w:type="paragraph" w:styleId="Heading2">
  <w:name w:val="heading 2"/>
  <w:pPr><w:spacing w:before="999"/></w:pPr>
</w:style>`
	newXML := buildStylesXML(t, newBody)
	pkg.MarkModified("word/styles.xml", newXML)

	// Second call — must detect IsModified, invalidate cache, re-parse.
	ppr2, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("second ResolveParagraph: %v", err)
	}
	assertPPrSpacing(t, ppr2, "Before", 999)
}

// ---------- Test: ThemeColor pass-through (NOT cleared) ----------

func TestResolveParagraph_ThemeColorPassThrough(t *testing.T) {
	stylesBody := `
<w:docDefaults>
  <w:rPrDefault><w:rPr><w:sz w:val="22"/></w:rPr></w:rPrDefault>
</w:docDefaults>
<w:style w:type="paragraph" w:styleId="Normal">
  <w:name w:val="Normal"/>
</w:style>
<w:style w:type="paragraph" w:styleId="Heading1">
  <w:name w:val="heading 1"/>
  <w:basedOn w:val="Normal"/>
  <w:rPr><w:color w:themeColor="accent1"/></w:rPr>
</w:style>`
	stylesXML := buildStylesXML(t, stylesBody)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("Heading1")},
	}}
	run := &wml.CT_R{} // no explicit rPr

	rpr, err := resolver.ResolveRun(p, run)
	if err != nil {
		t.Fatalf("ResolveRun: %v", err)
	}
	if rpr == nil || rpr.Color == nil {
		t.Fatal("rpr.Color is nil")
	}
	// ThemeColor must be "accent1" — NOT cleared by this plan.
	if rpr.Color.ThemeColor == nil {
		t.Fatal("ThemeColor was cleared (nil) — it must be passed through for plan 02-02")
	}
	if *rpr.Color.ThemeColor != "accent1" {
		t.Fatalf("ThemeColor = %q, want %q", *rpr.Color.ThemeColor, "accent1")
	}
}

// ---------- Test: NumPr pass-through (no lvl.PPr merge) ----------

func TestResolveParagraph_NumPrPassThrough(t *testing.T) {
	stylesBody := `
<w:docDefaults>
  <w:pPrDefault><w:pPr><w:spacing w:after="160"/></w:pPr></w:pPrDefault>
</w:docDefaults>`
	stylesXML := buildStylesXML(t, stylesBody)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	// Paragraph with direct NumPr — no style-based NumPr.
	numId := int64(1)
	ilvl := int64(0)
	p := &wml.CT_P{PPr: &wml.CT_PPr{
		NumPr: &wml.CT_NumPr{
			NumId: &wml.CT_NumId{Val: &numId},
			ILvl:  &wml.CT_ILvl{Val: &ilvl},
		},
	}}

	ppr, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if ppr == nil || ppr.NumPr == nil {
		t.Fatal("ppr.NumPr is nil — should be passed through")
	}
	if ppr.NumPr.NumId == nil || ppr.NumPr.NumId.Val == nil || *ppr.NumPr.NumId.Val != 1 {
		t.Fatalf("ppr.NumPr.NumId = %+v, want Val=1", ppr.NumPr.NumId)
	}
	if ppr.NumPr.ILvl == nil || ppr.NumPr.ILvl.Val == nil || *ppr.NumPr.ILvl.Val != 0 {
		t.Fatalf("ppr.NumPr.ILvl = %+v, want Val=0", ppr.NumPr.ILvl)
	}
}

// ---------- Test: no styles part ----------

func TestResolveParagraph_NoStylesPart(t *testing.T) {
	// Build a minimal package WITHOUT word/styles.xml.
	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
	rootRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	document := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="%s"><w:body><w:p/></w:body></w:document>`, stylesNS)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name string, payload []byte) {
		t.Helper()
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.Modified = time.Date(2020, 3, 4, 5, 6, 7, 0, time.UTC)
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	add("[Content_Types].xml", []byte(contentTypes))
	add("_rels/.rels", []byte(rootRels))
	add("word/document.xml", []byte(document))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	pkg, err := opc.Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	resolver := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("Anything")},
	}}
	ppr, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if ppr == nil {
		t.Fatal("ppr is nil")
	}
	// No docDefaults, no styles — result should be empty CT_PPr{}.
	// But warnings should include "not found" for the styleId.
	assertWarningContains(t, resolver.Warnings(), "not found in styles.xml or latentStyles")
}

// ---------- Test: Warnings accumulation ----------

func TestResolver_Warnings(t *testing.T) {
	// Set up a resolver that will trigger multiple warnings:
	// CycleA→CycleB→CycleA (circular) + Ghost→Missing (dangling).
	stylesBody := `
<w:style w:type="paragraph" w:styleId="CycleA">
  <w:name w:val="CycleA"/>
  <w:basedOn w:val="CycleB"/>
</w:style>
<w:style w:type="paragraph" w:styleId="CycleB">
  <w:name w:val="CycleB"/>
  <w:basedOn w:val="CycleA"/>
</w:style>
<w:style w:type="paragraph" w:styleId="Ghost">
  <w:name w:val="Ghost"/>
  <w:basedOn w:val="Missing"/>
</w:style>`
	stylesXML := buildStylesXML(t, stylesBody)
	pkg := buildPkgWithStyles(t, stylesXML)
	resolver := NewResolver(pkg)

	// Resolve circular.
	p1 := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("CycleA")},
	}}
	_, _ = resolver.ResolveParagraph(p1)

	// Resolve dangling.
	p2 := &wml.CT_P{PPr: &wml.CT_PPr{
		PStyle: &wml.CT_PStyle{Val: strPtr("Ghost")},
	}}
	_, _ = resolver.ResolveParagraph(p2)

	warnings := resolver.Warnings()
	if len(warnings) < 2 {
		t.Fatalf("expected at least 2 warnings, got %d: %v", len(warnings), warnings)
	}
	assertWarningContains(t, warnings, "circular basedOn chain")
	assertWarningContains(t, warnings, "not found in styles.xml or latentStyles")
}

// ---------- Test: real-fixture placeholder ----------

func TestResolveParagraph_RealFixtures(t *testing.T) {
	// Per D-10/D-11, real .docx fixtures are user-authored (Word not
	// available on this CI machine).  This placeholder signals plan
	// 02-03 where the real-fixture equality harness is wired.
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Walk up to project root looking for testdata/word/style-rich/.
	// The test runner cwd may be the internal/style directory itself.
	fixtureDir := filepath.Join(root, "..", "..", "testdata", "word", "style-rich")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		// Also check from repo root.
		fixtureDir = filepath.Join(root, "testdata", "word", "style-rich")
		if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
			t.Skip("testdata/word/style-rich/ not present — real-fixture tests require user-authored .docx per D-10")
		}
	}
	// If fixtures exist, we'd run equality assertions here.  Placeholder
	// for 02-03 to wire.
	t.Log("style-rich fixtures present — 02-03 will wire equality harness here")
}

// ---------- Helpers ----------

func strPtr(s string) *string { return &s }
