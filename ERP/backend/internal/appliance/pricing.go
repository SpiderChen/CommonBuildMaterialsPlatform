package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

type PricingQuote struct {
	CustomerID       int64   `json:"customerId"`
	ProjectID        int64   `json:"projectId"`
	ProductID        int64   `json:"productId"`
	PolicyID         int64   `json:"policyId"`
	CustomerGrade    string  `json:"customerGrade"`
	Region           string  `json:"region"`
	MinQuantity      float64 `json:"minQuantity"`
	MaxQuantity      float64 `json:"maxQuantity"`
	Source           string  `json:"source"`
	ListPrice        float64 `json:"listPrice"`
	UnitPrice        float64 `json:"unitPrice"`
	FloorPrice       float64 `json:"floorPrice"`
	PromotionName    string  `json:"promotionName"`
	PromotionType    string  `json:"promotionType"`
	PromotionValue   float64 `json:"promotionValue"`
	PromotionAmount  float64 `json:"promotionAmount"`
	TaxRateID        int64   `json:"taxRateId"`
	TaxRateName      string  `json:"taxRateName"`
	TaxRate          float64 `json:"taxRate"`
	BelowFloor       bool    `json:"belowFloor"`
	ApprovalRequired bool    `json:"approvalRequired"`
	Reason           string  `json:"reason"`
}

type pricingEvaluateRequest struct {
	CustomerID   int64   `json:"customerId"`
	ProjectID    int64   `json:"projectId"`
	ProductID    int64   `json:"productId"`
	PlanTime     string  `json:"planTime"`
	PlanQuantity float64 `json:"planQuantity"`
	UnitPrice    float64 `json:"unitPrice"`
}

func (a *App) evaluatePricing(w http.ResponseWriter, r *http.Request, session Session) {
	var req pricingEvaluateRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid pricing payload")
		return
	}
	data := scopedData(a.mustSnapshot(), session.User)
	customer, ok := findCustomer(data, req.CustomerID)
	if !ok {
		writeError(w, http.StatusBadRequest, "客户不存在")
		return
	}
	project, ok := findProject(data, req.ProjectID)
	if !ok {
		writeError(w, http.StatusBadRequest, "项目不存在")
		return
	}
	product, ok := findProduct(data, req.ProductID)
	if !ok {
		writeError(w, http.StatusBadRequest, "产品不存在")
		return
	}
	contract, _ := activeContract(data, req.CustomerID, req.ProjectID, req.ProductID)
	order := SalesOrder{CustomerID: req.CustomerID, ProjectID: req.ProjectID, ProductID: req.ProductID, PlanTime: req.PlanTime, PlanQuantity: req.PlanQuantity, UnitPrice: req.UnitPrice}
	writeJSON(w, http.StatusOK, priceQuote(data, order, customer, project, product, contract))
}

func (a *App) createPricePolicy(w http.ResponseWriter, r *http.Request, session Session) {
	var item PricePolicy
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid price policy")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if _, ok := findProduct(*data, item.ProductID); !ok {
			return fmt.Errorf("产品不存在")
		}
		if item.CustomerID != 0 {
			if _, ok := findCustomer(*data, item.CustomerID); !ok {
				return fmt.Errorf("客户不存在")
			}
		}
		if item.ProjectID != 0 {
			if _, ok := findProject(*data, item.ProjectID); !ok {
				return fmt.Errorf("项目不存在")
			}
		}
		if item.TaxRateID != 0 {
			if _, ok := findTaxRate(*data, item.TaxRateID); !ok {
				return fmt.Errorf("税率不存在")
			}
		}
		if item.SalePrice <= 0 {
			return fmt.Errorf("销售价必须大于 0")
		}
		if item.FloorPrice < 0 {
			return fmt.Errorf("底价不能小于 0")
		}
		if item.MinQuantity < 0 || item.MaxQuantity < 0 {
			return fmt.Errorf("阶梯数量不能小于 0")
		}
		if item.MaxQuantity > 0 && item.MaxQuantity < item.MinQuantity {
			return fmt.Errorf("阶梯最大量不能小于最小量")
		}
		item.ID = nextID(data, "pricePolicy")
		item.CustomerGrade = strings.ToUpper(strings.TrimSpace(item.CustomerGrade))
		item.Region = strings.TrimSpace(item.Region)
		item.PromotionName = strings.TrimSpace(item.PromotionName)
		item.PromotionType = strings.ToLower(strings.TrimSpace(item.PromotionType))
		switch item.PromotionType {
		case "", "none":
			item.PromotionType = ""
			item.PromotionValue = 0
		case "fixed":
			if item.PromotionValue <= 0 {
				return fmt.Errorf("固定促销金额必须大于 0")
			}
			if item.PromotionValue >= item.SalePrice {
				return fmt.Errorf("固定促销金额必须小于销售价")
			}
		case "percent":
			if item.PromotionValue <= 0 || item.PromotionValue >= 1 {
				return fmt.Errorf("百分比促销必须在 0 到 1 之间")
			}
		default:
			return fmt.Errorf("促销类型仅支持 fixed 或 percent")
		}
		item.Status = fallback(item.Status, "active")
		item.EffectiveFrom = fallback(item.EffectiveFrom, todayString())
		item.EffectiveTo = fallback(item.EffectiveTo, "2099-12-31")
		data.PricePolicies = append(data.PricePolicies, item)
		addAudit(data, session.User.Username, "create", "price_policy", item.ID, fmt.Sprintf("product=%d price=%.2f", item.ProductID, item.SalePrice), clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.price_policy.created")
}

func (a *App) createTaxRate(w http.ResponseWriter, r *http.Request, session Session) {
	var item TaxRate
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid tax rate")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if item.Rate < 0 || item.Rate > 1 {
			return fmt.Errorf("税率必须在 0 到 1 之间")
		}
		item.ID = nextID(data, "taxRate")
		item.Scope = fallback(item.Scope, "sales")
		item.Status = fallback(item.Status, "active")
		data.TaxRates = append(data.TaxRates, item)
		addAudit(data, session.User.Username, "create", "tax_rate", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "master.tax_rate.created")
}

func priceQuote(data AppData, order SalesOrder, customer Customer, project Project, product Product, contract Contract) PricingQuote {
	grade := customerGrade(data, customer.ID)
	contractPrice := contractUnitPrice(contract, product.ID)
	source := "contract"
	unitPrice := contractPrice
	if unitPrice == 0 {
		unitPrice = product.BasePrice
		source = "product"
	}
	listPrice := unitPrice
	var promotionAmount float64
	floorPrice := product.CostPrice
	var region, promotionName, promotionType string
	var minQuantity, maxQuantity, promotionValue float64
	var policyID, taxRateID int64
	if policy, ok := bestPricePolicy(data, customer, project, product.ID, grade, priceDate(order.PlanTime), order.PlanQuantity); ok {
		policyID = policy.ID
		taxRateID = policy.TaxRateID
		listPrice = policy.SalePrice
		unitPrice, promotionAmount = applyPolicyPromotion(policy)
		floorPrice = policy.FloorPrice
		region = policy.Region
		minQuantity = policy.MinQuantity
		maxQuantity = policy.MaxQuantity
		promotionName = policy.PromotionName
		promotionType = policy.PromotionType
		promotionValue = policy.PromotionValue
		source = "price_policy"
	}
	if order.UnitPrice > 0 {
		unitPrice = order.UnitPrice
		source += "_manual"
	}
	taxRate, _ := effectiveTaxRate(data, taxRateID)
	quote := PricingQuote{
		CustomerID: order.CustomerID, ProjectID: order.ProjectID, ProductID: order.ProductID,
		PolicyID: policyID, CustomerGrade: grade, Region: region, MinQuantity: round(minQuantity), MaxQuantity: round(maxQuantity),
		Source: source, ListPrice: round(listPrice), UnitPrice: round(unitPrice), FloorPrice: round(floorPrice),
		PromotionName: promotionName, PromotionType: promotionType, PromotionValue: round(promotionValue), PromotionAmount: round(promotionAmount),
		TaxRateID: taxRate.ID, TaxRateName: taxRate.Name, TaxRate: taxRate.Rate,
	}
	if quote.FloorPrice > 0 && quote.UnitPrice < quote.FloorPrice {
		quote.BelowFloor = true
		quote.ApprovalRequired = true
		quote.Reason = fmt.Sprintf("销售单价 %.2f 低于底价 %.2f", quote.UnitPrice, quote.FloorPrice)
	}
	return quote
}

func bestPricePolicy(data AppData, customer Customer, project Project, productID int64, grade string, date string, quantity float64) (PricePolicy, bool) {
	var best PricePolicy
	bestScore := -1
	for _, item := range data.PricePolicies {
		if item.Status != "active" || item.ProductID != productID || !policyDateActive(item, date) {
			continue
		}
		if item.MinQuantity > 0 && quantity < item.MinQuantity {
			continue
		}
		if item.MaxQuantity > 0 && quantity > item.MaxQuantity {
			continue
		}
		score := 0
		if item.CustomerID != 0 {
			if item.CustomerID != customer.ID {
				continue
			}
			score += 1000
		}
		if item.ProjectID != 0 {
			if item.ProjectID != project.ID {
				continue
			}
			score += 500
		}
		if item.CustomerGrade != "" {
			if !strings.EqualFold(item.CustomerGrade, grade) {
				continue
			}
			score += 300
		}
		if item.Region != "" {
			if !projectMatchesRegion(project, item.Region) {
				continue
			}
			score += 200
		}
		if item.MinQuantity > 0 || item.MaxQuantity > 0 {
			score += 100
		}
		if item.PromotionType != "" {
			score += 50
		}
		score += item.Priority
		if score > bestScore || (score == bestScore && item.ID > best.ID) {
			best = item
			bestScore = score
		}
	}
	return best, bestScore >= 0
}

func applyPolicyPromotion(policy PricePolicy) (float64, float64) {
	unitPrice := policy.SalePrice
	promotionAmount := 0.0
	switch policy.PromotionType {
	case "fixed":
		promotionAmount = policy.PromotionValue
		unitPrice = policy.SalePrice - promotionAmount
	case "percent":
		promotionAmount = policy.SalePrice * policy.PromotionValue
		unitPrice = policy.SalePrice - promotionAmount
	}
	if unitPrice < 0 {
		unitPrice = 0
	}
	return round(unitPrice), round(promotionAmount)
}

func projectMatchesRegion(project Project, region string) bool {
	needle := strings.ToLower(strings.TrimSpace(region))
	if needle == "" {
		return true
	}
	haystack := strings.ToLower(strings.TrimSpace(project.Name + " " + project.Address))
	return strings.Contains(haystack, needle)
}

func policyDateActive(item PricePolicy, date string) bool {
	if date == "" {
		date = todayString()
	}
	if item.EffectiveFrom != "" && date < item.EffectiveFrom {
		return false
	}
	if item.EffectiveTo != "" && date > item.EffectiveTo {
		return false
	}
	return true
}

func priceDate(planTime string) string {
	if len(planTime) >= 10 {
		return planTime[:10]
	}
	return todayString()
}

func customerGrade(data AppData, customerID int64) string {
	for _, item := range data.CustomerProfiles {
		if item.CustomerID == customerID && item.Status == "active" && item.Grade != "" {
			return strings.ToUpper(item.Grade)
		}
	}
	return "C"
}

func contractUnitPrice(contract Contract, productID int64) float64 {
	for _, item := range contract.Items {
		if item.ProductID == productID {
			return item.UnitPrice
		}
	}
	if len(contract.Items) > 0 {
		return contract.Items[0].UnitPrice
	}
	return 0
}

func effectiveTaxRate(data AppData, taxRateID int64) (TaxRate, bool) {
	if taxRateID != 0 {
		if item, ok := findTaxRate(data, taxRateID); ok && item.Status == "active" {
			return item, true
		}
	}
	for _, item := range data.TaxRates {
		if item.Status == "active" && item.Scope == "sales" {
			return item, true
		}
	}
	return TaxRate{}, false
}

func findTaxRate(data AppData, id int64) (TaxRate, bool) {
	for _, item := range data.TaxRates {
		if item.ID == id {
			return item, true
		}
	}
	return TaxRate{}, false
}
