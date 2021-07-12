package binary

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Reader struct {
	Order binary.ByteOrder
	Src   io.Reader
	Index int
	Err   error
	buf   []byte
}

func (br *Reader) Read(n int) (ok bool) {
	if br.Err != nil {
		return false
	}

	if cap(br.buf) <= n {
		br.buf = make([]byte, n)
	} else {
		br.buf = br.buf[:n]
	}

	nread, err := br.Src.Read(br.buf)
	if err != nil {
		br.Err = err
		ok = false
	}

	br.Index += nread

	return br.Err == nil
}

func (br *Reader) ReadUInt8(i *int) (ok bool) {
	if !br.Read(1) {
		return false
	}
	*i = int(br.buf[0])
	return true
}

func (br *Reader) ReadUInt16(i *int) (ok bool) {
	if !br.Read(2) {
		return false
	}
	*i = int(br.Order.Uint16(br.buf))
	return true
}

func (br *Reader) ReadInt16(i *int) (ok bool) {
	if !br.Read(2) {
		return false
	}
	*i = int(int16(br.Order.Uint16(br.buf)))
	return true
}

func (br *Reader) ReadUInt32(i *int) (ok bool) {
	if !br.Read(4) {
		return false
	}
	*i = int(br.Order.Uint32(br.buf))
	return true
}

func (br *Reader) ReadBits(i *uint8) (ok bool) {
	if !br.Read(1) {
		return false
	}
	*i = br.buf[0]
	return true
}

func (br *Reader) ReadRune(r *rune) (ok bool) {
	if !br.Read(4) {
		return false
	}
	*r = rune(int32(br.Order.Uint32(br.buf)))
	return true
}

func (br *Reader) ReadNullString(s *string, max int) (ok bool) {
	buf := make([]byte, max)
	i := 0

	for ; i <= max; i++ {
		if i == max {
			br.Err = fmt.Errorf("string was not null terminated")
			return false
		}

		if !br.Read(1) {
			return false
		}

		if br.buf[0] == 0 {
			break
		}

		buf[i] = br.buf[0]
	}

	*s = string(buf[0:i])

	return true
}

func (br *Reader) ReadString(s *string, c int) (ok bool) {
	if !br.Read(c) {
		return false
	}
	*s = string(br.buf)
	return true
}

type Writer struct {
	Order binary.ByteOrder
	Dst   io.Writer
	Err   error
}

func (bw *Writer) Write(p []byte) (ok bool) {
	if bw.Err != nil {
		return false
	}

	_, err := bw.Dst.Write(p)
	if err != nil {
		bw.Err = err
		ok = false
	}
	return
}

func (bw *Writer) WriteString(s string) (ok bool) {
	return bw.Write([]byte(s))
}

func (bw *Writer) WriteNullString(s string) (ok bool) {
	return bw.Write(append([]byte(s), 0))
}

func (bw *Writer) WriteBits(b uint8) (ok bool) {
	return bw.Write([]byte{b})
}

func (bw *Writer) WriteUInt8(i uint8) (ok bool) {
	return bw.Write([]byte{i})
}

func (bw *Writer) WriteInt8(i int8) (ok bool) {
	return bw.Write([]byte{byte(i)})
}

func (bw *Writer) WriteInt16(i int16) (ok bool) {
	buf := make([]byte, 2)
	bw.Order.PutUint16(buf, uint16(i))
	return bw.Write(buf)
}

func (bw *Writer) WriteUInt16(i uint16) (ok bool) {
	buf := make([]byte, 2)
	bw.Order.PutUint16(buf, i)
	return bw.Write(buf)
}

func (bw *Writer) WriteInt32(i int32) (ok bool) {
	buf := make([]byte, 4)
	bw.Order.PutUint32(buf, uint32(i))
	return bw.Write(buf)
}

func (bw *Writer) WriteUInt32(i uint32) (ok bool) {
	buf := make([]byte, 4)
	bw.Order.PutUint32(buf, i)
	return bw.Write(buf)
}
