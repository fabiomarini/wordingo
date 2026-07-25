package wordingo

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

//go:embed defaults/styles.xml
var defaultStyles []byte

//go:embed defaults/theme1.xml
var defaultTheme []byte

//go:embed defaults/fontTable.xml
var defaultFontTable []byte

//go:embed defaults/settings.xml
var defaultSettings []byte

//go:embed defaults/webSettings.xml
var defaultWebSettings []byte

const (
	ctMain        = "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"
	ctStyles      = "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"
	ctSettings    = "application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"
	ctWebSettings = "application/vnd.openxmlformats-officedocument.wordprocessingml.webSettings+xml"
	ctFontTable   = "application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"
	ctTheme       = "application/vnd.openxmlformats-officedocument.theme+xml"
	ctRels        = "application/vnd.openxmlformats-package.relationships+xml"
	ctXML         = "application/xml"
)

const (
	relOfficeDocument = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	relStyles         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"
	relSettings       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings"
	relWebSettings    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/webSettings"
	relFontTable      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable"
	relTheme          = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme"
)

func newBlankPackage() *opc.Package {
	pkg := &opc.Package{
		Parts: make(map[string]*opc.Part),
		ContentTypes: &opc.ContentTypes{
			Defaults: map[string]string{
				"rels": ctRels,
				"xml":  ctXML,
			},
			Overrides: map[string]string{
				"/word/document.xml":     ctMain,
				"/word/styles.xml":       ctStyles,
				"/word/settings.xml":     ctSettings,
				"/word/webSettings.xml":  ctWebSettings,
				"/word/fontTable.xml":    ctFontTable,
				"/word/theme/theme1.xml": ctTheme,
			},
		},
		Rels:        make(map[string]*opc.Relationships),
		Conformance: opc.Transitional,
	}

	docXML := buildDocumentXML()

	addPart(pkg, "[Content_Types].xml", nil)
	addPart(pkg, "_rels/.rels", nil)
	addPart(pkg, "word/document.xml", docXML)
	addPart(pkg, "word/_rels/document.xml.rels", nil)
	addPart(pkg, "word/styles.xml", defaultStyles)
	addPart(pkg, "word/settings.xml", defaultSettings)
	addPart(pkg, "word/webSettings.xml", defaultWebSettings)
	addPart(pkg, "word/fontTable.xml", defaultFontTable)
	addPart(pkg, "word/theme/theme1.xml", defaultTheme)

	pkg.Rels[""] = &opc.Relationships{
		Rels: []opc.Relationship{
			{ID: "rId1", Type: relOfficeDocument, Target: "word/document.xml"},
		},
	}
	pkg.Rels["word/document.xml"] = &opc.Relationships{
		Rels: []opc.Relationship{
			{ID: "rId1", Type: relStyles, Target: "styles.xml"},
			{ID: "rId2", Type: relSettings, Target: "settings.xml"},
			{ID: "rId3", Type: relWebSettings, Target: "webSettings.xml"},
			{ID: "rId4", Type: relFontTable, Target: "fontTable.xml"},
			{ID: "rId5", Type: relTheme, Target: "theme/theme1.xml"},
		},
	}

	return pkg
}

// newTemplateTarget returns a fresh OPC package with infrastructure
// parts (content types, relationships) but NO style parts — no
// word/styles.xml, settings.xml, webSettings.xml, fontTable.xml, or
// theme/theme1.xml.  This satisfies CloneStyles' fresh-empty-target
// precondition (D-08) for FromTemplate/OpenTemplate operations.
//
// The package has:
//   - [Content_Types].xml (defaults: rels/xml; override: document.xml)
//   - _rels/.rels (root → word/document.xml)
//   - word/_rels/document.xml.rels (empty — CloneStyles adds relationships)
//
// word/document.xml must be added via MarkModified before Save.
func newTemplateTarget() *opc.Package {
	pkg := &opc.Package{
		Parts: make(map[string]*opc.Part),
		ContentTypes: &opc.ContentTypes{
			Defaults: map[string]string{
				"rels": ctRels,
				"xml":  ctXML,
			},
			Overrides: map[string]string{
				"/word/document.xml": ctMain,
			},
		},
		Rels:        make(map[string]*opc.Relationships),
		Conformance: opc.Transitional,
	}

	addPart(pkg, "[Content_Types].xml", nil)
	addPart(pkg, "_rels/.rels", nil)
	addPart(pkg, "word/_rels/document.xml.rels", nil)

	pkg.Rels[""] = &opc.Relationships{
		Rels: []opc.Relationship{
			{ID: "rId1", Type: relOfficeDocument, Target: "word/document.xml"},
		},
	}
	pkg.Rels["word/document.xml"] = &opc.Relationships{
		Rels: []opc.Relationship{},
	}

	return pkg
}

func addPart(pkg *opc.Package, name string, data []byte) {
	pkg.MarkModified(name, data)
}

func buildDocumentXML() []byte {
	sz12240 := int64(12240)
	sz15840 := int64(15840)
	margin1440 := int64(1440)
	valNormal := "Normal"

	doc := &wml.CT_Document{
		Body: &wml.CT_Body{
			P: []*wml.CT_P{
				{
					PPr: &wml.CT_PPr{
						PStyle: &wml.CT_PStyle{Val: &valNormal},
					},
				},
			},
			SectPr: &wml.CT_SectPr{
				PgSz: &wml.CT_PgSz{
					W: &sz12240,
					H: &sz15840,
				},
				PgMar: &wml.CT_PgMar{
					Top:    &margin1440,
					Right:  &margin1440,
					Bottom: &margin1440,
					Left:   &margin1440,
				},
				Cols:    &wml.CT_Cols{},
				DocGrid: &wml.CT_DocGrid{},
			},
		},
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(doc); err != nil {
		panic(fmt.Sprintf("wordingo: encode document.xml: %v", err))
	}
	if err := enc.Flush(); err != nil {
		panic(fmt.Sprintf("wordingo: flush document.xml: %v", err))
	}
	return buf.Bytes()
}

func strPtr(s string) *string { return &s }

func ptrInt64(v int64) *int64 { return &v }

// defaultSectPr returns a CT_SectPr with Letter page size (12240×15840 twips),
// 1-inch margins (1440 twips), empty columns, and no header/footer references.
// Used by buildEmptyBodyXML and OpenTemplateReader (Pitfall 4 — no HdrFtrRef/FtrRef).
func defaultSectPr() *wml.CT_SectPr {
	return &wml.CT_SectPr{
		PgSz:    &wml.CT_PgSz{W: ptrInt64(12240), H: ptrInt64(15840)},
		PgMar:   &wml.CT_PgMar{Top: ptrInt64(1440), Right: ptrInt64(1440), Bottom: ptrInt64(1440), Left: ptrInt64(1440)},
		Cols:    &wml.CT_Cols{},
		DocGrid: &wml.CT_DocGrid{},
	}
}

// buildEmptyBodyXML returns serialized CT_Document with no paragraphs or tables,
// only a single CT_SectPr with Letter page + 1in margins. Used by FromTemplate
// (CREATE-03) to provide a clean body with template styles.
func buildEmptyBodyXML() []byte {
	doc := &wml.CT_Document{
		Body: &wml.CT_Body{
			SectPr: defaultSectPr(),
		},
	}
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(doc); err != nil {
		panic(fmt.Sprintf("wordingo: encode empty body: %v", err))
	}
	if err := enc.Flush(); err != nil {
		panic(fmt.Sprintf("wordingo: flush empty body: %v", err))
	}
	return buf.Bytes()
}
