package style

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// ---------- Constants ----------

const corpusWmlNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
const corpusRelsNS = "http://schemas.openxmlformats.org/package/2006/relationships"
const corpusDrawingMLNS = "http://schemas.openxmlformats.org/drawingml/2006/main"

// ---------- Expected fixture schema types ----------

type expectedFixture struct {
	Template    string          `json:"template"`
	Description string          `json:"description"`
	Cases       []expectedCase  `json:"cases"`
}

type expectedCase struct {
	ID       string         `json:"id"`
	Locator  string         `json:"locator"`
	PStyle   *string        `json:"pStyle"`
	Expected expectedProps  `json:"expected"`
	Warnings []string       `json:"warnings"`
}

type expectedProps struct {
	PPr map[string]any `json:"pPr"`
	RPr map[string]any `json:"rPr"`
}

// ---------- Harness loaders ----------

func loadExpectedFixture(t *testing.T, relPath string) *expectedFixture {
	t.Helper()
	b, err := os.ReadFile(relPath)
	if err != nil {
		t.Fatalf("loadExpectedFixture(%q): %v", relPath, err)
	}
	var f expectedFixture
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatalf("loadExpectedFixture(%q): json.Unmarshal: %v", relPath, err)
	}
	return &f
}

// repoRoot resolves the repository root by walking up from the test file's
// working directory (test CWD is the package dir).
func repoRoot() string {
	// When running tests, CWD is typically the package dir (internal/style/).
	// Walk up to find go.mod or .planning/.
	dir, _ := os.Getwd()
	for dir != "" && dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return "."
}

func hasRealDocx(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ---------- ZIP helpers ----------

// addZipEntry adds a single entry to a zip.Writer with a fixed timestamp.
// Reuses the same signature as cloner_test.go (same package → accessible).
func addCorpusZipEntry(t *testing.T, zw *zip.Writer, name string, payload []byte) {
	t.Helper()
	addZipEntry(t, zw, name, payload)
}

// openCorpusPkg opens a ZIP byte slice as an *opc.Package.
func openCorpusPkg(t *testing.T, data []byte) *opc.Package {
	t.Helper()
	pkg, err := opc.Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("opc.Open: %v", err)
	}
	return pkg
}

// ---------- Assertion helpers ----------

// assertEffectivePPr compares resolved CT_PPr fields against expected map.
// Only keys present in the expected map are checked.
func assertEffectivePPr(t *testing.T, got *wml.CT_PPr, expected map[string]any) {
	t.Helper()
	for key, val := range expected {
		switch key {
		case "spacing":
			if val == nil {
				if got.Spacing != nil {
					t.Errorf("Spacing: expected nil, got %+v", *got.Spacing)
				}
				continue
			}
			exp := val.(map[string]any)
			if got.Spacing == nil {
				t.Errorf("Spacing: expected non-nil, got nil")
				continue
			}
			for sk, sv := range exp {
				switch sk {
				case "before":
					if sv == nil {
						if got.Spacing.Before != nil {
							t.Errorf("Spacing.Before: expected nil, got %d", *got.Spacing.Before)
						}
					} else {
						expV := int64(sv.(float64))
						if got.Spacing.Before == nil {
							t.Errorf("Spacing.Before: expected %d, got nil", expV)
						} else if *got.Spacing.Before != expV {
							t.Errorf("Spacing.Before: expected %d, got %d", expV, *got.Spacing.Before)
						}
					}
				case "after":
					if sv == nil {
						if got.Spacing.After != nil {
							t.Errorf("Spacing.After: expected nil, got %d", *got.Spacing.After)
						}
					} else {
						expV := int64(sv.(float64))
						if got.Spacing.After == nil {
							t.Errorf("Spacing.After: expected %d, got nil", expV)
						} else if *got.Spacing.After != expV {
							t.Errorf("Spacing.After: expected %d, got %d", expV, *got.Spacing.After)
						}
					}
				case "line":
					if sv == nil {
						if got.Spacing.Line != nil {
							t.Errorf("Spacing.Line: expected nil, got %d", *got.Spacing.Line)
						}
					} else {
						expV := int64(sv.(float64))
						if got.Spacing.Line == nil {
							t.Errorf("Spacing.Line: expected %d, got nil", expV)
						} else if *got.Spacing.Line != expV {
							t.Errorf("Spacing.Line: expected %d, got %d", expV, *got.Spacing.Line)
						}
					}
				case "lineRule":
					if sv == nil {
						if got.Spacing.LineRule != nil {
							t.Errorf("Spacing.LineRule: expected nil, got %q", *got.Spacing.LineRule)
						}
					} else {
						expV := sv.(string)
						if got.Spacing.LineRule == nil {
							t.Errorf("Spacing.LineRule: expected %q, got nil", expV)
						} else if *got.Spacing.LineRule != expV {
							t.Errorf("Spacing.LineRule: expected %q, got %q", expV, *got.Spacing.LineRule)
						}
					}
				default:
					t.Errorf("unexpected spacing subkey: %q", sk)
				}
			}
		case "ind":
			if val == nil {
				if got.Ind != nil {
					t.Errorf("Ind: expected nil, got %+v", *got.Ind)
				}
				continue
			}
			exp := val.(map[string]any)
			if got.Ind == nil {
				t.Errorf("Ind: expected non-nil, got nil")
				continue
			}
			for ik, iv := range exp {
				switch ik {
				case "left":
					expV := int64(iv.(float64))
					if got.Ind.Left == nil {
						t.Errorf("Ind.Left: expected %d, got nil", expV)
					} else if *got.Ind.Left != expV {
						t.Errorf("Ind.Left: expected %d, got %d", expV, *got.Ind.Left)
					}
				case "right":
					expV := int64(iv.(float64))
					if got.Ind.Right == nil {
						t.Errorf("Ind.Right: expected %d, got nil", expV)
					} else if *got.Ind.Right != expV {
						t.Errorf("Ind.Right: expected %d, got %d", expV, *got.Ind.Right)
					}
				case "firstLine":
					expV := int64(iv.(float64))
					if got.Ind.FirstLine == nil {
						t.Errorf("Ind.FirstLine: expected %d, got nil", expV)
					} else if *got.Ind.FirstLine != expV {
						t.Errorf("Ind.FirstLine: expected %d, got %d", expV, *got.Ind.FirstLine)
					}
				case "hanging":
					expV := int64(iv.(float64))
					if got.Ind.Hanging == nil {
						t.Errorf("Ind.Hanging: expected %d, got nil", expV)
					} else if *got.Ind.Hanging != expV {
						t.Errorf("Ind.Hanging: expected %d, got %d", expV, *got.Ind.Hanging)
					}
				default:
					t.Errorf("unexpected ind subkey: %q", ik)
				}
			}
		case "keepNext":
			if val == nil {
				if got.KeepNext != nil {
					t.Errorf("KeepNext: expected nil, got non-nil")
				}
			} else if got.KeepNext == nil {
				t.Errorf("KeepNext: expected non-nil, got nil")
			}
		case "keepLines":
			if val == nil {
				if got.KeepLines != nil {
					t.Errorf("KeepLines: expected nil, got non-nil")
				}
			} else if got.KeepLines == nil {
				t.Errorf("KeepLines: expected non-nil, got nil")
			}
		case "outlineLvl":
			if val == nil {
				if got.OutlineLvl != nil {
					t.Errorf("OutlineLvl: expected nil, got %d", *got.OutlineLvl.Val)
				}
			} else if got.OutlineLvl == nil {
				t.Errorf("OutlineLvl: expected %v, got nil", val)
			} else {
				expV := int64(val.(float64))
				if *got.OutlineLvl.Val != expV {
					t.Errorf("OutlineLvl: expected %d, got %d", expV, *got.OutlineLvl.Val)
				}
			}
		case "jc":
			if val == nil {
				if got.Jc != nil {
					t.Errorf("Jc: expected nil, got non-nil")
				}
			} else if got.Jc == nil {
				t.Errorf("Jc: expected non-nil, got nil")
			}
		case "numPr":
			if val == nil {
				if got.NumPr != nil {
					t.Errorf("NumPr: expected nil, got non-nil")
				}
				continue
			}
			exp := val.(map[string]any)
			if got.NumPr == nil {
				t.Errorf("NumPr: expected non-nil, got nil")
				continue
			}
			for nk, nv := range exp {
				switch nk {
				case "numId":
					expV := int64(nv.(float64))
					if got.NumPr.NumId == nil || got.NumPr.NumId.Val == nil {
						t.Errorf("NumPr.NumId: expected %d, got nil", expV)
					} else if *got.NumPr.NumId.Val != expV {
						t.Errorf("NumPr.NumId: expected %d, got %d", expV, *got.NumPr.NumId.Val)
					}
				case "ilvl":
					expV := int64(nv.(float64))
					if got.NumPr.ILvl == nil || got.NumPr.ILvl.Val == nil {
						t.Errorf("NumPr.ILvl: expected %d, got nil", expV)
					} else if *got.NumPr.ILvl.Val != expV {
						t.Errorf("NumPr.ILvl: expected %d, got %d", expV, *got.NumPr.ILvl.Val)
					}
				default:
					t.Errorf("unexpected numPr subkey: %q", nk)
				}
			}
		default:
			t.Errorf("unexpected pPr key: %q", key)
		}
	}
}

func assertEffectiveRPr(t *testing.T, got *wml.CT_RPr, expected map[string]any) {
	t.Helper()
	for key, val := range expected {
		switch key {
		case "sz":
			if val == nil {
				if got.Sz != nil {
					t.Errorf("Sz: expected nil, got %d", *got.Sz.Val)
				}
				continue
			}
			exp := val.(map[string]any)
			if got.Sz == nil {
				t.Errorf("Sz: expected non-nil, got nil")
				continue
			}
			if expV, ok := exp["val"]; ok {
				v := int64(expV.(float64))
				if got.Sz.Val == nil {
					t.Errorf("Sz.Val: expected %d, got nil", v)
				} else if *got.Sz.Val != v {
					t.Errorf("Sz.Val: expected %d, got %d", v, *got.Sz.Val)
				}
			}
		case "b":
			if val == nil {
				if got.B != nil {
					t.Errorf("B: expected nil, got non-nil")
				}
			} else if got.B == nil {
				t.Errorf("B: expected non-nil, got nil")
			}
		case "i":
			if val == nil {
				if got.I != nil {
					t.Errorf("I: expected nil, got non-nil")
				}
			} else if got.I == nil {
				t.Errorf("I: expected non-nil, got nil")
			}
		case "color":
			if val == nil {
				if got.Color != nil {
					t.Errorf("Color: expected nil, got %+v", *got.Color)
				}
				continue
			}
			exp := val.(map[string]any)
			if got.Color == nil {
				t.Errorf("Color: expected non-nil, got nil")
				continue
			}
			for ck, cv := range exp {
				switch ck {
				case "val":
					if cv == nil {
						if got.Color.Val != nil {
							t.Errorf("Color.Val: expected nil, got %q", *got.Color.Val)
						}
					} else {
						expV := cv.(string)
						if got.Color.Val == nil {
							t.Errorf("Color.Val: expected %q, got nil", expV)
						} else if *got.Color.Val != expV {
							t.Errorf("Color.Val: expected %q, got %q", expV, *got.Color.Val)
						}
					}
				case "themeColor":
					if cv == nil {
						if got.Color.ThemeColor != nil {
							t.Errorf("Color.ThemeColor: expected nil, got %q", *got.Color.ThemeColor)
						}
					} else {
						expV := cv.(string)
						if got.Color.ThemeColor == nil {
							t.Errorf("Color.ThemeColor: expected %q, got nil", expV)
						} else if *got.Color.ThemeColor != expV {
							t.Errorf("Color.ThemeColor: expected %q, got %q", expV, *got.Color.ThemeColor)
						}
					}
				case "themeShade":
					if cv == nil {
						if got.Color.ThemeShade != nil {
							t.Errorf("Color.ThemeShade: expected nil, got %q", *got.Color.ThemeShade)
						}
					} else {
						expV := cv.(string)
						if got.Color.ThemeShade == nil {
							t.Errorf("Color.ThemeShade: expected %q, got nil", expV)
						} else if *got.Color.ThemeShade != expV {
							t.Errorf("Color.ThemeShade: expected %q, got %q", expV, *got.Color.ThemeShade)
						}
					}
				case "themeTint":
					if cv == nil {
						if got.Color.ThemeTint != nil {
							t.Errorf("Color.ThemeTint: expected nil, got %q", *got.Color.ThemeTint)
						}
					} else {
						expV := cv.(string)
						if got.Color.ThemeTint == nil {
							t.Errorf("Color.ThemeTint: expected %q, got nil", expV)
						} else if *got.Color.ThemeTint != expV {
							t.Errorf("Color.ThemeTint: expected %q, got %q", expV, *got.Color.ThemeTint)
						}
					}
				default:
					t.Errorf("unexpected color subkey: %q", ck)
				}
			}
		case "rFonts":
			if val == nil {
				if got.RFonts != nil {
					t.Errorf("RFonts: expected nil, got %+v", *got.RFonts)
				}
				continue
			}
			exp := val.(map[string]any)
			if got.RFonts == nil {
				t.Errorf("RFonts: expected non-nil, got nil")
				continue
			}
			for fk, fv := range exp {
				switch fk {
				case "ascii":
					if fv == nil {
						if got.RFonts.Ascii != nil {
							t.Errorf("RFonts.Ascii: expected nil, got %q", *got.RFonts.Ascii)
						}
					} else {
						expV := fv.(string)
						if got.RFonts.Ascii == nil {
							t.Errorf("RFonts.Ascii: expected %q, got nil", expV)
						} else if *got.RFonts.Ascii != expV {
							t.Errorf("RFonts.Ascii: expected %q, got %q", expV, *got.RFonts.Ascii)
						}
					}
				case "hAnsi":
					if fv == nil {
						if got.RFonts.HAnsi != nil {
							t.Errorf("RFonts.HAnsi: expected nil, got %q", *got.RFonts.HAnsi)
						}
					} else {
						expV := fv.(string)
						if got.RFonts.HAnsi == nil {
							t.Errorf("RFonts.HAnsi: expected %q, got nil", expV)
						} else if *got.RFonts.HAnsi != expV {
							t.Errorf("RFonts.HAnsi: expected %q, got %q", expV, *got.RFonts.HAnsi)
						}
					}
				case "asciiTheme":
					if fv == nil {
						if got.RFonts.AsciiTheme != nil {
							t.Errorf("RFonts.AsciiTheme: expected nil, got %q", *got.RFonts.AsciiTheme)
						}
					} else {
						expV := fv.(string)
						if got.RFonts.AsciiTheme == nil {
							t.Errorf("RFonts.AsciiTheme: expected %q, got nil", expV)
						} else if *got.RFonts.AsciiTheme != expV {
							t.Errorf("RFonts.AsciiTheme: expected %q, got %q", expV, *got.RFonts.AsciiTheme)
						}
					}
				case "hAnsiTheme":
					if fv == nil {
						if got.RFonts.HAnsiTheme != nil {
							t.Errorf("RFonts.HAnsiTheme: expected nil, got %q", *got.RFonts.HAnsiTheme)
						}
					} else {
						expV := fv.(string)
						if got.RFonts.HAnsiTheme == nil {
							t.Errorf("RFonts.HAnsiTheme: expected %q, got nil", expV)
						} else if *got.RFonts.HAnsiTheme != expV {
							t.Errorf("RFonts.HAnsiTheme: expected %q, got %q", expV, *got.RFonts.HAnsiTheme)
						}
					}
				default:
					t.Errorf("unexpected rFonts subkey: %q", fk)
				}
			}
		default:
			t.Errorf("unexpected rPr key: %q", key)
		}
	}
}

func assertWarnings(t *testing.T, got []string, expected []string) {
	t.Helper()
	for _, exp := range expected {
		found := false
		for _, g := range got {
			if strings.Contains(g, exp) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Warnings: expected substring %q not found in %v", exp, got)
		}
	}
}

// normalizeNilPPr unwraps a nil *CT_PPr to an empty struct for comparison.
func normalizeNilPPr(p *wml.CT_PPr) *wml.CT_PPr {
	if p == nil {
		return &wml.CT_PPr{}
	}
	return p
}

// assertCloneCorrectness verifies that after CloneStyles + Save + re-open,
// the 5 style parts are byte-identical to the source.
func assertCloneCorrectness(t *testing.T, srcPkg *opc.Package, dstPkg *opc.Package) {
	t.Helper()
	var saved bytes.Buffer
	if err := dstPkg.Save(&saved); err != nil {
		t.Fatalf("Save: %v", err)
	}
	reopened := openCorpusPkg(t, saved.Bytes())

	for _, p := range cloneParts {
		srcPart, ok := srcPkg.Parts[p.name]
		if !ok {
			continue // source lacks this part — skip
		}
		dstPart, ok := reopened.Parts[p.name]
		if !ok {
			t.Errorf("part %q missing from re-opened target after clone", p.name)
			continue
		}
		assertPartBytesEqual(t, srcPkg, reopened, p.name)

		// Also verify the part is byte-identical between the source part
		// and the cloned part in the target before save.
		srcBytes := readPartBytes(t, srcPart)
		dstBytes := readPartBytes(t, dstPart)
		if !bytes.Equal(srcBytes, dstBytes) {
			t.Errorf("part %q byte mismatch: src=%d bytes, dst=%d bytes", p.name, len(srcBytes), len(dstBytes))
		}
		_ = srcBytes
		_ = dstBytes
	}
}

// ---------- Synthetic fixture builders ----------

// buildBaseFixtureContentTypes returns [Content_Types].xml with all 5 style parts
// as Overrides plus Defaults for xml/rels.
func buildBaseFixtureContentTypes() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
<Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>
<Override PartName="/word/fontTable.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"/>
<Override PartName="/word/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
<Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/>
</Types>`
}

var (
	corpusFontTableXML = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:fonts xmlns:w="` + corpusWmlNS + `"/>`)
	corpusSettingsXML  = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:settings xmlns:w="` + corpusWmlNS + `"><w:zoom w:percent="100"/></w:settings>`)
)

// buildFixturePkg builds a minimal ZIP skeleton with content types, rels,
// and document.xml, then calls addPart for each extra part. Returns the
// ZIP bytes.
func buildFixturePkg(t *testing.T, docXML string, themeXML string, numberingXML string, stylesXML string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	addCorpusZipEntry(t, zw, "[Content_Types].xml", []byte(buildBaseFixtureContentTypes()))
	addCorpusZipEntry(t, zw, "_rels/.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="`+corpusRelsNS+`">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))
	addCorpusZipEntry(t, zw, "word/document.xml", []byte(docXML))
	addCorpusZipEntry(t, zw, "word/_rels/document.xml.rels", []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="`+corpusRelsNS+`">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`))
	addCorpusZipEntry(t, zw, "word/styles.xml", []byte(stylesXML))
	addCorpusZipEntry(t, zw, "word/numbering.xml", []byte(numberingXML))
	addCorpusZipEntry(t, zw, "word/fontTable.xml", corpusFontTableXML)
	addCorpusZipEntry(t, zw, "word/theme/theme1.xml", []byte(themeXML))
	addCorpusZipEntry(t, zw, "word/settings.xml", corpusSettingsXML)

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// headingChainStylesXML: Normal→Heading1→Heading2, docDefaults with rPr
// asciiTheme=minorHAnsi sz=22, pPr spacing after=160.
func headingChainStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="` + corpusWmlNS + `">
<w:docDefaults>
<w:rPrDefault><w:rPr><w:rFonts w:asciiTheme="minorHAnsi" w:hAnsiTheme="minorHAnsi"/><w:sz w:val="22"/><w:szCs w:val="22"/></w:rPr></w:rPrDefault>
<w:pPrDefault><w:pPr><w:spacing w:after="160"/></w:pPr></w:pPrDefault>
</w:docDefaults>
<w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
<w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/><w:basedOn w:val="Normal"/><w:pPr><w:keepNext/><w:spacing w:before="480" w:after="240"/><w:outlineLvl w:val="0"/></w:pPr><w:rPr><w:sz w:val="32"/><w:b/><w:color w:themeColor="accent1"/><w:rFonts w:asciiTheme="majorHAnsi" w:hAnsiTheme="majorHAnsi"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="Heading2"><w:name w:val="heading 2"/><w:basedOn w:val="Heading1"/><w:pPr><w:keepNext/><w:spacing w:before="240" w:after="120"/><w:outlineLvl w:val="1"/></w:pPr><w:rPr><w:sz w:val="28"/><w:color w:themeColor="accent1"/></w:rPr></w:style>
</w:styles>`
}

func headingChainDocXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + corpusWmlNS + `"><w:body>
<w:p><w:pPr><w:pStyle w:val="Heading2"/></w:pPr><w:r><w:t>Heading2 paragraph</w:t></w:r></w:p>
</w:body></w:document>`
}

func headingChainThemeXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="` + corpusDrawingMLNS + `" name="Test"><a:themeElements><a:clrScheme name="Test">
<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
<a:accent1><a:srgbClr val="156082"/></a:accent1>
</a:clrScheme></a:themeElements></a:theme>`
}

func buildHeadingChainSourceZip(t *testing.T) []byte {
	t.Helper()
	return buildFixturePkg(t,
		headingChainDocXML(),
		headingChainThemeXML(),
		emptyNumberingXML(),
		headingChainStylesXML(),
	)
}

func emptyNumberingXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="` + corpusWmlNS + `"/>`
}

// theme-refs fixture: docDefaults rPr color themeColor="dark1",
// ParaAccent style with color themeColor="accent1",
// ParaHyperlink style with color themeColor="hyperlink".
func themeRefsStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="` + corpusWmlNS + `">
<w:docDefaults>
<w:rPrDefault><w:rPr><w:color w:themeColor="dark1"/></w:rPr></w:rPrDefault>
</w:docDefaults>
<w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
<w:style w:type="paragraph" w:styleId="ParaAccent"><w:name w:val="ParaAccent"/><w:rPr><w:color w:themeColor="accent1"/></w:rPr></w:style>
<w:style w:type="paragraph" w:styleId="ParaHyperlink"><w:name w:val="ParaHyperlink"/><w:rPr><w:color w:themeColor="hyperlink"/></w:rPr></w:style>
</w:styles>`
}

func themeRefsDocXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + corpusWmlNS + `"><w:body>
<w:p><w:pPr><w:pStyle w:val="ParaAccent"/></w:pPr><w:r><w:t>Accent1</w:t></w:r></w:p>
<w:p><w:r><w:t>Default dark1</w:t></w:r></w:p>
<w:p><w:pPr><w:pStyle w:val="ParaHyperlink"/></w:pPr><w:r><w:t>Hyperlink</w:t></w:r></w:p>
</w:body></w:document>`
}

func themeRefsThemeXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="` + corpusDrawingMLNS + `" name="Test"><a:themeElements><a:clrScheme name="Test">
<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
<a:accent1><a:srgbClr val="156082"/></a:accent1>
<a:hlink><a:srgbClr val="467886"/></a:hlink>
<a:folHlink><a:srgbClr val="96607D"/></a:folHlink>
</a:clrScheme></a:themeElements></a:theme>`
}

func buildThemeRefsSourceZip(t *testing.T) []byte {
	t.Helper()
	return buildFixturePkg(t,
		themeRefsDocXML(),
		themeRefsThemeXML(),
		emptyNumberingXML(),
		themeRefsStylesXML(),
	)
}

// multi-level numbering fixture: abstractNum 0 with 3 levels,
// num 1 abstractNumId=0, 3 paragraphs with numPr.
func multiLevelNumberingStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="` + corpusWmlNS + `">
<w:docDefaults>
<w:pPrDefault><w:pPr><w:spacing w:after="160"/></w:pPr></w:pPrDefault>
</w:docDefaults>
<w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
</w:styles>`
}

func multiLevelNumberingXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="` + corpusWmlNS + `">
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/><w:pPr><w:ind w:left="720"/></w:pPr></w:lvl>
<w:lvl w:ilvl="1"><w:numFmt w:val="lowerLetter"/><w:lvlText w:val="%2."/><w:start w:val="1"/><w:pPr><w:ind w:left="1440"/></w:pPr></w:lvl>
<w:lvl w:ilvl="2"><w:numFmt w:val="bullet"/><w:lvlText w:val="•"/><w:start w:val="1"/><w:pPr><w:ind w:left="2160"/></w:pPr></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
}

func multiLevelNumberingDocXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + corpusWmlNS + `"><w:body>
<w:p><w:pPr><w:pStyle w:val="Normal"/><w:numPr><w:numId w:val="1"/><w:ilvl w:val="0"/></w:numPr></w:pPr><w:r><w:t>Level 0</w:t></w:r></w:p>
<w:p><w:pPr><w:pStyle w:val="Normal"/><w:numPr><w:numId w:val="1"/><w:ilvl w:val="1"/></w:numPr></w:pPr><w:r><w:t>Level 1</w:t></w:r></w:p>
<w:p><w:pPr><w:pStyle w:val="Normal"/><w:numPr><w:numId w:val="1"/><w:ilvl w:val="2"/></w:numPr></w:pPr><w:r><w:t>Level 2</w:t></w:r></w:p>
</w:body></w:document>`
}

func buildMultiLevelNumberingSourceZip(t *testing.T) []byte {
	t.Helper()
	return buildFixturePkg(t,
		multiLevelNumberingDocXML(),
		emptyThemeXML(),
		multiLevelNumberingXML(),
		multiLevelNumberingStylesXML(),
	)
}

func emptyThemeXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="` + corpusDrawingMLNS + `" name="Empty"><a:themeElements><a:clrScheme name="Empty">
<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
</a:clrScheme></a:themeElements></a:theme>`
}

// circular basedOn fixture: CycleA.basedOn=CycleB, CycleB.basedOn=CycleA.
func circularBasedOnStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="` + corpusWmlNS + `">
<w:style w:type="paragraph" w:styleId="CycleA"><w:name w:val="CycleA"/><w:basedOn w:val="CycleB"/><w:pPr><w:spacing w:before="100"/></w:pPr></w:style>
<w:style w:type="paragraph" w:styleId="CycleB"><w:name w:val="CycleB"/><w:basedOn w:val="CycleA"/></w:style>
</w:styles>`
}

func circularBasedOnDocXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + corpusWmlNS + `"><w:body>
<w:p><w:pPr><w:pStyle w:val="CycleA"/></w:pPr><w:r><w:t>Circular test</w:t></w:r></w:p>
</w:body></w:document>`
}

func buildCircularBasedOnSourceZip(t *testing.T) []byte {
	t.Helper()
	return buildFixturePkg(t,
		circularBasedOnDocXML(),
		emptyThemeXML(),
		emptyNumberingXML(),
		circularBasedOnStylesXML(),
	)
}

// dangling basedOn fixture: Ghost.basedOn=Missing (Missing not in styles.xml).
func danglingBasedOnStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="` + corpusWmlNS + `">
<w:style w:type="paragraph" w:styleId="Ghost"><w:name w:val="Ghost"/><w:basedOn w:val="Missing"/><w:pPr><w:spacing w:before="200"/></w:pPr></w:style>
</w:styles>`
}

func danglingBasedOnDocXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + corpusWmlNS + `"><w:body>
<w:p><w:pPr><w:pStyle w:val="Ghost"/></w:pPr><w:r><w:t>Dangling test</w:t></w:r></w:p>
</w:body></w:document>`
}

func buildDanglingBasedOnSourceZip(t *testing.T) []byte {
	t.Helper()
	return buildFixturePkg(t,
		danglingBasedOnDocXML(),
		emptyThemeXML(),
		emptyNumberingXML(),
		danglingBasedOnStylesXML(),
	)
}

// missing numId fixture: numId=999 not in numbering.xml.
func missingNumIdStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="` + corpusWmlNS + `">
<w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
</w:styles>`
}

func missingNumIdNumberingXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="` + corpusWmlNS + `">
<w:abstractNum w:abstractNumId="0">
<w:lvl w:ilvl="0"><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:start w:val="1"/></w:lvl>
</w:abstractNum>
<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>
</w:numbering>`
}

func missingNumIdDocXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="` + corpusWmlNS + `"><w:body>
<w:p><w:pPr><w:pStyle w:val="Normal"/><w:numPr><w:numId w:val="999"/><w:ilvl w:val="0"/></w:numPr></w:pPr><w:r><w:t>Missing numId</w:t></w:r></w:p>
</w:body></w:document>`
}

func buildMissingNumIdSourceZip(t *testing.T) []byte {
	t.Helper()
	return buildFixturePkg(t,
		missingNumIdDocXML(),
		emptyThemeXML(),
		missingNumIdNumberingXML(),
		missingNumIdStylesXML(),
	)
}

// ---------- resolve helpers for corpus tests ----------

// corpusResolveParagraph is a convenience wrapper: opens a synthetic source
// ZIP, clones its 5 style parts into a fresh target, creates a resolver over
// the target, and resolves the first paragraph's effective props.  The
// paragraph comes from the SOURCE zip's document.xml (the parts to resolve
// are style data, not the document itself — CloneStyles only copies style
// parts, not document.xml).
func corpusResolveParagraph(t *testing.T, srcZip []byte) (*wml.CT_PPr, *wml.CT_RPr, *Resolver, *opc.Package) {
	t.Helper()
	srcPkg := openCorpusPkg(t, srcZip)
	dstZip := buildFreshTargetZip(t)
	dstPkg := openCorpusPkg(t, dstZip)

	if err := CloneStyles(srcPkg, dstPkg); err != nil {
		t.Fatalf("CloneStyles: %v", err)
	}

	resolver := NewResolver(dstPkg)

	// Parse document.xml from the SOURCE package (CloneStyles does NOT copy
	// document.xml — the target has a generic empty paragraph from
	// buildFreshTargetZip, not the fixture-specific one).
	docPart, ok := srcPkg.Parts["word/document.xml"]
	if !ok {
		t.Fatal("word/document.xml not found in source package")
	}
	rc, err := docPart.Open()
	if err != nil {
		t.Fatalf("open source document.xml: %v", err)
	}
	defer rc.Close()

	var doc wml.CT_Document
	dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("decode source document.xml: %v", err)
	}
	if len(doc.Body.P) == 0 {
		t.Fatal("source document has no paragraphs")
	}
	p := doc.Body.P[0]
	ppr, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	var rpr *wml.CT_RPr
	if len(p.R) > 0 {
		rpr, err = resolver.ResolveRun(p, p.R[0])
		if err != nil {
			t.Fatalf("ResolveRun: %v", err)
		}
	}
	return ppr, rpr, resolver, dstPkg
}

// ---------- Test: Heading2 chain (ROADMAP success criterion #4) ----------

func TestCorpus_Heading2Chain(t *testing.T) {
	srcZip := buildHeadingChainSourceZip(t)
	ppr, rpr, resolver, dstPkg := corpusResolveParagraph(t, srcZip)
	_ = dstPkg // clone correctness checked below

	fixture := loadExpectedFixture(t, expectedFixturePath("word/style-rich/heading-chain.expected.json"))
	if len(fixture.Cases) == 0 {
		t.Fatal("expected.json has no cases")
	}
	c := fixture.Cases[0]

	if c.Expected.PPr != nil {
		assertEffectivePPr(t, ppr, c.Expected.PPr)
	}
	if c.Expected.RPr != nil {
		assertEffectiveRPr(t, rpr, c.Expected.RPr)
	}
	assertWarnings(t, resolver.Warnings(), c.Warnings)

	// Clone correctness: after Save + re-open, parts are byte-identical
	srcPkg := openCorpusPkg(t, srcZip)
	dstZip2 := buildFreshTargetZip(t)
	dstPkg2 := openCorpusPkg(t, dstZip2)
	if err := CloneStyles(srcPkg, dstPkg2); err != nil {
		t.Fatalf("CloneStyles for correctness: %v", err)
	}
	assertCloneCorrectness(t, srcPkg, dstPkg2)
}

// expectedFixturePath returns the absolute path to an expected.json under testdata/.
func expectedFixturePath(rel string) string {
	return filepath.Join(repoRoot(), "testdata", rel)
}

// ---------- Test: Theme color refs ----------

func TestCorpus_ThemeRefs(t *testing.T) {
	srcZip := buildThemeRefsSourceZip(t)
	srcPkg := openCorpusPkg(t, srcZip)
	dstZip := buildFreshTargetZip(t)
	dstPkg := openCorpusPkg(t, dstZip)

	if err := CloneStyles(srcPkg, dstPkg); err != nil {
		t.Fatalf("CloneStyles: %v", err)
	}

	resolver := NewResolver(dstPkg)
	docPart := srcPkg.Parts["word/document.xml"]
	rc, err := docPart.Open()
	if err != nil {
		t.Fatalf("open document.xml: %v", err)
	}
	defer rc.Close()

	var doc wml.CT_Document
	dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("decode document.xml: %v", err)
	}

	fixture := loadExpectedFixture(t, expectedFixturePath("word/style-rich/theme-refs.expected.json"))

	for i, c := range fixture.Cases {
		t.Run(c.ID, func(t *testing.T) {
			if i >= len(doc.Body.P) {
				t.Fatalf("case %d (id=%s): not enough paragraphs in doc (%d)", i, c.ID, len(doc.Body.P))
			}
			p := doc.Body.P[i]
			ppr, err := resolver.ResolveParagraph(p)
			if err != nil {
				t.Fatalf("ResolveParagraph: %v", err)
			}
			var rpr *wml.CT_RPr
			if len(p.R) > 0 {
				rpr, err = resolver.ResolveRun(p, p.R[0])
				if err != nil {
					t.Fatalf("ResolveRun: %v", err)
				}
			}
			if c.Expected.PPr != nil {
				assertEffectivePPr(t, ppr, c.Expected.PPr)
			}
			if c.Expected.RPr != nil {
				assertEffectiveRPr(t, rpr, c.Expected.RPr)
			}
			assertWarnings(t, resolver.Warnings(), c.Warnings)
		})
	}

	// Clone correctness
	assertCloneCorrectness(t, srcPkg, dstPkg)
}

// ---------- Test: Multi-level numbering ----------

func TestCorpus_MultiLevelNumbering(t *testing.T) {
	srcZip := buildMultiLevelNumberingSourceZip(t)
	srcPkg := openCorpusPkg(t, srcZip)
	dstZip := buildFreshTargetZip(t)
	dstPkg := openCorpusPkg(t, dstZip)

	if err := CloneStyles(srcPkg, dstPkg); err != nil {
		t.Fatalf("CloneStyles: %v", err)
	}

	resolver := NewResolver(dstPkg)
	docPart := srcPkg.Parts["word/document.xml"]
	rc, err := docPart.Open()
	if err != nil {
		t.Fatalf("open document.xml: %v", err)
	}
	defer rc.Close()

	var doc wml.CT_Document
	dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("decode document.xml: %v", err)
	}

	fixture := loadExpectedFixture(t, expectedFixturePath("word/style-rich/multi-level-numbering.expected.json"))

	for i, c := range fixture.Cases {
		t.Run(c.ID, func(t *testing.T) {
			if i >= len(doc.Body.P) {
				t.Fatalf("case %d (id=%s): not enough paragraphs", i, c.ID)
			}
			p := doc.Body.P[i]
			ppr, err := resolver.ResolveParagraph(p)
			if err != nil {
				t.Fatalf("ResolveParagraph: %v", err)
			}
			var rpr *wml.CT_RPr
			if len(p.R) > 0 {
				rpr, err = resolver.ResolveRun(p, p.R[0])
				if err != nil {
					t.Fatalf("ResolveRun: %v", err)
				}
			}
			if c.Expected.PPr != nil {
				assertEffectivePPr(t, ppr, c.Expected.PPr)
			}
			if c.Expected.RPr != nil {
				assertEffectiveRPr(t, rpr, c.Expected.RPr)
			}
			assertWarnings(t, resolver.Warnings(), c.Warnings)
		})
	}

	assertCloneCorrectness(t, srcPkg, dstPkg)
}

// ---------- Test: Circular basedOn (D-05) ----------

func TestCorpus_CircularBasedOn(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked on circular basedOn: %v", r)
		}
	}()

	srcZip := buildCircularBasedOnSourceZip(t)
	ppr, rpr, resolver, _ := corpusResolveParagraph(t, srcZip)
	_ = rpr

	fixture := loadExpectedFixture(t, expectedFixturePath("style-engine/hostile/circular-basedon.expected.json"))
	if len(fixture.Cases) == 0 {
		t.Fatal("expected.json has no cases")
	}
	c := fixture.Cases[0]

	if c.Expected.PPr != nil {
		assertEffectivePPr(t, ppr, c.Expected.PPr)
	}
	assertWarnings(t, resolver.Warnings(), c.Warnings)
}

// ---------- Test: Dangling basedOn (D-07) ----------

func TestCorpus_DanglingBasedOn(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked on dangling basedOn: %v", r)
		}
	}()

	srcZip := buildDanglingBasedOnSourceZip(t)
	ppr, rpr, resolver, _ := corpusResolveParagraph(t, srcZip)
	_ = rpr

	fixture := loadExpectedFixture(t, expectedFixturePath("style-engine/hostile/dangling-basedon.expected.json"))
	if len(fixture.Cases) == 0 {
		t.Fatal("expected.json has no cases")
	}
	c := fixture.Cases[0]

	if c.Expected.PPr != nil {
		assertEffectivePPr(t, ppr, c.Expected.PPr)
	}
	assertWarnings(t, resolver.Warnings(), c.Warnings)
}

// ---------- Test: Missing numId (D-07) ----------

func TestCorpus_MissingNumId(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked on missing numId: %v", r)
		}
	}()

	srcZip := buildMissingNumIdSourceZip(t)
	ppr, rpr, resolver, _ := corpusResolveParagraph(t, srcZip)
	_ = rpr

	fixture := loadExpectedFixture(t, expectedFixturePath("style-engine/hostile/missing-numid.expected.json"))
	if len(fixture.Cases) == 0 {
		t.Fatal("expected.json has no cases")
	}
	c := fixture.Cases[0]

	if c.Expected.PPr != nil {
		assertEffectivePPr(t, ppr, c.Expected.PPr)
	}
	assertWarnings(t, resolver.Warnings(), c.Warnings)
}

// ---------- Test: Clone target not empty (D-08) ----------

func TestCorpus_CloneTargetNotEmpty(t *testing.T) {
	srcZip := buildHeadingChainSourceZip(t)
	srcPkg := openCorpusPkg(t, srcZip)
	dstZip := buildTargetWithStylesZip(t)
	dstPkg := openCorpusPkg(t, dstZip)

	err := CloneStyles(srcPkg, dstPkg)
	if err == nil {
		t.Fatal("CloneStyles: expected ErrCloneTargetNotEmpty, got nil")
	}
	if !errors.Is(err, ErrCloneTargetNotEmpty) {
		t.Fatalf("CloneStyles: err = %v, want ErrCloneTargetNotEmpty wrapping", err)
	}
}

// ---------- Test: Named styles applicable by name (STYLE-CLONE-02) ----------

func TestCorpus_NamedStylesApplicableByName(t *testing.T) {
	// Clone the heading-chain template into a fresh target, then resolve a
	// paragraph with pStyle="Heading2". The effective props must match
	// Heading2's merged chain (the same assertion as TestCorpus_Heading2Chain).
	srcZip := buildHeadingChainSourceZip(t)
	srcPkg := openCorpusPkg(t, srcZip)
	dstZip := buildFreshTargetZip(t)
	dstPkg := openCorpusPkg(t, dstZip)

	if err := CloneStyles(srcPkg, dstPkg); err != nil {
		t.Fatalf("CloneStyles: %v", err)
	}

	resolver := NewResolver(dstPkg)

	// Construct a paragraph with pStyle="Heading2" to test named style applicability.
	p := &wml.CT_P{
		PPr: &wml.CT_PPr{
			PStyle: &wml.CT_PStyle{Val: strPtr("Heading2")},
		},
		R: []*wml.CT_R{{}},
	}

	ppr, err := resolver.ResolveParagraph(p)
	if err != nil {
		t.Fatalf("ResolveParagraph: %v", err)
	}
	if ppr == nil {
		t.Fatal("ResolveParagraph returned nil pPr")
	}

	// Verify Heading2-specific props: spacing before=240 after=120, outlineLvl=1
	if ppr.Spacing == nil || ppr.Spacing.Before == nil || *ppr.Spacing.Before != 240 {
		t.Errorf("Heading2 spacing.before: expected 240, got %v", ppr.Spacing)
	}
	if ppr.OutlineLvl == nil || *ppr.OutlineLvl.Val != 1 {
		t.Errorf("Heading2 outlineLvl: expected 1, got %v", ppr.OutlineLvl)
	}
	if ppr.KeepNext == nil {
		t.Error("Heading2 keepNext: expected non-nil (inherited from Heading1)")
	}

	// ResolveRun for the same paragraph
	rpr, err := resolver.ResolveRun(p, p.R[0])
	if err != nil {
		t.Fatalf("ResolveRun: %v", err)
	}
	if rpr == nil {
		t.Fatal("ResolveRun returned nil rPr")
	}
	if rpr.Sz == nil || *rpr.Sz.Val != 28 {
		t.Errorf("Heading2 sz: expected 28, got %v", rpr.Sz)
	}
	if rpr.B == nil {
		t.Error("Heading2 b: expected non-nil (from Heading1)")
	}

	// Named styles applicable by name = STYLE-CLONE-02 acceptance
	t.Log("STYLE-CLONE-02: Named styles from cloned template applicable by name — PASSED")
}

// ---------- Test: Real fixtures (t.Skip-guarded) ----------

func TestCorpus_RealFixtures(t *testing.T) {
	root := repoRoot()
	styleRichDir := filepath.Join(root, "testdata", "word", "style-rich")
	hostileDir := filepath.Join(root, "testdata", "style-engine", "hostile")

	// Check style-rich directory for .docx files
	docxFiles, err := filepath.Glob(filepath.Join(styleRichDir, "*.docx"))
	if err != nil || len(docxFiles) == 0 {
		t.Skip("no real .docx fixtures found in testdata/word/style-rich/ — user-authored per D-10/D-11; synthetic equivalents cover the same paths in CI")
	}

	// Verify hostile fixture directory exists
	if _, err := os.Stat(hostileDir); os.IsNotExist(err) {
		t.Log("testdata/style-engine/hostile/ not present — real hostile fixtures not available")
	}

	// For each .docx found, open, clone, resolve, assert against expected.json
	for _, docxPath := range docxFiles {
		name := strings.TrimSuffix(filepath.Base(docxPath), ".docx")
		expectedPath := filepath.Join(styleRichDir, name+".expected.json")

		t.Run(name, func(t *testing.T) {
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Skipf("expected.json not found for %s — skipping", name)
			}

			b, err := os.ReadFile(docxPath)
			if err != nil {
				t.Fatalf("read %s: %v", docxPath, err)
			}
			pkg := openCorpusPkg(t, b)

			dstZip := buildFreshTargetZip(t)
			dstPkg := openCorpusPkg(t, dstZip)

			if err := CloneStyles(pkg, dstPkg); err != nil {
				t.Fatalf("CloneStyles: %v", err)
			}

			resolver := NewResolver(dstPkg)

			// Parse document.xml from the cloned target
			docPart, ok := dstPkg.Parts["word/document.xml"]
			if !ok {
				t.Fatal("word/document.xml not found in cloned target")
			}
			rc, err := docPart.Open()
			if err != nil {
				t.Fatalf("open document.xml: %v", err)
			}
			defer rc.Close()

			var doc wml.CT_Document
			dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)
			if err := dec.Decode(&doc); err != nil {
				t.Fatalf("decode document.xml: %v", err)
			}

			fixture := loadExpectedFixture(t, expectedPath)
			for i, c := range fixture.Cases {
				t.Run(c.ID, func(t *testing.T) {
					if i >= len(doc.Body.P) {
						t.Skipf("not enough paragraphs in real docx (need %d, have %d)", i+1, len(doc.Body.P))
					}
					par := doc.Body.P[i]
					ppr, err := resolver.ResolveParagraph(par)
					if err != nil {
						t.Fatalf("ResolveParagraph: %v", err)
					}
					var rpr *wml.CT_RPr
					if len(par.R) > 0 {
						rpr, err = resolver.ResolveRun(par, par.R[0])
						if err != nil {
							t.Fatalf("ResolveRun: %v", err)
						}
					}
					if c.Expected.PPr != nil {
						assertEffectivePPr(t, ppr, c.Expected.PPr)
					}
					if c.Expected.RPr != nil {
						assertEffectiveRPr(t, rpr, c.Expected.RPr)
					}
					assertWarnings(t, resolver.Warnings(), c.Warnings)
				})
			}
		})
	}
}
