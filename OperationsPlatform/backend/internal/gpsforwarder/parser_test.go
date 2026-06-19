package forwarder

import (
	"testing"
	"time"
)

func TestParseJSONLocation(t *testing.T) {
	raw := `{"deviceNo":"GPS1000002","plateNo":"粤B22336","longitude":113.9412,"latitude":22.5428,"speed":38.5,"direction":96,"mileage":2001.5,"accStatus":1,"locationTime":"2026-06-18 12:05:00"}`
	loc, err := ParseLocation(raw, "test", "http", "gps-json", time.Now())
	if err != nil {
		t.Fatalf("ParseLocation returned error: %v", err)
	}
	if loc.DeviceNo != "GPS1000002" || loc.PlateNo != "粤B22336" || loc.Longitude != 113.9412 {
		t.Fatalf("unexpected location: %+v", loc)
	}
}

func TestParseCSVLocation(t *testing.T) {
	raw := "GPS,GPS1000002,粤B22336,113.9412,22.5428,38.5,96,2001.5,1,2026-06-18 12:05:00"
	loc, err := ParseLocation(raw, "test", "tcp", "gps-csv", time.Now())
	if err != nil {
		t.Fatalf("ParseLocation returned error: %v", err)
	}
	if loc.Protocol != "gps-csv" || loc.AccStatus != 1 || loc.Mileage != 2001.5 {
		t.Fatalf("unexpected location: %+v", loc)
	}
}

func TestParseRejectsBadCoordinate(t *testing.T) {
	_, err := ParseLocation(`{"deviceNo":"GPS1","longitude":999,"latitude":22.5}`, "test", "http", "gps-json", time.Now())
	if err == nil {
		t.Fatal("expected invalid coordinate error")
	}
}
