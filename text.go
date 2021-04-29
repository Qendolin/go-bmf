package bmf

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// TextParseError contains info about where and why a parsing error occured
type TextParseError struct {
	LineNumber int
	Line       string
	Err        error
}

func (e TextParseError) Error() string {
	return fmt.Sprintf("format error in line %v: '%v'", e.LineNumber, e.Line)
}

func (e TextParseError) Unwrap() error {
	return e.Err
}

// ParseText parses a bmf font file in text format
func ParseText(data []byte) (fnt *Font, err error) {
	var lineNr int
	var line string
	defer func() {
		if err != nil {
			err = TextParseError{
				Line:       line,
				LineNumber: lineNr,
				Err:        err,
			}
		}
	}()

	fnt = &Font{}

	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		lineNr++
		line = sc.Text()
		tag, attribs, strs, err := parseTagText(line)
		if err != nil {
			return nil, err
		}
		switch tag {
		case "info":
			fnt.Info = parseInfoText(attribs, strs)
		case "char":
			fnt.Chars = append(fnt.Chars, parseCharText(attribs))
		case "common":
			fnt.Common = parseCommonText(attribs)
		case "page":
			fnt.Pages = append(fnt.Pages, parsePageText(attribs, strs))
		case "kerning":
			fnt.Kernings = append(fnt.Kernings, parseKerningPairText(attribs))
		}
	}
	return fnt, sc.Err()
}

func parsePageText(attribs map[string]int, strs []string) Page {
	page := Page{}
	for k, v := range attribs {
		switch k {
		case "id":
			page.Id = v
		case "file":
			page.File = strs[v]
		}
	}
	return page
}

func parseInfoText(attribs map[string]int, strs []string) Info {
	info := Info{}
	for k, v := range attribs {
		switch k {
		case "size":
			info.Size = v
		case "face":
			info.Face = strs[v]
		case "bold":
			info.Bold = itob(v)
		case "italic":
			info.Italic = itob(v)
		case "charset":
			info.Charset = strs[v]
		case "unicode":
			info.Unicode = itob(v)
		case "stretchH":
			info.StretchH = v
		case "smooth":
			info.Smooth = itob(v)
		case "aa":
			info.AA = v
		case "padding":
			info.Padding = parsePadding(strs[v])
		case "spacing":
			info.Spacing = parseSpacing(strs[v])
		case "outline":
			info.Outline = v
		}
	}

	return info
}

func parseCharText(attribs map[string]int) Char {
	char := Char{}
	for k, v := range attribs {
		switch k {
		case "id":
			char.Id = rune(v)
		case "x":
			char.X = v
		case "y":
			char.Y = v
		case "width":
			char.Width = v
		case "height":
			char.Height = v
		case "xoffset":
			char.XOffset = v
		case "yoffset":
			char.YOffset = v
		case "xadvance":
			char.XAdvance = v
		case "page":
			char.Page = v
		case "chnl":
			char.Channel = Channel(v)
		}
	}

	return char
}

func parseCommonText(attribs map[string]int) Common {
	common := Common{}

	for k, v := range attribs {
		switch k {
		case "lineHeight":
			common.LineHeight = v
		case "base":
			common.Base = v
		case "scaleW":
			common.ScaleW = v
		case "scaleH":
			common.ScaleH = v
		case "pages":
			common.Pages = v
		case "packed":
			common.Packed = itob(v)
		case "alphaChnl":
			common.AlphaChannel = ChannelData(v)
		case "redChnl":
			common.RedChannel = ChannelData(v)
		case "greenChnl":
			common.GreenChannel = ChannelData(v)
		case "blueChnl":
			common.BlueChannel = ChannelData(v)
		}
	}

	return common
}

func parseKerningPairText(attribs map[string]int) Kerning {
	kern := Kerning{}
	for k, v := range attribs {
		switch k {
		case "first":
			kern.First = rune(v)
		case "second":
			kern.Second = rune(v)
		case "amount":
			kern.Amount = v
		}
	}
	return kern
}

func parseTagText(line string) (name string, values map[string]int, strs []string, err error) {
	values = map[string]int{}
	strs = []string{}

	var stripped string
	rd := bufio.NewReader(strings.NewReader(line))
	for {
		start, err := rd.ReadString('"')
		stripped += start
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return "", nil, nil, err
		}
		str, err := rd.ReadString('"')
		if errors.Is(err, io.EOF) {
			return "", nil, nil, fmt.Errorf("expected \"")
		}
		strs = append(strs, str[:len(str)-1])
	}

	fields := strings.Fields(stripped)
	if len(fields) == 0 {
		return "", nil, nil, fmt.Errorf("expected non empty tag")
	}

	strIdx := 0
	for i, f := range fields {
		if i == 0 {
			name = f
			continue
		}

		kv := strings.Split(f, "=")
		if len(kv) != 2 {
			return "", nil, nil, fmt.Errorf("expected key-value pair")
		}
		key, value := kv[0], kv[1]
		if value == "\"" {
			values[key] = strIdx
			strIdx++
		} else if num, err := strconv.Atoi(value); err == nil {
			values[key] = num
		} else {
			strs = append(strs, value)
			values[key] = len(strs) - 1
		}
	}

	return
}
