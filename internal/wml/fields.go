package wml

import (
	"encoding/xml"
	"strings"
)

// Field and bookmark structures used by the table-of-contents feature.

// CT_FldChar is a field character (w:fldChar): begin, separate, or end.
// The begin character of a TOC field carries w:dirty="true" so Word
// refreshes the field when the document is opened.
type CT_FldChar struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main fldChar"`
	Type    *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main fldCharType,attr,omitempty"`
	Dirty   *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main dirty,attr,omitempty"`
}

// CT_InstrText is field instruction text (w:instrText), e.g.
// ` TOC \o "1-3" \h \z \u `. Leading/trailing whitespace is
// significant to Word, so xml:space="preserve" is emitted when needed.
type CT_InstrText struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main instrText"`
	Space   *string  `xml:"http://www.w3.org/XML/1998/namespace space,attr"`
	Value   string   `xml:",chardata"`
}

// MarshalXML writes the instruction text, adding xml:space="preserve"
// when the value contains leading/trailing whitespace or double spaces.
func (t *CT_InstrText) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	needsSpace := (t.Space != nil && *t.Space == "preserve") ||
		(len(t.Value) > 0 && (t.Value[0] == ' ' || t.Value[len(t.Value)-1] == ' ' || strings.Contains(t.Value, "  ")))
	if needsSpace {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Space: NSXML, Local: "space"},
			Value: "preserve",
		})
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if err := e.EncodeToken(xml.CharData(t.Value)); err != nil {
		return err
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// CT_BookmarkStart is the opening half of a bookmark (w:bookmarkStart).
// Bookmark ids must be unique within a document; names are internal
// anchors for hyperlinks (Word uses "_Toc…" for TOC targets).
type CT_BookmarkStart struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bookmarkStart"`
	ID      *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr,omitempty"`
	Name    *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main name,attr,omitempty"`
}

// CT_BookmarkEnd is the closing half of a bookmark (w:bookmarkEnd).
type CT_BookmarkEnd struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main bookmarkEnd"`
	ID      *int64   `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main id,attr,omitempty"`
}

// CT_UpdateFields is the settings switch (w:updateFields) that asks
// Word to refresh all fields — including the table of contents — when
// the document is opened.
type CT_UpdateFields struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main updateFields"`
	Val     *string  `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main val,attr,omitempty"`
}
