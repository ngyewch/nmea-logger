package format

import "github.com/BertoldVdb/go-ais/aisnmea"

type AISRecord struct {
	Timestamp int64              `json:"timestamp"`
	AIS       *aisnmea.VdmPacket `json:"ais"`
}
