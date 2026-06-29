package websocket

import (
	"bufio"
	"encoding/binary"
	"io"
)

func Write(w io.Writer, payload []byte) error {
	header := []byte{0x81}
	size := len(payload)

	switch {
	case size < 126:
		header = append(header, byte(size))
	case size <= 65535:
		header = append(header, 126, byte(size>>8), byte(size))
	default:
		header = append(header, 127)
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(size))
		header = append(header, length...)
	}

	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func Read(r *bufio.Reader) error {
	for {
		header, err := r.ReadByte()
		if err != nil {
			return err
		}
		opcode := header & 0x0f

		size, err := size(r)
		if err != nil {
			return err
		}
		if size > 0 {
			if _, err := io.CopyN(io.Discard, r, size); err != nil {
				return err
			}
		}
		if opcode == 0x8 {
			return nil
		}
	}
}

func size(r *bufio.Reader) (int64, error) {
	sizeByte, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	masked := sizeByte&0x80 != 0
	size := int64(sizeByte & 0x7f)

	switch size {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(r, ext[:]); err != nil {
			return 0, err
		}
		size = int64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(r, ext[:]); err != nil {
			return 0, err
		}
		size = int64(binary.BigEndian.Uint64(ext[:]))
	}

	if masked {
		var mask [4]byte
		if _, err := io.ReadFull(r, mask[:]); err != nil {
			return 0, err
		}
	}

	return size, nil
}
