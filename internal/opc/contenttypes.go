package opc

import (
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

// contentTypesNS is the fixed namespace of [Content_Types].xml.
const contentTypesNS = "http://schemas.openxmlformats.org/package/2006/content-types"

// ContentTypes models [Content_Types].xml: extension Defaults plus
// per-part Overrides. Part names in Overrides are OPC-absolute
// ("/word/document.xml").
type ContentTypes struct {
	Defaults  map[string]string // extension (no dot) -> MIME type
	Overrides map[string]string // part name (leading /) -> MIME type
}

type ctDefault struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

type ctOverride struct {
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}

type ctXML struct {
	XMLName   xml.Name     `xml:"http://schemas.openxmlformats.org/package/2006/content-types Types"`
	Defaults  []ctDefault  `xml:"http://schemas.openxmlformats.org/package/2006/content-types Default"`
	Overrides []ctOverride `xml:"http://schemas.openxmlformats.org/package/2006/content-types Override"`
}

// parseContentTypes decodes [Content_Types].xml from r.
func parseContentTypes(r io.Reader) (*ContentTypes, error) {
	var x ctXML
	if err := xml.NewDecoder(r).Decode(&x); err != nil {
		return nil, fmt.Errorf("opc: parse [Content_Types].xml: %w", err)
	}
	ct := &ContentTypes{
		Defaults:  make(map[string]string, len(x.Defaults)),
		Overrides: make(map[string]string, len(x.Overrides)),
	}
	for _, d := range x.Defaults {
		ct.Defaults[d.Extension] = d.ContentType
	}
	for _, o := range x.Overrides {
		ct.Overrides[o.PartName] = o.ContentType
	}
	return ct, nil
}

// TypeFor returns the MIME type for a package-relative part name,
// checking Overrides first, then Defaults by extension.
// Empty string if unknown.
func (c *ContentTypes) TypeFor(partName string) string {
	if t, ok := c.Overrides["/"+partName]; ok {
		return t
	}
	ext := ""
	for i := len(partName) - 1; i >= 0; i-- {
		if partName[i] == '.' {
			ext = partName[i+1:]
			break
		}
		if partName[i] == '/' {
			break
		}
	}
	return c.Defaults[ext]
}

// serialize renders [Content_Types].xml as a canonical XML byte slice.
// Builds XML directly to avoid Go's encoding/xml redundant xmlns
// on child elements — Word's OPC parser rejects those.
func (c *ContentTypes) serialize() ([]byte, error) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteByte('\n')
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)

	exts := make([]string, 0, len(c.Defaults))
	for e := range c.Defaults {
		exts = append(exts, e)
	}
	sort.Strings(exts)
	for _, e := range exts {
		fmt.Fprintf(&b, `<Default Extension="%s" ContentType="%s"></Default>`,
			xmlEscape(e), xmlEscape(c.Defaults[e]))
	}

	names := make([]string, 0, len(c.Overrides))
	for n := range c.Overrides {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(&b, `<Override PartName="%s" ContentType="%s"></Override>`,
			xmlEscape(n), xmlEscape(c.Overrides[n]))
	}

	b.WriteString(`</Types>`)
	return []byte(b.String()), nil
}

// xmlEscape escapes special characters for XML attribute values.
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
