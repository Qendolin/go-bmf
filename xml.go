package bmf

import (
	"encoding/xml"
)

func (pad *Padding) UnmarshalXMLAttr(attr xml.Attr) error {
	*pad = parsePadding(attr.Value)
	return nil
}

func (sp *Spacing) UnmarshalXMLAttr(attr xml.Attr) error {
	*sp = parseSpacing(attr.Value)
	return nil
}

func ParseXML(data []byte) (*Font, error) {
	fnt := &Font{}
	if err := xml.Unmarshal(data, fnt); err != nil {
		return nil, err
	}
	return fnt, nil
}
