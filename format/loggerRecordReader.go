package format

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

type LoggerRecordReader struct {
	scanner *bufio.Scanner
}

func NewLoggerRecordReader(r io.Reader) *LoggerRecordReader {
	scanner := bufio.NewScanner(r)
	return &LoggerRecordReader{
		scanner: scanner,
	}
}

func (reader *LoggerRecordReader) ReadLoggerRecord() (*LoggerRecord, error) {
	if !reader.scanner.Scan() {
		return nil, reader.scanner.Err()
	}

	logLineBytes := reader.scanner.Bytes()

	jsonDecoder := json.NewDecoder(bytes.NewReader(logLineBytes))
	jsonDecoder.DisallowUnknownFields()

	var loggerRecord LoggerRecord
	err := jsonDecoder.Decode(&loggerRecord)
	if err != nil {
		return nil, err
	}

	return &loggerRecord, nil
}
