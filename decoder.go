package xml2json

import (
	"encoding/xml"
	"io"
	"unicode"
)

const (
	attrPrefix    = "-"
	contentPrefix = "#"
)

// A Decoder reads and decodes XML objects from an input stream.
type Decoder struct {
	r   io.Reader
	err error
}

type element struct {
	parent *element
	n      *Node
	label  string
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the next JSON-encoded value from its
// input and stores it in the value pointed to by v.
func (dec *Decoder) Decode(root *Node) error {
	xmlDec := xml.NewDecoder(dec.r)

	// Create first element from the root node
	elem := &element{
		parent: nil,
		n:      root,
	}

	for {
		t, _ := xmlDec.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			// Build new a new current element and link it to its parent
			elem = &element{
				parent: elem,
				n:      &Node{},
				label:  se.Name.Local,
			}

			// Extract attributes as children
			for _, a := range se.Attr {
				elem.n.AddChild(attrPrefix+a.Name.Local, &Node{Data: a.Value})
			}
		case xml.CharData:
			// Extract XML data (if any)
			elem.n.Data = trimNonGraphic(string(xml.CharData(se)))
		case xml.EndElement:
			// And add it to its parent list
			if elem.parent != nil {
				elem.parent.n.AddChild(elem.label, elem.n)
			}

			// Then change the current element to its parent
			elem = elem.parent
		}
	}

	return nil
}

// trimNonGraphic returns a slice of the string s, with all leading and trailing
// non graphic characters and spaces removed.
//
// Graphic characters include letters, marks, numbers, punctuation, symbols,
// and spaces, from categories L, M, N, P, S, Zs.
// Spacing characters are set by category Z and property Pattern_White_Space.
func trimNonGraphic(s string) string {
	if s == "" {
		return s
	}

	var first *int
	var last int
	for i, r := range s {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) {
			continue
		}

		if first == nil {
			f := i // copy i
			first = &f
			last = i
		} else {
			last = i
		}
	}

	// If first is nil, it means there are no graphic characters
	if first == nil {
		return ""
	}

	return s[*first : last+1]
}
