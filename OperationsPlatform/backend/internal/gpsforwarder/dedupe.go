package forwarder

import (
	"fmt"
	"sync"
	"time"
)

type Deduper struct {
	window time.Duration
	mu     sync.Mutex
	seen   map[string]time.Time
}

func NewDeduper(window time.Duration) *Deduper {
	return &Deduper{window: window, seen: map[string]time.Time{}}
}

func (d *Deduper) Seen(loc Location, now time.Time) bool {
	if d.window <= 0 {
		return false
	}
	key := fmt.Sprintf("%s|%s|%.6f|%.6f", loc.DeviceNo, loc.LocationTime, loc.Longitude, loc.Latitude)
	d.mu.Lock()
	defer d.mu.Unlock()
	for existing, ts := range d.seen {
		if now.Sub(ts) > d.window {
			delete(d.seen, existing)
		}
	}
	if _, ok := d.seen[key]; ok {
		return true
	}
	d.seen[key] = now
	return false
}
