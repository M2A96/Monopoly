package dao

//go:generate mockgen -destination=../../test/v2/cursor.go -package=test -mock_names=Cursorer=MockCursor . Cursorer

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"

	"github/M2A96/Monopoly.git/object"

	"gorm.io/gorm"
)

type (
	// Cursorer is an interface.
	Cursorer interface {
		encoding.BinaryMarshaler
		encoding.BinaryUnmarshaler
		// GetOffset is a function.
		GetOffset() uint32
		// Query is a function.
		Query(
			table string,
		) func(*gorm.DB) *gorm.DB
	}

	// GetCursorer is an interface.
	GetCursorer interface {
		// GetCursorer is a function.
		GetCursorer() Cursorer
	}

	cursor struct {
		offset uint32
	}
)

var (
	_ Cursorer                   = (*cursor)(nil)
	_ encoding.BinaryMarshaler   = (*cursor)(nil)
	_ encoding.BinaryUnmarshaler = (*cursor)(nil)
	_ json.Marshaler             = (*cursor)(nil)
	_ object.GetMapper           = (*cursor)(nil)
)

// NewCursor is a function.
func NewCursor(
	offset uint32,
) *cursor {
	return &cursor{
		offset: offset,
	}
}

// CursorerComparer is a function.
func CursorerComparer(
	first Cursorer,
	second Cursorer,
) bool {
	return first.GetOffset() == second.GetOffset()
}

// GetOffset is a function.
func (dao *cursor) GetOffset() uint32 {
	return dao.offset
}

// Query is a function.
func (dao *cursor) Query(
	_ string,
) func(*gorm.DB) *gorm.DB {
	return func(
		gormDB *gorm.DB,
	) *gorm.DB {
		gormDB.Offset(int(dao.GetOffset()))

		return gormDB
	}
}

// GetMap is a function.
func (dao *cursor) GetMap() map[string]any {
	return map[string]any{
		"offset": dao.GetOffset(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (dao *cursor) MarshalJSON() ([]byte, error) {
	return json.Marshal(dao.GetMap())
}

// MarshalBinary is a function.
// read more https://pkg.go.dev/encoding#BinaryMarshaler
func (dao *cursor) MarshalBinary() ([]byte, error) {
	var bytesBuffer bytes.Buffer
	_, err := fmt.Fprintln(&bytesBuffer, dao.offset)

	return bytesBuffer.Bytes(), err
}

// UnmarshalBinary is a function.
// read more https://pkg.go.dev/encoding#BinaryUnmarshaler
func (dao *cursor) UnmarshalBinary(
	data []byte,
) error {
	bytesBuffer := bytes.NewBuffer(data)
	_, err := fmt.Fscanln(bytesBuffer, &dao.offset)

	return err
}
