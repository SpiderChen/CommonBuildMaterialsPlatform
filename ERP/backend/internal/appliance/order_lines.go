package appliance

import (
	"fmt"
	"strings"
)

func prepareSalesOrderLines(data *AppData, item *SalesOrder, customer Customer, project Project) ([]string, []string, error) {
	inputs := item.Lines
	if len(inputs) == 0 {
		inputs = []SalesOrderLine{{
			ProductID:     item.ProductID,
			StrengthGrade: item.StrengthGrade,
			Slump:         item.Slump,
			PouringPart:   item.PouringPart,
			Quantity:      item.PlanQuantity,
			UnitPrice:     item.UnitPrice,
		}}
	}

	var riskFlags []string
	var riskReasons []string
	var totalQty float64
	var totalAmount float64
	lines := make([]SalesOrderLine, 0, len(inputs))

	for index, input := range inputs {
		productID := nonZeroInt(input.ProductID, item.ProductID)
		product, ok := findProduct(*data, productID)
		if !ok {
			return nil, nil, fmt.Errorf("订单明细第 %d 行产品不存在", index+1)
		}
		quantity := nonZero(input.Quantity, item.PlanQuantity)
		if quantity <= 0 {
			return nil, nil, fmt.Errorf("订单明细第 %d 行数量必须大于 0", index+1)
		}
		contract, ok := activeContract(*data, item.CustomerID, item.ProjectID, product.ID)
		if !ok {
			return nil, nil, fmt.Errorf("订单明细第 %d 行无有效合同不允许下单", index+1)
		}
		quoteOrder := *item
		quoteOrder.ProductID = product.ID
		quoteOrder.PlanQuantity = quantity
		quoteOrder.UnitPrice = input.UnitPrice
		quote := priceQuote(*data, quoteOrder, customer, project, product, contract)

		line := SalesOrderLine{
			ID:            nextID(data, "orderLine"),
			Seq:           index + 1,
			ProductID:     product.ID,
			ProductLine:   product.Line,
			ProductName:   productName(product),
			StrengthGrade: fallback(input.StrengthGrade, item.StrengthGrade),
			Slump:         fallback(input.Slump, item.Slump),
			PouringPart:   fallback(input.PouringPart, item.PouringPart),
			Quantity:      round(quantity),
			Unit:          product.Unit,
			UnitPrice:     quote.UnitPrice,
			FloorPrice:    quote.FloorPrice,
			TaxRate:       quote.TaxRate,
			Amount:        round(quantity * quote.UnitPrice),
			PriceSource:   quote.Source,
		}
		if quote.BelowFloor {
			line.RiskFlag = "price_below_floor"
			riskFlags = appendOrderRisk(riskFlags, "price_below_floor")
			riskReasons = append(riskReasons, fmt.Sprintf("第 %d 行 %s：%s", line.Seq, line.ProductName, quote.Reason))
		}

		if index == 0 {
			item.ProductID = line.ProductID
			item.ProductLine = line.ProductLine
			item.Unit = line.Unit
			item.StrengthGrade = fallback(item.StrengthGrade, line.StrengthGrade)
			item.Slump = fallback(item.Slump, line.Slump)
			item.PouringPart = fallback(item.PouringPart, line.PouringPart)
		}
		totalQty = round(totalQty + line.Quantity)
		totalAmount = round(totalAmount + line.Amount)
		lines = append(lines, line)
	}

	item.PlanQuantity = totalQty
	item.TotalAmount = totalAmount
	if totalQty > 0 {
		item.UnitPrice = round(totalAmount / totalQty)
	}
	item.Lines = lines
	return riskFlags, riskReasons, nil
}

func orderLines(order SalesOrder) []SalesOrderLine {
	if len(order.Lines) > 0 {
		return order.Lines
	}
	if order.ProductID == 0 {
		return nil
	}
	return []SalesOrderLine{{
		ID:            order.ID * 1000,
		Seq:           1,
		ProductID:     order.ProductID,
		ProductLine:   order.ProductLine,
		StrengthGrade: order.StrengthGrade,
		Slump:         order.Slump,
		PouringPart:   order.PouringPart,
		Quantity:      order.PlanQuantity,
		Unit:          order.Unit,
		UnitPrice:     order.UnitPrice,
		Amount:        orderTotalAmount(order),
		PriceSource:   "legacy",
	}}
}

func findOrderLine(order SalesOrder, lineID int64) (SalesOrderLine, bool) {
	for _, line := range orderLines(order) {
		if line.ID == lineID {
			return line, true
		}
	}
	return SalesOrderLine{}, false
}

func dispatchedQtyForOrderLine(data AppData, orderID int64, lineID int64) float64 {
	total := 0.0
	for _, dispatch := range data.DispatchOrders {
		if dispatch.OrderID == orderID && dispatch.LineID == lineID && dispatch.Status != "cancelled" {
			total = round(total + dispatch.PlanQuantity)
		}
	}
	return total
}

func orderTotalAmount(order SalesOrder) float64 {
	if order.TotalAmount > 0 {
		return order.TotalAmount
	}
	return round(order.PlanQuantity * order.UnitPrice)
}

func productName(product Product) string {
	return strings.TrimSpace(product.Name + " " + product.Spec)
}

func appendOrderRisk(flags []string, flag string) []string {
	for _, existing := range flags {
		if existing == flag {
			return flags
		}
	}
	return append(flags, flag)
}
