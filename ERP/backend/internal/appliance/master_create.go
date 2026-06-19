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
		item.Line = fallback(item.Line, "concrete")
		item.Unit = fallback(item.Unit, "m3")
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

func (a *App) createSite(w http.ResponseWriter, r *http.Request, session Session) {
	var item Site
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid site")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if item.CompanyID == 0 {
			item.CompanyID = 1
		}
		if !companyIDExists(data.Companies, item.CompanyID) {
			return fmt.Errorf("公司不存在")
		}
		if strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Code) == "" {
			return fmt.Errorf("站点名称和编码不能为空")
		}
		item.ID = nextID(data, "site")
		item.Status = fallback(item.Status, "running")
		data.Sites = append(data.Sites, item)
		addAudit(data, session.User.Username, "create", "site", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.site.created")
}

func (a *App) createInventoryItem(w http.ResponseWriter, r *http.Request, session Session) {
	var item InventoryItem
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid inventory")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if item.SiteID == 0 {
			return fmt.Errorf("库存必须关联站点")
		}
		if _, ok := findSite(*data, item.SiteID); !ok {
			return fmt.Errorf("站点不存在")
		}
		if item.MaterialID == 0 {
			return fmt.Errorf("库存必须关联物料")
		}
		material, ok := findMaterial(*data, item.MaterialID)
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
