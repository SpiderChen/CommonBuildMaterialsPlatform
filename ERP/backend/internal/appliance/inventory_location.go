package appliance

type inventoryLotConsumption struct {
	InventoryItemID int64
	RawReceiptID    int64
	SiteID          int64
	MaterialID      int64
	SupplierID      int64
	BatchNo         string
	Warehouse       string
	Silo            string
	Quantity        float64
	Unit            string
}

func inventoryStorageLocation(data AppData, siteID, materialID int64, quantity float64) (string, string) {
	for _, item := range data.Inventory {
		if item.SiteID == siteID && item.MaterialID == materialID && siloCanAccept(data, siteID, item.Silo, quantity) {
			return fallback(item.Warehouse, "自动入库仓"), fallback(item.Silo, "AUTO")
		}
	}
	for _, warehouse := range data.Warehouses {
		if warehouse.SiteID != siteID || warehouse.Status != "active" {
			continue
		}
		for _, silo := range data.Silos {
			if silo.WarehouseID == warehouse.ID && silo.MaterialID == materialID && silo.Status != "blocked" && siloCanAccept(data, siteID, silo.Code, quantity) {
				return warehouse.Name, silo.Code
			}
		}
	}
	for _, warehouse := range data.Warehouses {
		if warehouse.SiteID == siteID && warehouse.Status == "active" {
			return warehouse.Name, "AUTO"
		}
	}
	return "自动入库仓", "AUTO"
}

func siloCanAccept(data AppData, siteID int64, siloCode string, quantity float64) bool {
	if quantity <= 0 || siloCode == "" || siloCode == "AUTO" {
		return true
	}
	silo, ok := siloByCodeForSite(data, siteID, siloCode)
	if !ok || silo.Capacity <= 0 {
		return true
	}
	return round(silo.CurrentQty+quantity) <= silo.Capacity
}

func adjustSiloCurrentQty(data *AppData, siteID int64, siloCode string, delta float64) {
	if delta == 0 || siloCode == "" || siloCode == "AUTO" {
		return
	}
	for i := range data.Silos {
		if !siloBelongsToSite(*data, data.Silos[i], siteID) {
			continue
		}
		if data.Silos[i].Code != siloCode && data.Silos[i].Name != siloCode {
			continue
		}
		data.Silos[i].CurrentQty = round(data.Silos[i].CurrentQty + delta)
		if data.Silos[i].CurrentQty < 0 {
			data.Silos[i].CurrentQty = 0
		}
		if data.Silos[i].Capacity > 0 && data.Silos[i].CurrentQty >= data.Silos[i].Capacity {
			data.Silos[i].Status = "warning"
		}
		return
	}
}

func siloByCodeForSite(data AppData, siteID int64, siloCode string) (Silo, bool) {
	for _, silo := range data.Silos {
		if !siloBelongsToSite(data, silo, siteID) {
			continue
		}
		if silo.Code == siloCode || silo.Name == siloCode {
			return silo, true
		}
	}
	return Silo{}, false
}

func siloBelongsToSite(data AppData, silo Silo, siteID int64) bool {
	for _, warehouse := range data.Warehouses {
		if warehouse.ID == silo.WarehouseID {
			return warehouse.SiteID == siteID
		}
	}
	return false
}
