import { useState } from "react";

import { fetchSettings } from "../lib/api";

export function useSettingsData({ onError, onTrustedNetworkChange }) {
  const [settings, setSettings] = useState(null);

  async function loadSettings() {
    try {
      onError?.("");
      const payload = await fetchSettings();
      setSettings(payload);
      onTrustedNetworkChange?.(Boolean(payload?.appSettings?.trustedNetwork));
      return payload;
    } catch (requestError) {
      onError?.(requestError.message);
      return null;
    }
  }

  return {
    loadSettings,
    settings,
  };
}
