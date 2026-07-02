package appliance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestMasterResourceCreationCoversDeliveryBasics(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/master/products", `{"name":"UHPC 超高性能混凝土","spec":"C120","basePrice":1200,"costPrice":860}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create product status %d: %s", rec.Code, rec.Body.String())
	}
	var product Product
	if err := json.Unmarshal(rec.Body.Bytes(), &product); err != nil {
		t.Fatalf("decode product: %v", err)
	}
	if product.ID == 0 || product.Line != "concrete" || product.Unit != "m3" || product.Status != "active" {
		t.Fatalf("unexpected product defaults: %+v", product)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/materials", `{"name":"硅灰","spec":"S95","safeStock":20}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create material status %d: %s", rec.Code, rec.Body.String())
	}
	var material Material
	if err := json.Unmarshal(rec.Body.Bytes(), &material); err != nil {
		t.Fatalf("decode material: %v", err)
	}
	if material.ID == 0 || material.Unit != "t" || material.Status != "active" {
		t.Fatalf("unexpected material defaults: %+v", material)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/projects", `{"customerId":1,"name":"深圳北站扩建","address":"深圳北站","longitude":114.03,"latitude":22.61}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create project status %d: %s", rec.Code, rec.Body.String())
	}
	var project Project
	if err := json.Unmarshal(rec.Body.Bytes(), &project); err != nil {
		t.Fatalf("decode project: %v", err)
	}
	if project.ID == 0 || project.Contact == "" || project.Phone == "" || project.Status != "active" {
		t.Fatalf("unexpected project defaults: %+v", project)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/sites", `{"companyId":1,"name":"龙岗交付站","code":"LG-OPS","address":"深圳龙岗"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create site status %d: %s", rec.Code, rec.Body.String())
	}
	var site Site
	if err := json.Unmarshal(rec.Body.Bytes(), &site); err != nil {
		t.Fatalf("decode site: %v", err)
	}
	if site.ID == 0 || site.Status != "running" {
		t.Fatalf("unexpected site defaults: %+v", site)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/inventory", `{"siteId":`+strconv.FormatInt(site.ID, 10)+`,"materialId":`+strconv.FormatInt(material.ID, 10)+`,"quantity":32}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create inventory status %d: %s", rec.Code, rec.Body.String())
	}
	var inventory InventoryItem
	if err := json.Unmarshal(rec.Body.Bytes(), &inventory); err != nil {
		t.Fatalf("decode inventory: %v", err)
	}
	if inventory.ID == 0 || inventory.Unit != material.Unit || inventory.QualityStatus != "passed" || inventory.AvailableStatus != "available" || inventory.UpdatedAt == "" {
		t.Fatalf("unexpected inventory defaults: %+v", inventory)
	}

	for _, item := range []struct {
		path string
		body string
		want string
	}{
		{"/api/master/drivers", `{"name":"赵师傅","phone":"13900001111","licenseNo":"D-10086"}`, "赵师傅"},
		{"/api/master/carriers", `{"name":"湾区承运服务","contact":"刘调度","phone":"13900002222"}`, "湾区承运服务"},
	} {
		rec = testRequest(t, app, token, http.MethodPost, item.path, item.body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create %s status %d: %s", item.path, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), item.want) {
			t.Fatalf("created payload missing %s: %s", item.want, rec.Body.String())
		}
	}

	for _, item := range []struct {
		path string
		want string
	}{
		{"/api/master/products", "UHPC 超高性能混凝土"},
		{"/api/master/materials", "硅灰"},
		{"/api/master/projects", "深圳北站扩建"},
		{"/api/master/sites", "龙岗交付站"},
		{"/api/master/inventory", strconv.FormatInt(material.ID, 10)},
		{"/api/master/drivers", "赵师傅"},
		{"/api/master/carriers", "湾区承运服务"},
	} {
		rec = testRequest(t, app, token, http.MethodGet, item.path, "")
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), item.want) {
			t.Fatalf("list %s missing %s, status %d: %s", item.path, item.want, rec.Code, rec.Body.String())
		}
	}
}

func TestMasterResourcesSupportUpdateAndDelete(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodPost, "/api/master/customers", `{"name":"CRUD Customer","contact":"Alice","phone":"13800000000","creditLimit":10000}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create customer status %d: %s", rec.Code, rec.Body.String())
	}
	var customer Customer
	if err := json.Unmarshal(rec.Body.Bytes(), &customer); err != nil {
		t.Fatalf("decode customer: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPut, "/api/master/customers/"+strconv.FormatInt(customer.ID, 10), `{"name":"CRUD Customer Updated","contact":"Bob","phone":"13900000000","creditLimit":12000}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "CRUD Customer Updated") {
		t.Fatalf("update customer failed status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/customers/"+strconv.FormatInt(customer.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete customer status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodGet, "/api/master/customers", "")
	if strings.Contains(rec.Body.String(), "CRUD Customer Updated") {
		t.Fatalf("deleted customer still listed: %s", rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/products", `{"name":"CRUD Product","spec":"C35","basePrice":410}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create product status %d: %s", rec.Code, rec.Body.String())
	}
	var product Product
	if err := json.Unmarshal(rec.Body.Bytes(), &product); err != nil {
		t.Fatalf("decode product: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPut, "/api/master/products/"+strconv.FormatInt(product.ID, 10), `{"name":"CRUD Product Updated","spec":"C40","basePrice":430,"unit":"m3"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "CRUD Product Updated") {
		t.Fatalf("update product failed status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/products/"+strconv.FormatInt(product.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete product status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/materials", `{"name":"CRUD Material","spec":"M1","safeStock":10}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create material status %d: %s", rec.Code, rec.Body.String())
	}
	var material Material
	if err := json.Unmarshal(rec.Body.Bytes(), &material); err != nil {
		t.Fatalf("decode material: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/master/sites", `{"companyId":1,"name":"CRUD Site","code":"CRUD-SITE","address":"site address"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create site status %d: %s", rec.Code, rec.Body.String())
	}
	var site Site
	if err := json.Unmarshal(rec.Body.Bytes(), &site); err != nil {
		t.Fatalf("decode site: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPost, "/api/master/inventory", `{"siteId":`+strconv.FormatInt(site.ID, 10)+`,"materialId":`+strconv.FormatInt(material.ID, 10)+`,"warehouse":"A","quantity":12}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create inventory status %d: %s", rec.Code, rec.Body.String())
	}
	var inventory InventoryItem
	if err := json.Unmarshal(rec.Body.Bytes(), &inventory); err != nil {
		t.Fatalf("decode inventory: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPut, "/api/master/inventory/"+strconv.FormatInt(inventory.ID, 10), `{"siteId":`+strconv.FormatInt(site.ID, 10)+`,"materialId":`+strconv.FormatInt(material.ID, 10)+`,"warehouse":"B","quantity":18}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"warehouse":"B"`) {
		t.Fatalf("update inventory failed status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/inventory/"+strconv.FormatInt(inventory.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete inventory status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/materials/"+strconv.FormatInt(material.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete material status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/sites/"+strconv.FormatInt(site.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete site status %d: %s", rec.Code, rec.Body.String())
	}
}
