package bmf

import (
	"encoding/xml"
)

// Parses the padding string <up>,<right>,<down>,<left> to a Padding struct
func (pad *Padding) UnmarshalXMLAttr(attr xml.Attr) error {
	*pad = parsePadding(attr.Value)
	return nil
}

// Parses the spacing string <up>,<right>,<down>,<left> to a Spacing struct
func (sp *Spacing) UnmarshalXMLAttr(attr xml.Attr) error {
	*sp = parseSpacing(attr.Value)
	return nil
}

// Parses a bmf font file in XML format
func ParseXML(data []byte) (*Font, error) {
	fnt := &Font{}
	if err := xml.Unmarshal(data, fnt); err != nil {
		return nil, err
	}
	return fnt, nil
}
