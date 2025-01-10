package sql2schema

import (
	"database/sql"
	"reflect"
)

type Schema []ColumnType

type Rows struct{ *sql.Rows }

func (r Rows) ToSchema() (Schema, error) {
	typs, e := r.Rows.ColumnTypes()
	rs := RawSchema(typs)
	return rs.ToSchema(), e
}

type RawSchema []*sql.ColumnType

func (r RawSchema) ToSchema() Schema {
	ret := make([]ColumnType, 0, len(r))
	buf := RawColumnType{ColumnType: nil}
	for _, raw := range r {
		buf.ColumnType = raw
		ret = append(ret, buf.ToColumnType())
	}
	return ret
}

type RawColumnType struct {
	*sql.ColumnType
}

func (r RawColumnType) ToDecimalSize() sql.Null[DecimalSize] {
	var ret sql.Null[DecimalSize]
	prec, sc, ok := r.ColumnType.DecimalSize()
	ret.Valid = ok
	ret.V = DecimalSize{
		Precision: prec,
		Scale:     sc,
	}
	return ret
}

func (r RawColumnType) ToLength() Length {
	var ret Length
	l, ok := r.ColumnType.Length()
	ret.Valid = ok
	ret.V = l
	return ret
}

func (r RawColumnType) ToNullable() Nullable {
	var ret Nullable
	n, ok := r.ColumnType.Nullable()
	ret.Valid = ok
	ret.V = n
	return ret
}

func (r RawColumnType) ToType() Type {
	return ReflectToType(r.ColumnType.ScanType())
}

func (r RawColumnType) ToColumnType() ColumnType {
	return ColumnType{
		DatabaseTypeName: r.ColumnType.DatabaseTypeName(),
		DecimalSize:      r.ToDecimalSize(),
		Length:           r.ToLength(),
		Name:             r.ColumnType.Name(),
		Nullable:         r.ToNullable(),
		Type:             r.ToType(),
	}
}

type DecimalSize struct {
	Precision int64
	Scale     int64
}

type Length sql.Null[int64]

type Nullable sql.Null[bool]

type Type struct {
	Align      int
	FieldAlign int
	Name       string
	PkgPath    string
	Size       uintptr
	String     string
	reflect.Kind
	Comparable bool
}

func ReflectToType(t reflect.Type) Type {
	return Type{
		Align:      t.Align(),
		FieldAlign: t.FieldAlign(),
		Name:       t.Name(),
		PkgPath:    t.PkgPath(),
		Size:       t.Size(),
		String:     t.String(),
		Kind:       t.Kind(),
		Comparable: t.Comparable(),
	}
}

type ColumnType struct {
	DatabaseTypeName string
	DecimalSize      sql.Null[DecimalSize]
	Length
	Name string
	Nullable
	Type `json:"typ"`
}
