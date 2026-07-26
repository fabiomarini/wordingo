package wordingo

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeBluePNG() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.NRGBA{R: 50, G: 100, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic("encode blue PNG: " + err.Error())
	}
	return buf.Bytes()
}

func TestRichContentStory(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	doc.AddHeader(HeaderDefault).AddParagraph("Wordingo Storybook — Header")

	doc.AddFooter(FooterDefault).AddParagraph("Page ")

	doc.AddParagraph("The Little Programmer").SetStyle("Title")

	doc.AddParagraph("A Tale of Bugs and Triumph").SetStyle("Subtitle")

	doc.AddParagraph("Chapter 1: The Discovery").SetStyle("Heading1")

	doc.AddParagraph("Once upon a time, in a land of ones and zeros, there lived a little programmer named Ada. Ada loved crafting beautiful code, but one day she encountered a bug that made her program crash whenever the moon was full.")

	doc.AddParagraph("She sat at her desk, staring at the screen, determined to find the source of the glitch. The stack trace was long and winding, like a dragon's tail.")

	doc.AddParagraph("Chapter 2: The Chase").SetStyle("Heading2")

	doc.AddParagraph("Ada started her investigation. She sprinkled log statements throughout the code, hoping to catch the culprit in the act. She followed the trail from main() all the way down to a tiny helper function that had a null pointer just waiting to strike.")

	doc.AddParagraph("Chapter 3: The Fix").SetStyle("Heading3")

	doc.AddParagraph("With a steady hand and a clear mind, Ada added a simple nil check. She recompiled, ran the tests, and watched the green checkmark appear. The bug was vanquished!")

	doc.AddParagraph("The moral of the story: even the mightiest bugs fall to patience, perseverance, and a well-placed guard clause.")

	pngData := makeBluePNG()
	doc.AddParagraph("")
	run, err := doc.AddImageBytes("blue-square.png", pngData, "image/png")
	if err != nil {
		t.Fatalf("AddImageBytes: %v", err)
	}
	if run == nil {
		t.Fatal("AddImageBytes returned nil Run")
	}
	run.SetImageWidth(2.0).SetImageHeight(2.0)

	tbl, err := doc.AddTable([][]string{
		{"Character", "Role", "Lesson"},
		{"Ada", "Programmer", "Patience"},
		{"Bug", "Antagonist", "Nil checks matter"},
		{"Moon", "Catalyst", "Edge cases hiding"},
	})
	if err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	tbl.SetTableStyle("LightGridAccent1")
	tbl.SetBorders(&TableBorders{
		Top:    &BorderDef{Style: "single", Size: 4, Color: "4472C4"},
		Bottom: &BorderDef{Style: "single", Size: 4, Color: "4472C4"},
	})

	doc.AddParagraph("")

	doc.AddParagraph("Key Takeaways").SetStyle("Heading2")

	list := doc.AddList(false)
	list.AddItem("Always check for nil", 0)
	list.AddItem("Log generously", 0)
	list.AddItem("Trust the stack trace", 0)
	list.AddItem("Sleep on hard problems", 0)

	para := doc.AddParagraph("")
	para.SetStyle("Normal")
	para.AddRun("You can read more at ")

	para.AddHyperlink("Ada's Blog", "https://example.com/ada").SetColor("0563C1").SetUnderline("single")

	doc.AddParagraph("")

	doc.AddParagraph("Fin").SetStyle("Heading1")

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	writePath := filepath.Join("testdata", "roundtrip", "rich-content-story.docx")
	if err := os.MkdirAll(filepath.Dir(writePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(writePath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}

	parts := make(map[string]string)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		var content bytes.Buffer
		_, _ = content.ReadFrom(rc)
		rc.Close()
		parts[f.Name] = content.String()
	}

	tests := []struct {
		name    string
		part    string
		needle  string
		present bool
	}{
		{"Title style applied", "word/document.xml", `w:val="Title"`, true},
		{"Subtitle style applied", "word/document.xml", `w:val="Subtitle"`, true},
		{"Heading1 style applied", "word/document.xml", `w:val="Heading1"`, true},
		{"Heading2 style applied", "word/document.xml", `w:val="Heading2"`, true},
		{"Heading3 style applied", "word/document.xml", `w:val="Heading3"`, true},
		{"Story text present", "word/document.xml", "little programmer", true},
		{"Image media part", "word/media/image1.png", "", true},
		{"PNG content type", "[Content_Types].xml", "image/png", true},
		{"Image rel type", "word/_rels/document.xml.rels", relImage, true},
		{"DrawingML inline", "word/document.xml", "wp:inline", true},
		{"Table present", "word/document.xml", "Character", true},
		{"Table has grid", "word/document.xml", "tblGrid", true},
		{"List paragraph", "word/document.xml", "numId", true},
		{"Hyperlink relationship", "word/_rels/document.xml.rels", "External", true},
		{"Hyperlink in doc", "word/document.xml", "w:hyperlink", true},
		{"Header part exists", "word/header1.xml", "", true},
		{"Footer part exists", "word/footer1.xml", "", true},
		{"Header reference in sectPr", "word/document.xml", "w:headerReference", true},
		{"Footer reference in sectPr", "word/document.xml", "w:footerReference", true},
	}

	for _, tt := range tests {
		content, ok := parts[tt.part]
		if !ok {
			t.Errorf("missing part %s", tt.part)
			continue
		}
		if tt.present {
			if !strings.Contains(content, tt.needle) {
				t.Errorf("%s: %s missing %q", tt.part, tt.name, tt.needle)
			}
		} else {
			if strings.Contains(content, tt.needle) {
				t.Errorf("%s: %s has unexpected %q", tt.part, tt.name, tt.needle)
			}
		}
	}

	if err := doc.Close(); err != nil {
		t.Fatal(err)
	}

	doc2, err := Open("testdata/roundtrip/rich-content-story.docx")
	if err != nil {
		t.Fatalf("Open roundtrip: %v", err)
	}
	defer doc2.Close()

	paras := doc2.Paragraphs()
	if len(paras) < 2 {
		t.Fatal("expected at least 2 paragraphs (empty default + Title)")
	}
	if para := paras[1]; para.Style() != "Title" {
		t.Errorf("first styled paragraph: got %q, want Title", para.Style())
	}
}
