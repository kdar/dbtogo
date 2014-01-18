package main

import (
	"bytes"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"testing"
)

func TestSqlite3(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	sqls := []string{
		`create table Foo (
      id integer not null primary key, 
      name text
     );
    `,
		`CREATE TABLE Bar(
    message varchar(255),
    date datetime,
    count int,
    awesome bit
    );`,
	}
	for _, sql := range sqls {
		_, err = db.Exec(sql)
		if err != nil {
			t.Fatalf("%q: %s\n", err, sql)
			return
		}
	}

	md, err := sqlite3(db)
	if err != nil {
		t.Fatal(err)
	}

	md.Package = "test"

	byts := &bytes.Buffer{}
	err = render(byts, md, "")
	if err != nil {
		t.Fatal(err)
	}

	err = format(ioutil.Discard, byts.Bytes())
	if err != nil {
		t.Fatal(err)
	}
}
