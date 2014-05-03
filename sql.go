package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	timeType = reflect.TypeOf(time.Now())
)

type Field struct {
	Name string
	Type reflect.Type
}

type Table struct {
	Name   string
	Fields []Field
}

type Metadata struct {
	Args []string
	// Args without connect string
	SafeArgs []string
	Package  string
	Tables   []Table
}

// parses the mysql type string and returns
// the basic type and its sign (if any)
func parseMysqlType(s string) (typ string, sign string) {
	fields := strings.Split(s, " ")
	if len(fields) > 1 {
		sign = fields[1]
	}

	fields2 := strings.Split(fields[0], "(")
	typ = fields2[0]

	return
}

// connect to mysql and return all
// of the tables and their fields
func mysql(db *sql.DB) (*Metadata, error) {
	rows, err := db.Query("show tables")
	if err != nil {
		return nil, err
	}

	md := &Metadata{}

	var tableNames []string
	var tableName string
	for rows.Next() {
		rows.Scan(&tableName)
		tableNames = append(tableNames, tableName)
	}

	for _, tableName := range tableNames {
		rows, err := db.Query(fmt.Sprintf("show columns from `%s`", tableName))
		if err != nil {
			return nil, err
		}

		table := Table{
			Name: tableName,
		}
		var field, mtyp, null string
		for rows.Next() {
			rows.Scan(&field, &mtyp, &null, &null, &null, &null)

			// default type
			vtype := reflect.TypeOf("")

			typ, _ := parseMysqlType(mtyp)
			switch typ {
			case "tinyint", "smallint", "mediumint", "int", "bigint":
				// if sign == "unsigned" {
				//   vtype = reflect.TypeOf(uint64(0))
				// } else {
				vtype = reflect.TypeOf(int64(0))
				// }
			case "decimal", "float", "double":
				vtype = reflect.TypeOf(float64(0))
			case "blob", "tinyblog", "mediumblob", "longblob":
				vtype = reflect.TypeOf([]byte{})
			case "datetime":
				vtype = timeType
			}

			table.Fields = append(table.Fields, Field{
				Name: field,
				Type: vtype,
			})
		}

		md.Tables = append(md.Tables, table)
	}

	return md, nil
}

// connect to postgresql and return all
// of the tables and their fields
func postgresql(db *sql.DB) (*Metadata, error) {
	//"SELECT column_name, data_type FROM information_schema.columns WHERE table_name = ?"
	return nil, nil
}

// parses the sqlite3 type string and returns
// the basic type
func parseSqlite3Type(s string) (typ string, sign string) {
	fields := strings.Split(s, "(")
	typ = fields[0]

	if strings.Contains(s, "unsigned") {
		sign = "unsigned"
	}

	return
}

func sqlite3(db *sql.DB) (*Metadata, error) {
	rows, err := db.Query("SELECT tbl_name FROM sqlite_master WHERE type = ?", "table")
	if err != nil {
		return nil, err
	}

	md := &Metadata{}

	var tableNames []string
	var tableName string
	for rows.Next() {
		rows.Scan(&tableName)
		tableNames = append(tableNames, tableName)
	}

	for _, tableName := range tableNames {
		rows, err := db.Query(fmt.Sprintf("PRAGMA TABLE_INFO(`%s`)", tableName))
		if err != nil {
			return nil, err
		}

		table := Table{
			Name: tableName,
		}
		var field, styp, null string
		for rows.Next() {
			rows.Scan(&null, &field, &styp, &null, &null, &null)

			// default type
			vtype := reflect.TypeOf("")

			typ, _ := parseSqlite3Type(strings.ToLower(styp))
			switch typ {
			case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "unsigned big int", "big int", "int2", "int8":
				// if sign == "unsigned" {
				//   vtype = reflect.TypeOf(uint64(0))
				// } else {
				vtype = reflect.TypeOf(int64(0))
				// }
			case "real", "double", "double precision", "float", "numeric", "decimal":
				vtype = reflect.TypeOf(float64(0))
			case "bit", "boolean", "bool":
				vtype = reflect.TypeOf(true)
			case "date", "datetime":
				vtype = timeType
			}

			table.Fields = append(table.Fields, Field{
				Name: field,
				Type: vtype,
			})
		}

		md.Tables = append(md.Tables, table)
	}

	return md, nil
}
