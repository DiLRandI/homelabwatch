import { useEffect, useState } from "react";

import { fetchUIBootstrap } from "../lib/api";

export function useUIBootstrap({ onError }) {
  const [initialized, setInitialized] = useState(false);
  const [trustedNetwork, setTrustedNetwork] = useState(false);
  const [csrfToken, setCsrfToken] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    void loadBootstrapState();
  }, []);

  async function loadBootstrapState() {
    try {
      setLoading(true);
      onError?.("");
      const payload = await fetchUIBootstrap();
      setInitialized(Boolean(payload.initialized));
      setTrustedNetwork(Boolean(payload.trustedNetwork));
      setCsrfToken(payload.csrfToken || "");
      return payload;
    } catch (requestError) {
      onError?.(requestError.message);
      return null;
    } finally {
      setLoading(false);
    }
  }

  return {
    csrfToken,
    initialized,
    loadBootstrapState,
    loading,
    markInitialized() {
      setInitialized(true);
    },
    setTrustedNetwork,
    trustedNetwork,
  };
}
