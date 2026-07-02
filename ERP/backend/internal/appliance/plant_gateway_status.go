package appliance

import "strings"

func plantsWithGatewayStatus(data AppData) []Plant {
	batchesByID := make(map[int64]ProductionBatch, len(data.ProductionBatches))
	for _, batch := range data.ProductionBatches {
		batchesByID[batch.ID] = batch
	}

	plants := make([]Plant, len(data.Plants))
	for i, plant := range data.Plants {
		plants[i] = plantWithGatewayStatus(data, plant, batchesByID)
	}
	return plants
}

func plantWithGatewayStatus(data AppData, plant Plant, batchesByID map[int64]ProductionBatch) Plant {
	plant.Interface = ""

	for i := len(data.DeviceProtocolFrames) - 1; i >= 0; i-- {
		frame := data.DeviceProtocolFrames[i]
		if !plantMatchesProtocolFrame(plant, frame, batchesByID) {
			continue
		}
		plant.GatewayStatus = plantGatewayStatusFromFrameStatus(frame.Status)
		plant.GatewayDeviceNo = frame.DeviceNo
		plant.GatewayChannel = frame.Channel
		plant.GatewayProtocol = frame.Protocol
		plant.LastFrameAt = frame.ReceivedAt
		plant.LastFrameStatus = frame.Status
		plant.GatewayError = frame.Error
		return plant
	}

	for _, device := range data.VehicleDevices {
		if !plantDeviceNoMatches(plant.Code, device.DeviceNo) {
			continue
		}
		plant.GatewayStatus = plantGatewayStatusFromDeviceStatus(device.Status)
		plant.GatewayDeviceNo = device.DeviceNo
		plant.GatewayChannel = "industrial-control-gateway"
		plant.GatewayProtocol = device.Protocol
		plant.LastFrameAt = device.LastSeenAt
		plant.LastFrameStatus = device.Status
		return plant
	}

	for _, credential := range data.DeviceCredentials {
		if !plantDeviceNoMatches(plant.Code, credential.DeviceNo) {
			continue
		}
		plant.GatewayStatus = "registered"
		plant.GatewayDeviceNo = credential.DeviceNo
		plant.GatewayChannel = "industrial-control-gateway"
		plant.LastFrameAt = credential.LastUsedAt
		plant.LastFrameStatus = credential.Status
		return plant
	}

	plant.GatewayStatus = "not_connected"
	return plant
}

func plantMatchesProtocolFrame(plant Plant, frame DeviceProtocolFrame, batchesByID map[int64]ProductionBatch) bool {
	code := strings.TrimSpace(plant.Code)
	if code == "" {
		return false
	}
	if frame.ParsedResource == "production_batch" && frame.ParsedID > 0 {
		if batch, ok := batchesByID[frame.ParsedID]; ok && strings.EqualFold(strings.TrimSpace(batch.PlantCode), code) {
			return true
		}
	}
	if plantDeviceNoMatches(code, frame.DeviceNo) {
		return true
	}
	return strings.Contains(strings.ToLower(frame.Raw), strings.ToLower(code))
}

func plantDeviceNoMatches(code string, deviceNo string) bool {
	code = strings.ToLower(strings.TrimSpace(code))
	deviceNo = strings.ToLower(strings.TrimSpace(deviceNo))
	if code == "" || deviceNo == "" {
		return false
	}
	return deviceNo == code || deviceNo == "plant-"+code || strings.HasSuffix(deviceNo, "-"+code) || strings.Contains(deviceNo, code)
}

func plantGatewayStatusFromFrameStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "accepted", "online":
		return "online"
	case "rejected", "failed", "error":
		return "error"
	case "":
		return "unknown"
	default:
		return strings.TrimSpace(status)
	}
}

func plantGatewayStatusFromDeviceStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "online":
		return "online"
	case "offline", "":
		return "offline"
	default:
		return strings.TrimSpace(status)
	}
}
