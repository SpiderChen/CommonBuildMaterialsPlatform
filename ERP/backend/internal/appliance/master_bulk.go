package appliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type MasterDataExport struct {
	Resource string           `json:"resource"`
	Count    int              `json:"count"`
	Fields   []string         `json:"fields"`
	Rows     []map[string]any `json:"rows"`
}

type masterDataImportRequest struct {
	Resource string           `json:"resource"`
	Mode     string           `json:"mode"`
	Rows     []map[string]any `json:"rows"`
}

type MasterDataImportResult struct {
	Resource string   `json:"resource"`
	Mode     string   `json:"mode"`
	Created  int      `json:"created"`
	Updated  int      `json:"updated"`
	Errors   []string `json:"errors"`
}

func (a *App) exportMasterData(w http.ResponseWriter, r *http.Request, session Session) {
	resource := normalizeMasterBulkResource(r.URL.Query().Get("resource"))
	data := scopedData(a.mustSnapshot(), session.User)
	result, err := buildMasterDataExport(data, resource)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) importMasterData(w http.ResponseWriter, r *http.Request, session Session) {
	var req masterDataImportRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid master import payload")
		return
	}
	req.Resource = normalizeMasterBulkResource(req.Resource)
	req.Mode = strings.ToLower(strings.TrimSpace(fallback(req.Mode, "create")))
	if req.Mode != "create" && req.Mode != "upsert" {
		writeError(w, http.StatusBadRequest, "导入模式仅支持 create 或 upsert")
		return
	}
	if len(req.Rows) == 0 {
		writeError(w, http.StatusBadRequest, "导入数据不能为空")
		return
	}
	result := MasterDataImportResult{Resource: req.Resource, Mode: req.Mode}
	err := a.store.Mutate(func(data *AppData) error {
		for index, row := range req.Rows {
			created, updated, err := importMasterDataRow(data, req.Resource, req.Mode, row)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("第 %d 行：%s", index+1, err.Error()))
				continue
			}
			if created {
				result.Created++
			}
			if updated {
				result.Updated++
			}
		}
		if result.Created == 0 && result.Updated == 0 && len(result.Errors) > 0 {
			return fmt.Errorf("导入失败：%s", strings.Join(result.Errors, "；"))
		}
		addAudit(data, session.User.Username, "import", "master_"+req.Resource, int64(result.Created+result.Updated), fmt.Sprintf("created=%d updated=%d errors=%d", result.Created, result.Updated, len(result.Errors)), clientIP(r))
		return nil
	})
	a.respondMutation(w, err, result, "master.bulk.imported")
}

func buildMasterDataExport(data AppData, resource string) (MasterDataExport, error) {
	result := MasterDataExport{Resource: resource}
	switch resource {
	case "customers":
		result.Fields = []string{"id", "companyId", "name", "contact", "phone", "creditLimit", "receivable", "paymentTerm", "status"}
		for _, item := range data.Customers {
			result.Rows = append(result.Rows, map[string]any{"id": item.ID, "companyId": item.CompanyID, "name": item.Name, "contact": item.Contact, "phone": item.Phone, "creditLimit": item.CreditLimit, "receivable": item.Receivable, "paymentTerm": item.PaymentTerm, "status": item.Status})
		}
	case "products":
		result.Fields = []string{"id", "line", "name", "spec", "unit", "basePrice", "costPrice", "requiresMix", "status"}
		for _, item := range data.Products {
			result.Rows = append(result.Rows, map[string]any{"id": item.ID, "line": item.Line, "name": item.Name, "spec": item.Spec, "unit": item.Unit, "basePrice": item.BasePrice, "costPrice": item.CostPrice, "requiresMix": item.RequiresMix, "status": item.Status})
		}
	case "materials":
		result.Fields = []string{"id", "name", "spec", "unit", "safeStock", "status"}
		for _, item := range data.Materials {
			result.Rows = append(result.Rows, map[string]any{"id": item.ID, "name": item.Name, "spec": item.Spec, "unit": item.Unit, "safeStock": item.SafeStock, "status": item.Status})
		}
	case "vehicles":
		result.Fields = []string{"id", "plateNo", "vehicleType", "capacity", "carrier", "siteId", "driverId", "onlineStatus", "businessStatus", "certExpiresAt", "status"}
		for _, item := range data.Vehicles {
			result.Rows = append(result.Rows, map[string]any{"id": item.ID, "plateNo": item.PlateNo, "vehicleType": item.VehicleType, "capacity": item.Capacity, "carrier": item.Carrier, "siteId": item.SiteID, "driverId": item.DriverID, "onlineStatus": item.OnlineStatus, "businessStatus": item.BusinessStatus, "certExpiresAt": item.CertExpiresAt, "status": item.Status})
		}
	case "drivers":
		result.Fields = []string{"id", "name", "phone", "licenseNo", "licenseExpire", "status"}
		for _, item := range data.Drivers {
			result.Rows = append(result.Rows, map[string]any{"id": item.ID, "name": item.Name, "phone": item.Phone, "licenseNo": item.LicenseNo, "licenseExpire": item.LicenseExpire, "status": item.Status})
		}
	default:
		return MasterDataExport{}, fmt.Errorf("不支持的主数据资源")
	}
	result.Count = len(result.Rows)
	return result, nil
}

func importMasterDataRow(data *AppData, resource string, mode string, row map[string]any) (bool, bool, error) {
	switch resource {
	case "customers":
		var item Customer
		if err := decodeMasterRow(row, &item); err != nil {
			return false, false, err
		}
		if strings.TrimSpace(item.Name) == "" {
			return false, false, fmt.Errorf("客户名称不能为空")
		}
		item.CompanyID = nonZeroInt(item.CompanyID, 1)
		item.Status = fallback(item.Status, "active")
		if mode == "upsert" && item.ID != 0 {
			for i := range data.Customers {
				if data.Customers[i].ID == item.ID {
					data.Customers[i] = item
					return false, true, nil
				}
			}
		}
		item.ID = nextID(data, "customer")
		data.Customers = append(data.Customers, item)
		return true, false, nil
	case "products":
		var item Product
		if err := decodeMasterRow(row, &item); err != nil {
			return false, false, err
		}
		if strings.TrimSpace(item.Name) == "" {
			return false, false, fmt.Errorf("产品名称不能为空")
		}
		item.Line = fallback(item.Line, "concrete")
		item.Unit = fallback(item.Unit, "m3")
		item.Status = fallback(item.Status, "active")
		if mode == "upsert" && item.ID != 0 {
			for i := range data.Products {
				if data.Products[i].ID == item.ID {
					data.Products[i] = item
					return false, true, nil
				}
			}
		}
		item.ID = nextID(data, "product")
		data.Products = append(data.Products, item)
		return true, false, nil
	case "materials":
		var item Material
		if err := decodeMasterRow(row, &item); err != nil {
			return false, false, err
		}
		if strings.TrimSpace(item.Name) == "" {
			return false, false, fmt.Errorf("物料名称不能为空")
		}
		item.Unit = fallback(item.Unit, "t")
		item.Status = fallback(item.Status, "active")
		if mode == "upsert" && item.ID != 0 {
			for i := range data.Materials {
				if data.Materials[i].ID == item.ID {
					data.Materials[i] = item
					return false, true, nil
				}
			}
		}
		item.ID = nextID(data, "material")
		data.Materials = append(data.Materials, item)
		return true, false, nil
	case "vehicles":
		var item Vehicle
		if err := decodeMasterRow(row, &item); err != nil {
			return false, false, err
		}
		if strings.TrimSpace(item.PlateNo) == "" {
			return false, false, fmt.Errorf("车牌号不能为空")
		}
		item.OnlineStatus = fallback(item.OnlineStatus, "offline")
		item.BusinessStatus = fallback(item.BusinessStatus, "idle")
		item.Status = fallback(item.Status, "active")
		if mode == "upsert" && item.ID != 0 {
			for i := range data.Vehicles {
				if data.Vehicles[i].ID == item.ID {
					data.Vehicles[i] = item
					return false, true, nil
				}
			}
		}
		item.ID = nextID(data, "vehicle")
		data.Vehicles = append(data.Vehicles, item)
		return true, false, nil
	case "drivers":
		var item Driver
		if err := decodeMasterRow(row, &item); err != nil {
			return false, false, err
		}
		if strings.TrimSpace(item.Name) == "" {
			return false, false, fmt.Errorf("司机姓名不能为空")
		}
		item.Status = fallback(item.Status, "active")
		if mode == "upsert" && item.ID != 0 {
			for i := range data.Drivers {
				if data.Drivers[i].ID == item.ID {
					data.Drivers[i] = item
					return false, true, nil
				}
			}
		}
		item.ID = nextID(data, "driver")
		data.Drivers = append(data.Drivers, item)
		return true, false, nil
	default:
		return false, false, fmt.Errorf("不支持的主数据资源")
	}
}

func decodeMasterRow(row map[string]any, target any) error {
	payload, err := json.Marshal(row)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func normalizeMasterBulkResource(resource string) string {
	return strings.ToLower(strings.TrimSpace(resource))
}
