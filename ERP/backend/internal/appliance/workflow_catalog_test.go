package appliance

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestWorkflowCatalogEndpointListsSupportedEvents(t *testing.T) {
	app := newTestHTTPApp(t)
	token := testLogin(t, app, "admin", "admin123")

	rec := testRequest(t, app, token, http.MethodGet, "/api/system/workflows/catalog", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("workflow catalog status %d: %s", rec.Code, rec.Body.String())
	}
	var catalog WorkflowCatalog
	if err := json.Unmarshal(rec.Body.Bytes(), &catalog); err != nil {
		t.Fatalf("decode workflow catalog: %v", err)
	}
	expectedBusinessEvents := []struct {
		eventType string
		resource  string
	}{
		{"sales_order.risk_detected", "sales_order"},
		{"ticket_void.requested", "ticket_void"},
		{"statement.confirm_requested", "statement"},
		{"system_user.status_change_requested", "system_user"},
		{"contract.submitted", "contract"},
		{"customer_blacklist.submitted", "customer_blacklist"},
		{"customer_blacklist_release.requested", "customer_blacklist_release"},
		{"delivery_sign.completed", "delivery_sign"},
		{"delivery_note.status_change_requested", "delivery_note"},
		{"gateway_route.status_change_requested", "gateway_route"},
		{"inventory_transfer.submitted", "inventory_transfer"},
		{"inventory_stocktake.review_requested", "inventory_stocktake"},
		{"mix_design.submitted", "mix_design"},
		{"mix_design.retire_requested", "mix_design"},
		{"mix_design_plant_profile.submitted", "mix_design_plant_profile"},
		{"mix_design_plant_profile.retire_requested", "mix_design_plant_profile"},
		{"laboratory_test.review_requested", "laboratory_test"},
		{"quality_exception.close_requested", "quality_exception"},
		{"quality_exception.submitted", "quality_exception"},
		{"oidc_provider.status_change_requested", "oidc_provider"},
		{"plant_buffer_adjustment.requested", "plant_buffer_adjustment"},
		{"production_plan.cancel_requested", "production_plan"},
		{"raw_material_inspection.review_requested", "raw_material_inspection"},
		{"red_letter_info.requested", "red_letter_info"},
		{"scim_provider.status_change_requested", "scim_provider"},
		{"stock_yard_adjustment.requested", "stock_yard_adjustment"},
		{"supplier_statement.submitted", "supplier_statement"},
	}
	for _, expected := range expectedBusinessEvents {
		if !hasWorkflowCatalogEvent(catalog.Events, expected.eventType, expected.resource) {
			t.Fatalf("expected business event %s/%s in catalog, got %+v", expected.eventType, expected.resource, catalog.Events)
		}
	}
	for _, item := range catalog.Events {
		if !item.Integration.HasTrigger || item.Integration.Status == "missing_trigger" {
			t.Fatalf("catalog event %s has no runtime trigger integration: %+v", item.EventType, item.Integration)
		}
		if item.Integration.ResultPolicy == "write_back" && !item.Integration.HasResultHandler {
			t.Fatalf("catalog event %s has no result write-back handler: %+v", item.EventType, item.Integration)
		}
		if item.Integration.Status == "missing_result_handler" {
			t.Fatalf("catalog event %s reports missing result handler: %+v", item.EventType, item.Integration)
		}
	}
	if event, ok := findWorkflowCatalogEvent(catalog.Events, "delivery_sign.completed", "delivery_sign"); !ok || !event.Integration.HasResultHandler || event.Integration.Status != "ready" {
		t.Fatalf("expected delivery sign catalog to report write-back integration, got %+v", event)
	}
	if event, ok := findWorkflowCatalogEvent(catalog.Events, "quality_exception.submitted", "quality_exception"); !ok || event.Integration.ResultPolicy != "event_only" || event.Integration.Status != "event_only" {
		t.Fatalf("expected quality exception submit catalog to report event-only integration, got %+v", event)
	}
	if !hasWorkflowCatalogEvent(catalog.Events, "ticket_void.requested", "ticket_void") {
		t.Fatalf("expected ticket void event in catalog, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogEvent(catalog.Events, "quality_exception.close_requested", "quality_exception") {
		t.Fatalf("expected quality exception close event in catalog, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogEvent(catalog.Events, "quality_exception.submitted", "quality_exception") {
		t.Fatalf("expected quality exception submitted event in catalog, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogEvent(catalog.Events, "delivery_sign.completed", "delivery_sign") {
		t.Fatalf("expected delivery sign event in catalog, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogEvent(catalog.Events, "mix_design_plant_profile.retire_requested", "mix_design_plant_profile") {
		t.Fatalf("expected plant profile retire event in catalog, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogResource(catalog.Resources, "raw_material_inspection") {
		t.Fatalf("expected raw material inspection resource in catalog, got %+v", catalog.Resources)
	}
	if !hasWorkflowCatalogVariable(catalog.Events, "sales_order.risk_detected", "riskFlags") {
		t.Fatalf("expected sales order catalog variables for form builder, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogTrigger(catalog.Events, "sales_order.risk_detected", http.MethodPost, "/api/orders") {
		t.Fatalf("expected sales order workflow trigger in catalog, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogTrigger(catalog.Events, "delivery_sign.completed", http.MethodPost, "/api/delivery/sign") {
		t.Fatalf("expected delivery sign workflow trigger in catalog, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogTrigger(catalog.Events, "supplier_statement.submitted", http.MethodPost, "/api/finance/supplier-statements") {
		t.Fatalf("expected supplier statement workflow trigger in catalog, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogTrigger(catalog.Events, "mix_design_plant_profile.retire_requested", http.MethodPost, "/api/laboratory/mix-design-plant-profiles/{id}/retire") {
		t.Fatalf("expected plant profile retire workflow trigger in catalog, got %+v", catalog.Events)
	}
	if !hasWorkflowCatalogOutboxEvent(catalog.OutboxEvents, "workflow.task_created") {
		t.Fatalf("expected workflow task created outbox event in catalog, got %+v", catalog.OutboxEvents)
	}
	if !hasWorkflowCatalogOutboxEvent(catalog.OutboxEvents, "workflow.result_failed") {
		t.Fatalf("expected workflow result failed outbox event in catalog, got %+v", catalog.OutboxEvents)
	}
	if !hasWorkflowCatalogOutboxPayload(catalog.OutboxEvents, "workflow.result_applied", "resourceId") {
		t.Fatalf("expected workflow result outbox payload fields in catalog, got %+v", catalog.OutboxEvents)
	}
	if !hasWorkflowCatalogOutboxPayload(catalog.OutboxEvents, "workflow.task_created", "triggerEventKey") {
		t.Fatalf("expected workflow outbox catalog to include trigger event identity fields, got %+v", catalog.OutboxEvents)
	}

	rec = testRequest(t, app, token, http.MethodGet, "/api/system/workflows", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("workflow overview status %d: %s", rec.Code, rec.Body.String())
	}
	var overview struct {
		Catalog WorkflowCatalog `json:"catalog"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode workflow overview catalog: %v", err)
	}
	if !hasWorkflowCatalogEvent(overview.Catalog.Events, "statement.confirm_requested", "statement") {
		t.Fatalf("expected statement event in overview catalog, got %+v", overview.Catalog.Events)
	}
	if !hasWorkflowCatalogOutboxEvent(overview.Catalog.OutboxEvents, "workflow.*") {
		t.Fatalf("expected outbox event catalog in overview, got %+v", overview.Catalog.OutboxEvents)
	}
}

func hasWorkflowCatalogEvent(items []WorkflowCatalogEvent, eventType string, resource string) bool {
	_, ok := findWorkflowCatalogEvent(items, eventType, resource)
	return ok
}

func findWorkflowCatalogEvent(items []WorkflowCatalogEvent, eventType string, resource string) (WorkflowCatalogEvent, bool) {
	for _, item := range items {
		if item.EventType == eventType && item.Resource == resource {
			return item, true
		}
	}
	return WorkflowCatalogEvent{}, false
}

func hasWorkflowCatalogVariable(items []WorkflowCatalogEvent, eventType string, key string) bool {
	for _, item := range items {
		if item.EventType != eventType {
			continue
		}
		for _, variable := range item.Variables {
			if variable.Key == key {
				return true
			}
		}
	}
	return false
}

func hasWorkflowCatalogTrigger(items []WorkflowCatalogEvent, eventType string, method string, path string) bool {
	for _, item := range items {
		if item.EventType != eventType {
			continue
		}
		for _, trigger := range item.Triggers {
			if trigger.Method == method && trigger.Path == path {
				return true
			}
		}
	}
	return false
}

func hasWorkflowCatalogOutboxEvent(items []WorkflowCatalogOutboxEvent, eventType string) bool {
	for _, item := range items {
		if item.EventType == eventType {
			return true
		}
	}
	return false
}

func hasWorkflowCatalogOutboxPayload(items []WorkflowCatalogOutboxEvent, eventType string, key string) bool {
	for _, item := range items {
		if item.EventType != eventType {
			continue
		}
		for _, field := range item.PayloadFields {
			if field.Key == key {
				return true
			}
		}
	}
	return false
}

func hasWorkflowCatalogResource(items []WorkflowCatalogResource, value string) bool {
	for _, item := range items {
		if item.Value == value {
			return true
		}
	}
	return false
}
