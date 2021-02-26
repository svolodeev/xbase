package xbase

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/encoding"
)

const (
	maxFieldNameLen = 10
	maxCFieldLen    = 254
	maxNFieldLen    = 19
)

const (
	defaultLFieldLen = 1
	defaultDFieldLen = 8
)

type field struct {
	Name   [11]byte
	Type   byte
	Offset uint32
	Len    byte
	Dec    byte
	Filler [14]byte
}

func (f *field) name() string {
	i := bytes.IndexByte(f.Name[:], 0)
	return string(f.Name[:i])
}

// String utils

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// New field

func newField(name string, typ string, length, dec int) *field {
	f := &field{}
	// do not change the call order
	f.setName(name)
	f.setType(typ)
	f.setLen(length)
	f.setDec(dec)
	return f
}

func (f *field) setName(name string) {
	name = strings.ToUpper(strings.TrimSpace(name))
	if len(name) == 0 {
		panic(fmt.Errorf("empty field name"))
	}
	if len(name) > maxFieldNameLen {
		panic(fmt.Errorf("too long field name: %q, want len(name) <= %d", name, maxFieldNameLen))
	}
	copy(f.Name[:], name)
}

func (f *field) setType(typ string) {
	typ = strings.ToUpper(strings.TrimSpace(typ))
	if len(typ) == 0 {
		panic(fmt.Errorf("empty field type"))
	}
	t := typ[0]
	if bytes.IndexByte([]byte("CNLD"), t) < 0 {
		panic(fmt.Errorf("invalid field type: got %s, want C, N, L, D", string(t)))
	}
	f.Type = t
}

func (f *field) setLen(length int) {
	switch f.Type {
	case 'C':
		if length <= 0 || length > maxCFieldLen {
			panic(fmt.Errorf("invalid field len: got %d, want 0 < len <= %d", length, maxCFieldLen))
		}
	case 'N':
		if length <= 0 || length > maxNFieldLen {
			panic(fmt.Errorf("invalid field len: got %d, want 0 < len <= %d", length, maxNFieldLen))
		}
	case 'L':
		length = defaultLFieldLen
	case 'D':
		length = defaultDFieldLen
	}
	f.Len = byte(length)
}

func (f *field) setDec(dec int) {
	if f.Type == 'N' {
		if dec < 0 {
			panic(fmt.Errorf("invalid field dec: got %d, want dec > 0", dec))
		}
		length := int(f.Len)
		if length <= 2 && dec > 0 {
			panic(fmt.Errorf("invalid field dec: got %d, want 0", dec))
		}
		if length > 2 && (dec > length-2) {
			panic(fmt.Errorf("invalid field dec: got %d, want dec <= %d", dec, length-2))
		}
	} else {
		dec = 0
	}
	f.Dec = byte(dec)
}

// Read/write

func (f *field) read(reader io.Reader) {
	if err := binary.Read(reader, binary.LittleEndian, f); err != nil {
		panic(err)
	}
}

func (f *field) write(writer io.Writer) {
	tmp := f.Offset
	f.Offset = 0
	defer func() { f.Offset = tmp }()

	if err := binary.Write(writer, binary.LittleEndian, f); err != nil {
		panic(err)
	}
}

// Buffer

func (f *field) buffer(recordBuf []byte) []byte {
	return recordBuf[int(f.Offset) : int(f.Offset)+int(f.Len)]
}

func (f *field) setBuffer(recordBuf []byte, value string) {
	copy(recordBuf[int(f.Offset):int(f.Offset)+int(f.Len)], value)
}

// Check

func (f *field) checkType(t byte) {
	if f.Type != t {
		panic(fmt.Errorf("type mismatch"))
	}
}

func (f *field) checkLen(value string) {
	if len(value) > int(f.Len) {
		panic(fmt.Errorf("field overflow"))
	}
}

// Get value

func (f *field) stringValue(recordBuf []byte, dec *encoding.Decoder) string {
	s := string(f.buffer(recordBuf))

	switch f.Type {
	case 'C':
		s = strings.TrimRight(s, " ")
	case 'N':
		s = strings.TrimLeft(s, " ")
	}

	if dec != nil && f.Type == 'C' && !isASCII(s) {
		ds, err := dec.String(s)
		if err != nil {
			panic(err)
		}
		s = ds
	}
	return s
}

func (f *field) boolValue(recordBuf []byte) bool {
	f.checkType('L')
	fieldBuf := f.buffer(recordBuf)
	b := fieldBuf[0]
	return (b == 'T' || b == 't' || b == 'Y' || b == 'y')
}

func (f *field) dateValue(recordBuf []byte) time.Time {
	f.checkType('D')
	s := string(f.buffer(recordBuf))
	var d time.Time
	if strings.Trim(s, " ") == "" {
		return d
	}
	d, err := time.Parse("20060102", s)
	if err != nil {
		panic(err)
	}
	return d
}

func (f *field) intValue(recordBuf []byte) int64 {
	f.checkType('N')
	s := string(f.buffer(recordBuf))
	s = strings.TrimSpace(s)
	if s == "" || s[0] == '.' {
		return 0
	}
	i := strings.IndexByte(s, '.')
	if i > 0 {
		s = s[0:i]
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return n
}

func (f *field) floatValue(recordBuf []byte) float64 {
	f.checkType('N')
	s := string(f.buffer(recordBuf))
	s = strings.TrimSpace(s)
	if s == "" || s[0] == '.' {
		return 0
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return n
}

// Set value

func (f *field) setStringValue(recordBuf []byte, value string, enc *encoding.Encoder) {
	f.checkType('C')

	if enc != nil && !isASCII(value) {
		s, err := enc.String(value)
		if err != nil {
			panic(err)
		}
		value = s
	}
	f.checkLen(value)
	f.setBuffer(recordBuf, padRight(value, int(f.Len)))
}

func (f *field) setBoolValue(recordBuf []byte, value bool) {
	f.checkType('L')
	s := "F"
	if value {
		s = "T"
	}
	f.setBuffer(recordBuf, s)
}

func (f *field) setDateValue(recordBuf []byte, value time.Time) {
	f.checkType('D')
	f.setBuffer(recordBuf, value.Format("20060102"))
}

func (f *field) setIntValue(recordBuf []byte, value int64) {
	f.checkType('N')
	s := strconv.FormatInt(value, 10)
	if f.Dec > 0 {
		s += "." + strings.Repeat("0", int(f.Dec))
	}
	f.checkLen(s)
	f.setBuffer(recordBuf, padLeft(s, int(f.Len)))
}

func (f *field) setFloatValue(recordBuf []byte, value float64) {
	f.checkType('N')
	s := strconv.FormatFloat(value, 'f', int(f.Dec), 64)
	f.checkLen(s)
	f.setBuffer(recordBuf, padLeft(s, int(f.Len)))
}

func (f *field) setValue(recordBuf []byte, value interface{}, enc *encoding.Encoder) {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("field %q: %w", f.name(), r))
		}
	}()
	switch v := value.(type) {
	case string:
		f.setStringValue(recordBuf, v, enc)
	case bool:
		f.setBoolValue(recordBuf, v)
	case int:
		f.setIntValue(recordBuf, int64(v))
	case int8:
		f.setIntValue(recordBuf, int64(v))
	case int16:
		f.setIntValue(recordBuf, int64(v))
	case int32:
		f.setIntValue(recordBuf, int64(v))
	case int64:
		f.setIntValue(recordBuf, int64(v))
	case uint8:
		f.setIntValue(recordBuf, int64(v))
	case uint16:
		f.setIntValue(recordBuf, int64(v))
	case uint32:
		f.setIntValue(recordBuf, int64(v))
	case uint64:
		f.setIntValue(recordBuf, int64(v))
	case float32:
		f.setFloatValue(recordBuf, float64(v))
	case float64:
		f.setFloatValue(recordBuf, float64(v))
	case time.Time:
		f.setDateValue(recordBuf, v)
	default:
		panic(fmt.Errorf("unsupport type value"))
	}
}
