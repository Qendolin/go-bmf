package bmf

import (
	"encoding/binary"
	"fmt"
	"io"
)

type binaryBlockLabel string

const (
	blockHeader       binaryBlockLabel = "header"
	blockInfo         binaryBlockLabel = "info"
	blockCommon       binaryBlockLabel = "common"
	blockPages        binaryBlockLabel = "pages"
	blockChars        binaryBlockLabel = "chars"
	blockKerningPairs binaryBlockLabel = "kerning pairs"
)

var binaryBlockTable = map[byte]binaryBlockLabel{
	1: blockInfo,
	2: blockCommon,
	3: blockPages,
	4: blockChars,
	5: blockKerningPairs,
}

// BinaryParseError contains info about where and why a parsing error occured
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

func (br *byteReader) read(n int) (ok bool) {
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

	return br.Err == nil
}

func (br *byteReader) readInt16(i *int, bo binary.ByteOrder) (ok bool) {
	if !br.read(2) {
		return false
	}
	*i = int(int16(bo.Uint16(br.Buffer)))
	return true
}

func (br *byteReader) readInt8(i *int, bo binary.ByteOrder) (ok bool) {
	if !br.read(1) {
		return false
	}
	*i = int(int8(br.Buffer[0]))
	return true
}

func (br *byteReader) readBits(i *uint8) (ok bool) {
	if !br.read(1) {
		return false
	}
	*i = br.Buffer[0]
	return true
}

func (br *byteReader) readRune(r *rune, bo binary.ByteOrder) (ok bool) {
	if !br.read(4) {
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

// ParseBinary parses a bmf font definition in binary format.
// For more information see http://www.angelcode.com/products/bmfont/doc/file_format.html#bin
func ParseBinary(data []byte) (fnt *Font, err error) {
	fileReader := &byteReader{Data: data}
	fnt = &Font{}

	if err := parseHeaderBinary(fileReader); err != nil {
		return nil, err
	}

	for fileReader.read(5) {
		err = parseBlockBinary(fnt, fileReader)
		if err != nil {
			return nil, err
		}
	}

	return fnt, nil
}

func parseBlockBinary(fnt *Font, fileReader *byteReader) (err error) {
	var blockLabel binaryBlockLabel
	order := binary.LittleEndian

	defer func() {
		if err != nil {
			err = BinaryParseError{
				Offset:    fileReader.Index,
				Block:     fileReader.Buffer,
				BlockName: string(blockHeader),
				Err:       err,
			}
		}
	}()

	typ := fileReader.Buffer[0]
	if label, ok := binaryBlockTable[typ]; !ok {
		return fmt.Errorf("expected block type to be one of 1,2,3,4,5 but was %d", typ)
	} else {
		blockLabel = label
	}

	blockLen := int(order.Uint32(fileReader.Buffer[1:]))
	if !fileReader.read(blockLen) {
		return fmt.Errorf("expected a %v block with length of %d", blockLabel, blockLen)
	}
	blockReader := &byteReader{Data: fileReader.Buffer}

	switch blockLabel {
	case "info":
		info, err := parseInfoBinary(blockReader, order, blockLen)
		if err != nil {
			return err
		}
		fnt.Info = *info
	case "common":
		common, err := parseCommonBinary(blockReader, order)
		if err != nil {
			return err
		}
		fnt.Common = *common
	case "pages":
		pages, err := parsePagesBinary(blockReader, blockLen)
		if err != nil {
			return err
		}
		fnt.Pages = pages
	case "chars":
		chars, err := parseCharsBinary(blockReader, order, blockLen)
		if err != nil {
			return err
		}
		fnt.Chars = chars
	case "kerning pairs":
		kernings, err := parseKerningPairsBinary(blockReader, order, blockLen)
		if err != nil {
			return err
		}
		fnt.Kernings = kernings
	}

	return nil
}

func parseHeaderBinary(frd *byteReader) (err error) {
	defer func() {
		if err != nil {
			err = BinaryParseError{
				Offset:    frd.Index,
				Block:     frd.Buffer,
				BlockName: string(blockHeader),
				Err:       err,
			}
		}
	}()

	if !frd.read(3) {
		return fmt.Errorf("expected three bytes for the file identifier")
	}
	if string(frd.Buffer) != "BMF" {
		return fmt.Errorf("expected 'BMF'")
	}

	if !frd.read(1) {
		return fmt.Errorf("expected one byte for the format version")
	}
	if frd.Buffer[0] != 3 {
		return fmt.Errorf("expected version to be one 3")
	}

	return nil
}

func parseInfoBinary(brd *byteReader, order binary.ByteOrder, blockLength int) (*Info, error) {
	info := Info{}

	if !brd.readInt16(&info.Size, order) {
		return nil, fmt.Errorf("expected two bytes for fontSize")
	}
	var flags uint8
	if !brd.readBits(&flags) {
		return nil, fmt.Errorf("expected one byte for bitField")
	}
	info.Smooth = itob(int(flags >> 7 & 0x1))
	info.Unicode = itob(int(flags >> 6 & 0x1))
	info.Italic = itob(int(flags >> 5 & 0x1))
	info.Bold = itob(int(flags >> 4 & 0x1))
	//FIXME: Unused "fixedHeigth" bit
	if !brd.read(1) {
		return nil, fmt.Errorf("expected one byte for charSet")
	}

	if !brd.readInt16(&info.StretchH, order) {
		return nil, fmt.Errorf("expected two bytes for stretchH")
	}
	if !brd.readInt8(&info.AA, order) {
		return nil, fmt.Errorf("expected one byte for aa")
	}
	if !brd.readInt8(&info.Padding.Up, order) {
		return nil, fmt.Errorf("expected one byte for paddingUp")
	}
	if !brd.readInt8(&info.Padding.Right, order) {
		return nil, fmt.Errorf("expected one byte for paddingRight")
	}
	if !brd.readInt8(&info.Padding.Down, order) {
		return nil, fmt.Errorf("expected one byte for paddingDown")
	}
	if !brd.readInt8(&info.Padding.Left, order) {
		return nil, fmt.Errorf("expected one byte for paddingLeft")
	}
	if !brd.readInt8(&info.Spacing.Horizontal, order) {
		return nil, fmt.Errorf("expected one byte for spacingHoriz")
	}
	if !brd.readInt8(&info.Spacing.Vertical, order) {
		return nil, fmt.Errorf("expected one byte for spacingVert")
	}
	if !brd.readInt8(&info.Outline, order) {
		return nil, fmt.Errorf("expected one byte for outline")
	}
	if len := blockLength - brd.Index; !brd.read(len) {
		return nil, fmt.Errorf("expected %d bytes for fontName", len)
	}
	if brd.Buffer[len(brd.Buffer)-1] != 0 {
		return nil, fmt.Errorf("expected fontName to be null terminated")
	}
	info.Face = string(brd.Buffer[:len(brd.Buffer)-1])
	return &info, nil
}

func parseCommonBinary(brd *byteReader, order binary.ByteOrder) (*Common, error) {
	common := Common{}

	if !brd.readInt16(&common.LineHeight, order) {
		return nil, fmt.Errorf("expected two bytes for lineHeight")
	}
	if !brd.readInt16(&common.Base, order) {
		return nil, fmt.Errorf("expected two bytes for base")
	}
	if !brd.readInt16(&common.ScaleW, order) {
		return nil, fmt.Errorf("expected two bytes for scaleW")
	}
	if !brd.readInt16(&common.ScaleH, order) {
		return nil, fmt.Errorf("expected two bytes for scaleH")
	}
	if !brd.readInt16(&common.Pages, order) {
		return nil, fmt.Errorf("expected two bytes for pages")
	}
	var flags uint8
	if !brd.readBits(&flags) {
		return nil, fmt.Errorf("expected one byte for bitField")
	}
	common.Packed = itob(int(flags >> 0 & 1))
	if !brd.readInt8((*int)(&common.AlphaChannel), order) {
		return nil, fmt.Errorf("expected one byte for alphaChnl")
	}
	if !brd.readInt8((*int)(&common.RedChannel), order) {
		return nil, fmt.Errorf("expected one byte for redChnl")
	}
	if !brd.readInt8((*int)(&common.GreenChannel), order) {
		return nil, fmt.Errorf("expected one byte for greenChnl")
	}
	if !brd.readInt8((*int)(&common.BlueChannel), order) {
		return nil, fmt.Errorf("expected one byte for blueChnl")
	}

	return &common, nil
}

func parsePagesBinary(brd *byteReader, blockLength int) ([]Page, error) {
	pages := []Page{}
	nameLen := 0
	file0 := ""
	start := brd.Index
	if !brd.read(blockLength - start - 1) {
		return nil, fmt.Errorf("expected %d bytes for pageNames", blockLength-start-1)
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
	pages = append(pages, Page{
		Id:   0,
		File: file0,
	})

	brd.Index = start + nameLen
	for brd.Index < blockLength {
		if !brd.read(nameLen) {
			return nil, fmt.Errorf("expected %d bytes for pageName", nameLen)
		}
		pages = append(pages, Page{
			Id:   len(pages),
			File: string(brd.Buffer[:len(brd.Buffer)-1]),
		})
	}
	if brd.Index != blockLength {
		return nil, fmt.Errorf("pageNames is longer than block size")
	}

	return pages, nil
}

func parseCharsBinary(brd *byteReader, order binary.ByteOrder, blockLength int) ([]Char, error) {
	chars := make([]Char, 0, blockLength/20)

	for brd.Index < blockLength {
		chr := Char{}
		if !brd.readRune(&chr.Id, order) {
			return nil, fmt.Errorf("expected four bytes for id")
		}
		if !brd.readInt16(&chr.X, order) {
			return nil, fmt.Errorf("expected two bytes for x")
		}
		if !brd.readInt16(&chr.Y, order) {
			return nil, fmt.Errorf("expected two bytes for y")
		}
		if !brd.readInt16(&chr.Width, order) {
			return nil, fmt.Errorf("expected two bytes for width")
		}
		if !brd.readInt16(&chr.Height, order) {
			return nil, fmt.Errorf("expected two bytes for height")
		}
		if !brd.readInt16(&chr.XOffset, order) {
			return nil, fmt.Errorf("expected two bytes for xoffset")
		}
		if !brd.readInt16(&chr.YOffset, order) {
			return nil, fmt.Errorf("expected two bytes for yoffset")
		}
		if !brd.readInt16(&chr.XAdvance, order) {
			return nil, fmt.Errorf("expected two bytes for xadvance")
		}
		if !brd.readInt8(&chr.Page, order) {
			return nil, fmt.Errorf("expected one byte for page")
		}
		if !brd.readInt8((*int)(&chr.Channel), order) {
			return nil, fmt.Errorf("expected one byte for chnl")
		}
		chars = append(chars, chr)
	}

	if brd.Index != blockLength {
		return nil, fmt.Errorf("chars is longer than block size")
	}

	return chars, nil
}

func parseKerningPairsBinary(brd *byteReader, order binary.ByteOrder, blockLength int) ([]Kerning, error) {
	kernings := []Kerning{}

	for brd.Index < blockLength {
		kern := Kerning{}
		if !brd.readRune(&kern.First, order) {
			return nil, fmt.Errorf("expected four bytes for first")
		}
		if !brd.readRune(&kern.Second, order) {
			return nil, fmt.Errorf("expected four bytes for second")
		}
		if !brd.readInt16(&kern.Amount, order) {
			return nil, fmt.Errorf("expected two bytes for amount")
		}
		kernings = append(kernings, kern)
	}

	if brd.Index != blockLength {
		return nil, fmt.Errorf("kerning pairs is longer than block size")
	}

	return kernings, nil
}
