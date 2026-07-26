package wml

import (
	"encoding/xml"

	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// CT_Hyperlink is a hyperlink element (w:hyperlink) with an r:id relationship
// attribute and child runs.
type CT_Hyperlink struct {
	XMLName xml.Name        `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main hyperlink"`
	ID      string          `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr,omitempty"`
	Anchor  *string         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main anchor,attr,omitempty"`
	R       []*CT_R         `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
	Raw     []xmlutil.RawXML `xml:",any"`
}
