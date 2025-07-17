package main

import (
	"context"
	"fmt"
	slogUtils "github.com/ngyewch/go-clibase/slog-utils"
	"github.com/urfave/cli/v3"
	"log/slog"
	"os"
)

var (
	version string

	log = slogUtils.GetLoggerForCurrentPackage()

	logLevelFlag = &cli.StringFlag{
		Name:     "log-level",
		Usage:    "log level",
		Category: "Logging",
		Value:    "INFO",
		Sources:  cli.EnvVars("LOG_LEVEL"),
		Action: func(ctx context.Context, cmd *cli.Command, s string) error {
			slogUtils.SetLevel(slogUtils.ToLevel(s))
			return nil
		},
	}

	outputDirFlag = &cli.StringFlag{
		Name:    "output-dir",
		Usage:   "output directory",
		Value:   "./logs",
		Sources: cli.EnvVars("OUTPUT_DIR"),
	}
	serialPortFlag = &cli.StringFlag{
		Name:     "serial-port",
		Usage:    "serial port",
		Required: true,
		Sources:  cli.EnvVars("SERIAL_PORT"),
	}
	baudRateFlag = &cli.IntFlag{
		Name:     "baud-rate",
		Usage:    "baud rate",
		Required: true,
		Sources:  cli.EnvVars("BAUD_RATE"),
	}
	dataBitsFlag = &cli.IntFlag{
		Name:    "data-bits",
		Usage:   "data bits",
		Value:   8,
		Sources: cli.EnvVars("DATA_BITS"),
	}
	parityFlag = &cli.StringFlag{
		Name:    "parity",
		Usage:   "parity",
		Value:   "N",
		Sources: cli.EnvVars("PARITY"),
		Action: func(ctx context.Context, cmd *cli.Command, s string) error {
			switch s {
			case "N", "O", "E", "M", "S":
			default:
				return fmt.Errorf("invalid parity")
			}
			return nil
		},
	}
	stopBitsFlag = &cli.StringFlag{
		Name:    "stop-bits",
		Usage:   "stop bits",
		Value:   "1",
		Sources: cli.EnvVars("STOP_BITS"),
		Action: func(ctx context.Context, cmd *cli.Command, s string) error {
			switch s {
			case "1", "1.5", "2":
			default:
				return fmt.Errorf("invalid stop bits")
			}
			return nil
		},
	}

	app = &cli.Command{
		Name:    "nmea-logger",
		Usage:   "NMEA logger",
		Version: version,
		Action:  doMain,
		Flags: []cli.Flag{
			serialPortFlag,
			baudRateFlag,
			dataBitsFlag,
			parityFlag,
			stopBitsFlag,
			outputDirFlag,
			logLevelFlag,
		},
	}
)

func main() {
	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Error("error",
			slog.Any("err", err),
		)
	}
}
