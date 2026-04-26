import { useState } from "react";

import { fetchStatusPage, fetchStatusPages } from "../lib/api";

export function useStatusPagesData({ onError } = {}) {
  const [statusPages, setStatusPages] = useState({ list: [], selected: null });

  async function loadStatusPages(selectedId = statusPages.selected?.id) {
    try {
      const list = await fetchStatusPages();
      let selected = null;
      const nextId = selectedId || list[0]?.id || "";
      if (nextId) {
        selected = await fetchStatusPage(nextId).catch(() => null);
      }
      setStatusPages({ list, selected });
      return { list, selected };
    } catch (error) {
      onError?.(error.message);
      return null;
    }
  }

  async function loadStatusPage(id) {
    try {
      const selected = await fetchStatusPage(id);
      setStatusPages((current) => ({ ...current, selected }));
      return selected;
    } catch (error) {
      onError?.(error.message);
      return null;
    }
  }

  return { loadStatusPage, loadStatusPages, statusPages };
}
