package appliance

import (
	"net/http"
)

type locationBatchReportRequest struct {
	Reports []locationReportPayload `json:"reports"`
}

type locationBatchReportItemResult struct {
	Index    int                   `json:"index"`
	Status   string                `json:"status"`
	Error    string                `json:"error,omitempty"`
	Location *VehicleLocationEvent `json:"location,omitempty"`
}

type locationBatchReportResponse struct {
	Total    int                             `json:"total"`
	Accepted int                             `json:"accepted"`
	Rejected int                             `json:"rejected"`
	Results  []locationBatchReportItemResult `json:"results"`
}

func (a *App) reportLocationBatch(w http.ResponseWriter, r *http.Request, session Session) {
	var req locationBatchReportRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid location batch")
		return
	}
	if len(req.Reports) == 0 {
		writeError(w, http.StatusBadRequest, "location batch is empty")
		return
	}
	if len(req.Reports) > 500 {
		writeError(w, http.StatusBadRequest, "location batch exceeds 500 points")
		return
	}
	out := locationBatchReportResponse{Total: len(req.Reports)}
	for index, report := range req.Reports {
		event, latest, err := a.recordLocationReport(r, session, report)
		if err != nil {
			out.Rejected++
			out.Results = append(out.Results, locationBatchReportItemResult{Index: index, Status: "rejected", Error: err.Error()})
			continue
		}
		a.runtime.CacheLatestLocation(latest)
		a.runtime.StoreTrackPoint(event)
		out.Accepted++
		out.Results = append(out.Results, locationBatchReportItemResult{Index: index, Status: "accepted", Location: &event})
	}
	writeJSON(w, http.StatusCreated, out)
}
