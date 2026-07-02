package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

func (a *App) createContract(w http.ResponseWriter, r *http.Request, session Session) {
	var item Contract
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid contract")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if err := normalizeContractDraft(data, &item); err != nil {
			return err
		}
		item.ID = nextID(data, "contract")
		item.ContractNo = fallback(item.ContractNo, number("CT", item.ID))
		item.Version = nonZeroIntAsInt(item.Version, 1)
		item.Status = fallback(item.Status, "draft")
		if item.Status == "active" {
			item.Status = "draft"
		}
		data.Contracts = append(data.Contracts, item)
		addAudit(data, session.User.Username, "create", "contract", item.ID, item.ContractNo, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "contract.created")
}

func (a *App) submitContract(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req struct {
		Reason string `json:"reason"`
	}
	_ = readJSON(r, &req)
	var item Contract
	err := a.store.Mutate(func(data *AppData) error {
		index := contractIndex(*data, id)
		if index < 0 {
			return fmt.Errorf("合同不存在")
		}
		if data.Contracts[index].Status == "active" || data.Contracts[index].Status == "superseded" {
			return fmt.Errorf("只有草稿或驳回的合同版本可以提交审批")
		}
		if len(data.Contracts[index].Items) == 0 {
			return fmt.Errorf("合同版本必须包含价格明细")
		}
		data.Contracts[index].Status = "pending_approval"
		data.Contracts[index].SubmittedAt = nowString()
		data.Contracts[index].ChangeReason = fallback(strings.TrimSpace(req.Reason), data.Contracts[index].ChangeReason)
		item = data.Contracts[index]
		_, instances, err := publishWorkflowEvent(data, workflowEventRequest{
			EventType:  "contract.submitted",
			Resource:   "contract",
			ResourceID: item.ID,
			ResourceNo: contractVersionNo(item),
			Title:      "客户合同版本审批",
			Actor:      session.User.Username,
			Reason:     fallback(item.ChangeReason, "合同版本提交审批"),
			Variables: map[string]string{
				"customerId": fmt.Sprintf("%d", item.CustomerID),
				"projectId":  fmt.Sprintf("%d", item.ProjectID),
				"version":    fmt.Sprintf("%d", item.Version),
			},
		})
		if err != nil {
			return err
		}
		if len(instances) == 0 {
			return fmt.Errorf("合同版本工作流未配置")
		}
		addAudit(data, session.User.Username, "submit", "contract", item.ID, contractVersionNo(item), clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "contract.submitted")
}

func (a *App) reviseContract(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req Contract
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid contract revision")
		return
	}
	var item Contract
	err := a.store.Mutate(func(data *AppData) error {
		baseIndex := contractIndex(*data, id)
		if baseIndex < 0 {
			return fmt.Errorf("合同不存在")
		}
		base := data.Contracts[baseIndex]
		item = base
		mergeContractRevision(&item, req)
		item.ID = nextID(data, "contract")
		item.ParentID = base.ID
		item.ContractNo = fallback(base.ContractNo, number("CT", item.ID))
		item.Version = nextContractVersion(*data, item.ContractNo)
		item.Status = "draft"
		item.SubmittedAt = ""
		item.ApprovedAt = ""
		item.ApprovedBy = ""
		if err := normalizeContractDraft(data, &item); err != nil {
			return err
		}
		data.Contracts = append(data.Contracts, item)
		addAudit(data, session.User.Username, "revise", "contract", item.ID, contractVersionNo(item), clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "contract.revised")
}

func normalizeContractDraft(data *AppData, item *Contract) error {
	if _, ok := findCustomer(*data, item.CustomerID); !ok {
		return fmt.Errorf("客户不存在")
	}
	if _, ok := findProject(*data, item.ProjectID); !ok {
		return fmt.Errorf("项目不存在")
	}
	item.Name = fallback(strings.TrimSpace(item.Name), "客户供应合同")
	item.ValidFrom = strings.TrimSpace(item.ValidFrom)
	item.ValidTo = strings.TrimSpace(item.ValidTo)
	item.CreditPolicy = fallback(strings.TrimSpace(item.CreditPolicy), "按合同和授信执行")
	item.ChangeReason = strings.TrimSpace(item.ChangeReason)
	if len(item.Items) == 0 {
		return fmt.Errorf("合同必须包含产品价格明细")
	}
	for i := range item.Items {
		if item.Items[i].ProductID == 0 {
			return fmt.Errorf("合同明细第 %d 行产品不能为空", i+1)
		}
		product, ok := findProduct(*data, item.Items[i].ProductID)
		if !ok {
			return fmt.Errorf("合同明细第 %d 行产品不存在", i+1)
		}
		item.Items[i].Unit = fallback(strings.TrimSpace(item.Items[i].Unit), product.Unit)
		if item.Items[i].Quantity <= 0 {
			return fmt.Errorf("合同明细第 %d 行数量必须大于 0", i+1)
		}
		if item.Items[i].UnitPrice <= 0 {
			return fmt.Errorf("合同明细第 %d 行单价必须大于 0", i+1)
		}
	}
	return nil
}

func mergeContractRevision(target *Contract, req Contract) {
	target.Name = fallback(req.Name, target.Name)
	target.ValidFrom = fallback(req.ValidFrom, target.ValidFrom)
	target.ValidTo = fallback(req.ValidTo, target.ValidTo)
	target.CreditPolicy = fallback(req.CreditPolicy, target.CreditPolicy)
	target.ChangeReason = fallback(req.ChangeReason, target.ChangeReason)
	if req.TotalAmount > 0 {
		target.TotalAmount = req.TotalAmount
	}
	if len(req.Items) > 0 {
		target.Items = req.Items
	}
}

func contractIndex(data AppData, id int64) int {
	for i := range data.Contracts {
		if data.Contracts[i].ID == id {
			return i
		}
	}
	return -1
}

func nextContractVersion(data AppData, contractNo string) int {
	version := 0
	for _, item := range data.Contracts {
		if item.ContractNo == contractNo && item.Version > version {
			version = item.Version
		}
	}
	if version == 0 {
		version = 1
	}
	return version + 1
}

func contractVersionNo(item Contract) string {
	version := item.Version
	if version <= 0 {
		version = 1
	}
	return fmt.Sprintf("%s-v%d", item.ContractNo, version)
}

func nonZeroIntAsInt(value, fallbackValue int) int {
	if value == 0 {
		return fallbackValue
	}
	return value
}
