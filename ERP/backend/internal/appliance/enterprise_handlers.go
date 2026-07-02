package appliance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (a *App) procurement(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 || parts[0] == "overview" {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"requests":       data.PurchaseRequests,
			"orders":         data.PurchaseOrders,
			"receipts":       data.RawMaterialReceipts,
			"flows":          data.InventoryFlows,
			"inventory":      data.Inventory,
			"stockYards":     data.StockYards,
			"stockYardPiles": data.StockYardPiles,
			"stockYardFlows": data.StockYardFlows,
			"transfers":      data.InventoryTransfers,
			"stocktakes":     data.InventoryStocktakes,
			"traces":         data.InventoryBatchTraces,
			"suppliers":      data.Suppliers,
		})
		return
	}
	if r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		switch parts[0] {
		case "requests":
			writeJSON(w, http.StatusOK, data.PurchaseRequests)
		case "orders":
			writeJSON(w, http.StatusOK, data.PurchaseOrders)
		case "receipts":
			writeJSON(w, http.StatusOK, data.RawMaterialReceipts)
		case "flows":
			writeJSON(w, http.StatusOK, data.InventoryFlows)
		case "traces":
			writeJSON(w, http.StatusOK, data.InventoryBatchTraces)
		case "transfers":
			writeJSON(w, http.StatusOK, data.InventoryTransfers)
		case "stocktakes":
			writeJSON(w, http.StatusOK, data.InventoryStocktakes)
		case "stock-yards":
			writeJSON(w, http.StatusOK, data.StockYards)
		case "stock-yard-piles":
			writeJSON(w, http.StatusOK, data.StockYardPiles)
		case "stock-yard-flows":
			writeJSON(w, http.StatusOK, data.StockYardFlows)
		default:
			writeError(w, http.StatusNotFound, "unknown procurement resource")
		}
		return
	}
	switch parts[0] {
	case "requests":
		a.createPurchaseRequest(w, r, session)
	case "orders":
		a.createPurchaseOrder(w, r, session)
	case "receipts":
		a.createRawMaterialReceipt(w, r, session)
	case "transfers":
		if len(parts) == 3 && parts[2] == "complete" {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			a.completeInventoryTransfer(w, r, session, id)
			return
		}
		a.createInventoryTransfer(w, r, session)
	case "stocktakes":
		if len(parts) == 3 && parts[2] == "review" {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			a.reviewInventoryStocktake(w, r, session, id)
			return
		}
		a.createInventoryStocktake(w, r, session)
	case "yard-receipts":
		a.createStockYardReceipt(w, r, session)
	case "yard-adjustments":
		a.createStockYardAdjustment(w, r, session)
	default:
		writeError(w, http.StatusNotFound, "unknown procurement route")
	}
}

func (a *App) createPurchaseRequest(w http.ResponseWriter, r *http.Request, session Session) {
	var item PurchaseRequest
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid purchase request")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if _, ok := findMaterial(*data, item.MaterialID); !ok {
			return fmt.Errorf("物料不存在")
		}
		item.ID = nextID(data, "purchaseRequest")
		item.RequestNo = number("PR", item.ID)
		item.Unit = fallback(item.Unit, "t")
		item.Status = fallback(item.Status, "submitted")
		item.CreatedAt = nowString()
		data.PurchaseRequests = append(data.PurchaseRequests, item)
		addAudit(data, session.User.Username, "create", "purchase_request", item.ID, item.RequestNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "procurement.request.created")
}

func (a *App) createPurchaseOrder(w http.ResponseWriter, r *http.Request, session Session) {
	var item PurchaseOrder
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid purchase order")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if _, ok := findSupplier(*data, item.SupplierID); !ok {
			return fmt.Errorf("供应商不存在")
		}
		if _, ok := findMaterial(*data, item.MaterialID); !ok {
			return fmt.Errorf("物料不存在")
		}
		item.ID = nextID(data, "purchaseOrder")
		item.OrderNo = number("PO", item.ID)
		item.Unit = fallback(item.Unit, "t")
		item.Status = fallback(item.Status, "approved")
		item.CreatedAt = nowString()
		data.PurchaseOrders = append(data.PurchaseOrders, item)
		addAudit(data, session.User.Username, "create", "purchase_order", item.ID, item.OrderNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "procurement.order.created")
}

func (a *App) createRawMaterialReceipt(w http.ResponseWriter, r *http.Request, session Session) {
	var item RawMaterialReceipt
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid raw material receipt")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		var requestSiteID int64
		if item.PurchaseOrderID != 0 {
			if po, ok := findPurchaseOrder(*data, item.PurchaseOrderID); ok {
				item.SupplierID = nonZeroInt(item.SupplierID, po.SupplierID)
				item.MaterialID = nonZeroInt(item.MaterialID, po.MaterialID)
				requestSiteID = purchaseRequestSiteID(*data, po.RequestID)
				item.SiteID = nonZeroInt(item.SiteID, requestSiteID)
			}
		}
		var err error
		item.SiteID, err = writableSiteID(*data, session.User, item.SiteID)
		if err != nil {
			return err
		}
		if requestSiteID != 0 && item.SiteID != requestSiteID {
			return fmt.Errorf("入库站点必须与采购申请站点一致")
		}
		if _, ok := scopedSupplier(*data, session.User, item.SupplierID); !ok {
			return fmt.Errorf("供应商不存在")
		}
		if _, ok := scopedMaterial(*data, session.User, item.MaterialID); !ok {
			return fmt.Errorf("物料不存在")
		}
		item.ID = nextID(data, "receipt")
		item.ReceiptNo = number("RI", item.ID)
		item.PlateNo = fallback(item.PlateNo, fmt.Sprintf("原料车-%s", item.ReceiptNo))
		item.NetWeight = round(item.GrossWeight - item.TareWeight)
		if item.NetWeight <= 0 {
			return fmt.Errorf("净重必须大于 0")
		}
		item.QualityStatus = fallback(item.QualityStatus, "pending")
		item.Status = fallback(item.Status, "received")
		item.CreatedAt = nowString()
		ticket := appendRawMaterialTicket(data, item)
		item.TicketID = ticket.ID
		data.RawMaterialReceipts = append(data.RawMaterialReceipts, item)
		balance := increaseInventoryWithLot(data, item.SiteID, item.MaterialID, item.SupplierID, item.NetWeight, item.ReceiptNo, item.ID)
		data.InventoryFlows = append(data.InventoryFlows, InventoryFlow{
			ID: nextID(data, "inventoryFlow"), FlowNo: number("IF", data.Next["inventoryFlow"]),
			SiteID: item.SiteID, MaterialID: item.MaterialID, SourceType: "raw_material_receipt",
			SourceID: item.ID, Direction: "in", Quantity: item.NetWeight, BalanceQty: balance,
			Remark: "原料入厂过磅入库", CreatedAt: nowString(),
		})
		if po, ok := findPurchaseOrder(*data, item.PurchaseOrderID); ok {
			amount := round(item.NetWeight * po.UnitPrice)
			data.Payables = append(data.Payables, Payable{
				ID: nextID(data, "payable"), BillNo: number("AP", data.Next["payable"]),
				SupplierID: item.SupplierID, Amount: amount, DueDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"), Status: "open",
			})
		}
		addAudit(data, session.User.Username, "create", "scale_ticket", ticket.ID, ticket.TicketNo, clientIP(r))
		addAudit(data, session.User.Username, "create", "raw_material_receipt", item.ID, item.ReceiptNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "procurement.receipt.created")
}

func (a *App) finance(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 || parts[0] == "overview" {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, financePayload(data))
		return
	}
	if len(parts) == 3 && parts[0] == "invoices" && parts[2] == "submit-tax" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.submitTaxInvoice(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "invoices" && parts[2] == "red-offset" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.redOffsetInvoice(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "invoices" && parts[2] == "download" && r.Method == http.MethodGet {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.downloadInvoice(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "red-letters" && parts[2] == "approve" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.approveRedLetterInfo(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "supplier-statements" && parts[2] == "approve" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.approveSupplierStatement(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "payment-plans" && parts[2] == "settle" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.settlePaymentPlan(w, r, session, id)
		return
	}
	if len(parts) == 2 && parts[0] == "collections" && parts[1] == "generate" && r.Method == http.MethodPost {
		a.generateCollectionTasks(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "collections" && parts[2] == "handle" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.handleCollectionTask(w, r, session, id)
		return
	}
	if len(parts) == 3 && parts[0] == "collections" && parts[2] == "send" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.sendCollectionTask(w, r, session, id)
		return
	}
	if r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		switch parts[0] {
		case "invoices":
			writeJSON(w, http.StatusOK, data.SalesInvoices)
		case "red-letters":
			writeJSON(w, http.StatusOK, data.RedLetterInfos)
		case "receivables":
			writeJSON(w, http.StatusOK, data.Receivables)
		case "receipts":
			writeJSON(w, http.StatusOK, data.Receipts)
		case "payment-plans":
			writeJSON(w, http.StatusOK, data.PaymentPlans)
		case "collections":
			writeJSON(w, http.StatusOK, data.CollectionTasks)
		case "collection-templates":
			writeJSON(w, http.StatusOK, data.CollectionTemplates)
		case "collection-dispatches":
			writeJSON(w, http.StatusOK, data.CollectionDispatches)
		case "payables":
			writeJSON(w, http.StatusOK, data.Payables)
		case "payments":
			writeJSON(w, http.StatusOK, data.Payments)
		case "supplier-statements":
			writeJSON(w, http.StatusOK, data.SupplierStatements)
		case "profit":
			writeJSON(w, http.StatusOK, data.ProjectProfits)
		default:
			writeError(w, http.StatusNotFound, "unknown finance resource")
		}
		return
	}
	switch parts[0] {
	case "invoices":
		a.createInvoice(w, r, session)
	case "red-letters":
		a.createRedLetterInfo(w, r, session)
	case "receipts":
		a.createReceipt(w, r, session)
	case "payments":
		a.createPayment(w, r, session)
	case "payment-plans":
		a.createPaymentPlan(w, r, session)
	case "collection-templates":
		a.createCollectionTemplate(w, r, session)
	case "supplier-statements":
		a.createSupplierStatement(w, r, session)
	default:
		writeError(w, http.StatusNotFound, "unknown finance route")
	}
}

func financePayload(data AppData) map[string]interface{} {
	return map[string]interface{}{
		"invoices":                 data.SalesInvoices,
		"redLetterInfos":           data.RedLetterInfos,
		"invoiceTypes":             dataDictionariesByType(data, "invoice_type"),
		"taxGatewaySubmissions":    data.TaxGatewaySubmissions,
		"statements":               data.Statements,
		"receivables":              data.Receivables,
		"receipts":                 data.Receipts,
		"paymentPlans":             data.PaymentPlans,
		"collectionTasks":          data.CollectionTasks,
		"collectionTemplates":      data.CollectionTemplates,
		"collectionDispatches":     data.CollectionDispatches,
		"agingBuckets":             financeAgingBuckets(data),
		"supplierStatements":       data.SupplierStatements,
		"payables":                 data.Payables,
		"payments":                 data.Payments,
		"transportSettlements":     data.TransportSettlements,
		"transportSettlementItems": data.TransportSettlementItems,
		"costCalcs":                data.CostCalcs,
		"projectProfits":           data.ProjectProfits,
	}
}

func (a *App) createInvoice(w http.ResponseWriter, r *http.Request, session Session) {
	var req struct {
		StatementID     int64   `json:"statementId"`
		TaxRate         float64 `json:"taxRate"`
		InvoiceCategory string  `json:"invoiceCategory"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice")
		return
	}
	var invoice SalesInvoice
	err := a.store.Mutate(func(data *AppData) error {
		statement, ok := findStatement(*data, req.StatementID)
		if !ok {
			return fmt.Errorf("对账单不存在")
		}
		if statement.Status != "confirmed" {
			return fmt.Errorf("对账单未确认或已开票")
		}
		for _, existing := range data.SalesInvoices {
			if existing.StatementID == statement.ID && existing.InvoiceType != "red" && existing.Status != "cancelled" {
				return fmt.Errorf("对账单已存在蓝字发票")
			}
		}
		rate := req.TaxRate
		if rate == 0 {
			rate = 0.13
		}
		invoice = SalesInvoice{
			ID: nextID(data, "invoice"), InvoiceNo: number("INV", data.Next["invoice"]),
			StatementID: statement.ID, CustomerID: statement.CustomerID, Amount: statement.TotalAmount,
			TaxRate: rate, TaxAmount: round(statement.TotalAmount * rate), TaxStatus: "pending",
			Status: "issued", IssuedAt: nowString(), InvoiceType: "blue", InvoiceCategory: fallback(req.InvoiceCategory, "blue_vat_special"),
		}
		data.SalesInvoices = append(data.SalesInvoices, invoice)
		receivable := Receivable{
			ID: nextID(data, "receivable"), BillNo: number("AR", data.Next["receivable"]),
			CustomerID: statement.CustomerID, StatementID: statement.ID, InvoiceID: invoice.ID,
			Amount: invoice.Amount, DueDate: time.Now().AddDate(0, 0, customerPaymentTerm(*data, statement.CustomerID)).Format("2006-01-02"),
			Status: "open", CreatedAt: nowString(),
		}
		data.Receivables = append(data.Receivables, receivable)
		for i := range data.Statements {
			if data.Statements[i].ID == statement.ID {
				data.Statements[i].Status = "invoiced"
			}
		}
		addAudit(data, session.User.Username, "create", "sales_invoice", invoice.ID, invoice.InvoiceNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, invoice, "finance.invoice.created")
}

func (a *App) submitTaxInvoice(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var invoice SalesInvoice
	var submission TaxGatewaySubmission
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.SalesInvoices {
			if data.SalesInvoices[i].ID != id {
				continue
			}
			if !userCanAccessInvoice(*data, session.User, data.SalesInvoices[i]) {
				return fmt.Errorf("无权操作该发票")
			}
			if data.SalesInvoices[i].TaxStatus == "submitted" {
				invoice = data.SalesInvoices[i]
				return nil
			}
			data.SalesInvoices[i].TaxStatus = "submitting"
			invoice = data.SalesInvoices[i]
			submissionID := nextID(data, "taxSubmission")
			submission = TaxGatewaySubmission{
				ID:           submissionID,
				SubmissionNo: number("TGS", submissionID),
				InvoiceID:    invoice.ID,
				InvoiceNo:    invoice.InvoiceNo,
				Action:       "issue",
				Provider:     a.runtime.taxGateway.Provider,
				Endpoint:     sanitizedTaxEndpoint(a.runtime.taxGateway.URL),
				Status:       "submitting",
				SubmittedAt:  nowString(),
				Actor:        session.User.Username,
			}
			if submission.Provider == "" {
				submission.Provider = defaultTaxGatewayProvider(submission.Endpoint)
			}
			data.TaxGatewaySubmissions = append(data.TaxGatewaySubmissions, submission)
			addAudit(data, session.User.Username, "submit_tax_start", "sales_invoice", id, submission.SubmissionNo, clientIP(r))
			return nil
		}
		return fmt.Errorf("发票不存在")
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if invoice.TaxStatus == "submitted" {
		a.emit("finance.invoice.tax_submitted", invoice)
		writeJSON(w, http.StatusCreated, invoice)
		return
	}

	result := a.runtime.SubmitTaxInvoice(r.Context(), invoice)
	err = a.store.Mutate(func(data *AppData) error {
		updated, updatedSubmission, err := applyTaxGatewayResultLocked(data, id, submission.ID, "issue", result, session.User.Username, clientIP(r))
		invoice = updated
		submission = updatedSubmission
		return err
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if result.Status == "accepted" {
		a.emit("finance.invoice.tax_accepted", map[string]interface{}{"invoice": invoice, "submission": submission})
		writeJSON(w, http.StatusCreated, invoice)
		return
	}
	if result.Status != "submitted" {
		a.emit("finance.invoice.tax_failed", map[string]interface{}{"invoice": invoice, "submission": submission})
		writeError(w, http.StatusBadGateway, "税控网关提交失败: "+result.Error)
		return
	}
	a.emit("finance.invoice.tax_submitted", map[string]interface{}{"invoice": invoice, "submission": submission})
	writeJSON(w, http.StatusCreated, invoice)
}

func (a *App) redOffsetInvoice(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Reason          string `json:"reason"`
		RedLetterInfoID int64  `json:"redLetterInfoId"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid red offset request")
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "客户红冲申请"
	}
	var original SalesInvoice
	var redInfo RedLetterInfo
	var redInvoice SalesInvoice
	var submission TaxGatewaySubmission
	err := a.store.Mutate(func(data *AppData) error {
		for _, item := range data.SalesInvoices {
			if item.OriginalInvoiceID == id && item.InvoiceType == "red" && item.TaxStatus != "failed" {
				return fmt.Errorf("该发票已有红冲单")
			}
		}
		for i := range data.SalesInvoices {
			if data.SalesInvoices[i].ID != id {
				continue
			}
			if !userCanAccessInvoice(*data, session.User, data.SalesInvoices[i]) {
				return fmt.Errorf("无权操作该发票")
			}
			if data.SalesInvoices[i].InvoiceType == "red" {
				return fmt.Errorf("红票不能再次红冲")
			}
			if data.SalesInvoices[i].TaxStatus != "submitted" || data.SalesInvoices[i].TaxControlNo == "" {
				return fmt.Errorf("发票尚未完成税控提交，不能红冲")
			}
			original = data.SalesInvoices[i]
			break
		}
		if original.ID == 0 {
			return fmt.Errorf("发票不存在")
		}
		var infoErr error
		redInfo, infoErr = ensureRedLetterInfoForRedOffset(data, original, req.RedLetterInfoID, reason, fallback(session.User.DisplayName, session.User.Username))
		if infoErr != nil {
			return infoErr
		}
		if redInfo.Reason != "" && strings.TrimSpace(req.Reason) == "" {
			reason = redInfo.Reason
		}
		amount := -original.Amount
		if amount > 0 {
			amount = -amount
		}
		taxAmount := -original.TaxAmount
		if taxAmount > 0 {
			taxAmount = -taxAmount
		}
		redID := nextID(data, "invoice")
		redInvoice = SalesInvoice{
			ID:                redID,
			InvoiceNo:         number("RINV", redID),
			StatementID:       original.StatementID,
			CustomerID:        original.CustomerID,
			Amount:            amount,
			TaxRate:           original.TaxRate,
			TaxAmount:         taxAmount,
			TaxStatus:         "submitting",
			Status:            "red_pending",
			IssuedAt:          nowString(),
			InvoiceType:       "red",
			InvoiceCategory:   redInvoiceCategory(original.InvoiceCategory),
			OriginalInvoiceID: original.ID,
			RedLetterInfoID:   redInfo.ID,
			RedLetterInfoNo:   redInfo.InfoNo,
			RedReason:         reason,
		}
		data.SalesInvoices = append(data.SalesInvoices, redInvoice)
		receivable := Receivable{
			ID: nextID(data, "receivable"), BillNo: number("AR", data.Next["receivable"]),
			CustomerID: original.CustomerID, StatementID: original.StatementID, InvoiceID: redInvoice.ID,
			Amount: redInvoice.Amount, DueDate: todayString(), Status: "credited", CreatedAt: nowString(),
		}
		data.Receivables = append(data.Receivables, receivable)
		submissionID := nextID(data, "taxSubmission")
		submission = TaxGatewaySubmission{
			ID:           submissionID,
			SubmissionNo: number("TGS", submissionID),
			InvoiceID:    redInvoice.ID,
			InvoiceNo:    redInvoice.InvoiceNo,
			Action:       "red_offset",
			Provider:     a.runtime.taxGateway.Provider,
			Endpoint:     sanitizedTaxEndpoint(a.runtime.taxGateway.URL),
			Status:       "submitting",
			SubmittedAt:  nowString(),
			Actor:        session.User.Username,
		}
		if submission.Provider == "" {
			submission.Provider = defaultTaxGatewayProvider(submission.Endpoint)
		}
		data.TaxGatewaySubmissions = append(data.TaxGatewaySubmissions, submission)
		addAudit(data, session.User.Username, "red_offset_start", "sales_invoice", original.ID, redInvoice.InvoiceNo, clientIP(r))
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result := a.runtime.SubmitTaxRedInvoice(r.Context(), redInvoice, original)
	err = a.store.Mutate(func(data *AppData) error {
		updated, updatedSubmission, err := applyTaxGatewayResultLocked(data, redInvoice.ID, submission.ID, "red_offset", result, session.User.Username, clientIP(r))
		redInvoice = updated
		submission = updatedSubmission
		return err
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if result.Status == "accepted" {
		a.emit("finance.invoice.red_accepted", map[string]interface{}{"invoice": redInvoice, "submission": submission})
		writeJSON(w, http.StatusCreated, redInvoice)
		return
	}
	if result.Status != "submitted" {
		a.emit("finance.invoice.red_failed", map[string]interface{}{"invoice": redInvoice, "submission": submission})
		writeError(w, http.StatusBadGateway, "税控红冲失败: "+result.Error)
		return
	}
	a.emit("finance.invoice.red_submitted", map[string]interface{}{"invoice": redInvoice, "submission": submission})
	writeJSON(w, http.StatusCreated, redInvoice)
}

func (a *App) taxGatewayCallback(w http.ResponseWriter, r *http.Request) {
	if a.runtime == nil {
		writeError(w, http.StatusServiceUnavailable, "runtime services unavailable")
		return
	}
	payload, err := parseTaxGatewayCallback(r, a.runtime.taxGateway)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var invoice SalesInvoice
	var submission TaxGatewaySubmission
	err = a.store.Mutate(func(data *AppData) error {
		submissionIndex := -1
		for i := range data.TaxGatewaySubmissions {
			if payload.RequestID != "" && data.TaxGatewaySubmissions[i].RequestID == payload.RequestID {
				submissionIndex = i
				break
			}
		}
		if submissionIndex < 0 && payload.InvoiceNo != "" {
			for i := range data.TaxGatewaySubmissions {
				if data.TaxGatewaySubmissions[i].InvoiceNo == payload.InvoiceNo {
					submissionIndex = i
				}
			}
		}
		if submissionIndex < 0 {
			return fmt.Errorf("税控提交流水不存在")
		}
		current := data.TaxGatewaySubmissions[submissionIndex]
		status := payload.Status
		if status == "" {
			if payload.Error != "" {
				status = "failed"
			} else {
				status = "submitted"
			}
		}
		result := taxGatewayResult{
			Provider:     fallback(payload.Provider, current.Provider),
			Endpoint:     current.Endpoint,
			RequestID:    fallback(payload.RequestID, current.RequestID),
			Status:       status,
			TaxControlNo: payload.TaxControlNo,
			FileURL:      payload.FileURL,
			Error:        payload.Error,
			Attempt:      current.Attempt,
			DurationMs:   current.DurationMs,
		}
		action := fallback(payload.Action, current.Action)
		updated, updatedSubmission, err := applyTaxGatewayResultLocked(data, current.InvoiceID, current.ID, action, result, "tax-gateway", clientIP(r))
		invoice = updated
		submission = updatedSubmission
		return err
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit("finance.invoice.tax_callback", map[string]interface{}{"invoice": invoice, "submission": submission})
	writeJSON(w, http.StatusOK, map[string]interface{}{"invoice": invoice, "submission": submission})
}

func applyTaxGatewayResultLocked(data *AppData, invoiceID, submissionID int64, action string, result taxGatewayResult, actor, ip string) (SalesInvoice, TaxGatewaySubmission, error) {
	status := result.Status
	if status == "" {
		status = "failed"
	}
	if status != "submitted" && status != "accepted" && status != "failed" {
		if result.Error != "" {
			status = "failed"
		}
	}
	if status == "submitted" && result.TaxControlNo == "" {
		return SalesInvoice{}, TaxGatewaySubmission{}, fmt.Errorf("税控平台未返回税控流水号")
	}
	submissionIndex := -1
	for i := range data.TaxGatewaySubmissions {
		if data.TaxGatewaySubmissions[i].ID == submissionID {
			submissionIndex = i
			break
		}
	}
	if submissionIndex < 0 {
		return SalesInvoice{}, TaxGatewaySubmission{}, fmt.Errorf("税控提交流水不存在")
	}
	if action == "" {
		action = data.TaxGatewaySubmissions[submissionIndex].Action
	}
	if action == "" {
		action = "issue"
	}
	data.TaxGatewaySubmissions[submissionIndex].Action = action
	data.TaxGatewaySubmissions[submissionIndex].Provider = fallback(result.Provider, data.TaxGatewaySubmissions[submissionIndex].Provider)
	data.TaxGatewaySubmissions[submissionIndex].Endpoint = fallback(result.Endpoint, data.TaxGatewaySubmissions[submissionIndex].Endpoint)
	data.TaxGatewaySubmissions[submissionIndex].RequestID = fallback(result.RequestID, data.TaxGatewaySubmissions[submissionIndex].RequestID)
	data.TaxGatewaySubmissions[submissionIndex].Status = status
	data.TaxGatewaySubmissions[submissionIndex].TaxControlNo = result.TaxControlNo
	data.TaxGatewaySubmissions[submissionIndex].FileURL = result.FileURL
	data.TaxGatewaySubmissions[submissionIndex].Error = result.Error
	if result.Attempt > 0 {
		data.TaxGatewaySubmissions[submissionIndex].Attempt = result.Attempt
	}
	if result.DurationMs > 0 {
		data.TaxGatewaySubmissions[submissionIndex].DurationMs = result.DurationMs
	}
	if status != "accepted" {
		data.TaxGatewaySubmissions[submissionIndex].CompletedAt = nowString()
	}
	submission := data.TaxGatewaySubmissions[submissionIndex]

	for i := range data.SalesInvoices {
		if data.SalesInvoices[i].ID != invoiceID {
			continue
		}
		if status == "submitted" {
			data.SalesInvoices[i].TaxControlNo = fallback(result.TaxControlNo, data.SalesInvoices[i].TaxControlNo)
			data.SalesInvoices[i].TaxStatus = "submitted"
			data.SalesInvoices[i].FileURL = fallback(result.FileURL, data.SalesInvoices[i].FileURL)
			if action == "red_offset" || data.SalesInvoices[i].InvoiceType == "red" {
				data.SalesInvoices[i].Status = "red_issued"
				data.SalesInvoices[i].RedAt = nowString()
				markRedLetterInfoUsed(data, data.SalesInvoices[i], data.SalesInvoices[i].TaxControlNo)
				markOriginalInvoiceRedOffset(data, data.SalesInvoices[i])
				addAudit(data, actor, "red_offset", "sales_invoice", data.SalesInvoices[i].ID, data.SalesInvoices[i].TaxControlNo, ip)
			} else {
				data.SalesInvoices[i].Status = fallback(data.SalesInvoices[i].Status, "issued")
				addAudit(data, actor, "submit_tax", "sales_invoice", data.SalesInvoices[i].ID, data.SalesInvoices[i].TaxControlNo, ip)
			}
			updateTaxIntegrationEndpoint(data, result, "online")
		} else if status == "accepted" {
			data.SalesInvoices[i].TaxStatus = "accepted"
			addAudit(data, actor, "tax_accepted", "sales_invoice", data.SalesInvoices[i].ID, submission.SubmissionNo, ip)
			updateTaxIntegrationEndpoint(data, result, "online")
		} else {
			data.SalesInvoices[i].TaxStatus = "failed"
			if result.Error == "" {
				result.Error = "税控平台未返回成功状态"
				data.TaxGatewaySubmissions[submissionIndex].Error = result.Error
				submission = data.TaxGatewaySubmissions[submissionIndex]
			}
			addAudit(data, actor, "submit_tax_failed", "sales_invoice", data.SalesInvoices[i].ID, result.Error, ip)
			updateTaxIntegrationEndpoint(data, result, "degraded")
		}
		return data.SalesInvoices[i], submission, nil
	}
	return SalesInvoice{}, submission, fmt.Errorf("发票不存在")
}

func markOriginalInvoiceRedOffset(data *AppData, redInvoice SalesInvoice) {
	if redInvoice.OriginalInvoiceID == 0 {
		return
	}
	for i := range data.SalesInvoices {
		if data.SalesInvoices[i].ID == redInvoice.OriginalInvoiceID {
			data.SalesInvoices[i].Status = "red_offset"
			data.SalesInvoices[i].TaxStatus = "red_offset"
			data.SalesInvoices[i].RedAt = redInvoice.RedAt
			data.SalesInvoices[i].RedReason = redInvoice.RedReason
			data.SalesInvoices[i].RedLetterInfoID = redInvoice.RedLetterInfoID
			data.SalesInvoices[i].RedLetterInfoNo = redInvoice.RedLetterInfoNo
			return
		}
	}
}

func (a *App) downloadInvoice(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var invoice SalesInvoice
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.SalesInvoices {
			if data.SalesInvoices[i].ID != id {
				continue
			}
			if !userCanAccessInvoice(*data, session.User, data.SalesInvoices[i]) {
				return fmt.Errorf("无权下载该发票")
			}
			if data.SalesInvoices[i].TaxStatus != "submitted" || data.SalesInvoices[i].FileURL == "" {
				return fmt.Errorf("发票尚未完成税控提交")
			}
			data.SalesInvoices[i].DownloadedAt = nowString()
			invoice = data.SalesInvoices[i]
			addAudit(data, session.User.Username, "download", "sales_invoice", id, invoice.InvoiceNo, clientIP(r))
			return nil
		}
		return fmt.Errorf("发票不存在")
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"invoice":      invoice,
		"fileName":     invoice.InvoiceNo + ".pdf",
		"url":          invoice.FileURL,
		"downloadedAt": invoice.DownloadedAt,
	})
}

func (a *App) createReceipt(w http.ResponseWriter, r *http.Request, session Session) {
	var item Receipt
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid receipt")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		created, err := appendReceiptLocked(data, item)
		if err != nil {
			return err
		}
		item = created
		addAudit(data, session.User.Username, "create", "receipt", item.ID, item.ReceiptNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "finance.receipt.created")
}

func (a *App) createPayment(w http.ResponseWriter, r *http.Request, session Session) {
	var item Payment
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		payable, ok := findPayable(*data, item.PayableID)
		if !ok {
			return fmt.Errorf("应付账款不存在")
		}
		item.ID = nextID(data, "payment")
		item.PaymentNo = number("PY", item.ID)
		item.SupplierID = nonZeroInt(item.SupplierID, payable.SupplierID)
		item.Method = fallback(item.Method, "bank")
		item.PaidAt = nowString()
		data.Payments = append(data.Payments, item)
		for i := range data.Payables {
			if data.Payables[i].ID == item.PayableID {
				data.Payables[i].PaidAmount = round(data.Payables[i].PaidAmount + item.Amount)
				if data.Payables[i].PaidAmount >= data.Payables[i].Amount {
					data.Payables[i].Status = "paid"
				} else {
					data.Payables[i].Status = "partial"
				}
			}
		}
		addAudit(data, session.User.Username, "create", "payment", item.ID, item.PaymentNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "finance.payment.created")
}

func (a *App) rules(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 && r.Method == http.MethodGet {
		data := scopedData(a.mustSnapshot(), session.User)
		writeJSON(w, http.StatusOK, map[string]interface{}{"rules": data.RuleDefinitions, "alarms": data.VehicleAlarms, "notifications": data.Notifications})
		return
	}
	if len(parts) == 1 && parts[0] == "evaluate" && r.Method == http.MethodPost {
		a.evaluateRules(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "alarms" && parts[2] == "handle" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		a.handleAlarm(w, r, session, id)
		return
	}
	if len(parts) == 1 && parts[0] == "definitions" {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, a.mustSnapshot().RuleDefinitions)
			return
		}
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var item RuleDefinition
		if err := readJSON(r, &item); err != nil {
			writeError(w, http.StatusBadRequest, "invalid rule")
			return
		}
		var saved RuleDefinition
		err := a.store.Mutate(func(data *AppData) error {
			normalized, err := normalizeRuleDefinition(item)
			if err != nil {
				return err
			}
			if item.ID > 0 {
				for i := range data.RuleDefinitions {
					if data.RuleDefinitions[i].ID != item.ID {
						continue
					}
					for j := range data.RuleDefinitions {
						if j != i && strings.EqualFold(data.RuleDefinitions[j].Code, normalized.Code) {
							return fmt.Errorf("规则编码已存在")
						}
					}
					data.RuleDefinitions[i] = normalized
					saved = normalized
					addAudit(data, session.User.Username, "update", "rule_definition", normalized.ID, normalized.Code, clientIP(r))
					return nil
				}
				return fmt.Errorf("规则不存在")
			}
			for _, existing := range data.RuleDefinitions {
				if strings.EqualFold(existing.Code, normalized.Code) {
					return fmt.Errorf("规则编码已存在")
				}
			}
			normalized.ID = nextID(data, "rule")
			data.RuleDefinitions = append(data.RuleDefinitions, normalized)
			saved = normalized
			addAudit(data, session.User.Username, "create", "rule_definition", normalized.ID, normalized.Code, clientIP(r))
			return nil
		})
		a.respondMutation(w, err, saved, "rule.definition.saved")
		return
	}
	if len(parts) == 3 && parts[0] == "definitions" && parts[2] == "status" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req struct {
			Enabled bool `json:"enabled"`
		}
		_ = readJSON(r, &req)
		var saved RuleDefinition
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.RuleDefinitions {
				if data.RuleDefinitions[i].ID != id {
					continue
				}
				data.RuleDefinitions[i].Enabled = req.Enabled
				saved = data.RuleDefinitions[i]
				addAudit(data, session.User.Username, "status", "rule_definition", id, saved.Code, clientIP(r))
				return nil
			}
			return fmt.Errorf("规则不存在")
		})
		a.respondMutation(w, err, saved, "rule.definition.status")
		return
	}
	if len(parts) == 2 && parts[0] == "definitions" && r.Method == http.MethodDelete {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var deleted RuleDefinition
		err := a.store.Mutate(func(data *AppData) error {
			for i, item := range data.RuleDefinitions {
				if item.ID != id {
					continue
				}
				if item.Enabled {
					return fmt.Errorf("启用中的规则不能删除，请先停用")
				}
				deleted = item
				data.RuleDefinitions = append(data.RuleDefinitions[:i], data.RuleDefinitions[i+1:]...)
				addAudit(data, session.User.Username, "delete", "rule_definition", id, item.Code, clientIP(r))
				return nil
			}
			return fmt.Errorf("规则不存在")
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.emit("rule.definition.deleted", deleted)
		writeJSON(w, http.StatusOK, deleted)
		return
	}
	writeError(w, http.StatusNotFound, "unknown rule route")
}

func normalizeRuleDefinition(item RuleDefinition) (RuleDefinition, error) {
	code := strings.TrimSpace(item.Code)
	name := strings.TrimSpace(item.Name)
	metric := strings.TrimSpace(item.Metric)
	if code == "" || name == "" || metric == "" {
		return RuleDefinition{}, fmt.Errorf("规则编码、名称和指标不能为空")
	}
	return RuleDefinition{
		ID:          item.ID,
		Code:        code,
		Name:        name,
		Category:    fallback(strings.TrimSpace(item.Category), "vehicle"),
		Metric:      metric,
		Operator:    fallback(strings.TrimSpace(item.Operator), ">"),
		Threshold:   item.Threshold,
		Level:       fallback(strings.TrimSpace(item.Level), "warning"),
		Enabled:     item.Enabled,
		NotifyRoles: normalizeStringList(item.NotifyRoles),
		Description: strings.TrimSpace(item.Description),
	}, nil
}

func (a *App) evaluateRules(w http.ResponseWriter, r *http.Request, session Session) {
	var created []VehicleAlarm
	err := a.store.Mutate(func(data *AppData) error {
		for _, latest := range data.LatestLocations {
			for _, rule := range data.RuleDefinitions {
				if !rule.Enabled {
					continue
				}
				if rule.Metric == "speed" && latest.Speed > rule.Threshold && !openAlarmExists(*data, latest.VehicleID, rule.Code) {
					alarm := appendAlarm(data, latest.VehicleID, 0, rule.Code, rule.Level, fmt.Sprintf("%s 触发规则：%s", latest.PlateNo, rule.Name))
					created = append(created, alarm)
					appendRuleNotifications(data, rule, alarm.Message)
				}
				if rule.Metric == "offline_minutes" && latest.OnlineStatus == "offline" && !openAlarmExists(*data, latest.VehicleID, rule.Code) {
					alarm := appendAlarm(data, latest.VehicleID, 0, rule.Code, rule.Level, fmt.Sprintf("%s 触发规则：%s", latest.PlateNo, rule.Name))
					created = append(created, alarm)
					appendRuleNotifications(data, rule, alarm.Message)
				}
			}
		}
		addAudit(data, session.User.Username, "evaluate", "rule_engine", 0, fmt.Sprintf("%d alarms", len(created)), clientIP(r))
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.hub.Broadcast("vehicle.alarm.created", created)
	writeJSON(w, http.StatusOK, map[string]interface{}{"created": created})
}

func (a *App) handleAlarm(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Remark string `json:"remark"`
	}
	_ = readJSON(r, &req)
	var updated VehicleAlarm
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.VehicleAlarms {
			if data.VehicleAlarms[i].ID == id {
				data.VehicleAlarms[i].Status = "handled"
				data.VehicleAlarms[i].HandledBy = session.User.DisplayName
				data.VehicleAlarms[i].HandledAt = nowString()
				if req.Remark != "" {
					data.VehicleAlarms[i].Message = data.VehicleAlarms[i].Message + "；处理：" + req.Remark
				}
				updated = data.VehicleAlarms[i]
				addAudit(data, session.User.Username, "handle", "vehicle_alarm", id, req.Remark, clientIP(r))
				return nil
			}
		}
		return fmt.Errorf("预警不存在")
	})
	a.respondMutation(w, err, updated, "vehicle.alarm.handled")
}

func (a *App) integrations(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	data := scopedData(a.mustSnapshot(), session.User)
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "overview") {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"endpoints":      data.IntegrationEndpoints,
			"vehicleDevices": data.VehicleDevices,
			"scaleDevices":   data.ScaleDevices,
			"plants":         plantsWithGatewayStatus(data),
			"protocolFrames": data.DeviceProtocolFrames,
		})
		return
	}
	switch parts[0] {
	case "endpoints":
		if len(parts) == 1 && r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, data.IntegrationEndpoints)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodPost {
			var req IntegrationEndpoint
			if err := readJSON(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid integration endpoint payload")
				return
			}
			var saved IntegrationEndpoint
			err := a.store.Mutate(func(data *AppData) error {
				endpoint, err := normalizeIntegrationEndpoint(req)
				if err != nil {
					return err
				}
				if req.ID > 0 {
					for i := range data.IntegrationEndpoints {
						if data.IntegrationEndpoints[i].ID != req.ID {
							continue
						}
						for j := range data.IntegrationEndpoints {
							if j != i && sameIntegrationEndpointKey(data.IntegrationEndpoints[j], endpoint) {
								return fmt.Errorf("集成端点已存在")
							}
						}
						endpoint.LastSyncAt = data.IntegrationEndpoints[i].LastSyncAt
						data.IntegrationEndpoints[i] = endpoint
						saved = endpoint
						addAudit(data, session.User.Username, "update", "integration_endpoint", endpoint.ID, endpoint.Name+"/"+endpoint.Type, clientIP(r))
						return nil
					}
					return fmt.Errorf("集成端点不存在")
				}
				for _, existing := range data.IntegrationEndpoints {
					if sameIntegrationEndpointKey(existing, endpoint) {
						return fmt.Errorf("集成端点已存在")
					}
					ensureCounterAtLeast(data, "integration", existing.ID)
				}
				endpoint.ID = nextID(data, "integration")
				data.IntegrationEndpoints = append(data.IntegrationEndpoints, endpoint)
				saved = endpoint
				addAudit(data, session.User.Username, "create", "integration_endpoint", endpoint.ID, endpoint.Name+"/"+endpoint.Type, clientIP(r))
				return nil
			})
			a.respondMutation(w, err, saved, "integration.endpoint.saved")
			return
		}
		if len(parts) == 3 && parts[2] == "status" && r.Method == http.MethodPost {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			var req struct {
				Status string `json:"status"`
			}
			_ = readJSON(r, &req)
			var saved IntegrationEndpoint
			err := a.store.Mutate(func(data *AppData) error {
				status := fallback(strings.TrimSpace(req.Status), "online")
				for i := range data.IntegrationEndpoints {
					if data.IntegrationEndpoints[i].ID != id {
						continue
					}
					data.IntegrationEndpoints[i].Status = status
					saved = data.IntegrationEndpoints[i]
					addAudit(data, session.User.Username, "status", "integration_endpoint", id, saved.Name+"/"+status, clientIP(r))
					return nil
				}
				return fmt.Errorf("集成端点不存在")
			})
			a.respondMutation(w, err, saved, "integration.endpoint.status")
			return
		}
		if len(parts) == 2 && r.Method == http.MethodDelete {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			var deleted IntegrationEndpoint
			err := a.store.Mutate(func(data *AppData) error {
				for i, item := range data.IntegrationEndpoints {
					if item.ID != id {
						continue
					}
					if item.Status != "disabled" {
						return fmt.Errorf("启用中的集成端点不能删除，请先停用")
					}
					deleted = item
					data.IntegrationEndpoints = append(data.IntegrationEndpoints[:i], data.IntegrationEndpoints[i+1:]...)
					addAudit(data, session.User.Username, "delete", "integration_endpoint", id, item.Name+"/"+item.Type, clientIP(r))
					return nil
				}
				return fmt.Errorf("集成端点不存在")
			})
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			a.emit("integration.endpoint.deleted", deleted)
			writeJSON(w, http.StatusOK, deleted)
			return
		}
		writeError(w, http.StatusNotFound, "unknown integration endpoint route")
	case "vehicle-devices":
		writeJSON(w, http.StatusOK, data.VehicleDevices)
	case "scale-devices":
		writeJSON(w, http.StatusOK, data.ScaleDevices)
	case "plants":
		writeJSON(w, http.StatusOK, plantsWithGatewayStatus(data))
	case "protocol-frames":
		writeJSON(w, http.StatusOK, data.DeviceProtocolFrames)
	default:
		writeError(w, http.StatusNotFound, "unknown integration resource")
	}
}

func normalizeIntegrationEndpoint(req IntegrationEndpoint) (IntegrationEndpoint, error) {
	name := strings.TrimSpace(req.Name)
	endpointType := strings.TrimSpace(req.Type)
	if name == "" || endpointType == "" {
		return IntegrationEndpoint{}, fmt.Errorf("集成端点名称和类型不能为空")
	}
	url := strings.TrimSpace(req.URL)
	if url != "" {
		if err := validateNoMockEndpoint(url, "集成端点 URL"); err != nil {
			return IntegrationEndpoint{}, err
		}
	}
	status := fallback(strings.TrimSpace(req.Status), "online")
	if url == "" && status != "disabled" {
		return IntegrationEndpoint{}, fmt.Errorf("启用集成端点必须配置 URL")
	}
	return IntegrationEndpoint{
		ID:       req.ID,
		Name:     name,
		Type:     endpointType,
		Protocol: fallback(strings.TrimSpace(req.Protocol), "rest/http"),
		URL:      url,
		Status:   status,
	}, nil
}

func sameIntegrationEndpointKey(left IntegrationEndpoint, right IntegrationEndpoint) bool {
	return strings.EqualFold(strings.TrimSpace(left.Type), strings.TrimSpace(right.Type)) &&
		strings.EqualFold(strings.TrimSpace(left.Name), strings.TrimSpace(right.Name))
}

func findSupplier(data AppData, id int64) (Supplier, bool) {
	for _, item := range data.Suppliers {
		if item.ID == id {
			return item, true
		}
	}
	return Supplier{}, false
}

func findPurchaseOrder(data AppData, id int64) (PurchaseOrder, bool) {
	for _, item := range data.PurchaseOrders {
		if item.ID == id {
			return item, true
		}
	}
	return PurchaseOrder{}, false
}

func findStatement(data AppData, id int64) (Statement, bool) {
	for _, item := range data.Statements {
		if item.ID == id {
			return item, true
		}
	}
	return Statement{}, false
}

func findReceivable(data AppData, id int64) (Receivable, bool) {
	for _, item := range data.Receivables {
		if item.ID == id {
			return item, true
		}
	}
	return Receivable{}, false
}

func findPayable(data AppData, id int64) (Payable, bool) {
	for _, item := range data.Payables {
		if item.ID == id {
			return item, true
		}
	}
	return Payable{}, false
}

func increaseInventory(data *AppData, siteID, materialID, supplierID int64, quantity float64) float64 {
	return increaseInventoryWithLot(data, siteID, materialID, supplierID, quantity, "", 0)
}

func increaseInventoryWithLot(data *AppData, siteID, materialID, supplierID int64, quantity float64, batchNo string, rawReceiptID int64) float64 {
	warehouse, silo := inventoryStorageLocation(*data, siteID, materialID, quantity)
	for i := range data.Inventory {
		if data.Inventory[i].SiteID == siteID && data.Inventory[i].MaterialID == materialID && data.Inventory[i].Warehouse == warehouse && data.Inventory[i].Silo == silo && sameInventoryLot(data.Inventory[i], batchNo, rawReceiptID, supplierID) {
			data.Inventory[i].Quantity = round(data.Inventory[i].Quantity + quantity)
			data.Inventory[i].UpdatedAt = nowString()
			if material, ok := findMaterial(*data, materialID); ok && data.Inventory[i].Quantity >= material.SafeStock {
				data.Inventory[i].AvailableStatus = "available"
			}
			adjustSiloCurrentQty(data, siteID, silo, quantity)
			return inventoryBalance(*data, siteID, materialID)
		}
	}
	item := InventoryItem{
		ID: nextID(data, "inventory"), SiteID: siteID, Warehouse: warehouse, Silo: silo,
		MaterialID: materialID, BatchNo: batchNo, RawReceiptID: rawReceiptID, SupplierID: supplierID, Quantity: round(quantity), Unit: "t",
		QualityStatus: "pending", AvailableStatus: "available", UpdatedAt: nowString(),
	}
	data.Inventory = append(data.Inventory, item)
	adjustSiloCurrentQty(data, siteID, silo, quantity)
	return inventoryBalance(*data, siteID, materialID)
}

func sameInventoryLot(item InventoryItem, batchNo string, rawReceiptID, supplierID int64) bool {
	if batchNo == "" && rawReceiptID == 0 {
		return true
	}
	if item.BatchNo != batchNo || item.RawReceiptID != rawReceiptID {
		return false
	}
	return supplierID == 0 || item.SupplierID == 0 || item.SupplierID == supplierID
}

func customerPaymentTerm(data AppData, customerID int64) int {
	for _, customer := range data.Customers {
		if customer.ID == customerID && customer.PaymentTerm > 0 {
			return customer.PaymentTerm
		}
	}
	return 30
}

func appendAlarm(data *AppData, vehicleID, dispatchID int64, alarmType, level, message string) VehicleAlarm {
	alarm := VehicleAlarm{
		ID: nextID(data, "alarm"), VehicleID: vehicleID, DispatchID: dispatchID,
		AlarmType: alarmType, Level: level, Message: message, Status: "open", CreatedAt: nowString(),
	}
	data.VehicleAlarms = append(data.VehicleAlarms, alarm)
	return alarm
}

func openAlarmExists(data AppData, vehicleID int64, alarmType string) bool {
	for _, alarm := range data.VehicleAlarms {
		if alarm.VehicleID == vehicleID && alarm.AlarmType == alarmType && alarm.Status == "open" {
			return true
		}
	}
	return false
}

func appendRuleNotifications(data *AppData, rule RuleDefinition, content string) {
	for _, role := range rule.NotifyRoles {
		data.Notifications = append(data.Notifications, Notification{
			ID: nextID(data, "notification"), TargetRole: role, Channel: "system",
			Title: rule.Name, Content: content, Status: "unread", CreatedAt: nowString(),
		})
	}
}

func checksumVerified(checksum, signature string) bool {
	checksum = strings.TrimSpace(checksum)
	signature = strings.TrimSpace(signature)
	return strings.HasPrefix(checksum, "sha256:") && strings.HasPrefix(signature, "sig:")
}
