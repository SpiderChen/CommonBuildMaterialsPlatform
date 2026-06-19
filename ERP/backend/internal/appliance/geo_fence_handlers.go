package appliance

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
)

func (a *App) geoFences(w http.ResponseWriter, r *http.Request, session Session, parts []string) {
	if len(parts) == 0 {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, normalizedGeoFences(scopedData(a.mustSnapshot(), session.User).GeoFences))
			return
		case http.MethodPost:
			a.createGeoFence(w, r, session)
			return
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}
	if len(parts) != 1 {
		writeError(w, http.StatusNotFound, "unknown fence route")
		return
	}
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	if id == 0 {
		writeError(w, http.StatusBadRequest, "invalid fence id")
		return
	}
	switch r.Method {
	case http.MethodPut, http.MethodPatch:
		a.updateGeoFence(w, r, session, id)
	case http.MethodDelete:
		a.archiveGeoFence(w, r, session, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) createGeoFence(w http.ResponseWriter, r *http.Request, session Session) {
	var item GeoFence
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid fence")
		return
	}
	err := a.store.Mutate(func(data *AppData) error {
		if err := validateGeoFence(*data, &item); err != nil {
			return err
		}
		item.ID = nextID(data, "fence")
		data.GeoFences = append(data.GeoFences, item)
		addAudit(data, session.User.Username, "create", "geo_fence", item.ID, item.Name, clientIP(r))
		return nil
	})
	a.respondMutation(w, err, item, "vehicle.fence.created")
}

func (a *App) updateGeoFence(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item GeoFence
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid fence")
		return
	}
	item.ID = id
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.GeoFences {
			if data.GeoFences[i].ID != id {
				continue
			}
			if err := validateGeoFence(*data, &item); err != nil {
				return err
			}
			data.GeoFences[i] = item
			addAudit(data, session.User.Username, "update", "geo_fence", item.ID, item.Name, clientIP(r))
			return nil
		}
		return errors.New("围栏不存在")
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit("vehicle.fence.updated", item)
	writeJSON(w, http.StatusOK, item)
}

func (a *App) archiveGeoFence(w http.ResponseWriter, r *http.Request, session Session, id int64) {
	var item GeoFence
	err := a.store.Mutate(func(data *AppData) error {
		for i := range data.GeoFences {
			if data.GeoFences[i].ID != id {
				continue
			}
			data.GeoFences[i] = normalizeGeoFence(data.GeoFences[i])
			data.GeoFences[i].Status = "inactive"
			item = data.GeoFences[i]
			addAudit(data, session.User.Username, "archive", "geo_fence", item.ID, item.Name, clientIP(r))
			return nil
		}
		return errors.New("围栏不存在")
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.emit("vehicle.fence.archived", item)
	writeJSON(w, http.StatusOK, item)
}

func (a *App) geoFenceEvents(w http.ResponseWriter, r *http.Request, session Session) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	data := scopedData(a.mustSnapshot(), session.User)
	vehicleID, _ := strconv.ParseInt(r.URL.Query().Get("vehicleId"), 10, 64)
	fenceID, _ := strconv.ParseInt(r.URL.Query().Get("fenceId"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	out := []GeoFenceEvent{}
	for _, item := range data.GeoFenceEvents {
		if vehicleID != 0 && item.VehicleID != vehicleID {
			continue
		}
		if fenceID != 0 && item.FenceID != fenceID {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].EventTime == out[j].EventTime {
			return out[i].ID > out[j].ID
		}
		return out[i].EventTime > out[j].EventTime
	})
	if len(out) > limit {
		out = out[:limit]
	}
	writeJSON(w, http.StatusOK, out)
}

func normalizedGeoFences(items []GeoFence) []GeoFence {
	out := make([]GeoFence, 0, len(items))
	for _, item := range items {
		out = append(out, normalizeGeoFence(item))
	}
	return out
}
