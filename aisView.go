package main

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BertoldVdb/go-ais"
	"github.com/BertoldVdb/go-ais/aisnmea"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/ngyewch/nmea-logger/resources"
	"github.com/ulikunitz/xz"
	"github.com/urfave/cli/v3"
)

type PlaybackRecord interface {
	GetTimestamp() int64
}

type PositionReportRecord struct {
	Type           string             `json:"type"`
	T              int64              `json:"t"`
	PositionReport ais.PositionReport `json:"positionReport"`
}

func (record *PositionReportRecord) GetTimestamp() int64 {
	return record.T
}

type ShipStaticDataRecord struct {
	Type           string             `json:"type"`
	T              int64              `json:"t"`
	ShipStaticData ais.ShipStaticData `json:"shipStaticData"`
}

func (record *ShipStaticDataRecord) GetTimestamp() int64 {
	return record.T
}

func doAisView(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return fmt.Errorf("insufficient arguments")
	}

	listenAddr := cmd.String(listenAddrFlag.Name)
	playbackSpeed := cmd.Float64(playbackSpeedFlag.Name)
	playbackUpdatePeriod := cmd.Duration(playbackUpdatePeriodFlag.Name)
	collectionPeriodInMs := (time.Duration(playbackUpdatePeriod.Seconds()*playbackSpeed) * time.Second).Milliseconds()

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

		var reader io.Reader
		reader = f
		if strings.HasSuffix(logFile, ".gz") {
			gzipReader, err := gzip.NewReader(reader)
			if err != nil {
				slog.Warn("error opening log file",
					slog.Any("err", err),
				)
				return
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
				slog.Warn("error opening log file",
					slog.Any("err", err),
				)
				return
			}
			reader = xzReader
		}

		aisCodec := ais.CodecNew(false, false)
		aisCodec.DropSpace = true
		nmeaCodec := aisnmea.NMEACodecNew(aisCodec)

		scanner := bufio.NewScanner(reader)
		var records []PlaybackRecord
		for scanner.Scan() {
			logLineBytes := scanner.Bytes()

			jsonDecoder := json.NewDecoder(bytes.NewReader(logLineBytes))
			jsonDecoder.DisallowUnknownFields()

			var record LoggerRecord
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
				if len(records) > 0 {
					tFirstInBatch := records[0].GetTimestamp()
					tCurrent := record.Timestamp
					dt := tCurrent - tFirstInBatch
					if dt > collectionPeriodInMs {
						err = wsjson.Write(context.Background(), c, records)
						if err != nil {
							slog.Warn("error writing playback records",
								slog.Any("err", err),
							)
							break
						}
						records = nil
						sleepDuration := time.Duration(float64(dt)*1000/playbackSpeed) * time.Microsecond
						<-time.After(sleepDuration)
					}
				}

				switch packet := decoded.Packet.(type) {
				case ais.PositionReport:
					record := PositionReportRecord{
						Type:           "positionReport",
						T:              record.Timestamp,
						PositionReport: packet,
					}
					records = append(records, &record)
				case ais.ShipStaticData:
					record := ShipStaticDataRecord{
						Type:           "shipStaticData",
						T:              record.Timestamp,
						ShipStaticData: packet,
					}
					records = append(records, &record)
				}
			}
		}
		if len(records) > 0 {
			err = wsjson.Write(context.Background(), c, records)
			if err != nil {
				slog.Warn("error writing playback records",
					slog.Any("err", err),
				)
			}
			records = nil
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

	httpListener, err := net.Listen("tcp4", listenAddr)
	defer func(httpListener net.Listener) {
		_ = httpListener.Close()
	}(httpListener)
	if err != nil {
		return err
	}
	fmt.Printf("URL: http://%s\n", httpListener.Addr().String())
	return http.Serve(httpListener, nil)
}
