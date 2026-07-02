package appliance

import "reflect"

func purgeDemoBusinessSeed(data *AppData, seed AppData) bool {
	if !demoSeedPayloadMatches(*data, seed) {
		return false
	}
	license := data.License
	menuLabels := data.MenuLabels
	admin := initialAdminUser(data.Users)
	admin.ID = 1
	admin.DisplayName = builtinSuperAdminRoleName
	admin.RoleCode = builtinSuperAdminRoleCode
	admin.Status = "active"
	admin.TenantID = 0
	admin.CompanyID = 0
	admin.SiteID = 0
	admin.CustomerID = 0
	admin.DriverID = 0

	next := InitialData()
	next.License = license
	next.Users = []User{admin}
	if len(menuLabels) > 0 {
		next.MenuLabels = menuLabels
	}
	*data = next
	return true
}

func demoSeedPayloadMatches(data AppData, seed AppData) bool {
	skip := map[string]bool{
		"SchemaVersion":    true,
		"License":          true,
		"Modules":          true,
		"Plugins":          true,
		"Users":            true,
		"DataDictionaries": true,
		"GatewayRoutes":    true,
		"MenuLabels":       true,
		"Next":             true,
	}
	left := reflect.ValueOf(data)
	right := reflect.ValueOf(seed)
	for i := 0; i < left.NumField(); i++ {
		field := left.Type().Field(i).Name
		if skip[field] {
			continue
		}
		if !reflect.DeepEqual(left.Field(i).Interface(), right.Field(i).Interface()) {
			return false
		}
	}
	return true
}
