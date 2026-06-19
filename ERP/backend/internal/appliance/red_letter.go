package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

func (a *App) createRedLetterInfo(w http.ResponseWriter, r *http.Request, session Session) {
	var req struct {
		OriginalInvoiceID int64   `json:"originalInvoiceId"`
		Reason            string  `json:"reason"`
		Applicant         string  `json:"applicant"`
		Amount            float64 `json:"amount"`
		TaxAmount         float64 `json:"taxAmount"`
		Remark            string  `json:"remark"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid red letter info")
		return
	}
	if req.OriginalInvoiceID == 0 {
		writeError(w, http.StatusBadRequest, "原发票不能为空")
		return
	}
	var item RedLetterInfo
	err := a.store.Mutate(func(data *AppData) error {
		original, ok := findSalesInvoice(*data, req.OriginalInvoiceID)
		if !ok {
			return fmt.Errorf("原发票不存在")
		}
		if !userCanAccessInvoice(*data, session.User, original) {
			return fmt.Errorf("无权申请该发票红字信息表")
		}
		if err := validateOriginalInvoiceForRedLetter(original); err != nil {
			return err
		}
		if hasActiveRedLetterInfo(*data, original.ID) {
			return fmt.Errorf("该发票已有待处理红字信息表")
		}
		actor := fallback(req.Applicant, fallback(session.User.DisplayName, session.User.Username))
		item = newRedLetterInfoLocked(data, original, req.Reason, actor, "requested")
		if req.Amount != 0 {
			item.Amount = negative(req.Amount)
		}
		if req.TaxAmount != 0 {
			item.TaxAmount = negative(req.TaxAmount)
		}
		item.Remark = strings.TrimSpace(req.Remark)
		data.RedLetterInfos = append(data.RedLetterInfos, item)
		addAudit(data, session.User.Username, "create", "red_letter_info", item.ID, item.InfoNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "finance.red_letter.created")
}

func (a *App) approveRedLetterInfo(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		TaxControlNo string `json:"taxControlNo"`
		ApprovedBy   string `json:"approvedBy"`
		Remark       string `json:"remark"`
	}
	_ = readJSON(r, &req)
	var item RedLetterInfo
	err := a.store.Mutate(func(data *AppData) error {
		index := redLetterInfoIndex(*data, id)
		if index < 0 {
			return fmt.Errorf("红字信息表不存在")
		}
		current := data.RedLetterInfos[index]
		if current.Status == "used" {
			return fmt.Errorf("红字信息表已被红票使用")
		}
		original, ok := findSalesInvoice(*data, current.OriginalInvoiceID)
		if !ok {
			return fmt.Errorf("原发票不存在")
		}
		if !userCanAccessInvoice(*data, session.User, original) {
			return fmt.Errorf("无权审批该红字信息表")
		}
		if err := validateOriginalInvoiceForRedLetter(original); err != nil {
			return err
		}
		current.Status = "approved"
		current.ApprovedBy = fallback(req.ApprovedBy, fallback(session.User.DisplayName, session.User.Username))
		current.ApprovedAt = nowString()
		current.TaxControlNo = fallback(req.TaxControlNo, fallback(current.TaxControlNo, number("RLITAX", current.ID)))
		if req.Remark != "" {
			current.Remark = strings.TrimSpace(req.Remark)
		}
		data.RedLetterInfos[index] = current
		item = current
		addAudit(data, session.User.Username, "approve", "red_letter_info", item.ID, item.InfoNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "finance.red_letter.approved")
}

func ensureRedLetterInfoForRedOffset(data *AppData, original SalesInvoice, requestedID int64, reason, actor string) (RedLetterInfo, error) {
	if requestedID > 0 {
		index := redLetterInfoIndex(*data, requestedID)
		if index < 0 {
			return RedLetterInfo{}, fmt.Errorf("红字信息表不存在")
		}
		item := data.RedLetterInfos[index]
		if item.OriginalInvoiceID != original.ID {
			return RedLetterInfo{}, fmt.Errorf("红字信息表与原发票不匹配")
		}
		if item.RedInvoiceID != 0 || item.Status == "used" {
			return RedLetterInfo{}, fmt.Errorf("红字信息表已被红票使用")
		}
		if item.Status != "approved" {
			return RedLetterInfo{}, fmt.Errorf("红字信息表尚未审批")
		}
		return item, nil
	}
	for _, item := range data.RedLetterInfos {
		if item.OriginalInvoiceID == original.ID && item.Status == "approved" && item.RedInvoiceID == 0 {
			return item, nil
		}
		if item.OriginalInvoiceID == original.ID && item.Status == "requested" {
			return RedLetterInfo{}, fmt.Errorf("该发票已有待审批红字信息表")
		}
	}
	item := newRedLetterInfoLocked(data, original, reason, actor, "approved")
	item.ApprovedBy = actor
	item.ApprovedAt = nowString()
	item.TaxControlNo = number("RLITAX", item.ID)
	data.RedLetterInfos = append(data.RedLetterInfos, item)
	return item, nil
}

func markRedLetterInfoUsed(data *AppData, redInvoice SalesInvoice, taxControlNo string) {
	if redInvoice.RedLetterInfoID == 0 {
		return
	}
	for i := range data.RedLetterInfos {
		if data.RedLetterInfos[i].ID != redInvoice.RedLetterInfoID {
			continue
		}
		data.RedLetterInfos[i].Status = "used"
		data.RedLetterInfos[i].RedInvoiceID = redInvoice.ID
		data.RedLetterInfos[i].UsedAt = nowString()
		data.RedLetterInfos[i].TaxControlNo = fallback(taxControlNo, data.RedLetterInfos[i].TaxControlNo)
		return
	}
}

func newRedLetterInfoLocked(data *AppData, original SalesInvoice, reason, applicant, status string) RedLetterInfo {
	id := nextID(data, "redLetter")
	return RedLetterInfo{
		ID:                id,
		InfoNo:            number("RLI", id),
		OriginalInvoiceID: original.ID,
		OriginalInvoiceNo: original.InvoiceNo,
		CustomerID:        original.CustomerID,
		Amount:            negative(original.Amount),
		TaxAmount:         negative(original.TaxAmount),
		Reason:            fallback(reason, "客户红冲申请"),
		Applicant:         applicant,
		Status:            fallback(status, "requested"),
		RequestedAt:       nowString(),
	}
}

func validateOriginalInvoiceForRedLetter(invoice SalesInvoice) error {
	if invoice.InvoiceType == "red" {
		return fmt.Errorf("红票不能申请红字信息表")
	}
	if invoice.TaxStatus != "submitted" || invoice.TaxControlNo == "" {
		return fmt.Errorf("原发票尚未完成税控提交")
	}
	return nil
}

func hasActiveRedLetterInfo(data AppData, originalInvoiceID int64) bool {
	for _, item := range data.RedLetterInfos {
		if item.OriginalInvoiceID == originalInvoiceID && item.Status != "cancelled" && item.Status != "rejected" && item.Status != "used" {
			return true
		}
	}
	return false
}

func findSalesInvoice(data AppData, id int64) (SalesInvoice, bool) {
	for _, item := range data.SalesInvoices {
		if item.ID == id {
			return item, true
		}
	}
	return SalesInvoice{}, false
}

func redLetterInfoIndex(data AppData, id int64) int {
	for i := range data.RedLetterInfos {
		if data.RedLetterInfos[i].ID == id {
			return i
		}
	}
	return -1
}

func negative(value float64) float64 {
	if value < 0 {
		return round(value)
	}
	return round(-value)
}

func redInvoiceCategory(originalCategory string) string {
	if strings.HasPrefix(originalCategory, "red_") {
		return originalCategory
	}
	if originalCategory == "blue_e_invoice" {
		return "red_e_invoice"
	}
	if originalCategory == "blue_vat_normal" {
		return "red_vat_normal"
	}
	return "red_vat_special"
}

func dataDictionariesByType(data AppData, typ string) []DataDictionary {
	items := make([]DataDictionary, 0)
	for _, item := range data.DataDictionaries {
		if item.Type == typ && item.Status != "disabled" {
			items = append(items, item)
		}
	}
	return items
}
