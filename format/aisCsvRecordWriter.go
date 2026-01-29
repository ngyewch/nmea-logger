package format

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/BertoldVdb/go-ais"
)

type CsvAISRecordWriter struct {
	csvWriter         *csv.Writer
	shipStaticDataMap map[uint32]*ais.ShipStaticData
}

func NewCsvAISRecordWriter(w io.Writer) (*CsvAISRecordWriter, error) {
	csvWriter := csv.NewWriter(w)
	err := csvWriter.Write([]string{
		"DATE TIME (UTC)",
		"EPOCH TIME",
		"MMSI",
		"LATITUDE",
		"LONGITUDE",
		"COURSE",
		"SPEED",
		"HEADING",
		"NAVSTAT",
		"IMO",
		"NAME",
		"CALLSIGN",
		"AISTYPE",
		"A",
		"B",
		"C",
		"D",
		"DRAUGHT",
		"DESTINATION",
		"ETA",
	})
	if err != nil {
		return nil, err
	}
	return &CsvAISRecordWriter{
		csvWriter:         csvWriter,
		shipStaticDataMap: make(map[uint32]*ais.ShipStaticData),
	}, nil
}

func (writer *CsvAISRecordWriter) Close() error {
	writer.csvWriter.Flush()
	return nil
}

func (writer *CsvAISRecordWriter) PreprocessRecord(record *AISRecord) error {
	switch report := record.AIS.Packet.(type) {
	case ais.ShipStaticData:
		_, ok := writer.shipStaticDataMap[report.UserID]
		if !ok {
			writer.shipStaticDataMap[report.UserID] = &report
		}
	}
	return nil
}

func (writer *CsvAISRecordWriter) WriteAISRecord(record *AISRecord) error {
	const dateTimeLayout = "2006-01-02 15:04:05"
	switch report := record.AIS.Packet.(type) {
	case ais.PositionReport:
		t := time.UnixMilli(record.Timestamp)
		var cells []string
		cells = append(cells,
			t.Format(dateTimeLayout),
			strconv.FormatInt(record.Timestamp/1000, 10),
			strconv.FormatInt(int64(report.UserID), 10),
			strconv.FormatFloat(roundToDecimalPoints(float64(report.Latitude), 5), 'f', -1, 64),
			strconv.FormatFloat(roundToDecimalPoints(float64(report.Longitude), 5), 'f', -1, 64),
			strconv.FormatFloat(float64(report.Cog), 'f', -1, 64),
			strconv.FormatFloat(float64(report.Sog), 'f', -1, 64),
			strconv.FormatInt(int64(report.TrueHeading), 10),
			strconv.FormatInt(int64(report.NavigationalStatus), 10),
		)
		shipStaticData := writer.shipStaticDataMap[report.UserID]
		if shipStaticData != nil {
			cells = append(cells,
				strconv.FormatInt(int64(shipStaticData.ImoNumber), 10),
				shipStaticData.Name,
				shipStaticData.CallSign,
				strconv.FormatInt(int64(shipStaticData.Type), 10),
				strconv.FormatInt(int64(shipStaticData.Dimension.A), 10),
				strconv.FormatInt(int64(shipStaticData.Dimension.B), 10),
				strconv.FormatInt(int64(shipStaticData.Dimension.C), 10),
				strconv.FormatInt(int64(shipStaticData.Dimension.D), 10),
				strconv.FormatFloat(float64(shipStaticData.MaximumStaticDraught), 'f', -1, 64),
				shipStaticData.Destination,
				fmt.Sprintf("%02d-%02d %02d:%02d", shipStaticData.Eta.Month, shipStaticData.Eta.Day, shipStaticData.Eta.Hour, shipStaticData.Eta.Minute),
			)
		}

		err := writer.csvWriter.Write(cells)
		if err != nil {
			return err
		}

	case ais.ShipStaticData:
		writer.shipStaticDataMap[report.UserID] = &report
	}
	return nil
}

func roundToDecimalPoints(v float64, decimalPoints int) float64 {
	multiplier := math.Pow10(decimalPoints)
	return math.Round(v*multiplier) / multiplier
}
