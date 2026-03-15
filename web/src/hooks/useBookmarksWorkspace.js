import { useDeferredValue, useMemo, useState } from "react";

function buildFolderTree(folders, bookmarks) {
  const byParent = new Map();
  const bookmarkCounts = new Map();

  for (const bookmark of bookmarks) {
    const key = bookmark.folderId || "";
    bookmarkCounts.set(key, (bookmarkCounts.get(key) || 0) + 1);
  }

  for (const folder of folders) {
    const key = folder.parentId || "";
    const items = byParent.get(key) || [];
    items.push({
      ...folder,
      bookmarkCount:
        typeof folder.bookmarkCount === "number"
          ? folder.bookmarkCount
          : bookmarkCounts.get(folder.id) || 0,
    });
    byParent.set(key, items);
  }

  for (const items of byParent.values()) {
    items.sort((left, right) => {
      if (left.position !== right.position) {
        return left.position - right.position;
      }
      return left.name.localeCompare(right.name);
    });
  }

  function attachChildren(parentId = "") {
    return (byParent.get(parentId) || []).map((folder) => ({
      ...folder,
      children: attachChildren(folder.id),
    }));
  }

  return attachChildren("");
}

function matchesQuery(bookmark, query) {
  if (!query) {
    return true;
  }
  const haystack = [
    bookmark.name,
    bookmark.deviceName,
    bookmark.serviceName,
    ...(bookmark.tags || []),
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();

  return haystack.includes(query);
}

export function useBookmarksWorkspace({ bookmarks, folders, tags }) {
  const [search, setSearch] = useState("");
  const [selectedFolderId, setSelectedFolderId] = useState("");
  const [selectedTag, setSelectedTag] = useState("");
  const [favoritesOnly, setFavoritesOnly] = useState(false);
  const deferredSearch = useDeferredValue(search.trim().toLowerCase());

  const filteredBookmarks = useMemo(() => {
    return bookmarks.filter((bookmark) => {
      if (favoritesOnly && !bookmark.isFavorite) {
        return false;
      }
      if (selectedFolderId && bookmark.folderId !== selectedFolderId) {
        return false;
      }
      if (
        selectedTag &&
        !(bookmark.tags || []).some(
          (tag) =>
            tag.toLowerCase() === selectedTag.toLowerCase() ||
            tag.toLowerCase() === selectedTag.replaceAll("-", " ").toLowerCase(),
        )
      ) {
        return false;
      }
      return matchesQuery(bookmark, deferredSearch);
    });
  }, [bookmarks, deferredSearch, favoritesOnly, selectedFolderId, selectedTag]);

  const favorites = useMemo(
    () =>
      filteredBookmarks
        .filter((bookmark) => bookmark.isFavorite)
        .sort((left, right) => {
          if (left.favoritePosition !== right.favoritePosition) {
            return left.favoritePosition - right.favoritePosition;
          }
          return left.name.localeCompare(right.name);
        }),
    [filteredBookmarks],
  );

  const folderTree = useMemo(
    () => buildFolderTree(folders, bookmarks),
    [bookmarks, folders],
  );

  const activeFolder = useMemo(
    () => folders.find((folder) => folder.id === selectedFolderId) || null,
    [folders, selectedFolderId],
  );

  return {
    activeFolder,
    favorites,
    favoritesOnly,
    filteredBookmarks,
    folderTree,
    search,
    selectedFolderId,
    selectedTag,
    setFavoritesOnly,
    setSearch,
    setSelectedFolderId,
    setSelectedTag,
    tags,
  };
}
