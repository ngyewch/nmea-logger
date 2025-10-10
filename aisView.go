package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/BertoldVdb/go-ais"
	"github.com/BertoldVdb/go-ais/aisnmea"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/ngyewch/nmea-logger/resources"
	"github.com/urfave/cli/v3"
)

func doAisView(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("insufficient arguments")
	}

	listenAddr := cmd.String(listenAddrFlag.Name)
	timeDilation := cmd.Float32(timeDilationFlag.Name)

	uiFs, err := fs.Sub(resources.UIFs, "gen/ui")
	if err != nil {
		return err
	}
	templates, err := template.ParseFS(uiFs, "index.html")
	if err != nil {
		return err
	}

	tmpl := templates.Lookup("index.html")
	if tmpl == nil {
		return errors.New("template not found")
	}

	logFile := cmd.Args().First()
	_, err = os.Stat(logFile)
	if err != nil {
		return err
	}

	type PositionReportRecord struct {
		Type           string             `json:"type"`
		T              int64              `json:"t"`
		PositionReport ais.PositionReport `json:"positionReport"`
	}

	type ShipStaticDataRecord struct {
		Type           string             `json:"type"`
		T              int64              `json:"t"`
		ShipStaticData ais.ShipStaticData `json:"shipStaticData"`
	}

	httpUIFs := http.FileServer(http.FS(uiFs))
	http.Handle("/assets/", httpUIFs)

	serveIndex := func(w http.ResponseWriter, r *http.Request) {
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{
				"localhost:8080",
			},
		})
		if err != nil {
			slog.Warn("error accepting websocket",
				slog.Any("err", err),
			)
			return
		}
		defer func(c *websocket.Conn) {
			err = c.CloseNow()
			if err != nil {
				slog.Warn("error closing websocket",
					slog.Any("err", err),
				)
			}
		}(c)

		f, err := os.Open(logFile)
		if err != nil {
			slog.Warn("error opening log file",
				slog.Any("err", err),
			)
			return
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		aisCodec := ais.CodecNew(false, false)
		aisCodec.DropSpace = true
		nmeaCodec := aisnmea.NMEACodecNew(aisCodec)

		scanner := bufio.NewScanner(f)
		var t int64 = 0
		for scanner.Scan() {
			logLineBytes := scanner.Bytes()

			jsonDecoder := json.NewDecoder(bytes.NewReader(logLineBytes))
			jsonDecoder.DisallowUnknownFields()

			var record Record
			err = jsonDecoder.Decode(&record)
			if err != nil {
				slog.Warn("error reading log line",
					slog.Any("err", err),
				)
				continue
			}

			decoded, err := nmeaCodec.ParseSentence(record.NMEA)
			if err != nil {
				slog.Warn("error parsing NMEA sentence",
					slog.Any("err", err),
				)
				continue
			}

			if decoded != nil {
				if (t != 0) && (record.Timestamp > t) {
					dt := float32(record.Timestamp-t) / timeDilation
					<-time.After(time.Duration(dt) * time.Millisecond)
					t = record.Timestamp
				}

				closed := false
				switch packet := decoded.Packet.(type) {
				case ais.PositionReport:
					record := PositionReportRecord{
						Type:           "positionReport",
						T:              record.Timestamp,
						PositionReport: packet,
					}
					err = wsjson.Write(context.Background(), c, record)
					if err != nil {
						var closeError *websocket.CloseError
						if errors.Is(err, closeError) {
							closed = true
						}
						slog.Warn("error writing message",
							slog.Any("err", err),
						)
					}
				case ais.ShipStaticData:
					record := ShipStaticDataRecord{
						Type:           "shipStaticData",
						T:              record.Timestamp,
						ShipStaticData: packet,
					}
					err = wsjson.Write(context.Background(), c, record)
					if err != nil {
						var closeError *websocket.CloseError
						if errors.Is(err, closeError) {
							closed = true
						}
						slog.Warn("error writing message",
							slog.Any("err", err),
						)
					}
				}

				if closed {
					break
				}
			}
		}

		err = c.Close(websocket.StatusNormalClosure, "")
		if err != nil {
			slog.Warn("error closing websocket",
				slog.Any("err", err),
			)
		}
	})
	http.HandleFunc("/index.html", serveIndex)
	http.HandleFunc("/", serveIndex)

	fmt.Printf("Listening on %s\n", listenAddr)
	return http.ListenAndServe(listenAddr, nil)
}
