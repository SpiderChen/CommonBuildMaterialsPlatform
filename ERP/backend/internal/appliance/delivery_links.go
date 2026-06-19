package appliance

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type deliverySignRequest struct {
	DeliverySign
	Attachments []DeliverySignAttachment `json:"attachments"`
}

type deliverySignLinkRequest struct {
	DispatchID int64  `json:"dispatchId"`
	TicketID   int64  `json:"ticketId"`
	Channel    string `json:"channel"`
	Phone      string `json:"phone"`
	ExpiresAt  string `json:"expiresAt"`
}

type publicDeliverySignDetail struct {
	Link        DeliverySignLink         `json:"link"`
	Dispatch    DispatchOrder            `json:"dispatch"`
	Ticket      ScaleTicket              `json:"ticket"`
	Order       SalesOrder               `json:"order"`
	Customer    string                   `json:"customer"`
	Project     string                   `json:"project"`
	Product     string                   `json:"product"`
	PlateNo     string                   `json:"plateNo"`
	Attachments []DeliverySignAttachment `json:"attachments"`
}

func (a *App) publicDeliverySign(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) != 1 {
		writeError(w, http.StatusNotFound, "unknown public sign route")
		return
	}
	token := strings.TrimSpace(parts[0])
	if token == "" {
		writeError(w, http.StatusNotFound, "sign token not found")
		return
	}
	if r.Method == http.MethodGet {
		detail, err := buildPublicDeliverySignDetail(a.mustSnapshot(), token)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, detail)
		return
	}
	if r.Method == http.MethodPost {
		a.signDeliveryByToken(w, r, token)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (a *App) listDeliverySignLinks(w http.ResponseWriter, r *http.Request, session Session) {
	writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).DeliverySignLinks)
}

func (a *App) listDeliverySignAttachments(w http.ResponseWriter, r *http.Request, session Session) {
	writeJSON(w, http.StatusOK, scopedData(a.mustSnapshot(), session.User).DeliverySignAttachments)
}

func (a *App) createDeliverySignLink(w http.ResponseWriter, r *http.Request, session Session) {
	var req deliverySignLinkRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid sign link")
		return
	}
	var link DeliverySignLink
	err := a.store.Mutate(func(data *AppData) error {
		next, err := buildDeliverySignLink(data, req, session.User.Username)
		if err != nil {
			return err
		}
		link = next
		data.DeliverySignLinks = append(data.DeliverySignLinks, link)
		upsertDeliveryNoteQRCode(data, link)
		addAudit(data, session.User.Username, "send", "delivery_sign_link", link.ID, link.LinkNo+" "+link.Channel, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, link, "delivery.sign_link.sent")
}

func (a *App) addDeliverySignAttachment(w http.ResponseWriter, r *http.Request, session Session, signID int64) {
	var item DeliverySignAttachment
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid sign attachment")
		return
	}
	var created DeliverySignAttachment
	err := a.store.Mutate(func(data *AppData) error {
		sign, ok := findDeliverySign(*data, signID)
		if !ok {
			return fmt.Errorf("签收单不存在")
		}
		item.SignID = sign.ID
		item.DispatchID = sign.DispatchID
		item.TicketID = sign.TicketID
		created = createDeliverySignAttachment(data, item, session.User.Username)
		addAudit(data, session.User.Username, "upload", "delivery_sign_attachment", created.ID, created.FileName, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, created, "delivery.sign_attachment.created")
}

func (a *App) signDeliveryByToken(w http.ResponseWriter, r *http.Request, token string) {
	var req deliverySignRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid sign")
		return
	}
	var created DeliverySign
	err := a.store.Mutate(func(data *AppData) error {
		linkIndex, link, err := findUsableDeliverySignLink(*data, token)
		if err != nil {
			return err
		}
		req.DeliverySign.DispatchID = link.DispatchID
		req.DeliverySign.LinkID = link.ID
		req.DeliverySign.TicketID = link.TicketID
		req.DeliverySign.LineID = link.LineID
		req.DeliverySign.LineSeq = link.LineSeq
		req.DeliverySign.ProductID = link.ProductID
		req.DeliverySign.ProductName = link.ProductName
		if req.DeliverySign.Phone == "" {
			req.DeliverySign.Phone = link.Phone
		}
		sign, err := completeDeliverySign(data, req.DeliverySign, req.Attachments, "public:"+link.LinkNo, clientIP(r))
		if err != nil {
			return err
		}
		created = sign
		data.DeliverySignLinks[linkIndex].Status = "used"
		data.DeliverySignLinks[linkIndex].UsedAt = sign.SignedAt
		return nil
	})
	a.respondMutation(w, err, created, "delivery.signed")
}

func completeDeliverySign(data *AppData, item DeliverySign, attachments []DeliverySignAttachment, actor string, ip string) (DeliverySign, error) {
	dispatch, ok := findDispatch(*data, item.DispatchID)
	if !ok {
		return DeliverySign{}, fmt.Errorf("派车单不存在")
	}
	if _, ok := findDeliverySignByDispatch(*data, dispatch.ID); ok {
		return DeliverySign{}, fmt.Errorf("派车单已签收")
	}
	order, ok := findOrder(*data, dispatch.OrderID)
	if !ok {
		return DeliverySign{}, fmt.Errorf("订单不存在")
	}
	if item.TicketID == 0 {
		for _, ticket := range data.ScaleTickets {
			if ticket.DispatchID == dispatch.ID && ticket.Status == "valid" {
				item.TicketID = ticket.ID
				break
			}
		}
	}
	item.ID = nextID(data, "sign")
	item.SignNo = number("DS", item.ID)
	item.OrderID = order.ID
	item.LineID = dispatch.LineID
	item.LineSeq = dispatch.LineSeq
	item.ProductID = dispatch.ProductID
	item.ProductName = dispatch.ProductName
	if item.ProductID == 0 {
		item.ProductID = order.ProductID
	}
	if item.ProductName == "" && item.ProductID != 0 {
		if product, ok := findProduct(*data, item.ProductID); ok {
			item.ProductName = productName(product)
		}
	}
	item.CustomerID = order.CustomerID
	item.ProjectID = order.ProjectID
	item.SignedQty = nonZero(item.SignedQty, dispatch.PlanQuantity)
	if item.SignedQty <= 0 {
		return DeliverySign{}, fmt.Errorf("签收数量必须大于 0")
	}
	if item.SignedQty > dispatch.PlanQuantity {
		return DeliverySign{}, fmt.Errorf("签收数量不能大于派车数量")
	}
	item.SignedAt = nowString()
	data.DeliverySigns = append(data.DeliverySigns, item)
	for i := range data.DispatchOrders {
		if data.DispatchOrders[i].ID == dispatch.ID {
			data.DispatchOrders[i].SignedQty = item.SignedQty
			data.DispatchOrders[i].Status = "completed"
			data.DispatchOrders[i].UpdatedAt = nowString()
		}
	}
	for i := range data.ScaleTickets {
		if data.ScaleTickets[i].ID == item.TicketID {
			data.ScaleTickets[i].SignStatus = "signed"
		}
	}
	for i := range data.Orders {
		if data.Orders[i].ID == order.ID {
			data.Orders[i].SignedQty += item.SignedQty
			if data.Orders[i].SignedQty >= data.Orders[i].PlanQuantity {
				data.Orders[i].Status = "completed"
			}
		}
	}
	for i := range data.DeliveryNotes {
		if data.DeliveryNotes[i].DispatchID == dispatch.ID {
			data.DeliveryNotes[i].Status = "signed"
		}
	}
	if len(attachments) == 0 && item.Photo != "" {
		attachments = append(attachments, DeliverySignAttachment{FileName: "现场签收照片", FileType: "photo", URL: item.Photo})
	}
	for _, attachment := range attachments {
		attachment.SignID = item.ID
		attachment.DispatchID = item.DispatchID
		attachment.TicketID = item.TicketID
		createDeliverySignAttachment(data, attachment, firstNonEmpty(item.Signer, actor))
	}
	updateVehicleStatus(data, dispatch.VehicleID, "returning")
	upsertStatement(data, item, order)
	addAudit(data, actor, "create", "delivery_sign", item.ID, item.SignNo, ip)
	return item, nil
}

func buildDeliverySignLink(data *AppData, req deliverySignLinkRequest, actor string) (DeliverySignLink, error) {
	dispatch, ok := findDispatch(*data, req.DispatchID)
	if !ok {
		return DeliverySignLink{}, fmt.Errorf("派车单不存在")
	}
	if _, ok := findDeliverySignByDispatch(*data, dispatch.ID); ok {
		return DeliverySignLink{}, fmt.Errorf("派车单已签收")
	}
	order, ok := findOrder(*data, dispatch.OrderID)
	if !ok {
		return DeliverySignLink{}, fmt.Errorf("订单不存在")
	}
	if req.TicketID == 0 {
		for _, ticket := range data.ScaleTickets {
			if ticket.DispatchID == dispatch.ID && ticket.Status == "valid" {
				req.TicketID = ticket.ID
				break
			}
		}
	}
	channel := strings.TrimSpace(req.Channel)
	if channel == "" {
		channel = "sms"
	}
	expiresAt := strings.TrimSpace(req.ExpiresAt)
	if expiresAt == "" {
		expiresAt = time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	}
	id := nextID(data, "signLink")
	token := tokenString()
	url := deliverySignPublicURL(token)
	link := DeliverySignLink{
		ID: id, LinkNo: number("SL", id), DispatchID: dispatch.ID, TicketID: req.TicketID,
		OrderID: order.ID, LineID: dispatch.LineID, LineSeq: dispatch.LineSeq, ProductID: dispatch.ProductID, ProductName: dispatch.ProductName,
		CustomerID: order.CustomerID, ProjectID: order.ProjectID,
		Channel: channel, Phone: firstNonEmpty(req.Phone, order.Phone), Token: token,
		URL: url, QRCode: "qr://" + url, Status: "sent", SentAt: nowString(),
		ExpiresAt: expiresAt, CreatedBy: actor, CreatedAt: nowString(),
	}
	return link, nil
}

func buildPublicDeliverySignDetail(data AppData, token string) (publicDeliverySignDetail, error) {
	_, link, err := findUsableDeliverySignLink(data, token)
	if err != nil {
		return publicDeliverySignDetail{}, err
	}
	dispatch, _ := findDispatch(data, link.DispatchID)
	order, _ := findOrder(data, link.OrderID)
	customer, _ := findCustomer(data, link.CustomerID)
	project, _ := findProject(data, link.ProjectID)
	product, _ := findProduct(data, order.ProductID)
	ticket, _ := findScaleTicket(data, link.TicketID)
	vehicle, _ := findVehicle(data, dispatch.VehicleID)
	productDisplay := product.Name
	if link.ProductName != "" {
		productDisplay = link.ProductName
	}
	return publicDeliverySignDetail{
		Link: link, Dispatch: dispatch, Ticket: ticket, Order: order,
		Customer: customer.Name, Project: project.Name, Product: productDisplay, PlateNo: vehicle.PlateNo,
		Attachments: deliverySignAttachmentsForDispatch(data, dispatch.ID),
	}, nil
}

func findUsableDeliverySignLink(data AppData, token string) (int, DeliverySignLink, error) {
	for i, link := range data.DeliverySignLinks {
		if link.Token != token {
			continue
		}
		if link.Status == "used" || link.Status == "revoked" || link.Status == "expired" {
			return i, link, fmt.Errorf("签收链接已失效")
		}
		if link.ExpiresAt != "" {
			if expiresAt, err := time.ParseInLocation("2006-01-02 15:04:05", link.ExpiresAt, time.Local); err == nil && time.Now().After(expiresAt) {
				return i, link, fmt.Errorf("签收链接已过期")
			}
		}
		return i, link, nil
	}
	return -1, DeliverySignLink{}, fmt.Errorf("签收链接不存在")
}

func createDeliverySignAttachment(data *AppData, item DeliverySignAttachment, actor string) DeliverySignAttachment {
	item.ID = nextID(data, "signAttachment")
	item.FileType = fallback(item.FileType, "photo")
	item.UploadedBy = firstNonEmpty(item.UploadedBy, actor)
	item.UploadedAt = firstNonEmpty(item.UploadedAt, nowString())
	data.DeliverySignAttachments = append(data.DeliverySignAttachments, item)
	return item
}

func upsertDeliveryNoteQRCode(data *AppData, link DeliverySignLink) {
	for i := range data.DeliveryNotes {
		if data.DeliveryNotes[i].DispatchID == link.DispatchID {
			data.DeliveryNotes[i].QRCode = link.QRCode
			data.DeliveryNotes[i].Status = "pending"
			return
		}
	}
	id := nextID(data, "deliveryNote")
	data.DeliveryNotes = append(data.DeliveryNotes, DeliveryNote{
		ID: id, NoteNo: number("DN", id), TicketID: link.TicketID, OrderID: link.OrderID,
		DispatchID: link.DispatchID, QRCode: link.QRCode, Status: "pending", CreatedAt: nowString(),
	})
}

func deliverySignPublicURL(token string) string {
	base := strings.TrimRight(os.Getenv("CBMP_PUBLIC_BASE_URL"), "/")
	path := "/public/sign/" + token
	if base == "" {
		return path
	}
	return base + path
}

func findDeliverySign(data AppData, id int64) (DeliverySign, bool) {
	for _, item := range data.DeliverySigns {
		if item.ID == id {
			return item, true
		}
	}
	return DeliverySign{}, false
}

func findDeliverySignByDispatch(data AppData, dispatchID int64) (DeliverySign, bool) {
	for _, item := range data.DeliverySigns {
		if item.DispatchID == dispatchID {
			return item, true
		}
	}
	return DeliverySign{}, false
}

func deliverySignAttachmentsForDispatch(data AppData, dispatchID int64) []DeliverySignAttachment {
	items := []DeliverySignAttachment{}
	for _, item := range data.DeliverySignAttachments {
		if item.DispatchID == dispatchID {
			items = append(items, item)
		}
	}
	return items
}

func parseSignID(parts []string) (int64, bool) {
	if len(parts) != 3 || parts[0] != "sign" || parts[2] != "attachments" {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	return id, err == nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
