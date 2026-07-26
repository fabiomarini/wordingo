// Package wml implements the ~60 essential WordprocessingML struct
// types (CT_*) from ISO/IEC 29500 Part 1, with URI-based struct tags,
// whitespace-fidelity CT_Text, and RawXML hoarding for unknown
// children (WML-01..04).
package wml

// Namespace URI constants used in WML struct tags.
const (
	// NSWMLMain is the WordprocessingML main namespace
	// (Transitional conformance).
	NSWMLMain = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

	// NSRels is the Office Document Relationships namespace.
	NSRels = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"

	// NSDrawingML is the DrawingML main namespace.
	NSDrawingML = "http://schemas.openxmlformats.org/drawingml/2006/main"

	// NSWP is the Wordprocessing Drawing namespace.
	NSWP = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"

	// NSMC is the Markup Compatibility namespace.
	NSMC = "http://schemas.openxmlformats.org/markup-compatibility/2006"

	// NSW14 is the Office 2010 WordprocessingML extensions namespace.
	NSW14 = "http://schemas.microsoft.com/office/word/2010/wordml"

	// NSW15 is the Office 2012 WordprocessingML extensions namespace.
	NSW15 = "http://schemas.microsoft.com/office/word/2012/wordml"

	// NSXML is the xml: namespace (xml:space, xml:lang, etc.)
	NSXML = "http://www.w3.org/XML/1998/namespace"

	// NSPicture is the DrawingML picture namespace.
	NSPicture = "http://schemas.openxmlformats.org/drawingml/2006/picture"
)
