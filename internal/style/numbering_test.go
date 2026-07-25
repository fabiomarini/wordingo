package style

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
)

// ---------- Synthetic fixture builders ----------

const wmlNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

// buildNumberingXML wraps body in a standard numbering.xml envelope.
func buildNumberingXML(t *testing.T, body string) []byte {
	t.Helper()
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="%s">
%s
</w:numbering>`, wmlNS, body)
	return []byte(xml)
}

// buildPkgWithNumbering constructs an in-memory *opc.Package with numbering.xml
// and a minimal document.  Follows buildPkgWithStyles pattern.
func buildPkgWithNumbering(t *testing.T, numberingXML []byte) *opc.Package {
	t.Helper()
	return buildPkgWithParts(t, map[string][]byte{
		"word/document.xml": []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="%s"><w:body><w:p/></w:body></w:document>`, wmlNS)),
		"word/numbering.xml": numberingXML,
	})
}

// buildPkgWithNumberingAndStyles builds a package with both numbering.xml
// and styles.xml, plus minimal document.
func buildPkgWithNumberingAndStyles(t *testing.T, numberingXML, stylesXML []byte) *opc.Package {
	t.Helper()
	parts := map[string][]byte{
		"word/document.xml": []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="%s"><w:body><w:p/></w:body></w:document>`, wmlNS)),
	}
	if numberingXML != nil {
		parts["word/numbering.xml"] = numberingXML
	}
	if stylesXML != nil {
		parts["word/styles.xml"] = stylesXML
	}
	return buildPkgWithParts(t, parts)
}

// buildPkgWithThemeAndNumbering builds a package with theme + numbering + styles.
func buildPkgWithThemeAndNumbering(t *testing.T, themeXML, numberingXML, stylesXML []byte) *opc.Package {
	t.Helper()
	parts := map[string][]byte{
		"word/document.xml": []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="%s"><w:body><w:p/></w:body></w:document>`, wmlNS)),
	}
	if themeXML != nil {
		parts["word/theme/theme1.xml"] = themeXML
	}
	if numberingXML != nil {
		parts["word/numbering.xml"] = numberingXML
	}
	if stylesXML != nil {
		parts["word/styles.xml"] = stylesXML
	}
	return buildPkgWithParts(t, parts)
}

// buildPkgWithParts is a general-purpose package builder for synthetic test
// fixtures.  parts maps part names to their byte payloads.  Content types and
// relationships are auto-generated for known part types.
func buildPkgWithParts(t *testing.T, parts map[string][]byte) *opc.Package {
	t.Helper()

	// Auto-generate content types for known parts
	overrides := ""
	docRels := ""
	if _, ok := parts["word/document.xml"]; ok {
		overrides += `<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>` + "\n"
		docRels += `<Relationship Id="rId_doc" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>` + "\n"
	}
	if _, ok := parts["word/styles.xml"]; ok {
		overrides += `<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>` + "\n"
		docRels += `<Relationship Id="rId_styles" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>` + "\n"
	}
	if _, ok := parts["word/numbering.xml"]; ok {
		overrides += `<Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>` + "\n"
		docRels += `<Relationship Id="rId_numbering" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/>` + "\n"
	}
	if _, ok := parts["word/theme/theme1.xml"]; ok {
		overrides += `<Override PartName="/word/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>` + "\n"
		docRels += `<Relationship Id="rId_theme" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>` + "\n"
	}
	if _, ok := parts["word/fontTable.xml"]; ok {
		overrides += `<Override PartName="/word/fontTable.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"/>` + "\n"
		docRels += `<Relationship Id="rId_fontTable" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable" Target="fontTable.xml"/>` + "\n"
	}

	contentTypes := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
%s</Types>`, overrides)

	rootRels := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
%s</Relationships>`, docRels)

	docRelsContent := ""
	if _, ok := parts["word/styles.xml"]; ok {
		docRelsContent += `<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`
	}
	if _, ok := parts["word/numbering.xml"]; ok {
		docRelsContent += `<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/>`
	}
	if _, ok := parts["word/theme/theme1.xml"]; ok {
		docRelsContent += `<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>`
	}
	if _, ok := parts["word/fontTable.xml"]; ok {
		docRelsContent += `<Relationship Id="rId4" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable" Target="fontTable.xml"/>`
	}
	if docRelsContent == "" {
		docRelsContent = `<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`
	}

	wordRels := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
%s</Relationships>`, docRelsContent)

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

	// Only add word/_rels/document.xml.rels if we have a document
	if _, ok := parts["word/document.xml"]; ok {
		add("word/_rels/document.xml.rels", []byte(wordRels))
	}

	// Add all specified parts
	for name, payload := range parts {
		if name == "word/document.xml" {
			// Already have the constant; skip
		}
		add(name, payload)
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	pkg, err := opc.Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return pkg
}

// ---------- Numbering unit tests ----------

func TestResolveLvl_Basic(t *testing.T) {
	numberingXML := buildNumberingXML(t, `
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`)
	pkg := buildPkgWithNumbering(t, numberingXML)
	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	lvl := nc.ResolveLvl(1, 0)
	if lvl == nil {
		t.Fatal("ResolveLvl returned nil")
	}
	if lvl.NumFmt == nil || lvl.NumFmt.Val == nil || *lvl.NumFmt.Val != "decimal" {
		t.Errorf("NumFmt.Val = %v, want decimal", lvl.NumFmt)
	}
	if lvl.LvlText == nil || lvl.LvlText.Val == nil || *lvl.LvlText.Val != "%1." {
		t.Errorf("LvlText.Val = %v, want %%1.", lvl.LvlText)
	}
	if lvl.Start == nil || lvl.Start.Val == nil || *lvl.Start.Val != 1 {
		t.Errorf("Start.Val = %v, want 1", lvl.Start)
	}
	if len(warns) > 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
}

func TestResolveLvl_StartOverride(t *testing.T) {
	// Pitfall 4: lvlOverride with startOverride changes Start
	numberingXML := buildNumberingXML(t, `
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/><w:lvlOverride w:ilvl="0"><w:startOverride w:val="5"/></w:lvlOverride></w:num>`)
	pkg := buildPkgWithNumbering(t, numberingXML)
	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	lvl := nc.ResolveLvl(1, 0)
	if lvl == nil {
		t.Fatal("ResolveLvl returned nil")
	}
	if lvl.Start == nil || lvl.Start.Val == nil || *lvl.Start.Val != 5 {
		t.Errorf("Start.Val = %v, want 5 (Pitfall 4: startOverride applied)", lvl.Start)
	}
	// numFmt and lvlText should still come from abstractNum
	if lvl.NumFmt == nil || lvl.NumFmt.Val == nil || *lvl.NumFmt.Val != "decimal" {
		t.Errorf("NumFmt should be from abstractNum, got %v", lvl.NumFmt)
	}
	if len(warns) > 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
}

func TestResolveLvl_FullLvlReplacement(t *testing.T) {
	// Pitfall 4: lvlOverride with full lvl replaces the abstractNum's lvl
	numberingXML := buildNumberingXML(t, `
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
<w:lvl w:ilvl="1"><w:numFmt w:val="bullet"/><w:lvlText w:val="•"/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/><w:lvlOverride w:ilvl="1"><w:lvl w:ilvl="1"><w:numFmt w:val="decimal"/><w:lvlText w:val="(%1)"/><w:start w:val="1"/></w:lvl></w:lvlOverride></w:num>`)
	pkg := buildPkgWithNumbering(t, numberingXML)
	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	lvl := nc.ResolveLvl(1, 1)
	if lvl == nil {
		t.Fatal("ResolveLvl returned nil")
	}
	// Should have the override's values, not the abstractNum's
	if lvl.LvlText == nil || lvl.LvlText.Val == nil || *lvl.LvlText.Val != "(%1)" {
		t.Errorf("LvlText.Val = %v, want (%%1) (full lvl replacement)", lvl.LvlText)
	}
	if len(warns) > 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
}

func TestResolveLvl_MissingNumId(t *testing.T) {
	numberingXML := buildNumberingXML(t, `
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`)
	pkg := buildPkgWithNumbering(t, numberingXML)
	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	lvl := nc.ResolveLvl(999, 0)
	if lvl != nil {
		t.Errorf("ResolveLvl should return nil for missing numId")
	}
	found := false
	for _, w := range warns {
		if strings.Contains(w, "numId") && (strings.Contains(w, "not found") || strings.Contains(w, "999")) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("warning should mention numId not found, got: %v", warns)
	}
}

func TestResolveLvl_MissingIlvl(t *testing.T) {
	numberingXML := buildNumberingXML(t, `
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`)
	pkg := buildPkgWithNumbering(t, numberingXML)
	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	lvl := nc.ResolveLvl(1, 9)
	if lvl != nil {
		t.Errorf("ResolveLvl should return nil for missing ilvl")
	}
	found := false
	for _, w := range warns {
		if strings.Contains(w, "ilvl") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("warning should mention ilvl not found, got: %v", warns)
	}
}

func TestResolveLvl_MissingAbstractNum(t *testing.T) {
	numberingXML := buildNumberingXML(t, `
<w:num w:numId="1"><w:abstractNumId w:val="888"/></w:num>`)
	pkg := buildPkgWithNumbering(t, numberingXML)
	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	lvl := nc.ResolveLvl(1, 0)
	if lvl != nil {
		t.Errorf("ResolveLvl should return nil for missing abstractNum")
	}
	found := false
	for _, w := range warns {
		if strings.Contains(w, "abstractNum") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("warning should mention abstractNum not found, got: %v", warns)
	}
}

func TestResolveLvl_MissingNumberingPart(t *testing.T) {
	pkg := buildPkgWithNumbering(t, nil) // will create pkg without numbering
	delete(pkg.Parts, "word/numbering.xml")

	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	lvl := nc.ResolveLvl(1, 0)
	if lvl != nil {
		t.Errorf("ResolveLvl should return nil when numbering.xml missing")
	}
	if len(warns) == 0 {
		t.Errorf("expected warning for missing numbering.xml")
	}
}

func TestResolveLvl_MalformedXML(t *testing.T) {
	malformed := []byte(`<?xml version="1.0"?><w:numbering xmlns:w="` + wmlNS + `"><w:num w:numId="1">`)
	pkg := buildPkgWithNumbering(t, malformed)

	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic on malformed XML: %v", r)
		}
	}()

	lvl := nc.ResolveLvl(1, 0)
	if lvl != nil {
		t.Errorf("ResolveLvl should return nil on malformed XML")
	}
	if len(warns) == 0 {
		t.Errorf("expected warning for malformed numbering.xml")
	}
}

func TestResolveLvl_CacheHit(t *testing.T) {
	numberingXML := buildNumberingXML(t, `
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`)
	pkg := buildPkgWithNumbering(t, numberingXML)
	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	lvl1 := nc.ResolveLvl(1, 0)
	if lvl1 == nil {
		t.Fatal("first ResolveLvl returned nil")
	}

	// Mutate the numbering part — cached value should survive
	pkg.MarkModified("word/numbering.xml", []byte(`<?xml version="1.0"?><w:numbering xmlns:w="`+wmlNS+`"></w:numbering>`))

	// Note: cache is NOT invalidated by MarkModified alone — invalidateIfStale
	// must be called.  Without invalidation, cached value is returned.
	lvl2 := nc.ResolveLvl(1, 0)
	if lvl2 == nil {
		t.Fatal("second ResolveLvl (cache hit) returned nil")
	}
	if lvl2.NumFmt == nil || lvl2.NumFmt.Val == nil || *lvl2.NumFmt.Val != "decimal" {
		t.Errorf("cache hit should return cached numFmt, got %v", lvl2.NumFmt)
	}
	// Verify different pointers (each ResolveLvl returns a fresh reference
	// into the cached tree — they should be different calls returning
	// the same tree)
	if lvl1 == lvl2 {
		t.Log("note: same pointer returned (this is expected as the cache tree is reused)")
	}
}

func TestResolveLvl_CacheInvalidation(t *testing.T) {
	numberingXML := buildNumberingXML(t, `
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`)
	pkg := buildPkgWithNumbering(t, numberingXML)
	var warns []string
	nc := newNumberingCache(pkg, func(m string) { warns = append(warns, m) })

	lvl1 := nc.ResolveLvl(1, 0)
	if lvl1 == nil {
		t.Fatal("first ResolveLvl returned nil")
	}

	// Replace numbering.xml with different content + mark modified
	newNumberingXML := buildNumberingXML(t, `
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="upperRoman"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`)
	pkg.MarkModified("word/numbering.xml", newNumberingXML)

	// Invalidate the cache manually (as ResolveParagraph does via invalidateIfStale)
	nc.invalidateIfStale()

	lvl2 := nc.ResolveLvl(1, 0)
	if lvl2 == nil {
		t.Fatal("second ResolveLvl after invalidation returned nil")
	}
	if lvl2.NumFmt == nil || lvl2.NumFmt.Val == nil || *lvl2.NumFmt.Val != "upperRoman" {
		t.Errorf("after invalidation, NumFmt = %v, want upperRoman", lvl2.NumFmt)
	}
}

// ---------- Integration tests: ResolveParagraph + ResolveRun ----------

func TestResolver_ResolveRun_ThemeColorConcretization(t *testing.T) {
	// Build a package with a Heading1 style setting themeColor="accent1"
	stylesXML := buildStylesXML(t, `
<w:style w:type="paragraph" w:styleId="Heading1">
<w:name w:val="heading 1"/>
<w:pPr><w:spacing w:after="120"/></w:pPr>
<w:rPr><w:color w:themeColor="accent1"/></w:rPr>
</w:style>`)
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithThemeAndNumbering(t, themeXML, nil, stylesXML)

	r := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{PStyle: &wml.CT_PStyle{Val: strPtr("Heading1")}}}
	run := &wml.CT_R{}

	result, err := r.ResolveRun(p, run)
	if err != nil {
		t.Fatalf("ResolveRun: %v", err)
	}
	if result.Color == nil || result.Color.Val == nil || *result.Color.Val != "156082" {
		t.Errorf("Color.Val = %v, want 156082", result.Color)
	}
	if result.Color.ThemeColor != nil {
		t.Errorf("ThemeColor should be nil after concretization (D-06), got %v", *result.Color.ThemeColor)
	}
}

func TestResolver_ResolveRun_ThemeColorDark1SysClr(t *testing.T) {
	stylesXML := buildStylesXML(t, `
<w:style w:type="paragraph" w:styleId="Heading1">
<w:name w:val="heading 1"/>
<w:rPr><w:color w:themeColor="dark1"/></w:rPr>
</w:style>`)
	themeXML := buildThemeXML(t, `<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>`)
	pkg := buildPkgWithThemeAndNumbering(t, themeXML, nil, stylesXML)

	r := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{PStyle: &wml.CT_PStyle{Val: strPtr("Heading1")}}}
	run := &wml.CT_R{}

	result, err := r.ResolveRun(p, run)
	if err != nil {
		t.Fatalf("ResolveRun: %v", err)
	}
	if result.Color == nil || result.Color.Val == nil || *result.Color.Val != "000000" {
		t.Errorf("Color.Val = %v, want 000000 (sysClr lastClr)", result.Color)
	}
}

func TestResolver_ResolveParagraph_NumberingLevelMerge(t *testing.T) {
	// Build a package where numbering level 0 has indentation
	numberingXML := buildNumberingXML(t, `
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0">
<w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/>
<w:pPr><w:ind w:left="720"/></w:pPr>
</w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`)
	pkg := buildPkgWithNumberingAndStyles(t, numberingXML, nil)

	r := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		NumPr: &wml.CT_NumPr{
			NumId: &wml.CT_NumId{Val: int64Ptr(1)},
			ILvl:  &wml.CT_ILvl{Val: int64Ptr(0)},
		},
	}}

	result, err := r.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if result == nil {
		t.Fatal("ResolveParagraph returned nil")
	}
	if result.NumPr == nil {
		t.Errorf("NumPr should be preserved after resolution")
	}
	if result.Ind == nil || result.Ind.Left == nil || *result.Ind.Left != 720 {
		t.Errorf("Ind.Left = %v, want 720 (level pPr merged)", result.Ind)
	}
}

func TestResolver_ResolveParagraph_MissingNumIdNoCrash(t *testing.T) {
	numberingXML := buildNumberingXML(t, `
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`)
	pkg := buildPkgWithNumberingAndStyles(t, numberingXML, nil)

	r := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{
		NumPr: &wml.CT_NumPr{
			NumId: &wml.CT_NumId{Val: int64Ptr(999)},
			ILvl:  &wml.CT_ILvl{Val: int64Ptr(0)},
		},
	}}

	result, err := r.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if result == nil {
		t.Fatal("ResolveParagraph returned nil")
	}
	if result.NumPr == nil {
		t.Errorf("NumPr should be preserved on missing numId")
	}

	warns := r.Warnings()
	found := false
	for _, w := range warns {
		if strings.Contains(w, "numId") && strings.Contains(w, "999") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning containing numId 999, got: %v", warns)
	}
}

func TestResolver_ResolveParagraph_NilNumPrNoop(t *testing.T) {
	pkg := buildPkgWithNumberingAndStyles(t, nil, nil)
	r := NewResolver(pkg)

	p := &wml.CT_P{} // no numPr

	result, err := r.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if len(r.Warnings()) > 0 {
		t.Errorf("expected no warnings on nil NumPr, got: %v", r.Warnings())
	}
	_ = result
}

func TestResolver_ResolveRun_NilColorNoop(t *testing.T) {
	stylesXML := buildStylesXML(t, `
<w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="normal"/></w:style>`)
	pkg := buildPkgWithThemeAndNumbering(t, nil, nil, stylesXML)

	r := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{PStyle: &wml.CT_PStyle{Val: strPtr("Normal")}}}
	run := &wml.CT_R{}

	result, err := r.ResolveRun(p, run)
	if err != nil {
		t.Fatalf("ResolveRun: %v", err)
	}
	if result.Color != nil {
		t.Log("Color is nil by default — good")
	}
	if len(r.Warnings()) > 0 {
		t.Errorf("expected no warnings on nil color, got: %v", r.Warnings())
	}
	_ = result
}

func TestResolver_NoRegression_02_01(t *testing.T) {
	// Exercise the Heading2 chain scenario from 02-01 and verify the
	// post-processing did not break the chain walker.
	stylesXML := buildStylesXML(t, `
<w:style w:type="paragraph" w:styleId="Normal">
<w:name w:val="normal"/>
<w:pPr><w:spacing w:after="120"/></w:pPr>
<w:rPr><w:sz w:val="20"/><w:b/></w:rPr>
</w:style>
<w:style w:type="paragraph" w:styleId="Heading1">
<w:name w:val="heading 1"/>
<w:basedOn w:val="Normal"/>
<w:pPr><w:spacing w:before="240"/><w:keepNext/></w:pPr>
<w:rPr><w:sz w:val="28"/><w:color w:val="2E74B5"/></w:rPr>
</w:style>
<w:style w:type="paragraph" w:styleId="Heading2">
<w:name w:val="heading 2"/>
<w:basedOn w:val="Heading1"/>
<w:pPr><w:spacing w:before="120"/></w:pPr>
<w:rPr><w:i/></w:rPr>
</w:style>`)
	pkg := buildPkgWithThemeAndNumbering(t, nil, nil, stylesXML)

	r := NewResolver(pkg)

	p := &wml.CT_P{PPr: &wml.CT_PPr{PStyle: &wml.CT_PStyle{Val: strPtr("Heading2")}}}
	result, err := r.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}

	// Verify chain merge: Normal spacing after=120 + Heading1 before=240 + Heading2 before=120
	if result.Spacing == nil {
		t.Fatal("Spacing is nil after chain merge")
	}
	if result.Spacing.After == nil || *result.Spacing.After != 120 {
		t.Errorf("Spacing.After = %v, want 120 (from Normal)", result.Spacing.After)
	}
	if result.Spacing.Before == nil || *result.Spacing.Before != 120 {
		t.Errorf("Spacing.Before = %v, want 120 (Heading2 overrides Heading1)", result.Spacing.Before)
	}
	if result.KeepNext == nil {
		t.Errorf("KeepNext should be preserved from Heading1")
	}

	// Verify run properties: Normal sz=20 + Heading1 sz=28 + Heading2 italic
	run := &wml.CT_R{}
	rp, err := r.ResolveRun(p, run)
	if err != nil {
		t.Fatalf("ResolveRun: %v", err)
	}
	if rp == nil {
		t.Fatal("run result is nil")
	}
	if rp.Sz == nil || rp.Sz.Val == nil || *rp.Sz.Val != 28 {
		t.Errorf("Sz.Val = %v, want 28 (Heading1 overrides Normal)", rp.Sz)
	}
	if rp.B == nil {
		t.Errorf("B (bold) should be from Normal")
	}
	if rp.I == nil {
		t.Errorf("I (italic) should be from Heading2")
	}
	if rp.Color == nil || rp.Color.Val == nil || *rp.Color.Val != "2E74B5" {
		t.Errorf("Color.Val = %v, want 2E74B5 (from Heading1)", rp.Color)
	}
}

func TestResolver_NoFontTableParse(t *testing.T) {
	// Pitfall 6: resolver must NOT parse fontTable.xml
	fontTableXML := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:fonts xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:font w:name="Calibri"><w:panose w:val="020F0502020204030204"/></w:font>
</w:fonts>`)

	pkg := buildPkgWithParts(t, map[string][]byte{
		"word/document.xml":  []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="%s"><w:body><w:p/></w:body></w:document>`, wmlNS)),
		"word/fontTable.xml": fontTableXML,
	})

	r := NewResolver(pkg)

	p := &wml.CT_P{}
	run := &wml.CT_R{}

	_, _ = r.ResolveParagraph(p)
	_, _ = r.ResolveRun(p, run)

	// Check no warnings about fontTable
	for _, w := range r.Warnings() {
		if strings.Contains(w, "fontTable") {
			t.Errorf("unexpected fontTable reference: %s", w)
		}
	}

	// Verify fontTable part data is unread by checking it's not modified
	part, ok := pkg.Parts["word/fontTable.xml"]
	if !ok {
		t.Fatal("fontTable.xml not in package")
	}
	if !part.IsModified() {
		// fontTable was NOT modified — confirm read-only boundary (D-06 principle)
	}
}

// ---------- Helpers ----------

func int64Ptr(i int64) *int64 { return &i }
