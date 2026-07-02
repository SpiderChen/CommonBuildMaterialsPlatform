package appliance

import (
	"fmt"
	"net/http"
	"strings"
)

func (a *App) account(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 1 && parts[0] == "profile" {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, publicUser(session.User))
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			a.updateAccountProfile(w, r, session)
			return
		}
	}
	if len(parts) == 1 && parts[0] == "password" && r.Method == http.MethodPost {
		a.changeAccountPassword(w, r, session)
		return
	}
	writeError(w, http.StatusNotFound, "unknown account route")
}

func (a *App) updateAccountProfile(w http.ResponseWriter, r *http.Request, session Session) {
	var req struct {
		DisplayName string `json:"displayName"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid account profile payload")
		return
	}
	var updated User
	err := a.store.Mutate(func(data *AppData) error {
		displayName := strings.TrimSpace(req.DisplayName)
		if displayName == "" {
			return fmt.Errorf("显示名称不能为空")
		}
		for i := range data.Users {
			if data.Users[i].ID != session.User.ID {
				continue
			}
			data.Users[i].DisplayName = displayName
			updated = data.Users[i]
			addAudit(data, session.User.Username, "update_profile", "user", updated.ID, updated.Username, clientIP(r))
			return nil
		}
		return fmt.Errorf("用户不存在")
	})
	if err == nil {
		a.replaceUserInSessions(updated)
	}
	a.respondMutation(w, err, publicUser(updated), "account.profile.updated")
}

func (a *App) changeAccountPassword(w http.ResponseWriter, r *http.Request, session Session) {
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid account password payload")
		return
	}
	var updated User
	err := a.store.Mutate(func(data *AppData) error {
		newPassword := strings.TrimSpace(req.NewPassword)
		if len(newPassword) < 6 {
			return fmt.Errorf("新密码至少 6 位")
		}
		for i := range data.Users {
			if data.Users[i].ID != session.User.ID {
				continue
			}
			if !verifyPassword(req.CurrentPassword, data.Users[i]) {
				return fmt.Errorf("当前密码不正确")
			}
			data.Users[i].PasswordSalt, data.Users[i].PasswordHash = makePassword(newPassword)
			updated = data.Users[i]
			addAudit(data, session.User.Username, "change_password", "user", updated.ID, updated.Username, clientIP(r))
			return nil
		}
		return fmt.Errorf("用户不存在")
	})
	if err == nil {
		a.replaceUserInSessions(updated)
	}
	a.respondMutation(w, err, publicUser(updated), "account.password.changed")
}

func (a *App) replaceUserInSessions(user User) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for token, session := range a.sessions {
		if session.User.ID != user.ID {
			continue
		}
		session.User = user
		a.sessions[token] = session
	}
}
