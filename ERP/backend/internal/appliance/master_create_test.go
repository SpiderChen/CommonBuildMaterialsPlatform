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

	rec := testRequest(t, app, token, http.MethodPost, "/api/master/products", `{"name":"SMA 改性沥青混合料","spec":"SMA-13","basePrice":1200,"costPrice":860}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create product status %d: %s", rec.Code, rec.Body.String())
	}
	var product Product
	if err := json.Unmarshal(rec.Body.Bytes(), &product); err != nil {
		t.Fatalf("decode product: %v", err)
	}
	if product.ID == 0 || product.Line != "asphalt" || product.Unit != "t" || product.Status != "active" {
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
	if site.ID == 0 || site.Status != "active" {
		t.Fatalf("unexpected site defaults: %+v", site)
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/plants", `{"siteId":`+strconv.FormatInt(site.ID, 10)+`,"name":"龙岗 240 沥青线","code":"LG-AMP240","capacity":"240t/h"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant status %d: %s", rec.Code, rec.Body.String())
	}
	var plant Plant
	if err := json.Unmarshal(rec.Body.Bytes(), &plant); err != nil {
		t.Fatalf("decode plant: %v", err)
	}
	if plant.ID == 0 || plant.SiteID != site.ID || plant.Status != "running" || plant.Interface != "" {
		t.Fatalf("unexpected plant defaults: %+v", plant)
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
		{"/api/master/products", "SMA 改性沥青混合料"},
		{"/api/master/materials", "硅灰"},
		{"/api/master/projects", "深圳北站扩建"},
		{"/api/master/sites", "龙岗交付站"},
		{"/api/master/plants", "LG-AMP240"},
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

func TestSiteCreateAndUpdateSyncsGeoFence(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	body := `{"companyId":1,"name":"Fence Site","code":"FENCE-SITE","address":"site address","longitude":113.9123,"latitude":22.5123,"fenceRadius":650}`
	rec := testRequest(t, app, token, http.MethodPost, "/api/master/sites", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create site status %d: %s", rec.Code, rec.Body.String())
	}
	var site Site
	if err := json.Unmarshal(rec.Body.Bytes(), &site); err != nil {
		t.Fatalf("decode site: %v", err)
	}
	fence := loadSiteFence(t, app, token, site.ID)
	if fence.Type != "site" || fence.Shape != "circle" || fence.Radius != 650 || fence.Longitude != 113.9123 || fence.Latitude != 22.5123 {
		t.Fatalf("unexpected created site fence: %+v", fence)
	}

	update := `{"companyId":1,"name":"Fence Site Updated","code":"FENCE-SITE","address":"site address","longitude":113.9234,"latitude":22.5234,"fenceRadius":880}`
	rec = testRequest(t, app, token, http.MethodPut, "/api/master/sites/"+strconv.FormatInt(site.ID, 10), update)
	if rec.Code != http.StatusOK {
		t.Fatalf("update site status %d: %s", rec.Code, rec.Body.String())
	}
	updatedFence := loadSiteFence(t, app, token, site.ID)
	if updatedFence.ID != fence.ID || updatedFence.Name != "Fence Site Updated围栏" || updatedFence.Radius != 880 || updatedFence.Longitude != 113.9234 || updatedFence.Latitude != 22.5234 {
		t.Fatalf("unexpected updated site fence: %+v", updatedFence)
	}
}

func loadSiteFence(t *testing.T, app *App, token string, siteID int64) GeoFence {
	t.Helper()
	rec := testRequest(t, app, token, http.MethodGet, "/api/vehicle/fences", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list fences status %d: %s", rec.Code, rec.Body.String())
	}
	var fences []GeoFence
	if err := json.Unmarshal(rec.Body.Bytes(), &fences); err != nil {
		t.Fatalf("decode fences: %v", err)
	}
	for _, fence := range fences {
		if fence.Type == "site" && fence.SiteID == siteID && fence.Status == "active" {
			return fence
		}
	}
	t.Fatalf("site fence for %d not found in %+v", siteID, fences)
	return GeoFence{}
}

func TestLegacyWarehouseSiloMasterEndpointsRemoved(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	for _, item := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/master/warehouses", ""},
		{http.MethodPost, "/api/master/warehouses", `{"siteId":1,"name":"旧仓库","code":"OLD-WH"}`},
		{http.MethodPut, "/api/master/warehouses/1", `{"siteId":1,"name":"旧仓库","code":"OLD-WH"}`},
		{http.MethodDelete, "/api/master/warehouses/1", ""},
		{http.MethodGet, "/api/master/silos", ""},
		{http.MethodPost, "/api/master/silos", `{"warehouseId":1,"materialId":1,"name":"旧筒仓","code":"OLD-SILO"}`},
		{http.MethodPut, "/api/master/silos/1", `{"warehouseId":1,"materialId":1,"name":"旧筒仓","code":"OLD-SILO"}`},
		{http.MethodDelete, "/api/master/silos/1", ""},
	} {
		rec := testRequest(t, app, token, item.method, item.path, item.body)
		if rec.Code == http.StatusOK || rec.Code == http.StatusCreated {
			t.Fatalf("%s %s should be removed, got %d: %s", item.method, item.path, rec.Code, rec.Body.String())
		}
	}

	for _, path := range []string{"/api/bootstrap", "/api/procurement/overview"} {
		rec := testRequest(t, app, token, http.MethodGet, path, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("load %s status %d: %s", path, rec.Code, rec.Body.String())
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		if _, ok := payload["warehouses"]; ok {
			t.Fatalf("%s should not expose warehouses: %s", path, rec.Body.String())
		}
		if _, ok := payload["silos"]; ok {
			t.Fatalf("%s should not expose silos: %s", path, rec.Body.String())
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
	rec = testRequest(t, app, token, http.MethodPut, "/api/master/products/"+strconv.FormatInt(product.ID, 10), `{"name":"CRUD Product Updated","spec":"AC-20","basePrice":430,"unit":"t"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "CRUD Product Updated") {
		t.Fatalf("update product failed status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/products/"+strconv.FormatInt(product.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete product status %d: %s", rec.Code, rec.Body.String())
	}

	rec = testRequest(t, app, token, http.MethodPost, "/api/master/carriers", `{"name":"CRUD Carrier","contact":"Dispatch","phone":"13600000000"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create carrier status %d: %s", rec.Code, rec.Body.String())
	}
	var carrier Carrier
	if err := json.Unmarshal(rec.Body.Bytes(), &carrier); err != nil {
		t.Fatalf("decode carrier: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPut, "/api/master/carriers/"+strconv.FormatInt(carrier.ID, 10), `{"name":"CRUD Carrier Updated","contact":"Dispatch Updated","phone":"13700000000","settleMode":"per_trip","status":"active"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "CRUD Carrier Updated") {
		t.Fatalf("update carrier failed status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/carriers/"+strconv.FormatInt(carrier.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete carrier status %d: %s", rec.Code, rec.Body.String())
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
	rec = testRequest(t, app, token, http.MethodPost, "/api/master/plants", `{"siteId":`+strconv.FormatInt(site.ID, 10)+`,"name":"CRUD Plant","code":"CRUD-PLANT","capacity":"160t/h"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create plant status %d: %s", rec.Code, rec.Body.String())
	}
	var plant Plant
	if err := json.Unmarshal(rec.Body.Bytes(), &plant); err != nil {
		t.Fatalf("decode plant: %v", err)
	}
	rec = testRequest(t, app, token, http.MethodPut, "/api/master/plants/"+strconv.FormatInt(plant.ID, 10), `{"siteId":`+strconv.FormatInt(site.ID, 10)+`,"name":"CRUD Plant Updated","code":"CRUD-PLANT-2","capacity":"200t/h","status":"active"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "CRUD Plant Updated") {
		t.Fatalf("update plant failed status %d: %s", rec.Code, rec.Body.String())
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
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/plants/"+strconv.FormatInt(plant.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete plant status %d: %s", rec.Code, rec.Body.String())
	}
	rec = testRequest(t, app, token, http.MethodDelete, "/api/master/sites/"+strconv.FormatInt(site.ID, 10), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete site status %d: %s", rec.Code, rec.Body.String())
	}
}
