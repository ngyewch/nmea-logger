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

	type PositionReportEntry struct {
		T              int64              `json:"t"`
		PositionReport ais.PositionReport `json:"positionReport"`
	}

	type TemplateData struct {
		ShipStaticDataMap  map[uint32]ais.ShipStaticData    `json:"shipStaticDataMap"`
		PositionReportsMap map[uint32][]PositionReportEntry `json:"positionReportsMap"`
	}

	templateData := &TemplateData{
		ShipStaticDataMap:  make(map[uint32]ais.ShipStaticData),
		PositionReportsMap: make(map[uint32][]PositionReportEntry),
	}

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

		decoded, _ := nmeaCodec.ParseSentence(record.NMEA)
		if decoded != nil {
			switch packet := decoded.Packet.(type) {
			case ais.PositionReport:
				positionReports := templateData.PositionReportsMap[packet.UserID]
				positionReports = append(positionReports, PositionReportEntry{
					T:              record.Timestamp,
					PositionReport: packet,
				})
				templateData.PositionReportsMap[packet.UserID] = positionReports
			case ais.ShipStaticData:
				templateData.ShipStaticDataMap[packet.UserID] = packet
			}
		}
	}

	httpUIFs := http.FileServer(http.FS(uiFs))
	http.Handle("/assets/", httpUIFs)

	serveIndex := func(w http.ResponseWriter, r *http.Request) {
		err = tmpl.Execute(w, templateData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
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

		// Set the context as needed. Use of r.Context() is not recommended
		// to avoid surprising behavior (see http.Hijacker).
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		var v any
		err = wsjson.Read(ctx, c, &v)
		if err != nil {
			// ...
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

type AisDataPlayback struct {
}
