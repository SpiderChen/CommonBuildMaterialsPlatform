package appliance

func InitialData() AppData {
	seed := SeedData()
	admin := initialAdminUser(seed.Users)
	admin.CompanyID = 0
	admin.SiteID = 0
	admin.CustomerID = 0
	admin.DriverID = 0

	data := AppData{
		SchemaVersion: seed.SchemaVersion,
		License: LicenseInfo{
			CustomerName:          "待导入授权客户",
			Edition:               "ERP Appliance",
			LastVerificationState: "missing",
			LastVerificationError: "待导入客户授权包",
		},
		Modules:          seed.Modules,
		Plugins:          seed.Plugins,
		Roles:            []Role{builtinSuperAdminRole(1)},
		Users:            []User{admin},
		GroupProfile:     unconfiguredGroupProfile(),
		SecurityPolicies: seed.SecurityPolicies,
		DataDictionaries: seed.DataDictionaries,
		GatewayRoutes:    seed.GatewayRoutes,
	}
	data.Next = initialDataCounters(data)
	return data
}

func unconfiguredGroupProfile() GroupProfile {
	return GroupProfile{
		Name:             "待配置集团",
		Code:             "UNCONFIGURED",
		Edition:          "ERP Appliance",
		OperatingMode:    "private_deployment",
		DataArchitecture: "single_customer",
	}
}

func initialAdminUser(users []User) User {
	for _, user := range users {
		if user.Username == builtinAdminUsername {
			user.ID = 1
			user.DisplayName = builtinSuperAdminRoleName
			user.RoleCode = builtinSuperAdminRoleCode
			user.Status = "active"
			return user
		}
	}
	salt, hash, status := seedUserCredential(builtinAdminUsername)
	return User{ID: 1, Username: builtinAdminUsername, DisplayName: builtinSuperAdminRoleName, RoleCode: builtinSuperAdminRoleCode, PasswordSalt: salt, PasswordHash: hash, Status: status}
}

func initialDataCounters(data AppData) map[string]int64 {
	counters := map[string]int64{
		"role": 1,
		"user": 1,
	}
	for _, item := range data.DataDictionaries {
		if counters["dict"] < item.ID {
			counters["dict"] = item.ID
		}
	}
	for _, item := range data.SecurityPolicies {
		if counters["securityPolicy"] < item.ID {
			counters["securityPolicy"] = item.ID
		}
	}
	for _, item := range data.GatewayRoutes {
		if counters["gatewayRoute"] < item.ID {
			counters["gatewayRoute"] = item.ID
		}
	}
	return counters
}
