package json

import (
	"context"
	"encoding/json"
	"io"

	s2 "github.com/takanoriyanagitani/go-sql2schema"
	. "github.com/takanoriyanagitani/go-sql2schema/util"

	ser "github.com/takanoriyanagitani/go-sql2schema/ser"
)

func TypesToWriter(typs []s2.ColumnType, w io.Writer) error {
	var enc *json.Encoder = json.NewEncoder(w)
	for _, typ := range typs {
		e := enc.Encode(typ)
		if nil != e {
			return e
		}
	}

	return nil
}

func SchemaToTypesToWriter(s s2.Schema, w io.Writer) error {
	return TypesToWriter(s, w)
}

func JsonSchemaWriterNew(w io.Writer) ser.SchemaWriter {
	return func(s s2.Schema) IO[Void] {
		return func(_ context.Context) (Void, error) {
			return Empty, SchemaToTypesToWriter(s, w)
		}
	}
}

var CreateSchemaWriter ser.WriterToSchemaWriter = JsonSchemaWriterNew
