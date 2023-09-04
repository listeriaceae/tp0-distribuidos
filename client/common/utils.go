package common

import (
	"encoding/binary"
	"io"
	"strconv"
)

// A lottery bet registry.
type Bet struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    int
}

func (b *Bet) Marshal(w io.Writer, agency int) error {
	data := [...]string{
		strconv.Itoa(agency),
		b.FirstName,
		b.LastName,
		b.Document,
		b.Birthdate,
		strconv.Itoa(b.Number),
	}

	const header_size = len(data) * 2
	offset := uint16(header_size)

	// Consecutive offsets indicate the start and end of attributes.
	// There's an implicit initial offset of `header_size`.
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
