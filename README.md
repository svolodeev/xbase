# xbase

[![Go Reference](https://pkg.go.dev/badge/github.com/svolodeev/xbase.svg)](https://pkg.go.dev/github.com/svolodeev/xbase)

A pure Go library for working with [DBF](http://en.wikipedia.org/wiki/DBase#File_formats) files. The main purpose of the library is to organize export-import of data from files in DBF format.

The XBase type is used to work with DBF files. In addition to working with existing files, the XBase object allows you to create a new file of the given structure. Each XBase object can be linked with only one file.

### Writing changes to a file
The XBase object contains data for one current record. Changing field values ​​does not cause an immediate change to the file. Changes are saved when the __Save()__ method is called.

### Deleting records
Deleting a record does not physically destroy it on disk. The deletion mark is put in a special field of the record.

### Error processing
If an error occurs when calling the method, use the __Error()__ method to get its value. By default, methods don't panic. This behavior can be changed. If you call __SetPanic(true)__, then when an error occurs, the methods will cause a panic. Use whichever is more convenient for you.

### Limitations
The following field types are supported: __C__, __N__, __L__, __D__. Memo fields are not supported. Index files are not supported.

## Examples
File creation.
    
    // Create file
    db := xbase.New()
    db.AddField("NAME", "C", 30)
    db.AddField("SALARY", "N", 9, 2)
    db.AddField("BDATE", "D")
    db.SetCodePage(1251)
    db.CreateFile("persons.dbf")
    if db.Error() != nil {
        return db.Error()
    }
    defer db.CloseFile()
    
    // Add record
    db.Add()
    db.SetFieldValue(1, "John Smith")
    db.SetFieldValue(2, 1234.56)
    db.SetFieldValue(3, time.Date(1998, 2, 20, 0, 0, 0, 0, time.UTC))
    db.Save()
    if db.Error() != nil {
        return db.Error()
    }

Reading file.

    db := xbase.New()
    db.SetPanic(true)
    
    db.OpenFile("persons.dbf", true)
    defer db.CloseFile()
    
    db.First()
    for !db.EOF() {
        name := db.FieldValueAsString(1)
        salary := db.FieldValueAsFloat(2)
        bDate := db.FieldValueAsDate(3)
        fmt.Println(name, salary, bDate)
        db.Next()
    }

File information.

    db := xbase.New()
    db.SetPanic(true)
    
    db.OpenFile("persons.dbf", true)
    defer db.CloseFile()

    fmt.Println("Record count:", db.RecCount())
    fmt.Println("Field count:", db.FieldCount())
    fmt.Println("Code page:", db.CodePage())

    // File structure
    for n := 1; n <= db.FieldCount(); n++ {
        name, typ, length, dec := db.FieldInfo(n)
        fmt.Println(name, typ, length, dec)
    }

## License
Copyright (C) Sergey Volodeev. Released under MIT license.