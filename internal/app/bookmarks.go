package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/deleema/homelabwatch/internal/domain"
)

func (a *App) QueryBookmarks(ctx context.Context, options domain.BookmarkListOptions) ([]domain.Bookmark, error) {
	return a.store.QueryBookmarks(ctx, options)
}

func (a *App) ListBookmarks(ctx context.Context) ([]domain.Bookmark, error) {
	return a.store.ListBookmarks(ctx)
}

func (a *App) GetBookmark(ctx context.Context, id string) (domain.Bookmark, error) {
	return a.store.GetBookmark(ctx, id)
}

func (a *App) SaveBookmark(ctx context.Context, input domain.BookmarkInput) (domain.Bookmark, error) {
	item, err := a.store.SaveBookmark(ctx, input)
	if err != nil {
		return domain.Bookmark{}, err
	}
	a.publish("bookmark", item.ID, "upserted", item)
	return item, nil
}

func (a *App) CreateBookmarkFromService(ctx context.Context, input domain.CreateBookmarkFromServiceInput) (domain.Bookmark, error) {
	item, err := a.store.CreateBookmarkFromService(ctx, input)
	if err != nil {
		return domain.Bookmark{}, err
	}
	a.publish("bookmark", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteBookmark(ctx context.Context, id string) error {
	if err := a.store.DeleteBookmark(ctx, id); err != nil {
		return err
	}
	a.publish("bookmark", id, "deleted", nil)
	return nil
}

func (a *App) OpenBookmark(ctx context.Context, id string) (domain.Bookmark, error) {
	item, err := a.store.OpenBookmark(ctx, id)
	if err != nil {
		return domain.Bookmark{}, err
	}
	a.publish("bookmark", item.ID, "opened", item)
	return item, nil
}

func (a *App) ReorderBookmarks(ctx context.Context, items []domain.BookmarkReorderItem) error {
	if err := a.store.ReorderBookmarks(ctx, items); err != nil {
		return err
	}
	a.publish("bookmark", "all", "reordered", items)
	return nil
}

func (a *App) ListFolders(ctx context.Context) ([]domain.Folder, error) {
	return a.store.ListFolders(ctx)
}

func (a *App) SaveFolder(ctx context.Context, input domain.FolderInput) (domain.Folder, error) {
	item, err := a.store.SaveFolder(ctx, input)
	if err != nil {
		return domain.Folder{}, err
	}
	a.publish("folder", item.ID, "upserted", item)
	return item, nil
}

func (a *App) DeleteFolder(ctx context.Context, id string) error {
	if err := a.store.DeleteFolder(ctx, id); err != nil {
		return err
	}
	a.publish("folder", id, "deleted", nil)
	return nil
}

func (a *App) ReorderFolders(ctx context.Context, items []domain.FolderReorderItem) error {
	if err := a.store.ReorderFolders(ctx, items); err != nil {
		return err
	}
	a.publish("folder", "all", "reordered", items)
	return nil
}

func (a *App) ListTags(ctx context.Context) ([]domain.Tag, error) {
	return a.store.ListTags(ctx)
}

func (a *App) SaveBookmarkAsset(filename string, data []byte) (string, string, error) {
	if len(data) == 0 {
		return "", "", errors.New("bookmark asset is empty")
	}
	if err := os.MkdirAll(a.bookmarkAssetsDir(), 0o755); err != nil {
		return "", "", err
	}
	extension := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	if extension == "" {
		extension = ".bin"
	}
	assetName := randomName("bookmark_asset") + extension
	target := filepath.Join(a.bookmarkAssetsDir(), assetName)
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return "", "", err
	}
	contentType := http.DetectContentType(data)
	return a.BookmarkAssetURL(assetName), contentType, nil
}

func (a *App) LoadBookmarkAsset(name string) ([]byte, string, error) {
	cleanName := filepath.Base(strings.TrimSpace(name))
	if cleanName == "." || cleanName == "" {
		return nil, "", os.ErrNotExist
	}
	data, err := os.ReadFile(filepath.Join(a.bookmarkAssetsDir(), cleanName))
	if err != nil {
		return nil, "", err
	}
	return data, http.DetectContentType(data), nil
}

func (a *App) BookmarkAssetURL(name string) string {
	return "/api/ui/v1/bookmark-assets/" + filepath.Base(strings.TrimSpace(name))
}

func (a *App) ExportBookmarks(ctx context.Context) (domain.BookmarkExport, error) {
	folders, err := a.store.ListFolders(ctx)
	if err != nil {
		return domain.BookmarkExport{}, err
	}
	tags, err := a.store.ListTags(ctx)
	if err != nil {
		return domain.BookmarkExport{}, err
	}
	items, err := a.store.ListBookmarks(ctx)
	if err != nil {
		return domain.BookmarkExport{}, err
	}
	payload := domain.BookmarkExport{
		Folders: make([]domain.FolderExportItem, 0, len(folders)),
		Tags:    make([]domain.TagExportItem, 0, len(tags)),
		Items:   make([]domain.BookmarkExportItem, 0, len(items)),
	}
	for _, folder := range folders {
		payload.Folders = append(payload.Folders, domain.FolderExportItem{
			ID:       folder.ID,
			ParentID: folder.ParentID,
			Name:     folder.Name,
			Position: folder.Position,
		})
	}
	for _, tag := range tags {
		payload.Tags = append(payload.Tags, domain.TagExportItem{
			ID:   tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}
	for _, item := range items {
		payload.Items = append(payload.Items, domain.BookmarkExportItem{
			ID:                      item.ID,
			FolderID:                item.FolderID,
			ServiceID:               item.ServiceID,
			ServiceSource:           item.ServiceSource,
			ServiceSourceRef:        item.ServiceSourceRef,
			DeviceID:                item.DeviceID,
			Name:                    item.ManualName,
			URL:                     item.ManualURL,
			Description:             item.Description,
			IconMode:                item.IconMode,
			IconValue:               item.IconValue,
			Tags:                    append([]string(nil), item.Tags...),
			IsFavorite:              item.IsFavorite,
			FavoritePosition:        item.FavoritePosition,
			Position:                item.Position,
			ClickCount:              item.ClickCount,
			LastOpenedAt:            item.LastOpenedAt,
			UseDevicePrimaryAddress: item.UseDevicePrimaryAddress,
		})
		if item.IconMode == "uploaded" && item.IconValue != "" {
			data, contentType, assetErr := a.LoadBookmarkAsset(filepath.Base(item.IconValue))
			if assetErr == nil {
				payload.Assets = append(payload.Assets, domain.BookmarkAsset{
					Name:        filepath.Base(item.IconValue),
					ContentType: contentType,
					Data:        base64.StdEncoding.EncodeToString(data),
				})
			}
		}
	}
	return payload, nil
}

func (a *App) ImportBookmarks(ctx context.Context, payload domain.BookmarkImport) (domain.BookmarkImportResult, error) {
	result := domain.BookmarkImportResult{}
	assetURLs := map[string]string{}
	for _, asset := range payload.Assets {
		data, err := base64.StdEncoding.DecodeString(asset.Data)
		if err != nil {
			return result, err
		}
		assetURL, _, err := a.SaveBookmarkAsset(asset.Name, data)
		if err != nil {
			return result, err
		}
		assetURLs[asset.Name] = assetURL
		result.AssetsImported++
	}
	folders := append([]domain.FolderExportItem(nil), payload.Folders...)
	slices.SortFunc(folders, func(a, b domain.FolderExportItem) int {
		return strings.Compare(a.ParentID+a.ID, b.ParentID+b.ID)
	})
	for _, folder := range folders {
		if _, err := a.store.SaveFolder(ctx, domain.FolderInput{
			ID:       folder.ID,
			ParentID: folder.ParentID,
			Name:     folder.Name,
			Position: folder.Position,
		}); err != nil {
			return result, err
		}
		result.FoldersImported++
	}
	for _, item := range payload.Items {
		input := domain.BookmarkInput{
			ID:                      item.ID,
			FolderID:                item.FolderID,
			DeviceID:                item.DeviceID,
			Name:                    item.Name,
			URL:                     item.URL,
			Description:             item.Description,
			Tags:                    item.Tags,
			IconMode:                item.IconMode,
			IconValue:               item.IconValue,
			IsFavorite:              item.IsFavorite,
			FavoritePosition:        item.FavoritePosition,
			Position:                item.Position,
			UseDevicePrimaryAddress: item.UseDevicePrimaryAddress,
		}
		if input.IconMode == "uploaded" {
			if importedURL := assetURLs[filepath.Base(input.IconValue)]; importedURL != "" {
				input.IconValue = importedURL
			}
		}
		if item.ServiceSourceRef != "" {
			service, err := a.store.FindServiceBySource(ctx, item.ServiceSource, item.ServiceSourceRef)
			if err == nil {
				input.ServiceID = service.ID
				input.Name = firstNonEmpty(item.Name, service.Name)
			}
		}
		if _, err := a.store.SaveBookmark(ctx, input); err != nil {
			return result, err
		}
		result.BookmarksImported++
	}
	return result, nil
}

func (a *App) bookmarkAssetsDir() string {
	return filepath.Join(a.config.DataDir, "bookmark-assets")
}

func randomName(prefix string) string {
	buffer := make([]byte, 6)
	_, _ = rand.Read(buffer)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(buffer))
}
