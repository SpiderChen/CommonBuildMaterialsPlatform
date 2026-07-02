package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type OrganizationOverview struct {
	Group       GroupProfile        `json:"group"`
	Metrics     OrganizationMetrics `json:"metrics"`
	Nodes       []OrganizationNode  `json:"nodes"`
	Companies   []Company           `json:"companies"`
	Departments []Department        `json:"departments"`
	Sites       []Site              `json:"sites"`
}

type OrganizationMetrics struct {
	CompanyCount       int `json:"companyCount"`
	ActiveCompanyCount int `json:"activeCompanyCount"`
	SiteCount          int `json:"siteCount"`
	RunningSiteCount   int `json:"runningSiteCount"`
	DepartmentCount    int `json:"departmentCount"`
	UserCount          int `json:"userCount"`
}

type OrganizationNode struct {
	ID        string `json:"id"`
	ParentID  string `json:"parentId"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	CompanyID int64  `json:"companyId"`
	SiteID    int64  `json:"siteId"`
}

func (a *App) systemOrg(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		data.Companies = publicCompanies(data.Companies)
		writeJSON(w, http.StatusOK, buildOrganizationOverview(data))
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
		item.Name = strings.TrimSpace(item.Name)
		item.Code = strings.TrimSpace(item.Code)
		item.Level = fallback(strings.TrimSpace(item.Level), "subsidiary")
		item.Region = strings.TrimSpace(item.Region)
		if item.Name == "" || item.Code == "" {
			return fmt.Errorf("公司名称和编码不能为空")
		}
		if item.ParentID != 0 && !companyIDExists(data.Companies, item.ParentID) {
			return fmt.Errorf("上级公司不存在")
		}
		if !userCanManageCompany(*data, session.User, nonZeroInt(item.ParentID, item.ID)) {
			return fmt.Errorf("无权维护该公司")
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
		item.Name = strings.TrimSpace(item.Name)
		item.Code = strings.TrimSpace(item.Code)
		if item.CompanyID == 0 || item.Name == "" || item.Code == "" {
			return fmt.Errorf("部门必须包含公司、名称和编码")
		}
		if !companyIDExists(data.Companies, item.CompanyID) {
			return fmt.Errorf("公司不存在")
		}
		if !userCanManageCompany(*data, session.User, item.CompanyID) {
			return fmt.Errorf("无权维护该公司部门")
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
		if !userCanManageCompany(*data, session.User, id) {
			return fmt.Errorf("无权维护该公司")
		}
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
				if !userCanManageCompany(*data, session.User, data.Departments[i].CompanyID) {
					return fmt.Errorf("无权维护该公司部门")
				}
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

func buildOrganizationOverview(data AppData) OrganizationOverview {
	group := data.GroupProfile
	if group.Name == "" {
		group.Name = fallback(data.License.CustomerName, "建材集团")
	}
	if group.Code == "" {
		group.Code = "GROUP"
	}
	if group.Edition == "" {
		group.Edition = "集团版"
	}
	if group.OperatingMode == "" {
		group.OperatingMode = "集团总部统管"
	}
	if group.DataArchitecture == "" {
		group.DataArchitecture = "group-company-department"
	}

	nodes := []OrganizationNode{{
		ID:     "group:" + group.Code,
		Kind:   "group",
		Name:   group.Name,
		Code:   group.Code,
		Status: "active",
	}}
	rootID := nodes[0].ID
	metrics := OrganizationMetrics{
		CompanyCount:    len(data.Companies),
		SiteCount:       len(data.Sites),
		DepartmentCount: len(data.Departments),
		UserCount:       len(data.Users),
	}
	for _, company := range data.Companies {
		parentID := rootID
		if company.ParentID != 0 {
			parentID = orgNodeID("company", company.ParentID)
		}
		if company.Status == "active" {
			metrics.ActiveCompanyCount++
		}
		nodes = append(nodes, OrganizationNode{
			ID:        orgNodeID("company", company.ID),
			ParentID:  parentID,
			Kind:      fallback(company.Level, "company"),
			Name:      company.Name,
			Code:      company.Code,
			Region:    company.Region,
			Status:    company.Status,
			CompanyID: company.ID,
		})
	}
	for _, department := range data.Departments {
		parentID := orgNodeID("company", department.CompanyID)
		if department.ParentID != 0 {
			parentID = orgNodeID("department", department.ParentID)
		}
		nodes = append(nodes, OrganizationNode{
			ID:        orgNodeID("department", department.ID),
			ParentID:  parentID,
			Kind:      "department",
			Name:      department.Name,
			Code:      department.Code,
			Status:    department.Status,
			CompanyID: department.CompanyID,
		})
	}
	for _, site := range data.Sites {
		if site.Status == "running" || site.Status == "active" {
			metrics.RunningSiteCount++
		}
		nodes = append(nodes, OrganizationNode{
			ID:        orgNodeID("site", site.ID),
			ParentID:  orgNodeID("company", site.CompanyID),
			Kind:      "site",
			Name:      site.Name,
			Code:      site.Code,
			Region:    site.Address,
			Status:    orgSiteStatus(site.Status),
			CompanyID: site.CompanyID,
			SiteID:    site.ID,
		})
	}
	return OrganizationOverview{
		Group:       group,
		Metrics:     metrics,
		Nodes:       nodes,
		Companies:   publicCompanies(data.Companies),
		Departments: data.Departments,
		Sites:       data.Sites,
	}
}

func orgNodeID(kind string, id int64) string {
	return kind + ":" + strconv.FormatInt(id, 10)
}

func orgSiteStatus(status string) string {
	switch status {
	case "disabled", "inactive", "retired":
		return "disabled"
	default:
		return "active"
	}
}

func userCanManageCompany(data AppData, user User, companyID int64) bool {
	switch normalizeDataScope(roleDataScope(data.Roles, user.RoleCode)) {
	case "group":
		return true
	case "company":
		return descendantCompanyIDs(data.Companies, user.CompanyID)[companyID]
	case "site":
		site, ok := findSite(data, user.SiteID)
		return ok && site.CompanyID == companyID
	default:
		return false
	}
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
