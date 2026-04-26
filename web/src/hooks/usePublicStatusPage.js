import { useEffect, useState } from "react";

import { fetchPublicStatusPage } from "../lib/api";

export function usePublicStatusPage(slug) {
  const [state, setState] = useState({
    error: "",
    loading: true,
    missing: false,
    page: null,
  });

  async function load() {
    if (!slug) {
      setState({ error: "Status page not found.", loading: false, missing: true, page: null });
      return;
    }
    try {
      const page = await fetchPublicStatusPage(slug);
      setState({ error: "", loading: false, missing: false, page });
    } catch (error) {
      setState({
        error: error.message,
        loading: false,
        missing: /not found|no rows/i.test(error.message),
        page: null,
      });
    }
  }

  useEffect(() => {
    let active = true;
    async function guardedLoad() {
      if (active) {
        await load();
      }
    }
    void guardedLoad();
    const intervalID = window.setInterval(guardedLoad, 30000);
    function handleVisibility() {
      if (document.visibilityState === "visible") {
        void guardedLoad();
      }
    }
    document.addEventListener("visibilitychange", handleVisibility);
    return () => {
      active = false;
      window.clearInterval(intervalID);
      document.removeEventListener("visibilitychange", handleVisibility);
    };
  }, [slug]);

  return { ...state, refresh: load };
}
