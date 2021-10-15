package xbase

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"
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

func TestSetPanic(t *testing.T) {
	db := New()
	db.SetPanic(true)
	require.Equal(t, true, db.IsPanic())
	require.Panics(t, func() { db.CreateFile("./testdata/test.dbf") })
	require.Error(t, db.Error())
}

func TestSetFieldValueError(t *testing.T) {
	db := New()
	db.OpenFile("./testdata/rec0.dbf", true)
	db.Add()

	db.SetFieldValue(0, true)
	require.Error(t, db.Error())
	require.Equal(t, "xbase: SetFieldValue: field 0: field number out of range", db.Error().Error())
}

func TestAddFieldError(t *testing.T) {
	db := New()
	db.AddField("NAME", "X", 10)
	require.Error(t, db.Error())
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

func copyFile(src, dst string) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		panic(err)
	}
}

func TestOpenEditRec(t *testing.T) {
	copyFile("./testdata/rec3.dbf", "./testdata/test1.dbf")

	db := New()
	db.OpenFile("./testdata/test1.dbf", false)

	db.GoTo(2)
	db.SetFieldValue(1, "Edit")
	db.Save()

	db.First()
	db.Next()
	require.Equal(t, "Edit", db.FieldValueAsString(1))
	require.Equal(t, int64(3), db.RecCount())

	db.CloseFile()
	require.NoError(t, db.Error())
}

func TestOpenAddRec(t *testing.T) {
	copyFile("./testdata/rec3.dbf", "./testdata/test2.dbf")

	db := New()
	db.OpenFile("./testdata/test2.dbf", false)

	db.Add()
	db.SetFieldValue(1, "Add")
	db.Save()

	db.First()
	db.Last()
	require.Equal(t, "Add", db.FieldValueAsString(1))
	require.Equal(t, int64(4), db.RecCount())

	db.CloseFile()
	require.NoError(t, db.Error())
}

func TestCreateEditRec(t *testing.T) {
	db := New()
	db.AddField("NAME", "C", 3)
	db.CreateFile("./testdata/test.dbf")

	db.Add()
	db.Save()

	db.Add()
	db.Save()

	db.Add()
	db.Save()

	db.First()
	db.Next()
	db.SetFieldValue(1, "Abc")
	db.Save()

	db.First()
	require.Equal(t, "", db.FieldValueAsString(1))

	db.GoTo(2)
	require.Equal(t, "Abc", db.FieldValueAsString(1))

	require.Equal(t, int64(3), db.RecCount())

	db.CloseFile()
	require.NoError(t, db.Error())
}

func TestTryGoTo(t *testing.T) {
	db := New()
	db.SetPanic(true)
	defer db.CloseFile()
	db.AddField("NAME", "C", 30)
	db.CreateFile("./testdata/test.dbf")
	db.Add()
	db.SetFieldValue(1, "Abc")
	db.Save()
	db.Flush()
	//time.Sleep(time.Second)
	var wg sync.WaitGroup
	wg.Add(1)
	long := strings.Repeat("abc", 10)
	go func() {
		otdb := New()
		otdb.SetPanic(true)
		otdb.OpenFile("./testdata/test.dbf", false)
		otdb.Add()
		otdb.SetFieldValue(1, long)
		otdb.Save()
		otdb.Flush()
		otdb.Add()
		otdb.SetFieldValue(1, "abc")
		otdb.Save()
		otdb.Flush()
		otdb.CloseFile()
		wg.Done()
	}()
	wg.Wait()
	if db.TryGoTo(2) {
		require.Equal(t, long, db.FieldValueAsString(1))
	} else {
		t.Fatal()
	}
	db.goTo(3)
	require.Equal(t, int64(3), db.recCount())
	db.Add()
	db.SetFieldValue(1, "AbcAbc")
	db.Save()
	db.Flush()
	if db.TryGoTo(4) {
		require.Equal(t, "AbcAbc", db.FieldValueAsString(1))
	} else {
		t.Fatal()
	}
	db.goTo(5)
	if db.err == nil {
		t.Fatal("must error")
	}
	require.Equal(t, int64(4), db.recCount())
}
