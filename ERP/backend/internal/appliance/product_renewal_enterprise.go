package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type productRenewalApprovalRequest struct {
	Action       string  `json:"action"`
	ApprovalID   int64   `json:"approvalId"`
	QuoteID      int64   `json:"quoteId"`
	ContractID   int64   `json:"contractId"`
	ApprovalType string  `json:"approvalType"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	CurrentRole  string  `json:"currentRole"`
	Comment      string  `json:"comment"`
}

type productRenewalESignRequest struct {
	Action     string `json:"action"`
	SignID     int64  `json:"signId"`
	ContractID int64  `json:"contractId"`
	Signer     string `json:"signer"`
	Phone      string `json:"phone"`
	Channel    string `json:"channel"`
	Signature  string `json:"signature"`
	Remark     string `json:"remark"`
}

func (a *App) productOpsRenewalApproval(w http.ResponseWriter, r *http.Request, session Session, taskIDPart string) {
	taskID, _ := strconv.ParseInt(taskIDPart, 10, 64)
	var req productRenewalApprovalRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid renewal approval payload")
		return
	}
	action := fallback(strings.ToLower(strings.TrimSpace(req.Action)), "submit")
	var approval ProductRenewalApproval
	err := a.store.Mutate(func(data *AppData) error {
		taskIndex := productRenewalTaskIndex(*data, taskID)
		if taskIndex < 0 {
			return fmt.Errorf("续费任务不存在")
		}
		task := &data.ProductRenewalTasks[taskIndex]
		switch action {
		case "submit":
			approval = buildProductRenewalApproval(data, task, req, session.User.DisplayName)
			if approval.Amount <= 0 {
				return fmt.Errorf("续费审批金额必须大于 0")
			}
			task.Stage = "续费审批中"
			task.Status = "open"
			task.Amount = approval.Amount
			task.Currency = approval.Currency
			task.LastContactAt = approval.RequestedAt
			productRenewalSyncInstanceStage(data, task.InstanceID, task.Stage)
			data.ProductRenewalApprovals = append(data.ProductRenewalApprovals, approval)
			addAudit(data, session.User.Username, "submit", "renewal_approval", approval.ID, approval.ApprovalNo+" "+approval.CustomerName, clientIP(r))
			return nil
		case "approve", "reject":
			index := latestProductRenewalApprovalIndex(*data, task.ID, req.ApprovalID, "pending")
			if index < 0 {
				return fmt.Errorf("没有待处理的续费审批")
			}
			now := nowString()
			approval = data.ProductRenewalApprovals[index]
			approval.Status = map[bool]string{true: "approved", false: "rejected"}[action == "approve"]
			approval.ApprovedBy = session.User.DisplayName
			approval.ApprovedAt = now
			approval.Comment = fallback(strings.TrimSpace(req.Comment), approval.Comment)
			data.ProductRenewalApprovals[index] = approval
			task.LastContactAt = now
			if action == "approve" {
				task.Stage = "审批通过"
				if approval.QuoteID != 0 {
					productRenewalApproveQuote(data, task.ID, approval.QuoteID, session.User.DisplayName, now)
				}
				if approval.ContractID != 0 {
					productRenewalApproveContract(data, task.ID, approval.ContractID)
				}
			} else {
				task.Stage = "审批驳回"
			}
			productRenewalSyncInstanceStage(data, task.InstanceID, task.Stage)
			addAudit(data, session.User.Username, action, "renewal_approval", approval.ID, approval.ApprovalNo+" "+approval.CustomerName, clientIP(r))
			return nil
		default:
			return fmt.Errorf("续费审批动作必须是 submit、approve 或 reject")
		}
	})
	a.respondMutation(w, err, approval, "product_ops.renewal.approval.changed")
}

func (a *App) productOpsRenewalInvoice(w http.ResponseWriter, r *http.Request, session Session, taskIDPart string) {
	taskID, _ := strconv.ParseInt(taskIDPart, 10, 64)
	var req ProductRenewalInvoice
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid renewal invoice payload")
		return
	}
	var invoice ProductRenewalInvoice
	err := a.store.Mutate(func(data *AppData) error {
		taskIndex := productRenewalTaskIndex(*data, taskID)
		if taskIndex < 0 {
			return fmt.Errorf("续费任务不存在")
		}
		task := &data.ProductRenewalTasks[taskIndex]
		contractIndex := latestProductRenewalContractIndex(*data, task.ID, req.ContractID)
		if contractIndex < 0 {
			return fmt.Errorf("请先确认续费合同")
		}
		contract := data.ProductRenewalContracts[contractIndex]
		paymentIndex := latestProductRenewalPaymentIndex(*data, task.ID, contract.ID, req.PaymentID)
		var payment ProductRenewalPayment
		if paymentIndex >= 0 {
			payment = data.ProductRenewalPayments[paymentIndex]
		}
		amount := req.Amount
		if amount <= 0 && payment.ID != 0 {
			amount = payment.Amount
		}
		if amount <= 0 {
			amount = productRenewalPaidAmount(*data, contract.ID)
		}
		if amount <= 0 {
			amount = contract.Amount
		}
		if amount <= 0 {
			return fmt.Errorf("续费开票金额必须大于 0")
		}
		now := nowString()
		id := nextID(data, "renewalInvoice")
		taxRate := req.TaxRate
		if taxRate <= 0 {
			taxRate = 0.06
		}
		invoiceNo := fallback(strings.TrimSpace(req.InvoiceNo), number("RI", id))
		invoice = ProductRenewalInvoice{
			ID:              id,
			InvoiceNo:       invoiceNo,
			TaskID:          task.ID,
			ContractID:      contract.ID,
			PaymentID:       payment.ID,
			InstanceID:      task.InstanceID,
			CustomerName:    task.CustomerName,
			LicenseID:       task.LicenseID,
			Amount:          amount,
			TaxRate:         taxRate,
			TaxAmount:       round(amount * taxRate),
			InvoiceType:     fallback(strings.TrimSpace(req.InvoiceType), "blue_e_invoice"),
			Status:          fallback(strings.TrimSpace(req.Status), "issued"),
			TaxStatus:       fallback(strings.TrimSpace(req.TaxStatus), "accepted"),
			FileURL:         fallback(strings.TrimSpace(req.FileURL), "renewal-invoice://"+invoiceNo+".pdf"),
			CreatedBy:       session.User.DisplayName,
			CreatedAt:       now,
			IssuedAt:        fallback(strings.TrimSpace(req.IssuedAt), now),
			ExternalRequest: fallback(strings.TrimSpace(req.ExternalRequest), "local-tax-"+invoiceNo),
			Remark:          strings.TrimSpace(req.Remark),
		}
		data.ProductRenewalInvoices = append(data.ProductRenewalInvoices, invoice)
		_, _ = enqueueProductRenewalSyncRecord(data, productRenewalSyncRequest{
			Scenario:     "tax",
			ResourceType: "invoice",
			ResourceID:   invoice.ID,
			ResourceNo:   invoice.InvoiceNo,
			Task:         *task,
			Action:       "issue",
			Payload: map[string]interface{}{
				"invoice":  invoice,
				"contract": contract,
				"payment":  payment,
				"task":     *task,
			},
		})
		for i := range data.ProductRenewalInvoices {
			if data.ProductRenewalInvoices[i].ID == invoice.ID {
				invoice = data.ProductRenewalInvoices[i]
				break
			}
		}
		task.Stage = "已开票"
		task.LastContactAt = now
		productRenewalSyncInstanceStage(data, task.InstanceID, task.Stage)
		addAudit(data, session.User.Username, "invoice", "renewal_task", task.ID, invoice.InvoiceNo+" "+invoice.CustomerName, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, invoice, "product_ops.renewal.invoice.created")
}

func (a *App) productOpsRenewalESign(w http.ResponseWriter, r *http.Request, session Session, taskIDPart string) {
	taskID, _ := strconv.ParseInt(taskIDPart, 10, 64)
	var req productRenewalESignRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid renewal e-sign payload")
		return
	}
	action := fallback(strings.ToLower(strings.TrimSpace(req.Action)), "send")
	var sign ProductRenewalESign
	err := a.store.Mutate(func(data *AppData) error {
		taskIndex := productRenewalTaskIndex(*data, taskID)
		if taskIndex < 0 {
			return fmt.Errorf("续费任务不存在")
		}
		task := &data.ProductRenewalTasks[taskIndex]
		contractIndex := latestProductRenewalContractIndex(*data, task.ID, req.ContractID)
		if contractIndex < 0 {
			return fmt.Errorf("请先确认续费合同")
		}
		contract := &data.ProductRenewalContracts[contractIndex]
		now := nowString()
		switch action {
		case "send":
			id := nextID(data, "renewalESign")
			signNo := number("RS", id)
			sign = ProductRenewalESign{
				ID:           id,
				SignNo:       signNo,
				TaskID:       task.ID,
				ContractID:   contract.ID,
				InstanceID:   task.InstanceID,
				CustomerName: task.CustomerName,
				LicenseID:    task.LicenseID,
				Signer:       fallback(strings.TrimSpace(req.Signer), task.CustomerName+" 经办人"),
				Phone:        strings.TrimSpace(req.Phone),
				Channel:      fallback(strings.TrimSpace(req.Channel), "local_esign"),
				Status:       "sent",
				LinkURL:      "/public/renewal-sign/" + signNo,
				SentBy:       session.User.DisplayName,
				SentAt:       now,
				Remark:       strings.TrimSpace(req.Remark),
			}
			contract.Status = "signing"
			task.Stage = "电子签待签"
			task.LastContactAt = now
			productRenewalSyncInstanceStage(data, task.InstanceID, task.Stage)
			data.ProductRenewalESigns = append(data.ProductRenewalESigns, sign)
			_, _ = enqueueProductRenewalSyncRecord(data, productRenewalSyncRequest{
				Scenario:     "esign",
				ResourceType: "esign",
				ResourceID:   sign.ID,
				ResourceNo:   sign.SignNo,
				Task:         *task,
				Action:       "send",
				Code:         sign.Channel,
				Payload: map[string]interface{}{
					"esign":    sign,
					"contract": *contract,
					"task":     *task,
				},
			})
			for i := range data.ProductRenewalESigns {
				if data.ProductRenewalESigns[i].ID == sign.ID {
					sign = data.ProductRenewalESigns[i]
					break
				}
			}
			addAudit(data, session.User.Username, "send", "renewal_esign", sign.ID, sign.SignNo+" "+sign.CustomerName, clientIP(r))
			return nil
		case "complete":
			index := latestProductRenewalESignIndex(*data, task.ID, contract.ID, req.SignID, "sent")
			if index < 0 {
				return fmt.Errorf("没有待签署的续费电子签")
			}
			sign = data.ProductRenewalESigns[index]
			sign.Status = "signed"
			sign.SignedAt = now
			sign.Signature = fallback(strings.TrimSpace(req.Signature), sign.Signer+" 电子签名")
			sign.Remark = fallback(strings.TrimSpace(req.Remark), sign.Remark)
			data.ProductRenewalESigns[index] = sign
			contract.Status = "signed"
			contract.SignedBy = sign.Signer
			contract.SignedAt = now
			task.Stage = "合同已签"
			task.LastContactAt = now
			productRenewalSyncInstanceStage(data, task.InstanceID, task.Stage)
			addAudit(data, session.User.Username, "complete", "renewal_esign", sign.ID, sign.SignNo+" "+sign.CustomerName, clientIP(r))
			return nil
		default:
			return fmt.Errorf("续费电子签动作必须是 send 或 complete")
		}
	})
	a.respondMutation(w, err, sign, "product_ops.renewal.esign.changed")
}

func buildProductRenewalApproval(data *AppData, task *ProductRenewalTask, req productRenewalApprovalRequest, actor string) ProductRenewalApproval {
	now := nowString()
	id := nextID(data, "renewalApproval")
	quoteIndex := latestProductRenewalQuoteIndex(*data, task.ID, req.QuoteID)
	contractIndex := latestProductRenewalContractIndex(*data, task.ID, req.ContractID)
	var quote ProductRenewalQuote
	var contract ProductRenewalContract
	if quoteIndex >= 0 {
		quote = data.ProductRenewalQuotes[quoteIndex]
	}
	if contractIndex >= 0 {
		contract = data.ProductRenewalContracts[contractIndex]
	}
	amount := req.Amount
	if amount <= 0 {
		amount = nonZero(contract.Amount, quote.Amount)
	}
	if amount <= 0 {
		amount = task.Amount
	}
	return ProductRenewalApproval{
		ID:           id,
		ApprovalNo:   number("RA", id),
		TaskID:       task.ID,
		QuoteID:      nonZeroInt(req.QuoteID, quote.ID),
		ContractID:   nonZeroInt(req.ContractID, contract.ID),
		InstanceID:   task.InstanceID,
		CustomerName: task.CustomerName,
		LicenseID:    task.LicenseID,
		ApprovalType: fallback(strings.TrimSpace(req.ApprovalType), productRenewalApprovalType(quote.ID, contract.ID)),
		Amount:       amount,
		Currency:     fallback(strings.TrimSpace(req.Currency), fallback(task.Currency, "CNY")),
		Status:       "pending",
		CurrentRole:  fallback(strings.TrimSpace(req.CurrentRole), "boss"),
		RequestedBy:  actor,
		RequestedAt:  now,
		Comment:      strings.TrimSpace(req.Comment),
	}
}

func productRenewalApprovalType(quoteID, contractID int64) string {
	if contractID != 0 {
		return "contract"
	}
	if quoteID != 0 {
		return "quote"
	}
	return "renewal"
}

func latestProductRenewalApprovalIndex(data AppData, taskID, approvalID int64, status string) int {
	index := -1
	for i, approval := range data.ProductRenewalApprovals {
		if approvalID > 0 {
			if approval.ID == approvalID && approval.TaskID == taskID {
				return i
			}
			continue
		}
		if approval.TaskID != taskID {
			continue
		}
		if status != "" && approval.Status != status {
			continue
		}
		if index < 0 || approval.RequestedAt > data.ProductRenewalApprovals[index].RequestedAt {
			index = i
		}
	}
	return index
}

func latestProductRenewalPaymentIndex(data AppData, taskID, contractID, paymentID int64) int {
	index := -1
	for i, payment := range data.ProductRenewalPayments {
		if paymentID > 0 {
			if payment.ID == paymentID && payment.TaskID == taskID {
				return i
			}
			continue
		}
		if payment.TaskID != taskID || (contractID != 0 && payment.ContractID != contractID) {
			continue
		}
		if index < 0 || payment.CreatedAt > data.ProductRenewalPayments[index].CreatedAt {
			index = i
		}
	}
	return index
}

func latestProductRenewalESignIndex(data AppData, taskID, contractID, signID int64, status string) int {
	index := -1
	for i, sign := range data.ProductRenewalESigns {
		if signID > 0 {
			if sign.ID == signID && sign.TaskID == taskID {
				return i
			}
			continue
		}
		if sign.TaskID != taskID || (contractID != 0 && sign.ContractID != contractID) {
			continue
		}
		if status != "" && sign.Status != status {
			continue
		}
		if index < 0 || sign.SentAt > data.ProductRenewalESigns[index].SentAt {
			index = i
		}
	}
	return index
}

func productRenewalApproveQuote(data *AppData, taskID, quoteID int64, actor, now string) {
	for i := range data.ProductRenewalQuotes {
		if data.ProductRenewalQuotes[i].TaskID == taskID && data.ProductRenewalQuotes[i].ID == quoteID {
			data.ProductRenewalQuotes[i].Status = "approved"
			data.ProductRenewalQuotes[i].ApprovedBy = actor
			data.ProductRenewalQuotes[i].ApprovedAt = now
			return
		}
	}
}

func productRenewalApproveContract(data *AppData, taskID, contractID int64) {
	for i := range data.ProductRenewalContracts {
		if data.ProductRenewalContracts[i].TaskID == taskID && data.ProductRenewalContracts[i].ID == contractID && data.ProductRenewalContracts[i].Status == "pending_approval" {
			data.ProductRenewalContracts[i].Status = "approved"
			return
		}
	}
}
