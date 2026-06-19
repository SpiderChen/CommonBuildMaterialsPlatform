package ops

import "testing"

func TestRenewCustomerUpdatesLicenseAndAudit(t *testing.T) {
	data := SeedData()
	renewal, err := RenewCustomer(&data, 3, RenewLicenseRequest{
		ExpiresAt: "2027-06-30", Edition: "Enterprise Appliance", MaxSites: 16, MaxVehicles: 2400, Operator: "ops", Note: "已完成年度续费",
	})
	if err != nil {
		t.Fatalf("renew customer: %v", err)
	}
	if renewal.OldExpiresAt != "2026-06-10" || renewal.NewExpiresAt != "2027-06-30" {
		t.Fatalf("unexpected renewal: %+v", renewal)
	}
	customer := data.Customers[findCustomerIndex(data, 3)]
	if customer.RenewalStatus != "active" || customer.MaxSites != 16 || customer.MaxVehicles != 2400 {
		t.Fatalf("customer was not updated: %+v", customer)
	}
	if len(data.AuditLogs) == 0 || data.AuditLogs[0].Action != "license.renewed" {
		t.Fatalf("missing audit log: %+v", data.AuditLogs)
	}
}

func TestAlertLifecycle(t *testing.T) {
	data := SeedData()
	alert, err := AcknowledgeAlert(&data, 1, "交付运维")
	if err != nil {
		t.Fatalf("ack alert: %v", err)
	}
	if alert.Status != "acknowledged" || alert.AcknowledgedAt == "" {
		t.Fatalf("unexpected acknowledged alert: %+v", alert)
	}
	alert, err = ResolveAlert(&data, 1, ResolveAlertRequest{Operator: "交付运维", Resolution: "已完成服务端重启和补丁发布"})
	if err != nil {
		t.Fatalf("resolve alert: %v", err)
	}
	if alert.Status != "resolved" || alert.ResolvedAt == "" || alert.Resolution == "" {
		t.Fatalf("unexpected resolved alert: %+v", alert)
	}
}

func TestPackagePublishAndAssign(t *testing.T) {
	data := SeedData()
	pkg, err := CreatePackage(&data, CreateUpdatePackageRequest{
		Target: "server", Version: "1.4.4", FileName: "cbmp-appliance-1.4.4.tar.gz", Checksum: "sha256:test", RolloutPct: 50,
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	pkg, err = PublishPackage(&data, pkg.ID)
	if err != nil {
		t.Fatalf("publish package: %v", err)
	}
	if pkg.Status != "published" || pkg.PublishedAt == "" {
		t.Fatalf("unexpected package: %+v", pkg)
	}
	assignments, err := AssignPackage(&data, pkg.ID, AssignUpdatePackageRequest{CustomerIDs: []int64{1, 2}})
	if err != nil {
		t.Fatalf("assign package: %v", err)
	}
	if len(assignments) != 2 {
		t.Fatalf("expected two assignments, got %+v", assignments)
	}
}
