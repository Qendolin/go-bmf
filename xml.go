package bmf

import (
	"encoding/xml"
	"fmt"
)

// Converts from format <up>,<right>,<down>,<left>
func (pad *Padding) UnmarshalXMLAttr(attr xml.Attr) error {
	*pad = parsePadding(attr.Value)
	return nil
}

// Converts to format <up>,<right>,<down>,<left>
func (pad Padding) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: fmt.Sprintf("%d,%d,%d,%d", pad.Up, pad.Right, pad.Down, pad.Left)}, nil
}

// Converts from format <horizontal>,<vertical>
func (sp *Spacing) UnmarshalXMLAttr(attr xml.Attr) error {
	*sp = parseSpacing(attr.Value)
	return nil
}

// Converts to format <horizontal>,<vertical>
func (sp Spacing) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: fmt.Sprintf("%d,%d", sp.Horizontal, sp.Vertical)}, nil
}

// Converts bool to number
func (nb BinBool) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if nb {
		return xml.Attr{Name: name, Value: "1"}, nil
	} else {
		return xml.Attr{Name: name, Value: "0"}, nil
	}
}

type xmlFont struct {
	XMLName  xml.Name  `xml:"font"`
	Info     Info      `xml:"info"`
	Common   Common    `xml:"common"`
	Pages    []Page    `xml:"pages>page"`
	Chars    []Char    `xml:"chars>char"`
	Kernings []Kerning `xml:"kernings>kerning"`
}

func (font Font) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "font"
	return e.EncodeElement(xmlFont{
		Info:     font.Info,
		Common:   font.Common,
		Pages:    font.Pages,
		Chars:    font.Chars,
		Kernings: font.Kernings,
	}, start)
}

// Parses a bmf font file in XML format
func ParseXML(data []byte) (*Font, error) {
	fnt := &Font{}
	if err := xml.Unmarshal(data, fnt); err != nil {
		return nil, err
	}
	return fnt, nil
}
