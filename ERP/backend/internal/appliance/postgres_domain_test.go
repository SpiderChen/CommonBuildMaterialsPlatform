package appliance

import (
	"strings"
	"testing"
)

func TestDomainRowsRoundTripTopLevelAppDataResources(t *testing.T) {
	data := SeedData()
	rows, err := domainRowsFromAppData(data, "checksum-test")
	if err != nil {
		t.Fatalf("domain rows: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("expected domain rows")
	}
	if got, want := domainRowCount(data), len(rows); got != want {
		t.Fatalf("domain row count mismatch: got %d want %d", got, want)
	}

	seen := map[string]int{}
	for _, row := range rows {
		if row.Payload == nil || row.RowID == "" || row.Checksum != "checksum-test" {
			t.Fatalf("invalid domain row: %+v", row)
		}
		seen[row.Resource]++
	}
	for _, resource := range []string{"schemaVersion", "license", "next", "modules", "users", "orders", "salesInvoices", "taxGatewaySubmissions", "inventory", "auditLogs"} {
		if seen[resource] == 0 {
			t.Fatalf("expected domain resource %s in rows", resource)
		}
	}

	restored, err := appDataFromDomainRows(rows)
	if err != nil {
		t.Fatalf("restore domain rows: %v", err)
	}
	if restored.SchemaVersion != data.SchemaVersion || restored.License.LicenseID != data.License.LicenseID {
		t.Fatalf("restored singleton mismatch: %+v", restored.License)
	}
	if len(restored.Modules) != len(data.Modules) || len(restored.Orders) != len(data.Orders) || len(restored.Inventory) != len(data.Inventory) {
		t.Fatalf("restored core resource counts mismatch: modules=%d/%d orders=%d/%d inventory=%d/%d",
			len(restored.Modules), len(data.Modules), len(restored.Orders), len(data.Orders), len(restored.Inventory), len(data.Inventory))
	}
	if len(restored.SalesInvoices) != len(data.SalesInvoices) || restored.SalesInvoices[0].InvoiceType != "blue" {
		t.Fatalf("restored invoices mismatch: %+v", restored.SalesInvoices)
	}
	if len(restored.TaxGatewaySubmissions) != len(data.TaxGatewaySubmissions) || restored.TaxGatewaySubmissions[0].Action != "issue" {
		t.Fatalf("restored tax submissions mismatch: %+v", restored.TaxGatewaySubmissions)
	}
	if restored.Next["invoice"] != data.Next["invoice"] || restored.Next["taxSubmission"] != data.Next["taxSubmission"] {
		t.Fatalf("restored Next map mismatch: %+v", restored.Next)
	}
}

func TestPostgresDomainSchemaIncludesReadWriteTables(t *testing.T) {
	schema := postgresDomainSchemaSQL()
	for _, fragment := range []string{
		"create table if not exists cbmp_domain_rows",
		"primary key (resource, row_id)",
		"create table if not exists cbmp_domain_status",
		"idx_cbmp_domain_rows_resource_ordinal",
	} {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("domain schema missing %q", fragment)
		}
	}
}
