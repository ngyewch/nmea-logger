package format

import (
	"io"
)

type JsonlAISRecordWriter struct {
	jsonlWriter *JsonlWriter
}

func NewJsonlAISRecordWriter(w io.Writer) *JsonlAISRecordWriter {
	return &JsonlAISRecordWriter{
		jsonlWriter: NewJsonlWriter(w),
	}
}

func (writer *JsonlAISRecordWriter) Close() error {
	return writer.jsonlWriter.Close()
}

func (writer *JsonlAISRecordWriter) WriteAISRecord(record *AISRecord) error {
	return writer.jsonlWriter.WriteRecord(record)
}
