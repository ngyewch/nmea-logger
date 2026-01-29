package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	slogUtils "github.com/ngyewch/go-clibase/slog-utils"
	"github.com/urfave/cli/v3"
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
		Category: "Serial port",
		Required: true,
		Sources:  cli.EnvVars("SERIAL_PORT"),
	}
	baudRateFlag = &cli.IntFlag{
		Name:     "baud-rate",
		Usage:    "baud rate",
		Category: "Serial port",
		Required: true,
		Sources:  cli.EnvVars("BAUD_RATE"),
	}
	dataBitsFlag = &cli.IntFlag{
		Name:     "data-bits",
		Usage:    "data bits",
		Category: "Serial port",
		Value:    8,
		Sources:  cli.EnvVars("DATA_BITS"),
	}
	parityFlag = &cli.StringFlag{
		Name:     "parity",
		Usage:    "parity",
		Category: "Serial port",
		Value:    "N",
		Sources:  cli.EnvVars("PARITY"),
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
		Name:     "stop-bits",
		Usage:    "stop bits",
		Category: "Serial port",
		Value:    "1",
		Sources:  cli.EnvVars("STOP_BITS"),
		Action: func(ctx context.Context, cmd *cli.Command, s string) error {
			switch s {
			case "1", "1.5", "2":
			default:
				return fmt.Errorf("invalid stop bits")
			}
			return nil
		},
	}

	listenAddrFlag = &cli.StringFlag{
		Name:    "listen-addr",
		Usage:   "listen address",
		Value:   "127.0.0.1:8080",
		Sources: cli.EnvVars("LISTEN_ADDR"),
	}
	playbackSpeedFlag = &cli.Float64Flag{
		Name:    "playback-speed",
		Usage:   "playback speed",
		Value:   60,
		Sources: cli.EnvVars("PLAYBACK_SPEED"),
	}
	playbackUpdatePeriodFlag = &cli.DurationFlag{
		Name:    "playback-update-period",
		Usage:   "playback update period",
		Value:   1 * time.Second,
		Sources: cli.EnvVars("PLAYBACK_UPDATE_PERIOD"),
	}

	inputFileArg = &cli.StringArg{
		Name:      "input-file",
		UsageText: "(input-file)",
	}
	outputFileArg = &cli.StringArg{
		Name:      "output-file",
		UsageText: "(output-file)",
	}

	app = &cli.Command{
		Name:    "nmea-logger",
		Usage:   "NMEA logger",
		Version: version,
		Commands: []*cli.Command{
			{
				Name:   "log",
				Usage:  "log",
				Action: doLog,
				Flags: []cli.Flag{
					serialPortFlag,
					baudRateFlag,
					dataBitsFlag,
					parityFlag,
					stopBitsFlag,
					outputDirFlag,
				},
			},
			{
				Name:  "ais",
				Usage: "ais",
				Commands: []*cli.Command{
					{
						Name:   "convert",
						Usage:  "convert",
						Action: doAisConvert,
						Arguments: []cli.Argument{
							inputFileArg,
							outputFileArg,
						},
					},
					{
						Name:   "view",
						Usage:  "view",
						Action: doAisView,
						Arguments: []cli.Argument{
							inputFileArg,
						},
						Flags: []cli.Flag{
							listenAddrFlag,
							playbackSpeedFlag,
							playbackUpdatePeriodFlag,
						},
					},
				},
			},
		},
		DefaultCommand: "log",
		Flags: []cli.Flag{
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
