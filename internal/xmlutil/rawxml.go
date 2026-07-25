package xmlutil

import (
	"encoding/xml"
	"fmt"
)

// RawXML captures an XML element subtree as buffered tokens (D-06).
// It implements xml.Marshaler and xml.Unmarshaler for use as the
// `xml:",any"` catch-all field type on WML container structs.
//
// On unmarshal the entire element subtree (start → matching end) is
// captured token-by-token via xml.CopyToken, preserving attribute
// order and namespace declarations.
//
// On marshal the captured tokens are replayed verbatim through the
// xml.Encoder.  Unknown namespaces carry their own xmlns declarations
// from the source and are never remapped.
type RawXML struct {
	Tokens []xml.Token
}

// UnmarshalXML captures start and all tokens until the matching end
// element.  Each token is Deep-copied so the capture is independent
// of the decoder's buffer.
func (r *RawXML) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	r.Tokens = append(r.Tokens[:0], xml.CopyToken(start))
	depth := 1
	for depth > 0 {
		tok, err := d.Token()
		if err != nil {
			return fmt.Errorf("rawxml: %w", err)
		}
		tok = xml.CopyToken(tok)
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
		r.Tokens = append(r.Tokens, tok)
	}
	return nil
}

// MarshalXML replays all captured tokens verbatim.
func (r *RawXML) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	for _, tok := range r.Tokens {
		if err := e.EncodeToken(tok); err != nil {
			return fmt.Errorf("rawxml: %w", err)
		}
	}
	return nil
}

// compile-time interface checks
var _ xml.Marshaler = (*RawXML)(nil)
var _ xml.Unmarshaler = (*RawXML)(nil)
