package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/deleema/homelabwatch/internal/domain"
)

func (r *Router) handleDockerEndpoints(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListDockerEndpoints(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.DockerEndpoint
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveDockerEndpoint(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleDockerEndpointByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var item domain.DockerEndpoint
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveDockerEndpoint(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteDockerEndpoint(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleScanTargets(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListScanTargets(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.ScanTarget
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveScanTarget(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleScanTargetByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		var item domain.ScanTarget
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveScanTarget(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteScanTarget(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleTopologySources(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListTopologySources(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.TopologySource
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveTopologySource(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleTopologySourceByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPatch:
		existing, err := r.app.GetTopologySourceForEdit(req.Context(), id)
		if err != nil {
			writeLookupError(w, err)
			return
		}
		item, err := patchTopologySource(req.Body, existing)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveTopologySource(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteTopologySource(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleTopologyDiscoveryRun(w http.ResponseWriter, req *http.Request) {
	if err := r.app.TriggerTopologyDiscovery(req.Context()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "topology-discovery-triggered"})
}

func (r *Router) handleTopologyAutoDiscover(w http.ResponseWriter, req *http.Request) {
	var input domain.TopologyAutoDiscoverInput
	if err := decodeJSON(req.Body, &input); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := r.app.AutoDiscoverAndRunTopology(req.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}

func (r *Router) handleDiscoveredServices(w http.ResponseWriter, req *http.Request) {
	items, err := r.app.ListDiscoveredServices(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func patchTopologySource(body io.ReadCloser, existing domain.TopologySource) (domain.TopologySource, error) {
	defer body.Close()
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return domain.TopologySource{}, err
	}
	item := existing
	for key, value := range raw {
		switch key {
		case "id", "hasCommunity", "hasAuthPassphrase", "hasPrivacyPassphrase", "lastSuccessAt", "lastError", "createdAt", "updatedAt":
			continue
		case "name":
			if err := json.Unmarshal(value, &item.Name); err != nil {
				return domain.TopologySource{}, err
			}
		case "address":
			if err := json.Unmarshal(value, &item.Address); err != nil {
				return domain.TopologySource{}, err
			}
		case "port":
			if err := json.Unmarshal(value, &item.Port); err != nil {
				return domain.TopologySource{}, err
			}
		case "enabled":
			if err := json.Unmarshal(value, &item.Enabled); err != nil {
				return domain.TopologySource{}, err
			}
		case "pollIntervalSeconds":
			if err := json.Unmarshal(value, &item.PollIntervalSeconds); err != nil {
				return domain.TopologySource{}, err
			}
		case "timeoutMs":
			if err := json.Unmarshal(value, &item.TimeoutMS); err != nil {
				return domain.TopologySource{}, err
			}
		case "retries":
			if err := json.Unmarshal(value, &item.Retries); err != nil {
				return domain.TopologySource{}, err
			}
		case "snmpVersion":
			if err := json.Unmarshal(value, &item.SNMPVersion); err != nil {
				return domain.TopologySource{}, err
			}
		case "community":
			if err := json.Unmarshal(value, &item.Community); err != nil {
				return domain.TopologySource{}, err
			}
		case "username":
			if err := json.Unmarshal(value, &item.Username); err != nil {
				return domain.TopologySource{}, err
			}
		case "authProtocol":
			if err := json.Unmarshal(value, &item.AuthProtocol); err != nil {
				return domain.TopologySource{}, err
			}
		case "authPassphrase":
			if err := json.Unmarshal(value, &item.AuthPassphrase); err != nil {
				return domain.TopologySource{}, err
			}
		case "privacyProtocol":
			if err := json.Unmarshal(value, &item.PrivacyProtocol); err != nil {
				return domain.TopologySource{}, err
			}
		case "privacyPassphrase":
			if err := json.Unmarshal(value, &item.PrivacyPassphrase); err != nil {
				return domain.TopologySource{}, err
			}
		case "role":
			if err := json.Unmarshal(value, &item.Role); err != nil {
				return domain.TopologySource{}, err
			}
		case "root":
			if err := json.Unmarshal(value, &item.Root); err != nil {
				return domain.TopologySource{}, err
			}
		default:
			return domain.TopologySource{}, errors.New("unknown topology source field: " + key)
		}
	}
	return item, nil
}

func (r *Router) handleDiscoveredServiceBookmark(w http.ResponseWriter, req *http.Request) {
	var input domain.CreateBookmarkFromDiscoveredServiceInput
	if err := decodeJSON(req.Body, &input); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	input.DiscoveredServiceID = req.PathValue("id")
	item, err := r.app.CreateBookmarkFromDiscoveredService(req.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (r *Router) handleDiscoveredServiceIgnore(w http.ResponseWriter, req *http.Request) {
	item, err := r.app.IgnoreDiscoveredService(req.Context(), req.PathValue("id"))
	if err != nil {
		writeLookupError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleDiscoveredServiceRestore(w http.ResponseWriter, req *http.Request) {
	item, err := r.app.RestoreDiscoveredService(req.Context(), req.PathValue("id"))
	if err != nil {
		writeLookupError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleDiscoverySettings(w http.ResponseWriter, req *http.Request) {
	var input domain.DiscoverySettings
	if err := decodeJSON(req.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := r.app.SaveDiscoverySettings(req.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleDiscoveryRun(w http.ResponseWriter, req *http.Request) {
	if err := r.app.TriggerDiscovery(req.Context()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "discovery-triggered"})
}

func (r *Router) handleMonitoringRun(w http.ResponseWriter, req *http.Request) {
	if err := r.app.TriggerMonitoring(req.Context()); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "monitoring-triggered"})
}
