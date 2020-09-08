// package bmf implements BMF .fnt file parsing.
// It supports only version 3 of the binary format.
// For more infromation see http://www.angelcode.com/products/bmfont/doc/file_format.html
package bmf

import (
	"fmt"
	"strconv"
	"strings"
)

type ChannelData int

const (
	Glyph ChannelData = iota
	Outline
	GlyphAndOutline
	Zero
	One
)

type Channel int

const (
	Blue  = 0x1
	Green = 0x2
	Red   = 0x4
	Alpha = 0x8
	All   = 0xf
)

type Padding struct {
	Up    int
	Right int
	Down  int
	Left  int
}

type Spacing struct {
	Horizontal int
	Vertical   int
}

type Font struct {
	Info     Info      `xml:"info"`
	Common   Common    `xml:"common"`
	Pages    []Page    `xml:"pages>page"`
	Chars    []Char    `xml:"chars>char"`
	Kernings []Kerning `xml:"kernings>kerning"`
}

type Info struct {
	Face     string  `xml:"face,attr"`
	Size     int     `xml:"size,attr"`
	Bold     bool    `xml:"bold,attr"`
	Italic   bool    `xml:"italic,attr"`
	Charset  string  `xml:"charset,attr"`
	Unicode  bool    `xml:"unicode,attr"`
	StretchH int     `xml:"stretchH,attr"`
	Smooth   bool    `xml:"smooth,attr"`
	AA       int     `xml:"aa,attr"`
	Padding  Padding `xml:"padding,attr"`
	Spacing  Spacing `xml:"spacing,attr"`
	Outline  int     `xml:"outline,attr"`
}

type Common struct {
	LineHeight   int         `xml:"lineHeight,attr"`
	Base         int         `xml:"base,attr"`
	ScaleW       int         `xml:"scaleW,attr"`
	ScaleH       int         `xml:"scaleH,attr"`
	Pages        int         `xml:"pages,attr"`
	Packed       bool        `xml:"packed,attr"`
	AlphaChannel ChannelData `xml:"alphaChnl,attr"`
	RedChannel   ChannelData `xml:"redChnl,attr"`
	GreenChannel ChannelData `xml:"greenChnl,attr"`
	BlueChannel  ChannelData `xml:"blueChnl,attr"`
}

type Char struct {
	Id       rune    `xml:"id,attr"`
	X        int     `xml:"x,attr"`
	Y        int     `xml:"y,attr"`
	Width    int     `xml:"width,attr"`
	Height   int     `xml:"height,attr"`
	XOffset  int     `xml:"xoffset,attr"`
	YOffset  int     `xml:"yoffset,attr"`
	XAdvance int     `xml:"xadvance,attr"`
	Page     int     `xml:"page,attr"`
	Channel  Channel `xml:"chnl,attr"`
}

type Kerning struct {
	First  rune `xml:"first,attr"`
	Second rune `xml:"second,attr"`
	Amount int  `xml:"amount,attr"`
}

type Page struct {
	Id   int    `xml:"id,attr"`
	File string `xml:"file,attr"`
}

func parsePadding(s string) Padding {
	pad := Padding{}

	v := strings.SplitN(s, ",", 4)
	v = append(v, "0", "0", "0", "0")

	atoi(&pad.Up, v[0])
	atoi(&pad.Right, v[1])
	atoi(&pad.Down, v[2])
	atoi(&pad.Left, v[3])

	return pad
}

func parseSpacing(s string) Spacing {
	sp := Spacing{}

	v := strings.SplitN(s, ",", 4)
	v = append(v, "0", "0")

	atoi(&sp.Horizontal, v[0])
	atoi(&sp.Vertical, v[1])

	return sp
}

func atoi(i *int, a string) {
	v, err := strconv.Atoi(a)
	if err == nil {
		*i = v
	} else {
		*i = 0
	}
}

func itob(i int) bool {
	if i == 1 {
		return true
	}
	return false
}

// Parses bmf font file and detects the format automatically
func Parse(data []byte) (*Font, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("data must have length of at least 5 bytes")
	}

	if string(data[:3]) == "BMF" {
		return ParseBinary(data)
	}
	if string(data[:5]) == "<?xml" {
		return ParseXML(data)
	}
	return ParseText(data)
}
