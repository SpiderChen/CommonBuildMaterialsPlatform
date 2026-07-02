package appliance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type RuntimeServices struct {
	redisAddr      string
	rabbitURL      string
	clickhouseHTTP string
	taxGateway     TaxGatewayConfig
	mapProvider    MapProviderConfig
	redis          *redis.Client
	rabbit         *amqp.Connection
	httpClient     *http.Client
}

type RuntimeStatus struct {
	Storage                string                `json:"storage"`
	BusinessTables         string                `json:"businessTables"`
	BusinessTableCount     int                   `json:"businessTableCount"`
	BusinessProjectionRows int                   `json:"businessProjectionRows"`
	DomainTables           string                `json:"domainTables"`
	DomainResourceCount    int                   `json:"domainResourceCount"`
	DomainRowCount         int                   `json:"domainRowCount"`
	RedisAddr              string                `json:"redisAddr"`
	Redis                  string                `json:"redis"`
	RabbitURL              string                `json:"rabbitUrl"`
	RabbitMQ               string                `json:"rabbitmq"`
	ClickHouseURL          string                `json:"clickhouseUrl"`
	ClickHouse             string                `json:"clickhouse"`
	EventBus               string                `json:"eventBus"`
	TaxGatewayProvider     string                `json:"taxGatewayProvider"`
	TaxGatewayURL          string                `json:"taxGatewayUrl"`
	TaxGateway             string                `json:"taxGateway"`
	MapProvider            string                `json:"mapProvider"`
	MapTiles               string                `json:"mapTiles"`
	MapTileURL             string                `json:"mapTileUrl"`
	MapCoordinateSystem    string                `json:"mapCoordinateSystem"`
	MapAPIKeyConfigured    bool                  `json:"mapApiKeyConfigured"`
	DeviceGateways         []DeviceGatewayStatus `json:"deviceGateways"`
}

func NewRuntimeServicesFromEnv() *RuntimeServices {
	services := &RuntimeServices{
		redisAddr:      os.Getenv("CBMP_REDIS_ADDR"),
		rabbitURL:      os.Getenv("CBMP_RABBITMQ_URL"),
		clickhouseHTTP: os.Getenv("CBMP_CLICKHOUSE_HTTP_URL"),
		taxGateway:     NewTaxGatewayConfigFromEnv(),
		mapProvider:    NewMapProviderConfigFromEnv(),
		httpClient:     &http.Client{Timeout: 2 * time.Second},
	}
	if services.redisAddr != "" {
		services.redis = redis.NewClient(&redis.Options{Addr: services.redisAddr})
	}
	if services.rabbitURL != "" {
		conn, err := amqp.DialConfig(services.rabbitURL, amqp.Config{Heartbeat: 10 * time.Second})
		if err == nil {
			services.rabbit = conn
		}
	}
	return services
}

func (r *RuntimeServices) Status() RuntimeStatus {
	status := RuntimeStatus{
		Storage:             "vault",
		BusinessTables:      "disabled",
		DomainTables:        "disabled",
		RedisAddr:           r.redisAddr,
		RabbitURL:           maskedURL(r.rabbitURL),
		ClickHouseURL:       maskedURL(r.clickhouseHTTP),
		Redis:               "disabled",
		RabbitMQ:            "disabled",
		ClickHouse:          "disabled",
		EventBus:            "local-sse",
		TaxGateway:          "unconfigured",
		MapProvider:         r.mapProvider.Provider,
		MapTiles:            r.mapProvider.TileStatus(),
		MapTileURL:          r.mapProvider.PublicRuntimeTileURL(),
		MapCoordinateSystem: r.mapProvider.CoordinateSystem,
		MapAPIKeyConfigured: r.mapProvider.APIKeyConfigured,
	}
	if r.taxGateway.Provider != "" {
		status.TaxGatewayProvider = r.taxGateway.Provider
	}
	if r.taxGateway.URL != "" {
		status.TaxGatewayURL = maskedURL(r.taxGateway.URL)
		status.TaxGateway = "configured"
	}
	if r.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		if err := r.redis.Ping(ctx).Err(); err != nil {
			status.Redis = "degraded"
		} else {
			status.Redis = "online"
		}
	}
	if r.rabbitURL != "" {
		if r.rabbit != nil && !r.rabbit.IsClosed() {
			status.RabbitMQ = "online"
			status.EventBus = "rabbitmq+sse"
		} else {
			status.RabbitMQ = "degraded"
		}
	}
	if r.clickhouseHTTP != "" {
		if r.clickHousePing() {
			status.ClickHouse = "online"
		} else {
			status.ClickHouse = "degraded"
		}
	}
	return status
}

func (r *RuntimeServices) MapConfig() MapProviderConfig {
	if r == nil {
		return NewMapProviderConfigFromEnv()
	}
	cfg := r.mapProvider
	cfg.Subdomains = append([]string(nil), cfg.Subdomains...)
	if cfg.Subdomains == nil {
		cfg.Subdomains = []string{}
	}
	return cfg
}

func (r *RuntimeServices) Publish(topic string, payload interface{}) {
	if r == nil {
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	if r.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_ = r.redis.XAdd(ctx, &redis.XAddArgs{
			Stream: "cbmp:events",
			MaxLen: 10000,
			Approx: true,
			Values: map[string]interface{}{"topic": topic, "payload": string(raw), "at": nowString()},
		}).Err()
		cancel()
	}
	if r.rabbit != nil && !r.rabbit.IsClosed() {
		ch, err := r.rabbit.Channel()
		if err != nil {
			return
		}
		defer ch.Close()
		_ = ch.ExchangeDeclare("cbmp.events", "topic", true, false, false, false, nil)
		_ = ch.PublishWithContext(context.Background(), "cbmp.events", topic, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        raw,
			Timestamp:   time.Now(),
		})
	}
}

func (r *RuntimeServices) CacheLatestLocation(latest VehicleLatestLocation) {
	if r == nil || r.redis == nil {
		return
	}
	raw, err := json.Marshal(latest)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = r.redis.HSet(ctx, "cbmp:vehicle:latest", latest.PlateNo, raw).Err()
}

func (r *RuntimeServices) StoreTrackPoint(event VehicleLocationEvent) {
	if r == nil || r.clickhouseHTTP == "" {
		return
	}
	row := map[string]interface{}{
		"id":            event.ID,
		"vehicle_id":    event.VehicleID,
		"plate_no":      event.PlateNo,
		"driver_id":     event.DriverID,
		"dispatch_id":   event.DispatchID,
		"device_id":     event.DeviceID,
		"source_type":   event.SourceType,
		"longitude":     event.Longitude,
		"latitude":      event.Latitude,
		"speed":         event.Speed,
		"direction":     event.Direction,
		"mileage":       event.Mileage,
		"online_status": event.OnlineStatus,
		"address":       event.Address,
		"is_abnormal":   event.IsAbnormal,
		"abnormal_type": event.AbnormalType,
		"location_time": event.LocationTime,
		"receive_time":  event.ReceiveTime,
	}
	raw, err := json.Marshal(row)
	if err != nil {
		return
	}
	query := `create table if not exists cbmp_vehicle_track_point (
		id Int64,
		vehicle_id Int64,
		plate_no String,
		driver_id Int64,
		dispatch_id Int64,
		device_id String,
		source_type String,
		longitude Float64,
		latitude Float64,
		speed Float64,
		direction Float64,
		mileage Float64,
		online_status String,
		address String,
		is_abnormal Bool,
		abnormal_type String,
		location_time String,
		receive_time String
	) engine = MergeTree order by (vehicle_id, receive_time, id);
	insert into cbmp_vehicle_track_point format JSONEachRow
`
	_, _ = r.clickhousePost(query + string(raw) + "\n")
}

func (r *RuntimeServices) clickHousePing() bool {
	if r == nil || r.clickhouseHTTP == "" {
		return false
	}
	body, err := r.clickhousePost("select 1")
	return err == nil && strings.TrimSpace(body) == "1"
}

func (r *RuntimeServices) clickhousePost(query string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, r.clickhouseHTTP, bytes.NewBufferString(query))
	if err != nil {
		return "", err
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	if resp.StatusCode >= 300 {
		return buf.String(), fmt.Errorf("clickhouse status %d: %s", resp.StatusCode, buf.String())
	}
	return buf.String(), nil
}

func maskedURL(value string) string {
	if value == "" {
		return ""
	}
	return "<configured>"
}
