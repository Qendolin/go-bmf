package bmf

import (
	"encoding/binary"
	"fmt"
	"io"
)

type BinaryParseError struct {
	Offset    int
	Block     []byte
	BlockName string
	Err       error
}

type byteReader struct {
	Buffer []byte
	Data   []byte
	Index  int
	Err    error
}

func (br *byteReader) Read(n int) (ok bool) {
	if br.Err != nil {
		return false
	}
	if len(br.Data) < br.Index+n {
		br.Err = io.EOF
		n = len(br.Data) - br.Index
	}
	if cap(br.Buffer) <= n {
		br.Buffer = make([]byte, n)
	} else {
		br.Buffer = br.Buffer[:n]
	}
	br.Index += copy(br.Buffer, br.Data[br.Index:br.Index+n])
	if br.Err != nil {
		return false
	}
	return true
}

func (br *byteReader) ReadInt16(i *int, bo binary.ByteOrder) (ok bool) {
	if !br.Read(2) {
		return false
	}
	*i = int(int16(bo.Uint16(br.Buffer)))
	return true
}

func (br *byteReader) ReadInt8(i *int, bo binary.ByteOrder) (ok bool) {
	if !br.Read(1) {
		return false
	}
	*i = int(int8(br.Buffer[0]))
	return true
}

func (br *byteReader) ReadBits(i *uint8) (ok bool) {
	if !br.Read(1) {
		return false
	}
	*i = br.Buffer[0]
	return true
}

func (br *byteReader) ReadRune(r *rune, bo binary.ByteOrder) (ok bool) {
	if !br.Read(4) {
		return false
	}
	*r = rune(int32(bo.Uint32(br.Buffer)))
	return true
}

func (e BinaryParseError) Error() string {
	blockStr := ""
	for _, b := range e.Block {
		blockStr += fmt.Sprintf("%02x ", b)
	}
	blockStr = blockStr[:len(blockStr)-1]
	return fmt.Sprintf("format error at %v in %v during %v: %v", e.Offset, e.BlockName, blockStr, e.Err)
}

func ParseBinary(data []byte) (fnt *Font, err error) {
	frd := &byteReader{Data: data}
	rd := frd
	fnt = &Font{}
	blockName := "header"

	defer func() {
		if err != nil {
			err = BinaryParseError{
				Offset:    rd.Index,
				Block:     rd.Buffer,
				BlockName: blockName,
				Err:       err,
			}
		}
	}()

	if !frd.Read(3) {
		return nil, fmt.Errorf("expected three bytes for the file identifier")
	}
	if string(frd.Buffer) != "BMF" {
		return nil, fmt.Errorf("expected 'BMF'")
	}

	if !frd.Read(1) {
		return nil, fmt.Errorf("expected one byte for the format version")
	}
	if frd.Buffer[0] != 3 {
		return nil, fmt.Errorf("expected version to be one 3")
	}

	bin := binary.LittleEndian

	for frd.Read(5) {
		typ := rd.Buffer[0]
		switch typ {
		case 1:
			blockName = "info"
		case 2:
			blockName = "common"
		case 3:
			blockName = "pages"
		case 4:
			blockName = "chars"
		case 5:
			blockName = "kerning pairs"
		default:
			return nil, fmt.Errorf("expected block type to be one of 1,2,3,4,5")
		}

		blockLen := int(bin.Uint32(rd.Buffer[1:]))
		if !rd.Read(blockLen) {
			return nil, fmt.Errorf("expected a %v block with length of %d", blockName, blockLen)
		}
		brd := &byteReader{Data: rd.Buffer}
		rd = brd

		switch blockName {
		case "info":
			if !brd.ReadInt16(&fnt.Info.Size, bin) {
				return nil, fmt.Errorf("expected two bytes for fontSize")
			}
			var flags uint8
			if !brd.ReadBits(&flags) {
				return nil, fmt.Errorf("expected one byte for bitField")
			}
			fnt.Info.Smooth = itob(int(flags >> 7 & 0x1))
			fnt.Info.Unicode = itob(int(flags >> 6 & 0x1))
			fnt.Info.Italic = itob(int(flags >> 5 & 0x1))
			fnt.Info.Bold = itob(int(flags >> 4 & 0x1))
			//FIXME: Unused "fixedHeigth" bit
			if !brd.Read(1) {
				return nil, fmt.Errorf("expected one byte for charSet")
			}

			if !brd.ReadInt16(&fnt.Info.StretchH, bin) {
				return nil, fmt.Errorf("expected two bytes for stretchH")
			}
			if !brd.ReadInt8(&fnt.Info.AA, bin) {
				return nil, fmt.Errorf("expected one byte for aa")
			}
			if !brd.ReadInt8(&fnt.Info.Padding.Up, bin) {
				return nil, fmt.Errorf("expected one byte for paddingUp")
			}
			if !brd.ReadInt8(&fnt.Info.Padding.Right, bin) {
				return nil, fmt.Errorf("expected one byte for paddingRight")
			}
			if !brd.ReadInt8(&fnt.Info.Padding.Down, bin) {
				return nil, fmt.Errorf("expected one byte for paddingDown")
			}
			if !brd.ReadInt8(&fnt.Info.Padding.Left, bin) {
				return nil, fmt.Errorf("expected one byte for paddingLeft")
			}
			if !brd.ReadInt8(&fnt.Info.Spacing.Horizontal, bin) {
				return nil, fmt.Errorf("expected one byte for spacingHoriz")
			}
			if !brd.ReadInt8(&fnt.Info.Spacing.Vertical, bin) {
				return nil, fmt.Errorf("expected one byte for spacingVert")
			}
			if !brd.ReadInt8(&fnt.Info.Outline, bin) {
				return nil, fmt.Errorf("expected one byte for outline")
			}
			if len := blockLen - brd.Index; !brd.Read(len) {
				return nil, fmt.Errorf("expected %d bytes for fontName", len)
			}
			if brd.Buffer[len(brd.Buffer)-1] != 0 {
				return nil, fmt.Errorf("expected fontName to be null terminated")
			}
			fnt.Info.Face = string(brd.Buffer[:len(brd.Buffer)-1])
		case "common":
			if !brd.ReadInt16(&fnt.Common.LineHeight, bin) {
				return nil, fmt.Errorf("expected two bytes for lineHeight")
			}
			if !brd.ReadInt16(&fnt.Common.Base, bin) {
				return nil, fmt.Errorf("expected two bytes for base")
			}
			if !brd.ReadInt16(&fnt.Common.ScaleW, bin) {
				return nil, fmt.Errorf("expected two bytes for scaleW")
			}
			if !brd.ReadInt16(&fnt.Common.ScaleH, bin) {
				return nil, fmt.Errorf("expected two bytes for scaleH")
			}
			if !brd.ReadInt16(&fnt.Common.Pages, bin) {
				return nil, fmt.Errorf("expected two bytes for pages")
			}
			var flags uint8
			if !brd.ReadBits(&flags) {
				return nil, fmt.Errorf("expected one byte for bitField")
			}
			fnt.Common.Packed = itob(int(flags >> 0 & 1))
			if !brd.ReadInt8((*int)(&fnt.Common.AlphaChannel), bin) {
				return nil, fmt.Errorf("expected one byte for alphaChnl")
			}
			if !brd.ReadInt8((*int)(&fnt.Common.RedChannel), bin) {
				return nil, fmt.Errorf("expected one byte for redChnl")
			}
			if !brd.ReadInt8((*int)(&fnt.Common.GreenChannel), bin) {
				return nil, fmt.Errorf("expected one byte for greenChnl")
			}
			if !brd.ReadInt8((*int)(&fnt.Common.BlueChannel), bin) {
				return nil, fmt.Errorf("expected one byte for blueChnl")
			}
		case "pages":
			nameLen := 0
			file0 := ""
			start := brd.Index
			if !brd.Read(blockLen - start - 1) {
				return nil, fmt.Errorf("expected %d bytes for pageNames", blockLen-start-1)
			}
			for i, b := range brd.Buffer {
				nameLen++
				if b == 0 {
					break
				}
				if i == len(brd.Buffer) {
					return nil, fmt.Errorf("expected null terminated pageName")
				}
				file0 += string(b)
			}
			fnt.Pages = append(fnt.Pages, Page{
				Id:   0,
				File: file0,
			})

			brd.Index = start + nameLen
			for brd.Index < blockLen {
				if !brd.Read(nameLen) {
					return nil, fmt.Errorf("expected %d bytes for pageName", nameLen)
				}
				fnt.Pages = append(fnt.Pages, Page{
					Id:   len(fnt.Pages),
					File: string(brd.Buffer[:len(brd.Buffer)-1]),
				})
			}
			if brd.Index != blockLen {
				return nil, fmt.Errorf("pageNames is longer than block size")
			}
		case "chars":
			for brd.Index < blockLen {
				chr := Char{}
				if !brd.ReadRune(&chr.Id, bin) {
					return nil, fmt.Errorf("expected four bytes for id")
				}
				if !brd.ReadInt16(&chr.X, bin) {
					return nil, fmt.Errorf("expected two bytes for x")
				}
				if !brd.ReadInt16(&chr.Y, bin) {
					return nil, fmt.Errorf("expected two bytes for y")
				}
				if !brd.ReadInt16(&chr.Width, bin) {
					return nil, fmt.Errorf("expected two bytes for width")
				}
				if !brd.ReadInt16(&chr.Height, bin) {
					return nil, fmt.Errorf("expected two bytes for height")
				}
				if !brd.ReadInt16(&chr.XOffset, bin) {
					return nil, fmt.Errorf("expected two bytes for xoffset")
				}
				if !brd.ReadInt16(&chr.YOffset, bin) {
					return nil, fmt.Errorf("expected two bytes for yoffset")
				}
				if !brd.ReadInt16(&chr.XAdvance, bin) {
					return nil, fmt.Errorf("expected two bytes for xadvance")
				}
				if !brd.ReadInt8(&chr.Page, bin) {
					return nil, fmt.Errorf("expected one byte for page")
				}
				if !brd.ReadInt8((*int)(&chr.Channel), bin) {
					return nil, fmt.Errorf("expected one byte for chnl")
				}
				fnt.Chars = append(fnt.Chars, chr)
			}
			if brd.Index != blockLen {
				return nil, fmt.Errorf("chars is longer than block size")
			}
		case "kerning pairs":
			for brd.Index < blockLen {
				kern := Kerning{}
				if !brd.ReadRune(&kern.First, bin) {
					return nil, fmt.Errorf("expected four bytes for first")
				}
				if !brd.ReadRune(&kern.Second, bin) {
					return nil, fmt.Errorf("expected four bytes for second")
				}
				if !brd.ReadInt16(&kern.Amount, bin) {
					return nil, fmt.Errorf("expected two bytes for amount")
				}
				fnt.Kernings = append(fnt.Kernings, kern)
			}
			if brd.Index != blockLen {
				return nil, fmt.Errorf("kerning pairs is longer than block size")
			}
		}
		rd = frd
	}

	return fnt, nil
}

// source https://docs.microsoft.com/en-us/previous-versions/windows/desktop/bb322881(v=vs.85)
func lookupCharset(chrset int) string {
	switch chrset {
	case 186:
		return "Baltic"
	case 77:
		return "Mac"
	case 204:
		return "Russian"
	case 238:
		return "EastEurope"
	case 222:
		return "Thai"
	case 163:
		return "Vietnamese"
	case 162:
		return "Turkish"
	case 161:
		return "Greek"
	case 178:
		return "Arabic"
	case 177:
		return "Hebrew"
	case 130:
		return "Johab"
	case 255:
		return "Oem"
	case 136:
		return "ChineseBig5"
	case 134:
		return "GB2312"
	case 129:
		return "Hangul"
	case 128:
		return "ShiftJIS"
	case 2:
		return "Symbol"
	case 1:
		return "Default"
	case 0:
		return "Ansi"
	}
	return fmt.Sprintf("%d", chrset)
}
