package appliance

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type LicensePackageDownload struct {
	FileName     string         `json:"fileName"`
	ContentType  string         `json:"contentType"`
	Package      LicensePackage `json:"package"`
	Valid        bool           `json:"valid"`
	Reason       string         `json:"reason"`
	Fingerprint  string         `json:"fingerprint"`
	DownloadedAt string         `json:"downloadedAt"`
}

type licenseRenewRequest struct {
	LicenseID   string   `json:"licenseId"`
	ExpiresAt   string   `json:"expiresAt"`
	Edition     string   `json:"edition"`
	Modules     []string `json:"modules"`
	MaxSites    int      `json:"maxSites"`
	MaxVehicles int      `json:"maxVehicles"`
	Issuer      string   `json:"issuer"`
	PrivateKey  string   `json:"privateKey"`
}

func (a *App) downloadLicensePackage(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var download LicensePackageDownload
	err := a.store.Mutate(func(data *AppData) error {
		for _, item := range data.LicensePackages {
			if item.ID != id {
				continue
			}
			download = buildLicensePackageDownload(item, *data)
			if !download.Valid {
				return fmt.Errorf("授权包验签失败: %s", download.Reason)
			}
			addAudit(data, session.User.Username, "download", "license_package", id, item.LicenseID, clientIP(r))
			return nil
		}
		return fmt.Errorf("授权包不存在")
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, download)
}

func buildLicensePackageDownload(item LicensePackage, data AppData) LicensePackageDownload {
	valid, reason := verifyLicensePackage(item, data)
	return LicensePackageDownload{
		FileName:     "cbmp-license-" + safeLicenseFilePart(item.LicenseID) + ".json",
		ContentType:  "application/json",
		Package:      item,
		Valid:        valid,
		Reason:       reason,
		Fingerprint:  licensePublicKeyFingerprint(item.PublicKey),
		DownloadedAt: nowString(),
	}
}

func (a *App) renewLicensePackage(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var req licenseRenewRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid license renewal payload")
		return
	}
	privateKey, err := decodeLicensePrivateKey(fallback(req.PrivateKey, os.Getenv("CBMP_LICENSE_ISSUER_PRIVATE_KEY")))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var renewed LicensePackage
	err = a.store.Mutate(func(data *AppData) error {
		source, ok := findLicensePackage(*data, id)
		if !ok {
			return fmt.Errorf("授权包不存在")
		}
		if source.Status == "revoked" {
			return fmt.Errorf("已吊销授权不能续期")
		}
		if valid, reason := verifyLicensePackage(source, *data); !valid {
			return fmt.Errorf("原授权包验签失败: %s", reason)
		}
		issueID := nextID(data, "licenseIssue")
		renewed = LicensePackage{
			ID:           nextID(data, "licensePackage"),
			LicenseID:    fallback(strings.TrimSpace(req.LicenseID), number("LIC", issueID)),
			CustomerName: source.CustomerName,
			Watermark:    source.Watermark,
			ExpiresAt:    renewalExpiresAt(source.ExpiresAt, req.ExpiresAt),
			Edition:      fallback(strings.TrimSpace(req.Edition), source.Edition),
			Modules:      renewalModules(source.Modules, req.Modules),
			MaxSites:     maxInt(req.MaxSites, source.MaxSites),
			MaxVehicles:  maxInt(req.MaxVehicles, source.MaxVehicles),
			IssuedAt:     todayString(),
			Issuer:       fallback(strings.TrimSpace(req.Issuer), source.Issuer),
			Status:       "issued",
		}
		if err := signLicensePackage(&renewed, privateKey); err != nil {
			return err
		}
		if valid, reason := verifyLicensePackage(renewed, *data); !valid {
			return fmt.Errorf("%s", reason)
		}
		data.LicensePackages = append(data.LicensePackages, renewed)
		data.LicenseIssues = append(data.LicenseIssues, LicenseIssueRecord{
			ID: issueID, IssueNo: number("LI", issueID), LicenseID: renewed.LicenseID,
			CustomerName: renewed.CustomerName, Watermark: renewed.Watermark, ExpiresAt: renewed.ExpiresAt,
			Edition: renewed.Edition, Modules: renewed.Modules, MaxSites: renewed.MaxSites, MaxVehicles: renewed.MaxVehicles,
			Issuer: renewed.Issuer, PublicKeyFingerprint: renewed.PublicKeyFingerprint, Status: "renewed",
			IssuedAt: renewed.IssuedAt, Actor: session.User.Username,
		})
		addAudit(data, session.User.Username, "renew", "license_package", renewed.ID, renewed.LicenseID, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, renewed, "system.license.renewed")
}

func findLicensePackage(data AppData, id int64) (LicensePackage, bool) {
	for _, item := range data.LicensePackages {
		if item.ID == id {
			return item, true
		}
	}
	return LicensePackage{}, false
}

func renewalExpiresAt(current, requested string) string {
	requested = strings.TrimSpace(requested)
	if requested != "" {
		return requested
	}
	expiresAt, err := time.Parse("2006-01-02", current)
	if err != nil {
		return time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	}
	return expiresAt.AddDate(1, 0, 0).Format("2006-01-02")
}

func renewalModules(current, requested []string) []string {
	if len(requested) == 0 {
		return current
	}
	modules := make([]string, 0, len(requested))
	for _, item := range requested {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			modules = append(modules, trimmed)
		}
	}
	if len(modules) == 0 {
		return current
	}
	return modules
}

func safeLicenseFilePart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return strconv.FormatInt(time.Now().Unix(), 10)
	}
	var builder strings.Builder
	for _, ch := range value {
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '_' {
			builder.WriteRune(ch)
		}
	}
	if builder.Len() == 0 {
		return strconv.FormatInt(time.Now().Unix(), 10)
	}
	return builder.String()
}
