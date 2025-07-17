package main

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/arthurkiller/rollingwriter"
	"github.com/urfave/cli/v3"
	"go.bug.st/serial"
	"io"
	"os"
	"time"
)

func doMain(ctx context.Context, cmd *cli.Command) error {
	outputDir := cmd.String(outputDirFlag.Name)
	serialPort := cmd.String(serialPortFlag.Name)
	baudRate := cmd.Int(baudRateFlag.Name)
	dataBits := cmd.Int(dataBitsFlag.Name)
	parity0 := cmd.String(parityFlag.Name)
	stopBits0 := cmd.String(stopBitsFlag.Name)

	parity := serial.NoParity
	switch parity0 {
	case "N":
		parity = serial.NoParity
	case "E":
		parity = serial.EvenParity
	case "O":
		parity = serial.OddParity
	case "M":
		parity = serial.MarkParity
	case "S":
		parity = serial.SpaceParity
	}
	stopBits := serial.OneStopBit
	switch stopBits0 {
	case "1":
		stopBits = serial.OneStopBit
	case "1.5":
		stopBits = serial.OnePointFiveStopBits
	case "2":
		stopBits = serial.TwoStopBits
	}

	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: dataBits,
		Parity:   parity,
		StopBits: stopBits,
	}
	port, err := serial.Open(serialPort, mode)
	if err != nil {
		return err
	}
	defer func(port serial.Port) {
		_ = port.Close()
	}(port)

	rollingWriterConfig := rollingwriter.NewDefaultConfig()
	rollingWriterConfig.LogPath = outputDir
	rollingWriterConfig.FileName = "nmea"

	rollingWriter, err := rollingwriter.NewWriterFromConfig(&rollingWriterConfig)
	if err != nil {
		return err
	}
	defer func(rollingWriter rollingwriter.RollingWriter) {
		_ = rollingWriter.Close()
	}(rollingWriter)
	w := io.MultiWriter(os.Stdout, rollingWriter)

	scanner := bufio.NewScanner(port)
	for scanner.Scan() {
		record := &Record{
			Timestamp: time.Now().UnixMilli(),
			NMEA:      scanner.Text(),
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

	return nil
}

type Record struct {
	Timestamp int64  `json:"timestamp"`
	NMEA      string `json:"nmea"`
}
