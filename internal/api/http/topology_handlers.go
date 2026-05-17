package httpapi

import "net/http"

func (r *Router) handleTopology(w http.ResponseWriter, req *http.Request) {
	topology, err := r.app.Topology(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, topology)
}
