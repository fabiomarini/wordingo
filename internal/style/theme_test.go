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

const drawingmlNS = "http://schemas.openxmlformats.org/drawingml/2006/main"

// ---------- Synthetic fixture builders ----------

// buildThemeXML wraps clrSchemeBody in a standard theme1.xml envelope.
func buildThemeXML(t *testing.T, clrSchemeBody string) []byte {
	t.Helper()
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="%s" name="Test"><a:themeElements><a:clrScheme name="Test">%s</a:clrScheme></a:themeElements></a:theme>`,
		drawingmlNS, clrSchemeBody)
	return []byte(xml)
}

// buildPkgWithTheme constructs an in-memory *opc.Package with a theme1.xml part.
// Follows the same pattern as buildPkgWithStyles in resolver_test.go.
func buildPkgWithTheme(t *testing.T, themeXML []byte) *opc.Package {
	t.Helper()

	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
</Types>`

	rootRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	docRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>
</Relationships>`

	document := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p/></w:body></w:document>`)

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
	add("word/theme/theme1.xml", themeXML)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	pkg, err := opc.Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return pkg
}

// ---------- Theme color resolution tests ----------

func TestResolveColor_Accent1(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("accent1")}
	result := tc.ResolveColor(c)

	if result.Val == nil || *result.Val != "156082" {
		t.Errorf("Val = %v, want 156082", result.Val)
	}
	if result.ThemeColor != nil {
		t.Errorf("ThemeColor should be nil after concretization, got %v", *result.ThemeColor)
	}
	if result.ThemeShade != nil {
		t.Errorf("ThemeShade should be nil after concretization")
	}
	if result.ThemeTint != nil {
		t.Errorf("ThemeTint should be nil after concretization")
	}
	if len(warns) > 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
}

func TestResolveColor_Dark1_SysClr(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("dark1")}
	result := tc.ResolveColor(c)

	if result.Val == nil || *result.Val != "000000" {
		t.Errorf("Val = %v, want 000000 (from lastClr, not val)", result.Val)
	}
	if len(warns) > 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
}

func TestResolveColor_Light1_SysClr(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("light1")}
	result := tc.ResolveColor(c)

	if result.Val == nil || *result.Val != "FFFFFF" {
		t.Errorf("Val = %v, want FFFFFF", result.Val)
	}
}

func TestResolveColor_Hyperlink(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:hlink><a:srgbClr val="467886"/></a:hlink>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("hyperlink")}
	result := tc.ResolveColor(c)

	if result.Val == nil || *result.Val != "467886" {
		t.Errorf("Val = %v, want 467886 (Pitfall 3: hyperlink → hlink)", result.Val)
	}
}

func TestResolveColor_FollowedHyperlink(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:folHlink><a:srgbClr val="96607D"/></a:folHlink>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("followedHyperlink")}
	result := tc.ResolveColor(c)

	if result.Val == nil || *result.Val != "96607D" {
		t.Errorf("Val = %v, want 96607D (Pitfall 3: followedHyperlink → folHlink)", result.Val)
	}
}

func TestResolveColor_Dark2Light2Accents(t *testing.T) {
	tests := []struct {
		themeColor string
		elemName   string
		wantHex    string
	}{
		{"dark2", "dk2", "0E2841"},
		{"light2", "lt2", "E8E8E8"},
		{"accent2", "accent2", "E97132"},
		{"accent3", "accent3", "196B24"},
		{"accent4", "accent4", "0F9ED5"},
		{"accent5", "accent5", "A02B93"},
		{"accent6", "accent6", "4EA72E"},
	}

	body := ""
	for _, tt := range tests {
		body += fmt.Sprintf(`<a:%s><a:srgbClr val="%s"/></a:%s>`, tt.elemName, tt.wantHex, tt.elemName)
	}
	themeXML := buildThemeXML(t, body)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	for _, tt := range tests {
		t.Run(tt.themeColor, func(t *testing.T) {
			c := &wml.CT_Color{ThemeColor: strPtr(tt.themeColor)}
			result := tc.ResolveColor(c)
			if result.Val == nil || *result.Val != tt.wantHex {
				t.Errorf("Val = %v, want %s", result.Val, tt.wantHex)
			}
		})
	}
	if len(warns) > 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
}

func TestResolveColor_UnknownEnum(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("nonexistent")}
	result := tc.ResolveColor(c)

	if result.ThemeColor == nil || *result.ThemeColor != "nonexistent" {
		t.Errorf("ThemeColor should be preserved: got %v, want nonexistent", result.ThemeColor)
	}
	found := false
	for _, w := range warns {
		if strings.Contains(w, "unknown themeColor") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("warning should contain 'unknown themeColor', got: %v", warns)
	}
}

func TestResolveColor_MissingElement(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("accent2")}
	result := tc.ResolveColor(c)

	if result.ThemeColor == nil || *result.ThemeColor != "accent2" {
		t.Errorf("ThemeColor should be preserved on missing element")
	}
	found := false
	for _, w := range warns {
		if strings.Contains(w, "not found in theme1.xml") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("warning should contain 'not found in theme1.xml', got: %v", warns)
	}
}

func TestResolveColor_NilColor(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	result := tc.ResolveColor(nil)
	if result != nil {
		t.Errorf("ResolveColor(nil) should return nil")
	}
	if len(warns) > 0 {
		t.Errorf("unexpected warnings on nil color")
	}
}

func TestResolveColor_EmptyThemeColor(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{Val: strPtr("FF0000"), ThemeColor: strPtr("")}
	result := tc.ResolveColor(c)

	if result.Val == nil || *result.Val != "FF0000" {
		t.Errorf("Val should be preserved on empty ThemeColor")
	}
	if len(warns) > 0 {
		t.Errorf("unexpected warnings on empty ThemeColor: %v", warns)
	}
}

func TestResolveColor_TintMath(t *testing.T) {
	// accent1=0x156082 (R=0x15=21, G=0x60=96, B=0x82=130)
	// ThemeTint="80" → 128/255 ≈ 0.502
	// R = 21 + (255-21)*0.502 = 21 + 234*0.502 = 21 + 117.5 = 138.5 → 139 (0x8B)
	// G = 96 + (255-96)*0.502 = 96 + 159*0.502 = 96 + 79.8 = 175.8 → 176 (0xB0)
	// B = 130 + (255-130)*0.502 = 130 + 125*0.502 = 130 + 62.8 = 192.8 → 193 (0xC1)
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("accent1"), ThemeTint: strPtr("80")}
	result := tc.ResolveColor(c)

	expected := "8BB0C1"
	if result.Val == nil || *result.Val != expected {
		t.Errorf("Val = %v, want %s (tint 0x80 applied)", result.Val, expected)
	}
	if result.ThemeTint != nil {
		t.Errorf("ThemeTint should be nil after concretization")
	}
}

func TestResolveColor_ShadeMath(t *testing.T) {
	// accent1=0x156082 (R=21, G=96, B=130)
	// ThemeShade="80" → 128/255 ≈ 0.502
	// R = 21*0.502 = 10.5 → 11 (0x0B)
	// G = 96*0.502 = 48.2 → 48 (0x30)
	// B = 130*0.502 = 65.3 → 65 (0x41)
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("accent1"), ThemeShade: strPtr("80")}
	result := tc.ResolveColor(c)

	expected := "0B3041"
	if result.Val == nil || *result.Val != expected {
		t.Errorf("Val = %v, want %s (shade 0x80 applied)", result.Val, expected)
	}
	if result.ThemeShade != nil {
		t.Errorf("ThemeShade should be nil after concretization")
	}
}

func TestResolveColor_TintShadeOrder(t *testing.T) {
	// accent1=0x156082 (R=21, G=96, B=130)
	// ThemeShade="80" first: R=10.5→11 (0x0B), G=48 (0x30), B=65 (0x41) → 0B3041
	// Then ThemeTint="80": R=11+(255-11)*0.502=11+122.5=133.5→134 (0x86)
	//                        G=48+(255-48)*0.502=48+103.9=151.9→152 (0x98)
	//                        B=65+(255-65)*0.502=65+95.4=160.4→160 (0xA0)
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("accent1"), ThemeShade: strPtr("80"), ThemeTint: strPtr("80")}
	result := tc.ResolveColor(c)

	expected := "8698A0"
	if result.Val == nil || *result.Val != expected {
		t.Errorf("Val = %v, want %s (shade FIRST then tint — Pitfall 5)", result.Val, expected)
	}
}

func TestResolveColor_MissingThemePart(t *testing.T) {
	// Package with no word/theme/theme1.xml
	pkg := buildPkgWithTheme(t, nil) // this will create a pkg without theme
	// Build a pkg without theme — use the builder but delete the part
	delete(pkg.Parts, "word/theme/theme1.xml")

	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("accent1")}
	result := tc.ResolveColor(c)

	if result.ThemeColor == nil || *result.ThemeColor != "accent1" {
		t.Errorf("ThemeColor should be preserved when theme part missing")
	}
	if len(warns) == 0 {
		t.Errorf("expected warning for missing theme part")
	}
}

func TestResolveColor_MalformedThemeXML(t *testing.T) {
	// Truncated XML
	malformed := []byte(`<?xml version="1.0"?><a:theme xmlns:a="` + drawingmlNS + `"><a:themeElements><a:clrScheme name="Test">`)
	pkg := buildPkgWithTheme(t, malformed)

	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic on malformed XML: %v", r)
		}
	}()

	c := &wml.CT_Color{ThemeColor: strPtr("accent1")}
	result := tc.ResolveColor(c)

	if result.ThemeColor == nil || *result.ThemeColor != "accent1" {
		t.Errorf("ThemeColor should be preserved on malformed XML")
	}
	if len(warns) == 0 {
		t.Errorf("expected warning for malformed theme XML")
	}
}

func TestResolveColor_CacheHit(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	// First call populates cache
	c := &wml.CT_Color{ThemeColor: strPtr("accent1")}
	result1 := tc.ResolveColor(c)

	// Mutate the theme part bytes — should NOT affect cached result
	pkg.MarkModified("word/theme/theme1.xml", []byte(`<?xml version="1.0"?><a:theme xmlns:a="`+drawingmlNS+`"><a:themeElements><a:clrScheme name="Test"><a:accent1><a:srgbClr val="FFFFFF"/></a:accent1></a:clrScheme></a:themeElements></a:theme>`))

	result2 := tc.ResolveColor(c)

	if result2.Val == nil || *result2.Val != "156082" {
		t.Errorf("Cache hit should return cached value, got Val = %v, want 156082", result2.Val)
	}
	if result1 == result2 {
		t.Errorf("ResolveColor should return a new clone each time, not the same pointer")
	}
}

func TestResolveColor_ReadOnly(t *testing.T) {
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string

	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("accent1")}
	tc.ResolveColor(c)

	// D-06: theme resolution is read-only — the part should not be marked modified
	if pkg.Parts["word/theme/theme1.xml"].IsModified() {
		t.Errorf("D-06: theme part should not be modified by ResolveColor")
	}
}

func TestResolveColor_NilShadeTint(t *testing.T) {
	// ThemeColor="accent1" with nil ThemeShade and nil ThemeTint — just resolves hex
	themeXML := buildThemeXML(t, `<a:accent1><a:srgbClr val="156082"/></a:accent1>`)
	pkg := buildPkgWithTheme(t, themeXML)
	var warns []string
	tc := newThemeCache(pkg, func(m string) { warns = append(warns, m) })

	c := &wml.CT_Color{ThemeColor: strPtr("accent1")}
	result := tc.ResolveColor(c)

	if result.Val == nil || *result.Val != "156082" {
		t.Errorf("Val = %v, want 156082", result.Val)
	}
	if result.ThemeColor != nil {
		t.Errorf("ThemeColor not cleared after resolution")
	}
}

// strPtr is already defined in resolver_test.go
