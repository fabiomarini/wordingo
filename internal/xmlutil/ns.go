// Package xmlutil extends encoding/xml with OOXML namespace handling:
// URI↔prefix registry, safe decoder, RawXML token capture, and
// canonical-prefix encoder.
package xmlutil

// CanonicalPrefixes maps OOXML namespace URIs to their canonical short
// prefixes (D-08).  Both Transitional (schemas.openxmlformats.org) and
// Strict (purl.oclc.org) URI families are registered and map to the
// same prefixes.
var CanonicalPrefixes = map[string]string{
	// WordprocessingML
	"http://schemas.openxmlformats.org/wordprocessingml/2006/main": "w",
	"http://purl.oclc.org/ooxml/wordprocessingml/main":            "w",

	// Office Document Relationships
	"http://schemas.openxmlformats.org/officeDocument/2006/relationships": "r",
	"http://purl.oclc.org/ooxml/officeDocument/relationships":            "r",

	// DrawingML
	"http://schemas.openxmlformats.org/drawingml/2006/main": "a",
	"http://purl.oclc.org/ooxml/drawingml/main":            "a",

	// Wordprocessing Drawing
	"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing": "wp",

	// Markup Compatibility
	"http://schemas.openxmlformats.org/markup-compatibility/2006": "mc",

	// Extended WML (Office 2010)
	"http://schemas.microsoft.com/office/word/2010/wordml": "w14",

	// Extended WML (Office 2012)
	"http://schemas.microsoft.com/office/word/2012/wordml": "w15",

	// Wordprocessing Drawing (Office 2010)
	"http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing": "wp14",

	// Wordprocessing Shape (Office 2010)
	"http://schemas.microsoft.com/office/word/2010/wordprocessingShape": "wps",

	// VML
	"urn:schemas-microsoft-com:vml": "v",

	// Office
	"urn:schemas-microsoft-com:office:office": "o",

	// XML namespace (always-present, no prefix rewriting needed)
	"http://www.w3.org/XML/1998/namespace": "xml",

	// Package relationships (fixed URIs, no canonical prefix needed in output)
	"http://schemas.openxmlformats.org/package/2006/relationships": "r",

	// Strict package relationships
	"http://purl.oclc.org/ooxml/package/2006/relationships": "r",
}

// strictToTransitional maps Strict purl.oclc.org URIs to their
// Transitional equivalents.
var strictToTransitional = map[string]string{
	"http://purl.oclc.org/ooxml/wordprocessingml/main":            "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
	"http://purl.oclc.org/ooxml/officeDocument/relationships":     "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
	"http://purl.oclc.org/ooxml/drawingml/main":                  "http://schemas.openxmlformats.org/drawingml/2006/main",
	"http://purl.oclc.org/ooxml/package/2006/relationships":      "http://schemas.openxmlformats.org/package/2006/relationships",
	"http://purl.oclc.org/ooxml/package/2006/content-types":      "http://schemas.openxmlformats.org/package/2006/content-types",
}

// PrefixFor returns the canonical prefix for uri, or "" if unknown.
func PrefixFor(uri string) string {
	if p, ok := CanonicalPrefixes[uri]; ok {
		return p
	}
	return ""
}

// NormalizeURI converts a Strict purl.oclc.org URI to its Transitional
// equivalent.  Transitional and unknown URIs are returned unchanged.
func NormalizeURI(uri string) string {
	if t, ok := strictToTransitional[uri]; ok {
		return t
	}
	return uri
}

// ExtensionPrefixes lists prefixes that should appear in mc:Ignorable.
var ExtensionPrefixes = []string{"w14", "w15", "wp14"}
