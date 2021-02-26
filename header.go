package xbase

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

type header struct {
	DbfId      byte
	ModYear    byte
	ModMonth   byte
	ModDay     byte
	RecCount   uint32
	DataOffset uint16
	RecSize    uint16
	Filler1    [17]byte
	CP         byte
	Filler2    [2]byte
}

func newHeader() *header {
	h := &header{}
	h.DbfId = dbfId
	h.setModDate(time.Now())
	return h
}

// Read/write

func (h *header) read(reader io.Reader) {
	if err := binary.Read(reader, binary.LittleEndian, h); err != nil {
		panic(err)
	}
	if h.DbfId != dbfId {
		panic(fmt.Errorf("not DBF file"))
	}
}

func (h *header) write(writer io.Writer) {
	if err := binary.Write(writer, binary.LittleEndian, h); err != nil {
		panic(err)
	}
}

// Field count

func (h *header) fieldCount() int {
	return (int(h.DataOffset) - headerSize - 1) / fieldSize
}

func (h *header) setFieldCount(count int) {
	h.DataOffset = uint16(count*fieldSize + headerSize + 1)
}

// Modified date

func (h *header) modDate() time.Time {
	year := int(h.ModYear) + 1900
	month := time.Month(h.ModMonth)
	day := int(h.ModDay)
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func (h *header) setModDate(d time.Time) {
	h.ModYear = byte(d.Year() - 1900)
	h.ModMonth = byte(d.Month())
	h.ModDay = byte(d.Day())
}

// Code page

func (h *header) codePage() int {
	return pageByCode(h.CP)
}

func (h *header) setCodePage(cp int) {
	h.CP = codeByPage(cp)
}
