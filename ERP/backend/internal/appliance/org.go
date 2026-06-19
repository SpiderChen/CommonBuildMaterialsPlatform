package appliance

import (
	"fmt"
	"net/http"
	"strconv"
)

func (a *App) systemOrg(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 && r.Method == http.MethodGet {
		data := a.mustSnapshot()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"companies":   publicCompanies(data.Companies),
			"departments": data.Departments,
			"sites":       data.Sites,
		})
		return
	}
	if len(parts) == 1 && r.Method == http.MethodPost {
		switch parts[0] {
		case "companies":
			a.createCompany(w, r, session)
		case "departments":
			a.createDepartment(w, r, session)
		default:
			writeError(w, http.StatusNotFound, "unknown org resource")
		}
		return
	}
	if len(parts) == 3 && parts[2] == "status" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req struct {
			Status string `json:"status"`
		}
		_ = readJSON(r, &req)
		status := fallback(req.Status, "active")
		switch parts[0] {
		case "companies":
			a.updateCompanyStatus(w, session, id, status, r)
		case "departments":
			a.updateDepartmentStatus(w, session, id, status, r)
		default:
			writeError(w, http.StatusNotFound, "unknown org resource")
		}
		return
	}
	writeError(w, http.StatusNotFound, "unknown org route")
}

func (a *App) createCompany(w http.ResponseWriter, r *http.Request, session Session) {
	var item Company
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid company")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if item.Name == "" || item.Code == "" {
			return fmt.Errorf("公司名称和编码不能为空")
		}
		if companyCodeExists(data.Companies, item.Code) {
			return fmt.Errorf("公司编码已存在")
		}
		item.ID = nextID(data, "company")
		item.Status = fallback(item.Status, "active")
		data.Companies = append(data.Companies, item)
		addAudit(data, session.User.Username, "create", "company", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, publicCompany(item), "system.org.company.created")
}

func (a *App) createDepartment(w http.ResponseWriter, r *http.Request, session Session) {
	var item Department
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid department")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if item.CompanyID == 0 || item.Name == "" || item.Code == "" {
			return fmt.Errorf("部门必须包含公司、名称和编码")
		}
		if !companyIDExists(data.Companies, item.CompanyID) {
			return fmt.Errorf("公司不存在")
		}
		if item.ParentID != 0 && !departmentIDExists(data.Departments, item.ParentID) {
			return fmt.Errorf("上级部门不存在")
		}
		if departmentCodeExists(data.Departments, item.CompanyID, item.Code) {
			return fmt.Errorf("部门编码已存在")
		}
		item.ID = nextID(data, "department")
		item.Status = fallback(item.Status, "active")
		data.Departments = append(data.Departments, item)
		addAudit(data, session.User.Username, "create", "department", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "system.org.department.created")
}

func (a *App) updateCompanyStatus(w http.ResponseWriter, session Session, id int64, status string, r *http.Request) {
	var updated Company
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.Companies {
			if data.Companies[i].ID == id {
				data.Companies[i].Status = status
				updated = publicCompany(data.Companies[i])
				addAudit(data, session.User.Username, "status", "company", id, status, clientIP(r))
				return nil
			}
		}
		return fmt.Errorf("公司不存在")
	})
	a.respondMutation(w, err, updated, "system.org.company.updated")
}

func (a *App) updateDepartmentStatus(w http.ResponseWriter, session Session, id int64, status string, r *http.Request) {
	var updated Department
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.Departments {
			if data.Departments[i].ID == id {
				data.Departments[i].Status = status
				updated = data.Departments[i]
				addAudit(data, session.User.Username, "status", "department", id, status, clientIP(r))
				return nil
			}
		}
		return fmt.Errorf("部门不存在")
	})
	a.respondMutation(w, err, updated, "system.org.department.updated")
}

func companyIDExists(items []Company, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func companyCodeExists(items []Company, code string) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}
	return false
}

func departmentIDExists(items []Department, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func departmentCodeExists(items []Department, companyID int64, code string) bool {
	for _, item := range items {
		if item.CompanyID == companyID && item.Code == code {
			return true
		}
	}
	return false
}
