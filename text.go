package bmf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// TextParseError contains info about where and why a parsing error occurred
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

// ParseText parses a bmf font file in text format
func ParseText(src io.Reader) (fnt *Font, err error) {
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

	sc := bufio.NewScanner(src)
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
			info.Bold = Bool(v)
		case "italic":
			info.Italic = Bool(v)
		case "charset":
			info.Charset = strs[v]
		case "unicode":
			info.Unicode = Bool(v)
		case "stretchH":
			info.StretchH = v
		case "smooth":
			info.Smooth = Bool(v)
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
			common.Packed = Bool(v)
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

// SerializeText serializes a bmf font file in text format
func SerializeText(fnt *Font, dst io.Writer) error {
	if err := serializeInfoBlockText(fnt, dst); err != nil {
		return err
	}
	if err := serializeCommonBlockText(fnt, dst); err != nil {
		return err
	}
	if err := serializePagesBlockText(fnt, dst); err != nil {
		return err
	}
	if err := serializeCharsBlockText(fnt, dst); err != nil {
		return err
	}
	if err := serializeKerningsBlockText(fnt, dst); err != nil {
		return err
	}
	return nil
}

func serializeInfoBlockText(fnt *Font, dst io.Writer) error {
	i := fnt.Info

	_, err := fmt.Fprintf(dst, "info face=%q size=%d bold=%d italic=%d charset=%q unicode=%d stretchH=%d smooth=%d aa=%d ",
		i.Face, i.Size, i.Bold.Byte(), i.Italic.Byte(), i.Charset, i.Unicode.Byte(), i.StretchH, i.Smooth.Byte(), i.AA)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(dst, "padding=%d,%d,%d,%d spacing=%d,%d outline=%d\n",
		i.Padding.Up, i.Padding.Right, i.Padding.Down, i.Padding.Left, i.Spacing.Horizontal, i.Spacing.Vertical, i.Outline)

	return err
}

func serializeCommonBlockText(fnt *Font, dst io.Writer) error {
	c := fnt.Common

	_, err := fmt.Fprintf(dst, "common lineHeight=%d base=%d scaleW=%d scaleH=%d pages=%d packed=%d ",
		c.LineHeight, c.Base, c.ScaleW, c.ScaleH, c.Pages, c.Packed.Byte())
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(dst, "alphaChnl=%d redChnl=%d greenChnl=%d blueChnl=%d\n",
		c.AlphaChannel, c.RedChannel, c.GreenChannel, c.BlueChannel)

	return err
}

func serializePagesBlockText(fnt *Font, dst io.Writer) error {
	for _, p := range fnt.Pages {
		_, err := fmt.Fprintf(dst, "page id=%d file=%q\n", p.Id, p.File)
		if err != nil {
			return err
		}
	}
	return nil
}

func serializeCharsBlockText(fnt *Font, dst io.Writer) error {
	_, err := fmt.Fprintf(dst, "chars count=%d\n", len(fnt.Chars))
	if err != nil {
		return err
	}

	for _, c := range fnt.Chars {
		_, err = fmt.Fprintf(dst, "char id=%d x=%d y=%d width=%d height=%d ",
			c.Id, c.X, c.Y, c.Width, c.Height)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(dst, "xoffset=%d yoffset=%d xadvance=%d page=%d chnl=%d\n",
			c.XOffset, c.YOffset, c.XAdvance, c.Page, c.Channel)
		if err != nil {
			return err
		}
	}
	return nil
}

func serializeKerningsBlockText(fnt *Font, dst io.Writer) error {
	_, err := fmt.Fprintf(dst, "kernings count=%d\n", len(fnt.Kernings))
	if err != nil {
		return err
	}

	for _, k := range fnt.Kernings {
		_, err := fmt.Fprintf(dst, "kerning first=%d second=%d amount=%d\n",
			k.First, k.Second, k.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}
