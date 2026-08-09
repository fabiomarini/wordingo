package wml_test

import (
	"bytes"
	"encoding/xml"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// ---- Type count check ----

func TestExportedTypeCount(t *testing.T) {
	// Compile-check: core types exist.
	var _ *wml.CT_Document
	var _ *wml.CT_P
	var _ *wml.CT_R
	var _ *wml.CT_Text
	var _ *wml.CT_Body
	var _ *wml.CT_Styles
	var _ *wml.CT_Numbering
	var _ *wml.CT_Tbl
	var _ *wml.CT_SectPr
	var _ *wml.CT_Settings

	// Count exported CT_* types by parsing this package's source.
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, 0)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	count := 0
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				gen, ok := n.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					return true
				}
				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if strings.HasPrefix(ts.Name.Name, "CT_") && ts.Name.IsExported() {
						count++
					}
				}
				return false
			})
		}
	}
	const floor = 55
	if count < floor {
		t.Errorf("exported CT_* type count = %d, want >= %d", count, floor)
	}
	t.Logf("exported CT_* type count = %d", count)
}

// ---- Round-trip tests ----

func TestRoundTrip_Document(t *testing.T) {
	input := `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p/></w:body></w:document>`

	var doc wml.CT_Document
	if err := xml.Unmarshal([]byte(input), &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if doc.Body == nil {
		t.Fatal("Body is nil after unmarshal")
	}
	if len(doc.Body.P) == 0 {
		t.Fatal("Body.P is empty")
	}

	// Marshal through xmlutil.Encoder
	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(&doc); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()

	outStr := buf.String()
	if !strings.Contains(outStr, "<w:document") {
		t.Errorf("output missing w:document, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "<w:body") {
		t.Errorf("output missing w:body, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "<w:p/>") && !strings.Contains(outStr, "<w:p></w:p>") {
		t.Errorf("output missing w:p, got:\n%s", outStr)
	}
}

func TestRoundTrip_Paragraph(t *testing.T) {
	input := `<w:p xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:r><w:t xml:space="preserve">Hello World</w:t></w:r></w:p>`

	var p wml.CT_P
	if err := xml.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(p.R) == 0 || p.R[0].T == nil {
		t.Fatal("expected run with text")
	}

	// Marshal and verify round-trip through xmlutil.Encoder
	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(&p); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()

	outStr := buf.String()
	if !strings.Contains(outStr, "<w:p") {
		t.Errorf("output missing w:p, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "<w:r") {
		t.Errorf("output missing w:r, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "<w:t") {
		t.Errorf("output missing w:t, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "Hello World") {
		t.Errorf("output missing text, got:\n%s", outStr)
	}
}

func TestWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		text  string
		hasSpace bool
	}{
		{"clean text", `<w:t xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">hello</w:t>`, "hello", false},
		{"leading space", `<w:t xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xml:space="preserve">  hello</w:t>`, "  hello", true},
		{"double space", `<w:t xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xml:space="preserve">hello  world</w:t>`, "hello  world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var txt wml.CT_Text
			if err := xml.Unmarshal([]byte(tt.input), &txt); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if txt.Value != tt.text {
				t.Errorf("Value = %q, want %q", txt.Value, tt.text)
			}

			// Marshal through xmlutil.Encoder and check xml:space
			var buf bytes.Buffer
			enc := xmlutil.NewEncoder(&buf)
			if err := enc.Encode(&txt); err != nil {
				t.Fatalf("Encode: %v", err)
			}
			enc.Flush()

			outStr := buf.String()
			if tt.hasSpace && !strings.Contains(outStr, `xml:space="preserve"`) {
				t.Errorf("expected xml:space=\"preserve\", got:\n%s", outStr)
			}
			if !tt.hasSpace && strings.Contains(outStr, `xml:space="preserve"`) {
				t.Errorf("unexpected xml:space=\"preserve\", got:\n%s", outStr)
			}
			// Text should survive round-trip
			if !strings.Contains(outStr, tt.text) {
				t.Errorf("text %q not found in output:\n%s", tt.text, outStr)
			}
		})
	}
}

func TestPrefixAgnostic(t *testing.T) {
	// Non-canonical prefix resolves to same type via URI
	input := `<word:p xmlns:word="http://schemas.openxmlformats.org/wordprocessingml/2006/main"/>`
	var p wml.CT_P
	if err := xml.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("Unmarshal with word: prefix: %v", err)
	}
}

func TestHoard(t *testing.T) {
	input := `<w:pPr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:pStyle w:val="Heading1"/><w14:paraId xmlns:w14="http://schemas.microsoft.com/office/word/2010/wordml" w14:val="12345"/></w:pPr>`

	var ppr wml.CT_PPr
	if err := xml.Unmarshal([]byte(input), &ppr); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if ppr.PStyle == nil || ppr.PStyle.Val == nil || *ppr.PStyle.Val != "Heading1" {
		t.Error("PStyle not properly unmarshaled")
	}
	// Unknown child (w14:paraId) should be hoarded in Raw
	if len(ppr.Raw) == 0 {
		t.Error("Raw hoard is empty - unknown child was lost")
	}

	// Marshal and verify unknown child survives
	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(&ppr); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()

	outStr := buf.String()
	if !strings.Contains(outStr, "paraId") {
		t.Errorf("hoarded child 'paraId' lost in round-trip, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "12345") {
		t.Errorf("hoarded attr '12345' lost in round-trip, got:\n%s", outStr)
	}
}

func TestRoundTrip_Style(t *testing.T) {
	input := `<w:style xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" w:type="paragraph" w:styleId="Heading1">
<w:name w:val="heading 1"/>
<w:basedOn w:val="Normal"/>
<w:next w:val="Normal"/>
<w:link w:val="Heading1Char"/>
<w:qFormat/>
</w:style>`

	var s wml.CT_Style
	if err := xml.Unmarshal([]byte(input), &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if s.StyleID == nil || *s.StyleID != "Heading1" {
		t.Errorf("StyleID = %v, want Heading1", s.StyleID)
	}
	if s.Name == nil || s.Name.Val == nil || *s.Name.Val != "heading 1" {
		t.Errorf("Name.Val = %v, want 'heading 1'", s.Name.Val)
	}
	if s.BasedOn == nil || s.BasedOn.Val == nil || *s.BasedOn.Val != "Normal" {
		t.Errorf("BasedOn.Val = %v, want Normal", s.BasedOn.Val)
	}

	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(&s); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()

	outStr := buf.String()
	if !strings.Contains(outStr, "Heading1") {
		t.Errorf("styleId lost, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, `heading 1`) {
		t.Errorf("style name lost, got:\n%s", outStr)
	}
}

func TestRoundTrip_Numbering(t *testing.T) {
	input := `<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`

	var n wml.CT_Numbering
	if err := xml.Unmarshal([]byte(input), &n); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(&n); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()

	outStr := buf.String()
	if !strings.Contains(outStr, `decimal`) {
		t.Errorf("numFmt lost, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, `%1.`) {
		t.Errorf("lvlText lost, got:\n%s", outStr)
	}
}

func TestRoundTrip_Table(t *testing.T) {
	input := `<w:tbl xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:tblPr><w:tblStyle w:val="TableGrid"/><w:tblW w:w="5000" w:type="pct"/></w:tblPr>
<w:tblGrid><w:gridCol w:w="2500"/><w:gridCol w:w="2500"/></w:tblGrid>
<w:tr><w:tc><w:p/></w:tc><w:tc><w:p/></w:tc></w:tr>
</w:tbl>`

	var tbl wml.CT_Tbl
	if err := xml.Unmarshal([]byte(input), &tbl); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(&tbl); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()

	outStr := buf.String()
	if !strings.Contains(outStr, `TableGrid`) {
		t.Errorf("tblStyle lost, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, `<w:tr`) {
		t.Errorf("table row missing, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, `<w:tc`) {
		t.Errorf("table cell missing, got:\n%s", outStr)
	}
}

func TestRunBoundaries(t *testing.T) {
	hello := "Hello"
	world := "World"
	p := &wml.CT_P{
		R: []*wml.CT_R{
			{T: &wml.CT_Text{Value: hello}},
			{T: &wml.CT_Text{Value: world}},
		},
	}

	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()

	outStr := buf.String()
	if got := strings.Count(outStr, "<w:r"); got != 2 {
		t.Errorf("expected 2 <w:r occurrences, got %d:\n%s", got, outStr)
	}

	var p2 wml.CT_P
	if err := xml.Unmarshal([]byte(outStr), &p2); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(p2.R) != 2 {
		t.Errorf("round-trip run count = %d, want 2", len(p2.R))
	}
}

// ---- LvlOverride round-trip (Pitfall 4 fix) ----

func TestCT_NumLvlOverrideRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		xml   string
		check func(t *testing.T, n *wml.CT_Numbering)
	}{
		{
			name: "no lvlOverride baseline",
			xml: `<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`,
			check: func(t *testing.T, n *wml.CT_Numbering) {
				if len(n.Num) != 1 {
					t.Fatalf("Num count = %d, want 1", len(n.Num))
				}
				if len(n.Num[0].LvlOverride) != 0 {
					t.Errorf("LvlOverride should be empty for baseline")
				}
			},
		},
		{
			name: "lvlOverride with startOverride",
			xml: `<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:num w:numId="1"><w:abstractNumId w:val="0"/><w:lvlOverride w:ilvl="0"><w:startOverride w:val="5"/></w:lvlOverride></w:num>
</w:numbering>`,
			check: func(t *testing.T, n *wml.CT_Numbering) {
				num := n.Num[0]
				if len(num.LvlOverride) != 1 {
					t.Fatalf("LvlOverride count = %d, want 1", len(num.LvlOverride))
				}
				lo := num.LvlOverride[0]
				if lo.ILvl == nil || *lo.ILvl != 0 {
					t.Errorf("ILvl = %v, want 0", lo.ILvl)
				}
				if lo.StartOverride == nil || lo.StartOverride.Val == nil || *lo.StartOverride.Val != 5 {
					t.Errorf("StartOverride.Val = %v, want 5", lo.StartOverride.Val)
				}
				if lo.Lvl != nil {
					t.Errorf("Lvl should be nil for startOverride-only")
				}
			},
		},
		{
			name: "lvlOverride with full lvl replacement",
			xml: `<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:abstractNum w:abstractNumId="0"><w:lvl w:ilvl="1"><w:numFmt w:val="bullet"/><w:lvlText w:val="&#x2022;"/></w:lvl></w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/><w:lvlOverride w:ilvl="1"><w:lvl w:ilvl="1"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/></w:lvl></w:lvlOverride></w:num>
</w:numbering>`,
			check: func(t *testing.T, n *wml.CT_Numbering) {
				num := n.Num[0]
				if len(num.LvlOverride) != 1 {
					t.Fatalf("LvlOverride count = %d, want 1", len(num.LvlOverride))
				}
				lo := num.LvlOverride[0]
				if lo.Lvl == nil {
					t.Fatal("Lvl should not be nil for full lvl replacement")
				}
				if lo.Lvl.NumFmt == nil || lo.Lvl.NumFmt.Val == nil || *lo.Lvl.NumFmt.Val != "decimal" {
					t.Errorf("Override Lvl.NumFmt.Val = %v, want decimal", lo.Lvl.NumFmt)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n wml.CT_Numbering
			dec := xmlutil.NewSafeDecoder(strings.NewReader(tt.xml), 4096)
			if err := dec.Decode(&n); err != nil {
				t.Fatalf("Decode: %v", err)
			}
			tt.check(t, &n)

			// Marshal and verify round-trip
			var buf bytes.Buffer
			enc := xmlutil.NewEncoder(&buf)
			if err := enc.Encode(&n); err != nil {
				t.Fatalf("Encode: %v", err)
			}
			enc.Flush()

			var n2 wml.CT_Numbering
			dec2 := xmlutil.NewSafeDecoder(strings.NewReader(buf.String()), 4096)
			if err := dec2.Decode(&n2); err != nil {
				t.Fatalf("Re-decode after marshal: %v", err)
			}
			if len(n2.Num) != len(n.Num) {
				t.Errorf("Num count mismatch after round-trip: got %d, want %d", len(n2.Num), len(n.Num))
			}
		})
	}
}

// TestRoundTrip_Fields verifies that TOC field characters, instruction
// text (with xml:space preservation), and bookmarks survive a
// decode/encode round trip.
func TestRoundTrip_Fields(t *testing.T) {
	input := `<w:p xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:pPr><w:spacing w:before="240" w:after="120"/></w:pPr><w:bookmarkStart w:id="7" w:name="_Toc00000007"/><w:r><w:fldChar w:fldCharType="begin" w:dirty="true"/></w:r><w:r><w:instrText xml:space="preserve"> TOC \o "1-3" \h \z \u </w:instrText></w:r><w:r><w:fldChar w:fldCharType="separate"/></w:r><w:r><w:t>Table of Contents</w:t></w:r><w:bookmarkEnd w:id="7"/></w:p>`

	var p wml.CT_P
	if err := xml.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(p.BookmarkStart) != 1 || p.BookmarkStart[0].ID == nil || *p.BookmarkStart[0].ID != 7 ||
		p.BookmarkStart[0].Name == nil || *p.BookmarkStart[0].Name != "_Toc00000007" {
		t.Errorf("bookmarkStart = %+v", p.BookmarkStart)
	}
	if len(p.BookmarkEnd) != 1 || p.BookmarkEnd[0].ID == nil || *p.BookmarkEnd[0].ID != 7 {
		t.Errorf("bookmarkEnd = %+v", p.BookmarkEnd)
	}
	if len(p.R) != 4 {
		t.Fatalf("got %d runs, want 4", len(p.R))
	}
	begin := p.R[0].FldChar
	if begin == nil || begin.Type == nil || *begin.Type != "begin" || begin.Dirty == nil || *begin.Dirty != "true" {
		t.Errorf("run[0] fldChar = %+v", begin)
	}
	if it := p.R[1].InstrText; it == nil || it.Value != ` TOC \o "1-3" \h \z \u ` {
		t.Errorf("run[1] instrText = %+v", p.R[1].InstrText)
	}
	if sep := p.R[2].FldChar; sep == nil || sep.Type == nil || *sep.Type != "separate" {
		t.Errorf("run[2] fldChar = %+v", sep)
	}

	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(&p); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()
	out := buf.String()
	for _, marker := range []string{
		`<w:bookmarkStart w:id="7" w:name="_Toc00000007">`,
		`w:fldCharType="begin" w:dirty="true"`,
		`xml:space="preserve"> TOC \o &#34;1-3&#34; \h \z \u `,
		`<w:bookmarkEnd w:id="7">`,
	} {
		if !strings.Contains(out, marker) {
			t.Errorf("output missing %q:\n%s", marker, out)
		}
	}
}

// TestRoundTrip_UpdateFields verifies the settings.xml updateFields
// element decodes and re-encodes.
func TestRoundTrip_UpdateFields(t *testing.T) {
	input := `<w:settings xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:zoom w:percent="100"/><w:updateFields w:val="true"/></w:settings>`
	var s wml.CT_Settings
	if err := xml.Unmarshal([]byte(input), &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if s.UpdateFields == nil || s.UpdateFields.Val == nil || *s.UpdateFields.Val != "true" {
		t.Fatalf("UpdateFields = %+v", s.UpdateFields)
	}
	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(&s); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()
	if out := buf.String(); !strings.Contains(out, `w:updateFields w:val="true"`) {
		t.Errorf("output missing updateFields:\n%s", out)
	}
}
