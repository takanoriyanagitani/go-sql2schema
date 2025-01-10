package ser

import (
	"io"
	"os"

	s2 "github.com/takanoriyanagitani/go-sql2schema"
	. "github.com/takanoriyanagitani/go-sql2schema/util"
)

type SchemaWriter func(s2.Schema) IO[Void]

type WriterToSchemaWriter func(io.Writer) SchemaWriter

func (w WriterToSchemaWriter) SchemaToStdout() SchemaWriter {
	return w(os.Stdout)
}
