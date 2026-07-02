package appliance

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (a *App) createCustomerContact(w http.ResponseWriter, r *http.Request, session Session) {
	var item CustomerContact
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer contact")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if err := normalizeCustomerContact(*data, &item, nil); err != nil {
			return err
		}
		item.ID = nextID(data, "contact")
		if item.IsDefault {
			clearDefaultCustomerContact(data, item.CustomerID)
			updateCustomerPrimaryContact(data, item.CustomerID, item.Name, item.Phone)
		}
		data.CustomerContacts = append(data.CustomerContacts, item)
		addAudit(data, session.User.Username, "create", "customer_contact", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.customer_contact.created")
}

func normalizeCustomerContact(data AppData, item *CustomerContact, current *CustomerContact) error {
	if current != nil && item.CustomerID == 0 {
		item.CustomerID = current.CustomerID
	}
	if _, ok := findCustomer(data, item.CustomerID); !ok {
		return fmt.Errorf("客户不存在")
	}
	item.Name = strings.TrimSpace(item.Name)
	item.Phone = strings.TrimSpace(item.Phone)
	item.Role = strings.TrimSpace(item.Role)
	item.Status = strings.TrimSpace(item.Status)
	if current != nil {
		item.Name = fallback(item.Name, current.Name)
		item.Phone = fallback(item.Phone, current.Phone)
		item.Role = fallback(item.Role, current.Role)
		item.Status = fallback(item.Status, current.Status)
		item.IsDefault = item.IsDefault || current.IsDefault
	}
	if item.Name == "" {
		return fmt.Errorf("联系人姓名不能为空")
	}
	if item.Phone == "" {
		return fmt.Errorf("联系人电话不能为空")
	}
	item.Role = fallback(item.Role, "业务联系人")
	item.Status = fallback(item.Status, "active")
	if item.IsDefault && item.Status != "active" {
		return fmt.Errorf("非启用联系人不能设为默认")
	}
	return nil
}

func (a *App) setDefaultCustomerContact(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item CustomerContact
	err := a.store.Mutate(func(data *AppData) error {
		index := customerContactIndex(*data, id)
		if index < 0 {
			return fmt.Errorf("客户联系人不存在")
		}
		if data.CustomerContacts[index].Status != "active" {
			return fmt.Errorf("非启用联系人不能设为默认")
		}
		customerID := data.CustomerContacts[index].CustomerID
		clearDefaultCustomerContact(data, customerID)
		data.CustomerContacts[index].IsDefault = true
		item = data.CustomerContacts[index]
		updateCustomerPrimaryContact(data, customerID, item.Name, item.Phone)
		addAudit(data, session.User.Username, "set_default", "customer_contact", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.customer_contact.default")
}

func (a *App) createCustomerBlacklist(w http.ResponseWriter, r *http.Request, session Session) {
	var item CustomerBlacklist
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer blacklist")
		return
	}
	topic := "master.customer_blacklist.created"
	err := a.store.Mutate(func(data *AppData) error {
		customer, ok := findCustomer(*data, item.CustomerID)
		if !ok {
			return fmt.Errorf("客户不存在")
		}
		if activeOrPendingCustomerBlacklist(*data, item.CustomerID, true) {
			return fmt.Errorf("客户已有生效黑名单")
		}
		item.ID = nextID(data, "customerBlacklist")
		item.CustomerName = fallback(item.CustomerName, customer.Name)
		item.Reason = fallback(item.Reason, "客户风险停供")
		item.Scope = fallback(item.Scope, "sales_order")
		item.Severity = fallback(item.Severity, "high")
		if !item.BlockOrders && !item.BlockDispatch {
			item.BlockOrders = true
		}
		item.Status = "pending_approval"
		item.CreatedAt = nowString()
		item.Actor = session.User.Username
		data.CustomerBlacklists = append(data.CustomerBlacklists, item)
		_, instances, err := publishCustomerBlacklistWorkflow(data, item, session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			topic = "master.customer_blacklist.workflow_requested"
			addAudit(data, session.User.Username, "request_create", "customer_blacklist", item.ID, item.CustomerName, clientIP(r))
			return nil
		}
		activated, err := applyCustomerBlacklistActivationLocked(data, item.ID)
		if err != nil {
			return err
		}
		item = activated
		addAudit(data, session.User.Username, "create", "customer_blacklist", item.ID, item.CustomerName, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, topic)
}

func publishCustomerBlacklistWorkflow(data *AppData, item CustomerBlacklist, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "customer_blacklist.submitted",
		Source:     "master",
		Resource:   "customer_blacklist",
		ResourceID: item.ID,
		ResourceNo: item.CustomerName,
		Title:      "客户黑名单审批",
		Actor:      actor,
		Reason:     fallback(item.Reason, "客户风险停供"),
		Variables: map[string]string{
			"customerId":    fmt.Sprintf("%d", item.CustomerID),
			"customerName":  item.CustomerName,
			"scope":         item.Scope,
			"severity":      item.Severity,
			"blockOrders":   fmt.Sprintf("%t", item.BlockOrders),
			"blockDispatch": fmt.Sprintf("%t", item.BlockDispatch),
			"reason":        item.Reason,
		},
	})
}

func applyCustomerBlacklistActivationLocked(data *AppData, id int64) (CustomerBlacklist, error) {
	index := customerBlacklistIndex(*data, id)
	if index < 0 {
		return CustomerBlacklist{}, fmt.Errorf("客户黑名单不存在")
	}
	data.CustomerBlacklists[index].Status = "active"
	setCustomerStatus(data, data.CustomerBlacklists[index].CustomerID, "blocked")
	return data.CustomerBlacklists[index], nil
}

func (a *App) releaseCustomerBlacklist(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var response interface{}
	topic := "master.customer_blacklist.released"
	err := a.store.Mutate(func(data *AppData) error {
		index := customerBlacklistIndex(*data, id)
		if index < 0 {
			return fmt.Errorf("客户黑名单不存在")
		}
		item := data.CustomerBlacklists[index]
		_, instances, err := publishCustomerBlacklistReleaseWorkflow(data, item, session.User.Username)
		if err != nil {
			return err
		}
		if len(instances) > 0 {
			response = instances[0]
			topic = "master.customer_blacklist.release.workflow_requested"
			addAudit(data, session.User.Username, "request_release", "customer_blacklist", item.ID, item.CustomerName, clientIP(r))
			return nil
		}
		released, err := applyCustomerBlacklistReleaseLocked(data, item.ID)
		if err != nil {
			return err
		}
		response = released
		addAudit(data, session.User.Username, "release", "customer_blacklist", released.ID, released.CustomerName, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, response, topic)
}

func publishCustomerBlacklistReleaseWorkflow(data *AppData, item CustomerBlacklist, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "customer_blacklist_release.requested",
		Source:     "master",
		Resource:   "customer_blacklist_release",
		ResourceID: item.ID,
		ResourceNo: item.CustomerName,
		Title:      "客户黑名单解除",
		Actor:      actor,
		Reason:     fallback(item.Reason, "客户风险解除"),
		Variables: map[string]string{
			"customerId":    fmt.Sprintf("%d", item.CustomerID),
			"customerName":  item.CustomerName,
			"scope":         item.Scope,
			"severity":      item.Severity,
			"blockOrders":   fmt.Sprintf("%t", item.BlockOrders),
			"blockDispatch": fmt.Sprintf("%t", item.BlockDispatch),
			"reason":        item.Reason,
		},
	})
}

func applyCustomerBlacklistReleaseLocked(data *AppData, id int64) (CustomerBlacklist, error) {
	index := customerBlacklistIndex(*data, id)
	if index < 0 {
		return CustomerBlacklist{}, fmt.Errorf("客户黑名单不存在")
	}
	data.CustomerBlacklists[index].Status = "released"
	data.CustomerBlacklists[index].ReleasedAt = nowString()
	item := data.CustomerBlacklists[index]
	if !activeCustomerBlacklist(*data, item.CustomerID, true) {
		setCustomerStatus(data, item.CustomerID, "active")
	}
	return item, nil
}

func (a *App) createCustomerProfile(w http.ResponseWriter, r *http.Request, session Session) {
	var item CustomerProfile
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer profile")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		customer, ok := findCustomer(*data, item.CustomerID)
		if !ok {
			return fmt.Errorf("客户不存在")
		}
		item.CustomerName = fallback(item.CustomerName, customer.Name)
		item.Grade = fallback(strings.ToUpper(strings.TrimSpace(item.Grade)), "B")
		item.RiskLevel = fallback(item.RiskLevel, "medium")
		if item.CreditScore == 0 {
			item.CreditScore = 75
		}
		item.Status = fallback(item.Status, "active")
		item.UpdatedAt = nowString()
		item.Actor = session.User.Username
		if index := customerProfileIndex(*data, item.CustomerID); index >= 0 {
			item.ID = data.CustomerProfiles[index].ID
			data.CustomerProfiles[index] = item
		} else {
			item.ID = nextID(data, "customerProfile")
			data.CustomerProfiles = append(data.CustomerProfiles, item)
		}
		addAudit(data, session.User.Username, "upsert", "customer_profile", item.ID, item.CustomerName, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.customer_profile.saved")
}

func (a *App) evaluateCustomerProfiles(w http.ResponseWriter, r *http.Request, session Session) {
	var profiles []CustomerProfile
	err := a.store.Mutate(func(data *AppData) error {
		for _, customer := range data.Customers {
			profile := evaluateCustomerProfile(*data, customer, session.User.Username)
			if index := customerProfileIndex(*data, customer.ID); index >= 0 {
				profile.ID = data.CustomerProfiles[index].ID
				data.CustomerProfiles[index] = profile
			} else {
				profile.ID = nextID(data, "customerProfile")
				data.CustomerProfiles = append(data.CustomerProfiles, profile)
			}
			profiles = append(profiles, profile)
		}
		addAudit(data, session.User.Username, "evaluate", "customer_profiles", 0, fmt.Sprintf("%d profiles", len(profiles)), clientIP(r))
		return nil
	})
	a.respondMutation(w, err, profiles, "master.customer_profiles.evaluated")
}

func (a *App) createCustomerComplaint(w http.ResponseWriter, r *http.Request, session Session) {
	var item CustomerComplaint
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer complaint")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if _, ok := findCustomer(*data, item.CustomerID); !ok {
			return fmt.Errorf("客户不存在")
		}
		if item.ProjectID > 0 {
			if _, ok := findProject(*data, item.ProjectID); !ok {
				return fmt.Errorf("项目不存在")
			}
		}
		item.ID = nextID(data, "customerComplaint")
		item.ComplaintNo = number("CP", item.ID)
		item.Title = fallback(strings.TrimSpace(item.Title), "客户投诉")
		item.Level = fallback(item.Level, "medium")
		item.Status = "open"
		item.Owner = fallback(item.Owner, fallback(session.User.DisplayName, session.User.Username))
		applyComplaintSLAOnCreate(&item, time.Now())
		data.CustomerComplaints = append(data.CustomerComplaints, item)
		addAudit(data, session.User.Username, "create", "customer_complaint", item.ID, item.Title, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.customer_complaint.created")
}

func (a *App) closeCustomerComplaint(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Resolution string `json:"resolution"`
	}
	_ = readJSON(r, &req)
	var item CustomerComplaint
	err := a.store.Mutate(func(data *AppData) error {
		index := customerComplaintIndex(*data, id)
		if index < 0 {
			return fmt.Errorf("客户投诉不存在")
		}
		closeComplaintWithSLA(&data.CustomerComplaints[index], time.Now())
		data.CustomerComplaints[index].Resolution = fallback(req.Resolution, "已处理并回访")
		item = data.CustomerComplaints[index]
		addAudit(data, session.User.Username, "close", "customer_complaint", item.ID, item.Resolution, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.customer_complaint.closed")
}

func applyComplaintSLAOnCreate(item *CustomerComplaint, now time.Time) {
	if item.SLAHours <= 0 {
		item.SLAHours = defaultComplaintSLAHours(item.Level)
	}
	item.CreatedAt = fallback(item.CreatedAt, now.Format("2006-01-02 15:04:05"))
	if item.DueAt == "" {
		item.DueAt = now.Add(time.Duration(item.SLAHours) * time.Hour).Format("2006-01-02 15:04:05")
	}
	item.SLAStatus = complaintSLAStatus(*item, now)
}

func closeComplaintWithSLA(item *CustomerComplaint, now time.Time) {
	item.Status = "closed"
	item.ClosedAt = now.Format("2006-01-02 15:04:05")
	if isComplaintOverdue(*item, now) {
		item.SLAStatus = "breached"
		item.OverdueHours = complaintOverdueHours(*item, now)
		return
	}
	item.SLAStatus = "met"
	item.OverdueHours = 0
}

func complaintsWithSLAStatus(items []CustomerComplaint, now time.Time) []CustomerComplaint {
	out := append([]CustomerComplaint(nil), items...)
	for i := range out {
		if out[i].Status == "closed" {
			continue
		}
		if out[i].SLAHours <= 0 {
			out[i].SLAHours = defaultComplaintSLAHours(out[i].Level)
		}
		if out[i].DueAt == "" {
			out[i].DueAt = complaintCreatedAt(out[i]).Add(time.Duration(out[i].SLAHours) * time.Hour).Format("2006-01-02 15:04:05")
		}
		out[i].SLAStatus = complaintSLAStatus(out[i], now)
		out[i].OverdueHours = complaintOverdueHours(out[i], now)
	}
	return out
}

func defaultComplaintSLAHours(level string) int {
	switch strings.TrimSpace(level) {
	case "critical":
		return 2
	case "high":
		return 4
	case "low":
		return 72
	default:
		return 24
	}
}

func complaintSLAStatus(item CustomerComplaint, now time.Time) string {
	if item.Status == "closed" && item.SLAStatus != "" {
		return item.SLAStatus
	}
	if isComplaintOverdue(item, now) {
		return "overdue"
	}
	return "on_track"
}

func isComplaintOverdue(item CustomerComplaint, now time.Time) bool {
	due, ok := parseComplaintTime(item.DueAt)
	return ok && now.After(due)
}

func complaintOverdueHours(item CustomerComplaint, now time.Time) int {
	due, ok := parseComplaintTime(item.DueAt)
	if !ok || !now.After(due) {
		return 0
	}
	hours := int(now.Sub(due).Hours())
	if hours < 1 {
		return 1
	}
	return hours
}

func complaintCreatedAt(item CustomerComplaint) time.Time {
	if created, ok := parseComplaintTime(item.CreatedAt); ok {
		return created
	}
	return time.Now()
}

func parseComplaintTime(value string) (time.Time, bool) {
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(value), time.Local)
	return parsed, err == nil
}

func (a *App) createContractAttachment(w http.ResponseWriter, r *http.Request, session Session, contractID int64) {
	var item ContractAttachment
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid contract attachment")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		contract, ok := findContract(*data, contractID)
		if !ok {
			return fmt.Errorf("合同不存在")
		}
		item.ID = nextID(data, "contractAttachment")
		item.ContractID = contract.ID
		item.CustomerID = contract.CustomerID
		item.FileName = strings.TrimSpace(item.FileName)
		item.URL = strings.TrimSpace(item.URL)
		if item.FileName == "" || item.URL == "" {
			return fmt.Errorf("合同附件名称和 URL 必填")
		}
		item.FileType = fallback(strings.TrimSpace(item.FileType), "contract_pdf")
		item.Checksum = strings.TrimSpace(item.Checksum)
		item.Status = fallback(item.Status, "active")
		item.UploadedBy = fallback(session.User.DisplayName, session.User.Username)
		item.UploadedAt = nowString()
		data.ContractAttachments = append(data.ContractAttachments, item)
		addAudit(data, session.User.Username, "upload", "contract_attachment", item.ID, item.FileName, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "contract.attachment.created")
}

func orderBlockedByCustomerRisk(data AppData, customer Customer) (CustomerBlacklist, bool) {
	if customer.Status == "blocked" || customer.Status == "suspended" || customer.Status == "blacklisted" {
		if item, ok := findActiveCustomerBlacklist(data, customer.ID, true); ok {
			return item, true
		}
		return CustomerBlacklist{CustomerID: customer.ID, CustomerName: customer.Name, Reason: "客户已停供", BlockOrders: true, Status: customer.Status}, true
	}
	return findActiveCustomerBlacklist(data, customer.ID, true)
}

func evaluateCustomerProfile(data AppData, customer Customer, actor string) CustomerProfile {
	score := 100
	tags := []string{}
	if customer.CreditLimit > 0 {
		ratio := customer.Receivable / customer.CreditLimit
		if ratio >= 1 {
			score -= 35
			tags = append(tags, "超授信")
		} else if ratio >= 0.8 {
			score -= 20
			tags = append(tags, "授信占用高")
		} else if ratio >= 0.5 {
			score -= 10
			tags = append(tags, "授信占用中")
		} else {
			tags = append(tags, "回款健康")
		}
	}
	if activeCustomerBlacklist(data, customer.ID, false) {
		score -= 40
		tags = append(tags, "黑名单")
	}
	openComplaints := 0
	for _, complaint := range data.CustomerComplaints {
		if complaint.CustomerID == customer.ID && complaint.Status == "open" {
			openComplaints++
		}
	}
	if openComplaints > 0 {
		score -= openComplaints * 5
		tags = append(tags, fmt.Sprintf("%d 个未关闭投诉", openComplaints))
	}
	if score < 0 {
		score = 0
	}
	grade := "A"
	risk := "low"
	switch {
	case score < 55:
		grade = "D"
		risk = "critical"
	case score < 70:
		grade = "C"
		risk = "high"
	case score < 85:
		grade = "B"
		risk = "medium"
	}
	return CustomerProfile{
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		Grade:        grade,
		RiskLevel:    risk,
		CreditScore:  score,
		Tags:         tags,
		Status:       "active",
		UpdatedAt:    nowString(),
		Actor:        actor,
	}
}

func findActiveCustomerBlacklist(data AppData, customerID int64, ordersOnly bool) (CustomerBlacklist, bool) {
	for _, item := range data.CustomerBlacklists {
		if item.CustomerID != customerID || item.Status != "active" {
			continue
		}
		if ordersOnly && !item.BlockOrders {
			continue
		}
		return item, true
	}
	return CustomerBlacklist{}, false
}

func activeCustomerBlacklist(data AppData, customerID int64, ordersOnly bool) bool {
	_, ok := findActiveCustomerBlacklist(data, customerID, ordersOnly)
	return ok
}

func activeOrPendingCustomerBlacklist(data AppData, customerID int64, ordersOnly bool) bool {
	for _, item := range data.CustomerBlacklists {
		if item.CustomerID != customerID || (item.Status != "active" && item.Status != "pending_approval") {
			continue
		}
		if ordersOnly && !item.BlockOrders {
			continue
		}
		return true
	}
	return false
}

func customerContactIndex(data AppData, id int64) int {
	for i := range data.CustomerContacts {
		if data.CustomerContacts[i].ID == id {
			return i
		}
	}
	return -1
}

func customerBlacklistIndex(data AppData, id int64) int {
	for i := range data.CustomerBlacklists {
		if data.CustomerBlacklists[i].ID == id {
			return i
		}
	}
	return -1
}

func customerProfileIndex(data AppData, customerID int64) int {
	for i := range data.CustomerProfiles {
		if data.CustomerProfiles[i].CustomerID == customerID {
			return i
		}
	}
	return -1
}

func customerComplaintIndex(data AppData, id int64) int {
	for i := range data.CustomerComplaints {
		if data.CustomerComplaints[i].ID == id {
			return i
		}
	}
	return -1
}

func clearDefaultCustomerContact(data *AppData, customerID int64) {
	for i := range data.CustomerContacts {
		if data.CustomerContacts[i].CustomerID == customerID {
			data.CustomerContacts[i].IsDefault = false
		}
	}
}

func syncCustomerPrimaryContact(data *AppData, customerID int64) {
	selected := -1
	for i := range data.CustomerContacts {
		if data.CustomerContacts[i].CustomerID == customerID && data.CustomerContacts[i].Status == "active" && data.CustomerContacts[i].IsDefault {
			selected = i
			break
		}
	}
	if selected < 0 {
		for i := range data.CustomerContacts {
			if data.CustomerContacts[i].CustomerID == customerID && data.CustomerContacts[i].Status == "active" {
				selected = i
				break
			}
		}
	}
	for i := range data.CustomerContacts {
		if data.CustomerContacts[i].CustomerID == customerID {
			data.CustomerContacts[i].IsDefault = i == selected && data.CustomerContacts[i].Status == "active"
		}
	}
	for i := range data.Customers {
		if data.Customers[i].ID != customerID {
			continue
		}
		if selected >= 0 {
			data.Customers[i].Contact = data.CustomerContacts[selected].Name
			data.Customers[i].Phone = data.CustomerContacts[selected].Phone
			return
		}
		data.Customers[i].Contact = ""
		data.Customers[i].Phone = ""
		return
	}
}

func findContract(data AppData, id int64) (Contract, bool) {
	for _, item := range data.Contracts {
		if item.ID == id {
			return item, true
		}
	}
	return Contract{}, false
}

func updateCustomerPrimaryContact(data *AppData, customerID int64, name, phone string) {
	for i := range data.Customers {
		if data.Customers[i].ID == customerID {
			data.Customers[i].Contact = fallback(name, data.Customers[i].Contact)
			data.Customers[i].Phone = fallback(phone, data.Customers[i].Phone)
			return
		}
	}
}

func setCustomerStatus(data *AppData, customerID int64, status string) {
	for i := range data.Customers {
		if data.Customers[i].ID == customerID {
			data.Customers[i].Status = status
			return
		}
	}
}
