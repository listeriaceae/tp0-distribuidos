package common

import (
	"encoding/binary"
	"io"
	"net"
	"strconv"
)

func BetMarshal(w io.Writer, agency int, record []string) error {
	data := []string{strconv.Itoa(agency)}

	data = append(data, record...)
	offset := uint16(len(data) * 2)

	// Consecutive offsets indicate the start and end of attributes.
	// There's an implicit initial offset of `header size`.
	for _, v := range data {
		offset += uint16(len(v))
		if err := binary.Write(w, binary.BigEndian, offset); err != nil {
			return err
		}
	}

	// We do not account for alignment, attributes are packed together
	for _, v := range data {
		if _, err := io.WriteString(w, v); err != nil {
			return err
		}
	}

	return nil
}

func sendBatch(conn net.Conn, r io.Reader, length int) ([]byte, error) {
	var result [7]byte

	// The first 2 bytes of the message encode its length in big endian
	err := binary.Write(conn, binary.BigEndian, uint16(length))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(conn, r)
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(conn, result[:])
	return result[:], err
}

func RequestWinners(w io.Writer, agency int) error {
	var buf [4]byte

	// binary.BigEndian.PutUint16(buf[:], uint16(0))
	binary.BigEndian.PutUint16(buf[2:], uint16(agency))
	_, err := w.Write(buf[:])
	return err
}
