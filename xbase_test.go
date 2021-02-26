package xbase

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func readFile(name string) []byte {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	// ModDate
	b[1] = 0
	b[2] = 0
	b[3] = 0
	// CodePage
	// b[29] = 0
	return b
}

func addFields(db *XBase) {
	db.AddField("NAME", "C", 20)
	db.AddField("FLAG", "L")
	db.AddField("COUNT", "N", 5)
	db.AddField("PRICE", "N", 9, 2)
	db.AddField("DATE", "D")

	db.SetCodePage(866)
}

func TestCreateEmptyFile(t *testing.T) {
	db := New()
	addFields(db)
	db.CreateFile("./testdata/test.dbf")

	require.Equal(t, int64(0), db.RecCount())
	require.Equal(t, 5, db.FieldCount())
	require.Equal(t, int64(0), db.RecNo())
	require.Equal(t, true, db.EOF())
	require.Equal(t, true, db.BOF())

	db.CloseFile()
	require.NoError(t, db.Error())

	testBytes := readFile("./testdata/test.dbf")
	goldBytes := readFile("./testdata/rec0.dbf")
	require.Equal(t, goldBytes, testBytes)

}

func TestAddEmptyRec(t *testing.T) {
	db := New()
	addFields(db)
	db.CreateFile("./testdata/test.dbf")

	db.Add()
	db.Save()

	require.Equal(t, int64(1), db.RecCount())
	require.Equal(t, int64(1), db.RecNo())
	require.Equal(t, false, db.EOF())
	require.Equal(t, false, db.BOF())

	db.CloseFile()
	require.NoError(t, db.Error())

	testBytes := readFile("./testdata/test.dbf")
	goldBytes := readFile("./testdata/rec1.dbf")
	require.Equal(t, goldBytes, testBytes)
}

func TestAddRecords(t *testing.T) {
	db := New()
	addFields(db)
	db.CreateFile("./testdata/test.dbf")

	d := time.Date(2021, 2, 12, 0, 0, 0, 0, time.UTC)

	db.Add()
	db.SetFieldValue(1, "Abc")
	db.SetFieldValue(2, true)
	db.SetFieldValue(3, 123)
	db.SetFieldValue(4, 123.45)
	db.SetFieldValue(5, d)
	db.Save()

	db.Add()
	db.Save()

	db.Add()
	db.SetFieldValue(1, "Мышь")
	db.SetFieldValue(2, false)
	db.SetFieldValue(3, -321)
	db.SetFieldValue(4, -54.32)
	db.SetFieldValue(5, d)
	db.Save()

	require.Equal(t, int64(3), db.RecCount())

	db.CloseFile()
	require.NoError(t, db.Error())

	testBytes := readFile("./testdata/test.dbf")
	goldBytes := readFile("./testdata/rec3.dbf")
	require.Equal(t, goldBytes, testBytes)
}

func TestOpenEmptyFile(t *testing.T) {
	db := New()
	db.OpenFile("./testdata/rec0.dbf", true)

	require.Equal(t, int64(0), db.RecCount())
	require.Equal(t, 5, db.FieldCount())
	require.Equal(t, true, db.EOF())
	require.Equal(t, true, db.BOF())

	db.First()
	require.Equal(t, true, db.EOF())
	require.Equal(t, true, db.BOF())

	db.Next()
	require.Equal(t, true, db.EOF())
	require.Equal(t, true, db.BOF())

	db.Last()
	require.Equal(t, true, db.EOF())
	require.Equal(t, true, db.BOF())

	db.CloseFile()
	require.NoError(t, db.Error())
}

func TestReadEmptyRec(t *testing.T) {
	db := New()
	db.OpenFile("./testdata/rec1.dbf", true)

	db.First()
	require.Equal(t, "", db.FieldValueAsString(1))
	require.Equal(t, false, db.FieldValueAsBool(2))
	require.Equal(t, int64(0), db.FieldValueAsInt(3))
	require.Equal(t, float64(0), db.FieldValueAsFloat(4))
	var d time.Time
	require.Equal(t, d, db.FieldValueAsDate(5))

	db.CloseFile()
	require.NoError(t, db.Error())
}

func TestReadNext(t *testing.T) {
	db := New()
	db.OpenFile("./testdata/rec3.dbf", true)

	db.First()
	require.Equal(t, int64(1), db.RecNo())
	require.Equal(t, false, db.EOF())
	require.Equal(t, "Abc", db.FieldValueAsString(1))
	require.Equal(t, int64(123), db.FieldValueAsInt(3))

	db.Next()
	require.Equal(t, int64(2), db.RecNo())
	require.Equal(t, false, db.EOF())
	require.Equal(t, "", db.FieldValueAsString(1))
	require.Equal(t, int64(0), db.FieldValueAsInt(3))

	db.Next()
	require.Equal(t, int64(3), db.RecNo())
	require.Equal(t, false, db.EOF())
	require.Equal(t, "Мышь", db.FieldValueAsString(1))
	require.Equal(t, int64(-321), db.FieldValueAsInt(3))

	db.Next()
	require.Equal(t, true, db.EOF())

	db.Next()
	require.Equal(t, true, db.EOF())

	db.CloseFile()
	require.NoError(t, db.Error())
}

func TestReadPrev(t *testing.T) {
	db := New()
	db.OpenFile("./testdata/rec3.dbf", true)

	db.Last()
	require.Equal(t, int64(3), db.RecNo())
	require.Equal(t, false, db.EOF())
	require.Equal(t, "Мышь", db.FieldValueAsString(1))
	require.Equal(t, int64(-321), db.FieldValueAsInt(3))

	db.Prev()
	require.Equal(t, int64(2), db.RecNo())
	require.Equal(t, false, db.EOF())
	require.Equal(t, "", db.FieldValueAsString(1))
	require.Equal(t, int64(0), db.FieldValueAsInt(3))

	db.Prev()
	require.Equal(t, int64(1), db.RecNo())
	require.Equal(t, false, db.EOF())
	require.Equal(t, "Abc", db.FieldValueAsString(1))
	require.Equal(t, int64(123), db.FieldValueAsInt(3))

	db.Prev()
	require.Equal(t, true, db.BOF())

	db.Prev()
	require.Equal(t, true, db.BOF())

	db.CloseFile()
	require.NoError(t, db.Error())
}
