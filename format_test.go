package wordingo

import (
	"bytes"
	"math"
	"testing"

	"github.com/fabiomarini/wordingo/internal/wml"
)

func boolPtr(v bool) *bool          { return &v }
func testStrPtr(v string) *string   { return &v }
func float64Ptr(v float64) *float64 { return &v }

func TestAlignmentString(t *testing.T) {
	tests := []struct {
		a    Alignment
		want string
	}{
		{AlignmentLeft, "left"},
		{AlignmentCenter, "center"},
		{AlignmentRight, "right"},
		{AlignmentBoth, "both"},
		{Alignment(-1), ""},
	}
	for _, tt := range tests {
		got := tt.a.String()
		if got != tt.want {
			t.Errorf("Alignment(%d).String() = %q, want %q", tt.a, got, tt.want)
		}
	}
}

func TestRunFormatting(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("test")
	r := p.AddRun("hello")

	t.Run("SetBold", func(t *testing.T) {
		r.SetBold(true)
		if r.ct.RPr == nil || r.ct.RPr.B == nil || r.ct.RPr.B.Val == nil || !*r.ct.RPr.B.Val {
			t.Error("SetBold(true) did not set B.Val=true")
		}
		r.SetBold(false)
		if r.ct.RPr.B == nil || r.ct.RPr.B.Val == nil || *r.ct.RPr.B.Val {
			t.Error("SetBold(false) did not set B.Val=false")
		}
	})

	t.Run("SetItalic", func(t *testing.T) {
		r.SetItalic(true)
		if r.ct.RPr == nil || r.ct.RPr.I == nil || r.ct.RPr.I.Val == nil || !*r.ct.RPr.I.Val {
			t.Error("SetItalic(true) did not set I.Val=true")
		}
	})

	t.Run("SetUnderline", func(t *testing.T) {
		r.SetUnderline("single")
		if r.ct.RPr == nil || r.ct.RPr.U == nil || r.ct.RPr.U.Val == nil || *r.ct.RPr.U.Val != "single" {
			t.Error("SetUnderline did not set U.Val")
		}
	})

	t.Run("SetFont", func(t *testing.T) {
		r.SetFont("Arial")
		if r.ct.RPr == nil || r.ct.RPr.RFonts == nil {
			t.Fatal("RFonts not set")
		}
		if r.ct.RPr.RFonts.Ascii == nil || *r.ct.RPr.RFonts.Ascii != "Arial" {
			t.Error("SetFont did not set Ascii")
		}
		if r.ct.RPr.RFonts.HAnsi == nil || *r.ct.RPr.RFonts.HAnsi != "Arial" {
			t.Error("SetFont did not set HAnsi")
		}
	})

	t.Run("SetSize", func(t *testing.T) {
		r.SetSize(12)
		if r.ct.RPr == nil || r.ct.RPr.Sz == nil || r.ct.RPr.Sz.Val == nil {
			t.Fatal("Sz not set")
		}
		want := int64(math.Round(12 * 2))
		if *r.ct.RPr.Sz.Val != want {
			t.Errorf("SetSize(12) = %d half-pts, want %d", *r.ct.RPr.Sz.Val, want)
		}

		r.SetSize(12.5)
		want = int64(math.Round(12.5 * 2))
		if *r.ct.RPr.Sz.Val != want {
			t.Errorf("SetSize(12.5) = %d half-pts, want %d", *r.ct.RPr.Sz.Val, want)
		}
	})

	t.Run("SetColor", func(t *testing.T) {
		r.SetColor("FF0000")
		if r.ct.RPr == nil || r.ct.RPr.Color == nil || r.ct.RPr.Color.Val == nil {
			t.Fatal("Color not set")
		}
		if *r.ct.RPr.Color.Val != "FF0000" {
			t.Errorf("Color.Val = %q, want FF0000", *r.ct.RPr.Color.Val)
		}
	})

	t.Run("SetHighlight", func(t *testing.T) {
		r.SetHighlight("yellow")
		if r.ct.RPr == nil || r.ct.RPr.Highlight == nil || r.ct.RPr.Highlight.Val == nil {
			t.Fatal("Highlight not set")
		}
		if *r.ct.RPr.Highlight.Val != "yellow" {
			t.Errorf("Highlight.Val = %q, want yellow", *r.ct.RPr.Highlight.Val)
		}
	})

	t.Run("SetStyle", func(t *testing.T) {
		r.SetStyle("Emphasis")
		if r.ct.RPr == nil || r.ct.RPr.RStyle == nil || r.ct.RPr.RStyle.Val == nil {
			t.Fatal("RStyle not set")
		}
		if *r.ct.RPr.RStyle.Val != "Emphasis" {
			t.Errorf("RStyle.Val = %q, want Emphasis", *r.ct.RPr.RStyle.Val)
		}
	})

	t.Run("X", func(t *testing.T) {
		if r.X() != r.ct {
			t.Error("X() should return *wml.CT_R")
		}
	})
}

func TestParagraphFormatting(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("test")

	t.Run("SetAlignment", func(t *testing.T) {
		p.SetAlignment(AlignmentCenter)
		if p.ct.PPr == nil || p.ct.PPr.Jc == nil || p.ct.PPr.Jc.Val == nil {
			t.Fatal("Jc not set")
		}
		if *p.ct.PPr.Jc.Val != "center" {
			t.Errorf("Jc.Val = %q, want center", *p.ct.PPr.Jc.Val)
		}

		p.SetAlignment(AlignmentBoth)
		if *p.ct.PPr.Jc.Val != "both" {
			t.Errorf("Jc.Val = %q, want both", *p.ct.PPr.Jc.Val)
		}
	})

	t.Run("SetAlignmentUnknown", func(t *testing.T) {
		before := len(doc.warnings)
		p.SetAlignment(Alignment(-1))
		if len(doc.warnings) <= before {
			t.Error("expected warning for unknown alignment")
		}
	})

	t.Run("SetSpacing", func(t *testing.T) {
		p.SetSpacing(&ParSpacing{Before: 240, After: 120})
		if p.ct.PPr == nil || p.ct.PPr.Spacing == nil {
			t.Fatal("Spacing not set")
		}
		if p.ct.PPr.Spacing.Before == nil || *p.ct.PPr.Spacing.Before != 240 {
			t.Error("Spacing.Before != 240")
		}
		if p.ct.PPr.Spacing.After == nil || *p.ct.PPr.Spacing.After != 120 {
			t.Error("Spacing.After != 120")
		}
	})

	t.Run("SetSpacingNil", func(t *testing.T) {
		before := len(doc.warnings)
		p.SetSpacing(nil)
		if len(doc.warnings) <= before {
			t.Error("expected warning for nil ParSpacing")
		}
	})

	t.Run("SetIndent", func(t *testing.T) {
		p.SetIndent(&ParIndent{Left: 720})
		if p.ct.PPr == nil || p.ct.PPr.Ind == nil {
			t.Fatal("Ind not set")
		}
		if p.ct.PPr.Ind.Left == nil || *p.ct.PPr.Ind.Left != 720 {
			t.Error("Ind.Left != 720")
		}
	})

	t.Run("SetIndentNil", func(t *testing.T) {
		before := len(doc.warnings)
		p.SetIndent(nil)
		if len(doc.warnings) <= before {
			t.Error("expected warning for nil ParIndent")
		}
	})
}

func TestBuilderChain(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("chain")
	r := p.AddRun("a").SetBold(true).SetSize(12)
	if r.ct.RPr == nil || r.ct.RPr.B == nil || r.ct.RPr.B.Val == nil || !*r.ct.RPr.B.Val {
		t.Error("builder chain: SetBold not applied")
	}
	if r.ct.RPr.Sz == nil || r.ct.RPr.Sz.Val == nil || *r.ct.RPr.Sz.Val != int64(math.Round(12*2)) {
		t.Error("builder chain: SetSize not applied")
	}

	r2 := p.AddRun("b").SetFormatting(RunFormat{Bold: boolPtr(true), Color: testStrPtr("FF0000")})
	if r2.ct.RPr == nil || r2.ct.RPr.B == nil || r2.ct.RPr.B.Val == nil || !*r2.ct.RPr.B.Val {
		t.Error("SetFormatting: SetBold not applied")
	}
	if r2.ct.RPr.Color == nil || r2.ct.RPr.Color.Val == nil || *r2.ct.RPr.Color.Val != "FF0000" {
		t.Error("SetFormatting: SetColor not applied")
	}
}

func TestBulkSetFormatting(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("bulk")
	align := AlignmentCenter
	p.SetFormatting(ParFormat{
		Alignment: &align,
		Spacing:   &ParSpacing{Before: 240},
	})
	if p.ct.PPr.Jc == nil || *p.ct.PPr.Jc.Val != "center" {
		t.Error("ParFormat.SetFormatting: alignment not set")
	}
	if p.ct.PPr.Spacing == nil || p.ct.PPr.Spacing.Before == nil || *p.ct.PPr.Spacing.Before != 240 {
		t.Error("ParFormat.SetFormatting: spacing not set")
	}

	p2 := doc.AddParagraph("nil-fields")
	p2.SetFormatting(ParFormat{})
	if p2.ct.PPr != nil {
		t.Error("ParFormat with zero-value fields should not init PPr")
	}
}

func TestFormatWarnings(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("warn")
	r := p.AddRun("test")

	r.SetSize(-12)
	r.SetColor("XYZ123")
	p.SetAlignment(Alignment(-1))

	warns := doc.Warnings()
	if len(warns) == 0 {
		t.Error("expected warnings for invalid inputs")
	}
}

func TestAddParagraph(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("hello")
	if p == nil {
		t.Fatal("AddParagraph returned nil")
	}
	if len(doc.doc.Body.P) != 2 {
		t.Errorf("body has %d paragraphs, want 2", len(doc.doc.Body.P))
	}
	if p.Text() != "hello" {
		t.Errorf("Text() = %q, want %q", p.Text(), "hello")
	}
	if !doc.dirty {
		t.Error("AddParagraph should set dirty=true")
	}
}

func TestAddParagraphEmpty(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("")
	if len(doc.doc.Body.P) != 2 {
		t.Errorf("body has %d paragraphs, want 2", len(doc.doc.Body.P))
	}
	if len(doc.doc.Body.P[1].R) != 0 {
		t.Error("empty AddParagraph should have no runs")
	}
}

func TestAddParagraphDirty(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("test")
	if !doc.dirty {
		t.Error("AddParagraph should set dirty")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if doc.dirty {
		t.Error("dirty should be false after WriteTo")
	}
}

func TestSerializeBodyNoop(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if doc.dirty {
		t.Error("dirty should be false after WriteTo on non-dirty doc")
	}
}

func TestDocWarnings(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.warn("test warning %d", 1)
	w := doc.Warnings()
	found := false
	for _, msg := range w {
		if msg == "test warning 1" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Warnings() did not include doc warning: got %v", w)
	}
}

func TestNilStatePanic(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{"nil Document.AddParagraph", func() { (*Document)(nil).AddParagraph("x") }},
		{"nil Document.WriteTo", func() { (*Document)(nil).WriteTo(nil) }},
		{"nil Document.Warnings", func() { (*Document)(nil).Warnings() }},
		{"nil Paragraph.AddRun", func() { (*Paragraph)(nil).AddRun("x") }},
		{"nil Paragraph.SetAlignment", func() { (*Paragraph)(nil).SetAlignment(AlignmentLeft) }},
		{"nil Paragraph.SetStyle", func() { (*Paragraph)(nil).SetStyle("x") }},
		{"nil Paragraph.X", func() { (*Paragraph)(nil).X() }},
		{"nil Run.SetBold", func() { (*Run)(nil).SetBold(true) }},
		{"nil Run.X", func() { (*Run)(nil).X() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic, got none")
				}
			}()
			tt.fn()
		})
	}
}

func TestBuilderChainIntegration(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	p := doc.AddParagraph("Hello")
	p.AddRun(" World").SetBold(true).SetSize(14)
	p.SetAlignment(AlignmentCenter)
	p.SetSpacing(&ParSpacing{Before: 120, After: 60})

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	doc2, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	paras := doc2.Paragraphs()
	if len(paras) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(paras))
	}
	if paras[1].Text() != "Hello World" {
		t.Errorf("text = %q, want %q", paras[1].Text(), "Hello World")
	}
}

func TestSetStyleChain(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	p := doc.AddParagraph("styled")
	p.SetStyle("Heading1").SetAlignment(AlignmentCenter)
	if p.ct.PPr.PStyle == nil || p.ct.PPr.PStyle.Val == nil || *p.ct.PPr.PStyle.Val != "Heading1" {
		t.Error("SetStyle not applied in chain")
	}
	if p.ct.PPr.Jc == nil || *p.ct.PPr.Jc.Val != "center" {
		t.Error("SetAlignment not applied after SetStyle in chain")
	}
}

func TestAddRunExistingParagraph(t *testing.T) {
	data := buildSyntheticFixture(t, string(fixtureSingleParaXML(t)), nil)
	doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}

	paras := doc.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("expected 1 paragraph, got %d", len(paras))
	}

	r := paras[0].AddRun(" added")
	if r == nil {
		t.Fatal("AddRun returned nil")
	}
	if paras[0].Text() != "Hello World added" {
		t.Errorf("text = %q, want %q", paras[0].Text(), "Hello World added")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	pkg, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	reparas := pkg.Paragraphs()
	if len(reparas) != 1 {
		t.Fatalf("re-opened: expected 1 paragraph, got %d", len(reparas))
	}
	if reparas[0].Text() != "Hello World added" {
		t.Errorf("re-opened text = %q, want %q", reparas[0].Text(), "Hello World added")
	}
}

func TestAddParagraphNilBody(t *testing.T) {
	doc := &Document{doc: &wml.CT_Document{}}
	p := doc.AddParagraph("test")
	if p == nil {
		t.Fatal("AddParagraph on nil body returned nil")
	}
	if doc.doc.Body == nil {
		t.Fatal("Body should have been lazily initialized")
	}
	if len(doc.doc.Body.P) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(doc.doc.Body.P))
	}
}

func TestRoundTripSaveOpen(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	p := doc.AddParagraph("Round trip test")
	p.AddRun(" bold").SetBold(true)
	p.SetAlignment(AlignmentBoth)
	p.SetSpacing(&ParSpacing{Before: 240})

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	paras := reopened.Paragraphs()
	if len(paras) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(paras))
	}
	if paras[1].Text() != "Round trip test bold" {
		t.Errorf("text = %q, want %q", paras[1].Text(), "Round trip test bold")
	}
}
