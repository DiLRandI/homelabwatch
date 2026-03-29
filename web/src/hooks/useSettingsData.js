import { useRef, useState } from "react";

import { fetchSettings } from "../lib/api";

export function useSettingsData({ onError, onTrustedNetworkChange }) {
  const [settings, setSettings] = useState(null);
  const inFlightRef = useRef(null);
  const requestIdRef = useRef(0);

  async function loadSettings() {
    if (inFlightRef.current) {
      return inFlightRef.current;
    }

    const requestID = requestIdRef.current + 1;
    requestIdRef.current = requestID;

    const request = (async () => {
      try {
        onError?.("");
        const payload = await fetchSettings();
        if (requestIdRef.current === requestID) {
          setSettings(payload);
          onTrustedNetworkChange?.(Boolean(payload?.appSettings?.trustedNetwork));
        }
        return payload;
      } catch (requestError) {
        if (requestIdRef.current === requestID) {
          onError?.(requestError.message);
        }
        return null;
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
    loadSettings,
    settings,
  };
}
