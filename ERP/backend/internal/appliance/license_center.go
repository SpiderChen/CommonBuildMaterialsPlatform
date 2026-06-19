package appliance

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type licenseIssueRequest struct {
	LicenseID    string   `json:"licenseId"`
	CustomerName string   `json:"customerName"`
	Watermark    string   `json:"watermark"`
	ExpiresAt    string   `json:"expiresAt"`
	Edition      string   `json:"edition"`
	Modules      []string `json:"modules"`
	MaxSites     int      `json:"maxSites"`
	MaxVehicles  int      `json:"maxVehicles"`
	Issuer       string   `json:"issuer"`
	PrivateKey   string   `json:"privateKey"`
}

func (a *App) systemLicenseIssues(w http.ResponseWriter, r *http.Request, session Session) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.mustSnapshot().LicenseIssues)
	case http.MethodPost:
		a.issueLicense(w, r, session)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) issueLicense(w http.ResponseWriter, r *http.Request, session Session) {
	var req licenseIssueRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid license issue payload")
		return
	}
	privateKey, err := decodeLicensePrivateKey(fallback(req.PrivateKey, os.Getenv("CBMP_LICENSE_ISSUER_PRIVATE_KEY")))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var issued LicensePackage
	err = a.store.Mutate(func(data *AppData) error {
		id := nextID(data, "licenseIssue")
		issued = LicensePackage{
			ID:           nextID(data, "licensePackage"),
			LicenseID:    fallback(req.LicenseID, number("LIC", id)),
			CustomerName: strings.TrimSpace(req.CustomerName),
			Watermark:    strings.TrimSpace(req.Watermark),
			ExpiresAt:    req.ExpiresAt,
			Edition:      fallback(req.Edition, "Enterprise"),
			Modules:      req.Modules,
			MaxSites:     maxInt(req.MaxSites, len(data.Sites)),
			MaxVehicles:  maxInt(req.MaxVehicles, len(data.Vehicles)),
			IssuedAt:     todayString(),
			Issuer:       fallback(req.Issuer, "CBMP License Center"),
			PublicKey:    "ed25519:" + base64.RawStdEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey)),
			Status:       "issued",
		}
		if len(issued.Modules) == 0 {
			issued.Modules = enabledModuleCodes(data.Modules)
		}
		if err := signLicensePackage(&issued, privateKey); err != nil {
			return err
		}
		if valid, reason := verifyLicensePackage(issued, *data); !valid {
			return fmt.Errorf("%s", reason)
		}
		data.LicensePackages = append(data.LicensePackages, issued)
		data.LicenseIssues = append(data.LicenseIssues, LicenseIssueRecord{
			ID: id, IssueNo: number("LI", id), LicenseID: issued.LicenseID,
			CustomerName: issued.CustomerName, Watermark: issued.Watermark, ExpiresAt: issued.ExpiresAt,
			Edition: issued.Edition, Modules: issued.Modules, MaxSites: issued.MaxSites, MaxVehicles: issued.MaxVehicles,
			Issuer: issued.Issuer, PublicKeyFingerprint: issued.PublicKeyFingerprint, Status: "issued",
			IssuedAt: issued.IssuedAt, Actor: session.User.Username,
		})
		addAudit(data, session.User.Username, "issue", "license", id, issued.LicenseID, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, issued, "system.license.issued")
}

func (a *App) revokeLicense(w http.ResponseWriter, r *http.Request, session Session) {
	var req struct {
		LicenseID string `json:"licenseId"`
		Reason    string `json:"reason"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid license revoke payload")
		return
	}
	req.LicenseID = strings.TrimSpace(req.LicenseID)
	if req.LicenseID == "" {
		writeError(w, http.StatusBadRequest, "licenseId required")
		return
	}
	var revoked LicenseRevocation
	err := a.store.Mutate(func(data *AppData) error {
		for _, item := range data.LicenseRevocations {
			if item.LicenseID == req.LicenseID && item.Status == "active" {
				return fmt.Errorf("授权已在吊销列表中")
			}
		}
		id := nextID(data, "licenseRevocation")
		revoked = LicenseRevocation{
			ID: id, RevokeNo: number("LR", id), LicenseID: req.LicenseID,
			Reason: fallback(req.Reason, "customer revoked"), Status: "active", RevokedAt: nowString(), Actor: session.User.Username,
		}
		data.LicenseRevocations = append(data.LicenseRevocations, revoked)
		for i := range data.LicensePackages {
			if data.LicensePackages[i].LicenseID == req.LicenseID {
				data.LicensePackages[i].Status = "revoked"
				data.LicensePackages[i].LastVerificationState = "revoked"
				data.LicensePackages[i].LastVerificationError = revoked.Reason
			}
		}
		if data.License.LicenseID == req.LicenseID {
			data.License.LastVerificationState = "revoked"
			data.License.LastVerificationError = revoked.Reason
		}
		addAudit(data, session.User.Username, "revoke", "license", id, req.LicenseID, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, revoked, "system.license.revoked")
}

func decodeLicensePrivateKey(value string) (ed25519.PrivateKey, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(value), "ed25519:")
	if raw == "" {
		return nil, fmt.Errorf("授权签发私钥未配置")
	}
	decoded, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("授权签发私钥无效")
	}
	switch len(decoded) {
	case ed25519.PrivateKeySize:
		return ed25519.PrivateKey(decoded), nil
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(decoded), nil
	default:
		return nil, fmt.Errorf("授权签发私钥长度无效")
	}
}

func signLicensePackage(item *LicensePackage, privateKey ed25519.PrivateKey) error {
	item.PublicKey = "ed25519:" + base64.RawStdEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey))
	payload, err := licenseCanonicalPayload(*item)
	if err != nil {
		return err
	}
	item.Signature = "ed25519:" + base64.RawStdEncoding.EncodeToString(ed25519.Sign(privateKey, payload))
	item.PublicKeyFingerprint = licensePublicKeyFingerprint(item.PublicKey)
	return nil
}

func enabledModuleCodes(items []Module) []string {
	out := []string{}
	for _, item := range items {
		if item.Enabled {
			out = append(out, item.Code)
		}
	}
	return out
}

func maxInt(value, fallbackValue int) int {
	if value > fallbackValue {
		return value
	}
	return fallbackValue
}
