package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Reading struct {
	Name    string  `json:"name"`
	Value   float64 `json:"value"`
	Unit    string  `json:"unit,omitempty"`
	Quality string  `json:"quality,omitempty"`
}

type Message struct {
	ID         string            `json:"id"`
	Source     string            `json:"source"`
	Channel    string            `json:"channel"`
	Protocol   string            `json:"protocol"`
	DeviceNo   string            `json:"deviceNo"`
	AssetNo    string            `json:"assetNo,omitempty"`
	EventType  string            `json:"eventType"`
	EventTime  string            `json:"eventTime"`
	ReceivedAt string            `json:"receivedAt"`
	Readings   []Reading         `json:"readings"`
	Tags       map[string]string `json:"tags,omitempty"`
	RawHash    string            `json:"rawHash"`
	Raw        string            `json:"raw,omitempty"`
}

type ProtocolFrameEnvelope struct {
	Channel  string `json:"channel"`
	Protocol string `json:"protocol"`
	Raw      string `json:"raw"`
}

func newMessage(raw, source, channel, protocol string, now time.Time) Message {
	sum := sha256.Sum256([]byte(raw))
	id := hex.EncodeToString(sum[:12])
	return Message{
		ID:         fmt.Sprintf("icg-%s", id),
		Source:     source,
		Channel:    channel,
		Protocol:   protocol,
		EventType:  "telemetry",
		EventTime:  now.UTC().Format(time.RFC3339),
		ReceivedAt: now.UTC().Format(time.RFC3339),
		Tags:       map[string]string{},
		RawHash:    hex.EncodeToString(sum[:]),
		Raw:        raw,
	}
}

func (m Message) withoutRaw() Message {
	m.Raw = ""
	return m
}
