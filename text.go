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

type TextParseError struct {
	LineNumber int
	Line       string
	Err        error
}

func (e TextParseError) Error() string {
	return fmt.Sprintf("forrmat error in line %v: '%v'", e.LineNumber, e.Line)
}

func (e TextParseError) Unwrap() error {
	return e.Err
}

// Parses a bmf font file in text format
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
		tag, attribs, strs, err := parseTag(line)
		if err != nil {
			return nil, err
		}
		switch tag {
		case "info":
			for k, v := range attribs {
				switch k {
				case "size":
					fnt.Info.Size = v
				case "face":
					fnt.Info.Face = strs[v]
				case "bold":
					fnt.Info.Bold = itob(v)
				case "italic":
					fnt.Info.Italic = itob(v)
				case "charset":
					fnt.Info.Charset = strs[v]
				case "unicode":
					fnt.Info.Unicode = itob(v)
				case "stretchH":
					fnt.Info.StretchH = v
				case "smooth":
					fnt.Info.Smooth = itob(v)
				case "aa":
					fnt.Info.AA = v
				case "padding":
					fnt.Info.Padding = parsePadding(strs[v])
				case "spacing":
					fnt.Info.Spacing = parseSpacing(strs[v])
				case "outline":
					fnt.Info.Outline = v
				}
			}
		case "char":
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
			fnt.Chars = append(fnt.Chars, char)
		case "common":
			for k, v := range attribs {
				switch k {
				case "lineHeight":
					fnt.Common.LineHeight = v
				case "base":
					fnt.Common.Base = v
				case "scaleW":
					fnt.Common.ScaleW = v
				case "scaleH":
					fnt.Common.ScaleH = v
				case "pages":
					fnt.Common.Pages = v
				case "packed":
					fnt.Common.Packed = itob(v)
				case "alphaChnl":
					fnt.Common.AlphaChannel = ChannelData(v)
				case "redChnl":
					fnt.Common.RedChannel = ChannelData(v)
				case "greenChnl":
					fnt.Common.GreenChannel = ChannelData(v)
				case "blueChnl":
					fnt.Common.BlueChannel = ChannelData(v)
				}
			}
		case "page":
			page := Page{}
			for k, v := range attribs {
				switch k {
				case "id":
					page.Id = v
				case "file":
					page.File = strs[v]
				}
			}
			fnt.Pages = append(fnt.Pages, page)
		case "kerning":
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
			fnt.Kernings = append(fnt.Kernings, kern)
		}
	}
	return fnt, sc.Err()
}

func parseTag(line string) (name string, values map[string]int, strs []string, err error) {
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
