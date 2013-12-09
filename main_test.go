package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
)

func TestSqlite3(t *testing.T) {
	os.Remove("./foo.db")
	defer os.Remove("./foo.db")

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	sqls := []string{
		"create table foo (id integer not null primary key, name text);",
		`CREATE TABLE Product(
    ProductID                   int,
    AlchemyMarketedProductID    int,
    ReplacedByProductID         int,
    ReplacedByDate              datetime,
    ProductNameShort            varchar(30),
    ProductNameLong             varchar(255),
    ProductNameTypeID           int,
    PrescribingName             varchar(35),
    MarketerID                  int,
    LegendStatusID              int,
    BrandGenericStatusID        int,
    DEAClassificationID         int,
    OnMarket                    datetime,
    OffMarket                   datetime,
    DESIStatusID                int,
    CP_NUM                      int,
    LicenseTypeID               int,
    Repackaged                  bit,
    Innovator                   bit,
    PrivateLabel                bit, 
    ModifiedAction char(1));`,
	}
	for _, sql := range sqls {
		_, err = db.Exec(sql)
		if err != nil {
			fmt.Printf("%q: %s\n", err, sql)
			return
		}
	}

	structs, err := sqlite3(db)
	if err != nil {
		t.Fatal(err)
	}
	byts := &bytes.Buffer{}
	*omitgen = true
	generate(byts, structs)
	err = format(byts, byts.Bytes())
	if err != nil {
		t.Fatal(err)
	}
}
