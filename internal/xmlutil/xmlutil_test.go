package xmlutil_test

import (
	"bytes"
	"encoding/xml"
	"errors"
	"strings"
	"testing"

	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// ---- Registry tests ----

func TestPrefixFor_Transitional(t *testing.T) {
	tests := []struct {
		uri   string
		want  string
		label string
	}{
		{"http://schemas.openxmlformats.org/wordprocessingml/2006/main", "w", "WML main"},
		{"http://schemas.openxmlformats.org/officeDocument/2006/relationships", "r", "relationships"},
		{"http://schemas.openxmlformats.org/drawingml/2006/main", "a", "DrawingML"},
		{"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing", "wp", "wordprocessingDrawing"},
		{"http://schemas.openxmlformats.org/markup-compatibility/2006", "mc", "markup-compat"},
		{"http://schemas.microsoft.com/office/word/2010/wordml", "w14", "w14"},
		{"http://schemas.microsoft.com/office/word/2012/wordml", "w15", "w15"},
		{"http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing", "wp14", "wp14"},
		{"http://www.w3.org/XML/1998/namespace", "xml", "xml namespace"},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			got := xmlutil.PrefixFor(tt.uri)
			if got != tt.want {
				t.Errorf("PrefixFor(%q) = %q, want %q", tt.uri, got, tt.want)
			}
		})
	}
}

func TestPrefixFor_Strict(t *testing.T) {
	if got := xmlutil.PrefixFor("http://purl.oclc.org/ooxml/wordprocessingml/main"); got != "w" {
		t.Errorf("Strict WML main = %q, want %q", got, "w")
	}
	if got := xmlutil.PrefixFor("http://purl.oclc.org/ooxml/officeDocument/relationships"); got != "r" {
		t.Errorf("Strict relationships = %q, want %q", got, "r")
	}
	if got := xmlutil.PrefixFor("http://purl.oclc.org/ooxml/drawingml/main"); got != "a" {
		t.Errorf("Strict DrawingML = %q, want %q", got, "a")
	}
}

func TestPrefixFor_Unknown(t *testing.T) {
	if got := xmlutil.PrefixFor("urn:unknown"); got != "" {
		t.Errorf("unknown URI = %q, want empty", got)
	}
}

func TestNormalizeURI(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http://purl.oclc.org/ooxml/wordprocessingml/main",
			"http://schemas.openxmlformats.org/wordprocessingml/2006/main"},
		{"http://purl.oclc.org/ooxml/officeDocument/relationships",
			"http://schemas.openxmlformats.org/officeDocument/2006/relationships"},
		{"http://schemas.openxmlformats.org/wordprocessingml/2006/main",
			"http://schemas.openxmlformats.org/wordprocessingml/2006/main"},
		{"http://www.w3.org/XML/1998/namespace",
			"http://www.w3.org/XML/1998/namespace"},
	}
	for i, tt := range tests {
		got := xmlutil.NormalizeURI(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeURI[%d] = %q, want %q", i, got, tt.want)
		}
	}
}

// ---- SafeDecoder tests ----

func TestSafeDecoder_ValidXML(t *testing.T) {
	input := `<root><child attr="val">text</child></root>`
	d := xmlutil.NewSafeDecoder(strings.NewReader(input), 1<<20)
	var out struct {
		XMLName xml.Name `xml:"root"`
	}
	if err := d.Decode(&out); err != nil {
		t.Fatalf("valid xml: %v", err)
	}
}

func TestSafeDecoder_RejectsDOCTYPE(t *testing.T) {
	input := `<!DOCTYPE foo [<!ENTITY x "y">]><root/>`
	d := xmlutil.NewSafeDecoder(strings.NewReader(input), 1<<20)
	var v struct{ XMLName xml.Name }
	err := d.Decode(&v)
	if err == nil {
		t.Fatal("expected DOCTYPE rejection error, got nil")
	}
	if !errors.Is(err, xmlutil.ErrDOCTYPE) {
		t.Fatalf("expected ErrDOCTYPE, got %v", err)
	}
}

func TestSafeDecoder_DepthLimit(t *testing.T) {
	// Build 600-deep nesting via repeated elements
	var buf bytes.Buffer
	for i := 0; i < 600; i++ {
		buf.WriteString("<a>")
	}
	for i := 0; i < 600; i++ {
		buf.WriteString("</a>")
	}

	d := xmlutil.NewSafeDecoder(strings.NewReader(buf.String()), 1<<20)
	for {
		_, err := d.Token()
		if err != nil {
			if errors.Is(err, xmlutil.ErrXMLDepth) {
				return // success
			}
			t.Fatalf("expected ErrXMLDepth, got %v", err)
		}
	}
}

// ---- RawXML tests ----

type rawXMLTestContainer struct {
	XMLName xml.Name         `xml:"root"`
	Content []xmlutil.RawXML `xml:",any"`
}

func TestRawXML_RoundTrip(t *testing.T) {
	input := `<root xmlns:foo="urn:x"><foo:a foo:b="1" foo:c="2"><foo:d/></foo:a></root>`

	var c rawXMLTestContainer
	if err := xml.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	// Verify captured token count and attribute order
	if len(c.Content) != 1 {
		t.Fatalf("expected 1 RawXML content, got %d", len(c.Content))
	}

	// Marshal through standard encoder (loses non-canonical prefixes
	// but preserves namespace URIs, attribute order, and structure)
	out, err := xml.Marshal(&c)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	outStr := string(out)
	// Structure preserved: a element with correct namespace
	if !strings.Contains(outStr, `xmlns="urn:x"`) {
		t.Errorf("output should declare urn:x namespace, got:\n%s", outStr)
	}
	// Both attributes present in original order
	abIdx := strings.Index(outStr, `b="1"`)
	acIdx := strings.Index(outStr, `c="2"`)
	if abIdx < 0 || acIdx < 0 {
		t.Errorf("output missing expected attributes, got:\n%s", outStr)
	}
	if abIdx > acIdx {
		t.Errorf("attribute order changed: b after c, was b before c in input")
	}
	// Nested content preserved
	if !strings.Contains(outStr, `d`) {
		t.Errorf("output missing nested <d>, got:\n%s", outStr)
	}
}

// ---- Encoder tests ----

type encoderTestPara struct {
	XMLName xml.Name         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	R       *encoderTestRun  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type encoderTestRun struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
	T       *encoderTestText `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t"`
}

type encoderTestText struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t"`
	Value   string   `xml:",chardata"`
}

func TestEncoder_CanonicalPrefix(t *testing.T) {
	v := encoderTestPara{
		R: &encoderTestRun{
			T: &encoderTestText{Value: "hello"},
		},
	}

	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	output := buf.String()
	t.Logf("Encoder output:\n%s", output)

	// Must use canonical <w:p prefix form
	if !strings.Contains(output, "<w:p") {
		t.Errorf("output should contain '<w:p', got:\n%s", output)
	}
	// Must NOT contain default-namespace form for WML main
	if strings.Contains(output, `xmlns="http://schemas.openxmlformats.org/wordprocessingml/2006/main"`) {
		t.Errorf("output should NOT contain default xmlns= form for WML main, got:\n%s", output)
	}
	// Child elements should also use canonical prefixes
	if !strings.Contains(output, "<w:r") {
		t.Errorf("output should contain '<w:r', got:\n%s", output)
	}
	if !strings.Contains(output, "<w:t") {
		t.Errorf("output should contain '<w:t', got:\n%s", output)
	}
}

func TestEncoder_XMLNSDeclaredOnce(t *testing.T) {
	v := encoderTestPara{
		R: &encoderTestRun{
			T: &encoderTestText{Value: "world"},
		},
	}

	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	enc.Flush()

	output := buf.String()
	// xmlns:w should appear exactly once
	count := strings.Count(output, `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"`)
	if count != 1 {
		t.Errorf("xmlns:w should appear exactly once, got %d occurrences:\n%s", count, output)
	}
}

// encoderTestParaExt is a paragraph whose pPr carries a raw w14
// extension element (like w14:paraId in real Word output), forcing the
// encoder to declare w14 and emit mc:Ignorable.
type encoderTestParaExt struct {
	XMLName xml.Name           `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	PPr     *encoderTestPPrExt `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	R       *encoderTestRun    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type encoderTestPPrExt struct {
	XMLName xml.Name         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	Raw     []xmlutil.RawXML `xml:",any"`
}

func TestEncoder_MCIgnorableDeclaresMC(t *testing.T) {
	// Rebuild a w14:paraId element the way decoding a Word document
	// captures it: namespace-qualified name, no prefix.
	var raw xmlutil.RawXML
	{
		var buf bytes.Buffer
		enc := xml.NewEncoder(&buf)
		start := xml.StartElement{
			Name: xml.Name{Space: "http://schemas.microsoft.com/office/word/2010/wordml", Local: "paraId"},
			Attr: []xml.Attr{{
				Name:  xml.Name{Space: "http://schemas.microsoft.com/office/word/2010/wordml", Local: "val"},
				Value: "6B4DFA3C",
			}},
		}
		if err := enc.EncodeToken(start); err != nil {
			t.Fatal(err)
		}
		if err := enc.EncodeToken(start.End()); err != nil {
			t.Fatal(err)
		}
		if err := enc.Flush(); err != nil {
			t.Fatal(err)
		}
		dec := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
		if err := dec.Decode(&raw); err != nil {
			t.Fatal(err)
		}
	}

	v := encoderTestParaExt{PPr: &encoderTestPPrExt{Raw: []xmlutil.RawXML{raw}}}
	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	output := buf.String()

	// mc:Ignorable must be accompanied by an xmlns:mc declaration,
	// otherwise the output is not well-formed XML (unbound prefix).
	if !strings.Contains(output, `mc:Ignorable="w14"`) {
		t.Errorf("output missing mc:Ignorable=\"w14\":\n%s", output)
	}
	if !strings.Contains(output, `xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006"`) {
		t.Errorf("output missing xmlns:mc declaration:\n%s", output)
	}

	// Full well-formedness check.
	dec := xml.NewDecoder(strings.NewReader(output))
	for {
		if _, err := dec.Token(); err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("output is not well-formed XML: %v\n%s", err, output)
		}
	}
}

func TestEncoder_NoDuplicateMCDeclaration(t *testing.T) {
	// A paragraph with an mc:AlternateContent child: mc is then used as
	// an element prefix, so it is declared via the canonical loop and
	// must not be declared a second time for mc:Ignorable.
	var alt xmlutil.RawXML
	{
		var buf bytes.Buffer
		enc := xml.NewEncoder(&buf)
		start := xml.StartElement{
			Name: xml.Name{Space: "http://schemas.openxmlformats.org/markup-compatibility/2006", Local: "AlternateContent"},
		}
		if err := enc.EncodeToken(start); err != nil {
			t.Fatal(err)
		}
		if err := enc.EncodeToken(start.End()); err != nil {
			t.Fatal(err)
		}
		if err := enc.Flush(); err != nil {
			t.Fatal(err)
		}
		dec := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
		if err := dec.Decode(&alt); err != nil {
			t.Fatal(err)
		}
	}

	v := encoderTestParaExt{
		PPr: &encoderTestPPrExt{Raw: []xmlutil.RawXML{alt}},
		R:   &encoderTestRun{T: &encoderTestText{Value: "x"}},
	}
	var buf bytes.Buffer
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		t.Fatal(err)
	}
	if err := enc.Flush(); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	count := strings.Count(output, `xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006"`)
	if count != 1 {
		t.Errorf("xmlns:mc should appear exactly once, got %d:\n%s", count, output)
	}
}

func TestEncoder_DeterministicOutput(t *testing.T) {
	// Re-encoded output must be byte-identical across runs even though
	// xmlns declaration order comes from map iteration internally.
	v := encoderTestParaExt{
		PPr: &encoderTestPPrExt{Raw: []xmlutil.RawXML{rawW14ParaID(t)}},
		R:   &encoderTestRun{T: &encoderTestText{Value: "deterministic"}},
	}
	encode := func() string {
		var buf bytes.Buffer
		enc := xmlutil.NewEncoder(&buf)
		if err := enc.Encode(v); err != nil {
			t.Fatal(err)
		}
		if err := enc.Flush(); err != nil {
			t.Fatal(err)
		}
		return buf.String()
	}
	first := encode()
	for i := 0; i < 50; i++ {
		if got := encode(); got != first {
			t.Fatalf("iteration %d: output differs (nondeterministic encoding)", i)
		}
	}
}

// rawW14ParaID returns a RawXML capture of a w14:paraId element.
func rawW14ParaID(t *testing.T) xmlutil.RawXML {
	t.Helper()
	var raw xmlutil.RawXML
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	start := xml.StartElement{
		Name: xml.Name{Space: "http://schemas.microsoft.com/office/word/2010/wordml", Local: "paraId"},
		Attr: []xml.Attr{{
			Name:  xml.Name{Space: "http://schemas.microsoft.com/office/word/2010/wordml", Local: "val"},
			Value: "6B4DFA3C",
		}},
	}
	if err := enc.EncodeToken(start); err != nil {
		t.Fatal(err)
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		t.Fatal(err)
	}
	if err := enc.Flush(); err != nil {
		t.Fatal(err)
	}
	dec := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := dec.Decode(&raw); err != nil {
		t.Fatal(err)
	}
	return raw
}
