package bmf

import (
	encoding "encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Qendolin/go-bmf/internal/binary"
)

// The supportet version of the binary format
const SupportedVersion = 3

// BinaryParseError contains info about where and why a parsing error occurred
type BinaryParseError struct {
	Offset      int
	Block       BlockType
	BlockLength int
	Err         error
}

func (e BinaryParseError) Error() string {
	return fmt.Sprintf("format error at index %v in block %v with expected length %v: %v", e.Offset, e.Block.Name(), e.BlockLength, e.Err)
}

func (e BinaryParseError) Unwrap() error {
	return e.Err
}

type BlockType byte

const (
	blockHeader       BlockType = 0
	blockInfo         BlockType = 1
	blockCommon       BlockType = 2
	blockPages        BlockType = 3
	blockChars        BlockType = 4
	blockKerningPairs BlockType = 5
)

func (typ BlockType) Name() string {
	return blockNameTable[typ]
}

var blockNameTable = map[BlockType]string{
	blockHeader:       "header",
	blockInfo:         "info",
	blockCommon:       "common",
	blockPages:        "pages",
	blockChars:        "characters",
	blockKerningPairs: "kerning pairs",
}

// ParseBinary parses a bmf font definition in binary format.
// For more information see http://www.angelcode.com/products/bmfont/doc/file_format.html#bin
func ParseBinary(src io.Reader) (fnt *Font, err error) {
	fileReader := &binary.Reader{
		Src:   src,
		Order: encoding.LittleEndian,
	}
	fnt = &Font{}

	if err := parseHeaderBinary(fileReader); err != nil {
		return nil, err
	}

	for {
		err = parseBlockBinary(fnt, fileReader)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return fnt, nil
}

func parseBlockBinary(fnt *Font, fileReader *binary.Reader) (err error) {
	var (
		blockId   int
		blockType BlockType
		blockLen  int
	)

	defer func() {
		if err != nil {
			err = BinaryParseError{
				Offset:      fileReader.Index,
				Block:       blockType,
				BlockLength: blockLen,
				Err:         err,
			}
		}
	}()

	if !fileReader.ReadUInt8(&blockId) {
		if errors.Is(fileReader.Err, io.EOF) {
			return fileReader.Err
		}

		return fmt.Errorf("expected one byte for block type identifier")
	}
	if blockId < 1 || blockId > 5 {
		return fmt.Errorf("expected block type to be one of 1,2,3,4,5 but was %d", blockId)
	}

	blockType = BlockType(blockId)
	if !fileReader.ReadUInt32(&blockLen) {
		return fmt.Errorf("expected four bytes for block length")
	}

	blockReader := &binary.Reader{
		Src:   fileReader.Src,
		Order: encoding.LittleEndian,
	}

	switch blockType {
	case blockInfo:
		info, err := parseInfoBinary(blockReader, blockLen)
		fileReader.Index += blockReader.Index
		if err != nil {
			return err
		}
		fnt.Info = *info
	case blockCommon:
		common, err := parseCommonBinary(blockReader, blockLen)
		fileReader.Index += blockReader.Index
		if err != nil {
			return err
		}
		fnt.Common = *common
	case blockPages:
		pages, err := parsePagesBinary(blockReader, blockLen)
		fileReader.Index += blockReader.Index
		if err != nil {
			return err
		}
		fnt.Pages = pages
	case blockChars:
		chars, err := parseCharsBinary(blockReader, blockLen)
		fileReader.Index += blockReader.Index
		if err != nil {
			return err
		}
		fnt.Chars = chars
	case blockKerningPairs:
		kernings, err := parseKerningPairsBinary(blockReader, blockLen)
		fileReader.Index += blockReader.Index
		if err != nil {
			return err
		}
		fnt.Kernings = kernings
	}

	return nil
}

func parseHeaderBinary(frd *binary.Reader) (err error) {
	defer func() {
		if err != nil {
			err = BinaryParseError{
				Offset:      frd.Index,
				Block:       blockHeader,
				BlockLength: 4,
				Err:         err,
			}
		}
	}()

	var start string
	if !frd.ReadString(&start, 3) {
		return fmt.Errorf("expected three bytes for the file identifier")
	}
	if start != "BMF" {
		return fmt.Errorf("expected 'BMF'")
	}

	var version int
	if !frd.ReadUInt8(&version) {
		return fmt.Errorf("expected one byte for the format version")
	}
	if version != SupportedVersion {
		return fmt.Errorf("expected version to be one 3")
	}

	return nil
}

func parseInfoBinary(brd *binary.Reader, blockLength int) (*Info, error) {
	if blockLength < 15 {
		return nil, fmt.Errorf("expected at least 15 bytes for info block but was %d", blockLength)
	}

	info := Info{}

	if !brd.ReadInt16(&info.Size) {
		return nil, fmt.Errorf("expected two bytes for fontSize")
	}
	var flags uint8
	if !brd.ReadBits(&flags) {
		return nil, fmt.Errorf("expected one byte for bitField")
	}
	info.Smooth = Bool(int(flags >> 7 & 0x1))
	info.Unicode = Bool(int(flags >> 6 & 0x1))
	info.Italic = Bool(int(flags >> 5 & 0x1))
	info.Bold = Bool(int(flags >> 4 & 0x1))
	//FIXME: Unused "fixedHeigth" bit

	if charSet := 0; !brd.ReadUInt8(&charSet) {
		return nil, fmt.Errorf("expected one byte for charSet")
	} else if !info.Unicode {
		info.Charset, _ = LookupCharset(charSet)
	}

	if !brd.ReadUInt16(&info.StretchH) {
		return nil, fmt.Errorf("expected two bytes for stretchH")
	}
	if !brd.ReadUInt8(&info.AA) {
		return nil, fmt.Errorf("expected one byte for aa")
	}
	if !brd.ReadUInt8(&info.Padding.Up) {
		return nil, fmt.Errorf("expected one byte for paddingUp")
	}
	if !brd.ReadUInt8(&info.Padding.Right) {
		return nil, fmt.Errorf("expected one byte for paddingRight")
	}
	if !brd.ReadUInt8(&info.Padding.Down) {
		return nil, fmt.Errorf("expected one byte for paddingDown")
	}
	if !brd.ReadUInt8(&info.Padding.Left) {
		return nil, fmt.Errorf("expected one byte for paddingLeft")
	}
	if !brd.ReadUInt8(&info.Spacing.Horizontal) {
		return nil, fmt.Errorf("expected one byte for spacingHoriz")
	}
	if !brd.ReadUInt8(&info.Spacing.Vertical) {
		return nil, fmt.Errorf("expected one byte for spacingVert")
	}
	if !brd.ReadUInt8(&info.Outline) {
		return nil, fmt.Errorf("expected one byte for outline")
	}

	fontNameLen := blockLength - brd.Index
	fontName := ""
	if !brd.ReadString(&fontName, fontNameLen) {
		return nil, fmt.Errorf("expected %d bytes for fontName", fontNameLen)
	} else if fontName[fontNameLen-1] != 0 {
		return nil, fmt.Errorf("expected fontName to be null terminated")
	}
	info.Face = fontName[0 : fontNameLen-1]

	return &info, nil
}

func parseCommonBinary(brd *binary.Reader, blockLength int) (*Common, error) {
	if blockLength != 15 {
		return nil, fmt.Errorf("expected 15 bytes for common block but was %d", blockLength)
	}

	common := Common{}

	if !brd.ReadUInt16(&common.LineHeight) {
		return nil, fmt.Errorf("expected two bytes for lineHeight")
	}
	if !brd.ReadUInt16(&common.Base) {
		return nil, fmt.Errorf("expected two bytes for base")
	}
	if !brd.ReadUInt16(&common.ScaleW) {
		return nil, fmt.Errorf("expected two bytes for scaleW")
	}
	if !brd.ReadUInt16(&common.ScaleH) {
		return nil, fmt.Errorf("expected two bytes for scaleH")
	}
	if !brd.ReadUInt16(&common.Pages) {
		return nil, fmt.Errorf("expected two bytes for pages")
	}
	var flags uint8
	if !brd.ReadBits(&flags) {
		return nil, fmt.Errorf("expected one byte for bitField")
	}
	common.Packed = Bool(int(flags >> 0 & 1))
	if !brd.ReadUInt8((*int)(&common.AlphaChannel)) {
		return nil, fmt.Errorf("expected one byte for alphaChnl")
	}
	if !brd.ReadUInt8((*int)(&common.RedChannel)) {
		return nil, fmt.Errorf("expected one byte for redChnl")
	}
	if !brd.ReadUInt8((*int)(&common.GreenChannel)) {
		return nil, fmt.Errorf("expected one byte for greenChnl")
	}
	if !brd.ReadUInt8((*int)(&common.BlueChannel)) {
		return nil, fmt.Errorf("expected one byte for blueChnl")
	}

	return &common, nil
}

// All spaces at the start or end of a page file name are trimmed
func parsePagesBinary(brd *binary.Reader, blockLength int) ([]Page, error) {
	var file0 string
	if !brd.ReadNullString(&file0, blockLength) {
		return nil, fmt.Errorf("expected first null-terminated pageName")
	}

	nameLen := len(file0) + 1
	pages := []Page{{
		Id:   0,
		File: strings.Trim(file0, " "),
	}}

	if blockLength%nameLen != 0 {
		return nil, fmt.Errorf("expected multiple of %d bytes for pages block but was %d", nameLen, blockLength)
	}

	for i := 1; i < blockLength/nameLen; i++ {
		var name string

		if !brd.ReadString(&name, nameLen) {
			return nil, fmt.Errorf("expected %d bytes for pageName %d", nameLen, i)
		}
		pages = append(pages, Page{
			Id:   i,
			File: strings.Trim(name[:nameLen-1], " "),
		})
	}

	return pages, nil
}

func parseCharsBinary(brd *binary.Reader, blockLength int) (chars []Char, err error) {
	if blockLength%20 != 0 {
		return nil, fmt.Errorf("expected a multiple of 20 bytes for charater block but was %d", blockLength)
	}

	charCount := blockLength / 20
	charIdx := 0
	chars = make([]Char, charCount)

	defer func() {
		if err != nil {
			err = fmt.Errorf("at character %d/%d: %v", charIdx+1, charCount, err)
		}
	}()

	for ; charIdx < charCount; charIdx++ {
		char := Char{}
		if !brd.ReadRune(&char.Id) {
			return nil, fmt.Errorf("expected four bytes for id")
		}
		if !brd.ReadUInt16(&char.X) {
			return nil, fmt.Errorf("expected two bytes for x")
		}
		if !brd.ReadUInt16(&char.Y) {
			return nil, fmt.Errorf("expected two bytes for y")
		}
		if !brd.ReadUInt16(&char.Width) {
			return nil, fmt.Errorf("expected two bytes for width")
		}
		if !brd.ReadUInt16(&char.Height) {
			return nil, fmt.Errorf("expected two bytes for height")
		}
		if !brd.ReadInt16(&char.XOffset) {
			return nil, fmt.Errorf("expected two bytes for xoffset")
		}
		if !brd.ReadInt16(&char.YOffset) {
			return nil, fmt.Errorf("expected two bytes for yoffset")
		}
		if !brd.ReadInt16(&char.XAdvance) {
			return nil, fmt.Errorf("expected two bytes for xadvance")
		}
		if !brd.ReadUInt8(&char.Page) {
			return nil, fmt.Errorf("expected one byte for page")
		}
		if !brd.ReadUInt8((*int)(&char.Channel)) {
			return nil, fmt.Errorf("expected one byte for chnl")
		}
		chars[charIdx] = char
	}

	return chars, nil
}

func parseKerningPairsBinary(brd *binary.Reader, blockLength int) (kernings []Kerning, err error) {
	if blockLength%10 != 0 {
		return nil, fmt.Errorf("expected a multiple of 10 bytes for kerning pairs block but was %d", blockLength)
	}

	kernCount := blockLength / 10
	kernIdx := 0
	kernings = make([]Kerning, kernCount)

	defer func() {
		if err != nil {
			err = fmt.Errorf("at kerning pair %d/%d: %v", kernIdx+1, kernCount, err)
		}
	}()

	for ; kernIdx < kernCount; kernIdx++ {
		kern := Kerning{}
		if !brd.ReadRune(&kern.First) {
			return nil, fmt.Errorf("expected four bytes for first")
		}
		if !brd.ReadRune(&kern.Second) {
			return nil, fmt.Errorf("expected four bytes for second")
		}
		if !brd.ReadInt16(&kern.Amount) {
			return nil, fmt.Errorf("expected two bytes for amount")
		}
		kernings[kernIdx] = kern
	}

	return kernings, nil
}

// SerializeBinary serializes a bmf font definition in binary format.
func SerializeBinary(fnt *Font, dst io.Writer) error {
	bw := &binary.Writer{
		Order: encoding.LittleEndian,
		Dst:   dst,
	}
	bw.WriteString("BMF")
	bw.WriteUInt8(SupportedVersion)

	serializeInfoBlockBinary(fnt, bw)
	if bw.Err != nil {
		return bw.Err
	}
	serializeCommonBlockBinary(fnt, bw)
	if bw.Err != nil {
		return bw.Err
	}
	serializePagesBlockBinary(fnt, bw)
	if bw.Err != nil {
		return bw.Err
	}
	serializeCharsBlockBinary(fnt, bw)
	if bw.Err != nil {
		return bw.Err
	}
	serializeKerningsBlockBinary(fnt, bw)
	if bw.Err != nil {
		return bw.Err
	}

	return nil
}

func serializeInfoBlockBinary(fnt *Font, bw *binary.Writer) {
	i := fnt.Info
	bw.WriteUInt8(uint8(blockInfo))
	bw.WriteInt32(14 + int32(len(i.Face)) + 1)

	bw.WriteInt16(int16(i.Size))

	var flags uint8
	flags |= i.Smooth.Byte() << 7
	flags |= i.Unicode.Byte() << 6
	flags |= i.Italic.Byte() << 5
	flags |= i.Bold.Byte() << 4
	bw.WriteBits(flags)

	if charSet, found := LookupCharsetValue(i.Charset); found {
		bw.WriteUInt8(uint8(charSet))
	} else {
		bw.WriteUInt8(0)
	}

	bw.WriteUInt16(uint16(i.StretchH))
	bw.WriteUInt8(uint8(i.AA))

	bw.WriteUInt8(uint8(i.Padding.Up))
	bw.WriteUInt8(uint8(i.Padding.Right))
	bw.WriteUInt8(uint8(i.Padding.Down))
	bw.WriteUInt8(uint8(i.Padding.Left))

	bw.WriteUInt8(uint8(i.Spacing.Horizontal))
	bw.WriteUInt8(uint8(i.Spacing.Vertical))

	bw.WriteUInt8(uint8(i.Outline))

	bw.WriteNullString(i.Face)
}

func serializeCommonBlockBinary(fnt *Font, bw *binary.Writer) {
	c := fnt.Common
	bw.WriteUInt8(uint8(blockCommon))
	bw.WriteInt32(15)

	bw.WriteInt16(int16(c.LineHeight))
	bw.WriteInt16(int16(c.Base))

	bw.WriteInt16(int16(c.ScaleW))
	bw.WriteInt16(int16(c.ScaleH))

	bw.WriteInt16(int16(c.Pages))

	var flags uint8
	flags |= c.Packed.Byte()
	bw.WriteBits(flags)

	bw.WriteUInt8(uint8(c.AlphaChannel))
	bw.WriteUInt8(uint8(c.RedChannel))
	bw.WriteUInt8(uint8(c.GreenChannel))
	bw.WriteUInt8(uint8(c.BlueChannel))
}

// If a page name is shorter than the rest it is padded with spaces at the end
func serializePagesBlockBinary(fnt *Font, bw *binary.Writer) {
	var nameLen int32

	for _, p := range fnt.Pages {
		if len(p.File) > int(nameLen) {
			nameLen = int32(len(p.File))
		}
	}

	bw.WriteUInt8(uint8(blockPages))
	bw.WriteInt32((nameLen + 1) * int32(len(fnt.Pages)))

	for _, p := range fnt.Pages {
		bw.WriteNullString(fmt.Sprintf("%-*s", nameLen, p.File))
	}
}

func serializeCharsBlockBinary(fnt *Font, bw *binary.Writer) {
	bw.WriteUInt8(uint8(blockChars))
	bw.WriteInt32(20 * int32(len(fnt.Chars)))

	for _, c := range fnt.Chars {
		bw.WriteUInt32(uint32(c.Id))

		bw.WriteUInt16(uint16(c.X))
		bw.WriteUInt16(uint16(c.Y))

		bw.WriteUInt16(uint16(c.Width))
		bw.WriteUInt16(uint16(c.Height))

		bw.WriteInt16(int16(c.XOffset))
		bw.WriteInt16(int16(c.YOffset))

		bw.WriteInt16(int16(c.XAdvance))

		bw.WriteUInt8(uint8(c.Page))
		bw.WriteUInt8(uint8(c.Channel))
	}
}

func serializeKerningsBlockBinary(fnt *Font, bw *binary.Writer) {
	bw.WriteUInt8(uint8(blockKerningPairs))
	bw.WriteInt32(10 * int32(len(fnt.Kernings)))

	for _, k := range fnt.Kernings {
		bw.WriteUInt32(uint32(k.First))
		bw.WriteUInt32(uint32(k.Second))
		bw.WriteInt16(int16(k.Amount))
	}
}
