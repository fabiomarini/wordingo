package wordingo

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func encodePNG(t *testing.T, size int, c color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for x := range size {
		for y := range size {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// logoPNG / bodyPNG: two visually distinct deterministic PNGs.
func logoPNG(t *testing.T) []byte { return encodePNG(t, 4, color.RGBA{R: 200, A: 255}) }

func bodyPNG(t *testing.T) []byte { return encodePNG(t, 6, color.RGBA{B: 99, A: 255}) }

// buildBrandedTemplate is the realistic base template REFACTOR-HARNESS
// §3.3 describes: example content + a header carrying a LOGO + a footer
// with a placeholder.
func buildBrandedTemplate(t *testing.T) []byte {
	t.Helper()
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	hdr := doc.AddHeader(HeaderDefault)
	hdr.AddParagraph("GN Techonomy — Confidential")
	if _, err := hdr.AddImageBytes("logo.png", logoPNG(t), "image/png"); err != nil {
		t.Fatalf("header logo: %v", err)
	}
	ftr := doc.AddFooter(FooterDefault)
	ftr.AddParagraph("Prepared for {{customer_name}}")
	doc.AddParagraph("EXAMPLE CONTENT — must never reach the output").SetStyle("Title")
	doc.AddParagraph("Prepared for {{customer_name}}")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// letterheadMedia counts media parts and reports whether any header XML
// references an image.
func letterheadMedia(t *testing.T, data []byte) (media int, headerHasBlip bool) {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		switch {
		case strings.HasPrefix(f.Name, "word/media/"):
			media++
		case strings.HasPrefix(f.Name, "word/header") && strings.HasSuffix(f.Name, ".xml"):
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			var b bytes.Buffer
			if _, err := b.ReadFrom(rc); err != nil {
				t.Fatal(err)
			}
			rc.Close()
			if strings.Contains(b.String(), "a:blip") {
				headerHasBlip = true
			}
		}
	}
	return media, headerHasBlip
}

// TestFromTemplateKeepsLetterheadWithLogo is the REFACTOR-HARNESS §3.3
// requirement made executable: header/footer + LOGO survive
// FromTemplate, while the template's example body content does NOT.
func TestFromTemplateKeepsLetterheadWithLogo(t *testing.T) {
	src := buildBrandedTemplate(t)

	// Precondition: the source template has header media.
	if n, blip := letterheadMedia(t, src); n == 0 || !blip {
		t.Fatalf("fixture lacks header logo: media=%d blip=%v", n, blip)
	}

	doc, err := FromTemplateReader(bytes.NewReader(src), int64(len(src)))
	if err != nil {
		t.Fatalf("FromTemplate: %v", err)
	}
	doc.AddParagraph("Generated chapter text")
	var out bytes.Buffer
	if _, err := doc.WriteTo(&out); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	rendered := out.Bytes()

	// (1) the template's example body content is gone.
	od, err := OpenReader(bytes.NewReader(rendered), int64(len(rendered)))
	if err != nil {
		t.Fatalf("reopen output: %v", err)
	}
	defer od.Close()
	text, err := od.ExtractText(nil)
	if err != nil {
		t.Fatalf("ExtractText: %v", err)
	}
	if strings.Contains(text, "EXAMPLE CONTENT") {
		t.Error("template example body content leaked into the output")
	}
	if !strings.Contains(text, "Generated chapter text") {
		t.Error("generated content missing")
	}

	// (2) header/footer parts + logo media survived.
	n, blip := letterheadMedia(t, rendered)
	if n == 0 {
		t.Error("no media part in output — the logo was dropped")
	}
	if !blip {
		t.Error("header XML has no image reference — letterhead not cloned")
	}
	if hdr := extractPart(t, rendered, "word/header1.xml"); !strings.Contains(hdr, "Confidential") {
		t.Errorf("header text missing from output: %s", hdr)
	}
	if ftr := extractPart(t, rendered, "word/footer1.xml"); !strings.Contains(ftr, "{{customer_name}}") {
		t.Errorf("footer placeholder missing from output: %s", ftr)
	}

	// (3) the output's sectPr carries a headerReference.
	if docXML := extractPart(t, rendered, "word/document.xml"); !strings.Contains(docXML, "headerReference") {
		t.Error("output sectPr lost headerReference")
	}

	// (4) Merge fills the cloned footer placeholder.
	od.Merge(map[string]string{"customer_name": "Acme S.p.A."}, nil)
	var merged bytes.Buffer
	if _, err := od.WriteTo(&merged); err != nil {
		t.Fatal(err)
	}
	if ftr := extractPart(t, merged.Bytes(), "word/footer1.xml"); !strings.Contains(ftr, "Acme S.p.A.") {
		t.Errorf("footer placeholder not merged: %s", ftr)
	}
}

// TestFromTemplateCountersRespectClonedLetterhead: AddImage on a
// FromTemplate result must not overwrite the cloned logo part.
func TestFromTemplateCountersRespectClonedLetterhead(t *testing.T) {
	src := buildBrandedTemplate(t)
	logoBefore := extractPartRaw(t, src, "word/media/image1.png")

	doc, err := FromTemplateReader(bytes.NewReader(src), int64(len(src)))
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("body")
	if _, err := doc.AddImageBytes("body.png", bodyPNG(t), "image/png"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if _, err := doc.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	if logoAfter := extractPartRaw(t, out.Bytes(), "word/media/image1.png"); !bytes.Equal(logoBefore, logoAfter) {
		t.Error("body image overwrote the cloned header logo")
	}
}

func extractPart(t *testing.T, data []byte, name string) string {
	return string(extractPartRaw(t, data, name))
}

func extractPartRaw(t *testing.T, data []byte, name string) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			defer rc.Close()
			var b bytes.Buffer
			if _, err := b.ReadFrom(rc); err != nil {
				t.Fatal(err)
			}
			return b.Bytes()
		}
	}
	t.Fatalf("part %s missing", name)
	return nil
}
