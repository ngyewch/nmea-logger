package format

import (
	"github.com/BertoldVdb/go-ais"
	"github.com/BertoldVdb/go-ais/aisnmea"
)

type AISRecordReader struct {
	loggerRecordReader *LoggerRecordReader
	ignoreParseErrors  bool
	aisCodec           *ais.Codec
	nmeaCodec          *aisnmea.NMEACodec
}

func NewAISRecordReader(loggerRecordReader *LoggerRecordReader, ignoreParseErrors bool) *AISRecordReader {
	aisCodec := ais.CodecNew(false, false)
	aisCodec.DropSpace = true
	nmeaCodec := aisnmea.NMEACodecNew(aisCodec)
	return &AISRecordReader{
		loggerRecordReader: loggerRecordReader,
		ignoreParseErrors:  ignoreParseErrors,
		aisCodec:           aisCodec,
		nmeaCodec:          nmeaCodec,
	}
}

func (reader *AISRecordReader) ReadAISRecord() (*AISRecord, error) {
	for {
		loggerRecord, err := reader.loggerRecordReader.ReadLoggerRecord()
		if err != nil {
			return nil, err
		}
		if loggerRecord == nil {
			return nil, nil
		}

		decoded, err := reader.nmeaCodec.ParseSentence(loggerRecord.NMEA)
		if err != nil {
			if !reader.ignoreParseErrors {
				return nil, err
			}
		}

		if decoded != nil {
			aisRecord := &AISRecord{
				Timestamp: loggerRecord.Timestamp,
				AIS:       decoded,
			}
			return aisRecord, nil
		}
	}
}
