package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	s2 "github.com/takanoriyanagitani/go-sql2schema"
	. "github.com/takanoriyanagitani/go-sql2schema/util"

	ser "github.com/takanoriyanagitani/go-sql2schema/ser"
	js "github.com/takanoriyanagitani/go-sql2schema/ser/json/std"
)

var EnvValByKey func(string) IO[string] = Lift(
	func(key string) (string, error) {
		val, found := os.LookupEnv(key)
		switch found {
		case true:
			return val, nil
		default:
			return "", fmt.Errorf("env var %s missing", key)
		}
	},
)

var schemaName IO[string] = EnvValByKey("ENV_SCHEMA_NAME").
	Or(Of("public"))

var tableName IO[string] = EnvValByKey("ENV_TABLE_NAME")

const tableCheckQuery string = `
	SELECT table_name FROM information_schema.tables
	WHERE
	  table_schema = $1
	  AND table_name = $2
	LIMIT 1
`

type Config struct {
	TableName  string
	SchemaName string
}

var config IO[Config] = Bind(
	All(
		tableName,
		schemaName,
	),
	Lift(func(s []string) (Config, error) {
		return Config{
			TableName:  s[0],
			SchemaName: s[1],
		}, nil
	}),
)

type Database struct {
	*sql.DB
}

func (d Database) QueryRow(
	trustedQuery string,
	args ...any,
) IO[*sql.Row] {
	return func(ctx context.Context) (*sql.Row, error) {
		row := d.DB.QueryRowContext(ctx, trustedQuery, args...)
		return row, row.Err()
	}
}

func (d Database) FoundTableName(cfg Config) IO[string] {
	var buf string
	return func(ctx context.Context) (string, error) {
		row, e := d.QueryRow(
			tableCheckQuery,
			cfg.SchemaName,
			cfg.TableName,
		)(ctx)
		if nil != e {
			return "", e
		}

		e = row.Scan(&buf)
		return buf, e
	}
}

func (d Database) CheckedQuery(cfg Config) IO[string] {
	return Bind(
		d.FoundTableName(cfg),
		Lift(func(checkedTableName string) (string, error) {
			return fmt.Sprintf(
				`
					SELECT * FROM %s
					LIMIT 1
				`,
				checkedTableName,
			), nil
		}),
	)
}

func (d Database) toRows(cfg Config) IO[*sql.Rows] {
	return Bind(
		d.CheckedQuery(cfg),
		func(checkedQuery string) IO[*sql.Rows] {
			return func(ctx context.Context) (*sql.Rows, error) {
				return d.DB.QueryContext(
					ctx,
					checkedQuery,
				)
			}
		},
	)
}

func (d Database) ToSchema(cfg Config) IO[s2.Schema] {
	return Bind(
		d.toRows(cfg),
		Lift(func(rows *sql.Rows) (s2.Schema, error) {
			defer rows.Close()
			return s2.Rows{Rows: rows}.ToSchema()
		}),
	)
}

var writerGen ser.WriterToSchemaWriter = js.CreateSchemaWriter
var stdoutWriter func(s2.Schema) IO[Void] = writerGen.SchemaToStdout()

func (d Database) ConfigToSchemaToStdout(cfg Config) IO[Void] {
	return Bind(
		d.ToSchema(cfg),
		stdoutWriter,
	)
}

var database IO[Database] = func(_ context.Context) (Database, error) {
	db, err := sql.Open("pgx", "")
	return Database{
		DB: db,
	}, err
}

var cfg2db2schema2json2stdout IO[Void] = Bind(
	database,
	func(d Database) IO[Void] {
		return Bind(
			config,
			d.ConfigToSchemaToStdout,
		)
	},
)

var sub IO[Void] = func(ctx context.Context) (Void, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return cfg2db2schema2json2stdout(ctx)
}

func main() {
	_, e := sub(context.Background())
	if nil != e {
		log.Printf("%v\n", e)
	}
}
