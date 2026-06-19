package appliance

import "fmt"

func appendRawMaterialTicket(data *AppData, receipt RawMaterialReceipt) ScaleTicket {
	device := defaultScaleDeviceForSite(*data, receipt.SiteID)
	plateNo := fallback(receipt.PlateNo, fmt.Sprintf("原料车-%s", receipt.ReceiptNo))
	createdAt := fallback(receipt.CreatedAt, nowString())
	ticketID := nextID(data, "ticket")
	ticket := ScaleTicket{
		ID: ticketID, TicketNo: number("ST", ticketID), TicketType: "raw_material_in",
		SiteID: receipt.SiteID, PlateNo: plateNo, GrossWeight: receipt.GrossWeight, TareWeight: receipt.TareWeight,
		NetWeight: receipt.NetWeight, Unit: "t", SnapshotURL: rawMaterialCaptureURL(device.Code, receipt.ReceiptNo, "gross"),
		PrintCount: 1, SignStatus: "not_required", SettlementStatus: "pending", Status: "valid", CreatedAt: createdAt,
		ReceiptID: receipt.ID, SupplierID: receipt.SupplierID, MaterialID: receipt.MaterialID,
	}
	data.ScaleTickets = append(data.ScaleTickets, ticket)
	appendWeightRecord(data, device.ID, ticket.ID, plateNo, receipt.TareWeight, "tare", rawMaterialCaptureURL(device.Code, receipt.ReceiptNo, "tare"), createdAt)
	appendWeightRecord(data, device.ID, ticket.ID, plateNo, receipt.GrossWeight, "gross", ticket.SnapshotURL, createdAt)
	return ticket
}

func appendWeightRecord(data *AppData, deviceID, ticketID int64, plateNo string, weight float64, weightType, snapshotURL, createdAt string) {
	data.ScaleWeightRecords = append(data.ScaleWeightRecords, ScaleWeightRecord{
		ID: nextID(data, "weightRecord"), DeviceID: deviceID, TicketID: ticketID,
		PlateNo: plateNo, Weight: weight, WeightType: weightType, SnapshotURL: snapshotURL, CreatedAt: createdAt,
	})
}

func defaultScaleDeviceForSite(data AppData, siteID int64) ScaleDevice {
	for _, device := range data.ScaleDevices {
		if device.SiteID == siteID && device.Status == "online" {
			return device
		}
	}
	for _, device := range data.ScaleDevices {
		if device.SiteID == siteID {
			return device
		}
	}
	if len(data.ScaleDevices) > 0 {
		return data.ScaleDevices[0]
	}
	return ScaleDevice{ID: 0, SiteID: siteID, Code: "manual-scale", Name: "人工地磅", Status: "online"}
}

func rawMaterialCaptureURL(deviceCode, receiptNo, weightType string) string {
	return fmt.Sprintf("capture://%s/raw-%s-%s.jpg", fallback(deviceCode, "manual-scale"), receiptNo, weightType)
}

func scaleCaptureURL(deviceCode, sourceNo, weightType string) string {
	return fmt.Sprintf("capture://%s/%s-%s.jpg", fallback(deviceCode, "manual-scale"), sourceNo, weightType)
}

func backfillRawMaterialTickets(data *AppData) bool {
	changed := false
	for i := range data.RawMaterialReceipts {
		if data.RawMaterialReceipts[i].TicketID != 0 && hasScaleTicket(*data, data.RawMaterialReceipts[i].TicketID) {
			continue
		}
		ticket := appendRawMaterialTicket(data, data.RawMaterialReceipts[i])
		data.RawMaterialReceipts[i].TicketID = ticket.ID
		if data.RawMaterialReceipts[i].PlateNo == "" {
			data.RawMaterialReceipts[i].PlateNo = ticket.PlateNo
		}
		changed = true
	}
	return changed
}

func hasScaleTicket(data AppData, id int64) bool {
	for _, ticket := range data.ScaleTickets {
		if ticket.ID == id {
			return true
		}
	}
	return false
}
