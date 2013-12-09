package main

import (
	"bytes"
	"database/sql"
	goflag "flag"
	"fmt"
	_ "github.com/bmizerany/pq"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	helpMsg = `kdb is a tool to connect to a database and output Go structs
mirroring the tables and fields.

Usage: 

\tkdb [OPTIONS] <db> <db connect string>

Databases:

\tmysql     \thttp://github.com/go-sql-driver/mysql
\tpostgresql\thttp://github.com/bmizerany/pq
\tsqlite3   \thttp://github.com/mattn/go-sqlite3

Consult the above links to see what the db connect string should be.

Example db connect strings:

\tmysql     \tuser:pass@tcp(host:port)/db
\tpostgresql\tuser=pqgotest dbname=pqgotest sslmode=verify-full
\tsqlite3   \t./foo.db

Options:

\t-structname <name> \tset the format of the struct name
\t                   \tdefault: "capitalize,nounderscore"
\t-fieldname  <name> \tset the format of the field name
\t                   \tdefault: "lowercase,capitalize,nounderscore"
\t-output    <file>  \tset an output file
\t                   \twhen not set, outputs to stdout
\t-sqlstruct         \toutput structs that work with sqlstruct 
\t                   \t(github.com/kisielk/sqlstruct)
\t-omitgen           \tomit the generated comment at the top
\t-package <name>    \twhat the generated package should be
\t-types             \twhat the struct field types should be.
\t                   \tvalues: base, null, pointer
\t                   \tdefault: base

Available formats are the following:

\tcapitalize\tCapitalize the first letter of the name
\tlowercase\tConvert the whole name to lower case
\tnounderscore\tRemove all underscores

Note: the order of the formats specified matters. "lowercase,capitalize"
is not the same as "capitalize,lowercase".

Examples:

\tkdb -output model.go mysql "user:pass@tcp(localhost:3306)/test?charset=utf8"
\tkdb sqlite3 "./foo.db"

`
)

// var (
//   BUILTINS = []string{"break", "default", "func", "interface", "select",
//     "case", " defer", "go", "map", "struct",
//     "chan", " else", " goto", " package", "switch",
//     "const", "fallthrough", "if", "range", "type",
//     "continue", "for", "import", "return", "var"}
// )

var (
	flag             = goflag.NewFlagSet(os.Args[0], goflag.ExitOnError)
	tabWidth         = flag.Int("tabwidth", 2, "tab width")
	structNameFormat = flag.String("structname", "capitalize,nounderscore", "")
	fieldNameFormat  = flag.String("fieldname", "", "")
	output           = flag.String("output", "", "")
	sqlstruct        = flag.Bool("sqlstruct", false, "")
	omitgen          = flag.Bool("omitgen", false, "")
	types            = flag.String("types", "base", "")
	packge           = flag.String("package", "model", "")
)

// capitalize the first letter of the string
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// return a string all lower case
func lowercase(s string) string {
	return strings.ToLower(s)
}

// remove all underscores from the string
func nounderscore(s string) string {
	return strings.Replace(s, "_", "", -1)
}

// format a name based on the flags passed.
// the order of the flags matter.
func formatName(n string, flags string) string {
	formats := strings.Split(flags, ",")
	for _, f := range formats {
		switch f {
		case "capitalize":
			n = capitalize(n)
		case "nounderscore":
			n = nounderscore(n)
		case "lowercase":
			n = lowercase(n)
		}
	}

	return n
}

// format a struct name
func formatStructName(n string) string {
	return formatName(n, *structNameFormat)
}

// format a field name
func formatFieldName(n string) string {
	return formatName(n, *fieldNameFormat)
}

// "go fmt"'s the code
func format(w io.Writer, code []byte) error {
	fset := token.NewFileSet()

	ast, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return err
	}

	cfg := &printer.Config{Mode: printer.UseSpaces, Tabwidth: *tabWidth}
	err = cfg.Fprint(w, fset, ast)
	if err != nil {
		return err
	}

	return nil
}

func usage() {
	fmt.Fprint(os.Stderr, strings.Replace(helpMsg, "\\t", "\t", -1))
}

func fatal(err error) {
	fmt.Fprint(os.Stderr, err)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}

func main() {
	var err error

	flag.Usage = usage
	flag.Parse(os.Args[1:])

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	db, err := sql.Open(flag.Arg(0), flag.Arg(1))
	if err != nil {
		fatal(err)
	}

	md := &Metadata{
		Package: *packge,
		Args:    os.Args,
	}

	switch flag.Arg(0) {
	case "mysql":
		err = mysql(md, db)
	case "postgresql":
		err = postgresql(md, db)
	case "sqlite3":
		err = sqlite3(md, db)
	}

	if err != nil {
		fatal(err)
	}

	file := os.Stdout
	if *output != "" {
		file, err = os.Create(*output)
		if err != nil {
			fatal(err)
		}

		defer file.Close()
	}

	buffer := &bytes.Buffer{}
	md.Create().Output(buffer)
	//io.Copy(file, buffer)
	err = format(file, buffer.Bytes())
	if err != nil {
		fatal(err)
	}

	// funcMap := template.FuncMap{
	//   "format": format,
	// }

	// t, err := template.New("output").Funcs(funcMap).ParseFiles("output.tpl")
	// if err != nil {
	//   fatal(err)
	// }

	// err = t.ExecuteTemplate(file, "output.tpl", md.Create())
	// if err != nil {
	//   fatal(err)
	// }

	os.Exit(0)
}
