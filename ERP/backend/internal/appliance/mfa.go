package appliance

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const totpStepSeconds int64 = 30

func generateTOTPSecret() string {
	raw := make([]byte, 20)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		copy(raw, []byte(time.Now().Format(time.RFC3339Nano)))
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(raw), "=")
}

func (a *App) systemMFA(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "users" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, publicUsers(a.mustSnapshot().Users))
		return
	}
	if len(parts) != 3 || parts[0] != "users" || r.Method != http.MethodPost {
		writeError(w, http.StatusNotFound, "unknown mfa route")
		return
	}
	userID, _ := strconv.ParseInt(parts[1], 10, 64)
	switch parts[2] {
	case "enroll":
		a.enrollMFA(w, r, session, userID)
	case "enable":
		a.enableMFA(w, r, session, userID)
	case "disable":
		a.disableMFA(w, r, session, userID)
	default:
		writeError(w, http.StatusNotFound, "unknown mfa action")
	}
}

func (a *App) enrollMFA(w http.ResponseWriter, r *http.Request, session Session, userID int64) {
	secret := generateTOTPSecret()
	var user User
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.Users {
			if data.Users[i].ID == userID {
				data.Users[i].MFASecret = secret
				data.Users[i].MFAEnabled = false
				data.Users[i].MFALastUsedStep = 0
				user = data.Users[i]
				addAudit(data, session.User.Username, "enroll", "mfa", userID, data.Users[i].Username, clientIP(r))
				return nil
			}
		}
		return fmt.Errorf("用户不存在")
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user":       publicUser(user),
		"secret":     secret,
		"otpauthUrl": totpURL("CBMP", user.Username, secret),
	})
}

func (a *App) enableMFA(w http.ResponseWriter, r *http.Request, session Session, userID int64) {
	var req struct {
		Code string `json:"code"`
	}
	_ = readJSON(r, &req)
	var user User
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.Users {
			if data.Users[i].ID != userID {
				continue
			}
			if data.Users[i].MFASecret == "" {
				return fmt.Errorf("请先生成 MFA 密钥")
			}
			if _, ok := verifyTOTP(data.Users[i].MFASecret, req.Code, time.Now(), -1); !ok {
				return fmt.Errorf("MFA 动态码无效")
			}
			data.Users[i].MFAEnabled = true
			user = data.Users[i]
			addAudit(data, session.User.Username, "enable", "mfa", userID, data.Users[i].Username, clientIP(r))
			return nil
		}
		return fmt.Errorf("用户不存在")
	})
	a.respondMutation(w, err, publicUser(user), "system.mfa.enabled")
}

func (a *App) disableMFA(w http.ResponseWriter, r *http.Request, session Session, userID int64) {
	var user User
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.Users {
			if data.Users[i].ID == userID {
				data.Users[i].MFAEnabled = false
				data.Users[i].MFASecret = ""
				data.Users[i].MFALastUsedStep = 0
				user = data.Users[i]
				addAudit(data, session.User.Username, "disable", "mfa", userID, data.Users[i].Username, clientIP(r))
				return nil
			}
		}
		return fmt.Errorf("用户不存在")
	})
	a.respondMutation(w, err, publicUser(user), "system.mfa.disabled")
}

func totpURL(issuer, username, secret string) string {
	label := issuer + ":" + username
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", issuer)
	values.Set("algorithm", "SHA1")
	values.Set("digits", "6")
	values.Set("period", strconv.FormatInt(totpStepSeconds, 10))
	return "otpauth://totp/" + url.PathEscape(label) + "?" + values.Encode()
}

func verifyTOTP(secret, code string, now time.Time, lastUsedStep int64) (int64, bool) {
	code = strings.TrimSpace(code)
	if secret == "" || code == "" {
		return 0, false
	}
	step := now.Unix() / totpStepSeconds
	for offset := int64(-1); offset <= 1; offset++ {
		candidateStep := step + offset
		if candidateStep <= lastUsedStep {
			continue
		}
		if totpCodeAt(secret, candidateStep) == code {
			return candidateStep, true
		}
	}
	return 0, false
}

func totpCodeAt(secret string, step int64) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return ""
	}
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(step))
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(msg)
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)
	return fmt.Sprintf("%06d", value%1000000)
}
