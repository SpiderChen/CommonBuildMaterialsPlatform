package appliance

import (
	"errors"
	"math"
	"sort"
	"time"
)

const defaultTrackCompressionToleranceMeters = 35

func buildTrackReplay(data AppData, vehicleID int64, startTime, endTime string) TrackReplay {
	points := []VehicleLocationEvent{}
	for _, loc := range data.Locations {
		if vehicleID != 0 && loc.VehicleID != vehicleID {
			continue
		}
		if startTime != "" && loc.LocationTime < startTime {
			continue
		}
		if endTime != "" && loc.LocationTime > endTime {
			continue
		}
		points = append(points, loc)
	}
	sort.Slice(points, func(i, j int) bool {
		if points[i].LocationTime == points[j].LocationTime {
			return points[i].ID < points[j].ID
		}
		return points[i].LocationTime < points[j].LocationTime
	})

	replay := TrackReplay{
		VehicleID:        vehicleID,
		Points:           points,
		CompressedPoints: []VehicleLocationEvent{},
		Compression:      trackCompressionSummary(points, []VehicleLocationEvent{}, []TrackStopPoint{}, defaultTrackCompressionToleranceMeters),
		Stops:            []TrackStopPoint{},
		FenceEvents:      []GeoFenceEvent{},
		Tickets:          []ScaleTicket{},
		Signs:            []DeliverySign{},
	}
	if len(points) == 0 {
		if vehicle, ok := findVehicle(data, vehicleID); ok {
			replay.PlateNo = vehicle.PlateNo
		}
		return replay
	}
	replay.VehicleID = points[0].VehicleID
	replay.PlateNo = points[0].PlateNo
	replay.StartTime = points[0].LocationTime
	replay.EndTime = points[len(points)-1].LocationTime
	replay.ExportName = "track-" + replay.PlateNo + "-" + compactTime(replay.StartTime) + ".json"

	var distanceMetersTotal float64
	var speedTotal float64
	for i, point := range points {
		if point.Speed > replay.MaxSpeed {
			replay.MaxSpeed = point.Speed
		}
		speedTotal += point.Speed
		if i > 0 {
			prev := points[i-1]
			distanceMetersTotal += distanceMeters(prev.Latitude, prev.Longitude, point.Latitude, point.Longitude)
		}
	}
	replay.DistanceKm = round(distanceMetersTotal / 1000)
	replay.AverageSpeed = round(speedTotal / float64(len(points)))
	replay.DurationMinutes = round(trackDurationMinutes(replay.StartTime, replay.EndTime))
	replay.Stops = detectStops(points)
	replay.StopCount = len(replay.Stops)
	replay.CompressedPoints = compressTrackPoints(points, replay.Stops, defaultTrackCompressionToleranceMeters)
	replay.Compression = trackCompressionSummary(points, replay.CompressedPoints, replay.Stops, defaultTrackCompressionToleranceMeters)

	for _, event := range data.GeoFenceEvents {
		if event.VehicleID == replay.VehicleID && betweenTime(event.EventTime, replay.StartTime, replay.EndTime) {
			replay.FenceEvents = append(replay.FenceEvents, event)
		}
	}
	for _, ticket := range data.ScaleTickets {
		if ticket.VehicleID == replay.VehicleID && betweenTime(ticket.CreatedAt, replay.StartTime, replay.EndTime) {
			replay.Tickets = append(replay.Tickets, ticket)
		}
	}
	for _, sign := range data.DeliverySigns {
		dispatch, ok := findDispatch(data, sign.DispatchID)
		if ok && dispatch.VehicleID == replay.VehicleID && betweenTime(sign.SignedAt, replay.StartTime, replay.EndTime) {
			replay.Signs = append(replay.Signs, sign)
		}
	}
	return replay
}

func compressTrackPoints(points []VehicleLocationEvent, stops []TrackStopPoint, toleranceMeters float64) []VehicleLocationEvent {
	if len(points) <= 2 {
		return append([]VehicleLocationEvent{}, points...)
	}
	if toleranceMeters <= 0 {
		toleranceMeters = defaultTrackCompressionToleranceMeters
	}
	keep := make([]bool, len(points))
	protected := protectedTrackIndexes(points, stops)
	for index := range protected {
		keep[index] = true
	}
	markRDPTrackPoints(points, 0, len(points)-1, toleranceMeters, keep)

	compressed := []VehicleLocationEvent{}
	for index, point := range points {
		if keep[index] {
			compressed = append(compressed, point)
		}
	}
	if len(compressed) == 0 {
		return append([]VehicleLocationEvent{}, points...)
	}
	return compressed
}

func protectedTrackIndexes(points []VehicleLocationEvent, stops []TrackStopPoint) map[int]bool {
	indexes := map[int]bool{}
	if len(points) == 0 {
		return indexes
	}
	indexes[0] = true
	indexes[len(points)-1] = true

	stopTimes := map[string]bool{}
	for _, stop := range stops {
		if stop.StartTime != "" {
			stopTimes[stop.StartTime] = true
		}
		if stop.EndTime != "" {
			stopTimes[stop.EndTime] = true
		}
	}
	for index, point := range points {
		if point.IsAbnormal || point.AbnormalType != "" || stopTimes[point.LocationTime] {
			indexes[index] = true
		}
	}
	return indexes
}

func markRDPTrackPoints(points []VehicleLocationEvent, start, end int, toleranceMeters float64, keep []bool) {
	if end <= start {
		return
	}
	keep[start] = true
	keep[end] = true
	if end-start < 2 {
		return
	}
	maxDistance := 0.0
	maxIndex := -1
	for index := start + 1; index < end; index++ {
		distance := perpendicularTrackDistanceMeters(points[index], points[start], points[end])
		if distance > maxDistance {
			maxDistance = distance
			maxIndex = index
		}
	}
	if maxIndex >= 0 && maxDistance > toleranceMeters {
		keep[maxIndex] = true
		markRDPTrackPoints(points, start, maxIndex, toleranceMeters, keep)
		markRDPTrackPoints(points, maxIndex, end, toleranceMeters, keep)
	}
}

func perpendicularTrackDistanceMeters(point, start, end VehicleLocationEvent) float64 {
	if start.Latitude == end.Latitude && start.Longitude == end.Longitude {
		return distanceMeters(point.Latitude, point.Longitude, start.Latitude, start.Longitude)
	}
	x0, y0 := projectedTrackPointMeters(point, start)
	x1, y1 := 0.0, 0.0
	x2, y2 := projectedTrackPointMeters(end, start)
	dx := x2 - x1
	dy := y2 - y1
	if dx == 0 && dy == 0 {
		return math.Hypot(x0-x1, y0-y1)
	}
	t := ((x0-x1)*dx + (y0-y1)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))
	projX := x1 + t*dx
	projY := y1 + t*dy
	return math.Hypot(x0-projX, y0-projY)
}

func projectedTrackPointMeters(point, origin VehicleLocationEvent) (float64, float64) {
	const latMeters = 110540.0
	lonMeters := 111320.0 * math.Cos(origin.Latitude*math.Pi/180)
	return (point.Longitude - origin.Longitude) * lonMeters, (point.Latitude - origin.Latitude) * latMeters
}

func trackCompressionSummary(points, compressed []VehicleLocationEvent, stops []TrackStopPoint, toleranceMeters float64) TrackCompressionSummary {
	rawCount := len(points)
	compressedCount := len(compressed)
	abnormalCount := 0
	for _, point := range points {
		if point.IsAbnormal || point.AbnormalType != "" {
			abnormalCount++
		}
	}
	ratio := 0.0
	reduction := 0.0
	if rawCount > 0 {
		ratio = round(float64(compressedCount) / float64(rawCount) * 100)
		reduction = round((1 - float64(compressedCount)/float64(rawCount)) * 100)
	}
	return TrackCompressionSummary{
		Algorithm:               "rdp",
		ToleranceMeters:         round(toleranceMeters),
		RawPointCount:           rawCount,
		CompressedPointCount:    compressedCount,
		CompressionRatio:        ratio,
		ReductionPercent:        reduction,
		PreservedStops:          len(stops),
		PreservedAbnormalPoints: abnormalCount,
	}
}

func detectStops(points []VehicleLocationEvent) []TrackStopPoint {
	stops := []TrackStopPoint{}
	var start *VehicleLocationEvent
	var last *VehicleLocationEvent
	for i := range points {
		point := points[i]
		if point.Speed <= 3 {
			if start == nil {
				start = &points[i]
			}
			last = &points[i]
			continue
		}
		if start != nil && last != nil {
			if minutes := trackDurationMinutes(start.LocationTime, last.LocationTime); minutes >= 5 {
				stops = append(stops, TrackStopPoint{
					Longitude:       start.Longitude,
					Latitude:        start.Latitude,
					StartTime:       start.LocationTime,
					EndTime:         last.LocationTime,
					DurationMinutes: round(minutes),
					Address:         start.Address,
				})
			}
		}
		start = nil
		last = nil
	}
	if start != nil && last != nil {
		if minutes := trackDurationMinutes(start.LocationTime, last.LocationTime); minutes >= 5 {
			stops = append(stops, TrackStopPoint{
				Longitude:       start.Longitude,
				Latitude:        start.Latitude,
				StartTime:       start.LocationTime,
				EndTime:         last.LocationTime,
				DurationMinutes: round(minutes),
				Address:         start.Address,
			})
		}
	}
	return stops
}

func betweenTime(value, start, end string) bool {
	if value == "" {
		return false
	}
	if start != "" && value < start {
		return false
	}
	if end != "" && value > end {
		return false
	}
	return true
}

func trackDurationMinutes(start, end string) float64 {
	startAt, err := parseTime(start)
	if err != nil {
		return 0
	}
	endAt, err := parseTime(end)
	if err != nil || endAt.Before(startAt) {
		return 0
	}
	return endAt.Sub(startAt).Minutes()
}

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("empty time")
	}
	if parsed, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, value)
}

func compactTime(value string) string {
	out := ""
	for _, r := range value {
		if r >= '0' && r <= '9' {
			out += string(r)
		}
	}
	if out == "" {
		return "unknown"
	}
	return out
}
