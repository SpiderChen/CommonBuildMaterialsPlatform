package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

func (a *App) updateMasterResource(w http.ResponseWriter, r *http.Request, session Session, resource string, id int64) {
	switch resource {
	case "customers":
		var item Customer
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid customer")
			return
		}
		var updated Customer
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Customers {
				if data.Customers[i].ID != id {
					continue
				}
				current := data.Customers[i]
				if strings.TrimSpace(item.Name) == "" {
					return fmt.Errorf("客户名称不能为空")
				}
				item.ID = id
				item.CompanyID = nonZeroInt(item.CompanyID, current.CompanyID)
				item.Receivable = current.Receivable
				item.Status = fallback(item.Status, current.Status)
				data.Customers[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "customer", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("客户不存在")
		})
		a.respondUpdate(w, err, updated, "master.customer.updated")
	case "customer-contacts":
		var item CustomerContact
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid customer contact")
			return
		}
		var updated CustomerContact
		err := a.store.Mutate(func(data *AppData) error {
			index := customerContactIndex(*data, id)
			if index < 0 {
				return fmt.Errorf("客户联系人不存在")
			}
			current := data.CustomerContacts[index]
			oldCustomerID := current.CustomerID
			item.ID = id
			if err := normalizeCustomerContact(*data, &item, &current); err != nil {
				return err
			}
			if item.IsDefault {
				clearDefaultCustomerContact(data, item.CustomerID)
			}
			data.CustomerContacts[index] = item
			if oldCustomerID != item.CustomerID {
				syncCustomerPrimaryContact(data, oldCustomerID)
			}
			syncCustomerPrimaryContact(data, item.CustomerID)
			updated = data.CustomerContacts[index]
			addAudit(data, session.User.Username, "update", "customer_contact", id, item.Name, clientIP(r))
			return nil
		})
		a.respondUpdate(w, err, updated, "master.customer_contact.updated")
	case "projects":
		var item Project
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid project")
			return
		}
		var updated Project
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Projects {
				if data.Projects[i].ID != id {
					continue
				}
				if item.CustomerID == 0 {
					item.CustomerID = data.Projects[i].CustomerID
				}
				customer, ok := findCustomer(*data, item.CustomerID)
				if !ok {
					return fmt.Errorf("客户不存在")
				}
				if strings.TrimSpace(item.Name) == "" {
					return fmt.Errorf("项目名称不能为空")
				}
				item.ID = id
				item.Contact = fallback(item.Contact, customer.Contact)
				item.Phone = fallback(item.Phone, customer.Phone)
				item.Status = fallback(item.Status, data.Projects[i].Status)
				data.Projects[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "project", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("项目不存在")
		})
		a.respondUpdate(w, err, updated, "master.project.updated")
	case "products":
		var item Product
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid product")
			return
		}
		var updated Product
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Products {
				if data.Products[i].ID != id {
					continue
				}
				if strings.TrimSpace(item.Name) == "" {
					return fmt.Errorf("产品名称不能为空")
				}
				item.ID = id
				item.Line = fallback(item.Line, data.Products[i].Line)
				item.Unit = fallback(item.Unit, data.Products[i].Unit)
				item.Status = fallback(item.Status, data.Products[i].Status)
				data.Products[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "product", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("产品不存在")
		})
		a.respondUpdate(w, err, updated, "master.product.updated")
	case "price-policies":
		var item PricePolicy
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid price policy")
			return
		}
		var updated PricePolicy
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.PricePolicies {
				if data.PricePolicies[i].ID != id {
					continue
				}
				item.ID = id
				if err := normalizePricePolicy(*data, &item); err != nil {
					return err
				}
				data.PricePolicies[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "price_policy", id, fmt.Sprintf("product=%d price=%.2f", item.ProductID, item.SalePrice), clientIP(r))
				return nil
			}
			return fmt.Errorf("价格政策不存在")
		})
		a.respondUpdate(w, err, updated, "master.price_policy.updated")
	case "tax-rates":
		var item TaxRate
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid tax rate")
			return
		}
		var updated TaxRate
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.TaxRates {
				if data.TaxRates[i].ID != id {
					continue
				}
				item.ID = id
				if err := normalizeTaxRate(&item); err != nil {
					return err
				}
				data.TaxRates[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "tax_rate", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("税率不存在")
		})
		a.respondUpdate(w, err, updated, "master.tax_rate.updated")
	case "materials":
		var item Material
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid material")
			return
		}
		var updated Material
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Materials {
				if data.Materials[i].ID != id {
					continue
				}
				if strings.TrimSpace(item.Name) == "" {
					return fmt.Errorf("物料名称不能为空")
				}
				item.ID = id
				item.Unit = fallback(item.Unit, data.Materials[i].Unit)
				item.Status = fallback(item.Status, data.Materials[i].Status)
				data.Materials[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "material", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("物料不存在")
		})
		a.respondUpdate(w, err, updated, "master.material.updated")
	case "sites":
		var item Site
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid site")
			return
		}
		var updated Site
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Sites {
				if data.Sites[i].ID != id {
					continue
				}
				item.CompanyID = nonZeroInt(item.CompanyID, data.Sites[i].CompanyID)
				if !companyIDExists(data.Companies, item.CompanyID) {
					return fmt.Errorf("公司不存在")
				}
				if !userCanManageCompany(*data, session.User, item.CompanyID) {
					return fmt.Errorf("无权维护该公司站点")
				}
				if strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Code) == "" {
					return fmt.Errorf("站点名称和编码不能为空")
				}
				item.ID = id
				item.Status = fallback(item.Status, data.Sites[i].Status)
				data.Sites[i] = item
				updated = item
				syncSiteGeoFence(data, item)
				addAudit(data, session.User.Username, "update", "site", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("站点不存在")
		})
		a.respondUpdate(w, err, updated, "master.site.updated")
	case "plants":
		var item Plant
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid plant")
			return
		}
		var updated Plant
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Plants {
				if data.Plants[i].ID != id {
					continue
				}
				current := data.Plants[i]
				if item.SiteID == 0 {
					item.SiteID = current.SiteID
				}
				var err error
				item.SiteID, err = writableSiteID(*data, session.User, item.SiteID)
				if err != nil {
					return err
				}
				if strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Code) == "" {
					return fmt.Errorf("生产线名称和编码不能为空")
				}
				item.ID = id
				item.Capacity = fallback(item.Capacity, current.Capacity)
				item.Interface = ""
				item.Status = fallback(item.Status, current.Status)
				data.Plants[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "plant", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("生产线不存在")
		})
		a.respondUpdate(w, err, updated, "master.plant.updated")
	case "plant-buffer-locations":
		a.updatePlantBufferLocation(w, r, session, id)
	case "stock-yards":
		a.updateStockYard(w, r, session, id)
	case "stock-yard-piles":
		a.updateStockYardPile(w, r, session, id)
	case "carriers":
		var item Carrier
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid carrier")
			return
		}
		var updated Carrier
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Carriers {
				if data.Carriers[i].ID != id {
					continue
				}
				if strings.TrimSpace(item.Name) == "" {
					return fmt.Errorf("承运商名称不能为空")
				}
				item.ID = id
				item.SettleMode = fallback(strings.TrimSpace(item.SettleMode), data.Carriers[i].SettleMode)
				item.Status = fallback(strings.TrimSpace(item.Status), data.Carriers[i].Status)
				data.Carriers[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "carrier", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("承运商不存在")
		})
		a.respondUpdate(w, err, updated, "master.carrier.updated")
	case "drivers":
		var item Driver
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid driver")
			return
		}
		var updated Driver
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Drivers {
				if data.Drivers[i].ID != id {
					continue
				}
				if strings.TrimSpace(item.Name) == "" {
					return fmt.Errorf("司机姓名不能为空")
				}
				item.ID = id
				item.Status = fallback(item.Status, data.Drivers[i].Status)
				data.Drivers[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "driver", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("司机不存在")
		})
		a.respondUpdate(w, err, updated, "master.driver.updated")
	case "vehicles":
		var item Vehicle
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid vehicle")
			return
		}
		var updated Vehicle
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Vehicles {
				if data.Vehicles[i].ID != id {
					continue
				}
				if strings.TrimSpace(item.PlateNo) == "" {
					return fmt.Errorf("车牌号不能为空")
				}
				item.ID = id
				item.InternalNo = fallback(item.InternalNo, data.Vehicles[i].InternalNo)
				item.OnlineStatus = fallback(item.OnlineStatus, data.Vehicles[i].OnlineStatus)
				item.BusinessStatus = fallback(item.BusinessStatus, data.Vehicles[i].BusinessStatus)
				item.Status = fallback(item.Status, data.Vehicles[i].Status)
				data.Vehicles[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "vehicle", id, item.PlateNo, clientIP(r))
				return nil
			}
			return fmt.Errorf("车辆不存在")
		})
		a.respondUpdate(w, err, updated, "master.vehicle.updated")
	case "vehicle-devices":
		item, err := readVehicleDevicePayload(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid vehicle device")
			return
		}
		var updated VehicleDevice
		err = a.store.Mutate(func(data *AppData) error {
			for i := range data.VehicleDevices {
				if data.VehicleDevices[i].ID != id {
					continue
				}
				if strings.TrimSpace(item.DeviceNo) == "" {
					return fmt.Errorf("设备号不能为空")
				}
				vehicle, ok := findVehicle(*data, item.VehicleID)
				if !ok {
					return fmt.Errorf("车辆不存在")
				}
				if _, err := writableSiteID(*data, session.User, vehicle.SiteID); err != nil {
					return err
				}
				for _, existing := range data.VehicleDevices {
					if existing.ID != id && existing.DeviceNo == item.DeviceNo {
						return fmt.Errorf("设备号已绑定车辆")
					}
					if existing.ID != id && existing.VehicleID == item.VehicleID {
						return fmt.Errorf("车辆已绑定定位设备")
					}
				}
				item.ID = id
				item.LastSeenAt = fallback(item.LastSeenAt, data.VehicleDevices[i].LastSeenAt)
				data.VehicleDevices[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "vehicle_device", id, item.DeviceNo, clientIP(r))
				return nil
			}
			return fmt.Errorf("定位设备不存在")
		})
		a.respondUpdate(w, err, updated, "master.vehicle_device.updated")
	case "inventory":
		var item InventoryItem
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid inventory")
			return
		}
		var updated InventoryItem
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.Inventory {
				if data.Inventory[i].ID != id {
					continue
				}
				if item.SiteID == 0 {
					item.SiteID = data.Inventory[i].SiteID
				}
				if _, ok := findSite(*data, item.SiteID); !ok {
					return fmt.Errorf("站点不存在")
				}
				if item.MaterialID == 0 {
					item.MaterialID = data.Inventory[i].MaterialID
				}
				material, ok := findMaterial(*data, item.MaterialID)
				if !ok {
					return fmt.Errorf("物料不存在")
				}
				item.ID = id
				item.Warehouse = fallback(item.Warehouse, data.Inventory[i].Warehouse)
				item.Silo = fallback(item.Silo, data.Inventory[i].Silo)
				item.Unit = fallback(item.Unit, material.Unit)
				item.QualityStatus = fallback(item.QualityStatus, data.Inventory[i].QualityStatus)
				item.AvailableStatus = fallback(item.AvailableStatus, data.Inventory[i].AvailableStatus)
				item.UpdatedAt = nowString()
				data.Inventory[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "inventory", id, fmt.Sprintf("%d/%.2f", item.MaterialID, item.Quantity), clientIP(r))
				return nil
			}
			return fmt.Errorf("库存不存在")
		})
		a.respondUpdate(w, err, updated, "master.inventory.updated")
	default:
		writeError(w, http.StatusNotFound, "unknown master resource")
	}
}

func (a *App) deleteMasterResource(w http.ResponseWriter, r *http.Request, session Session, resource string, id int64) {
	var deleted interface{}
	err := a.store.Mutate(func(data *AppData) error {
		if masterResourceReferenced(*data, resource, id) {
			return fmt.Errorf("资源已被业务单据引用，不能删除")
		}
		switch resource {
		case "customers":
			for i, item := range data.Customers {
				if item.ID == id {
					deleted = item
					data.Customers = append(data.Customers[:i], data.Customers[i+1:]...)
					addAudit(data, session.User.Username, "delete", "customer", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("客户不存在")
		case "customer-contacts":
			for i, item := range data.CustomerContacts {
				if item.ID == id {
					deleted = item
					data.CustomerContacts = append(data.CustomerContacts[:i], data.CustomerContacts[i+1:]...)
					syncCustomerPrimaryContact(data, item.CustomerID)
					addAudit(data, session.User.Username, "delete", "customer_contact", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("客户联系人不存在")
		case "projects":
			for i, item := range data.Projects {
				if item.ID == id {
					deleted = item
					data.Projects = append(data.Projects[:i], data.Projects[i+1:]...)
					addAudit(data, session.User.Username, "delete", "project", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("项目不存在")
		case "products":
			for i, item := range data.Products {
				if item.ID == id {
					deleted = item
					data.Products = append(data.Products[:i], data.Products[i+1:]...)
					addAudit(data, session.User.Username, "delete", "product", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("产品不存在")
		case "price-policies":
			for i, item := range data.PricePolicies {
				if item.ID == id {
					deleted = item
					data.PricePolicies = append(data.PricePolicies[:i], data.PricePolicies[i+1:]...)
					addAudit(data, session.User.Username, "delete", "price_policy", id, fmt.Sprintf("product=%d price=%.2f", item.ProductID, item.SalePrice), clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("价格政策不存在")
		case "tax-rates":
			for i, item := range data.TaxRates {
				if item.ID == id {
					deleted = item
					data.TaxRates = append(data.TaxRates[:i], data.TaxRates[i+1:]...)
					addAudit(data, session.User.Username, "delete", "tax_rate", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("税率不存在")
		case "materials":
			for i, item := range data.Materials {
				if item.ID == id {
					deleted = item
					data.Materials = append(data.Materials[:i], data.Materials[i+1:]...)
					addAudit(data, session.User.Username, "delete", "material", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("物料不存在")
		case "sites":
			for i, item := range data.Sites {
				if item.ID == id {
					deleted = item
					data.Sites = append(data.Sites[:i], data.Sites[i+1:]...)
					archiveSiteGeoFences(data, id)
					addAudit(data, session.User.Username, "delete", "site", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("站点不存在")
		case "plants":
			for i, item := range data.Plants {
				if item.ID == id {
					deleted = item
					data.Plants = append(data.Plants[:i], data.Plants[i+1:]...)
					addAudit(data, session.User.Username, "delete", "plant", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("生产线不存在")
		case "plant-buffer-locations":
			for _, item := range data.PlantBufferFlows {
				if item.BufferID == id {
					return fmt.Errorf("暂存仓位已有流水，不能删除")
				}
			}
			for i, item := range data.PlantBufferLocations {
				if item.ID == id {
					if item.CurrentQty != 0 {
						return fmt.Errorf("暂存仓位有余额，不能删除")
					}
					deleted = item
					data.PlantBufferLocations = append(data.PlantBufferLocations[:i], data.PlantBufferLocations[i+1:]...)
					addAudit(data, session.User.Username, "delete", "plant_buffer", id, item.Code, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("暂存仓位不存在")
		case "stock-yards":
			for _, item := range data.StockYardPiles {
				if item.YardID == id {
					return fmt.Errorf("堆场下已有堆位，不能删除")
				}
			}
			for i, item := range data.StockYards {
				if item.ID == id {
					deleted = item
					data.StockYards = append(data.StockYards[:i], data.StockYards[i+1:]...)
					addAudit(data, session.User.Username, "delete", "stock_yard", id, item.Code, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("堆场不存在")
		case "stock-yard-piles":
			for _, item := range data.StockYardFlows {
				if item.PileID == id {
					return fmt.Errorf("堆位已有流水，不能删除")
				}
			}
			for i, item := range data.StockYardPiles {
				if item.ID == id {
					if item.CurrentQty != 0 {
						return fmt.Errorf("堆位有余额，不能删除")
					}
					deleted = item
					data.StockYardPiles = append(data.StockYardPiles[:i], data.StockYardPiles[i+1:]...)
					addAudit(data, session.User.Username, "delete", "stock_yard_pile", id, item.Code, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("堆位不存在")
		case "carriers":
			for i, item := range data.Carriers {
				if item.ID == id {
					deleted = item
					data.Carriers = append(data.Carriers[:i], data.Carriers[i+1:]...)
					addAudit(data, session.User.Username, "delete", "carrier", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("承运商不存在")
		case "drivers":
			for i, item := range data.Drivers {
				if item.ID == id {
					deleted = item
					data.Drivers = append(data.Drivers[:i], data.Drivers[i+1:]...)
					addAudit(data, session.User.Username, "delete", "driver", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("司机不存在")
		case "vehicles":
			for i, item := range data.Vehicles {
				if item.ID == id {
					deleted = item
					data.Vehicles = append(data.Vehicles[:i], data.Vehicles[i+1:]...)
					devices := data.VehicleDevices[:0]
					for _, device := range data.VehicleDevices {
						if device.VehicleID != id {
							devices = append(devices, device)
						}
					}
					data.VehicleDevices = devices
					addAudit(data, session.User.Username, "delete", "vehicle", id, item.PlateNo, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("车辆不存在")
		case "vehicle-devices":
			for i, item := range data.VehicleDevices {
				if item.ID == id {
					vehicle, ok := findVehicle(*data, item.VehicleID)
					if ok {
						if _, err := writableSiteID(*data, session.User, vehicle.SiteID); err != nil {
							return err
						}
					}
					deleted = item
					data.VehicleDevices = append(data.VehicleDevices[:i], data.VehicleDevices[i+1:]...)
					addAudit(data, session.User.Username, "delete", "vehicle_device", id, item.DeviceNo, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("定位设备不存在")
		case "inventory":
			for i, item := range data.Inventory {
				if item.ID == id {
					deleted = item
					data.Inventory = append(data.Inventory[:i], data.Inventory[i+1:]...)
					addAudit(data, session.User.Username, "delete", "inventory", id, fmt.Sprintf("%d/%.2f", item.MaterialID, item.Quantity), clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("库存不存在")
		default:
			return fmt.Errorf("unknown master resource")
		}
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit("master."+strings.TrimSuffix(resource, "s")+".deleted", deleted)
	writeJSON(w, http.StatusOK, deleted)
}

func (a *App) respondUpdate(w http.ResponseWriter, err error, payload interface{}, topic string) {
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit(topic, payload)
	writeJSON(w, http.StatusOK, payload)
}

func masterResourceReferenced(data AppData, resource string, id int64) bool {
	switch resource {
	case "customers":
		for _, item := range data.Projects {
			if item.CustomerID == id {
				return true
			}
		}
		for _, item := range data.Orders {
			if item.CustomerID == id {
				return true
			}
		}
		for _, item := range data.Contracts {
			if item.CustomerID == id {
				return true
			}
		}
		for _, item := range data.Receivables {
			if item.CustomerID == id {
				return true
			}
		}
	case "projects":
		for _, item := range data.Orders {
			if item.ProjectID == id {
				return true
			}
		}
		for _, item := range data.Contracts {
			if item.ProjectID == id {
				return true
			}
		}
		for _, item := range data.Statements {
			if item.ProjectID == id {
				return true
			}
		}
	case "products":
		for _, item := range data.Orders {
			if item.ProductID == id {
				return true
			}
			for _, line := range item.Lines {
				if line.ProductID == id {
					return true
				}
			}
		}
		for _, item := range data.Contracts {
			for _, line := range item.Items {
				if line.ProductID == id {
					return true
				}
			}
		}
		for _, item := range data.MixDesigns {
			if item.ProductID == id {
				return true
			}
		}
	case "tax-rates":
		for _, item := range data.PricePolicies {
			if item.TaxRateID == id {
				return true
			}
		}
	case "materials":
		for _, item := range data.Inventory {
			if item.MaterialID == id {
				return true
			}
		}
		for _, item := range data.PurchaseOrders {
			if item.MaterialID == id {
				return true
			}
		}
		for _, item := range data.RawMaterialReceipts {
			if item.MaterialID == id {
				return true
			}
		}
		for _, item := range data.StockYardPiles {
			if item.MaterialID == id {
				return true
			}
		}
		for _, mix := range data.MixDesigns {
			for _, material := range mix.Materials {
				if material.MaterialID == id {
					return true
				}
			}
		}
	case "sites":
		for _, item := range data.Orders {
			if item.SiteID == id {
				return true
			}
		}
		for _, item := range data.Plants {
			if item.SiteID == id {
				return true
			}
		}
		for _, item := range data.Vehicles {
			if item.SiteID == id {
				return true
			}
		}
		for _, item := range data.Inventory {
			if item.SiteID == id {
				return true
			}
		}
		for _, item := range data.PlantBufferLocations {
			if item.SiteID == id {
				return true
			}
		}
		for _, item := range data.StockYards {
			if item.SiteID == id {
				return true
			}
		}
		for _, item := range data.StockYardPiles {
			if item.SiteID == id {
				return true
			}
		}
		for _, item := range data.DispatchOrders {
			if item.SiteID == id {
				return true
			}
		}
	case "plants":
		plantCode := ""
		for _, item := range data.Plants {
			if item.ID == id {
				plantCode = item.Code
				break
			}
		}
		if plantCode == "" {
			return false
		}
		for _, item := range data.ProductionBatches {
			if item.PlantCode == plantCode {
				return true
			}
		}
		for _, item := range data.PlantBufferLocations {
			if item.PlantID == id || item.PlantCode == plantCode {
				return true
			}
		}
	case "carriers":
		for _, item := range data.DispatchSchedules {
			if item.CarrierID == id {
				return true
			}
		}
		for _, item := range data.TransportSettlements {
			if item.CarrierID == id {
				return true
			}
		}
		for _, item := range data.TransportSettlementItems {
			if item.CarrierID == id {
				return true
			}
		}
	case "drivers":
		for _, item := range data.Vehicles {
			if item.DriverID == id {
				return true
			}
		}
		for _, item := range data.DispatchOrders {
			if item.DriverID == id {
				return true
			}
		}
	case "vehicles":
		for _, item := range data.DispatchOrders {
			if item.VehicleID == id {
				return true
			}
		}
		for _, item := range data.ScaleTickets {
			if item.VehicleID == id {
				return true
			}
		}
	}
	return false
}
