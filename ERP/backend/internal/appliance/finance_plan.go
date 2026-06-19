package appliance

import (
	"fmt"
	"net/http"
	"time"
)

func (a *App) createPaymentPlan(w http.ResponseWriter, r *http.Request, session Session) {
	var item PaymentPlan
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payment plan")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		receivable, ok := findReceivable(*data, item.ReceivableID)
		if !ok {
			return fmt.Errorf("应收账款不存在")
		}
		remaining := remainingReceivable(receivable)
		if remaining <= 0 {
			return fmt.Errorf("应收账款已结清")
		}
		if item.Amount == 0 {
			item.Amount = remaining
		}
		item.Amount = round(item.Amount)
		if item.Amount <= 0 {
			return fmt.Errorf("计划金额必须大于 0")
		}
		if item.Amount > remaining {
			return fmt.Errorf("计划金额超过未收金额")
		}
		item.ID = nextID(data, "paymentPlan")
		item.PlanNo = number("PP", item.ID)
		item.CustomerID = nonZeroInt(item.CustomerID, receivable.CustomerID)
		item.DueDate = fallback(item.DueDate, receivable.DueDate)
		item.Method = fallback(item.Method, "bank")
		item.Status = "planned"
		item.CreatedAt = nowString()
		data.PaymentPlans = append(data.PaymentPlans, item)
		addAudit(data, session.User.Username, "create", "payment_plan", item.ID, item.PlanNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "finance.payment_plan.created")
}

func (a *App) settlePaymentPlan(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Amount float64 `json:"amount"`
		Method string  `json:"method"`
		Remark string  `json:"remark"`
	}
	_ = readJSON(r, &req)
	var item PaymentPlan
	var receipt Receipt
	err := a.store.Mutate(func(data *AppData) error {
		index := paymentPlanIndex(*data, id)
		if index < 0 {
			return fmt.Errorf("付款计划不存在")
		}
		current := data.PaymentPlans[index]
		if current.Status == "settled" {
			item = current
			return nil
		}
		amount := req.Amount
		if amount == 0 {
			amount = current.Amount
		}
		amount = round(amount)
		if amount <= 0 {
			return fmt.Errorf("结清金额必须大于 0")
		}
		receivable, ok := findReceivable(*data, current.ReceivableID)
		if !ok {
			return fmt.Errorf("应收账款不存在")
		}
		remaining := remainingReceivable(receivable)
		if amount > remaining {
			amount = remaining
		}
		if amount <= 0 {
			return fmt.Errorf("应收账款已结清")
		}
		var receiptErr error
		receipt, receiptErr = appendReceiptLocked(data, Receipt{
			ReceivableID: current.ReceivableID,
			CustomerID:   current.CustomerID,
			Amount:       amount,
			Method:       fallback(req.Method, current.Method),
		})
		if receiptErr != nil {
			return receiptErr
		}
		current.Status = "settled"
		current.SettledAt = nowString()
		if req.Remark != "" {
			current.Remark = req.Remark
		}
		data.PaymentPlans[index] = current
		item = current
		addAudit(data, session.User.Username, "settle", "payment_plan", item.ID, item.PlanNo, clientIP(r))
		addAudit(data, session.User.Username, "create", "receipt", receipt.ID, receipt.ReceiptNo, clientIP(r))
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit("finance.payment_plan.settled", map[string]interface{}{"paymentPlan": item, "receipt": receipt})
	writeJSON(w, http.StatusCreated, map[string]interface{}{"paymentPlan": item, "receipt": receipt})
}

func appendReceiptLocked(data *AppData, item Receipt) (Receipt, error) {
	receivable, ok := findReceivable(*data, item.ReceivableID)
	if !ok {
		return Receipt{}, fmt.Errorf("应收账款不存在")
	}
	item.Amount = round(item.Amount)
	if item.Amount <= 0 {
		return Receipt{}, fmt.Errorf("收款金额必须大于 0")
	}
	item.ID = nextID(data, "moneyReceipt")
	item.ReceiptNo = number("RC", item.ID)
	item.CustomerID = nonZeroInt(item.CustomerID, receivable.CustomerID)
	item.Method = fallback(item.Method, "bank")
	item.Status = "confirmed"
	item.ReceivedAt = nowString()
	data.Receipts = append(data.Receipts, item)
	for i := range data.Receivables {
		if data.Receivables[i].ID == item.ReceivableID {
			data.Receivables[i].ReceivedAmount = round(data.Receivables[i].ReceivedAmount + item.Amount)
			if data.Receivables[i].ReceivedAmount >= data.Receivables[i].Amount {
				data.Receivables[i].Status = "paid"
			} else {
				data.Receivables[i].Status = "partial"
			}
		}
	}
	for i := range data.Customers {
		if data.Customers[i].ID == item.CustomerID {
			data.Customers[i].Receivable = round(data.Customers[i].Receivable - item.Amount)
			if data.Customers[i].Receivable < 0 {
				data.Customers[i].Receivable = 0
			}
		}
	}
	return item, nil
}

func financeAgingBuckets(data AppData) []ReceivableAgingBucket {
	buckets := []ReceivableAgingBucket{
		{Bucket: "current", Label: "未逾期"},
		{Bucket: "1_30", Label: "逾期 1-30 天"},
		{Bucket: "31_60", Label: "逾期 31-60 天"},
		{Bucket: "61_90", Label: "逾期 61-90 天"},
		{Bucket: "over_90", Label: "逾期 90 天以上"},
	}
	index := map[string]int{}
	for i, bucket := range buckets {
		index[bucket.Bucket] = i
	}
	today, _ := time.Parse("2006-01-02", todayString())
	for _, receivable := range data.Receivables {
		remaining := remainingReceivable(receivable)
		if remaining <= 0 || receivable.Status == "paid" {
			continue
		}
		bucket := "current"
		if due, err := time.Parse("2006-01-02", receivable.DueDate); err == nil {
			days := int(today.Sub(due).Hours() / 24)
			switch {
			case days <= 0:
				bucket = "current"
			case days <= 30:
				bucket = "1_30"
			case days <= 60:
				bucket = "31_60"
			case days <= 90:
				bucket = "61_90"
			default:
				bucket = "over_90"
			}
		}
		i := index[bucket]
		buckets[i].Count++
		buckets[i].Amount = round(buckets[i].Amount + remaining)
		if bucket != "current" {
			buckets[i].OverdueAmount = round(buckets[i].OverdueAmount + remaining)
		}
	}
	return buckets
}

func remainingReceivable(item Receivable) float64 {
	return round(item.Amount - item.ReceivedAmount)
}

func paymentPlanIndex(data AppData, id int64) int {
	for i := range data.PaymentPlans {
		if data.PaymentPlans[i].ID == id {
			return i
		}
	}
	return -1
}
