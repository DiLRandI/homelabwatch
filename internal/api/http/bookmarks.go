package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/deleema/homelabwatch/internal/domain"
)

func (r *Router) handleBookmarkAssets(w http.ResponseWriter, req *http.Request) {
	file, header, err := req.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	assetURL, contentType, err := r.app.SaveBookmarkAsset(header.Filename, data)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{
		"url":         assetURL,
		"contentType": contentType,
	})
}

func (r *Router) handleBookmarks(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		options := domain.BookmarkListOptions{
			Query:     strings.TrimSpace(req.URL.Query().Get("q")),
			FolderID:  strings.TrimSpace(req.URL.Query().Get("folderId")),
			Tag:       strings.TrimSpace(req.URL.Query().Get("tag")),
			DeviceID:  strings.TrimSpace(req.URL.Query().Get("deviceId")),
			ServiceID: strings.TrimSpace(req.URL.Query().Get("serviceId")),
		}
		if rawFavorites := strings.TrimSpace(req.URL.Query().Get("favorites")); rawFavorites != "" {
			value := rawFavorites == "true"
			options.Favorites = &value
		}
		items, err := r.app.QueryBookmarks(req.Context(), options)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var item domain.BookmarkInput
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		saved, err := r.app.SaveBookmark(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleBookmarkByID(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	switch req.Method {
	case http.MethodPut, http.MethodPatch:
		var item domain.BookmarkInput
		if err := decodeJSON(req.Body, &item); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item.ID = id
		saved, err := r.app.SaveBookmark(req.Context(), item)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	case http.MethodDelete:
		if err := r.app.DeleteBookmark(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleBookmarkAssetByName(w http.ResponseWriter, req *http.Request) {
	data, contentType, err := r.app.LoadBookmarkAsset(req.PathValue("name"))
	if err != nil {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (r *Router) handleBookmarkFromService(w http.ResponseWriter, req *http.Request) {
	var input domain.CreateBookmarkFromServiceInput
	if err := decodeJSON(req.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := r.app.CreateBookmarkFromService(req.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (r *Router) handleBookmarkReorder(w http.ResponseWriter, req *http.Request) {
	var input []domain.BookmarkReorderItem
	if err := decodeJSON(req.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := r.app.ReorderBookmarks(req.Context(), input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handleBookmarkImport(w http.ResponseWriter, req *http.Request) {
	var payload domain.BookmarkImport
	if err := decodeJSON(req.Body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := r.app.ImportBookmarks(req.Context(), payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (r *Router) handleBookmarkExport(w http.ResponseWriter, req *http.Request) {
	payload, err := r.app.ExportBookmarks(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="homelabwatch-bookmarks.json"`)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func (r *Router) handleBookmarkOpen(w http.ResponseWriter, req *http.Request) {
	item, err := r.app.OpenBookmark(req.Context(), req.PathValue("id"))
	if err != nil {
		writeLookupError(w, err)
		return
	}
	http.Redirect(w, req, item.URL, http.StatusFound)
}

func (r *Router) handleFolders(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		items, err := r.app.ListFolders(req.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input domain.FolderInput
		if err := decodeJSON(req.Body, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := r.app.SaveFolder(req.Context(), input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleFolderByID(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimSpace(req.PathValue("id"))
	switch req.Method {
	case http.MethodPut:
		var input domain.FolderInput
		if err := decodeJSON(req.Body, &input); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		input.ID = id
		item, err := r.app.SaveFolder(req.Context(), input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if err := r.app.DeleteFolder(req.Context(), id); err != nil {
			writeLookupError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleFolderReorder(w http.ResponseWriter, req *http.Request) {
	var input []domain.FolderReorderItem
	if err := decodeJSON(req.Body, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := r.app.ReorderFolders(req.Context(), input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handleTags(w http.ResponseWriter, req *http.Request) {
	items, err := r.app.ListTags(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}
