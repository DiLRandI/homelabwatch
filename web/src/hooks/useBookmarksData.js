import { useRef, useState } from "react";

import { fetchBookmarks, fetchFolders, fetchTags } from "../lib/api";

export function useBookmarksData({ onError }) {
  const [bookmarks, setBookmarks] = useState([]);
  const [folders, setFolders] = useState([]);
  const [tags, setTags] = useState([]);
  const inFlightRef = useRef(null);
  const requestIdRef = useRef(0);

  async function loadBookmarksWorkspace() {
    if (inFlightRef.current) {
      return inFlightRef.current;
    }

    const requestID = requestIdRef.current + 1;
    requestIdRef.current = requestID;

    const request = (async () => {
      try {
        onError?.("");
        const [bookmarkItems, folderItems, tagItems] = await Promise.all([
          fetchBookmarks(),
          fetchFolders(),
          fetchTags(),
        ]);
        if (requestIdRef.current === requestID) {
          setBookmarks(Array.isArray(bookmarkItems) ? bookmarkItems : []);
          setFolders(Array.isArray(folderItems) ? folderItems : []);
          setTags(Array.isArray(tagItems) ? tagItems : []);
        }
        return true;
      } catch (requestError) {
        if (requestIdRef.current === requestID) {
          onError?.(requestError.message);
        }
        return false;
      } finally {
        if (inFlightRef.current === request) {
          inFlightRef.current = null;
        }
      }
    })();

    inFlightRef.current = request;
    return request;
  }

  return {
    bookmarks,
    folders,
    loadBookmarksWorkspace,
    tags,
  };
}
