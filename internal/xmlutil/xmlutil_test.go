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
		uri    string
		want   string
		label  string
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
		{"http://schemas.openxmlformats.org/package/2006/relationships", "", "package rels (not in canonical)"},
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
	for {
		tok, err := d.Token()
		if err != nil {
			t.Fatalf("unexpected error before DOCTYPE: %v", err)
		}
		if err := xmlutil.RejectDirective(tok); err != nil {
			if !errors.Is(err, xmlutil.ErrDOCTYPE) {
				t.Fatalf("expected ErrDOCTYPE, got %v", err)
			}
			return // success
		}
	}
}

func TestSafeDecoder_DepthLimit(t *testing.T) {
	// Build 600-deep nesting
	var buf bytes.Buffer
	buf.WriteString("<a>")
	for i := 0; i < 600; i++ {
		buf.WriteString("<b>")
	}
	for i := 0; i < 600; i++ {
		buf.WriteString("</b>")
	}
	buf.WriteString("</a>")

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
	XMLName xml.Name          `xml:"root"`
	Content []xmlutil.RawXML  `xml:",any"`
}

func TestRawXML_RoundTrip(t *testing.T) {
	input := `<root xmlns:foo="urn:x"><foo:a foo:b="1" foo:c="2"><foo:d/></foo:a></root>`

	var c rawXMLTestContainer
	if err := xml.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	out, err := xml.Marshal(&c)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Verify both occurrences of foo:a are present (during capture and replay)
	if !strings.Contains(string(out), `foo:a`) {
		t.Errorf("output missing 'foo:a', got: %s", out)
	}
	if !strings.Contains(string(out), `foo:b="1"`) {
		t.Errorf("output missing attr b=\"1\", got: %s", out)
	}
	if !strings.Contains(string(out), `foo:c="2"`) {
		t.Errorf("output missing attr c=\"2\", got: %s", out)
	}
}

// ---- Encoder tests ----

type encoderTestBlock struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
	Text    string   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t,chardata"`
}

func TestEncoder_CanonicalPrefix(t *testing.T) {
	v := encoderTestBlock{Text: "hello"}

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

	// Must start with canonical <w:p form
	if !strings.Contains(output, "<w:p") {
		t.Errorf("output should contain '<w:p', got:\n%s", output)
	}
	// Must NOT contain default-namespace form
	if strings.Contains(output, `xmlns="http://schemas.openxmlformats.org/wordprocessingml/2006/main"`) {
		t.Errorf("output should NOT contain default xmlns= form for WML main, got:\n%s", output)
	}
}

func TestEncoder_XMLNSDeclaredOnce(t *testing.T) {
	v := encoderTestBlock{Text: "world"}

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
