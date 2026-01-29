package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ngyewch/nmea-logger/format"
	"github.com/ngyewch/nmea-logger/ioutil"
	"github.com/urfave/cli/v3"
)

type RecordPreprocessor interface {
	PreprocessRecord(record *format.AISRecord) error
}

func doAisConvert(ctx context.Context, cmd *cli.Command) error {
	inputFile := cmd.StringArg(inputFileArg.Name)
	outputFile := cmd.StringArg(outputFileArg.Name)
	if inputFile == "" {
		return fmt.Errorf(inputFileArg.Name + " is required")
	}

	var recordWriter format.AISRecordWriter

	if outputFile != "" {
		f, outputFile1, err := ioutil.OpenFileForWriting(outputFile)
		if err != nil {
			return err
		}
		defer func(f io.WriteCloser) {
			_ = f.Close()
		}(f)

		ext := filepath.Ext(outputFile1)
		switch ext {
		case ".jsonl":
			recordWriter = format.NewJsonlAISRecordWriter(f)
			defer func(recordWriter format.AISRecordWriter) {
				_ = recordWriter.Close()
			}(recordWriter)

		case ".csv":
			recordWriter, err = format.NewCsvAISRecordWriter(f)
			if err != nil {
				return err
			}
			defer func(recordWriter format.AISRecordWriter) {
				_ = recordWriter.Close()
			}(recordWriter)

		default:
			return fmt.Errorf("unsupported file extension")
		}
	} else {
		recordWriter = format.NewJsonlAISRecordWriter(os.Stdout)
		defer func(recordWriter format.AISRecordWriter) {
			_ = recordWriter.Close()
		}(recordWriter)
	}

	const ignoreParseErrors = true

	recordPreprocessor, ok := recordWriter.(RecordPreprocessor)
	if ok {
		reader, err := ioutil.OpenFileForReading(inputFile)
		if err != nil {
			return err
		}
		defer func(reader io.ReadCloser) {
			_ = reader.Close()
		}(reader)
		loggerRecordReader := format.NewLoggerRecordReader(reader)
		aisRecordReader := format.NewAISRecordReader(loggerRecordReader, ignoreParseErrors)
		for {
			aisRecord, err := aisRecordReader.ReadAISRecord()
			if err != nil {
				return err
			}
			if aisRecord == nil {
				break
			}
			err = recordPreprocessor.PreprocessRecord(aisRecord)
			if err != nil {
				return err
			}
		}
	}

	reader, err := ioutil.OpenFileForReading(inputFile)
	if err != nil {
		return err
	}
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)
	loggerRecordReader := format.NewLoggerRecordReader(reader)
	aisRecordReader := format.NewAISRecordReader(loggerRecordReader, ignoreParseErrors)
	for {
		aisRecord, err := aisRecordReader.ReadAISRecord()
		if err != nil {
			return err
		}
		if aisRecord == nil {
			break
		}
		err = recordWriter.WriteAISRecord(aisRecord)
		if err != nil {
			return err
		}
	}

	return nil
}
