dbtogo
======

dbtogo is a tool to connect to a database and output Go structs mirroring the tables and fields

The generated code is configurable through modifying the template.

### Support

Right now only supports mysql and sqlite3. Haven't gotten around to implementing postgresql yet.

### Why?

I don't like using ORMs that do too much magic for me. This allows me to generate structs for my massive databases and use database/sql or some other light wrapper ([sqlx](https://github.com/jmoiron/sqlx)).

### Usage

`dbtogo --help` will tell you what you want to know.

    dbtogo mysql mysqluser:pass@tcp(host:port)/db
    dbtogo sqlite3 ./foo.db

### Example

Assume we have the following tables in a sqlite3 database:

    CREATE TABLE Foo (
    id integer not null primary key, 
    name text);

    CREATE TABLE Bar(
    message varchar(255),
    date datetime,
    count int,
    awesome bit);

dbtogo with the default template would output:

    package model

    // GENERATED BY dbtogo (github.com/kdar/dbtogo); DO NOT EDIT
    // ---args:

    import "database/sql"

    var InsertStmts = map[string]string{
        "Foo": "INSERT INTO Foo (Id,Name) VALUES (:Id,:Name)",
        "Bar": "INSERT INTO Bar (Message,Date,Count,Awesome) VALUES (:Message,:Date,:Count,:Awesome)",
    }

    type Foo struct {
        Id   sql.NullInt64
        Name sql.NullString
    }

    type Bar struct {
        Message sql.NullString
        Date    *time.Time
        Count   sql.NullInt64
        Awesome sql.NullBool
    }

### Templating

The example template should get you started.

    package {{.Package}}

    // GENERATED BY dbtogo (github.com/kdar/dbtogo); DO NOT EDIT
    // ---args: {{join .SafeArgs " "}}

    {{define "columns"}}{{$len:=len .}}{{range $i, $v := .}}{{$v.Name|capitalize}}{{if lt $i (sub $len 1)}},{{end}}{{end}}{{end}}
    {{define "values"}}{{$len:=len .}}{{range $i, $v := .}}:{{$v.Name|capitalize}}{{if lt $i (sub $len 1)}},{{end}}{{end}}{{end}}

    import "database/sql"

    var InsertStmts = map[string]string{
    {{range $_, $table := .Tables}}  "{{$table.Name}}": "INSERT INTO {{$table.Name}} ({{template "columns" $table.Fields}}) VALUES ({{template "values" $table.Fields}})",
    {{end}}}

    {{range $_, $table := .Tables}}
    type {{$table.Name}} struct {
    {{range $_, $field := $table.Fields}}  {{$field.Name|capitalize}} {{$field|typenull}}
    {{end}}}
    {{end}}

#### Available template functions

String functions

function          | description             
:-----------------|:--------------------------------------------------------
tolower           | "Awesome" -> "awesome"       
join              | strings.Join            
captialize        | "hello" -> "Hello"  
noundercore       | "im_cool" -> "imcool"
camelize          | "dino_party" -> "DinoParty"
camelizedownfirst | same as camelcase but with first letter downcased
pluralize         | returns the plural form of a singular word
singularize       | returns the singular form of a plural word
tableize          | "SuperPerson" -> "super_people"
typeify           | "something_like_this" -> "SomethingLikeThis"

Math functions

function          | description             
:-----------------|:--------------------------------------------------------
add               | adds two integers
sub               | substracts two integers

Field functions (these functions operate on dbtogo.Field)

function          | description             
:-----------------|:--------------------------------------------------------
typenull          | converts to sql.Null* when appropriate.
                  | makes it a pointer otherwise
typepointer       | converts to a pointer unless it is a slice type