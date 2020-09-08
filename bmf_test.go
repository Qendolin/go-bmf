package bmf_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Qendolin/go-bmf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var Expected = bmf.Font{
	Info: bmf.Info{
		Face:     "Arial",
		Size:     -26,
		Bold:     true,
		Italic:   true,
		Charset:  "",
		Unicode:  true,
		StretchH: 90,
		Smooth:   true,
		AA:       2,
		Padding: bmf.Padding{
			Up:    1,
			Right: 2,
			Down:  3,
			Left:  4,
		},
		Spacing: bmf.Spacing{
			Horizontal: 2,
			Vertical:   1,
		},
		Outline: 2,
	},
	Common: bmf.Common{
		LineHeight:   27,
		Base:         22,
		ScaleW:       32,
		ScaleH:       64,
		Pages:        2,
		Packed:       false,
		AlphaChannel: bmf.Glyph,
		RedChannel:   bmf.Outline,
		GreenChannel: bmf.Zero,
		BlueChannel:  bmf.One,
	},
	Pages: []bmf.Page{
		{
			Id:   0,
			File: "test-bin_0.png",
		},
		{
			Id:   1,
			File: "test-bin_1.png",
		},
	},
	Chars: []bmf.Char{
		{
			Id: -1,
			X:  0, Y: 26,
			Width: 24, Height: 23,
			XOffset: -3, YOffset: 4,
			XAdvance: 19,
			Page:     1, Channel: bmf.All,
		},
		{
			Id: 65,
			X:  0, Y: 0,
			Width: 29, Height: 25,
			XOffset: -6, YOffset: 2,
			XAdvance: 19,
			Page:     0, Channel: bmf.All,
		},
		{
			Id: 84,
			X:  0, Y: 0,
			Width: 26, Height: 25,
			XOffset: -3, YOffset: 2,
			XAdvance: 16,
			Page:     1, Channel: bmf.All,
		},
		{
			Id: 86,
			X:  0, Y: 26,
			Width: 29, Height: 25,
			XOffset: -4, YOffset: 2,
			XAdvance: 17,
			Page:     0, Channel: bmf.All,
		},
	},
	Kernings: []bmf.Kerning{
		{First: 86, Second: 65, Amount: -2},
		{First: 84, Second: 65, Amount: -2},
		{First: 65, Second: 86, Amount: -2},
		{First: 65, Second: 84, Amount: -2},
	},
}

func TestXML(t *testing.T) {
	f, err := os.Open("./testdata/test-xml.fnt")
	require.NoErrorf(t, err, "Unable to open testdata")
	data, err := ioutil.ReadAll(f)
	require.NoErrorf(t, err, "Unable to read testdata")
	fnt, err := bmf.ParseXML(data)
	require.NoError(t, err)
	assertFontEqual(t, Expected, *fnt)
}

func TestBinary(t *testing.T) {
	f, err := os.Open("./testdata/test-bin.fnt")
	require.NoErrorf(t, err, "Unable to open testdata")
	data, err := ioutil.ReadAll(f)
	require.NoErrorf(t, err, "Unable to read testdata")
	fnt, err := bmf.ParseBinary(data)
	require.NoError(t, err)
	assertFontEqual(t, Expected, *fnt)
}

func TestText(t *testing.T) {
	f, err := os.Open("./testdata/test-text.fnt")
	require.NoErrorf(t, err, "Unable to open testdata")
	data, err := ioutil.ReadAll(f)
	require.NoErrorf(t, err, "Unable to read testdata")
	fnt, err := bmf.ParseText(data)
	require.NoError(t, err)
	assertFontEqual(t, Expected, *fnt)
}

func assertFontEqual(t *testing.T, expected bmf.Font, actual bmf.Font) {
	assert.Equal(t, expected.Info, actual.Info)
	assert.Equal(t, expected.Common, actual.Common)
	assert.Equal(t, expected.Pages, actual.Pages)
	assert.Equal(t, expected.Chars, actual.Chars)
	assert.Equal(t, expected.Kernings, actual.Kernings)
}
