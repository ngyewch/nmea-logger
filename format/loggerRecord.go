package format

type LoggerRecord struct {
	Timestamp int64  `json:"timestamp"`
	NMEA      string `json:"nmea"`
}
