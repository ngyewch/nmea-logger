package main

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BertoldVdb/go-ais"
	"github.com/BertoldVdb/go-ais/aisnmea"
	"github.com/ulikunitz/xz"
	"github.com/urfave/cli/v3"
)

const (
	ignoreParseErrors = true
)

func doAisConvert(ctx context.Context, cmd *cli.Command) error {
	logFile := cmd.StringArg(inputFileArg.Name)
	if logFile == "" {
		return fmt.Errorf(inputFileArg.Name + " is required")
	}

	f, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	var reader io.Reader
	reader = f
	if strings.HasSuffix(logFile, ".gz") {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer func(reader io.ReadCloser) {
			_ = reader.Close()
		}(gzipReader)
		reader = gzipReader
	} else if strings.HasSuffix(logFile, ".bz2") {
		bzip2Reader := bzip2.NewReader(reader)
		reader = bzip2Reader
	} else if strings.HasSuffix(logFile, ".xz") {
		xzReader, err := xz.NewReader(reader)
		if err != nil {
			return err
		}
		reader = xzReader
	}

	aisCodec := ais.CodecNew(false, false)
	aisCodec.DropSpace = true
	nmeaCodec := aisnmea.NMEACodecNew(aisCodec)

	w := os.Stdout

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		logLineBytes := scanner.Bytes()

		jsonDecoder := json.NewDecoder(bytes.NewReader(logLineBytes))
		jsonDecoder.DisallowUnknownFields()

		var record LoggerRecord
		err = jsonDecoder.Decode(&record)
		if err != nil {
			return err
		}

		decoded, err := nmeaCodec.ParseSentence(record.NMEA)
		if err != nil {
			if !ignoreParseErrors {
				return err
			}
		}

		if decoded != nil {
			record := &AISRecord{
				Timestamp: record.Timestamp,
				AIS:       decoded,
			}
			jsonBytes, err := json.Marshal(record)
			if err != nil {
				return err
			}
			_, err = w.Write(jsonBytes)
			if err != nil {
				return err
			}
			_, err = w.Write([]byte{'\n'})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type AISRecord struct {
	Timestamp int64              `json:"timestamp"`
	AIS       *aisnmea.VdmPacket `json:"ais"`
}
