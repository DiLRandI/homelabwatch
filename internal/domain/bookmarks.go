package domain

import "time"

type BookmarkInput struct {
	ID                      string                `json:"id,omitempty"`
	FolderID                string                `json:"folderId,omitempty"`
	ServiceID               string                `json:"serviceId,omitempty"`
	DeviceID                string                `json:"deviceId,omitempty"`
	Name                    string                `json:"name,omitempty"`
	URL                     string                `json:"url,omitempty"`
	Description             string                `json:"description,omitempty"`
	Tags                    []string              `json:"tags,omitempty"`
	IconMode                string                `json:"iconMode,omitempty"`
	IconValue               string                `json:"iconValue,omitempty"`
	IsFavorite              bool                  `json:"isFavorite"`
	FavoritePosition        int                   `json:"favoritePosition,omitempty"`
	Position                int                   `json:"position,omitempty"`
	UseDevicePrimaryAddress bool                  `json:"useDevicePrimaryAddress,omitempty"`
	Monitor                 *BookmarkMonitorInput `json:"monitor,omitempty"`
}

type BookmarkMonitorInput struct {
	Enabled        bool   `json:"enabled"`
	ServiceID      string `json:"serviceId,omitempty"`
	ServiceName    string `json:"serviceName,omitempty"`
	ServiceVisible bool   `json:"serviceVisible"`
}

type CreateBookmarkFromServiceInput struct {
	ServiceID        string   `json:"serviceId"`
	FolderID         string   `json:"folderId,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	Name             string   `json:"name,omitempty"`
	IconMode         string   `json:"iconMode,omitempty"`
	IconValue        string   `json:"iconValue,omitempty"`
	IsFavorite       bool     `json:"isFavorite"`
	FavoritePosition int      `json:"favoritePosition,omitempty"`
}

type CreateBookmarkFromDiscoveredServiceInput struct {
	DiscoveredServiceID string   `json:"discoveredServiceId"`
	FolderID            string   `json:"folderId,omitempty"`
	Tags                []string `json:"tags,omitempty"`
	Name                string   `json:"name,omitempty"`
	IconMode            string   `json:"iconMode,omitempty"`
	IconValue           string   `json:"iconValue,omitempty"`
	IsFavorite          bool     `json:"isFavorite"`
	FavoritePosition    int      `json:"favoritePosition,omitempty"`
}

type BookmarkListOptions struct {
	Query     string
	FolderID  string
	Tag       string
	Favorites *bool
	DeviceID  string
	ServiceID string
}

type BookmarkReorderItem struct {
	ID               string `json:"id"`
	FolderID         string `json:"folderId,omitempty"`
	Position         int    `json:"position"`
	IsFavorite       bool   `json:"isFavorite"`
	FavoritePosition int    `json:"favoritePosition,omitempty"`
}

type Folder struct {
	ID            string    `json:"id"`
	ParentID      string    `json:"parentId,omitempty"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Position      int       `json:"position"`
	BookmarkCount int       `json:"bookmarkCount"`
	ChildCount    int       `json:"childCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type FolderInput struct {
	ID       string `json:"id,omitempty"`
	ParentID string `json:"parentId,omitempty"`
	Name     string `json:"name"`
	Position int    `json:"position,omitempty"`
}

type FolderReorderItem struct {
	ID       string `json:"id"`
	ParentID string `json:"parentId,omitempty"`
	Position int    `json:"position"`
}

type Tag struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	BookmarkCount int       `json:"bookmarkCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type BookmarkExport struct {
	Folders []FolderExportItem   `json:"folders"`
	Tags    []TagExportItem      `json:"tags"`
	Assets  []BookmarkAsset      `json:"assets,omitempty"`
	Items   []BookmarkExportItem `json:"items"`
}

type FolderExportItem struct {
	ID       string `json:"id"`
	ParentID string `json:"parentId,omitempty"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type TagExportItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type BookmarkAsset struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType,omitempty"`
	Data        string `json:"data"`
}

type BookmarkExportItem struct {
	ID                      string        `json:"id"`
	FolderID                string        `json:"folderId,omitempty"`
	ServiceID               string        `json:"serviceId,omitempty"`
	ServiceSource           ServiceSource `json:"serviceSource,omitempty"`
	ServiceSourceRef        string        `json:"serviceSourceRef,omitempty"`
	DeviceID                string        `json:"deviceId,omitempty"`
	Name                    string        `json:"name,omitempty"`
	URL                     string        `json:"url,omitempty"`
	Description             string        `json:"description,omitempty"`
	IconMode                string        `json:"iconMode,omitempty"`
	IconValue               string        `json:"iconValue,omitempty"`
	Tags                    []string      `json:"tags,omitempty"`
	IsFavorite              bool          `json:"isFavorite"`
	FavoritePosition        int           `json:"favoritePosition,omitempty"`
	Position                int           `json:"position"`
	ClickCount              int           `json:"clickCount"`
	LastOpenedAt            time.Time     `json:"lastOpenedAt"`
	UseDevicePrimaryAddress bool          `json:"useDevicePrimaryAddress"`
}

type BookmarkImport struct {
	Folders []FolderExportItem   `json:"folders"`
	Assets  []BookmarkAsset      `json:"assets,omitempty"`
	Items   []BookmarkExportItem `json:"items"`
}

type BookmarkImportResult struct {
	FoldersImported   int `json:"foldersImported"`
	BookmarksImported int `json:"bookmarksImported"`
	AssetsImported    int `json:"assetsImported"`
}
