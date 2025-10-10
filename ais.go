package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/BertoldVdb/go-ais"
	"github.com/BertoldVdb/go-ais/aisnmea"
	"github.com/urfave/cli/v3"
)

const (
	ignoreParseErrors = true
)

func doAis(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("insufficient arguments")
	}

	logFile := cmd.Args().First()

	f, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	aisCodec := ais.CodecNew(false, false)
	aisCodec.DropSpace = true
	nmeaCodec := aisnmea.NMEACodecNew(aisCodec)

	w := os.Stdout

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		logLineBytes := scanner.Bytes()

		jsonDecoder := json.NewDecoder(bytes.NewReader(logLineBytes))
		jsonDecoder.DisallowUnknownFields()

		var record Record
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
