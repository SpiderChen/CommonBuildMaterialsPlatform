package appliance

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type OIDCLoginState struct {
	State        string
	Nonce        string
	ProviderCode string
	RedirectURI  string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type oidcStartResponse struct {
	Provider         OIDCProvider `json:"provider"`
	AuthorizationURL string       `json:"authorizationUrl"`
	State            string       `json:"state"`
	Nonce            string       `json:"nonce"`
	ExpiresAt        string       `json:"expiresAt"`
}

type oidcCallbackRequest struct {
	Code    string                 `json:"code"`
	State   string                 `json:"state"`
	IDToken string                 `json:"idToken"`
	Claims  map[string]interface{} `json:"claims"`
}

func (a *App) authSSO(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 && parts[0] == "providers" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, activePublicOIDCProviders(a.mustSnapshot().OIDCProviders))
		return
	}
	if len(parts) == 2 && parts[1] == "start" && r.Method == http.MethodPost {
		a.startOIDCLogin(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[1] == "callback" && (r.Method == http.MethodPost || r.Method == http.MethodGet) {
		a.finishOIDCLogin(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "unknown sso route")
}

func (a *App) startOIDCLogin(w http.ResponseWriter, r *http.Request, providerCode string) {
	provider, ok := findOIDCProvider(a.mustSnapshot().OIDCProviders, providerCode)
	if !ok || provider.Status != "enabled" {
		writeError(w, http.StatusNotFound, "SSO 提供商不可用")
		return
	}
	redirectURI := fallback(provider.RedirectURI, callbackRedirectURI(r, provider.Code))
	state := tokenString()
	nonce := tokenString()
	expiresAt := time.Now().Add(10 * time.Minute)
	a.mu.Lock()
	a.ssoStates[state] = OIDCLoginState{
		State: state, Nonce: nonce, ProviderCode: provider.Code, RedirectURI: redirectURI,
		CreatedAt: time.Now(), ExpiresAt: expiresAt,
	}
	a.mu.Unlock()
	authURL, err := oidcAuthorizationURL(provider, redirectURI, state, nonce)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, oidcStartResponse{
		Provider: publicOIDCProvider(provider), AuthorizationURL: authURL,
		State: state, Nonce: nonce, ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (a *App) finishOIDCLogin(w http.ResponseWriter, r *http.Request, providerCode string) {
	req, err := readOIDCCallbackRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	loginState, ok := a.consumeOIDCState(req.State, providerCode)
	if !ok {
		writeError(w, http.StatusBadRequest, "SSO state 无效或已过期")
		return
	}
	provider, ok := findOIDCProvider(a.mustSnapshot().OIDCProviders, providerCode)
	if !ok || provider.Status != "enabled" {
		writeError(w, http.StatusNotFound, "SSO 提供商不可用")
		return
	}
	claims, err := oidcClaims(provider, loginState, req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	user, err := a.upsertOIDCUser(provider, claims, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	session := a.issueSessionWithDetail(w, r, user, "用户 SSO 登录: "+provider.Code)
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/?sso=ok", http.StatusFound)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func readOIDCCallbackRequest(r *http.Request) (oidcCallbackRequest, error) {
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		return oidcCallbackRequest{Code: query.Get("code"), State: query.Get("state"), IDToken: query.Get("id_token")}, nil
	}
	var req oidcCallbackRequest
	if err := readJSON(r, &req); err != nil {
		return req, fmt.Errorf("invalid sso callback payload")
	}
	return req, nil
}

func (a *App) consumeOIDCState(state, providerCode string) (OIDCLoginState, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	loginState, ok := a.ssoStates[state]
	if !ok || loginState.ProviderCode != providerCode || time.Now().After(loginState.ExpiresAt) {
		delete(a.ssoStates, state)
		return OIDCLoginState{}, false
	}
	delete(a.ssoStates, state)
	return loginState, true
}

func oidcClaims(provider OIDCProvider, loginState OIDCLoginState, req oidcCallbackRequest) (map[string]interface{}, error) {
	if req.IDToken != "" {
		return verifyOIDCIDToken(provider, req.IDToken, loginState.Nonce)
	}
	if req.Code == "" {
		return nil, fmt.Errorf("SSO 回调缺少 code 或 id_token")
	}
	return exchangeOIDCCode(provider, loginState, req.Code)
}

func oidcAuthorizationURL(provider OIDCProvider, redirectURI, state, nonce string) (string, error) {
	if strings.TrimSpace(provider.AuthURL) == "" {
		return "", fmt.Errorf("SSO 授权地址未配置")
	}
	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", provider.ClientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("scope", strings.Join(provider.Scopes, " "))
	values.Set("state", state)
	values.Set("nonce", nonce)
	parsed, err := url.Parse(provider.AuthURL)
	if err != nil {
		return "", fmt.Errorf("SSO 授权地址无效")
	}
	query := parsed.Query()
	for key, item := range values {
		for _, value := range item {
			query.Add(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func exchangeOIDCCode(provider OIDCProvider, loginState OIDCLoginState, code string) (map[string]interface{}, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", provider.ClientID)
	form.Set("client_secret", provider.ClientSecret)
	form.Set("redirect_uri", loginState.RedirectURI)
	client := http.Client{Timeout: 8 * time.Second}
	resp, err := client.PostForm(provider.TokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("SSO token 交换失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("SSO token 交换失败: %s", strings.TrimSpace(string(body)))
	}
	var token struct {
		IDToken     string `json:"id_token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("SSO token 响应无效")
	}
	if token.IDToken != "" {
		return verifyOIDCIDToken(provider, token.IDToken, loginState.Nonce)
	}
	if token.AccessToken != "" && provider.UserInfoURL != "" {
		return fetchOIDCUserInfo(provider, token.AccessToken)
	}
	return nil, fmt.Errorf("SSO token 响应缺少 id_token")
}

func fetchOIDCUserInfo(provider OIDCProvider, accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, provider.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("SSO userinfo 地址无效")
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SSO userinfo 获取失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("SSO userinfo 获取失败: %s", strings.TrimSpace(string(body)))
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(body, &claims); err != nil {
		return nil, fmt.Errorf("SSO userinfo 响应无效")
	}
	if iss := stringClaim(claims, "iss"); iss != "" && provider.Issuer != "" && iss != provider.Issuer {
		return nil, fmt.Errorf("SSO issuer 不匹配")
	}
	return claims, nil
}

func verifyOIDCIDToken(provider OIDCProvider, token, nonce string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("id_token 格式无效")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("id_token header 无效")
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("id_token header 无效")
	}
	signingInput := parts[0] + "." + parts[1]
	switch header.Alg {
	case "HS256":
		mac := hmac.New(sha256.New, []byte(provider.ClientSecret))
		mac.Write([]byte(signingInput))
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(expected), []byte(parts[2])) {
			return nil, fmt.Errorf("id_token 签名无效")
		}
	case "RS256":
		if err := verifyOIDCRS256(provider, header.Kid, signingInput, parts[2]); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的 id_token alg: %s", header.Alg)
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("id_token payload 无效")
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("id_token payload 无效")
	}
	if err := validateOIDCClaimSet(provider, claims, nonce, true); err != nil {
		return nil, err
	}
	return claims, nil
}

func verifyOIDCRS256(provider OIDCProvider, kid, signingInput, signaturePart string) error {
	if provider.JWKSURL == "" {
		return fmt.Errorf("RS256 id_token 需要配置 jwksUrl")
	}
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Get(provider.JWKSURL)
	if err != nil {
		return fmt.Errorf("JWKS 获取失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("JWKS 获取失败: %s", strings.TrimSpace(string(body)))
	}
	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Alg string `json:"alg"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("JWKS 响应无效")
	}
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || (kid != "" && key.Kid != kid) {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			continue
		}
		pub := &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: int(new(big.Int).SetBytes(eBytes).Int64())}
		signature, err := base64.RawURLEncoding.DecodeString(signaturePart)
		if err != nil {
			return fmt.Errorf("id_token 签名无效")
		}
		digest := sha256.Sum256([]byte(signingInput))
		if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], signature); err != nil {
			return fmt.Errorf("id_token 签名无效")
		}
		return nil
	}
	return fmt.Errorf("JWKS 未找到匹配 RSA key")
}

func validateOIDCClaimSet(provider OIDCProvider, claims map[string]interface{}, nonce string, requireNonce bool) error {
	if provider.Issuer != "" && stringClaim(claims, "iss") != provider.Issuer {
		return fmt.Errorf("SSO issuer 不匹配")
	}
	if provider.ClientID != "" && !audienceContains(claims["aud"], provider.ClientID) {
		return fmt.Errorf("SSO audience 不匹配")
	}
	if exp := numericClaim(claims, "exp"); exp > 0 && int64(exp) < time.Now().Unix() {
		return fmt.Errorf("id_token 已过期")
	}
	if nbf := numericClaim(claims, "nbf"); nbf > 0 && int64(nbf) > time.Now().Add(time.Minute).Unix() {
		return fmt.Errorf("id_token 尚未生效")
	}
	if requireNonce && nonce != "" && stringClaim(claims, "nonce") != nonce {
		return fmt.Errorf("SSO nonce 不匹配")
	}
	return nil
}

func (a *App) upsertOIDCUser(provider OIDCProvider, claims map[string]interface{}, r *http.Request) (User, error) {
	username := fallback(stringClaim(claims, provider.UsernameClaim), fallback(stringClaim(claims, "preferred_username"), fallback(stringClaim(claims, "email"), stringClaim(claims, "sub"))))
	if strings.TrimSpace(username) == "" {
		return User{}, fmt.Errorf("SSO claims 缺少用户名")
	}
	displayName := fallback(stringClaim(claims, provider.DisplayNameClaim), username)
	var user User
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.Users {
			if data.Users[i].Username != username {
				continue
			}
			if data.Users[i].Status != "active" {
				return fmt.Errorf("SSO 用户已停用")
			}
			user = data.Users[i]
			updateOIDCProviderLogin(data, provider.ID)
			addAudit(data, username, "sso", "oidc_provider", provider.ID, "SSO 映射已有用户", clientIP(r))
			return nil
		}
		if !provider.AutoProvision {
			return fmt.Errorf("SSO 用户未绑定且未开启自动开通")
		}
		user = User{
			ID: nextID(data, "user"), CompanyID: provider.CompanyID, SiteID: provider.SiteID,
			Username: username, DisplayName: displayName, RoleCode: fallback(provider.RoleCode, "customer"), Status: "active",
		}
		data.Users = append(data.Users, user)
		updateOIDCProviderLogin(data, provider.ID)
		addAudit(data, username, "provision", "sys_user", user.ID, "SSO 自动开通: "+provider.Code, clientIP(r))
		return nil
	})
	return user, err
}

func updateOIDCProviderLogin(data *AppData, providerID int64) {
	for i := range data.OIDCProviders {
		if data.OIDCProviders[i].ID == providerID {
			data.OIDCProviders[i].LastLoginAt = nowString()
			return
		}
	}
}

func (a *App) systemSSO(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "providers" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, publicOIDCProviders(a.mustSnapshot().OIDCProviders))
		return
	}
	if len(parts) == 1 && parts[0] == "providers" && r.Method == http.MethodPost {
		a.upsertOIDCProvider(w, r, session)
		return
	}
	if len(parts) == 3 && parts[0] == "providers" && parts[2] == "status" && r.Method == http.MethodPost {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req struct {
			Status string `json:"status"`
		}
		_ = readJSON(r, &req)
		var updated OIDCProvider
		topic := "system.sso.status"
		err := a.store.Mutate(func(data *AppData) error {
			for i := range data.OIDCProviders {
				if data.OIDCProviders[i].ID == id {
					status := fallback(strings.TrimSpace(req.Status), "enabled")
					if hasPendingWorkflowForResource(*data, "oidc_provider", id) {
						updated = publicOIDCProvider(data.OIDCProviders[i])
						topic = "system.sso.status_requested"
						return nil
					}
					_, instances, err := publishOIDCProviderStatusWorkflow(data, data.OIDCProviders[i], status, session.User.Username)
					if err != nil {
						return err
					}
					if len(instances) > 0 {
						updated = publicOIDCProvider(data.OIDCProviders[i])
						topic = "system.sso.status_requested"
						addAudit(data, session.User.Username, "request_status", "oidc_provider", id, data.OIDCProviders[i].Code+"/"+status, clientIP(r))
						return nil
					}
					next, err := applyOIDCProviderStatusLocked(data, id, status)
					if err != nil {
						return err
					}
					updated = publicOIDCProvider(next)
					addAudit(data, session.User.Username, "status", "oidc_provider", id, updated.Code+"/"+updated.Status, clientIP(r))
					return nil
				}
			}
			return fmt.Errorf("SSO 提供商不存在")
		})
		a.respondMutation(w, err, updated, topic)
		return
	}
	if len(parts) == 2 && parts[0] == "providers" && r.Method == http.MethodDelete {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var deleted OIDCProvider
		err := a.store.Mutate(func(data *AppData) error {
			for i, item := range data.OIDCProviders {
				if item.ID != id {
					continue
				}
				if item.Status == "enabled" {
					return fmt.Errorf("启用中的 SSO 提供商不能删除，请先停用")
				}
				if hasPendingWorkflowForResource(*data, "oidc_provider", id) {
					return fmt.Errorf("SSO 提供商存在待审批状态变更，不能删除")
				}
				deleted = publicOIDCProvider(item)
				data.OIDCProviders = append(data.OIDCProviders[:i], data.OIDCProviders[i+1:]...)
				addAudit(data, session.User.Username, "delete", "oidc_provider", id, item.Code, clientIP(r))
				return nil
			}
			return fmt.Errorf("SSO 提供商不存在")
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.emit("system.sso.provider.deleted", deleted)
		writeJSON(w, http.StatusOK, deleted)
		return
	}
	writeError(w, http.StatusNotFound, "unknown sso system route")
}

func publishOIDCProviderStatusWorkflow(data *AppData, item OIDCProvider, targetStatus string, actor string) (WorkflowEvent, []WorkflowInstance, error) {
	return publishWorkflowEvent(data, workflowEventRequest{
		EventType:  "oidc_provider.status_change_requested",
		Source:     "system",
		Resource:   "oidc_provider",
		ResourceID: item.ID,
		ResourceNo: item.Code,
		Title:      "SSO 提供商状态变更 " + item.Code,
		Actor:      actor,
		Reason:     "SSO 提供商状态变更审批",
		Variables: map[string]string{
			"targetStatus":  targetStatus,
			"currentStatus": item.Status,
			"code":          item.Code,
			"name":          item.Name,
			"issuer":        item.Issuer,
			"companyId":     fmt.Sprintf("%d", item.CompanyID),
			"siteId":        fmt.Sprintf("%d", item.SiteID),
		},
	})
}

func applyOIDCProviderStatusLocked(data *AppData, id int64, status string) (OIDCProvider, error) {
	status = fallback(strings.TrimSpace(status), "enabled")
	for i := range data.OIDCProviders {
		if data.OIDCProviders[i].ID == id {
			data.OIDCProviders[i].Status = status
			return data.OIDCProviders[i], nil
		}
	}
	return OIDCProvider{}, fmt.Errorf("SSO 提供商不存在")
}

func (a *App) upsertOIDCProvider(w http.ResponseWriter, r *http.Request, session Session) {
	var req OIDCProvider
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid sso provider payload")
		return
	}
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "SSO provider name/code required")
		return
	}
	var updated OIDCProvider
	err := a.store.Mutate(func(data *AppData) error {
		req.Code = strings.TrimSpace(req.Code)
		req.Status = fallback(req.Status, "enabled")
		req.UsernameClaim = fallback(req.UsernameClaim, "preferred_username")
		req.DisplayNameClaim = fallback(req.DisplayNameClaim, "name")
		req.RoleCode = fallback(req.RoleCode, "customer")
		if len(req.Scopes) == 0 {
			req.Scopes = []string{"openid", "profile", "email"}
		}
		for i := range data.OIDCProviders {
			if data.OIDCProviders[i].ID == req.ID || data.OIDCProviders[i].Code == req.Code {
				if req.ClientSecret == "" {
					req.ClientSecret = data.OIDCProviders[i].ClientSecret
				}
				req.ID = data.OIDCProviders[i].ID
				req.LastLoginAt = data.OIDCProviders[i].LastLoginAt
				data.OIDCProviders[i] = req
				updated = publicOIDCProvider(req)
				addAudit(data, session.User.Username, "update", "oidc_provider", req.ID, req.Code, clientIP(r))
				return nil
			}
		}
		req.ID = nextID(data, "oidcProvider")
		data.OIDCProviders = append(data.OIDCProviders, req)
		updated = publicOIDCProvider(req)
		addAudit(data, session.User.Username, "create", "oidc_provider", req.ID, req.Code, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, updated, "system.sso.provider")
}

func findOIDCProvider(items []OIDCProvider, code string) (OIDCProvider, bool) {
	for _, item := range items {
		if item.Code == code {
			return item, true
		}
	}
	return OIDCProvider{}, false
}

func publicOIDCProviders(items []OIDCProvider) []OIDCProvider {
	out := make([]OIDCProvider, 0, len(items))
	for _, item := range items {
		out = append(out, publicOIDCProvider(item))
	}
	return out
}

func activePublicOIDCProviders(items []OIDCProvider) []OIDCProvider {
	out := []OIDCProvider{}
	for _, item := range items {
		if item.Status == "enabled" {
			out = append(out, publicOIDCProvider(item))
		}
	}
	return out
}

func publicOIDCProvider(item OIDCProvider) OIDCProvider {
	item.TenantID = 0
	item.ClientSecret = ""
	return item
}

func callbackRedirectURI(r *http.Request, providerCode string) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
		if r.TLS != nil {
			scheme = "https"
		}
	}
	return fmt.Sprintf("%s://%s/api/auth/sso/%s/callback", scheme, r.Host, providerCode)
}

func stringClaim(claims map[string]interface{}, key string) string {
	if key == "" {
		return ""
	}
	value, ok := claims[key]
	if !ok {
		return ""
	}
	typed, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(typed)
}

func numericClaim(claims map[string]interface{}, key string) float64 {
	value, ok := claims[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case json.Number:
		v, _ := typed.Float64()
		return v
	default:
		return 0
	}
}

func audienceContains(value interface{}, expected string) bool {
	switch typed := value.(type) {
	case string:
		return typed == expected
	case []interface{}:
		for _, item := range typed {
			if text, ok := item.(string); ok && text == expected {
				return true
			}
		}
	}
	return false
}
