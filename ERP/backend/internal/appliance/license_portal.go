package appliance

import (
	"sort"
	"strings"
	"time"
)

func buildLicensePortalOverview(data AppData) LicensePortalOverview {
	requiredModules := enabledModuleCodes(data.Modules)
	recordsByLicense := map[string]*LicensePortalCustomer{}
	packageCountByLicense := map[string]int{}
	latestDownloadByLicense := map[string]string{}
	latestOperationByLicense := map[string]string{}

	overview := LicensePortalOverview{
		KPIs: LicensePortalKPI{
			TotalPackages: len(data.LicensePackages),
			RiskLevel:     "normal",
		},
		Customers:       []LicensePortalCustomer{},
		RecentEvents:    licensePortalEvents(data.AuditLogs),
		RequiredModules: requiredModules,
	}

	for _, item := range data.LicensePackages {
		licenseID := strings.TrimSpace(item.LicenseID)
		if licenseID == "" {
			continue
		}
		packageCountByLicense[licenseID]++
		switch item.Status {
		case "active":
			overview.KPIs.ActivePackages++
		case "revoked":
			overview.KPIs.RevokedPackages++
		case "issued":
			overview.KPIs.IssuedPackages++
		}
		current, exists := recordsByLicense[licenseID]
		if !exists || item.ID > current.LatestPackageID {
			recordsByLicense[licenseID] = licensePortalCustomerFromPackage(item, requiredModules)
		}
	}

	for _, audit := range data.AuditLogs {
		if audit.Resource == "license_package" && audit.Action == "download" {
			overview.KPIs.DownloadCount++
			if audit.Detail != "" && audit.CreatedAt > latestDownloadByLicense[audit.Detail] {
				latestDownloadByLicense[audit.Detail] = audit.CreatedAt
			}
		}
		if isLicensePortalAudit(audit) && audit.Detail != "" && audit.CreatedAt > latestOperationByLicense[audit.Detail] {
			latestOperationByLicense[audit.Detail] = audit.Action + " · " + audit.CreatedAt
		}
	}

	for _, revocation := range data.LicenseRevocations {
		if revocation.Status != "active" {
			continue
		}
		customer, ok := recordsByLicense[revocation.LicenseID]
		if !ok {
			customer = &LicensePortalCustomer{LicenseID: revocation.LicenseID}
			recordsByLicense[revocation.LicenseID] = customer
		}
		customer.Revoked = true
		customer.Status = "revoked"
		customer.VerificationState = "revoked"
		customer.VerificationError = fallback(revocation.Reason, "revoked")
		customer.RiskLevel = "revoked"
		customer.RenewalAvailable = false
		if revocation.RevokedAt != "" {
			customer.LastOperation = "revoke · " + revocation.RevokedAt
		}
	}

	totalCoverage := 0.0
	for licenseID, customer := range recordsByLicense {
		customer.PackageCount = packageCountByLicense[licenseID]
		customer.LatestDownloadAt = latestDownloadByLicense[licenseID]
		if customer.LastOperation == "" {
			customer.LastOperation = latestOperationByLicense[licenseID]
		}
		if customer.DaysToExpire < 0 {
			overview.KPIs.ExpiredPackages++
		} else if customer.DaysToExpire <= 30 {
			overview.KPIs.Expiring30Days++
		}
		totalCoverage += customer.ModuleCoverage
		if licenseRiskRank(customer.RiskLevel) > licenseRiskRank(overview.KPIs.RiskLevel) {
			overview.KPIs.RiskLevel = customer.RiskLevel
		}
		overview.Customers = append(overview.Customers, *customer)
	}
	overview.KPIs.TotalCustomers = len(overview.Customers)
	if overview.KPIs.TotalCustomers > 0 {
		overview.KPIs.ModuleCoverage = round(totalCoverage / float64(overview.KPIs.TotalCustomers))
	}
	sort.Slice(overview.Customers, func(i, j int) bool {
		if licenseRiskRank(overview.Customers[i].RiskLevel) != licenseRiskRank(overview.Customers[j].RiskLevel) {
			return licenseRiskRank(overview.Customers[i].RiskLevel) > licenseRiskRank(overview.Customers[j].RiskLevel)
		}
		if overview.Customers[i].ExpiresAt != overview.Customers[j].ExpiresAt {
			return overview.Customers[i].ExpiresAt < overview.Customers[j].ExpiresAt
		}
		return overview.Customers[i].LicenseID < overview.Customers[j].LicenseID
	})
	return overview
}

func licensePortalCustomerFromPackage(item LicensePackage, requiredModules []string) *LicensePortalCustomer {
	daysToExpire := licenseDaysToExpire(item.ExpiresAt)
	revoked := item.Status == "revoked"
	risk := licensePortalRisk(item.Status, daysToExpire, item.LastVerificationState, item.LastVerificationError)
	moduleCoverage := licenseModuleCoverage(item.Modules, requiredModules)
	return &LicensePortalCustomer{
		CustomerName:         item.CustomerName,
		Watermark:            item.Watermark,
		LicenseID:            item.LicenseID,
		Edition:              item.Edition,
		Status:               item.Status,
		ExpiresAt:            item.ExpiresAt,
		DaysToExpire:         daysToExpire,
		Modules:              item.Modules,
		ModuleCount:          len(item.Modules),
		ModuleCoverage:       moduleCoverage,
		MaxSites:             item.MaxSites,
		MaxVehicles:          item.MaxVehicles,
		LatestPackageID:      item.ID,
		LatestIssuedAt:       item.IssuedAt,
		LatestActivatedAt:    item.ActivatedAt,
		RenewalAvailable:     !revoked && daysToExpire >= 0,
		Revoked:              revoked,
		VerificationState:    fallback(item.LastVerificationState, "pending"),
		VerificationError:    item.LastVerificationError,
		RiskLevel:            risk,
		PublicKeyFingerprint: item.PublicKeyFingerprint,
	}
}

func licensePortalEvents(audits []AuditLog) []LicensePortalEvent {
	events := []LicensePortalEvent{}
	for i := len(audits) - 1; i >= 0; i-- {
		audit := audits[i]
		if !isLicensePortalAudit(audit) {
			continue
		}
		events = append(events, LicensePortalEvent{
			Action:    audit.Action,
			Resource:  audit.Resource,
			LicenseID: audit.Detail,
			Detail:    audit.Detail,
			Actor:     audit.User,
			IP:        audit.IP,
			CreatedAt: audit.CreatedAt,
		})
		if len(events) == 12 {
			break
		}
	}
	return events
}

func isLicensePortalAudit(audit AuditLog) bool {
	switch audit.Resource {
	case "license", "license_package":
		return true
	default:
		return false
	}
}

func licenseDaysToExpire(expiresAt string) int {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(expiresAt))
	if err != nil {
		return -9999
	}
	today := time.Now().Truncate(24 * time.Hour)
	return int(parsed.Sub(today).Hours() / 24)
}

func licensePortalRisk(status string, daysToExpire int, verificationState, verificationError string) string {
	if status == "revoked" || verificationState == "revoked" {
		return "revoked"
	}
	if daysToExpire < 0 {
		return "expired"
	}
	if strings.TrimSpace(verificationError) != "" && verificationState != "valid" {
		return "warning"
	}
	if daysToExpire <= 30 {
		return "warning"
	}
	return "normal"
}

func licenseModuleCoverage(modules, required []string) float64 {
	if len(required) == 0 {
		return 100
	}
	moduleSet := map[string]bool{}
	for _, module := range modules {
		moduleSet[module] = true
	}
	matched := 0
	for _, module := range required {
		if moduleSet[module] {
			matched++
		}
	}
	return percent(matched, len(required))
}

func licenseRiskRank(value string) int {
	switch value {
	case "revoked":
		return 4
	case "expired":
		return 3
	case "warning":
		return 2
	default:
		return 1
	}
}
