package appliance

import (
	"fmt"
	"net/http"
	"time"
)

func (a *App) createSupplierStatement(w http.ResponseWriter, r *http.Request, session Session) {
	var item SupplierStatement
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid supplier statement")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if _, ok := findSupplier(*data, item.SupplierID); !ok {
			return fmt.Errorf("供应商不存在")
		}
		item.ID = nextID(data, "supplierStatement")
		item.StatementNo = number("SS", item.ID)
		item.Period = fallback(item.Period, periodString())
		item.Amount = nonZero(item.Amount, unreconciledPayableAmount(*data, item.SupplierID))
		if item.Amount <= 0 {
			return fmt.Errorf("供应商暂无可对账应付")
		}
		item.Status = fallback(item.Status, "submitted")
		data.SupplierStatements = append(data.SupplierStatements, item)
		attachSupplierPayables(data, item, false)
		if event, instances, err := publishSupplierStatementWorkflow(data, item, session.User.Username); err != nil {
			return err
		} else if event.Status == "handled" || len(instances) > 0 {
			for i := range data.SupplierStatements {
				if data.SupplierStatements[i].ID == item.ID {
					data.SupplierStatements[i].Status = "pending_approval"
					item = data.SupplierStatements[i]
					break
				}
			}
		}
		addAudit(data, session.User.Username, "create", "supplier_statement", item.ID, item.StatementNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "finance.supplier_statement.created")
}

func (a *App) approveSupplierStatement(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item SupplierStatement
	err := a.store.Mutate(func(data *AppData) error {
		idx := -1
		for i := range data.SupplierStatements {
			if data.SupplierStatements[i].ID == id {
				idx = i
				break
			}
		}
		if idx < 0 {
			return fmt.Errorf("供应商对账单不存在")
		}
		if hasPendingWorkflowForResource(*data, "supplier_statement", id) {
			return fmt.Errorf("供应商对账单正在工作流审批中，请在工作流中处理")
		}
		if data.SupplierStatements[idx].Status == "approved" {
			item = data.SupplierStatements[idx]
			return nil
		}
		approved, err := approveSupplierStatementLocked(data, id, session.User.DisplayName)
		if err != nil {
			return err
		}
		item = approved
		addAudit(data, session.User.Username, "approve", "supplier_statement", item.ID, item.StatementNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "finance.supplier_statement.approved")
}

func publishSupplierStatementWorkflow(data *AppData, item SupplierStatement, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "supplier_statement.submitted",
		Source:     "finance",
		EventKey:   "supplier_statement:" + item.StatementNo,
		Resource:   "supplier_statement",
		ResourceID: item.ID,
		ResourceNo: item.StatementNo,
		Title:      "供应商对账单 " + item.StatementNo,
		Actor:      actor,
		Reason:     "供应商对账确认",
		Variables: map[string]string{
			"supplierId": fmt.Sprintf("%d", item.SupplierID),
			"period":     item.Period,
			"amount":     fmt.Sprintf("%.2f", item.Amount),
		},
	})
}

func approveSupplierStatementLocked(data *AppData, id int64, actor string) (SupplierStatement, error) {
	idx := -1
	for i := range data.SupplierStatements {
		if data.SupplierStatements[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return SupplierStatement{}, fmt.Errorf("供应商对账单不存在")
	}
	if data.SupplierStatements[idx].Status == "approved" {
		return data.SupplierStatements[idx], nil
	}
	data.SupplierStatements[idx].Status = "approved"
	data.SupplierStatements[idx].ApprovedBy = actor
	data.SupplierStatements[idx].ApprovedAt = nowString()
	item := data.SupplierStatements[idx]
	attached := attachSupplierPayables(data, item, true)
	if attached == 0 {
		payableID := nextID(data, "payable")
		data.Payables = append(data.Payables, Payable{
			ID: payableID, BillNo: number("AP", payableID), SupplierID: item.SupplierID, SupplierStatementID: item.ID,
			Amount: item.Amount, DueDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"), Status: "confirmed",
		})
	}
	return item, nil
}

func hasPendingWorkflowForResource(data AppData, resource string, resourceID int64) bool {
	for _, instance := range data.WorkflowInstances {
		if instance.Resource == resource && instance.ResourceID == resourceID && instance.Status == "pending" {
			return true
		}
	}
	return false
}

func unreconciledPayableAmount(data AppData, supplierID int64) float64 {
	total := 0.0
	for _, item := range data.Payables {
		if item.SupplierID == supplierID && item.SupplierStatementID == 0 && item.Status != "paid" {
			total += item.Amount - item.PaidAmount
		}
	}
	return round(total)
}

func attachSupplierPayables(data *AppData, statement SupplierStatement, confirm bool) int {
	count := 0
	remaining := statement.Amount
	for i := range data.Payables {
		if remaining <= 0 {
			break
		}
		if data.Payables[i].SupplierID != statement.SupplierID || data.Payables[i].Status == "paid" {
			continue
		}
		if data.Payables[i].SupplierStatementID != 0 && data.Payables[i].SupplierStatementID != statement.ID {
			continue
		}
		data.Payables[i].SupplierStatementID = statement.ID
		if confirm && data.Payables[i].Status == "open" {
			data.Payables[i].Status = "confirmed"
		}
		remaining = round(remaining - (data.Payables[i].Amount - data.Payables[i].PaidAmount))
		count++
	}
	return count
}
