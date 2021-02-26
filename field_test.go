package xbase

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFieldName(t *testing.T) {
	f := &field{
		Name: [11]byte{'N', 'A', 'M', 'E', 0, 0, 0, 0, 0, 0},
	}
	require.Equal(t, "NAME", f.name())
}

func TestFieldSetName(t *testing.T) {
	f := &field{}
	f.setName("name")
	require.Equal(t, "NAME", f.name())
}

func TestFieldSetType(t *testing.T) {
	f := &field{}
	f.setType("numeric")
	require.Equal(t, byte('N'), f.Type)
}

func TestFieldSetLen(t *testing.T) {
	f := &field{}
	f.setType("L")
	f.setLen(0)
	require.Equal(t, byte(1), f.Len)
}

func TestFieldSetDec(t *testing.T) {
	f := &field{}
	f.setType("N")
	f.setLen(5)
	f.setDec(2)
	require.Equal(t, byte(2), f.Dec)
}

func TestNewField(t *testing.T) {
	f := newField("Price", "N", 12, 2)
	require.Equal(t, "PRICE", f.name())
	require.Equal(t, byte('N'), f.Type)
	require.Equal(t, byte(12), f.Len)
	require.Equal(t, byte(2), f.Dec)
}

func TestReadField(t *testing.T) {
	b := make([]byte, fieldSize)
	copy(b[:], "NAME")
	b[11] = 'C'
	b[12] = 1
	b[16] = 14
	r := bytes.NewReader(b)

	f := &field{}
	f.read(r)

	require.Equal(t, "NAME", f.name())
	require.Equal(t, byte('C'), f.Type)
	require.Equal(t, uint32(1), f.Offset)
	require.Equal(t, byte(14), f.Len)
	require.Equal(t, byte(0), f.Dec)
}

func TestWriteField(t *testing.T) {
	f := &field{}
	copy(f.Name[:], "NAME")
	f.Type = 'C'
	f.Offset = 1
	f.Len = 14
	f.Dec = 0

	buf := bytes.NewBuffer(nil)
	f.write(buf)

	b := buf.Bytes()
	require.Equal(t, byte('N'), b[0])
	require.Equal(t, byte('A'), b[1])
	require.Equal(t, byte('M'), b[2])
	require.Equal(t, byte('E'), b[3])
	require.Equal(t, byte('C'), b[11])
	require.Equal(t, byte(0), b[12])
	require.Equal(t, byte(14), b[16])
}

func TestIsASCII(t *testing.T) {
	require.Equal(t, true, isASCII("ab12"))
	require.Equal(t, false, isASCII("шар"))
}

func TestFieldBuffer(t *testing.T) {
	f := newField("Log", "L", 1, 0)
	f.Offset = 6
	recordBuf := []byte(" Abc  T 12")
	buf := f.buffer(recordBuf)
	require.Equal(t, []byte("T"), buf)
}

func TestFieldStringValue(t *testing.T) {
	f := newField("Name", "C", 5, 0)
	f.Offset = 3
	recordBuf := []byte("   Abc    ")
	v := f.stringValue(recordBuf, nil)
	require.Equal(t, "Abc", v)
}

func TestFieldBoolValue(t *testing.T) {
	f := newField("Name", "L", 1, 0)
	f.Offset = 3
	recordBuf := []byte("   T    ")
	v := f.boolValue(recordBuf)
	require.Equal(t, true, v)
}

func TestFieldDateValue(t *testing.T) {
	f := newField("Name", "D", 8, 0)
	f.Offset = 3
	recordBuf := []byte("   20200923    ")

	d := time.Date(2020, 9, 23, 0, 0, 0, 0, time.UTC)
	v := f.dateValue(recordBuf)
	require.Equal(t, d, v)
}

func TestFieldIntValue(t *testing.T) {
	f := newField("Name", "N", 8, 0)
	f.Offset = 3
	recordBuf := []byte("      -2020    ")
	v := f.intValue(recordBuf)
	require.Equal(t, int64(-2020), v)
}

func TestFieldFloatValue(t *testing.T) {
	f := newField("Name", "N", 8, 2)
	f.Offset = 3
	recordBuf := []byte("     -20.21    ")
	v := f.floatValue(recordBuf)
	require.Equal(t, float64(-20.21), v)
}

func TestFieldSetBuffer(t *testing.T) {
	f := newField("Log", "L", 1, 0)
	f.Offset = 6
	recordBuf := []byte(" Abc  T 12")
	f.setBuffer(recordBuf, "F")
	require.Equal(t, []byte(" Abc  F 12"), recordBuf)
}

func TestFieldSetStringValue(t *testing.T) {
	recordBuf := make([]byte, 20)
	f := newField("NAME", "C", 5, 0)
	f.Offset = 5
	f.setStringValue(recordBuf, " Abc", nil)
	require.Equal(t, []byte(" Abc "), recordBuf[5:10])
}

func TestFieldSetBoolValue(t *testing.T) {
	recordBuf := make([]byte, 20)
	f := newField("NAME", "L", 1, 0)
	f.Offset = 5
	f.setBoolValue(recordBuf, true)
	require.Equal(t, []byte("T"), recordBuf[5:6])
}

func TestFieldSetDateValue(t *testing.T) {
	recordBuf := make([]byte, 20)
	f := newField("NAME", "D", 8, 0)
	f.Offset = 5
	d := time.Date(2020, 9, 23, 0, 0, 0, 0, time.UTC)
	f.setDateValue(recordBuf, d)
	require.Equal(t, []byte("20200923"), recordBuf[5:13])
}

func TestFieldSetIntValue(t *testing.T) {
	recordBuf := make([]byte, 20)
	f := newField("NAME", "N", 5, 0)
	f.Offset = 5
	f.setIntValue(recordBuf, 123)
	require.Equal(t, []byte("  123"), recordBuf[5:10])
}

func TestFieldSetFloatValue(t *testing.T) {
	recordBuf := make([]byte, 20)
	f := newField("NAME", "N", 8, 2)
	f.Offset = 5
	f.setFloatValue(recordBuf, 123.45)
	require.Equal(t, []byte("  123.45"), recordBuf[5:13])
}
