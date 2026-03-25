import { useState } from "react";

import { fetchBookmarks, fetchFolders, fetchTags } from "../lib/api";

export function useBookmarksData({ onError }) {
  const [bookmarks, setBookmarks] = useState([]);
  const [folders, setFolders] = useState([]);
  const [tags, setTags] = useState([]);

  async function loadBookmarksWorkspace() {
    try {
      onError?.("");
      const [bookmarkItems, folderItems, tagItems] = await Promise.all([
        fetchBookmarks(),
        fetchFolders(),
        fetchTags(),
      ]);
      setBookmarks(Array.isArray(bookmarkItems) ? bookmarkItems : []);
      setFolders(Array.isArray(folderItems) ? folderItems : []);
      setTags(Array.isArray(tagItems) ? tagItems : []);
      return true;
    } catch (requestError) {
      onError?.(requestError.message);
      return false;
    }
  }

  return {
    bookmarks,
    folders,
    loadBookmarksWorkspace,
    tags,
  };
}
