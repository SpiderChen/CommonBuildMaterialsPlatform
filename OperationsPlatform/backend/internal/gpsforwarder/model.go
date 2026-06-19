package forwarder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Location struct {
	ID           string  `json:"id"`
	Source       string  `json:"source"`
	Channel      string  `json:"channel"`
	Protocol     string  `json:"protocol"`
	DeviceNo     string  `json:"deviceNo"`
	PlateNo      string  `json:"plateNo,omitempty"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	Speed        float64 `json:"speed"`
	Direction    float64 `json:"direction"`
	Mileage      float64 `json:"mileage,omitempty"`
	AccStatus    int     `json:"accStatus"`
	LocationTime string  `json:"locationTime"`
	ReceivedAt   string  `json:"receivedAt"`
	RawHash      string  `json:"rawHash"`
	Raw          string  `json:"raw,omitempty"`
}

type ProtocolFrameEnvelope struct {
	Channel  string `json:"channel"`
	Protocol string `json:"protocol"`
	Raw      string `json:"raw"`
}

func newLocation(raw, source, channel, protocol string, now time.Time) Location {
	sum := sha256.Sum256([]byte(raw))
	id := hex.EncodeToString(sum[:12])
	return Location{
		ID:           fmt.Sprintf("gps-%s", id),
		Source:       source,
		Channel:      channel,
		Protocol:     protocol,
		LocationTime: now.UTC().Format(time.RFC3339),
		ReceivedAt:   now.UTC().Format(time.RFC3339),
		RawHash:      hex.EncodeToString(sum[:]),
		Raw:          raw,
	}
}

func (l Location) withoutRaw() Location {
	l.Raw = ""
	return l
}
