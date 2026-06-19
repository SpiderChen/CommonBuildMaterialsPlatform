package appliance

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

type PortalOverview struct {
	Dispatches      []DispatchOrder          `json:"dispatches"`
	Orders          []SalesOrder             `json:"orders"`
	Statements      []Statement              `json:"statements"`
	Invoices        []SalesInvoice           `json:"invoices"`
	Signs           []DeliverySign           `json:"signs"`
	SignLinks       []DeliverySignLink       `json:"signLinks"`
	SignAttachments []DeliverySignAttachment `json:"signAttachments"`
	Alarms          []VehicleAlarm           `json:"alarms"`
	Complaints      []CustomerComplaint      `json:"complaints"`
}

func (a *App) portal(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "overview" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, buildPortalOverview(scopedData(a.mustSnapshot(), session.User), session.User))
		return
	}
	if len(parts) == 1 && parts[0] == "complaints" {
		if r.Method == http.MethodGet {
			data := scopedData(a.mustSnapshot(), session.User)
			writeJSON(w, http.StatusOK, complaintsWithSLAStatus(data.CustomerComplaints, time.Now()))
			return
		}
		if r.Method == http.MethodPost {
			a.createPortalComplaint(w, r, session)
			return
		}
	}
	if len(parts) == 3 && parts[0] == "dispatches" && parts[2] == "exception" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.reportPortalDispatchException(w, r, session, id)
		return
	}
	writeError(w, http.StatusNotFound, "unknown portal route")
}

func buildPortalOverview(data AppData, user User) PortalOverview {
	if user.RoleCode == "driver" {
		customerIDs := map[int64]bool{}
		for _, order := range data.Orders {
			customerIDs[order.CustomerID] = true
		}
		for _, sign := range data.DeliverySigns {
			customerIDs[sign.CustomerID] = true
		}
		data.Statements = filter(data.Statements, func(item Statement) bool { return customerIDs[item.CustomerID] })
		data.SalesInvoices = filter(data.SalesInvoices, func(item SalesInvoice) bool { return customerIDs[item.CustomerID] })
		data.CustomerComplaints = nil
	}
	slices.SortFunc(data.DispatchOrders, func(a, b DispatchOrder) int { return strings.Compare(b.DispatchNo, a.DispatchNo) })
	slices.SortFunc(data.Orders, func(a, b SalesOrder) int { return strings.Compare(b.OrderNo, a.OrderNo) })
	slices.SortFunc(data.Statements, func(a, b Statement) int { return strings.Compare(b.StatementNo, a.StatementNo) })
	slices.SortFunc(data.SalesInvoices, func(a, b SalesInvoice) int { return strings.Compare(b.InvoiceNo, a.InvoiceNo) })
	return PortalOverview{
		Dispatches:      data.DispatchOrders,
		Orders:          data.Orders,
		Statements:      data.Statements,
		Invoices:        data.SalesInvoices,
		Signs:           data.DeliverySigns,
		SignLinks:       data.DeliverySignLinks,
		SignAttachments: data.DeliverySignAttachments,
		Alarms:          data.VehicleAlarms,
		Complaints:      complaintsWithSLAStatus(data.CustomerComplaints, time.Now()),
	}
}

func dispatchIndex(data AppData, id int64) int {
	for i := range data.DispatchOrders {
		if data.DispatchOrders[i].ID == id {
			return i
		}
	}
	return -1
}

func (a *App) createPortalComplaint(w http.ResponseWriter, r *http.Request, session Session) {
	var item CustomerComplaint
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid portal complaint")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		customerID := session.User.CustomerID
		if customerID == 0 {
			customerID = item.CustomerID
		}
		if customerID == 0 {
			return fmt.Errorf("客户不能为空")
		}
		customer, ok := findCustomer(*data, customerID)
		if !ok {
			return fmt.Errorf("客户不存在")
		}
		item.CustomerID = customer.ID
		if item.ProjectID == 0 {
			for _, project := range data.Projects {
				if project.CustomerID == customer.ID {
					item.ProjectID = project.ID
					break
				}
			}
		}
		if item.ProjectID != 0 {
			project, ok := findProject(*data, item.ProjectID)
			if !ok {
				return fmt.Errorf("项目不存在")
			}
			if project.CustomerID != customer.ID {
				return fmt.Errorf("项目不属于当前客户")
			}
		}
		item.ID = nextID(data, "customerComplaint")
		item.ComplaintNo = number("CP", item.ID)
		item.Title = fallback(strings.TrimSpace(item.Title), "客户自助投诉")
		item.Content = strings.TrimSpace(item.Content)
		item.Level = fallback(item.Level, "medium")
		item.Status = "open"
		item.Owner = fallback(item.Owner, "客服主管")
		item.CreatedAt = nowString()
		applyComplaintSLAOnCreate(&item, time.Now())
		data.CustomerComplaints = append(data.CustomerComplaints, item)
		addAudit(data, session.User.Username, "create", "portal_customer_complaint", item.ID, item.Title, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "portal.customer_complaint.created")
}

func (a *App) reportPortalDispatchException(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Exception string `json:"exception"`
		Level     string `json:"level"`
		AlarmType string `json:"alarmType"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid dispatch exception")
		return
	}
	var updated DispatchOrder
	var forbidden bool
	err := a.store.Mutate(func(data *AppData) error {
		index := dispatchIndex(*data, id)
		if index < 0 {
			return fmt.Errorf("派车单不存在")
		}
		dispatch := data.DispatchOrders[index]
		switch session.User.RoleCode {
		case "driver":
			if session.User.DriverID == 0 || dispatch.DriverID != session.User.DriverID {
				forbidden = true
				return fmt.Errorf("无权上报该派车单异常")
			}
		case "boss", "dispatcher":
		default:
			forbidden = true
			return fmt.Errorf("当前角色不能上报派车异常")
		}
		exception := fallback(strings.TrimSpace(req.Exception), "司机上报现场异常")
		level := fallback(strings.TrimSpace(req.Level), "medium")
		alarmType := fallback(strings.TrimSpace(req.AlarmType), "driver_exception")
		data.DispatchOrders[index].Exception = exception
		data.DispatchOrders[index].UpdatedAt = nowString()
		updated = data.DispatchOrders[index]
		appendAlarm(data, updated.VehicleID, updated.ID, alarmType, level, fmt.Sprintf("%s：%s", updated.DispatchNo, exception))
		addAudit(data, session.User.Username, "report_exception", "dispatch_order", updated.ID, exception, clientIP(r))
		return nil
	})
	if forbidden {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	a.respondMutation(w, err, updated, "dispatch.order.update")
}
