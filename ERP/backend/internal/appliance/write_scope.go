package appliance

import "fmt"

func writableSiteID(data AppData, user User, requested int64) (int64, error) {
	scoped := scopedData(data, user)
	if len(scoped.Sites) == 0 {
		return 0, fmt.Errorf("当前用户没有可写站点")
	}
	if requested == 0 {
		if len(scoped.Sites) == 1 {
			return scoped.Sites[0].ID, nil
		}
		return 0, fmt.Errorf("请选择站点")
	}
	for _, site := range scoped.Sites {
		if site.ID == requested {
			return requested, nil
		}
	}
	return 0, fmt.Errorf("无权操作该站点")
}

func scopedCustomer(data AppData, user User, id int64) (Customer, bool) {
	return findCustomer(scopedData(data, user), id)
}

func scopedProject(data AppData, user User, id int64) (Project, bool) {
	return findProject(scopedData(data, user), id)
}

func scopedMaterial(data AppData, user User, id int64) (Material, bool) {
	return findMaterial(scopedData(data, user), id)
}

func scopedSupplier(data AppData, user User, id int64) (Supplier, bool) {
	return findSupplier(scopedData(data, user), id)
}

func purchaseRequestSiteID(data AppData, requestID int64) int64 {
	for _, item := range data.PurchaseRequests {
		if item.ID == requestID {
			return item.SiteID
		}
	}
	return 0
}
