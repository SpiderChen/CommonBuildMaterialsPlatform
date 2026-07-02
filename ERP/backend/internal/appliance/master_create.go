package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

func (a *App) createProject(w http.ResponseWriter, r *http.Request, session Session) {
	var item Project
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid project")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if item.CustomerID == 0 {
			return fmt.Errorf("项目必须关联客户")
		}
		customer, ok := findCustomer(*data, item.CustomerID)
		if !ok {
			return fmt.Errorf("客户不存在")
		}
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("项目名称不能为空")
		}
		item.Contact = fallback(item.Contact, customer.Contact)
		item.Phone = fallback(item.Phone, customer.Phone)
		item.ID = nextID(data, "project")
		item.Status = fallback(item.Status, "active")
		data.Projects = append(data.Projects, item)
		addAudit(data, session.User.Username, "create", "project", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.project.created")
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request, session Session) {
	var item Product
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid product")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("产品名称不能为空")
		}
		item.Line = fallback(item.Line, "asphalt")
		item.Unit = fallback(item.Unit, "t")
		item.ID = nextID(data, "product")
		item.Status = fallback(item.Status, "active")
		data.Products = append(data.Products, item)
		addAudit(data, session.User.Username, "create", "product", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.product.created")
}

func (a *App) createMaterial(w http.ResponseWriter, r *http.Request, session Session) {
	var item Material
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid material")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("物料名称不能为空")
		}
		item.Unit = fallback(item.Unit, "t")
		item.ID = nextID(data, "material")
		item.Status = fallback(item.Status, "active")
		data.Materials = append(data.Materials, item)
		addAudit(data, session.User.Username, "create", "material", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.material.created")
}

func (a *App) createDriver(w http.ResponseWriter, r *http.Request, session Session) {
	var item Driver
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid driver")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("司机姓名不能为空")
		}
		item.ID = nextID(data, "driver")
		item.Status = fallback(item.Status, "active")
		data.Drivers = append(data.Drivers, item)
		addAudit(data, session.User.Username, "create", "driver", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.driver.created")
}

func (a *App) createCarrier(w http.ResponseWriter, r *http.Request, session Session) {
	var item Carrier
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid carrier")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("承运商名称不能为空")
		}
		item.ID = nextID(data, "carrier")
		item.SettleMode = fallback(item.SettleMode, "monthly")
		item.Status = fallback(item.Status, "active")
		data.Carriers = append(data.Carriers, item)
		addAudit(data, session.User.Username, "create", "carrier", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.carrier.created")
}

type vehicleDevicePayload struct {
	VehicleDevice
}

func readVehicleDevicePayload(r *http.Request) (VehicleDevice, error) {
	var payload vehicleDevicePayload
	if err := readJSON(r, &payload); err != nil {
		return VehicleDevice{}, err
	}
	payload.DeviceNo = strings.TrimSpace(payload.DeviceNo)
	payload.Protocol = fallback(strings.TrimSpace(payload.Protocol), "gps-forwarder")
	payload.Vendor = fallback(strings.TrimSpace(payload.Vendor), "GPS 转发器")
	payload.Status = fallback(strings.TrimSpace(payload.Status), "active")
	return payload.VehicleDevice, nil
}

func nextVehicleDeviceID(data *AppData) int64 {
	if data.Next == nil {
		data.Next = map[string]int64{}
	}
	if data.Next["vehicleDevice"] == 0 {
		for _, item := range data.VehicleDevices {
			if item.ID > data.Next["vehicleDevice"] {
				data.Next["vehicleDevice"] = item.ID
			}
		}
	}
	return nextID(data, "vehicleDevice")
}

func (a *App) createVehicleDevice(w http.ResponseWriter, r *http.Request, session Session) {
	item, err := readVehicleDevicePayload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid vehicle device")
		return
	}
	err = a.store.Mutate(func(data *AppData) error {
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
			if existing.DeviceNo == item.DeviceNo {
				return fmt.Errorf("设备号已绑定车辆")
			}
			if existing.VehicleID == item.VehicleID {
				return fmt.Errorf("车辆已绑定定位设备")
			}
		}
		item.ID = nextVehicleDeviceID(data)
		data.VehicleDevices = append(data.VehicleDevices, item)
		addAudit(data, session.User.Username, "create", "vehicle_device", item.ID, item.DeviceNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.vehicle_device.created")
}

func (a *App) createSite(w http.ResponseWriter, r *http.Request, session Session) {
	var item Site
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid site")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if item.CompanyID == 0 {
			return fmt.Errorf("站点必须绑定公司")
		}
		if !companyIDExists(data.Companies, item.CompanyID) {
			return fmt.Errorf("公司不存在")
		}
		if !userCanManageCompany(*data, session.User, item.CompanyID) {
			return fmt.Errorf("无权维护该公司站点")
		}
		if strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Code) == "" {
			return fmt.Errorf("站点名称和编码不能为空")
		}
		item.ID = nextID(data, "site")
		item.Status = fallback(item.Status, "active")
		data.Sites = append(data.Sites, item)
		syncSiteGeoFence(data, item)
		addAudit(data, session.User.Username, "create", "site", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.site.created")
}

func (a *App) createPlant(w http.ResponseWriter, r *http.Request, session Session) {
	var item Plant
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid plant")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		var err error
		item.SiteID, err = writableSiteID(*data, session.User, item.SiteID)
		if err != nil {
			return err
		}
		if strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Code) == "" {
			return fmt.Errorf("生产线名称和编码不能为空")
		}
		item.Capacity = fallback(item.Capacity, "0")
		item.Interface = ""
		item.ID = nextID(data, "plant")
		item.Status = fallback(item.Status, "running")
		data.Plants = append(data.Plants, item)
		addAudit(data, session.User.Username, "create", "plant", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.plant.created")
}

func (a *App) createInventoryItem(w http.ResponseWriter, r *http.Request, session Session) {
	var item InventoryItem
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid inventory")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		var err error
		item.SiteID, err = writableSiteID(*data, session.User, item.SiteID)
		if err != nil {
			return err
		}
		if item.MaterialID == 0 {
			return fmt.Errorf("库存必须关联物料")
		}
		material, ok := scopedMaterial(*data, session.User, item.MaterialID)
		if !ok {
			return fmt.Errorf("物料不存在")
		}
		item.Warehouse = fallback(item.Warehouse, "manual")
		item.Silo = fallback(item.Silo, "manual")
		item.Unit = fallback(item.Unit, material.Unit)
		item.QualityStatus = fallback(item.QualityStatus, "passed")
		item.AvailableStatus = fallback(item.AvailableStatus, "available")
		item.UpdatedAt = nowString()
		item.ID = nextID(data, "inventory")
		data.Inventory = append(data.Inventory, item)
		addAudit(data, session.User.Username, "create", "inventory", item.ID, fmt.Sprintf("%d/%.2f", item.MaterialID, item.Quantity), clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.inventory.created")
}
