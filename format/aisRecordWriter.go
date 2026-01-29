package format

import (
	"io"
)

type AISRecordWriter interface {
	io.Closer

	WriteAISRecord(record *AISRecord) error
}
