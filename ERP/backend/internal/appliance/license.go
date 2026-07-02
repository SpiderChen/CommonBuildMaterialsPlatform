package appliance

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type LicenseVerification struct {
	Valid       bool        `json:"valid"`
	Reason      string      `json:"reason"`
	License     LicenseInfo `json:"license"`
	ModuleCount int         `json:"moduleCount"`
	MaxSites    int         `json:"maxSites"`
	MaxVehicles int         `json:"maxVehicles"`
	Signature   string      `json:"signatureType"`
	Fingerprint string      `json:"fingerprint"`
	VerifiedAt  string      `json:"verifiedAt"`
}

type licenseSignedPayload struct {
	LicenseID    string   `json:"licenseId"`
	CustomerName string   `json:"customerName"`
	Watermark    string   `json:"watermark"`
	ExpiresAt    string   `json:"expiresAt"`
	Edition      string   `json:"edition"`
	Modules      []string `json:"modules"`
	MaxSites     int      `json:"maxSites"`
	MaxVehicles  int      `json:"maxVehicles"`
	IssuedAt     string   `json:"issuedAt"`
	Issuer       string   `json:"issuer"`
}

func (a *App) systemLicense(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	data := a.mustSnapshot()
	if len(parts) == 0 {
		writeJSON(w, http.StatusOK, data.License)
		return
	}
	switch parts[0] {
	case "verify":
		writeJSON(w, http.StatusOK, verifyActiveLicense(data))
	case "portal":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, buildLicensePortalOverview(data))
	case "packages":
		if len(parts) == 3 && parts[2] == "download" && r.Method == http.MethodGet {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			a.downloadLicensePackage(w, r, session, id)
			return
		}
		if len(parts) == 3 && parts[2] == "renew" && r.Method == http.MethodPost {
			id, _ := strconv.ParseInt(parts[1], 10, 64)
			a.renewLicensePackage(w, r, session, id)
			return
		}
		if len(parts) != 1 || r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, data.LicensePackages)
	case "issues":
		a.systemLicenseIssues(w, r, session)
	case "revocations":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, data.LicenseRevocations)
	case "revoke":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		a.revokeLicense(w, r, session)
	case "import":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		a.importLicensePackage(w, r, session)
	default:
		writeError(w, http.StatusNotFound, "unknown license route")
	}
}

func (a *App) importLicensePackage(w http.ResponseWriter, r *http.Request, session Session) {
	var item LicensePackage
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid license package")
		return
	}
	var activated LicensePackage
	err := a.store.Mutate(func(data *AppData) error {
		valid, reason := verifyLicensePackage(item, *data)
		if !valid {
			return fmt.Errorf("%s", reason)
		}
		item.ID = nextID(data, "licensePackage")
		item.PublicKeyFingerprint = licensePublicKeyFingerprint(item.PublicKey)
		item.Status = "active"
		item.ActivatedAt = nowString()
		item.LastVerificationState = "valid"
		item.LastVerificationError = ""
		for i := range data.LicensePackages {
			if data.LicensePackages[i].Status == "active" {
				data.LicensePackages[i].Status = "superseded"
			}
		}
		data.LicensePackages = append(data.LicensePackages, item)
		data.License = licenseInfoFromPackage(item)
		data.License.LastVerifiedAt = nowString()
		data.License.LastVerificationState = "valid"
		data.License.LastVerificationError = ""
		activated = item
		addAudit(data, session.User.Username, "import", "license", item.ID, item.LicenseID, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, activated, "system.license.imported")
}

func verifyActiveLicense(data AppData) LicenseVerification {
	pkg := LicensePackage{
		LicenseID: data.License.LicenseID, CustomerName: data.License.CustomerName, Watermark: data.License.Watermark,
		ExpiresAt: data.License.ExpiresAt, Edition: data.License.Edition, Modules: data.License.Modules,
		MaxSites: data.License.MaxSites, MaxVehicles: data.License.MaxVehicles, IssuedAt: data.License.IssuedAt,
		Issuer: data.License.Issuer, PublicKey: data.License.PublicKey, Signature: data.License.Signature,
	}
	valid, reason := verifyLicensePackage(pkg, data)
	return LicenseVerification{
		Valid: valid, Reason: reason, License: data.License,
		ModuleCount: len(data.License.Modules), MaxSites: data.License.MaxSites, MaxVehicles: data.License.MaxVehicles,
		Signature: "ed25519", Fingerprint: licensePublicKeyFingerprint(data.License.PublicKey), VerifiedAt: nowString(),
	}
}

func verifyLicensePackage(item LicensePackage, data AppData) (bool, string) {
	if item.LicenseID == "" || item.CustomerName == "" || item.Watermark == "" || item.ExpiresAt == "" {
		return false, "授权包缺少必要字段"
	}
	if revoked, reason := licenseRevoked(data, item.LicenseID); revoked {
		return false, "授权已吊销: " + reason
	}
	if len(item.Modules) == 0 {
		return false, "授权模块不能为空"
	}
	if item.MaxSites < len(data.Sites) {
		return false, "授权站点额度不足"
	}
	if item.MaxVehicles < len(data.Vehicles) {
		return false, "授权车辆额度不足"
	}
	expiresAt, err := time.Parse("2006-01-02", item.ExpiresAt)
	if err != nil {
		return false, "授权到期日格式错误"
	}
	if expiresAt.Before(time.Now().Truncate(24 * time.Hour)) {
		return false, "授权已过期"
	}
	publicKey, err := decodeLicensePublicKey(item.PublicKey)
	if err != nil {
		return false, err.Error()
	}
	if item.PublicKeyFingerprint != "" && item.PublicKeyFingerprint != licensePublicKeyFingerprint(item.PublicKey) {
		return false, "授权公钥指纹不匹配"
	}
	if trusted, reason := licensePublicKeyTrusted(publicKey); !trusted {
		return false, reason
	}
	signature, err := decodeLicenseSignature(item.Signature)
	if err != nil {
		return false, err.Error()
	}
	payload, err := licenseCanonicalPayload(item)
	if err != nil {
		return false, "授权内容无法序列化"
	}
	if !ed25519.Verify(publicKey, payload, signature) {
		return false, "授权签名无效"
	}
	return true, "valid"
}

func licensePublicKeyTrusted(publicKey ed25519.PublicKey) (bool, string) {
	trustedKeys, err := trustedLicensePublicKeys()
	if err != nil {
		return false, err.Error()
	}
	if len(trustedKeys) == 0 {
		return false, "授权签发公钥未配置"
	}
	for _, trustedKey := range trustedKeys {
		if bytes.Equal(trustedKey, publicKey) {
			return true, "trusted"
		}
	}
	return false, "授权签发公钥不受信任"
}

func trustedLicensePublicKeys() ([]ed25519.PublicKey, error) {
	values := []string{}
	values = append(values, splitLicenseKeyList(os.Getenv("CBMP_LICENSE_TRUSTED_PUBLIC_KEYS"))...)
	if value := strings.TrimSpace(os.Getenv("CBMP_LICENSE_ISSUER_PUBLIC_KEY")); value != "" {
		values = append(values, value)
	}
	keys := make([]ed25519.PublicKey, 0, len(values)+1)
	seen := map[string]bool{}
	for _, value := range values {
		key, err := decodeLicensePublicKey(value)
		if err != nil {
			return nil, fmt.Errorf("受信授权公钥配置无效")
		}
		fingerprint := licensePublicKeyFingerprint(value)
		if fingerprint == "" || seen[fingerprint] {
			continue
		}
		seen[fingerprint] = true
		keys = append(keys, key)
	}
	if value := strings.TrimSpace(os.Getenv("CBMP_LICENSE_ISSUER_PRIVATE_KEY")); value != "" {
		privateKey, err := decodeLicensePrivateKey(value)
		if err != nil {
			return nil, err
		}
		publicValue := "ed25519:" + base64.RawStdEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey))
		fingerprint := licensePublicKeyFingerprint(publicValue)
		if fingerprint != "" && !seen[fingerprint] {
			keys = append(keys, privateKey.Public().(ed25519.PublicKey))
		}
	}
	return keys, nil
}

func splitLicenseKeyList(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	out := []string{}
	for _, field := range fields {
		if field = strings.TrimSpace(field); field != "" {
			out = append(out, field)
		}
	}
	return out
}

func licenseRevoked(data AppData, licenseID string) (bool, string) {
	for _, item := range data.LicenseRevocations {
		if item.LicenseID == licenseID && item.Status == "active" {
			return true, fallback(item.Reason, "revoked")
		}
	}
	return false, ""
}

func licenseInfoFromPackage(item LicensePackage) LicenseInfo {
	return LicenseInfo{
		LicenseID: item.LicenseID, CustomerName: item.CustomerName, Watermark: item.Watermark,
		ExpiresAt: item.ExpiresAt, Edition: item.Edition, Modules: item.Modules, MaxSites: item.MaxSites,
		MaxVehicles: item.MaxVehicles, IssuedAt: item.IssuedAt, Issuer: item.Issuer, PublicKey: item.PublicKey,
		PublicKeyFingerprint: licensePublicKeyFingerprint(item.PublicKey), Signature: item.Signature,
	}
}

func licenseCanonicalPayload(item LicensePackage) ([]byte, error) {
	return json.Marshal(licenseSignedPayload{
		LicenseID: item.LicenseID, CustomerName: item.CustomerName, Watermark: item.Watermark,
		ExpiresAt: item.ExpiresAt, Edition: item.Edition, Modules: item.Modules, MaxSites: item.MaxSites,
		MaxVehicles: item.MaxVehicles, IssuedAt: item.IssuedAt, Issuer: item.Issuer,
	})
}

func decodeLicensePublicKey(value string) (ed25519.PublicKey, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(value), "ed25519:")
	decoded, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("授权公钥无效")
	}
	return ed25519.PublicKey(decoded), nil
}

func decodeLicenseSignature(value string) ([]byte, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(value), "ed25519:")
	decoded, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil || len(decoded) != ed25519.SignatureSize {
		return nil, fmt.Errorf("授权签名无效")
	}
	return decoded, nil
}

func licensePublicKeyFingerprint(value string) string {
	key, err := decodeLicensePublicKey(value)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:])[:16]
}
