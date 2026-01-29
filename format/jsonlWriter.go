package format

import (
	"encoding/json"
	"io"
)

type JsonlWriter struct {
	w io.Writer
}

func NewJsonlWriter(w io.Writer) *JsonlWriter {
	return &JsonlWriter{
		w: w,
	}
}

func (writer *JsonlWriter) Close() error {
	return nil
}

func (writer *JsonlWriter) WriteRecord(v any) error {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = writer.w.Write(jsonBytes)
	if err != nil {
		return err
	}
	_, err = writer.w.Write([]byte{'\n'})
	if err != nil {
		return err
	}
	return nil
}
