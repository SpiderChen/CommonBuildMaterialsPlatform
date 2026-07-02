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
				if strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Code) == "" {
					return fmt.Errorf("站点名称和编码不能为空")
				}
				item.ID = id
				item.Status = fallback(item.Status, data.Sites[i].Status)
				data.Sites[i] = item
				updated = item
				addAudit(data, session.User.Username, "update", "site", id, item.Name, clientIP(r))
				return nil
			}
			return fmt.Errorf("站点不存在")
		})
		a.respondUpdate(w, err, updated, "master.site.updated")
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
					addAudit(data, session.User.Username, "delete", "site", id, item.Name, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("站点不存在")
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
					addAudit(data, session.User.Username, "delete", "vehicle", id, item.PlateNo, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("车辆不存在")
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
		for _, item := range data.DispatchOrders {
			if item.SiteID == id {
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
