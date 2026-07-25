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

const (
	ctMain      = "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"
	ctStyles    = "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"
	ctSettings  = "application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"
	ctFontTable = "application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"
	ctTheme     = "application/vnd.openxmlformats-officedocument.theme+xml"
	ctRels      = "application/vnd.openxmlformats-package.relationships+xml"
	ctXML       = "application/xml"
)

const (
	relOfficeDocument = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	relStyles         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"
	relSettings       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings"
	relFontTable      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable"
	relTheme          = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme"
)

// newBlankPackage constructs a fresh opc.Package holding the 8-part
// minimal blank-document set: content types, root rels, document.xml
// (built from wml structs via xmlutil.Encoder), document.xml.rels, and
// four embedded default parts (styles, settings, fontTable, theme1).
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
				"/word/fontTable.xml":    ctFontTable,
				"/word/theme/theme1.xml": ctTheme,
			},
		},
		Rels:         make(map[string]*opc.Relationships),
		Conformance:  opc.Transitional,
	}

	// Build document.xml from wml structs — exercises the full
	// wml + xmlutil.Encoder write path.
	docXML := buildDocumentXML()

	// Register the 8 parts.  [Content_Types].xml and the two .rels
	// parts are re-serialized by opc.Package.Save from the ContentTypes
	// and Rels fields — they still need entries in Parts so the save
	// loop emits them.
	addPart(pkg, "[Content_Types].xml", nil)
	addPart(pkg, "_rels/.rels", nil)
	addPart(pkg, "word/document.xml", docXML)
	addPart(pkg, "word/_rels/document.xml.rels", nil)
	addPart(pkg, "word/styles.xml", defaultStyles)
	addPart(pkg, "word/settings.xml", defaultSettings)
	addPart(pkg, "word/fontTable.xml", defaultFontTable)
	addPart(pkg, "word/theme/theme1.xml", defaultTheme)

	// Register relationships.
	pkg.Rels[""] = &opc.Relationships{
		Rels: []opc.Relationship{
			{ID: "rId1", Type: relOfficeDocument, Target: "word/document.xml"},
		},
	}
	pkg.Rels["word/document.xml"] = &opc.Relationships{
		Rels: []opc.Relationship{
			{ID: "rId1", Type: relStyles, Target: "styles.xml"},
			{ID: "rId2", Type: relSettings, Target: "settings.xml"},
			{ID: "rId3", Type: relFontTable, Target: "fontTable.xml"},
			{ID: "rId4", Type: relTheme, Target: "theme/theme1.xml"},
		},
	}

	return pkg
}

// addPart inserts or marks a part as modified so that Save writes
// data (or the override payload) instead of raw-copying.
func addPart(pkg *opc.Package, name string, data []byte) {
	pkg.MarkModified(name, data)
}

// buildDocumentXML constructs word/document.xml from wml structs,
// serialized through xmlutil.Encoder for canonical w: prefix output.
func buildDocumentXML() []byte {
	sz12240 := int64(12240)
	sz15840 := int64(15840)
	margin1440 := int64(1440)

	doc := &wml.CT_Document{
		Body: &wml.CT_Body{
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
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(doc); err != nil {
		panic(fmt.Sprintf("wordingo: encode document.xml: %v", err))
	}
	if err := enc.Flush(); err != nil {
		panic(fmt.Sprintf("wordingo: flush document.xml: %v", err))
	}
	return buf.Bytes()
}


