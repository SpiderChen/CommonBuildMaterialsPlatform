package appliance

import (
	"sort"
	"time"
)

type ManagementReports struct {
	ProjectProfit      []ProjectProfitReport    `json:"projectProfit"`
	InventoryWarnings  []InventoryItem          `json:"inventoryWarnings"`
	VehicleEfficiency  []map[string]interface{} `json:"vehicleEfficiency"`
	CustomerStatements []Statement              `json:"customerStatements"`
	Operating          OperatingAnalysisReport  `json:"operating"`
	CustomerAging      []CustomerAgingReport    `json:"customerAging"`
	AgingBuckets       []ReceivableAgingBucket  `json:"agingBuckets"`
	Quality            QualityAnalysisReport    `json:"quality"`
	Energy             []ProductionEnergyReport `json:"energy"`
}

type ProjectProfitReport struct {
	ProjectID int64   `json:"projectId"`
	Revenue   float64 `json:"revenue"`
	Cost      float64 `json:"cost"`
	Profit    float64 `json:"profit"`
	Margin    float64 `json:"margin"`
	Period    string  `json:"period"`
}

type OperatingAnalysisReport struct {
	OrderCount            int     `json:"orderCount"`
	PlannedQty            float64 `json:"plannedQty"`
	SignedQty             float64 `json:"signedQty"`
	Revenue               float64 `json:"revenue"`
	MaterialCost          float64 `json:"materialCost"`
	TransportCost         float64 `json:"transportCost"`
	TotalCost             float64 `json:"totalCost"`
	GrossProfit           float64 `json:"grossProfit"`
	GrossMargin           float64 `json:"grossMargin"`
	ReceivableBalance     float64 `json:"receivableBalance"`
	OverdueReceivable     float64 `json:"overdueReceivable"`
	InventoryWarningCount int     `json:"inventoryWarningCount"`
	OpenQualityIssues     int     `json:"openQualityIssues"`
}

type CustomerAgingReport struct {
	CustomerID    int64   `json:"customerId"`
	CustomerName  string  `json:"customerName"`
	Current       float64 `json:"current"`
	Overdue1To30  float64 `json:"overdue1To30"`
	Overdue31To60 float64 `json:"overdue31To60"`
	Overdue61To90 float64 `json:"overdue61To90"`
	OverdueOver90 float64 `json:"overdueOver90"`
	Total         float64 `json:"total"`
	OverdueTotal  float64 `json:"overdueTotal"`
}

type QualityAnalysisReport struct {
	Inspections        int                 `json:"inspections"`
	Passed             int                 `json:"passed"`
	Pending            int                 `json:"pending"`
	Failed             int                 `json:"failed"`
	PassRate           float64             `json:"passRate"`
	Samples            int                 `json:"samples"`
	SamplePassed       int                 `json:"samplePassed"`
	SamplePending      int                 `json:"samplePending"`
	SampleFailed       int                 `json:"sampleFailed"`
	LaboratoryTests    int                 `json:"laboratoryTests"`
	OpenExceptionCount int                 `json:"openExceptionCount"`
	OpenExceptions     []QualityInspection `json:"openExceptions"`
	QualityExceptions  []QualityException  `json:"qualityExceptions"`
}

type ProductionEnergyReport struct {
	SiteID            int64   `json:"siteId"`
	ProducedQty       float64 `json:"producedQty"`
	BatchCount        int     `json:"batchCount"`
	MaterialUsageQty  float64 `json:"materialUsageQty"`
	MaterialCost      float64 `json:"materialCost"`
	EstimatedPowerKwh float64 `json:"estimatedPowerKwh"`
	UnitMaterialCost  float64 `json:"unitMaterialCost"`
	UnitPowerKwh      float64 `json:"unitPowerKwh"`
}

func buildManagementReports(data AppData) ManagementReports {
	reports := ManagementReports{
		InventoryWarnings:  inventoryWarnings(data),
		VehicleEfficiency:  vehicleEfficiency(data),
		CustomerStatements: data.Statements,
		AgingBuckets:       financeAgingBuckets(data),
		CustomerAging:      buildCustomerAgingReport(data),
		Quality:            buildQualityAnalysisReport(data),
		Energy:             buildProductionEnergyReports(data),
	}
	reports.ProjectProfit, reports.Operating = buildProjectProfitReports(data, reports.InventoryWarnings, reports.Quality)
	return reports
}

func buildProjectProfitReports(data AppData, inventoryWarnings []InventoryItem, quality QualityAnalysisReport) ([]ProjectProfitReport, OperatingAnalysisReport) {
	profits := map[int64]*ProjectProfitReport{}
	operating := OperatingAnalysisReport{OrderCount: len(data.Orders), InventoryWarningCount: len(inventoryWarnings), OpenQualityIssues: len(quality.OpenExceptions) + quality.OpenExceptionCount}
	for _, order := range data.Orders {
		operating.PlannedQty = round(operating.PlannedQty + order.PlanQuantity)
	}
	for _, receivable := range data.Receivables {
		remaining := remainingReceivable(receivable)
		if remaining <= 0 || receivable.Status == "paid" {
			continue
		}
		operating.ReceivableBalance = round(operating.ReceivableBalance + remaining)
		if receivableOverdueDays(receivable) > 0 {
			operating.OverdueReceivable = round(operating.OverdueReceivable + remaining)
		}
	}
	for _, settlement := range data.TransportSettlements {
		operating.TransportCost = round(operating.TransportCost + settlement.Amount)
	}
	for _, sign := range data.DeliverySigns {
		order, ok := findOrder(data, sign.OrderID)
		if !ok {
			continue
		}
		unitPrice, costPrice := signedUnitAndCost(data, sign, order)
		revenue := round(sign.SignedQty * unitPrice)
		cost := round(sign.SignedQty * costPrice)
		operating.SignedQty = round(operating.SignedQty + sign.SignedQty)
		operating.Revenue = round(operating.Revenue + revenue)
		operating.MaterialCost = round(operating.MaterialCost + cost)
		p := profits[order.ProjectID]
		if p == nil {
			p = &ProjectProfitReport{ProjectID: order.ProjectID, Period: periodString()}
			profits[order.ProjectID] = p
		}
		p.Revenue = round(p.Revenue + revenue)
		p.Cost = round(p.Cost + cost)
		p.Profit = round(p.Revenue - p.Cost)
		p.Margin = moneyPercent(p.Profit, p.Revenue)
	}
	operating.TotalCost = round(operating.MaterialCost + operating.TransportCost)
	operating.GrossProfit = round(operating.Revenue - operating.TotalCost)
	operating.GrossMargin = moneyPercent(operating.GrossProfit, operating.Revenue)
	out := make([]ProjectProfitReport, 0, len(profits))
	for _, p := range profits {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Profit == out[j].Profit {
			return out[i].ProjectID < out[j].ProjectID
		}
		return out[i].Profit > out[j].Profit
	})
	return out, operating
}

func buildCustomerAgingReport(data AppData) []CustomerAgingReport {
	items := map[int64]*CustomerAgingReport{}
	for _, customer := range data.Customers {
		items[customer.ID] = &CustomerAgingReport{CustomerID: customer.ID, CustomerName: customer.Name}
	}
	for _, receivable := range data.Receivables {
		remaining := remainingReceivable(receivable)
		if remaining <= 0 || receivable.Status == "paid" {
			continue
		}
		item := items[receivable.CustomerID]
		if item == nil {
			item = &CustomerAgingReport{CustomerID: receivable.CustomerID}
			if customer, ok := findCustomer(data, receivable.CustomerID); ok {
				item.CustomerName = customer.Name
			}
			items[receivable.CustomerID] = item
		}
		days := receivableOverdueDays(receivable)
		switch {
		case days <= 0:
			item.Current = round(item.Current + remaining)
		case days <= 30:
			item.Overdue1To30 = round(item.Overdue1To30 + remaining)
		case days <= 60:
			item.Overdue31To60 = round(item.Overdue31To60 + remaining)
		case days <= 90:
			item.Overdue61To90 = round(item.Overdue61To90 + remaining)
		default:
			item.OverdueOver90 = round(item.OverdueOver90 + remaining)
		}
		item.Total = round(item.Total + remaining)
		if days > 0 {
			item.OverdueTotal = round(item.OverdueTotal + remaining)
		}
	}
	out := make([]CustomerAgingReport, 0, len(items))
	for _, item := range items {
		if item.Total > 0 || item.OverdueTotal > 0 {
			out = append(out, *item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].OverdueTotal == out[j].OverdueTotal {
			return out[i].Total > out[j].Total
		}
		return out[i].OverdueTotal > out[j].OverdueTotal
	})
	return out
}

func buildQualityAnalysisReport(data AppData) QualityAnalysisReport {
	report := QualityAnalysisReport{}
	for _, item := range data.QualityInspections {
		report.Inspections++
		switch item.Result {
		case "passed":
			report.Passed++
		case "failed", "rejected":
			report.Failed++
			report.OpenExceptions = append(report.OpenExceptions, item)
		default:
			report.Pending++
			if item.Status != "completed" {
				report.OpenExceptions = append(report.OpenExceptions, item)
			}
		}
	}
	for _, item := range data.QualitySamples {
		report.Samples++
		switch item.Result {
		case "passed":
			report.SamplePassed++
		case "failed", "rejected":
			report.SampleFailed++
		default:
			report.SamplePending++
		}
	}
	for _, item := range data.LaboratoryTests {
		report.LaboratoryTests++
		if item.Result == "failed" && item.Status != "reviewed" {
			report.SampleFailed++
		}
	}
	for _, item := range data.QualityExceptions {
		if item.Status != "closed" {
			report.OpenExceptionCount++
			report.QualityExceptions = append(report.QualityExceptions, item)
		}
	}
	report.PassRate = moneyPercent(float64(report.Passed), float64(report.Inspections))
	return report
}

func buildProductionEnergyReports(data AppData) []ProductionEnergyReport {
	bySite := map[int64]*ProductionEnergyReport{}
	for _, batch := range data.ProductionBatches {
		if batch.Status != "completed" && batch.Status != "released" {
			continue
		}
		item := bySite[batch.SiteID]
		if item == nil {
			item = &ProductionEnergyReport{SiteID: batch.SiteID}
			bySite[batch.SiteID] = item
		}
		item.ProducedQty = round(item.ProducedQty + batch.Quantity)
		item.BatchCount++
		if product, ok := findProduct(data, batch.ProductID); ok {
			item.MaterialCost = round(item.MaterialCost + product.CostPrice*batch.Quantity)
		}
		item.EstimatedPowerKwh = round(item.EstimatedPowerKwh + batch.Quantity*2.6)
	}
	for _, flow := range data.InventoryFlows {
		if flow.Direction != "out" || flow.SourceType != "production_batch" {
			continue
		}
		batch, ok := findProductionBatch(data, flow.SourceID)
		if !ok {
			continue
		}
		item := bySite[batch.SiteID]
		if item == nil {
			item = &ProductionEnergyReport{SiteID: batch.SiteID}
			bySite[batch.SiteID] = item
		}
		item.MaterialUsageQty = round(item.MaterialUsageQty + flow.Quantity)
	}
	out := make([]ProductionEnergyReport, 0, len(bySite))
	for _, item := range bySite {
		if item.ProducedQty > 0 {
			item.UnitMaterialCost = round(item.MaterialCost / item.ProducedQty)
			item.UnitPowerKwh = round(item.EstimatedPowerKwh / item.ProducedQty)
		}
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ProducedQty > out[j].ProducedQty })
	return out
}

func signedUnitAndCost(data AppData, sign DeliverySign, order SalesOrder) (float64, float64) {
	productID := order.ProductID
	unitPrice := order.UnitPrice
	if sign.LineID != 0 {
		if line, ok := findOrderLine(order, sign.LineID); ok {
			productID = line.ProductID
			unitPrice = line.UnitPrice
		}
	}
	costPrice := 0.0
	if product, ok := findProduct(data, productID); ok {
		costPrice = product.CostPrice
	}
	return unitPrice, costPrice
}

func receivableOverdueDays(receivable Receivable) int {
	today, _ := time.Parse("2006-01-02", todayString())
	due, err := time.Parse("2006-01-02", receivable.DueDate)
	if err != nil {
		return 0
	}
	days := int(today.Sub(due).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

func moneyPercent(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return round(part / total * 100)
}
