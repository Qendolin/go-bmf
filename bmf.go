// Package bmf implements BMF .fnt file parsing.
// It supports only version 3 of the binary format.
// For more information see http://www.angelcode.com/products/bmfont/doc/file_format.html
package bmf

import (
	"bytes"
	"io"
	"strconv"
)

// BinBool represents a boolean as 0 or 1 in xml
type BinBool bool

// ChannelData specifies the type of data that a color channel holds
type ChannelData int

// Channel data types
const (
	Glyph ChannelData = iota
	Outline
	GlyphAndOutline
	Zero
	One
)

// Channel is a bitfield to specify color channels
type Channel int

// RGBA channel bits
const (
	Blue  Channel = 0x1
	Green Channel = 0x2
	Red   Channel = 0x4
	Alpha Channel = 0x8
	All   Channel = 0xf
)

// Padding specifies the padding for each character in pixels
// See https://www.angelcode.com/products/bmfont/doc/export_options.html
type Padding struct {
	Up    int
	Right int
	Down  int
	Left  int
}

// Spacing specifies the spacing for each character in pixels
// See https://www.angelcode.com/products/bmfont/doc/export_options.html
type Spacing struct {
	Horizontal int
	Vertical   int
}

// Font defines an AngelCode Bitmap Font
type Font struct {
	Info     Info      `xml:"info"`
	Common   Common    `xml:"common"`
	Pages    []Page    `xml:"pages>page"`
	Chars    []Char    `xml:"chars>char"`
	Kernings []Kerning `xml:"kernings>kerning"`
}

// Info holds information on how the font was generated
type Info struct {
	Face     string  `xml:"face,attr"`
	Size     int     `xml:"size,attr"`
	Bold     BinBool `xml:"bold,attr"`
	Italic   BinBool `xml:"italic,attr"`
	Charset  string  `xml:"charset,attr"`
	Unicode  BinBool `xml:"unicode,attr"`
	StretchH int     `xml:"stretchH,attr"`
	Smooth   BinBool `xml:"smooth,attr"`
	AA       int     `xml:"aa,attr"`
	Padding  Padding `xml:"padding,attr"`
	Spacing  Spacing `xml:"spacing,attr"`
	Outline  int     `xml:"outline,attr"`
}

// Common holds information common to all characters.
type Common struct {
	LineHeight   int         `xml:"lineHeight,attr"`
	Base         int         `xml:"base,attr"`
	ScaleW       int         `xml:"scaleW,attr"`
	ScaleH       int         `xml:"scaleH,attr"`
	Pages        int         `xml:"pages,attr"`
	Packed       BinBool     `xml:"packed,attr"`
	AlphaChannel ChannelData `xml:"alphaChnl,attr"`
	RedChannel   ChannelData `xml:"redChnl,attr"`
	GreenChannel ChannelData `xml:"greenChnl,attr"`
	BlueChannel  ChannelData `xml:"blueChnl,attr"`
}

// Char describes on character in the font. There is one for each included character in the font.
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

// Kerning specifies the distance between specific character pairs
type Kerning struct {
	First  rune `xml:"first,attr"`
	Second rune `xml:"second,attr"`
	Amount int  `xml:"amount,attr"`
}

// Page references a bitmap image that contains the glyphs.
// A font can contain multiple glyph pages.
type Page struct {
	Id   int    `xml:"id,attr"`
	File string `xml:"file,attr"`
}

func atoi(i *int, a string) {
	v, err := strconv.Atoi(a)
	if err == nil {
		*i = v
	} else {
		*i = 0
	}
}

func Bool(i int) BinBool {
	return i == 1
}

func (b BinBool) Byte() byte {
	if b {
		return 1
	}
	return 0
}

// Parse parses a bmf font file and detects the format automatically
func Parse(src io.Reader) (*Font, error) {
	start := make([]byte, 5)

	if _, err := src.Read(start); err != nil {
		return nil, err
	}

	src = io.MultiReader(bytes.NewBufferString(string(start)), src)

	if string(start[:3]) == "BMF" {
		return ParseBinary(src)
	}
	if string(start[:5]) == "<?xml" {
		return ParseXML(src)
	}
	return ParseText(src)
}
